package service

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"be-presensi/internal/modules/auth/dto"
	userEntity "be-presensi/internal/modules/user/entity"
	userRepository "be-presensi/internal/modules/user/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user is inactive")
	ErrUserDeleted        = errors.New("user not found")
	ErrAdminAlreadyExists = errors.New("admin already exists")
	ErrAdminRoleNotFound  = errors.New("admin role not found")
	ErrAdminEmailExists   = errors.New("email already exists")
	ErrAdminNIPExists     = errors.New("nip already exists")
)

type AuthService interface {
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
	GetCurrentUser(userID int64) (*dto.UserResponse, error)
	SetupAdmin(req dto.SetupAdminRequest) (*dto.UserResponse, error)
}

type authService struct {
	userRepo userRepository.UserRepository
	db       *gorm.DB
}

func NewAuthService(
	userRepo userRepository.UserRepository,
	db *gorm.DB,
) AuthService {
	return &authService{
		userRepo: userRepo,
		db:       db,
	}
}

func (s *authService) GetCurrentUser(userID int64) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserDeleted
		}

		return nil, err
	}

	return &dto.UserResponse{
		ID:             user.ID,
		RoleID:         user.RoleID,
		WorkScheduleID: user.WorkScheduleID,
		NIP:            user.NIP,
		FullName:       user.FullName,
		Email:          user.Email,
		PhoneNumber:    user.PhoneNumber,
		IsActive:       user.IsActive,
	}, nil
}

func (s *authService) Login(
	req dto.LoginRequest,
) (*dto.LoginResponse, error) {

	email := strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.findUserForLogin(email)
	if err != nil {
		return nil, err
	}

	if user.IsActive == false {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, expiresIn, err := generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	role := ""
	if user.Role != nil {
		role = *user.Role
	}

	return &dto.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User: dto.UserResponse{
			ID:             user.ID,
			RoleID:         user.RoleID,
			Role:           role,
			WorkScheduleID: user.WorkScheduleID,
			NIP:            user.NIP,
			FullName:       user.FullName,
			Email:          user.Email,
			PhoneNumber:    user.PhoneNumber,
			IsActive:       user.IsActive,
		},
	}, nil
}

func (s *authService) findUserForLogin(email string) (*userEntity.User, error) {
	user, err := s.userRepo.FindByEmail(email)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserDeleted
		}

		return nil, err
	}

	return user, nil
}

func generateAccessToken(userID int64) (string, int64, error) {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		return "", 0, errors.New("JWT_SECRET is not configured")
	}

	expirationHours := int64(24)

	if value := os.Getenv("JWT_EXPIRATION_HOURS"); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil && parsed > 0 {
			expirationHours = parsed
		}
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(expirationHours) * time.Hour)

	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}

	return signedToken, int64(time.Until(expiresAt).Seconds()), nil
}

func (s *authService) SetupAdmin(req dto.SetupAdminRequest) (*dto.UserResponse, error) {
	var adminRoleID int

	err := s.db.
		Table("roles").
		Where("LOWER(name) = ? AND is_deleted = ?", "admin", false).
		Pluck("id", &adminRoleID).Error

	if err != nil {
		return nil, err
	}

	if adminRoleID == 0 {
		return nil, ErrAdminRoleNotFound
	}

	var adminCount int64

	err = s.db.
		Table("users").
		Where("role_id = ? AND is_deleted = ?", adminRoleID, false).
		Count(&adminCount).Error

	if err != nil {
		return nil, err
	}

	if adminCount > 0 {
		return nil, ErrAdminAlreadyExists
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	_, err = s.userRepo.FindByEmail(email)
	if err == nil {
		return nil, ErrAdminEmailExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var nip *string
	if req.NIP != nil {
		value := strings.TrimSpace(*req.NIP)
		if value != "" {
			nip = &value

			_, err = s.userRepo.FindByNIP(value)
			if err == nil {
				return nil, ErrAdminNIPExists
			}

			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
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

	user := &userEntity.User{
		RoleID:       adminRoleID,
		NIP:          nip,
		FullName:     strings.TrimSpace(req.FullName),
		Email:        email,
		PasswordHash: string(passwordHash),
		PhoneNumber:  strings.TrimSpace(req.PhoneNumber),
		IsActive:     true,
		CreatedBy:    nil,
		CreatedAt:    now,
		IsDeleted:    false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:             user.ID,
		RoleID:         user.RoleID,
		WorkScheduleID: user.WorkScheduleID,
		NIP:            user.NIP,
		FullName:       user.FullName,
		Email:          user.Email,
		PhoneNumber:    user.PhoneNumber,
		IsActive:       user.IsActive,
	}, nil
}
