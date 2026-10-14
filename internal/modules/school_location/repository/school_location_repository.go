package repository

import (
	"context"

	"be-presensi/internal/modules/school_location/entity"

	"gorm.io/gorm"
)

type SchoolLocationRepository interface {
	FindAll(ctx context.Context) ([]entity.SchoolLocation, error)
	FindByID(ctx context.Context, id int) (*entity.SchoolLocation, error)
	FindByName(ctx context.Context, name string) (*entity.SchoolLocation, error)
	Create(ctx context.Context, location *entity.SchoolLocation) error
	Update(ctx context.Context, location *entity.SchoolLocation) error
	Delete(ctx context.Context, location *entity.SchoolLocation) error
	FindAllPaginated(
		ctx context.Context,
		search string,
		limit int,
		offset int,
	) ([]entity.SchoolLocation, int64, error)
}

type schoolLocationRepository struct {
	db *gorm.DB
}

func NewSchoolLocationRepository(db *gorm.DB) SchoolLocationRepository {
	return &schoolLocationRepository{
		db: db,
	}
}

func (r *schoolLocationRepository) FindAll(
	ctx context.Context,
) ([]entity.SchoolLocation, error) {
	var locations []entity.SchoolLocation

	err := r.db.WithContext(ctx).
		Where("is_deleted = ?", false).
		Order("id ASC").
		Find(&locations).Error

	return locations, err
}

func (r *schoolLocationRepository) FindAllPaginated(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) ([]entity.SchoolLocation, int64, error) {
	var locations []entity.SchoolLocation
	var total int64

	query := r.db.WithContext(ctx).
		Model(&entity.SchoolLocation{}).
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
		Find(&locations).Error

	if err != nil {
		return nil, 0, err
	}

	return locations, total, nil
}

func (r *schoolLocationRepository) FindByID(
	ctx context.Context,
	id int,
) (*entity.SchoolLocation, error) {
	var location entity.SchoolLocation

	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&location).Error

	if err != nil {
		return nil, err
	}

	return &location, nil
}

func (r *schoolLocationRepository) FindByName(
	ctx context.Context,
	name string,
) (*entity.SchoolLocation, error) {
	var location entity.SchoolLocation

	err := r.db.WithContext(ctx).
		Where("name = ? AND is_deleted = ?", name, false).
		First(&location).Error

	if err != nil {
		return nil, err
	}

	return &location, nil
}

func (r *schoolLocationRepository) Create(
	ctx context.Context,
	location *entity.SchoolLocation,
) error {
	return r.db.WithContext(ctx).
		Create(location).Error
}

func (r *schoolLocationRepository) Update(
	ctx context.Context,
	location *entity.SchoolLocation,
) error {
	return r.db.WithContext(ctx).
		Save(location).Error
}

func (r *schoolLocationRepository) Delete(
	ctx context.Context,
	location *entity.SchoolLocation,
) error {
	return r.db.WithContext(ctx).
		Model(&entity.SchoolLocation{}).
		Where("id = ? AND is_deleted = ?", location.ID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_at": location.UpdatedAt,
			"updated_by": location.UpdatedBy,
		}).Error
}
