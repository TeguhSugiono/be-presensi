package repository

import (
	"context"
	"strings"
	"time"

	"be-presensi/internal/modules/attendance/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AttendanceRepository interface {
	WithTx(tx *gorm.DB) AttendanceRepository

	FindAll(ctx context.Context) ([]entity.Attendance, error)
	FindByID(ctx context.Context, id int64) (*entity.Attendance, error)
	FindByUserAndDate(ctx context.Context, userID int64, attendanceDate time.Time) (*entity.Attendance, error)
	Create(ctx context.Context, attendance *entity.Attendance) error
	Update(ctx context.Context, attendance *entity.Attendance) error
	Delete(ctx context.Context, attendance *entity.Attendance) error
	FindAllPaginated(
		ctx context.Context,
		search string,
		limit, offset int,
	) ([]*entity.Attendance, int64, error)

	FindByIDForUpdate(
		ctx context.Context,
		id int64,
	) (*entity.Attendance, error)
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) WithTx(tx *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: tx}
}

func (r *attendanceRepository) FindAll(ctx context.Context) ([]entity.Attendance, error) {
	var attendances []entity.Attendance

	err := r.db.WithContext(ctx).
		Where("is_deleted = ?", false).
		Order("attendance_date DESC, id DESC").
		Find(&attendances).Error

	return attendances, err
}

func (r *attendanceRepository) FindByID(
	ctx context.Context,
	id int64,
) (*entity.Attendance, error) {
	var attendance entity.Attendance

	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&attendance).Error
	if err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (r *attendanceRepository) FindByUserAndDate(
	ctx context.Context,
	userID int64,
	attendanceDate time.Time,
) (*entity.Attendance, error) {
	var attendance entity.Attendance

	err := r.db.WithContext(ctx).
		Where(
			"user_id = ? AND attendance_date = ? AND is_deleted = ?",
			userID,
			attendanceDate.Format("2006-01-02"),
			false,
		).
		First(&attendance).Error
	if err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (r *attendanceRepository) Create(
	ctx context.Context,
	attendance *entity.Attendance,
) error {
	return r.db.WithContext(ctx).Create(attendance).Error
}

func (r *attendanceRepository) Update(
	ctx context.Context,
	attendance *entity.Attendance,
) error {
	return r.db.WithContext(ctx).Save(attendance).Error
}

func (r *attendanceRepository) Delete(
	ctx context.Context,
	attendance *entity.Attendance,
) error {
	return r.db.WithContext(ctx).
		Model(&entity.Attendance{}).
		Where("id = ? AND is_deleted = ?", attendance.ID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_by": attendance.UpdatedBy,
			"updated_at": attendance.UpdatedAt,
		}).Error
}

func (r *attendanceRepository) FindAllPaginated(
	ctx context.Context,
	search string,
	limit, offset int,
) ([]*entity.Attendance, int64, error) {
	var attendances []*entity.Attendance
	var total int64

	query := r.db.WithContext(ctx).
		Model(&entity.Attendance{}).
		Joins("LEFT JOIN users ON users.id = attendances.user_id").
		Where("attendances.is_deleted = ?", false).
		Where("users.is_deleted = ?", false)

	if search != "" {
		keyword := "%" + strings.TrimSpace(search) + "%"

		query = query.Where(
			`(
				users.full_name ILIKE ?
				OR CAST(attendances.user_id AS TEXT) ILIKE ?
				OR CAST(attendances.attendance_date AS TEXT) ILIKE ?
			)`,
			keyword,
			keyword,
			keyword,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Select("attendances.*, users.full_name AS user_name").
		Order("attendances.attendance_date DESC, attendances.id DESC").
		Limit(limit).
		Offset(offset).
		Find(&attendances).Error
	if err != nil {
		return nil, 0, err
	}

	return attendances, total, nil
}

func (r *attendanceRepository) FindByIDForUpdate(
	ctx context.Context,
	id int64,
) (*entity.Attendance, error) {
	var attendance entity.Attendance

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&attendance).Error
	if err != nil {
		return nil, err
	}

	return &attendance, nil
}
