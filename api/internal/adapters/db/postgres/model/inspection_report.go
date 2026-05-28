package model

import "time"

type InspectionReport struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	WorkOrderID     int64     `gorm:"column:work_order_id;not null"`
	InspectorID     int64     `gorm:"column:inspector_id;not null"`
	Findings        string    `gorm:"column:findings;not null"`
	Recommendations string    `gorm:"column:recommendations"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (InspectionReport) TableName() string { return "inspection_report" }
