package repository

import (
	"context"
	"time"

	"be-presensi/internal/modules/role/entity"

	"gorm.io/gorm"
)

type RoleRepository interface {
	FindAll(ctx context.Context) ([]entity.Role, error)
	FindByID(ctx context.Context, id int) (*entity.Role, error)
	FindByName(ctx context.Context, name string) (*entity.Role, error)
	Create(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, role *entity.Role) error
	FindAllPaginated(ctx context.Context, search string, limit int, offset int) ([]entity.Role, int64, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{
		db: db,
	}
}

func (r *roleRepository) FindAll(ctx context.Context) ([]entity.Role, error) {
	var roles []entity.Role

	err := r.db.WithContext(ctx).
		Where("is_deleted = ?", false).
		Order("id ASC").
		Find(&roles).Error

	return roles, err
}

func (r *roleRepository) FindAllPaginated(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) ([]entity.Role, int64, error) {
	var roles []entity.Role
	var total int64

	query := r.db.WithContext(ctx).
		Model(&entity.Role{}).
		Where("is_deleted = ?", false)

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&roles).Error

	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *roleRepository) FindByID(ctx context.Context, id int) (*entity.Role, error) {
	var role entity.Role

	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role

	err := r.db.WithContext(ctx).
		Where("name = ? AND is_deleted = ?", name, false).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).
		Create(role).Error
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).
		Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).
		Model(&entity.Role{}).
		Where("id = ? AND is_deleted = ?", role.ID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_at": time.Now(),
		}).Error
}
