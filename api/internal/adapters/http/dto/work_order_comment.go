package dto

import "time"

type CreateCommentRequest struct {
	Text string `json:"text" validate:"required,min=1"`
}

type CommentResponse struct {
	ID          int64     `json:"id"`
	WorkOrderID int64     `json:"work_order_id"`
	AuthorID    int64     `json:"author_id"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
}
