package repository

import (
	"context"

	"be-presensi/internal/modules/activity_comment/entity"

	"gorm.io/gorm"
)

type ActivityCommentRepository interface {
	WithTx(tx *gorm.DB) ActivityCommentRepository

	Create(
		ctx context.Context,
		comment *entity.ActivityComment,
	) error

	FindByDailyActivityID(
		ctx context.Context,
		dailyActivityID int64,
	) ([]*entity.ActivityComment, error)

	UpdateWANotified(ctx context.Context, id int64, notified bool) error
}

type activityCommentRepository struct {
	db *gorm.DB
}

func NewActivityCommentRepository(
	db *gorm.DB,
) ActivityCommentRepository {
	return &activityCommentRepository{db: db}
}

func (r *activityCommentRepository) WithTx(
	tx *gorm.DB,
) ActivityCommentRepository {
	return &activityCommentRepository{db: tx}
}

func (r *activityCommentRepository) Create(
	ctx context.Context,
	comment *entity.ActivityComment,
) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *activityCommentRepository) FindByDailyActivityID(
	ctx context.Context,
	dailyActivityID int64,
) ([]*entity.ActivityComment, error) {
	var comments []*entity.ActivityComment

	err := r.db.WithContext(ctx).
		Model(&entity.ActivityComment{}).
		Select(`
            activity_comments.*,
            daily_activities.activity_date AS daily_task_date
        `).
		Joins(`
            INNER JOIN daily_activities
            ON daily_activities.id = activity_comments.daily_activity_id
        `).
		Where(
			"activity_comments.daily_activity_id = ? AND activity_comments.is_deleted = ? AND daily_activities.is_deleted = ?",
			dailyActivityID,
			false,
			false,
		).
		Order("activity_comments.created_at ASC, activity_comments.id ASC").
		Find(&comments).Error

	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *activityCommentRepository) UpdateWANotified(
	ctx context.Context,
	id int64,
	notified bool,
) error {
	result := r.db.WithContext(ctx).
		Model(&entity.ActivityComment{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Update("wa_notified", notified)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
