package model

import "time"

type EquipmentType struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string    `gorm:"column:name;not null;uniqueIndex"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (EquipmentType) TableName() string { return "equipment_type" }
