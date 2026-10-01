package entity

import "time"

type WorkSchedule struct {
	ID         int     `gorm:"column:id;primaryKey"`
	Name       string  `gorm:"column:name;type:varchar(50);not null"`
	StartTime  *string `gorm:"column:start_time;type:time"`
	EndTime    *string `gorm:"column:end_time;type:time"`
	IsFlexible bool    `gorm:"column:is_flexible;not null;default:false"`

	CreatedBy *int64    `gorm:"column:created_by"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`

	UpdatedBy *int64    `gorm:"column:updated_by"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`

	IsDeleted bool `gorm:"column:is_deleted;not null;default:false"`
}

func (WorkSchedule) TableName() string {
	return "work_schedules"
}
