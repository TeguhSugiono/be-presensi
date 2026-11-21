package dto

type CreateQuarterTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Year        int    `json:"year" binding:"required"`
	Quarter     string `json:"quarter" binding:"required,oneof=Q1 Q2 Q3 Q4"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
}

type UpdateQuarterTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Year        int    `json:"year" binding:"required"`
	Quarter     string `json:"quarter" binding:"required,oneof=Q1 Q2 Q3 Q4"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
}
