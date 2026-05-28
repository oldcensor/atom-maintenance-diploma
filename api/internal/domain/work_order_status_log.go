package domain

import (
	"context"
	"time"
)

type WorkOrderStatusLog struct {
	ID          int64
	WorkOrderID int64
	FromStatus  WorkOrderStatus
	ToStatus    WorkOrderStatus
	ChangedBy   *int64
	Comment     string
	CreatedAt   time.Time
}

type WorkOrderStatusLogRepository interface {
	Create(ctx context.Context, l *WorkOrderStatusLog) (*WorkOrderStatusLog, error)
	ListByWorkOrderID(ctx context.Context, workOrderID int64) ([]WorkOrderStatusLog, error)
}
