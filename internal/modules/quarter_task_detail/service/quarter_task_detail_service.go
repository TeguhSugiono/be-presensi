package service

import (
	"context"
	"errors"
	"strings"
	"time"

	quarterTaskEntity "be-presensi/internal/modules/quarter_task/entity"
	quarterTaskRepository "be-presensi/internal/modules/quarter_task/repository"
	"be-presensi/internal/modules/quarter_task_detail/dto"
	"be-presensi/internal/modules/quarter_task_detail/entity"
	quarterTaskDetailRepository "be-presensi/internal/modules/quarter_task_detail/repository"

	"gorm.io/gorm"
)

var (
	ErrQuarterTaskNotFound       = errors.New("quarter task not found")
	ErrQuarterTaskDetailNotFound = errors.New("quarter task detail not found")
	ErrAchievementRequired       = errors.New("achievement text is required")
	ErrInvalidQuarterTaskAction  = errors.New("invalid quarter task action")
	ErrQuarterTaskDetailLocked   = errors.New("quarter task detail is already submitted")
)

const (
	QuarterTaskDetailStatusDraft     = "draft"
	QuarterTaskDetailStatusSubmitted = "submitted"
)

type QuarterTaskDetailService struct {
	db         *gorm.DB
	taskRepo   quarterTaskRepository.QuarterTaskRepository
	detailRepo quarterTaskDetailRepository.QuarterTaskDetailRepository
}

func NewQuarterTaskDetailService(
	db *gorm.DB,
	taskRepo quarterTaskRepository.QuarterTaskRepository,
	detailRepo quarterTaskDetailRepository.QuarterTaskDetailRepository,
) *QuarterTaskDetailService {
	return &QuarterTaskDetailService{
		db:         db,
		taskRepo:   taskRepo,
		detailRepo: detailRepo,
	}
}

// FindAllTasks mengambil seluruh Quarter Task aktif untuk pengguna.
func (s *QuarterTaskDetailService) FindAllTasks(
	ctx context.Context,
) ([]*quarterTaskEntity.QuarterTask, error) {
	return s.taskRepo.FindAll(ctx)
}

// Upsert menyimpan Draft atau melakukan Submit Quarter Task Detail.
func (s *QuarterTaskDetailService) Upsert(
	ctx context.Context,
	userID int64,
	quarterTaskID int64,
	req *dto.UpsertQuarterTaskDetailRequest,
) (*entity.QuarterTaskDetail, error) {
	action := strings.TrimSpace(req.Action)
	achievementText := strings.TrimSpace(req.AchievementText)

	if action != QuarterTaskDetailStatusDraft &&
		action != "submit" {
		return nil, ErrInvalidQuarterTaskAction
	}

	if action == "submit" && achievementText == "" {
		return nil, ErrAchievementRequired
	}

	var result *entity.QuarterTaskDetail

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txTaskRepo := s.taskRepo.WithTx(tx)
		txDetailRepo := s.detailRepo.WithTx(tx)

		// Lock header agar tidak diproses bersamaan dengan perubahan
		// atau penghapusan Quarter Task.
		task, err := txTaskRepo.FindByIDForUpdate(
			ctx,
			quarterTaskID,
		)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrQuarterTaskNotFound
			}
			return err
		}

		if task == nil {
			return ErrQuarterTaskNotFound
		}

		// Cari detail milik pengguna dan lock jika sudah tersedia.
		detail, err := txDetailRepo.FindByTaskAndUserForUpdate(
			ctx,
			quarterTaskID,
			userID,
		)

		now := time.Now()
		status := QuarterTaskDetailStatusDraft
		var submittedAt *time.Time

		if action == "submit" {
			status = QuarterTaskDetailStatusSubmitted
			submittedAt = &now
		}

		// Belum memiliki detail: buat data baru.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			detail = &entity.QuarterTaskDetail{
				QuarterTaskID:   &quarterTaskID,
				UserID:          &userID,
				AchievementText: &achievementText,
				Status:          &status,
				SubmittedAt:     submittedAt,
				CreatedBy:       &userID,
				CreatedAt:       &now,
				IsDeleted:       boolPointer(false),
			}

			// if err := txDetailRepo.Create(ctx, detail); err != nil {
			// 	return err
			// }

			// result = detail
			// return nil

			if err := txDetailRepo.Create(ctx, detail); err != nil {
				return err
			}

			// Ambil ulang data beserta relasi user dan quarter task.
			result, err = txDetailRepo.FindByTaskAndUser(
				ctx,
				quarterTaskID,
				userID,
			)
			if err != nil {
				return err
			}

			return nil
		}

		if err != nil {
			return err
		}

		if detail == nil {
			return ErrQuarterTaskDetailNotFound
		}

		// Detail yang sudah disubmit tidak boleh diedit kembali.
		if detail.Status == nil ||
			*detail.Status != QuarterTaskDetailStatusDraft {
			return ErrQuarterTaskDetailLocked
		}

		detail.AchievementText = &achievementText
		detail.Status = &status
		detail.SubmittedAt = submittedAt
		detail.UpdatedBy = &userID
		detail.UpdatedAt = &now

		if err := txDetailRepo.Update(ctx, detail); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrQuarterTaskDetailLocked
			}
			return err
		}

		// Ambil ulang data beserta relasi user dan quarter task.
		result, err = txDetailRepo.FindByTaskAndUser(
			ctx,
			quarterTaskID,
			userID,
		)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetByTask mengambil detail Quarter Task milik pengguna yang sedang login.
func (s *QuarterTaskDetailService) GetByTask(
	ctx context.Context,
	userID int64,
	quarterTaskID int64,
) (*entity.QuarterTaskDetail, error) {
	// Pastikan Quarter Task masih tersedia.
	_, err := s.taskRepo.FindByID(ctx, quarterTaskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuarterTaskNotFound
		}
		return nil, err
	}

	detail, err := s.detailRepo.FindByTaskAndUser(
		ctx,
		quarterTaskID,
		userID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuarterTaskDetailNotFound
		}
		return nil, err
	}

	return detail, nil
}

func boolPointer(value bool) *bool {
	return &value
}

func (s *QuarterTaskDetailService) FindAllSubmissionsPaginated(
	ctx context.Context,
	search string,
	year *int,
	quarterTaskID *int64,
	page int,
	perPage int,
) ([]*entity.QuarterTaskDetail, int64, error) {
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

	return s.detailRepo.FindAllSubmissionsPaginated(
		ctx,
		strings.TrimSpace(search),
		year,
		quarterTaskID,
		perPage,
		offset,
	)
}
