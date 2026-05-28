package domain

import (
	"context"
	"time"
)

type EquipmentStatus string

const (
	StatusActive           EquipmentStatus = "active"
	StatusInactive         EquipmentStatus = "inactive"
	StatusUnderMaintenance EquipmentStatus = "under_maintenance"
	StatusDecommissioned   EquipmentStatus = "decommissioned"
)

type Equipment struct {
	ID              int64
	Name            string
	Description     string
	SerialNumber    string
	EquipmentTypeID int64
	DepartmentID    *int64
	ResponsibleID   *int64
	Status          EquipmentStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type EquipmentRepository interface {
	Create(ctx context.Context, e *Equipment) (*Equipment, error)
	GetByID(ctx context.Context, id int64) (*Equipment, error)
	List(ctx context.Context) ([]Equipment, error)
	Update(ctx context.Context, e *Equipment) (*Equipment, error)
	Delete(ctx context.Context, id int64) error
}
