package model

import "time"

type WorkOrderStatusLog struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	WorkOrderID int64     `gorm:"column:work_order_id;not null"`
	FromStatus  string    `gorm:"column:from_status;not null"`
	ToStatus    string    `gorm:"column:to_status;not null"`
	ChangedBy   *int64    `gorm:"column:changed_by"`
	Comment     string    `gorm:"column:comment"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (WorkOrderStatusLog) TableName() string { return "work_order_status_log" }
