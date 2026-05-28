package dto

import "time"

type CreateChecklistItemRequest struct {
	Text      string `json:"text"       validate:"required,min=1"`
	SortOrder int    `json:"sort_order"`
}

type ToggleChecklistItemRequest struct {
	Checked bool `json:"checked"`
}

type ChecklistItemResponse struct {
	ID          int64      `json:"id"`
	WorkOrderID int64      `json:"work_order_id"`
	Text        string     `json:"text"`
	Checked     bool       `json:"checked"`
	CheckedBy   *int64     `json:"checked_by,omitempty"`
	CheckedAt   *time.Time `json:"checked_at,omitempty"`
	SortOrder   int        `json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
}
