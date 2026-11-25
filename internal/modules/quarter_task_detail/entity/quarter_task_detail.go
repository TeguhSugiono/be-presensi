package entity

import "time"

type QuarterTaskDetail struct {
	ID              *int64     `gorm:"column:id;primaryKey"`
	QuarterTaskID   *int64     `gorm:"column:quarter_task_id;not null"`
	Title           *string    `gorm:"column:title;->"`
	UserID          *int64     `gorm:"column:user_id;not null"`
	UserName        *string    `gorm:"column:user_name;->"`
	Year            *int       `gorm:"column:year;->"`
	Quarter         *string    `gorm:"column:quarter;->"`
	AchievementText *string    `gorm:"column:achievement_text;type:text;not null"`
	Status          *string    `gorm:"column:status;type:varchar(30);not null;default:draft"`
	SubmittedAt     *time.Time `gorm:"column:submitted_at"`

	CreatedBy *int64     `gorm:"column:created_by;not null"`
	CreatedAt *time.Time `gorm:"column:created_at;not null"`
	UpdatedBy *int64     `gorm:"column:updated_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime:false"`
	IsDeleted *bool      `gorm:"column:is_deleted;not null;default:false"`
}

func (*QuarterTaskDetail) TableName() string {
	return "quarter_task_details"
}
