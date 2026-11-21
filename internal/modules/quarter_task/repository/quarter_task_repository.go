package repository

import (
	"context"
	"strings"

	"be-presensi/internal/modules/quarter_task/entity"
	detailentity "be-presensi/internal/modules/quarter_task_detail/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QuarterTaskRepository interface {
	WithTx(tx *gorm.DB) QuarterTaskRepository

	Create(
		ctx context.Context,
		task *entity.QuarterTask,
	) error

	Update(
		ctx context.Context,
		task *entity.QuarterTask,
	) error

	Delete(
		ctx context.Context,
		id int64,
		updatedBy int64,
	) error

	FindByID(
		ctx context.Context,
		id int64,
	) (*entity.QuarterTask, error)

	FindByIDForUpdate(
		ctx context.Context,
		id int64,
	) (*entity.QuarterTask, error)

	// FindAllPaginated(
	// 	ctx context.Context,
	// 	search string,
	// 	year *int,
	// 	limit int,
	// 	offset int,
	// ) ([]*entity.QuarterTask, int64, error)

	FindAllPaginated(
		ctx context.Context,
		search string,
		year *int,
		quarterTaskID *int64,
		limit int,
		offset int,
	) ([]*entity.QuarterTask, int64, error)

	FindAll(
		ctx context.Context,
	) ([]*entity.QuarterTask, error)

	FindAllByYear(
		ctx context.Context,
		year int,
	) ([]*entity.QuarterTask, error)
}

type quarterTaskRepository struct {
	db *gorm.DB
}

func NewQuarterTaskRepository(
	db *gorm.DB,
) QuarterTaskRepository {
	return &quarterTaskRepository{db: db}
}

func (r *quarterTaskRepository) WithTx(
	tx *gorm.DB,
) QuarterTaskRepository {
	return &quarterTaskRepository{db: tx}
}

func (r *quarterTaskRepository) Create(
	ctx context.Context,
	task *entity.QuarterTask,
) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *quarterTaskRepository) Update(
	ctx context.Context,
	task *entity.QuarterTask,
) error {
	result := r.db.WithContext(ctx).
		Model(&entity.QuarterTask{}).
		Where(
			"id = ? AND is_deleted = ?",
			task.ID,
			false,
		).
		Updates(map[string]interface{}{
			"title":       task.Title,
			"description": task.Description,
			"year":        task.Year,
			"quarter":     task.Quarter,
			"start_date":  task.StartDate,
			"end_date":    task.EndDate,
			"updated_by":  task.UpdatedBy,
			"updated_at":  task.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *quarterTaskRepository) Delete(
	ctx context.Context,
	id int64,
	updatedBy int64,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task entity.QuarterTask

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(
				"id = ? AND is_deleted = ?",
				id,
				false,
			).
			First(&task).Error

		if err != nil {
			return err
		}

		var total int64

		if err := tx.Model(&detailentity.QuarterTaskDetail{}).
			Where(
				"quarter_task_id = ? AND is_deleted = ?",
				id,
				false,
			).
			Count(&total).Error; err != nil {
			return err
		}

		if total > 0 {
			return gorm.ErrInvalidData
		}

		return tx.Model(&entity.QuarterTask{}).
			Where("id = ? AND is_deleted = ?", id, false).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_by": updatedBy,
			}).Error
	})
}

func (r *quarterTaskRepository) FindByID(
	ctx context.Context,
	id int64,
) (*entity.QuarterTask, error) {
	var task entity.QuarterTask

	err := r.db.WithContext(ctx).
		Where(
			"id = ? AND is_deleted = ?",
			id,
			false,
		).
		First(&task).Error

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *quarterTaskRepository) FindByIDForUpdate(
	ctx context.Context,
	id int64,
) (*entity.QuarterTask, error) {
	var task entity.QuarterTask

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"id = ? AND is_deleted = ?",
			id,
			false,
		).
		First(&task).Error

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *quarterTaskRepository) FindAllPaginated(
	ctx context.Context,
	search string,
	year *int,
	quarterTaskID *int64,
	limit int,
	offset int,
) ([]*entity.QuarterTask, int64, error) {
	var tasks []*entity.QuarterTask
	var total int64

	query := r.db.WithContext(ctx).
		Model(&entity.QuarterTask{}).
		Where("is_deleted = ?", false)

	search = strings.TrimSpace(search)

	if search != "" {
		keyword := "%" + search + "%"
		query = query.Where("title ILIKE ?", keyword)
	}

	// if year != nil {
	// 	query = query.Where("year = ?", *year)
	// }

	if year != nil {
		query = query.Where("year = ?", *year)
	}

	if quarterTaskID != nil {
		query = query.Where("id = ?", *quarterTaskID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("year DESC").
		Order("quarter DESC").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error

	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *quarterTaskRepository) FindAll(
	ctx context.Context,
) ([]*entity.QuarterTask, error) {
	var tasks []*entity.QuarterTask

	err := r.db.WithContext(ctx).
		Where("is_deleted = ?", false).
		Order("year DESC").
		Order("quarter DESC").
		Order("id DESC").
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *quarterTaskRepository) FindAllByYear(
	ctx context.Context,
	year int,
) ([]*entity.QuarterTask, error) {
	var tasks []*entity.QuarterTask

	err := r.db.WithContext(ctx).
		Where(
			"year = ? AND is_deleted = ?",
			year,
			false,
		).
		Order("quarter ASC").
		Order("id ASC").
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}
