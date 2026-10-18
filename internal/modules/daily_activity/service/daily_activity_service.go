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

func (s *DailyActivityService) GetByDate(
	ctx context.Context,
	userID int64,
	activityDate time.Time,
) (*entity.DailyActivity, error) {
	activity, err := s.activityRepo.FindByUserAndDate(
		ctx,
		userID,
		activityDate,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivityNotFound
		}
		return nil, err
	}

	return activity, nil
}

const (
	DailyActivityStatusDraft     int16 = 1
	DailyActivityStatusSubmitted int16 = 2
	DailyActivityStatusApproved  int16 = 3
)

func (s *DailyActivityService) Upsert(
	ctx context.Context,
	userID int64,
	attendanceID int64,
	req dto.UpsertDailyActivityRequest,
) (*entity.DailyActivity, error) {
	action := strings.TrimSpace(req.Action)
	description := strings.TrimSpace(req.Description)

	if action != "draft" && action != "submit" {
		return nil, errors.New("invalid daily activity action")
	}

	// Laporan yang dikirim harus memiliki deskripsi.
	if action == "submit" && description == "" {
		return nil, ErrDescriptionRequired
	}

	var result *entity.DailyActivity

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		attendanceRepo := s.attendanceRepo.WithTx(tx)
		activityRepo := s.activityRepo.WithTx(tx)

		attendance, err := attendanceRepo.FindByIDForUpdate(
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

		activity, err := activityRepo.FindByAttendanceIDForUpdate(
			ctx,
			attendanceID,
			userID,
		)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			now := time.Now()

			status := DailyActivityStatusDraft
			isLocked := false

			if action == "submit" {
				status = DailyActivityStatusSubmitted
				isLocked = true
			}

			activity = &entity.DailyActivity{
				AttendanceID: attendanceID,
				UserID:       userID,
				ActivityDate: attendance.AttendanceDate,
				Description:  description,
				Status:       status,
				IsLocked:     isLocked,
				CreatedBy:    &userID,
				CreatedAt:    now,
				IsDeleted:    false,
			}

			if err := activityRepo.Create(ctx, activity); err != nil {
				return err
			}

			result = activity
			return nil
		}

		if err != nil {
			return err
		}

		// Laporan yang sudah disubmit atau approved tidak dapat
		// diedit maupun disubmit ulang oleh guru.
		if activity.Status != DailyActivityStatusDraft ||
			activity.IsLocked {
			return ErrActivityLocked
		}

		now := time.Now()

		activity.Description = description
		activity.UpdatedBy = &userID
		activity.UpdatedAt = &now

		if action == "submit" {
			activity.Status = DailyActivityStatusSubmitted
			activity.IsLocked = true
		}

		if err := activityRepo.Update(ctx, activity); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrActivityLocked
			}

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

func (s *DailyActivityService) FindAllPaginated(
	search string,
	page int,
	perPage int,
) ([]*entity.DailyActivity, int64, error) {
	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 10
	}

	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	return s.activityRepo.FindAllPaginated(
		context.Background(),
		strings.TrimSpace(search),
		perPage,
		offset,
	)
}
