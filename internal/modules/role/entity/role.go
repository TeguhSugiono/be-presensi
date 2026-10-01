package entity

import "time"

type Role struct {
	ID        int    `gorm:"column:id;primaryKey"`
	Name      string `gorm:"column:name;type:varchar(50);not null"`
	IsTeacher bool   `gorm:"column:is_teacher;not null;default:false"`

	CreatedBy *int64    `gorm:"column:created_by"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`

	UpdatedBy *int64    `gorm:"column:updated_by"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`

	IsDeleted bool `gorm:"column:is_deleted;not null;default:false"`
}

func (Role) TableName() string {
	return "roles"
}
