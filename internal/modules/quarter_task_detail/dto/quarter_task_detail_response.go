package dto

import (
	"time"

	"be-presensi/internal/modules/quarter_task_detail/entity"
)

type QuarterTaskDetailResponse struct {
	ID              *int64  `json:"id"`
	QuarterTaskID   *int64  `json:"quarter_task_id"`
	UserID          *int64  `json:"user_id"`
	UserName        *string `json:"user_name"`
	AchievementText *string `json:"achievement_text"`
	Status          *string `json:"status"`
	SubmittedAt     *string `json:"submitted_at"`
	CreatedBy       *int64  `json:"created_by"`
	CreatedAt       *string `json:"created_at"`
	UpdatedBy       *int64  `json:"updated_by"`
	UpdatedAt       *string `json:"updated_at"`
}

func ToQuarterTaskDetailResponse(
	detail *entity.QuarterTaskDetail,
) *QuarterTaskDetailResponse {
	if detail == nil {
		return nil
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}

	result := &QuarterTaskDetailResponse{
		ID:              detail.ID,
		QuarterTaskID:   detail.QuarterTaskID,
		UserID:          detail.UserID,
		UserName:        detail.UserName,
		AchievementText: detail.AchievementText,
		Status:          detail.Status,
		CreatedBy:       detail.CreatedBy,
		UpdatedBy:       detail.UpdatedBy,
	}

	if detail.SubmittedAt != nil {
		value := detail.SubmittedAt.In(loc).Format("2006-01-02 15:04:05")
		result.SubmittedAt = &value
	}

	if detail.CreatedAt != nil {
		value := detail.CreatedAt.In(loc).Format("2006-01-02 15:04:05")
		result.CreatedAt = &value
	}

	if detail.UpdatedAt != nil {
		value := detail.UpdatedAt.In(loc).Format("2006-01-02 15:04:05")
		result.UpdatedAt = &value
	}

	return result
}
