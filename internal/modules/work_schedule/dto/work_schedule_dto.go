package dto

type CreateWorkScheduleRequest struct {
	Name       string  `json:"name" binding:"required,max=50"`
	StartTime  *string `json:"start_time"`
	EndTime    *string `json:"end_time"`
	IsFlexible bool    `json:"is_flexible"`
}

type UpdateWorkScheduleRequest struct {
	Name       string  `json:"name" binding:"required,max=50"`
	StartTime  *string `json:"start_time"`
	EndTime    *string `json:"end_time"`
	IsFlexible bool    `json:"is_flexible"`
}

type WorkScheduleResponse struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	StartTime  *string `json:"start_time"`
	EndTime    *string `json:"end_time"`
	IsFlexible bool    `json:"is_flexible"`
}
