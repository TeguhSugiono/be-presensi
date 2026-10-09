package repository

import (
	"context"

	"be-presensi/internal/modules/work_schedule/entity"

	"gorm.io/gorm"
)

type WorkScheduleRepository interface {
	FindAll(ctx context.Context) ([]entity.WorkSchedule, error)
	FindByID(ctx context.Context, id int) (*entity.WorkSchedule, error)
	FindByName(ctx context.Context, name string) (*entity.WorkSchedule, error)
	Create(ctx context.Context, schedule *entity.WorkSchedule) error
	Update(ctx context.Context, schedule *entity.WorkSchedule) error
	Delete(ctx context.Context, schedule *entity.WorkSchedule) error
	FindAllPaginated(
		ctx context.Context,
		search string,
		limit int,
		offset int,
	) ([]entity.WorkSchedule, int64, error)
}

type workScheduleRepository struct {
	db *gorm.DB
}

func NewWorkScheduleRepository(db *gorm.DB) WorkScheduleRepository {
	return &workScheduleRepository{
		db: db,
	}
}

func (r *workScheduleRepository) FindAll(ctx context.Context) ([]entity.WorkSchedule, error) {
	var schedules []entity.WorkSchedule

	err := r.db.WithContext(ctx).
		Where("is_deleted = ?", false).
		Order("id ASC").
		Find(&schedules).Error

	return schedules, err
}

func (r *workScheduleRepository) FindAllPaginated(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) ([]entity.WorkSchedule, int64, error) {
	var schedules []entity.WorkSchedule
	var total int64

	query := r.db.WithContext(ctx).
		Model(&entity.WorkSchedule{}).
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
		Find(&schedules).Error

	if err != nil {
		return nil, 0, err
	}

	return schedules, total, nil
}

func (r *workScheduleRepository) FindByID(ctx context.Context, id int) (*entity.WorkSchedule, error) {
	var schedule entity.WorkSchedule

	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&schedule).Error

	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

func (r *workScheduleRepository) FindByName(ctx context.Context, name string) (*entity.WorkSchedule, error) {
	var schedule entity.WorkSchedule

	err := r.db.WithContext(ctx).
		Where("name = ? AND is_deleted = ?", name, false).
		First(&schedule).Error

	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

func (r *workScheduleRepository) Create(
	ctx context.Context,
	schedule *entity.WorkSchedule,
) error {
	return r.db.WithContext(ctx).
		Create(schedule).Error
}

func (r *workScheduleRepository) Update(
	ctx context.Context,
	schedule *entity.WorkSchedule,
) error {
	return r.db.WithContext(ctx).
		Save(schedule).Error
}

func (r *workScheduleRepository) Delete(
	ctx context.Context,
	schedule *entity.WorkSchedule,
) error {
	return r.db.WithContext(ctx).
		Model(&entity.WorkSchedule{}).
		Where("id = ? AND is_deleted = ?", schedule.ID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_at": schedule.UpdatedAt,
			"updated_by": schedule.UpdatedBy,
		}).Error
}
