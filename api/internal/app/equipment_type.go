package app

import (
	"context"
	"log/slog"

	"atom-maintenance/internal/domain"
	"atom-maintenance/platform/logger"
)

type EquipmentTypeApp struct {
	repo domain.EquipmentTypeRepository
	log  *slog.Logger
}

func NewEquipmentTypeApp(repo domain.EquipmentTypeRepository, log *slog.Logger) *EquipmentTypeApp {
	return &EquipmentTypeApp{repo: repo, log: log}
}

func (a *EquipmentTypeApp) Create(ctx context.Context, et *domain.EquipmentType) (*domain.EquipmentType, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.equipment_type", "op", "Create")
	out, err := a.repo.Create(ctx, et)
	if err != nil {
		log.Error("create failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EquipmentTypeApp) GetByID(ctx context.Context, id int64) (*domain.EquipmentType, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.equipment_type", "op", "GetByID")
	out, err := a.repo.GetByID(ctx, id)
	if err != nil {
		log.Warn("get failed", "id", id, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EquipmentTypeApp) List(ctx context.Context) ([]domain.EquipmentType, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.equipment_type", "op", "List")
	out, err := a.repo.List(ctx)
	if err != nil {
		log.Error("list failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EquipmentTypeApp) Update(ctx context.Context, et *domain.EquipmentType) (*domain.EquipmentType, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.equipment_type", "op", "Update")
	out, err := a.repo.Update(ctx, et)
	if err != nil {
		log.Error("update failed", "id", et.ID, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *EquipmentTypeApp) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, a.log).With("module", "app.equipment_type", "op", "Delete")
	if err := a.repo.Delete(ctx, id); err != nil {
		log.Error("delete failed", "id", id, "err", err)
		return err
	}
	return nil
}
