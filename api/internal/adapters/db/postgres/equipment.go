package postgres

import (
	"context"
	"log/slog"
	"time"

	"atom-maintenance/internal/adapters/db/postgres/model"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg"
	"atom-maintenance/platform/logger"

	"gorm.io/gorm"
)

func toEquipmentDomain(m model.Equipment) domain.Equipment {
	return domain.Equipment{
		ID:              m.ID,
		Name:            m.Name,
		Description:     m.Description,
		SerialNumber:    m.SerialNumber,
		EquipmentTypeID: m.EquipmentTypeID,
		DepartmentID:    m.DepartmentID,
		ResponsibleID:   m.ResponsibleID,
		Status:          domain.EquipmentStatus(m.Status),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

type EquipmentRepo struct {
	db      *gorm.DB
	log     *slog.Logger
	queryTO time.Duration
}

func NewEquipmentRepo(db *gorm.DB, log *slog.Logger, queryTO time.Duration) *EquipmentRepo {
	return &EquipmentRepo{db: db, log: log, queryTO: queryTO}
}

func (r *EquipmentRepo) Create(ctx context.Context, e *domain.Equipment) (*domain.Equipment, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.equipment", "op", "Create")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	m := model.Equipment{
		Name:            e.Name,
		Description:     e.Description,
		SerialNumber:    e.SerialNumber,
		EquipmentTypeID: e.EquipmentTypeID,
		DepartmentID:    e.DepartmentID,
		ResponsibleID:   e.ResponsibleID,
		Status:          string(e.Status),
	}
	if err := db.Create(&m).Error; err != nil {
		mapped := pkg.MapDB(err)
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil, mapped
	}

	out := toEquipmentDomain(m)
	log.Info("equipment created",
		"id", m.ID,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return &out, nil
}

func (r *EquipmentRepo) GetByID(ctx context.Context, id int64) (*domain.Equipment, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.equipment", "op", "GetByID")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var m model.Equipment
	if err := db.First(&m, id).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := toEquipmentDomain(m)
	log.Info("equipment fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *EquipmentRepo) List(ctx context.Context) ([]domain.Equipment, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.equipment", "op", "List")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.Equipment
	if err := db.Order("id asc").Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.Equipment, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEquipmentDomain(m))
	}

	log.Info("equipment listed", "count", len(out), "duration_ms", time.Since(start).Milliseconds())
	return out, nil
}

func (r *EquipmentRepo) Update(ctx context.Context, e *domain.Equipment) (*domain.Equipment, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.equipment", "op", "Update")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Model(&model.Equipment{}).
		Where("id = ?", e.ID).
		Updates(map[string]any{
			"name":              e.Name,
			"description":       e.Description,
			"serial_number":     e.SerialNumber,
			"equipment_type_id": e.EquipmentTypeID,
			"department_id":     e.DepartmentID,
			"responsible_id":    e.ResponsibleID,
			"status":            string(e.Status),
			"updated_at":        time.Now(),
		})

	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("update failed",
			"id", e.ID,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, mapped
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, e.ID)
}

func (r *EquipmentRepo) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.equipment", "op", "Delete")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Delete(&model.Equipment{}, id)
	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("delete failed",
			"id", id,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds())

		return mapped
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	log.Info("equipment deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}
