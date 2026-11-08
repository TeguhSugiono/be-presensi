package entity

import "time"

type WhatsAppConfig struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	Provider     string     `gorm:"column:provider;type:varchar(30);not null"`
	APIToken     string     `gorm:"column:api_token;type:text;not null"`
	SenderNumber *string    `gorm:"column:sender_number;type:varchar(20)"`
	Priority     int        `gorm:"column:priority;not null;default:1"`
	IsActive     bool       `gorm:"column:is_active;not null;default:true"`
	CreatedBy    *int64     `gorm:"column:created_by"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	UpdatedBy    *int64     `gorm:"column:updated_by"`
	UpdatedAt    *time.Time `gorm:"column:updated_at"`
	IsDeleted    bool       `gorm:"column:is_deleted;not null;default:false"`
}

func (WhatsAppConfig) TableName() string {
	return "whatsapp_configs"
}
