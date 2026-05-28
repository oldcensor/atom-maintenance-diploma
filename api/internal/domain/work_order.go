package domain

import (
	"context"
	"time"
)

type WorkOrderStatus string

const (
	WorkOrderStatusOpen       WorkOrderStatus = "open"
	WorkOrderStatusInProgress WorkOrderStatus = "in_progress"
	WorkOrderStatusCompleted  WorkOrderStatus = "completed"
	WorkOrderStatusCancelled  WorkOrderStatus = "cancelled"
)

type WorkOrderType string

const (
	WorkOrderTypeEmergency  WorkOrderType = "emergency"
	WorkOrderTypeCorrective WorkOrderType = "corrective"
	WorkOrderTypePlanned    WorkOrderType = "planned"
)

type WorkOrder struct {
	ID          int64
	ScheduleID  *int64
	EquipmentID int64
	Title       string
	Description string
	AssignedTo  *int64
	CreatedBy   *int64
	Status      WorkOrderStatus
	WorkType    WorkOrderType
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

type WorkOrderRepository interface {
	Create(ctx context.Context, w *WorkOrder) (*WorkOrder, error)
	GetByID(ctx context.Context, id int64) (*WorkOrder, error)
	List(ctx context.Context) ([]WorkOrder, error)
	Update(ctx context.Context, w *WorkOrder) (*WorkOrder, error)
	Delete(ctx context.Context, id int64) error
	ExistsByScheduleID(ctx context.Context, scheduleID int64, statuses []WorkOrderStatus) (bool, error)
}
