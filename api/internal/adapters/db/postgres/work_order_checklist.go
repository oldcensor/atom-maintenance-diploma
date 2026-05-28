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

func toChecklistDomain(m model.WorkOrderChecklistItem) domain.ChecklistItem {
	return domain.ChecklistItem{
		ID:          m.ID,
		WorkOrderID: m.WorkOrderID,
		Text:        m.Text,
		Checked:     m.Checked,
		CheckedBy:   m.CheckedBy,
		CheckedAt:   m.CheckedAt,
		SortOrder:   m.SortOrder,
		CreatedAt:   m.CreatedAt,
	}
}

type WorkOrderChecklistRepo struct {
	db      *gorm.DB
	log     *slog.Logger
	queryTO time.Duration
}

func NewWorkOrderChecklistRepo(db *gorm.DB, log *slog.Logger, queryTO time.Duration) *WorkOrderChecklistRepo {
	return &WorkOrderChecklistRepo{db: db, log: log, queryTO: queryTO}
}

func (r *WorkOrderChecklistRepo) Create(ctx context.Context, item *domain.ChecklistItem) (*domain.ChecklistItem, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.wo_checklist", "op", "Create")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	m := model.WorkOrderChecklistItem{
		WorkOrderID: item.WorkOrderID,
		Text:        item.Text,
		SortOrder:   item.SortOrder,
	}
	if err := db.Create(&m).Error; err != nil {
		mapped := pkg.MapDB(err)
		log.Error("create failed", "err", err, "duration_ms", time.Since(start).Milliseconds())
		return nil, mapped
	}

	out := toChecklistDomain(m)
	log.Info("checklist item created", "id", m.ID, "duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *WorkOrderChecklistRepo) ListByWorkOrderID(ctx context.Context, workOrderID int64) ([]domain.ChecklistItem, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.wo_checklist", "op", "ListByWorkOrderID")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.WorkOrderChecklistItem
	if err := db.Where("work_order_id = ?", workOrderID).Order("sort_order asc, id asc").Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.ChecklistItem, 0, len(rows))
	for _, m := range rows {
		out = append(out, toChecklistDomain(m))
	}

	log.Info("checklist listed", "work_order_id", workOrderID, "count", len(out), "duration_ms", time.Since(start).Milliseconds())
	return out, nil
}

func (r *WorkOrderChecklistRepo) Toggle(ctx context.Context, id int64, checked bool, checkedBy *int64) (*domain.ChecklistItem, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.wo_checklist", "op", "Toggle")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	updates := map[string]any{
		"checked": checked,
	}
	if checked {
		now := time.Now()
		updates["checked_by"] = checkedBy
		updates["checked_at"] = &now
	} else {
		updates["checked_by"] = nil
		updates["checked_at"] = nil
	}

	res := db.Model(&model.WorkOrderChecklistItem{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("toggle failed", "id", id, "err", res.Error, "duration_ms", time.Since(start).Milliseconds())
		return nil, mapped
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	var m model.WorkOrderChecklistItem
	if err := db.First(&m, id).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := toChecklistDomain(m)
	log.Info("checklist item toggled", "id", id, "checked", checked, "duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *WorkOrderChecklistRepo) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.wo_checklist", "op", "Delete")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Delete(&model.WorkOrderChecklistItem{}, id)
	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("delete failed", "id", id, "err", res.Error, "duration_ms", time.Since(start).Milliseconds())
		return mapped
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	log.Info("checklist item deleted", "id", id, "duration_ms", time.Since(start).Milliseconds())
	return nil
}
