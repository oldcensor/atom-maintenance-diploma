package app

import (
	"context"
	"log/slog"

	"atom-maintenance/internal/domain"
	"atom-maintenance/platform/logger"
)

type InspectionReportApp struct {
	repo   domain.InspectionReportRepository
	woRepo domain.WorkOrderRepository
	log    *slog.Logger
}

func NewInspectionReportApp(repo domain.InspectionReportRepository, woRepo domain.WorkOrderRepository, log *slog.Logger) *InspectionReportApp {
	return &InspectionReportApp{
		repo:   repo,
		woRepo: woRepo,
		log:    log,
	}
}

func (a *InspectionReportApp) Create(ctx context.Context, rep *domain.InspectionReport) (*domain.InspectionReport, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.inspection_report", "op", "Create")

	wo, err := a.woRepo.GetByID(ctx, rep.WorkOrderID)
	if err != nil {
		log.Warn("work order not found", "work_order_id", rep.WorkOrderID, "err", err)
		return nil, err
	}
	if wo.Status != domain.WorkOrderStatusOpen && wo.Status != domain.WorkOrderStatusInProgress {
		log.Warn("work order not in actionable state", "work_order_id", wo.ID, "status", wo.Status)
		return nil, domain.ErrBadRequest
	}

	out, err := a.repo.Create(ctx, rep)
	if err != nil {
		log.Error("create failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *InspectionReportApp) GetByID(ctx context.Context, id int64) (*domain.InspectionReport, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.inspection_report", "op", "GetByID")

	out, err := a.repo.GetByID(ctx, id)
	if err != nil {
		log.Warn("get failed", "id", id, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *InspectionReportApp) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, a.log).With("module", "app.inspection_report", "op", "Delete")

	if err := a.repo.Delete(ctx, id); err != nil {
		log.Warn("delete failed", "id", id, "err", err)
		return err
	}
	return nil
}

func (a *InspectionReportApp) List(ctx context.Context) ([]domain.InspectionReport, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.inspection_report", "op", "List")

	out, err := a.repo.List(ctx)
	if err != nil {
		log.Error("list failed", "err", err)
		return nil, err
	}
	return out, nil
}
