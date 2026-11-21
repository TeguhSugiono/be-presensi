package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"be-presensi/internal/modules/quarter_task/dto"
	"be-presensi/internal/modules/quarter_task/entity"
	quarterTaskRepository "be-presensi/internal/modules/quarter_task/repository"
	quarterTaskDetailEntity "be-presensi/internal/modules/quarter_task_detail/entity"
	quarterTaskDetailRepository "be-presensi/internal/modules/quarter_task_detail/repository"

	"gorm.io/gorm"
)

var (
	ErrQuarterTaskNotFound      = errors.New("quarter task not found")
	ErrQuarterTaskTitleRequired = errors.New("quarter task title is required")
	ErrQuarterTaskInvalidDate   = errors.New("invalid quarter task date")
	ErrQuarterTaskInvalidPeriod = errors.New("quarter task start date must be before or equal to end date")
	ErrQuarterTaskHasSubmission = errors.New("quarter task cannot be deleted because submissions already exist")
)

type QuarterTaskService struct {
	db         *gorm.DB
	taskRepo   quarterTaskRepository.QuarterTaskRepository
	detailRepo quarterTaskDetailRepository.QuarterTaskDetailRepository
}

func NewQuarterTaskService(
	db *gorm.DB,
	taskRepo quarterTaskRepository.QuarterTaskRepository,
	detailRepo quarterTaskDetailRepository.QuarterTaskDetailRepository,
) *QuarterTaskService {
	return &QuarterTaskService{
		db:         db,
		taskRepo:   taskRepo,
		detailRepo: detailRepo,
	}
}

func (s *QuarterTaskService) Create(
	ctx context.Context,
	userID int64,
	req *dto.CreateQuarterTaskRequest,
) (*entity.QuarterTask, error) {
	title := strings.TrimSpace(req.Title)

	if title == "" {
		return nil, ErrQuarterTaskTitleRequired
	}

	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.StartDate))
	if err != nil {
		return nil, ErrQuarterTaskInvalidDate
	}

	endDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.EndDate))
	if err != nil {
		return nil, ErrQuarterTaskInvalidDate
	}

	if startDate.After(endDate) {
		return nil, ErrQuarterTaskInvalidPeriod
	}

	now := time.Now()
	isDeleted := false

	task := &entity.QuarterTask{
		Title:       &title,
		Description: &req.Description,
		Year:        &req.Year,
		Quarter:     &req.Quarter,
		StartDate:   &startDate,
		EndDate:     &endDate,
		CreatedBy:   &userID,
		CreatedAt:   &now,
		IsDeleted:   &isDeleted,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *QuarterTaskService) Update(
	ctx context.Context,
	userID int64,
	id int64,
	req *dto.UpdateQuarterTaskRequest,
) (*entity.QuarterTask, error) {
	title := strings.TrimSpace(req.Title)

	if title == "" {
		return nil, ErrQuarterTaskTitleRequired
	}

	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.StartDate))
	if err != nil {
		return nil, ErrQuarterTaskInvalidDate
	}

	endDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.EndDate))
	if err != nil {
		return nil, ErrQuarterTaskInvalidDate
	}

	if startDate.After(endDate) {
		return nil, ErrQuarterTaskInvalidPeriod
	}

	var result *entity.QuarterTask

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txTaskRepo := s.taskRepo.WithTx(tx)

		task, err := txTaskRepo.FindByIDForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrQuarterTaskNotFound
			}
			return err
		}

		now := time.Now()

		task.Title = &title
		task.Description = &req.Description
		task.Year = &req.Year
		task.Quarter = &req.Quarter
		task.StartDate = &startDate
		task.EndDate = &endDate
		task.UpdatedBy = &userID
		task.UpdatedAt = &now

		if err := txTaskRepo.Update(ctx, task); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrQuarterTaskNotFound
			}
			return err
		}

		result = task
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *QuarterTaskService) Delete(
	ctx context.Context,
	userID int64,
	id int64,
) error {
	err := s.taskRepo.Delete(ctx, id, userID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrQuarterTaskNotFound
		}

		if errors.Is(err, gorm.ErrInvalidData) {
			return ErrQuarterTaskHasSubmission
		}

		return err
	}

	return nil
}

func (s *QuarterTaskService) FindByID(
	ctx context.Context,
	id int64,
) (*entity.QuarterTask, error) {
	task, err := s.taskRepo.FindByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuarterTaskNotFound
		}
		return nil, err
	}

	return task, nil
}

func (s *QuarterTaskService) FindAllPaginated(
	ctx context.Context,
	search string,
	year *int,
	quarterTaskID *int64,
	page int,
	perPage int,
) ([]*entity.QuarterTask, int64, error) {
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

	// return s.taskRepo.FindAllPaginated(
	// 	ctx,
	// 	strings.TrimSpace(search),
	// 	year,
	// 	perPage,
	// 	offset,
	// )

	return s.taskRepo.FindAllPaginated(
		ctx,
		strings.TrimSpace(search),
		year,
		quarterTaskID,
		perPage,
		offset,
	)
}

func (s *QuarterTaskService) FindAll(
	ctx context.Context,
) ([]*entity.QuarterTask, error) {
	return s.taskRepo.FindAll(ctx)
}

func (s *QuarterTaskService) FindSubmissions(
	ctx context.Context,
	quarterTaskID int64,
) ([]*quarterTaskDetailEntity.QuarterTaskDetail, error) {
	_, err := s.taskRepo.FindByID(ctx, quarterTaskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuarterTaskNotFound
		}
		return nil, err
	}

	return s.detailRepo.FindAllByTask(ctx, quarterTaskID)
}

func (s *QuarterTaskService) FindAllByYear(
	ctx context.Context,
	year int,
) ([]*entity.QuarterTask, error) {
	return s.taskRepo.FindAllByYear(ctx, year)
}
