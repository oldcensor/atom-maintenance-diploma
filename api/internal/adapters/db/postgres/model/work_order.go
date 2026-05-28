package model

import "time"

type WorkOrder struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	ScheduleID  *int64     `gorm:"column:schedule_id"`
	EquipmentID int64      `gorm:"column:equipment_id;not null"`
	Title       string     `gorm:"column:title;not null"`
	Description string     `gorm:"column:description"`
	AssignedTo  *int64     `gorm:"column:assigned_to"`
	CreatedBy   *int64     `gorm:"column:created_by"`
	Status      string     `gorm:"column:status;not null;default:open"`
	WorkType    string     `gorm:"column:work_type;not null;default:corrective"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	CompletedAt *time.Time `gorm:"column:completed_at"`
}

func (WorkOrder) TableName() string { return "work_order" }
