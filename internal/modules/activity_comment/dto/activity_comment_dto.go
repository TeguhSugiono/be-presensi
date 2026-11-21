package dto

type CreateActivityCommentRequest struct {
	DailyActivityID int64  `json:"daily_activity_id" binding:"required"`
	Comment         string `json:"comment" binding:"required"`
	Status          int16  `json:"status" binding:"required,oneof=1 2"`
}

type ActivityCommentResponse struct {
	ID              int64  `json:"id"`
	DailyActivityID int64  `json:"daily_activity_id"`
	DailyTaskDate   string `json:"daily_task_date"`
	Status          int16  `json:"status"`
	KepsekID        int64  `json:"kepsek_id"`
	Comment         string `json:"comment"`
	WANotified      bool   `json:"wa_notified"`
	CreatedAt       string `json:"created_at"`
}
