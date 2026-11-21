package entity

import "time"

type QuarterTask struct {
	ID          *int64     `gorm:"column:id;primaryKey"`
	Title       *string    `gorm:"column:title;type:varchar(150);not null"`
	Description *string    `gorm:"column:description;type:text"`
	Year        *int       `gorm:"column:year;not null"`
	Quarter     *string    `gorm:"column:quarter;type:varchar(2);not null"`
	StartDate   *time.Time `gorm:"column:start_date;type:date;not null"`
	EndDate     *time.Time `gorm:"column:end_date;type:date;not null"`

	CreatedBy *int64     `gorm:"column:created_by;not null"`
	CreatedAt *time.Time `gorm:"column:created_at;not null"`
	UpdatedBy *int64     `gorm:"column:updated_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime:false"`
	IsDeleted *bool      `gorm:"column:is_deleted;not null;default:false"`
}

func (*QuarterTask) TableName() string {
	return "quarter_tasks"
}
