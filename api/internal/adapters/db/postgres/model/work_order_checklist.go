package model

import "time"

type WorkOrderChecklistItem struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	WorkOrderID int64      `gorm:"column:work_order_id;not null"`
	Text        string     `gorm:"column:text;not null"`
	Checked     bool       `gorm:"column:checked;not null;default:false"`
	CheckedBy   *int64     `gorm:"column:checked_by"`
	CheckedAt   *time.Time `gorm:"column:checked_at"`
	SortOrder   int        `gorm:"column:sort_order;not null;default:0"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (WorkOrderChecklistItem) TableName() string { return "work_order_checklist_item" }
