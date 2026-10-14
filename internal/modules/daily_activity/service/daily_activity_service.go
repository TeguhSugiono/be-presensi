package service

import (
	"context"
	"errors"
	"strings"
	"time"

	attendanceRepository "be-presensi/internal/modules/attendance/repository"
	"be-presensi/internal/modules/daily_activity/dto"
	"be-presensi/internal/modules/daily_activity/entity"
	dailyActivityRepository "be-presensi/internal/modules/daily_activity/repository"

	"gorm.io/gorm"
)

var (
	ErrAttendanceNotFound       = errors.New("attendance not found")
	ErrAttendanceAlreadyChecked = errors.New("attendance has already been checked out")
	ErrActivityAlreadyExists    = errors.New("daily activity already exists")
	ErrActivityNotFound         = errors.New("daily activity not found")
	ErrActivityLocked           = errors.New("daily activity is already locked")
	ErrDescriptionRequired      = errors.New("description is required")
)

type DailyActivityService struct {
	db             *gorm.DB
	attendanceRepo attendanceRepository.AttendanceRepository
	activityRepo   dailyActivityRepository.DailyActivityRepository
}

func NewDailyActivityService(
	db *gorm.DB,
	attendanceRepo attendanceRepository.AttendanceRepository,
	activityRepo dailyActivityRepository.DailyActivityRepository,
) *DailyActivityService {
	return &DailyActivityService{
		db:             db,
		attendanceRepo: attendanceRepo,
		activityRepo:   activityRepo,
	}
}

func (s *DailyActivityService) Create(
	ctx context.Context,
	userID int64,
	attendanceID int64,
	req dto.CreateDailyActivityRequest,
) (*entity.DailyActivity, error) {
	description := strings.TrimSpace(req.Description)
	if description == "" {
		return nil, ErrDescriptionRequired
	}

	var result *entity.DailyActivity

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txAttendanceRepo := s.attendanceRepo.WithTx(tx)
		txActivityRepo := s.activityRepo.WithTx(tx)

		attendance, err := txAttendanceRepo.FindByIDForUpdate(
			ctx,
			attendanceID,
		)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAttendanceNotFound
			}
			return err
		}

		if attendance.UserID != userID {
			return ErrAttendanceNotFound
		}
		if attendance.CheckOutTime != nil {
			return ErrAttendanceAlreadyChecked
		}

		existing, err := txActivityRepo.FindByAttendanceID(
			ctx,
			attendanceID,
			userID,
		)
		if err == nil && existing != nil {
			return ErrActivityAlreadyExists
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now()
		createdBy := userID

		activity := &entity.DailyActivity{
			AttendanceID: attendance.ID,
			UserID:       userID,
			ActivityDate: attendance.AttendanceDate,
			Description:  description,
			IsLocked:     false,
			CreatedBy:    &createdBy,
			CreatedAt:    now,
			IsDeleted:    false,
		}

		if err := txActivityRepo.Create(ctx, activity); err != nil {
			return err
		}

		result = activity
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DailyActivityService) GetByAttendance(
	ctx context.Context,
	userID int64,
	attendanceID int64,
) (*entity.DailyActivity, error) {
	attendance, err := s.attendanceRepo.FindByID(ctx, attendanceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceNotFound
		}
		return nil, err
	}

	if attendance.UserID != userID {
		return nil, ErrAttendanceNotFound
	}

	activity, err := s.activityRepo.FindByAttendanceID(
		ctx,
		attendanceID,
		userID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivityNotFound
		}
		return nil, err
	}

	return activity, nil
}

func (s *DailyActivityService) Update(
	ctx context.Context,
	userID int64,
	attendanceID int64,
	req dto.UpdateDailyActivityRequest,
) (*entity.DailyActivity, error) {
	description := strings.TrimSpace(req.Description)
	if description == "" {
		return nil, ErrDescriptionRequired
	}

	var result *entity.DailyActivity

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txAttendanceRepo := s.attendanceRepo.WithTx(tx)
		txActivityRepo := s.activityRepo.WithTx(tx)

		attendance, err := txAttendanceRepo.FindByIDForUpdate(
			ctx,
			attendanceID,
		)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAttendanceNotFound
			}
			return err
		}

		if attendance.UserID != userID {
			return ErrAttendanceNotFound
		}

		activity, err := txActivityRepo.FindByAttendanceIDForUpdate(
			ctx,
			attendanceID,
			userID,
		)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrActivityNotFound
			}
			return err
		}

		if activity.IsLocked || attendance.CheckOutTime != nil {
			return ErrActivityLocked
		}

		now := time.Now()
		updatedBy := userID

		activity.Description = description
		activity.UpdatedBy = &updatedBy
		activity.UpdatedAt = &now

		if err := txActivityRepo.Update(ctx, activity); err != nil {
			return err
		}

		result = activity
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
