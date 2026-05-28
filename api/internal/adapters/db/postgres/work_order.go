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

func toWorkOrderDomain(m model.WorkOrder) domain.WorkOrder {
	return domain.WorkOrder{
		ID:          m.ID,
		ScheduleID:  m.ScheduleID,
		EquipmentID: m.EquipmentID,
		Title:       m.Title,
		Description: m.Description,
		AssignedTo:  m.AssignedTo,
		CreatedBy:   m.CreatedBy,
		Status:      domain.WorkOrderStatus(m.Status),
		WorkType:    domain.WorkOrderType(m.WorkType),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		CompletedAt: m.CompletedAt,
	}
}

type WorkOrderRepo struct {
	db      *gorm.DB
	log     *slog.Logger
	queryTO time.Duration
}

func NewWorkOrderRepo(db *gorm.DB, log *slog.Logger, queryTO time.Duration) *WorkOrderRepo {
	return &WorkOrderRepo{db: db, log: log, queryTO: queryTO}
}

func (r *WorkOrderRepo) Create(ctx context.Context, w *domain.WorkOrder) (*domain.WorkOrder, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.work_order", "op", "Create")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	m := model.WorkOrder{
		ScheduleID:  w.ScheduleID,
		EquipmentID: w.EquipmentID,
		Title:       w.Title,
		Description: w.Description,
		AssignedTo:  w.AssignedTo,
		CreatedBy:   w.CreatedBy,
		Status:      string(w.Status),
		WorkType:    string(w.WorkType),
		CompletedAt: w.CompletedAt,
	}
	if err := db.Create(&m).Error; err != nil {
		mapped := pkg.MapDB(err)
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, mapped
	}

	out := toWorkOrderDomain(m)
	log.Info("work order created",
		"id", m.ID,
		"duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *WorkOrderRepo) GetByID(ctx context.Context, id int64) (*domain.WorkOrder, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.work_order", "op", "GetByID")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var m model.WorkOrder
	if err := db.First(&m, id).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := toWorkOrderDomain(m)
	log.Info("work order fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *WorkOrderRepo) List(ctx context.Context) ([]domain.WorkOrder, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.work_order", "op", "List")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.WorkOrder
	if err := db.Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.WorkOrder, 0, len(rows))
	for _, m := range rows {
		out = append(out, toWorkOrderDomain(m))
	}

	log.Info("work orders listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())
	return out, nil
}

func (r *WorkOrderRepo) Update(ctx context.Context, w *domain.WorkOrder) (*domain.WorkOrder, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.work_order", "op", "Update")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	updates := map[string]any{
		"schedule_id":  w.ScheduleID,
		"equipment_id": w.EquipmentID,
		"title":        w.Title,
		"description":  w.Description,
		"assigned_to":  w.AssignedTo,
		"status":       string(w.Status),
		"work_type":    string(w.WorkType),
		"completed_at": w.CompletedAt,
		"updated_at":   time.Now(),
	}

	res := db.Model(&model.WorkOrder{}).Where("id = ?", w.ID).Updates(updates)
	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("update failed",
			"id", w.ID,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, mapped
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, w.ID)
}

func (r *WorkOrderRepo) ExistsByScheduleID(ctx context.Context, scheduleID int64, statuses []domain.WorkOrderStatus) (bool, error) {
	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	vals := make([]string, len(statuses))
	for i, s := range statuses {
		vals[i] = string(s)
	}

	var count int64
	if err := db.Model(&model.WorkOrder{}).
		Where("schedule_id = ? AND status IN ?", scheduleID, vals).
		Count(&count).Error; err != nil {
		return false, pkg.MapDB(err)
	}
	return count > 0, nil
}

func (r *WorkOrderRepo) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.work_order", "op", "Delete")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Delete(&model.WorkOrder{}, id)
	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("delete failed",
			"id", id,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds())
		return mapped
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	log.Info("work order deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}
