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

func toDepartmentDomain(m model.Department) domain.Department {
	return domain.Department{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
}

type DepartmentRepo struct {
	db      *gorm.DB
	log     *slog.Logger
	queryTO time.Duration
}

func NewDepartmentRepo(db *gorm.DB, log *slog.Logger, queryTO time.Duration) *DepartmentRepo {
	return &DepartmentRepo{db: db, log: log, queryTO: queryTO}
}

func (r *DepartmentRepo) Create(ctx context.Context, d *domain.Department) (*domain.Department, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.department", "op", "Create")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	m := model.Department{Name: d.Name, Description: d.Description}
	if err := db.Create(&m).Error; err != nil {
		mapped := pkg.MapDB(err)
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil, mapped
	}

	out := toDepartmentDomain(m)
	log.Info("department created",
		"id", m.ID,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return &out, nil
}

func (r *DepartmentRepo) GetByID(ctx context.Context, id int64) (*domain.Department, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.department", "op", "GetByID")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var m model.Department
	if err := db.First(&m, id).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := toDepartmentDomain(m)
	log.Info("department fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return &out, nil
}

func (r *DepartmentRepo) List(ctx context.Context) ([]domain.Department, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.department", "op", "List")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.Department
	if err := db.Order("id asc").Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.Department, 0, len(rows))
	for _, m := range rows {
		out = append(out, toDepartmentDomain(m))
	}

	log.Info("departments listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return out, nil
}

func (r *DepartmentRepo) Update(ctx context.Context, d *domain.Department) (*domain.Department, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.department", "op", "Update")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Model(&model.Department{}).
		Where("id = ?", d.ID).
		Updates(map[string]any{
			"name":        d.Name,
			"description": d.Description,
		})

	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("update failed",
			"id", d.ID,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil, mapped
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, d.ID)
}

func (r *DepartmentRepo) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.department", "op", "Delete")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Delete(&model.Department{}, id)
	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("delete failed",
			"id", id,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return mapped
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	log.Info("department deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return nil
}
