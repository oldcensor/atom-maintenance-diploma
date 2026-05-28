package model

import "time"

type WorkOrderComment struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	WorkOrderID int64     `gorm:"column:work_order_id;not null"`
	AuthorID    int64     `gorm:"column:author_id;not null"`
	Text        string    `gorm:"column:text;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (WorkOrderComment) TableName() string { return "work_order_comment" }
