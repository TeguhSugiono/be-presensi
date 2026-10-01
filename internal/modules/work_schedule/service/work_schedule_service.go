package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"be-presensi/internal/modules/work_schedule/dto"
	"be-presensi/internal/modules/work_schedule/entity"
	"be-presensi/internal/modules/work_schedule/repository"

	"gorm.io/gorm"
)

var (
	ErrWorkScheduleNotFound = errors.New("work schedule not found")
	ErrWorkScheduleExists   = errors.New("work schedule already exists")
	ErrInvalidStartTime     = errors.New("invalid start time")
	ErrInvalidEndTime       = errors.New("invalid end time")
	ErrInvalidTimeRange     = errors.New("start time must be before end time")
)

type WorkScheduleService interface {
	FindAll(ctx context.Context) ([]*dto.WorkScheduleResponse, error)
	FindByID(ctx context.Context, id int) (*dto.WorkScheduleResponse, error)
	Create(ctx context.Context, req *dto.CreateWorkScheduleRequest, userID *int64) (*dto.WorkScheduleResponse, error)
	Update(ctx context.Context, id int, req *dto.UpdateWorkScheduleRequest, userID *int64) (*dto.WorkScheduleResponse, error)
	Delete(ctx context.Context, id int, userID *int64) error
}

type workScheduleService struct {
	repository repository.WorkScheduleRepository
}

func NewWorkScheduleService(
	repository repository.WorkScheduleRepository,
) WorkScheduleService {
	return &workScheduleService{
		repository: repository,
	}
}

func (s *workScheduleService) FindAll(
	ctx context.Context,
) ([]*dto.WorkScheduleResponse, error) {

	schedules, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.WorkScheduleResponse, 0, len(schedules))

	for _, schedule := range schedules {
		result = append(result, toResponse(schedule))
	}

	return result, nil
}

func (s *workScheduleService) FindByID(
	ctx context.Context,
	id int,
) (*dto.WorkScheduleResponse, error) {

	schedule, err := s.repository.FindByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkScheduleNotFound
		}

		return nil, err
	}

	result := toResponse(*schedule)

	return result, nil
}

func (s *workScheduleService) Create(
	ctx context.Context,
	req *dto.CreateWorkScheduleRequest,
	userID *int64,
) (*dto.WorkScheduleResponse, error) {

	name := strings.TrimSpace(req.Name)

	existing, err := s.repository.FindByName(ctx, name)

	if err == nil && existing != nil {
		return nil, ErrWorkScheduleExists
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if !validateTime(req.StartTime) {
		return nil, ErrInvalidStartTime
	}

	if !validateTime(req.EndTime) {
		return nil, ErrInvalidEndTime
	}

	if !validateTimeRange(req.StartTime, req.EndTime) {
		return nil, ErrInvalidTimeRange
	}

	now := time.Now()

	schedule := &entity.WorkSchedule{
		Name:       name,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		IsFlexible: req.IsFlexible,
		CreatedBy:  userID,
		CreatedAt:  now,
		IsDeleted:  false,
	}

	if err := s.repository.Create(ctx, schedule); err != nil {
		return nil, err
	}

	result := toResponse(*schedule)

	return result, nil
}

func (s *workScheduleService) Update(
	ctx context.Context,
	id int,
	req *dto.UpdateWorkScheduleRequest,
	userID *int64,
) (*dto.WorkScheduleResponse, error) {

	schedule, err := s.repository.FindByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkScheduleNotFound
		}

		return nil, err
	}

	name := strings.TrimSpace(req.Name)

	existing, err := s.repository.FindByName(ctx, name)

	if err == nil && existing != nil && existing.ID != id {
		return nil, ErrWorkScheduleExists
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if !validateTime(req.StartTime) {
		return nil, ErrInvalidStartTime
	}

	if !validateTime(req.EndTime) {
		return nil, ErrInvalidEndTime
	}

	if !validateTimeRange(req.StartTime, req.EndTime) {
		return nil, ErrInvalidTimeRange
	}

	schedule.Name = name
	schedule.StartTime = req.StartTime
	schedule.EndTime = req.EndTime
	schedule.IsFlexible = req.IsFlexible

	// CreatedBy dan CreatedAt tetap.
	schedule.UpdatedBy = userID
	now := time.Now()
	schedule.UpdatedAt = &now

	if err := s.repository.Update(ctx, schedule); err != nil {
		return nil, err
	}

	result := toResponse(*schedule)

	return result, nil
}

func (s *workScheduleService) Delete(
	ctx context.Context,
	id int,
	userID *int64,
) error {

	schedule, err := s.repository.FindByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkScheduleNotFound
		}

		return err
	}

	schedule.IsDeleted = true
	schedule.UpdatedBy = userID
	now := time.Now()
	schedule.UpdatedAt = &now

	return s.repository.Delete(ctx, schedule)
}

func validateTime(value *string) bool {
	if value == nil || strings.TrimSpace(*value) == "" {
		return true
	}

	_, err := time.Parse(
		"15:04",
		strings.TrimSpace(*value),
	)

	return err == nil
}

func validateTimeRange(startTime *string, endTime *string) bool {
	if startTime == nil || endTime == nil {
		return true
	}

	start := strings.TrimSpace(*startTime)
	end := strings.TrimSpace(*endTime)

	if start == "" || end == "" {
		return true
	}

	startParsed, err := time.Parse("15:04", start)
	if err != nil {
		return false
	}

	endParsed, err := time.Parse("15:04", end)
	if err != nil {
		return false
	}

	return startParsed.Before(endParsed)
}

func toResponse(
	schedule entity.WorkSchedule,
) *dto.WorkScheduleResponse {

	return &dto.WorkScheduleResponse{
		ID:         schedule.ID,
		Name:       schedule.Name,
		StartTime:  schedule.StartTime,
		EndTime:    schedule.EndTime,
		IsFlexible: schedule.IsFlexible,
	}
}
