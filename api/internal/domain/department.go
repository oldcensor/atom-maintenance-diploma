package domain

import (
	"context"
	"time"
)

type Department struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

type DepartmentRepository interface {
	Create(ctx context.Context, d *Department) (*Department, error)
	GetByID(ctx context.Context, id int64) (*Department, error)
	List(ctx context.Context) ([]Department, error)
	Update(ctx context.Context, d *Department) (*Department, error)
	Delete(ctx context.Context, id int64) error
}
