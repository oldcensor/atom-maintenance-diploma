package domain

import (
	"context"
	"time"
)

type EquipmentType struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

type EquipmentTypeRepository interface {
	Create(ctx context.Context, et *EquipmentType) (*EquipmentType, error)
	GetByID(ctx context.Context, id int64) (*EquipmentType, error)
	List(ctx context.Context) ([]EquipmentType, error)
	Update(ctx context.Context, et *EquipmentType) (*EquipmentType, error)
	Delete(ctx context.Context, id int64) error
}
