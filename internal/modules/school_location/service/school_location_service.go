package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"be-presensi/internal/modules/school_location/dto"
	"be-presensi/internal/modules/school_location/entity"
	"be-presensi/internal/modules/school_location/repository"

	"gorm.io/gorm"
)

var (
	ErrSchoolLocationNotFound = errors.New("school location not found")
	ErrSchoolLocationExists   = errors.New("school location already exists")
	ErrInvalidLatitude        = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude       = errors.New("longitude must be between -180 and 180")
	ErrInvalidRadius          = errors.New("radius must be greater than zero")
	ErrInvalidLocationName    = errors.New("location name is required")
)

type SchoolLocationService interface {
	FindAll(ctx context.Context) ([]dto.SchoolLocationResponse, error)
	FindByID(ctx context.Context, id int) (*dto.SchoolLocationResponse, error)
	Create(
		ctx context.Context,
		req *dto.CreateSchoolLocationRequest,
		userID *int64,
	) (*dto.SchoolLocationResponse, error)
	Update(
		ctx context.Context,
		id int,
		req *dto.UpdateSchoolLocationRequest,
		userID *int64,
	) (*dto.SchoolLocationResponse, error)
	Delete(ctx context.Context, id int, userID *int64) error
	FindAllPaginated(
		ctx context.Context,
		search string,
		page int,
		perPage int,
	) ([]dto.SchoolLocationResponse, int64, error)
}

type schoolLocationService struct {
	repository repository.SchoolLocationRepository
}

func NewSchoolLocationService(
	repository repository.SchoolLocationRepository,
) SchoolLocationService {
	return &schoolLocationService{
		repository: repository,
	}
}

func (s *schoolLocationService) FindAll(
	ctx context.Context,
) ([]dto.SchoolLocationResponse, error) {
	locations, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.SchoolLocationResponse, 0, len(locations))

	for _, location := range locations {
		result = append(result, toResponse(location))
	}

	return result, nil
}

func (s *schoolLocationService) FindAllPaginated(
	ctx context.Context,
	search string,
	page int,
	perPage int,
) ([]dto.SchoolLocationResponse, int64, error) {
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

	locations, total, err := s.repository.FindAllPaginated(
		ctx,
		strings.TrimSpace(search),
		perPage,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.SchoolLocationResponse, 0, len(locations))

	for _, location := range locations {
		result = append(result, toResponse(location))
	}

	return result, total, nil
}

func (s *schoolLocationService) FindByID(
	ctx context.Context,
	id int,
) (*dto.SchoolLocationResponse, error) {
	location, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSchoolLocationNotFound
		}
		return nil, err
	}

	result := toResponse(*location)
	return &result, nil
}

func (s *schoolLocationService) Create(
	ctx context.Context,
	req *dto.CreateSchoolLocationRequest,
	userID *int64,
) (*dto.SchoolLocationResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidLocationName
	}

	if err := validateCoordinates(req.Latitude, req.Longitude); err != nil {
		return nil, err
	}

	if req.RadiusMeters <= 0 {
		return nil, ErrInvalidRadius
	}

	existing, err := s.repository.FindByName(ctx, name)
	if err == nil && existing != nil {
		return nil, ErrSchoolLocationExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now()

	location := &entity.SchoolLocation{
		Name:         name,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		RadiusMeters: req.RadiusMeters,
		IsActive:     req.IsActive,
		CreatedBy:    userID,
		CreatedAt:    now,
		IsDeleted:    false,
	}

	if err := s.repository.Create(ctx, location); err != nil {
		return nil, err
	}

	result := toResponse(*location)
	return &result, nil
}

func (s *schoolLocationService) Update(
	ctx context.Context,
	id int,
	req *dto.UpdateSchoolLocationRequest,
	userID *int64,
) (*dto.SchoolLocationResponse, error) {
	location, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSchoolLocationNotFound
		}
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidLocationName
	}

	if err := validateCoordinates(req.Latitude, req.Longitude); err != nil {
		return nil, err
	}

	if req.RadiusMeters <= 0 {
		return nil, ErrInvalidRadius
	}

	existing, err := s.repository.FindByName(ctx, name)
	if err == nil && existing != nil && existing.ID != id {
		return nil, ErrSchoolLocationExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	location.Name = name
	location.Latitude = req.Latitude
	location.Longitude = req.Longitude
	location.RadiusMeters = req.RadiusMeters
	location.IsActive = req.IsActive

	// CreatedBy dan CreatedAt tetap dipertahankan.
	location.UpdatedBy = userID

	now := time.Now()
	location.UpdatedAt = &now

	if err := s.repository.Update(ctx, location); err != nil {
		return nil, err
	}

	result := toResponse(*location)
	return &result, nil
}

func (s *schoolLocationService) Delete(
	ctx context.Context,
	id int,
	userID *int64,
) error {
	location, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSchoolLocationNotFound
		}
		return err
	}

	location.IsDeleted = true
	location.UpdatedBy = userID

	now := time.Now()
	location.UpdatedAt = &now

	return s.repository.Delete(ctx, location)
}

func validateCoordinates(latitude, longitude float64) error {
	if latitude < -90 || latitude > 90 {
		return ErrInvalidLatitude
	}

	if longitude < -180 || longitude > 180 {
		return ErrInvalidLongitude
	}

	return nil
}

func toResponse(
	location entity.SchoolLocation,
) dto.SchoolLocationResponse {
	return dto.SchoolLocationResponse{
		ID:           location.ID,
		Name:         location.Name,
		Latitude:     location.Latitude,
		Longitude:    location.Longitude,
		RadiusMeters: location.RadiusMeters,
		IsActive:     location.IsActive,
	}
}
