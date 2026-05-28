package app

import (
	"context"
	"log/slog"

	"atom-maintenance/internal/domain"
	"atom-maintenance/platform/logger"
)

type MaintenanceScheduleApp struct {
	repo   domain.MaintenanceScheduleRepository
	eqRepo domain.EquipmentRepository
	log    *slog.Logger
}

func NewMaintenanceScheduleApp(repo domain.MaintenanceScheduleRepository, eqRepo domain.EquipmentRepository, log *slog.Logger) *MaintenanceScheduleApp {
	return &MaintenanceScheduleApp{repo: repo, eqRepo: eqRepo, log: log}
}

func (a *MaintenanceScheduleApp) Create(ctx context.Context, s *domain.MaintenanceSchedule) (*domain.MaintenanceSchedule, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.schedule", "op", "Create")

	if eq, err := a.eqRepo.GetByID(ctx, s.EquipmentID); err == nil && eq.Status == domain.StatusDecommissioned {
		log.Warn("equipment decommissioned", "equipment_id", s.EquipmentID)
		return nil, domain.ErrBadRequest
	}

	out, err := a.repo.Create(ctx, s)
	if err != nil {
		log.Error("create failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *MaintenanceScheduleApp) GetByID(ctx context.Context, id int64) (*domain.MaintenanceSchedule, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.schedule", "op", "GetByID")

	out, err := a.repo.GetByID(ctx, id)
	if err != nil {
		log.Warn("get failed", "id", id, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *MaintenanceScheduleApp) List(ctx context.Context) ([]domain.MaintenanceSchedule, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.schedule", "op", "List")

	out, err := a.repo.List(ctx)
	if err != nil {
		log.Error("list failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *MaintenanceScheduleApp) Update(ctx context.Context, s *domain.MaintenanceSchedule) (*domain.MaintenanceSchedule, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.schedule", "op", "Update")

	out, err := a.repo.Update(ctx, s)
	if err != nil {
		log.Error("update failed", "id", s.ID, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *MaintenanceScheduleApp) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, a.log).With("module", "app.schedule", "op", "Delete")

	if err := a.repo.Delete(ctx, id); err != nil {
		log.Error("delete failed", "id", id, "err", err)
		return err
	}
	return nil
}
