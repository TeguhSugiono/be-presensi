package service

import (
	"context"
	"errors"
	"strings"

	"be-presensi/internal/modules/role/dto"
	"be-presensi/internal/modules/role/entity"
	"be-presensi/internal/modules/role/repository"

	"gorm.io/gorm"
)

var (
	ErrRoleNotFound = errors.New("role not found")
	ErrRoleExists   = errors.New("role already exists")
)

type RoleService interface {
	FindAll(ctx context.Context) ([]dto.RoleResponse, error)
	FindByID(ctx context.Context, id int) (*dto.RoleResponse, error)
	Create(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error)
	Update(ctx context.Context, id int, req dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	Delete(ctx context.Context, id int) error
}

type roleService struct {
	repository repository.RoleRepository
}

func NewRoleService(repository repository.RoleRepository) RoleService {
	return &roleService{
		repository: repository,
	}
}

func (s *roleService) FindAll(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.RoleResponse, 0, len(roles))

	for _, role := range roles {
		result = append(result, toResponse(role))
	}

	return result, nil
}

func (s *roleService) FindByID(ctx context.Context, id int) (*dto.RoleResponse, error) {
	role, err := s.repository.FindByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}

		return nil, err
	}

	result := toResponse(*role)

	return &result, nil
}

func (s *roleService) Create(
	ctx context.Context,
	req dto.CreateRoleRequest,
) (*dto.RoleResponse, error) {

	name := strings.TrimSpace(req.Name)

	existing, err := s.repository.FindByName(ctx, name)

	if err == nil && existing != nil {
		return nil, ErrRoleExists
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role := &entity.Role{
		Name:      name,
		IsTeacher: req.IsTeacher,
	}

	if err := s.repository.Create(ctx, role); err != nil {
		return nil, err
	}

	result := toResponse(*role)

	return &result, nil
}

func (s *roleService) Update(
	ctx context.Context,
	id int,
	req dto.UpdateRoleRequest,
) (*dto.RoleResponse, error) {

	role, err := s.repository.FindByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}

		return nil, err
	}

	name := strings.TrimSpace(req.Name)

	existing, err := s.repository.FindByName(ctx, name)

	if err == nil && existing != nil && existing.ID != id {
		return nil, ErrRoleExists
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role.Name = name
	role.IsTeacher = req.IsTeacher

	if err := s.repository.Update(ctx, role); err != nil {
		return nil, err
	}

	result := toResponse(*role)

	return &result, nil
}

func (s *roleService) Delete(ctx context.Context, id int) error {
	role, err := s.repository.FindByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}

		return err
	}

	return s.repository.Delete(ctx, role)
}

func toResponse(role entity.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID:        role.ID,
		Name:      role.Name,
		IsTeacher: role.IsTeacher,
	}
}
