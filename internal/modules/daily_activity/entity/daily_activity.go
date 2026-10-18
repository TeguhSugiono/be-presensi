package entity

import "time"

type DailyActivity struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	AttendanceID int64      `gorm:"column:attendance_id;not null"`
	UserID       int64      `gorm:"column:user_id;not null"`
	UserName     string     `gorm:"-" json:"user_name"`
	ActivityDate time.Time  `gorm:"column:activity_date;type:date;not null"`
	Description  string     `gorm:"column:description;type:text;not null"`
	Status       int16      `gorm:"column:status" json:"status"`
	IsLocked     bool       `gorm:"column:is_locked;not null;default:false"`
	CreatedBy    *int64     `gorm:"column:created_by"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	UpdatedBy    *int64     `gorm:"column:updated_by"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;autoUpdateTime:false"`
	IsDeleted    bool       `gorm:"column:is_deleted;not null;default:false"`
}

func (DailyActivity) TableName() string {
	return "daily_activities"
}
