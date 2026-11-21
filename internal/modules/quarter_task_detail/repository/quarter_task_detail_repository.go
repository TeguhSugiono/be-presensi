package repository

import (
	"context"
	"strings"

	"be-presensi/internal/modules/quarter_task_detail/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QuarterTaskDetailRepository interface {
	WithTx(tx *gorm.DB) QuarterTaskDetailRepository

	FindByTaskAndUser(
		ctx context.Context,
		quarterTaskID int64,
		userID int64,
	) (*entity.QuarterTaskDetail, error)

	FindByTaskAndUserForUpdate(
		ctx context.Context,
		quarterTaskID int64,
		userID int64,
	) (*entity.QuarterTaskDetail, error)

	Create(
		ctx context.Context,
		detail *entity.QuarterTaskDetail,
	) error

	Update(
		ctx context.Context,
		detail *entity.QuarterTaskDetail,
	) error

	FindByID(
		ctx context.Context,
		id int64,
	) (*entity.QuarterTaskDetail, error)

	FindAllByTask(
		ctx context.Context,
		quarterTaskID int64,
	) ([]*entity.QuarterTaskDetail, error)

	FindAllSubmissionsPaginated(
		ctx context.Context,
		search string,
		year *int,
		quarterTaskID *int64,
		limit int,
		offset int,
	) ([]*entity.QuarterTaskDetail, int64, error)
}

type quarterTaskDetailRepository struct {
	db *gorm.DB
}

func NewQuarterTaskDetailRepository(
	db *gorm.DB,
) QuarterTaskDetailRepository {
	return &quarterTaskDetailRepository{db: db}
}

func (r *quarterTaskDetailRepository) WithTx(
	tx *gorm.DB,
) QuarterTaskDetailRepository {
	return &quarterTaskDetailRepository{db: tx}
}

func (r *quarterTaskDetailRepository) FindByTaskAndUser(
	ctx context.Context,
	quarterTaskID int64,
	userID int64,
) (*entity.QuarterTaskDetail, error) {
	var detail entity.QuarterTaskDetail

	err := r.db.WithContext(ctx).
		Model(&entity.QuarterTaskDetail{}).
		Joins(
			"LEFT JOIN users ON users.id = quarter_task_details.user_id",
		).
		Joins(
			"INNER JOIN quarter_tasks ON quarter_tasks.id = quarter_task_details.quarter_task_id",
		).
		Where(
			"quarter_task_details.quarter_task_id = ? AND "+
				"quarter_task_details.user_id = ? AND "+
				"quarter_task_details.is_deleted = ? AND "+
				"quarter_tasks.is_deleted = ?",
			quarterTaskID,
			userID,
			false,
			false,
		).
		Select(
			"quarter_task_details.*, " +
				"users.full_name AS user_name, " +
				"quarter_tasks.year AS year, " +
				"quarter_tasks.quarter AS quarter",
		).
		First(&detail).Error

	if err != nil {
		return nil, err
	}

	return &detail, nil
}

// func (r *quarterTaskDetailRepository) FindByTaskAndUser(
// 	ctx context.Context,
// 	quarterTaskID int64,
// 	userID int64,
// ) (*entity.QuarterTaskDetail, error) {
// 	var detail entity.QuarterTaskDetail

// 	err := r.db.WithContext(ctx).
// 		Where(
// 			"quarter_task_id = ? AND user_id = ? AND is_deleted = ?",
// 			quarterTaskID,
// 			userID,
// 			false,
// 		).
// 		First(&detail).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	return &detail, nil
// }

func (r *quarterTaskDetailRepository) FindByTaskAndUserForUpdate(
	ctx context.Context,
	quarterTaskID int64,
	userID int64,
) (*entity.QuarterTaskDetail, error) {
	var detail entity.QuarterTaskDetail

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"quarter_task_id = ? AND user_id = ? AND is_deleted = ?",
			quarterTaskID,
			userID,
			false,
		).
		First(&detail).Error

	if err != nil {
		return nil, err
	}

	return &detail, nil
}

func (r *quarterTaskDetailRepository) Create(
	ctx context.Context,
	detail *entity.QuarterTaskDetail,
) error {
	return r.db.WithContext(ctx).Create(detail).Error
}

func (r *quarterTaskDetailRepository) Update(
	ctx context.Context,
	detail *entity.QuarterTaskDetail,
) error {
	result := r.db.WithContext(ctx).
		Model(&entity.QuarterTaskDetail{}).
		Where(
			"id = ? AND user_id = ? AND status = ? AND is_deleted = ?",
			detail.ID,
			detail.UserID,
			"draft",
			false,
		).
		Updates(map[string]interface{}{
			"achievement_text": detail.AchievementText,
			"status":           detail.Status,
			"submitted_at":     detail.SubmittedAt,
			"updated_by":       detail.UpdatedBy,
			"updated_at":       detail.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *quarterTaskDetailRepository) FindByID(
	ctx context.Context,
	id int64,
) (*entity.QuarterTaskDetail, error) {
	var detail entity.QuarterTaskDetail

	err := r.db.WithContext(ctx).
		Where(
			"id = ? AND is_deleted = ?",
			id,
			false,
		).
		First(&detail).Error

	if err != nil {
		return nil, err
	}

	return &detail, nil
}

func (r *quarterTaskDetailRepository) FindAllByTask(
	ctx context.Context,
	quarterTaskID int64,
) ([]*entity.QuarterTaskDetail, error) {
	var details []*entity.QuarterTaskDetail

	err := r.db.WithContext(ctx).
		Model(&entity.QuarterTaskDetail{}).
		Joins(
			"LEFT JOIN users ON users.id = quarter_task_details.user_id",
		).
		Where(
			"quarter_task_details.quarter_task_id = ? AND quarter_task_details.is_deleted = ?",
			quarterTaskID,
			false,
		).
		Select(
			"quarter_task_details.*, users.full_name AS user_name",
		).
		Order("users.full_name ASC").
		Find(&details).Error

	if err != nil {
		return nil, err
	}

	return details, nil
}

func (r *quarterTaskDetailRepository) FindAllSubmissionsPaginated(
	ctx context.Context,
	search string,
	year *int,
	quarterTaskID *int64,
	limit int,
	offset int,
) ([]*entity.QuarterTaskDetail, int64, error) {
	var details []*entity.QuarterTaskDetail
	var total int64

	query := r.db.WithContext(ctx).
		Model(&entity.QuarterTaskDetail{}).
		Joins(
			"LEFT JOIN users ON users.id = quarter_task_details.user_id",
		).
		Joins(
			"INNER JOIN quarter_tasks ON quarter_tasks.id = quarter_task_details.quarter_task_id",
		).
		Where(
			"quarter_task_details.is_deleted = ? AND quarter_tasks.is_deleted = ?",
			false,
			false,
		)

	search = strings.TrimSpace(search)

	if search != "" {
		keyword := "%" + search + "%"
		query = query.Where(
			"(users.full_name ILIKE ? OR quarter_task_details.achievement_text ILIKE ?)",
			keyword,
			keyword,
		)
	}

	if year != nil {
		query = query.Where("quarter_tasks.year = ?", *year)
	}

	if quarterTaskID != nil {
		query = query.Where(
			"quarter_task_details.quarter_task_id = ?",
			*quarterTaskID,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Select(
			"quarter_task_details.*, " +
				"users.full_name AS user_name, " +
				"quarter_tasks.year AS year, " +
				"quarter_tasks.quarter AS quarter",
		).
		Order("quarter_task_details.created_at DESC").
		Order("quarter_task_details.id DESC").
		Limit(limit).
		Offset(offset).
		Find(&details).Error

	if err != nil {
		return nil, 0, err
	}

	return details, total, nil
}
