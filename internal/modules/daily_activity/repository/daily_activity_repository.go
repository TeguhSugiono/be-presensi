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
	"time"

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

	FindByUserAndDate(
		ctx context.Context,
		userID int64,
		activityDate time.Time,
	) (*entity.DailyActivity, error)

	FindAllPaginated(
		ctx context.Context,
		search string,
		limit int,
		offset int,
	) ([]*entity.DailyActivity, int64, error)
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
	result := r.db.WithContext(ctx).
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
			"status":      activity.Status,
			"is_locked":   activity.IsLocked,
			"updated_by":  activity.UpdatedBy,
			"updated_at":  activity.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *dailyActivityRepository) FindByUserAndDate(
	ctx context.Context,
	userID int64,
	activityDate time.Time,
) (*entity.DailyActivity, error) {
	var activity entity.DailyActivity

	err := r.db.WithContext(ctx).
		Where(
			"user_id = ? AND activity_date = ? AND is_deleted = ?",
			userID,
			activityDate.Format("2006-01-02"),
			false,
		).
		First(&activity).Error

	if err != nil {
		return nil, err
	}

	return &activity, nil
}

func (r *dailyActivityRepository) FindAllPaginated(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) ([]*entity.DailyActivity, int64, error) {
	var activities []*entity.DailyActivity
	var total int64

	query := r.db.WithContext(ctx).
		Model(&entity.DailyActivity{}).
		Joins(
			"LEFT JOIN users ON users.id = daily_activities.user_id",
		).
		Where("daily_activities.is_deleted = ?", false)

	if search != "" {
		keyword := "%" + search + "%"
		query = query.Where(
			"users.full_name ILIKE ? OR daily_activities.description ILIKE ?",
			keyword,
			keyword,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type activityWithUser struct {
		entity.DailyActivity
		UserName string `gorm:"column:user_name"`
	}

	var rows []activityWithUser

	err := query.
		Select(
			"daily_activities.*, users.full_name AS user_name",
		).
		Order("daily_activities.activity_date DESC, daily_activities.id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error

	if err != nil {
		return nil, 0, err
	}

	for i := range rows {
		activity := rows[i].DailyActivity
		activity.UserName = rows[i].UserName
		activities = append(activities, &activity)
	}

	return activities, total, nil
}
