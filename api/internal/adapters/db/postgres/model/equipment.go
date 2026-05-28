package model

import "time"

type Equipment struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name            string    `gorm:"column:name;not null"`
	Description     string    `gorm:"column:description"`
	SerialNumber    string    `gorm:"column:serial_number;not null;uniqueIndex"`
	EquipmentTypeID int64     `gorm:"column:equipment_type_id;not null"`
	DepartmentID    *int64    `gorm:"column:department_id"`
	ResponsibleID   *int64    `gorm:"column:responsible_id"`
	Status          string    `gorm:"column:status;not null;default:active"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Equipment) TableName() string { return "equipment" }
