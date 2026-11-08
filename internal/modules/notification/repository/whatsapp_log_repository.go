package repository

import (
	"context"

	"be-presensi/internal/modules/notification/entity"

	"gorm.io/gorm"
)

type WhatsAppLogRepository interface {
	Create(ctx context.Context, log *entity.WhatsAppLog) error
}

type whatsappLogRepository struct {
	db *gorm.DB
}

func NewWhatsAppLogRepository(db *gorm.DB) WhatsAppLogRepository {
	return &whatsappLogRepository{
		db: db,
	}
}

func (r *whatsappLogRepository) Create(
	ctx context.Context,
	log *entity.WhatsAppLog,
) error {
	return r.db.WithContext(ctx).Create(log).Error
}
