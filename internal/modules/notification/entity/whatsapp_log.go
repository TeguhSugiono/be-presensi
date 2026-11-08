package entity

import "time"

type WhatsAppLog struct {
	ID                 int64      `gorm:"column:id;primaryKey"`
	AttendanceID       *int64     `gorm:"column:attendance_id"`
	UserID             *int64     `gorm:"column:user_id"`
	Provider           string     `gorm:"column:provider;type:varchar(30);not null"`
	Target             string     `gorm:"column:target;type:varchar(30);not null"`
	Message            string     `gorm:"column:message;type:text;not null"`
	Status             string     `gorm:"column:status;type:varchar(20);not null"`
	AttemptCount       int        `gorm:"column:attempt_count;not null;default:0"`
	SuccessfulConfigID *int64     `gorm:"column:successful_config_id"`
	AttemptDetails     *string    `gorm:"column:attempt_details;type:jsonb"`
	ErrorMessage       *string    `gorm:"column:error_message;type:text"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null"`
	SentAt             *time.Time `gorm:"column:sent_at"`
}

func (WhatsAppLog) TableName() string {
	return "whatsapp_logs"
}
