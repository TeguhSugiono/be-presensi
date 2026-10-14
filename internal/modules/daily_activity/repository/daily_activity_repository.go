// package repository

// import (
// 	"context"

// 	"be-presensi/internal/modules/daily_activity/entity"

// 	"gorm.io/gorm"
// 	"gorm.io/gorm/clause"
// )

// type DailyActivityRepository interface {
// 	WithTx(tx *gorm.DB) DailyActivityRepository
// 	FindByAttendanceIDForUpdate(
// 		ctx context.Context,
// 		attendanceID int64,
// 		userID int64,
// 	) (*entity.DailyActivity, error)
// 	Lock(
// 		ctx context.Context,
// 		activity *entity.DailyActivity,
// 	) error
// }

// type dailyActivityRepository struct {
// 	db *gorm.DB
// }

// func NewDailyActivityRepository(
// 	db *gorm.DB,
// ) DailyActivityRepository {
// 	return &dailyActivityRepository{db: db}
// }

// func (r *dailyActivityRepository) WithTx(
// 	tx *gorm.DB,
// ) DailyActivityRepository {
// 	return &dailyActivityRepository{db: tx}
// }

// func (r *dailyActivityRepository) FindByAttendanceIDForUpdate(
// 	ctx context.Context,
// 	attendanceID int64,
// 	userID int64,
// ) (*entity.DailyActivity, error) {
// 	var activity entity.DailyActivity

// 	err := r.db.WithContext(ctx).
// 		Clauses(clause.Locking{Strength: "UPDATE"}).
// 		Where(
// 			"attendance_id = ? AND user_id = ? AND is_deleted = ?",
// 			attendanceID,
// 			userID,
// 			false,
// 		).
// 		First(&activity).Error
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &activity, nil
// }

// func (r *dailyActivityRepository) Lock(
// 	ctx context.Context,
// 	activity *entity.DailyActivity,
// ) error {
// 	return r.db.WithContext(ctx).
// 		Model(&entity.DailyActivity{}).
// 		Where("id = ? AND is_deleted = ?", activity.ID, false).
// 		Updates(map[string]interface{}{
// 			"is_locked":  true,
// 			"updated_by": activity.UpdatedBy,
// 			"updated_at": activity.UpdatedAt,
// 		}).Error
// }

package repository

import (
	"context"

	"be-presensi/internal/modules/daily_activity/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DailyActivityRepository interface {
	WithTx(tx *gorm.DB) DailyActivityRepository
	FindByAttendanceIDForUpdate(
		ctx context.Context,
		attendanceID int64,
		userID int64,
	) (*entity.DailyActivity, error)
	Lock(
		ctx context.Context,
		activity *entity.DailyActivity,
	) error

	FindByAttendanceID(
		ctx context.Context,
		attendanceID int64,
		userID int64,
	) (*entity.DailyActivity, error)

	Create(
		ctx context.Context,
		activity *entity.DailyActivity,
	) error

	Update(
		ctx context.Context,
		activity *entity.DailyActivity,
	) error
}

type dailyActivityRepository struct {
	db *gorm.DB
}

func NewDailyActivityRepository(
	db *gorm.DB,
) DailyActivityRepository {
	return &dailyActivityRepository{db: db}
}

func (r *dailyActivityRepository) WithTx(
	tx *gorm.DB,
) DailyActivityRepository {
	return &dailyActivityRepository{db: tx}
}

func (r *dailyActivityRepository) FindByAttendanceIDForUpdate(
	ctx context.Context,
	attendanceID int64,
	userID int64,
) (*entity.DailyActivity, error) {
	var activity entity.DailyActivity

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"attendance_id = ? AND user_id = ? AND is_deleted = ?",
			attendanceID,
			userID,
			false,
		).
		First(&activity).Error
	if err != nil {
		return nil, err
	}

	return &activity, nil
}

func (r *dailyActivityRepository) Lock(
	ctx context.Context,
	activity *entity.DailyActivity,
) error {
	return r.db.WithContext(ctx).
		Model(&entity.DailyActivity{}).
		Where("id = ? AND is_deleted = ?", activity.ID, false).
		Updates(map[string]interface{}{
			"is_locked":  true,
			"updated_by": activity.UpdatedBy,
			"updated_at": activity.UpdatedAt,
		}).Error
}

func (r *dailyActivityRepository) FindByAttendanceID(
	ctx context.Context,
	attendanceID int64,
	userID int64,
) (*entity.DailyActivity, error) {
	var activity entity.DailyActivity

	err := r.db.WithContext(ctx).
		Where(
			"attendance_id = ? AND user_id = ? AND is_deleted = ?",
			attendanceID,
			userID,
			false,
		).
		First(&activity).Error
	if err != nil {
		return nil, err
	}

	return &activity, nil
}

func (r *dailyActivityRepository) Create(
	ctx context.Context,
	activity *entity.DailyActivity,
) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

func (r *dailyActivityRepository) Update(
	ctx context.Context,
	activity *entity.DailyActivity,
) error {
	return r.db.WithContext(ctx).
		Model(&entity.DailyActivity{}).
		Where(
			"id = ? AND user_id = ? AND is_deleted = ? AND is_locked = ?",
			activity.ID,
			activity.UserID,
			false,
			false,
		).
		Updates(map[string]interface{}{
			"description": activity.Description,
			"updated_by":  activity.UpdatedBy,
			"updated_at":  activity.UpdatedAt,
		}).Error
}
