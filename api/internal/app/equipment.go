package app

import (
	"context"
	"log/slog"

	"atom-maintenance/internal/domain"
	"atom-maintenance/platform/logger"
)

type EquipmentApp struct {
	repo domain.EquipmentRepository
	log  *slog.Logger
}

func NewEquipmentApp(repo domain.EquipmentRepository, log *slog.Logger) *EquipmentApp {
	return &EquipmentApp{repo: repo, log: log}
}

func (a *EquipmentApp) Create(ctx context.Context, e *domain.Equipment) (*domain.Equipment, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.equipment", "op", "Create")

	out, err := a.repo.Create(ctx, e)
	if err != nil {
		log.Error("create failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EquipmentApp) GetByID(ctx context.Context, id int64) (*domain.Equipment, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.equipment", "op", "GetByID")

	out, err := a.repo.GetByID(ctx, id)
	if err != nil {
		log.Warn("get failed", "id", id, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EquipmentApp) List(ctx context.Context) ([]domain.Equipment, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.equipment", "op", "List")

	out, err := a.repo.List(ctx)
	if err != nil {
		log.Error("list failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EquipmentApp) Update(ctx context.Context, e *domain.Equipment) (*domain.Equipment, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.equipment", "op", "Update")

	out, err := a.repo.Update(ctx, e)
	if err != nil {
		log.Error("update failed", "id", e.ID, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EquipmentApp) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, a.log).With("module", "app.equipment", "op", "Delete")

	if err := a.repo.Delete(ctx, id); err != nil {
		log.Error("delete failed", "id", id, "err", err)
		return err
	}
	return nil
}
