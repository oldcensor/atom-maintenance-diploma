package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"atom-maintenance/internal/domain"
	"atom-maintenance/internal/ports"
	"atom-maintenance/pkg"
	"atom-maintenance/platform/logger"
)

const revokedEmpPrefix = "revoked:emp:"

type EmployeeApp struct {
	repo      domain.EmployeeRepository
	cache     ports.Cache
	accessTTL time.Duration
	log       *slog.Logger
}

func NewEmployeeApp(repo domain.EmployeeRepository, cache ports.Cache, accessTTL time.Duration, log *slog.Logger) *EmployeeApp {
	return &EmployeeApp{
		repo:      repo,
		cache:     cache,
		accessTTL: accessTTL,
		log:       log,
	}
}

func (a *EmployeeApp) Create(ctx context.Context, e *domain.Employee, plainPassword string) (*domain.Employee, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.employee", "op", "Create")

	hash, err := pkg.HashPassword(plainPassword)
	if err != nil {
		log.Error("hash password", "err", err)
		return nil, domain.ErrInternal
	}
	e.PasswordHash = hash

	out, err := a.repo.Create(ctx, e)
	if err != nil {
		log.Error("create failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EmployeeApp) GetByID(ctx context.Context, id int64) (*domain.Employee, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.employee", "op", "GetByID")
	out, err := a.repo.GetByID(ctx, id)
	if err != nil {
		log.Warn("get failed", "id", id, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EmployeeApp) List(ctx context.Context) ([]domain.Employee, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.employee", "op", "List")
	out, err := a.repo.List(ctx)
	if err != nil {
		log.Error("list failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EmployeeApp) Update(ctx context.Context, e *domain.Employee) (*domain.Employee, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.employee", "op", "Update")
	out, err := a.repo.Update(ctx, e)
	if err != nil {
		log.Error("update failed", "id", e.ID, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EmployeeApp) SoftDelete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, a.log).With("module", "app.employee", "op", "SoftDelete")
	if err := a.repo.SoftDelete(ctx, id); err != nil {
		log.Error("soft-delete failed", "id", id, "err", err)
		return err
	}
	key := fmt.Sprintf("%s%d", revokedEmpPrefix, id)
	if err := a.cache.Set(ctx, key, "revoked", a.accessTTL); err != nil {
		log.Error("revoke employee tokens", "id", id, "err", err)
	}
	return nil
}
