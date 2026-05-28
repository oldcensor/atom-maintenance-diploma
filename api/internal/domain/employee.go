package domain

import (
	"context"
	"time"
)

type EmployeeRole string

const (
	RoleTechnician EmployeeRole = "technician"
	RoleEngineer   EmployeeRole = "engineer"
	RoleManager    EmployeeRole = "manager"
	RoleAdmin      EmployeeRole = "admin"
)

var roleOrder = []EmployeeRole{RoleTechnician, RoleEngineer, RoleManager, RoleAdmin}

func (r EmployeeRole) AtLeast(min EmployeeRole) bool {
	ri, mi := -1, -1
	for i, v := range roleOrder {
		if v == r {
			ri = i
		}
		if v == min {
			mi = i
		}
	}
	return ri >= mi && ri != -1
}

type Employee struct {
	ID             int64
	Email          string
	PasswordHash   string
	FullName       string
	Role           EmployeeRole
	DepartmentID   *int64
	FailedAttempts int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

type EmployeeRepository interface {
	Create(ctx context.Context, e *Employee) (*Employee, error)
	GetByID(ctx context.Context, id int64) (*Employee, error)
	GetByEmail(ctx context.Context, email string) (*Employee, error)
	List(ctx context.Context) ([]Employee, error)
	Update(ctx context.Context, e *Employee) (*Employee, error)
	SoftDelete(ctx context.Context, id int64) error
	IncrFailedAttempts(ctx context.Context, id int64) error
	ResetFailedAttempts(ctx context.Context, id int64) error
}
