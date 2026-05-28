package app

import (
	"context"
	"log/slog"

	"atom-maintenance/internal/domain"
	"atom-maintenance/platform/logger"
)

type WorkOrderCommentApp struct {
	repo domain.WorkOrderCommentRepository
	log  *slog.Logger
}

func NewWorkOrderCommentApp(repo domain.WorkOrderCommentRepository, log *slog.Logger) *WorkOrderCommentApp {
	return &WorkOrderCommentApp{repo: repo, log: log}
}

func (a *WorkOrderCommentApp) Create(ctx context.Context, c *domain.WorkOrderComment) (*domain.WorkOrderComment, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.wo_comment", "op", "Create")

	out, err := a.repo.Create(ctx, c)
	if err != nil {
		log.Error("create failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *WorkOrderCommentApp) ListByWorkOrderID(ctx context.Context, workOrderID int64) ([]domain.WorkOrderComment, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.wo_comment", "op", "ListByWorkOrderID")

	out, err := a.repo.ListByWorkOrderID(ctx, workOrderID)
	if err != nil {
		log.Error("list failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *WorkOrderCommentApp) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, a.log).With("module", "app.wo_comment", "op", "Delete")

	if err := a.repo.Delete(ctx, id); err != nil {
		log.Error("delete failed", "id", id, "err", err)
		return err
	}
	return nil
}
