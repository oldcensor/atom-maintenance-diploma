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

func toStatusLogDomain(m model.WorkOrderStatusLog) domain.WorkOrderStatusLog {
	return domain.WorkOrderStatusLog{
		ID:          m.ID,
		WorkOrderID: m.WorkOrderID,
		FromStatus:  domain.WorkOrderStatus(m.FromStatus),
		ToStatus:    domain.WorkOrderStatus(m.ToStatus),
		ChangedBy:   m.ChangedBy,
		Comment:     m.Comment,
		CreatedAt:   m.CreatedAt,
	}
}

type WorkOrderStatusLogRepo struct {
	db      *gorm.DB
	log     *slog.Logger
	queryTO time.Duration
}

func NewWorkOrderStatusLogRepo(db *gorm.DB, log *slog.Logger, queryTO time.Duration) *WorkOrderStatusLogRepo {
	return &WorkOrderStatusLogRepo{db: db, log: log, queryTO: queryTO}
}

func (r *WorkOrderStatusLogRepo) Create(ctx context.Context, l *domain.WorkOrderStatusLog) (*domain.WorkOrderStatusLog, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.wo_status_log", "op", "Create")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	m := model.WorkOrderStatusLog{
		WorkOrderID: l.WorkOrderID,
		FromStatus:  string(l.FromStatus),
		ToStatus:    string(l.ToStatus),
		ChangedBy:   l.ChangedBy,
		Comment:     l.Comment,
	}
	if err := db.Create(&m).Error; err != nil {
		mapped := pkg.MapDB(err)
		log.Error("create failed", "err", err, "duration_ms", time.Since(start).Milliseconds())
		return nil, mapped
	}

	out := toStatusLogDomain(m)
	log.Info("status log created", "id", m.ID, "duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *WorkOrderStatusLogRepo) ListByWorkOrderID(ctx context.Context, workOrderID int64) ([]domain.WorkOrderStatusLog, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.wo_status_log", "op", "ListByWorkOrderID")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.WorkOrderStatusLog
	if err := db.Where("work_order_id = ?", workOrderID).Order("created_at asc").Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.WorkOrderStatusLog, 0, len(rows))
	for _, m := range rows {
		out = append(out, toStatusLogDomain(m))
	}

	log.Info("status logs listed", "work_order_id", workOrderID, "count", len(out), "duration_ms", time.Since(start).Milliseconds())
	return out, nil
}
