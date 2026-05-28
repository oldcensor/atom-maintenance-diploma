package domain

import (
	"context"
	"time"
)

type WorkOrderComment struct {
	ID          int64
	WorkOrderID int64
	AuthorID    int64
	Text        string
	CreatedAt   time.Time
}

type WorkOrderCommentRepository interface {
	Create(ctx context.Context, c *WorkOrderComment) (*WorkOrderComment, error)
	ListByWorkOrderID(ctx context.Context, workOrderID int64) ([]WorkOrderComment, error)
	Delete(ctx context.Context, id int64) error
}
