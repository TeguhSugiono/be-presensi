package repository

import (
	"context"

	"be-presensi/internal/modules/notification/entity"

	"gorm.io/gorm"
)

type WhatsAppConfigRepository interface {
	FindActive(ctx context.Context) (*entity.WhatsAppConfig, error)
	FindByID(ctx context.Context, id int64) (*entity.WhatsAppConfig, error)
	Create(ctx context.Context, config *entity.WhatsAppConfig) error
	Update(ctx context.Context, config *entity.WhatsAppConfig) error
	Delete(ctx context.Context, config *entity.WhatsAppConfig) error
	FindAllActive() ([]entity.WhatsAppConfig, error)
}

type whatsappConfigRepository struct {
	db *gorm.DB
}

func NewWhatsAppConfigRepository(db *gorm.DB) WhatsAppConfigRepository {
	return &whatsappConfigRepository{
		db: db,
	}
}

func (r *whatsappConfigRepository) FindActive(
	ctx context.Context,
) (*entity.WhatsAppConfig, error) {
	var config *entity.WhatsAppConfig

	err := r.db.WithContext(ctx).
		Where(
			"is_active = ? AND is_deleted = ?",
			true,
			false,
		).
		Order("id DESC").
		First(&config).Error

	if err != nil {
		return nil, err
	}

	return config, nil
}

func (r *whatsappConfigRepository) FindByID(
	ctx context.Context,
	id int64,
) (*entity.WhatsAppConfig, error) {
	var config *entity.WhatsAppConfig

	err := r.db.WithContext(ctx).
		Where(
			"id = ? AND is_deleted = ?",
			id,
			false,
		).
		First(&config).Error

	if err != nil {
		return nil, err
	}

	return config, nil
}

func (r *whatsappConfigRepository) Create(
	ctx context.Context,
	config *entity.WhatsAppConfig,
) error {
	return r.db.WithContext(ctx).
		Create(config).Error
}

func (r *whatsappConfigRepository) Update(
	ctx context.Context,
	config *entity.WhatsAppConfig,
) error {
	return r.db.WithContext(ctx).
		Save(config).Error
}

func (r *whatsappConfigRepository) Delete(
	ctx context.Context,
	config *entity.WhatsAppConfig,
) error {
	return r.db.WithContext(ctx).
		Model(&entity.WhatsAppConfig{}).
		Where(
			"id = ? AND is_deleted = ?",
			config.ID,
			false,
		).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"is_active":  false,
			"updated_at": config.UpdatedAt,
			"updated_by": config.UpdatedBy,
		}).Error
}

func (r *whatsappConfigRepository) FindAllActive() ([]entity.WhatsAppConfig, error) {
	var configs []entity.WhatsAppConfig

	err := r.db.
		Where("is_active = ? AND is_deleted = ?", true, false).
		Order("priority ASC, id ASC").
		Find(&configs).Error

	return configs, err
}
