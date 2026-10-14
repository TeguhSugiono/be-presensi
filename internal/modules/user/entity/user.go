package entity

import "time"

type User struct {
	ID             int64   `gorm:"column:id;primaryKey"`
	RoleID         int     `gorm:"column:role_id;not null"`
	Role           *string `gorm:"-" json:"-"`
	WorkScheduleID *int    `gorm:"column:work_schedule_id"`
	NIP            *string `gorm:"column:nip;type:varchar(30)"`
	FullName       string  `gorm:"column:full_name;type:varchar(150);not null"`
	Email          string  `gorm:"column:email;type:varchar(100);not null"`
	PasswordHash   string  `gorm:"column:password_hash;type:varchar(255);not null"`
	PhoneNumber    string  `gorm:"column:phone_number;type:varchar(20);not null"`
	IsActive       bool    `gorm:"column:is_active;not null;default:true"`

	CreatedBy *int64     `gorm:"column:created_by"`
	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	UpdatedBy *int64     `gorm:"column:updated_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime:false"`

	IsDeleted bool `gorm:"column:is_deleted;not null;default:false"`
}

func (User) TableName() string {
	return "users"
}
