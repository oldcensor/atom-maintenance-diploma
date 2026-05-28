package dto

import "time"

type CreateInspectionReportRequest struct {
	WorkOrderID     int64  `json:"work_order_id" validate:"required,gt=0"`
	InspectorID     int64  `json:"inspector_id" validate:"required,gt=0"`
	Findings        string `json:"findings" validate:"required,min=1"`
	Recommendations string `json:"recommendations"`
}

type InspectionReportResponse struct {
	ID              int64     `json:"id"`
	WorkOrderID     int64     `json:"work_order_id"`
	InspectorID     int64     `json:"inspector_id"`
	Findings        string    `json:"findings"`
	Recommendations string    `json:"recommendations,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
