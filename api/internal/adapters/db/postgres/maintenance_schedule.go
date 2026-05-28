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

func toScheduleDomain(m model.MaintenanceSchedule) domain.MaintenanceSchedule {
	return domain.MaintenanceSchedule{
		ID:             m.ID,
		EquipmentID:    m.EquipmentID,
		ScheduledAt:    m.ScheduledAt,
		Description:    m.Description,
		AssignedTo:     m.AssignedTo,
		Status:         domain.ScheduleStatus(m.Status),
		IntervalUnit:   m.IntervalUnit,
		IntervalValue:  m.IntervalValue,
		LastMeterValue: m.LastMeterValue,
		NextDueAt:      m.NextDueAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

type MaintenanceScheduleRepo struct {
	db      *gorm.DB
	log     *slog.Logger
	queryTO time.Duration
}

func NewMaintenanceScheduleRepo(db *gorm.DB, log *slog.Logger, queryTO time.Duration) *MaintenanceScheduleRepo {
	return &MaintenanceScheduleRepo{db: db, log: log, queryTO: queryTO}
}

func (r *MaintenanceScheduleRepo) Create(ctx context.Context, s *domain.MaintenanceSchedule) (*domain.MaintenanceSchedule, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.schedule", "op", "Create")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	m := model.MaintenanceSchedule{
		EquipmentID:    s.EquipmentID,
		ScheduledAt:    s.ScheduledAt,
		Description:    s.Description,
		AssignedTo:     s.AssignedTo,
		Status:         string(s.Status),
		IntervalUnit:   s.IntervalUnit,
		IntervalValue:  s.IntervalValue,
		LastMeterValue: s.LastMeterValue,
		NextDueAt:      s.NextDueAt,
	}
	if err := db.Create(&m).Error; err != nil {
		mapped := pkg.MapDB(err)
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil, mapped
	}

	out := toScheduleDomain(m)
	log.Info("schedule created",
		"id", m.ID,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return &out, nil
}

func (r *MaintenanceScheduleRepo) GetByID(ctx context.Context, id int64) (*domain.MaintenanceSchedule, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.schedule", "op", "GetByID")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var m model.MaintenanceSchedule
	if err := db.First(&m, id).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := toScheduleDomain(m)
	log.Info("schedule fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *MaintenanceScheduleRepo) List(ctx context.Context) ([]domain.MaintenanceSchedule, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.schedule", "op", "List")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.MaintenanceSchedule
	if err := db.Order("scheduled_at asc").Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.MaintenanceSchedule, 0, len(rows))
	for _, m := range rows {
		out = append(out, toScheduleDomain(m))
	}

	log.Info("schedules listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())
	return out, nil
}

// ListActive returns scheduled-status schedules with interval_unit set — used by the scheduler.
func (r *MaintenanceScheduleRepo) ListActive(ctx context.Context) ([]domain.MaintenanceSchedule, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.schedule", "op", "ListActive")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.MaintenanceSchedule
	if err := db.Where("status = ? AND interval_unit IS NOT NULL", string(domain.ScheduleStatusScheduled)).
		Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.MaintenanceSchedule, 0, len(rows))
	for _, m := range rows {
		out = append(out, toScheduleDomain(m))
	}

	log.Info("active schedules listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())
	return out, nil
}

func (r *MaintenanceScheduleRepo) Update(ctx context.Context, s *domain.MaintenanceSchedule) (*domain.MaintenanceSchedule, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.schedule", "op", "Update")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Model(&model.MaintenanceSchedule{}).
		Where("id = ?", s.ID).
		Updates(map[string]any{
			"equipment_id":     s.EquipmentID,
			"scheduled_at":     s.ScheduledAt,
			"description":      s.Description,
			"assigned_to":      s.AssignedTo,
			"status":           string(s.Status),
			"interval_unit":    s.IntervalUnit,
			"interval_value":   s.IntervalValue,
			"last_meter_value": s.LastMeterValue,
			"next_due_at":      s.NextDueAt,
			"updated_at":       time.Now(),
		})

	if res.Error != nil {
		mapped := pkg.MapDB(res.Error)
		log.Error("update failed",
			"id", s.ID,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, mapped
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, s.ID)
}

// UpdateMeterFields updates only last_meter_value and next_due_at — called inside a transaction by the scheduler.
func (r *MaintenanceScheduleRepo) UpdateMeterFields(ctx context.Context, id int64, lastMeterValue *float64, nextDueAt *time.Time) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.schedule", "op", "UpdateMeterFields")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Model(&model.MaintenanceSchedule{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_meter_value": lastMeterValue,
			"next_due_at":      nextDueAt,
			"updated_at":       time.Now(),
		})

	if res.Error != nil {
		log.Error("update meter fields failed",
			"id", id,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds())
		return pkg.MapDB(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	log.Info("meter fields updated",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (r *MaintenanceScheduleRepo) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.schedule", "op", "Delete")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Delete(&model.MaintenanceSchedule{}, id)
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

	log.Info("schedule deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}
