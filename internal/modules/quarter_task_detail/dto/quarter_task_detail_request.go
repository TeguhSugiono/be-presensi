package dto

type UpsertQuarterTaskDetailRequest struct {
	AchievementText string `json:"achievement_text" binding:"required"`
	Action          string `json:"action" binding:"required,oneof=draft submit"`
}
