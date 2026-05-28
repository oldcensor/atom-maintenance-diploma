package dto

import "time"

type CreateDepartmentRequest struct {
	Name        string `json:"name"        validate:"required,max=255"`
	Description string `json:"description"`
}

type UpdateDepartmentRequest struct {
	Name        string `json:"name"        validate:"required,max=255"`
	Description string `json:"description"`
}

type DepartmentResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
