package domain

import (
	"context"
	"time"
)

type ChecklistItem struct {
	ID          int64
	WorkOrderID int64
	Text        string
	Checked     bool
	CheckedBy   *int64
	CheckedAt   *time.Time
	SortOrder   int
	CreatedAt   time.Time
}

type ChecklistItemRepository interface {
	Create(ctx context.Context, item *ChecklistItem) (*ChecklistItem, error)
	ListByWorkOrderID(ctx context.Context, workOrderID int64) ([]ChecklistItem, error)
	Toggle(ctx context.Context, id int64, checked bool, checkedBy *int64) (*ChecklistItem, error)
	Delete(ctx context.Context, id int64) error
}
