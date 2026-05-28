package dto

import "time"

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CreateEmployeeRequest struct {
	Email        string `json:"email"         validate:"required,email"`
	Password     string `json:"password"      validate:"required,min=8"`
	FullName     string `json:"full_name"     validate:"required,max=255"`
	Role         string `json:"role"          validate:"required,oneof=technician engineer manager admin"`
	DepartmentID *int64 `json:"department_id"`
}

type UpdateEmployeeRequest struct {
	FullName     string `json:"full_name"     validate:"required,max=255"`
	Role         string `json:"role"          validate:"required,oneof=technician engineer manager admin"`
	DepartmentID *int64 `json:"department_id"`
}

type EmployeeResponse struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	DepartmentID *int64    `json:"department_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
