package dto

import (
	"time"

	"be-presensi/internal/modules/quarter_task/entity"
)

type QuarterTaskResponse struct {
	ID          *int64  `json:"id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Year        *int    `json:"year"`
	Quarter     *string `json:"quarter"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	CreatedBy   *int64  `json:"created_by"`
	CreatedAt   *string `json:"created_at"`
	UpdatedBy   *int64  `json:"updated_by"`
	UpdatedAt   *string `json:"updated_at"`
}

func ToQuarterTaskResponse(
	task *entity.QuarterTask,
) *QuarterTaskResponse {
	if task == nil {
		return nil
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}

	result := &QuarterTaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Year:        task.Year,
		Quarter:     task.Quarter,
		CreatedBy:   task.CreatedBy,
	}

	if task.StartDate != nil {
		value := task.StartDate.Format("2006-01-02")
		result.StartDate = &value
	}

	if task.EndDate != nil {
		value := task.EndDate.Format("2006-01-02")
		result.EndDate = &value
	}

	if task.CreatedAt != nil {
		value := task.CreatedAt.In(loc).Format("2006-01-02 15:04:05")
		result.CreatedAt = &value
	}

	result.UpdatedBy = task.UpdatedBy

	if task.UpdatedAt != nil {
		value := task.UpdatedAt.In(loc).Format("2006-01-02 15:04:05")
		result.UpdatedAt = &value
	}

	return result
}
