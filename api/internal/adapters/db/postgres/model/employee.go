package model

import "time"

type Employee struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Email          string     `gorm:"column:email;not null;uniqueIndex"`
	PasswordHash   string     `gorm:"column:password_hash;not null"`
	FullName       string     `gorm:"column:full_name;not null"`
	Role           string     `gorm:"column:role;not null;default:technician"`
	DepartmentID   *int64     `gorm:"column:department_id"`
	FailedAttempts int        `gorm:"column:failed_attempts;not null;default:0"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (Employee) TableName() string { return "employee" }
