package model

import "time"

type MaintenanceSchedule struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement"`
	EquipmentID    int64      `gorm:"column:equipment_id;not null"`
	ScheduledAt    time.Time  `gorm:"column:scheduled_at;not null"`
	Description    string     `gorm:"column:description"`
	AssignedTo     *int64     `gorm:"column:assigned_to"`
	Status         string     `gorm:"column:status;not null;default:scheduled"`
	IntervalUnit   *string    `gorm:"column:interval_unit"`
	IntervalValue  *int       `gorm:"column:interval_value"`
	LastMeterValue *float64   `gorm:"column:last_meter_value"`
	NextDueAt      *time.Time `gorm:"column:next_due_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (MaintenanceSchedule) TableName() string { return "maintenance_schedule" }
