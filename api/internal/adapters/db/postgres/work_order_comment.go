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

func toCommentDomain(m model.WorkOrderComment) domain.WorkOrderComment {
	return domain.WorkOrderComment{
		ID:          m.ID,
		WorkOrderID: m.WorkOrderID,
		AuthorID:    m.AuthorID,
		Text:        m.Text,
		CreatedAt:   m.CreatedAt,
	}
}

type WorkOrderCommentRepo struct {
	db      *gorm.DB
	log     *slog.Logger
	queryTO time.Duration
}

func NewWorkOrderCommentRepo(db *gorm.DB, log *slog.Logger, queryTO time.Duration) *WorkOrderCommentRepo {
	return &WorkOrderCommentRepo{db: db, log: log, queryTO: queryTO}
}

func (r *WorkOrderCommentRepo) Create(ctx context.Context, c *domain.WorkOrderComment) (*domain.WorkOrderComment, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.wo_comment", "op", "Create")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	m := model.WorkOrderComment{
		WorkOrderID: c.WorkOrderID,
		AuthorID:    c.AuthorID,
		Text:        c.Text,
	}
	if err := db.Create(&m).Error; err != nil {
		mapped := pkg.MapDB(err)
		log.Error("create failed", "err", err, "duration_ms", time.Since(start).Milliseconds())
		return nil, mapped
	}

	out := toCommentDomain(m)
	log.Info("comment created", "id", m.ID, "duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *WorkOrderCommentRepo) ListByWorkOrderID(ctx context.Context, workOrderID int64) ([]domain.WorkOrderComment, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.wo_comment", "op", "ListByWorkOrderID")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.WorkOrderComment
	if err := db.Where("work_order_id = ?", workOrderID).Order("created_at asc").Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.WorkOrderComment, 0, len(rows))
	for _, m := range rows {
		out = append(out, toCommentDomain(m))
	}

	log.Info("comments listed", "work_order_id", workOrderID, "count", len(out), "duration_ms", time.Since(start).Milliseconds())
	return out, nil
}

func (r *WorkOrderCommentRepo) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.wo_comment", "op", "Delete")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Delete(&model.WorkOrderComment{}, id)
	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("delete failed", "id", id, "err", res.Error, "duration_ms", time.Since(start).Milliseconds())
		return mapped
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	log.Info("comment deleted", "id", id, "duration_ms", time.Since(start).Milliseconds())
	return nil
}
