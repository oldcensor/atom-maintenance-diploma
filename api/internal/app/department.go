package app

import (
	"context"
	"log/slog"

	"atom-maintenance/internal/domain"
	"atom-maintenance/platform/logger"
)

type DepartmentApp struct {
	repo domain.DepartmentRepository
	log  *slog.Logger
}

func NewDepartmentApp(repo domain.DepartmentRepository, log *slog.Logger) *DepartmentApp {
	return &DepartmentApp{repo: repo, log: log}
}

func (a *DepartmentApp) Create(ctx context.Context, d *domain.Department) (*domain.Department, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.department", "op", "Create")
	out, err := a.repo.Create(ctx, d)
	if err != nil {
		log.Error("create failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *DepartmentApp) GetByID(ctx context.Context, id int64) (*domain.Department, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.department", "op", "GetByID")
	out, err := a.repo.GetByID(ctx, id)
	if err != nil {
		log.Warn("get failed", "id", id, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *DepartmentApp) List(ctx context.Context) ([]domain.Department, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.department", "op", "List")
	out, err := a.repo.List(ctx)
	if err != nil {
		log.Error("list failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *DepartmentApp) Update(ctx context.Context, d *domain.Department) (*domain.Department, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.department", "op", "Update")
	out, err := a.repo.Update(ctx, d)
	if err != nil {
		log.Error("update failed", "id", d.ID, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *DepartmentApp) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, a.log).With("module", "app.department", "op", "Delete")
	if err := a.repo.Delete(ctx, id); err != nil {
		log.Error("delete failed", "id", id, "err", err)
		return err
	}
	return nil
}
