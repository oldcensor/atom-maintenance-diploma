package domain

import (
	"context"
	"time"
)

type InspectionReport struct {
	ID              int64
	WorkOrderID     int64
	InspectorID     int64
	Findings        string
	Recommendations string
	CreatedAt       time.Time
}

type InspectionReportRepository interface {
	Create(ctx context.Context, r *InspectionReport) (*InspectionReport, error)
	GetByID(ctx context.Context, id int64) (*InspectionReport, error)
	List(ctx context.Context) ([]InspectionReport, error)
	Delete(ctx context.Context, id int64) error
}
