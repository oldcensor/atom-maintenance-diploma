package dto

import "time"

type StatusLogResponse struct {
	ID          int64     `json:"id"`
	WorkOrderID int64     `json:"work_order_id"`
	FromStatus  string    `json:"from_status"`
	ToStatus    string    `json:"to_status"`
	ChangedBy   *int64    `json:"changed_by,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
