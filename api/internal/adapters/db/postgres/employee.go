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

func toEmployeeDomain(m model.Employee) domain.Employee {
	return domain.Employee{
		ID:             m.ID,
		Email:          m.Email,
		PasswordHash:   m.PasswordHash,
		FullName:       m.FullName,
		Role:           domain.EmployeeRole(m.Role),
		DepartmentID:   m.DepartmentID,
		FailedAttempts: m.FailedAttempts,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      m.DeletedAt,
	}
}

type EmployeeRepo struct {
	db      *gorm.DB
	log     *slog.Logger
	queryTO time.Duration
}

func NewEmployeeRepo(db *gorm.DB, log *slog.Logger, queryTO time.Duration) *EmployeeRepo {
	return &EmployeeRepo{db: db, log: log, queryTO: queryTO}
}

func (r *EmployeeRepo) Create(ctx context.Context, e *domain.Employee) (*domain.Employee, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.employee", "op", "Create")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	m := model.Employee{
		Email:        e.Email,
		PasswordHash: e.PasswordHash,
		FullName:     e.FullName,
		Role:         string(e.Role),
		DepartmentID: e.DepartmentID,
	}
	if err := db.Create(&m).Error; err != nil {
		mapped := pkg.MapDB(err)
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil, mapped
	}

	out := toEmployeeDomain(m)
	log.Info("employee created",
		"id", m.ID,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return &out, nil
}

func (r *EmployeeRepo) GetByID(ctx context.Context, id int64) (*domain.Employee, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.employee", "op", "GetByID")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var m model.Employee
	if err := db.Where("deleted_at IS NULL").First(&m, id).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := toEmployeeDomain(m)
	log.Info("employee fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *EmployeeRepo) GetByEmail(ctx context.Context, email string) (*domain.Employee, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.employee", "op", "GetByEmail")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var m model.Employee
	if err := db.Where("email = ? AND deleted_at IS NULL", email).First(&m).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := toEmployeeDomain(m)
	log.Info("employee fetched by email",
		"id", m.ID,
		"duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *EmployeeRepo) List(ctx context.Context) ([]domain.Employee, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.employee", "op", "List")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.Employee
	if err := db.Where("deleted_at IS NULL").Order("id asc").Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.Employee, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEmployeeDomain(m))
	}

	log.Info("employees listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())
	return out, nil
}

func (r *EmployeeRepo) Update(ctx context.Context, e *domain.Employee) (*domain.Employee, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.employee", "op", "Update")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Model(&model.Employee{}).
		Where("id = ? AND deleted_at IS NULL", e.ID).
		Updates(map[string]any{
			"full_name":     e.FullName,
			"role":          string(e.Role),
			"department_id": e.DepartmentID,
			"updated_at":    time.Now(),
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

func (r *EmployeeRepo) SoftDelete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.employee", "op", "SoftDelete")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	now := time.Now()
	res := db.Model(&model.Employee{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"deleted_at": now, "updated_at": now})

	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("soft-delete failed",
			"id", id,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds())
		return mapped
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	log.Info("employee soft-deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (r *EmployeeRepo) IncrFailedAttempts(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.employee", "op", "IncrFailedAttempts")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	if err := db.Model(&model.Employee{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"failed_attempts": gorm.Expr("failed_attempts + 1"),
			"updated_at":      time.Now(),
		}).Error; err != nil {
		log.Error("incr failed_attempts error",
			"id", id,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		return pkg.MapDB(err)
	}

	log.Info("failed_attempts incremented",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (r *EmployeeRepo) ResetFailedAttempts(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.employee", "op", "ResetFailedAttempts")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	if err := db.Model(&model.Employee{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"failed_attempts": 0,
			"updated_at":      time.Now(),
		}).Error; err != nil {
		log.Error("reset failed_attempts error",
			"id", id,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		return pkg.MapDB(err)
	}

	log.Info("failed_attempts reset",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}
