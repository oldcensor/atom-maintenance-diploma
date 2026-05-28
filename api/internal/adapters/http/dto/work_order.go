package dto

import "time"

type CreateWorkOrderRequest struct {
	ScheduleID  *int64 `json:"schedule_id"`
	EquipmentID int64  `json:"equipment_id" validate:"required,gt=0"`
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description"`
	AssignedTo  *int64 `json:"assigned_to"`
	Status      string `json:"status" validate:"omitempty,oneof=open in_progress"`
	WorkType    string `json:"work_type" validate:"omitempty,oneof=emergency corrective planned"`
}

type UpdateWorkOrderRequest struct {
	ScheduleID  *int64     `json:"schedule_id"`
	EquipmentID int64      `json:"equipment_id" validate:"required,gt=0"`
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description string     `json:"description"`
	AssignedTo  *int64     `json:"assigned_to"`
	Status      string     `json:"status" validate:"required,oneof=open in_progress completed cancelled"`
	WorkType    string     `json:"work_type" validate:"omitempty,oneof=emergency corrective planned"`
	CompletedAt *time.Time `json:"completed_at"`
}

type WorkOrderResponse struct {
	ID          int64      `json:"id"`
	ScheduleID  *int64     `json:"schedule_id,omitempty"`
	EquipmentID int64      `json:"equipment_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	AssignedTo  *int64     `json:"assigned_to,omitempty"`
	CreatedBy   *int64     `json:"created_by,omitempty"`
	Status      string     `json:"status"`
	WorkType    string     `json:"work_type"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
