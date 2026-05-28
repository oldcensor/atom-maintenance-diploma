package domain

import (
	"context"
	"time"
)

type ScheduleStatus string

const (
	ScheduleStatusScheduled  ScheduleStatus = "scheduled"
	ScheduleStatusInProgress ScheduleStatus = "in_progress"
	ScheduleStatusCompleted  ScheduleStatus = "completed"
	ScheduleStatusCancelled  ScheduleStatus = "cancelled"
)

type MaintenanceSchedule struct {
	ID              int64
	EquipmentID     int64
	ScheduledAt     time.Time
	Description     string
	AssignedTo      *int64
	Status          ScheduleStatus
	IntervalUnit    *string
	IntervalValue   *int
	LastMeterValue  *float64
	NextDueAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type MaintenanceScheduleRepository interface {
	Create(ctx context.Context, s *MaintenanceSchedule) (*MaintenanceSchedule, error)
	GetByID(ctx context.Context, id int64) (*MaintenanceSchedule, error)
	List(ctx context.Context) ([]MaintenanceSchedule, error)
	ListActive(ctx context.Context) ([]MaintenanceSchedule, error)
	Update(ctx context.Context, s *MaintenanceSchedule) (*MaintenanceSchedule, error)
	UpdateMeterFields(ctx context.Context, id int64, lastMeterValue *float64, nextDueAt *time.Time) error
	Delete(ctx context.Context, id int64) error
}
