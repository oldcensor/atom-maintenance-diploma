package dto

import "time"

type CreateEquipmentTypeRequest struct {
	Name        string `json:"name"        validate:"required,max=255"`
	Description string `json:"description"`
}

type UpdateEquipmentTypeRequest struct {
	Name        string `json:"name"        validate:"required,max=255"`
	Description string `json:"description"`
}

type EquipmentTypeResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
