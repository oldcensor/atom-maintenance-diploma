package app

import (
	"context"
	"log/slog"

	"atom-maintenance/internal/domain"
	"atom-maintenance/platform/logger"
)

type WorkOrderChecklistApp struct {
	repo domain.ChecklistItemRepository
	log  *slog.Logger
}

func NewWorkOrderChecklistApp(repo domain.ChecklistItemRepository, log *slog.Logger) *WorkOrderChecklistApp {
	return &WorkOrderChecklistApp{repo: repo, log: log}
}

func (a *WorkOrderChecklistApp) Create(ctx context.Context, item *domain.ChecklistItem) (*domain.ChecklistItem, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.wo_checklist", "op", "Create")

	out, err := a.repo.Create(ctx, item)
	if err != nil {
		log.Error("create failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *WorkOrderChecklistApp) ListByWorkOrderID(ctx context.Context, workOrderID int64) ([]domain.ChecklistItem, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.wo_checklist", "op", "ListByWorkOrderID")

	out, err := a.repo.ListByWorkOrderID(ctx, workOrderID)
	if err != nil {
		log.Error("list failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *WorkOrderChecklistApp) Toggle(ctx context.Context, id int64, checked bool, checkedBy *int64) (*domain.ChecklistItem, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.wo_checklist", "op", "Toggle")

	out, err := a.repo.Toggle(ctx, id, checked, checkedBy)
	if err != nil {
		log.Error("toggle failed", "id", id, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *WorkOrderChecklistApp) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, a.log).With("module", "app.wo_checklist", "op", "Delete")

	if err := a.repo.Delete(ctx, id); err != nil {
		log.Error("delete failed", "id", id, "err", err)
		return err
	}
	return nil
}
