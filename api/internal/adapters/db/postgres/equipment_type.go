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

func toEquipmentTypeDomain(m model.EquipmentType) domain.EquipmentType {
	return domain.EquipmentType{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
}

type EquipmentTypeRepo struct {
	db      *gorm.DB
	log     *slog.Logger
	queryTO time.Duration
}

func NewEquipmentTypeRepo(db *gorm.DB, log *slog.Logger, queryTO time.Duration) *EquipmentTypeRepo {
	return &EquipmentTypeRepo{db: db, log: log, queryTO: queryTO}
}

func (r *EquipmentTypeRepo) Create(ctx context.Context, et *domain.EquipmentType) (*domain.EquipmentType, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.equipment_type", "op", "Create")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	m := model.EquipmentType{Name: et.Name, Description: et.Description}
	if err := db.Create(&m).Error; err != nil {
		mapped := pkg.MapDB(err)
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil, mapped
	}

	out := toEquipmentTypeDomain(m)
	log.Info("equipment_type created",
		"id", m.ID,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return &out, nil
}

func (r *EquipmentTypeRepo) GetByID(ctx context.Context, id int64) (*domain.EquipmentType, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.equipment_type", "op", "GetByID")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var m model.EquipmentType
	if err := db.First(&m, id).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := toEquipmentTypeDomain(m)
	log.Info("equipment_type fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *EquipmentTypeRepo) List(ctx context.Context) ([]domain.EquipmentType, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.equipment_type", "op", "List")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.EquipmentType
	if err := db.Order("id asc").Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.EquipmentType, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEquipmentTypeDomain(m))
	}

	log.Info("equipment_types listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())
	return out, nil
}

func (r *EquipmentTypeRepo) Update(ctx context.Context, et *domain.EquipmentType) (*domain.EquipmentType, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.equipment_type", "op", "Update")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Model(&model.EquipmentType{}).
		Where("id = ?", et.ID).
		Updates(map[string]any{
			"name":        et.Name,
			"description": et.Description,
		})

	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("update failed",
			"id", et.ID,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, mapped
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, et.ID)
}

func (r *EquipmentTypeRepo) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.equipment_type", "op", "Delete")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Delete(&model.EquipmentType{}, id)
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

	log.Info("equipment_type deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}
