package dto

import "time"

type CreateEquipmentRequest struct {
	Name            string `json:"name"              validate:"required,max=255"`
	Description     string `json:"description"`
	SerialNumber    string `json:"serial_number"     validate:"required,max=255"`
	EquipmentTypeID int64  `json:"equipment_type_id" validate:"required,gt=0"`
	DepartmentID    *int64 `json:"department_id"`
	ResponsibleID   *int64 `json:"responsible_id"`
	Status          string `json:"status"            validate:"omitempty,oneof=active inactive under_maintenance decommissioned"`
}

type UpdateEquipmentRequest struct {
	Name            string `json:"name"              validate:"required,max=255"`
	Description     string `json:"description"`
	SerialNumber    string `json:"serial_number"     validate:"required,max=255"`
	EquipmentTypeID int64  `json:"equipment_type_id" validate:"required,gt=0"`
	DepartmentID    *int64 `json:"department_id"`
	ResponsibleID   *int64 `json:"responsible_id"`
	Status          string `json:"status"            validate:"required,oneof=active inactive under_maintenance decommissioned"`
}

type EquipmentResponse struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	SerialNumber    string    `json:"serial_number"`
	EquipmentTypeID int64     `json:"equipment_type_id"`
	DepartmentID    *int64    `json:"department_id,omitempty"`
	ResponsibleID   *int64    `json:"responsible_id,omitempty"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
