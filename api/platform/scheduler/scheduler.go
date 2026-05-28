package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"atom-maintenance/internal/adapters/simulator"
	"atom-maintenance/internal/domain"
)

type Scheduler struct {
	interval  time.Duration
	simClient *simulator.Client
	schedRepo domain.MaintenanceScheduleRepository
	woRepo    domain.WorkOrderRepository
	txm       domain.TxManager
	log       *slog.Logger
	stop      chan struct{}
	stopOnce  sync.Once
}

func New(
	interval time.Duration,
	simClient *simulator.Client,
	schedRepo domain.MaintenanceScheduleRepository,
	woRepo domain.WorkOrderRepository,
	txm domain.TxManager,
	log *slog.Logger,
) *Scheduler {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &Scheduler{
		interval:  interval,
		simClient: simClient,
		schedRepo: schedRepo,
		woRepo:    woRepo,
		txm:       txm,
		log:       log,
		stop:      make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.log.Info("scheduler started", "interval", s.interval)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-s.stop:
			s.log.Info("scheduler stopped")
			return
		case <-ctx.Done():
			s.log.Info("scheduler context done")
			return
		}
	}
}

func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stop)
	})
}
func (s *Scheduler) Tick(ctx context.Context) {
	s.tick(ctx)
}

func (s *Scheduler) tick(ctx context.Context) {
	log := s.log.With("module", "scheduler", "op", "tick")

	schedules, err := s.schedRepo.ListActive(ctx)
	if err != nil {
		log.Error("list active schedules failed", "err", err)
		return
	}

	for _, sched := range schedules {
		if err := s.processSchedule(ctx, sched); err != nil {
			log.Error("process schedule failed", "schedule_id", sched.ID, "err", err)
		}
	}
}

func (s *Scheduler) processSchedule(ctx context.Context, sched domain.MaintenanceSchedule) error {
	log := s.log.With("module", "scheduler", "op", "processSchedule", "schedule_id", sched.ID)

	if sched.IntervalUnit == nil || sched.IntervalValue == nil {
		return nil
	}

	due, newMeterValue, err := s.isDue(ctx, sched)
	if err != nil {
		return fmt.Errorf("check due: %w", err)
	}
	if !due {
		return nil
	}

	exists, err := s.woRepo.ExistsByScheduleID(ctx, sched.ID, []domain.WorkOrderStatus{
		domain.WorkOrderStatusOpen,
		domain.WorkOrderStatusInProgress,
	})
	if err != nil {
		return fmt.Errorf("check existing work order: %w", err)
	}
	if exists {
		log.Info("work order already exists, skipping")
		return nil
	}

	return s.txm.WithinTx(ctx, func(txCtx context.Context) error {
		wo := &domain.WorkOrder{
			ScheduleID:  &sched.ID,
			EquipmentID: sched.EquipmentID,
			Title:       fmt.Sprintf("Плановое ТО: %s", sched.Description),
			Status:      domain.WorkOrderStatusOpen,
			WorkType:    domain.WorkOrderTypePlanned,
		}
		created, err := s.woRepo.Create(txCtx, wo)
		if err != nil {
			return fmt.Errorf("create work order: %w", err)
		}
		log.Info("work order created", "work_order_id", created.ID)

		nextDueAt := calcNextDueAt(*sched.IntervalUnit, *sched.IntervalValue, newMeterValue)
		if err := s.schedRepo.UpdateMeterFields(txCtx, sched.ID, &newMeterValue, nextDueAt); err != nil {
			return fmt.Errorf("update meter fields: %w", err)
		}
		return nil
	})
}

func (s *Scheduler) isDue(ctx context.Context, sched domain.MaintenanceSchedule) (bool, float64, error) {
	unit := *sched.IntervalUnit
	interval := float64(*sched.IntervalValue)

	switch unit {
	case "days":
		now := time.Now()
		if sched.NextDueAt == nil {
			return !sched.ScheduledAt.After(now), float64(daysSinceEpoch(now)), nil
		}
		due := !now.Before(*sched.NextDueAt)
		return due, float64(daysSinceEpoch(now)), nil

	case "operating_hours", "cycles":
		telemetry, err := s.simClient.GetByEquipmentID(ctx, sched.EquipmentID)
		if err != nil {
			return false, 0, fmt.Errorf("get telemetry: %w", err)
		}
		if telemetry == nil {
			s.log.Warn("no telemetry for equipment", "equipment_id", sched.EquipmentID)
			return false, 0, nil
		}
		current := telemetry.CurrentValue
		last := float64(0)
		if sched.LastMeterValue != nil {
			last = *sched.LastMeterValue
		}
		return current >= last+interval, current, nil

	default:
		return false, 0, fmt.Errorf("unknown interval_unit: %s", unit)
	}
}

func calcNextDueAt(unit string, intervalValue int, currentValue float64) *time.Time {
	switch unit {
	case "days":
		t := time.Now().AddDate(0, 0, intervalValue)
		return &t
	case "operating_hours", "cycles":
		t := time.Now().AddDate(0, 0, intervalValue)
		return &t
	}
	return nil
}

func daysSinceEpoch(t time.Time) int {
	epoch := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return int(t.Sub(epoch).Hours() / 24)
}
