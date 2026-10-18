package service

import (
	"errors"
	"strings"
	"time"

	"be-presensi/internal/modules/user/dto"
	"be-presensi/internal/modules/user/entity"
	"be-presensi/internal/modules/user/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrNIPAlreadyExists   = errors.New("nip already exists")
	ErrRoleNotFound       = errors.New("role not found or inactive")
	ErrScheduleNotFound   = errors.New("work schedule not found or inactive")
	ErrAdminRequired      = errors.New("admin access required")
)

type UserService interface {
	FindAll() ([]dto.UserResponse, error)
	FindByID(id int64) (*dto.UserResponse, error)
	Create(req dto.CreateUserRequest, userID *int64) (*dto.UserResponse, error)
	Update(id int64, req dto.UpdateUserRequest, userID *int64) (*dto.UserResponse, error)
	Delete(id int64, userID *int64) error
	FindAllPaginated(
		search string,
		page int,
		perPage int,
	) ([]*dto.UserResponse, int64, error)

	ChangePassword(
		id int64,
		newPassword string,
		userID *int64,
	) error
}

type userService struct {
	repo repository.UserRepository
	db   *gorm.DB
}

func NewUserService(
	repo repository.UserRepository,
	db *gorm.DB,
) UserService {
	return &userService{
		repo: repo,
		db:   db,
	}
}

func (s *userService) FindAll() ([]dto.UserResponse, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	result := make([]dto.UserResponse, 0, len(users))

	for _, user := range users {
		result = append(result, toUserResponse(&user))
	}

	return result, nil
}

func (s *userService) FindAllPaginated(
	search string,
	page int,
	perPage int,
) ([]*dto.UserResponse, int64, error) {
	offset := (page - 1) * perPage

	users, total, err := s.repo.FindAllPaginated(
		search,
		perPage,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*dto.UserResponse, 0, len(users))

	for _, user := range users {
		response := toUserResponse(user)
		result = append(result, &response)
	}

	return result, total, nil
}

func (s *userService) FindByID(id int64) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	result := toUserResponse(user)

	return &result, nil
}

func (s *userService) Create(
	req dto.CreateUserRequest,
	userID *int64,
) (*dto.UserResponse, error) {

	email := strings.TrimSpace(strings.ToLower(req.Email))

	if err := s.validateRole(req.RoleID); err != nil {
		return nil, err
	}

	if err := s.validateWorkSchedule(req.WorkScheduleID); err != nil {
		return nil, err
	}

	existingEmail, err := s.repo.FindByEmail(email)

	if err == nil && existingEmail != nil {
		return nil, ErrEmailAlreadyExists
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if req.NIP != nil {
		nip := strings.TrimSpace(*req.NIP)

		if nip != "" {
			existingNIP, err := s.repo.FindByNIP(nip)

			if err == nil && existingNIP != nil {
				return nil, ErrNIPAlreadyExists
			}

			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}

			req.NIP = &nip
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	user := &entity.User{
		RoleID:         req.RoleID,
		WorkScheduleID: req.WorkScheduleID,
		NIP:            req.NIP,
		FullName:       strings.TrimSpace(req.FullName),
		Email:          email,
		PasswordHash:   string(passwordHash),
		PhoneNumber:    strings.TrimSpace(req.PhoneNumber),
		IsActive:       req.IsActive,

		CreatedBy: userID,
		CreatedAt: now,
		IsDeleted: false,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	result := toUserResponse(user)

	return &result, nil
}

func (s *userService) Update(
	id int64,
	req dto.UpdateUserRequest,
	userID *int64,
) (*dto.UserResponse, error) {

	user, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	if err := s.validateRole(req.RoleID); err != nil {
		return nil, err
	}

	if err := s.validateWorkSchedule(req.WorkScheduleID); err != nil {
		return nil, err
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	existingEmail, err := s.repo.FindByEmail(email)

	if err == nil && existingEmail != nil && existingEmail.ID != id {
		return nil, ErrEmailAlreadyExists
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if req.NIP != nil {
		nip := strings.TrimSpace(*req.NIP)

		if nip != "" {
			existingNIP, err := s.repo.FindByNIP(nip)

			if err == nil && existingNIP != nil && existingNIP.ID != id {
				return nil, ErrNIPAlreadyExists
			}

			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}

			req.NIP = &nip
		}
	}

	user.RoleID = req.RoleID
	user.WorkScheduleID = req.WorkScheduleID
	user.NIP = req.NIP
	user.FullName = strings.TrimSpace(req.FullName)
	user.Email = email
	user.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	user.IsActive = req.IsActive

	user.UpdatedBy = userID
	now := time.Now()
	user.UpdatedAt = &now

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	result := toUserResponse(user)

	return &result, nil
}

func (s *userService) Delete(id int64, userID *int64) error {
	user, err := s.repo.FindByID(id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}

		return err
	}

	user.IsDeleted = true
	user.IsActive = false
	user.UpdatedBy = userID
	now := time.Now()
	user.UpdatedAt = &now

	return s.repo.Update(user)
}

func (s *userService) ChangePassword(
	id int64,
	newPassword string,
	userID *int64,
) error {

	if userID == nil || *userID <= 0 {
		return ErrAdminRequired
	}

	var count int64

	err := s.db.
		Table("users").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where(
			"users.id = ? AND users.is_active = ? AND users.is_deleted = ? AND roles.name = ? AND roles.is_deleted = ?",
			*userID,
			true,
			false,
			"Admin",
			false,
		).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count == 0 {
		return ErrAdminRequired
	}

	user, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}

		return err
	}

	newPassword = strings.TrimSpace(newPassword)
	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	now := time.Now()

	user.PasswordHash = string(passwordHash)
	user.UpdatedBy = userID
	user.UpdatedAt = &now

	return s.repo.Update(user)
}

func (s *userService) validateRole(roleID int) error {
	var count int64

	err := s.db.
		Table("roles").
		Where("id = ? AND is_deleted = ?", roleID, false).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count == 0 {
		return ErrRoleNotFound
	}

	return nil
}

func (s *userService) validateWorkSchedule(scheduleID *int) error {
	if scheduleID == nil {
		return nil
	}

	var count int64

	err := s.db.
		Table("work_schedules").
		Where("id = ? AND is_deleted = ?", *scheduleID, false).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count == 0 {
		return ErrScheduleNotFound
	}

	return nil
}

func toUserResponse(user *entity.User) dto.UserResponse {
	role := ""
	if user.Role != nil {
		role = *user.Role
	}

	return dto.UserResponse{
		ID:             user.ID,
		RoleID:         user.RoleID,
		Role:           role,
		WorkScheduleID: user.WorkScheduleID,
		NIP:            user.NIP,
		FullName:       user.FullName,
		Email:          user.Email,
		PhoneNumber:    user.PhoneNumber,
		IsActive:       user.IsActive,
	}
}
