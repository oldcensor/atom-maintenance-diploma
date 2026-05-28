package dto

import "time"

type CreateMaintenanceScheduleRequest struct {
	EquipmentID   int64     `json:"equipment_id" validate:"required,gt=0"`
	ScheduledAt   time.Time `json:"scheduled_at" validate:"required"`
	Description   string    `json:"description"`
	AssignedTo    *int64    `json:"assigned_to"`
	Status        string    `json:"status" validate:"omitempty,oneof=scheduled in_progress completed cancelled"`
	IntervalUnit  *string   `json:"interval_unit" validate:"omitempty,oneof=days operating_hours cycles"`
	IntervalValue *int      `json:"interval_value" validate:"omitempty,gt=0"`
}

type UpdateMaintenanceScheduleRequest struct {
	EquipmentID   int64     `json:"equipment_id" validate:"required,gt=0"`
	ScheduledAt   time.Time `json:"scheduled_at" validate:"required"`
	Description   string    `json:"description"`
	AssignedTo    *int64    `json:"assigned_to"`
	Status        string    `json:"status" validate:"required,oneof=scheduled in_progress completed cancelled"`
	IntervalUnit  *string   `json:"interval_unit" validate:"omitempty,oneof=days operating_hours cycles"`
	IntervalValue *int      `json:"interval_value" validate:"omitempty,gt=0"`
}

type MaintenanceScheduleResponse struct {
	ID             int64      `json:"id"`
	EquipmentID    int64      `json:"equipment_id"`
	ScheduledAt    time.Time  `json:"scheduled_at"`
	Description    string     `json:"description,omitempty"`
	AssignedTo     *int64     `json:"assigned_to,omitempty"`
	Status         string     `json:"status"`
	IntervalUnit   *string    `json:"interval_unit,omitempty"`
	IntervalValue  *int       `json:"interval_value,omitempty"`
	LastMeterValue *float64   `json:"last_meter_value,omitempty"`
	NextDueAt      *time.Time `json:"next_due_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
