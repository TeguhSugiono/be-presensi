package dto

type CreateActivityCommentRequest struct {
	DailyActivityID int64  `json:"daily_activity_id" binding:"required"`
	Comment         string `json:"comment" binding:"required"`
}

type ActivityCommentResponse struct {
	ID              int64  `json:"id"`
	DailyActivityID int64  `json:"daily_activity_id"`
	KepsekID        int64  `json:"kepsek_id"`
	Comment         string `json:"comment"`
	WANotified      bool   `json:"wa_notified"`
	CreatedAt       string `json:"created_at"`
}
