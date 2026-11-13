package entity

import "time"

type ActivityComment struct {
	ID              int64     `gorm:"column:id;primaryKey"`
	DailyActivityID int64     `gorm:"column:daily_activity_id;not null"`
	KepsekID        int64     `gorm:"column:kepsek_id;not null"`
	Comment         string    `gorm:"column:comment;type:text;not null"`
	WANotified      bool      `gorm:"column:wa_notified;not null;default:false"`
	CreatedBy       *int64    `gorm:"column:created_by"`
	CreatedAt       time.Time `gorm:"column:created_at;not null"`
	UpdatedBy       *int64    `gorm:"column:updated_by"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null"`
	IsDeleted       bool      `gorm:"column:is_deleted;not null;default:false"`
}

func (ActivityComment) TableName() string {
	return "activity_comments"
}
