package dto

import (
	"time"

	"be-presensi/internal/modules/daily_activity/entity"
)

type CreateDailyActivityRequest struct {
	Description string `json:"description" binding:"required"`
}

type UpdateDailyActivityRequest struct {
	Description string `json:"description" binding:"required"`
}

type UpsertDailyActivityRequest struct {
	Description string `json:"description"`
	Action      string `json:"action" binding:"required,oneof=draft submit"`
}

type DailyActivityResponse struct {
	ID           int64   `json:"id"`
	AttendanceID int64   `json:"attendance_id"`
	UserID       int64   `json:"user_id"`
	ActivityDate string  `json:"activity_date"`
	Description  string  `json:"description"`
	Status       int16   `json:"status"`
	IsLocked     bool    `json:"is_locked"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    *string `json:"updated_at"`
}

func ToDailyActivityResponse(
	activity *entity.DailyActivity,
) *DailyActivityResponse {
	if activity == nil {
		return nil
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}

	result := &DailyActivityResponse{
		ID:           activity.ID,
		AttendanceID: activity.AttendanceID,
		UserID:       activity.UserID,
		ActivityDate: activity.ActivityDate.In(loc).Format("2006-01-02"),
		Description:  activity.Description,
		Status:       activity.Status,
		IsLocked:     activity.IsLocked,
		CreatedAt:    activity.CreatedAt.In(loc).Format("2006-01-02 15:04:05"),
	}

	if activity.UpdatedAt != nil {
		value := activity.UpdatedAt.In(loc).Format("2006-01-02 15:04:05")
		result.UpdatedAt = &value
	}

	return result
}
