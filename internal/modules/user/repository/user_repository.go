package repository

import (
	"be-presensi/internal/modules/user/entity"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindAll() ([]entity.User, error)
	FindByID(id int64) (*entity.User, error)
	FindByEmail(email string) (*entity.User, error)
	FindByNIP(nip string) (*entity.User, error)
	Create(user *entity.User) error
	Update(user *entity.User) error
	FindAllPaginated(
		search string,
		limit int,
		offset int,
	) ([]*entity.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) FindAll() ([]entity.User, error) {
	var users []entity.User

	err := r.db.
		Where("is_deleted = ?", false).
		Order("id ASC").
		Find(&users).Error

	return users, err
}

func (r *userRepository) FindAllPaginated(
	search string,
	limit int,
	offset int,
) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	query := r.db.Model(&entity.User{}).
		Where("is_deleted = ?", false)

	if search != "" {
		keyword := "%" + search + "%"
		query = query.Where(
			"full_name ILIKE ? OR email ILIKE ? OR nip ILIKE ?",
			keyword,
			keyword,
			keyword,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) FindByID(id int64) (*entity.User, error) {
	var user entity.User

	err := r.db.
		Where("id = ? AND is_deleted = ?", id, false).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

type userWithRole struct {
	entity.User
	RoleName *string `gorm:"column:role_name"`
}

func (r *userRepository) FindByEmail(email string) (*entity.User, error) {
	var result userWithRole

	err := r.db.
		Table("users").
		Select("users.*, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where(
			"users.email = ? AND users.is_deleted = ? AND roles.is_deleted = ?",
			email,
			false,
			false,
		).
		First(&result).Error

	if err != nil {
		return nil, err
	}

	result.User.Role = result.RoleName

	return &result.User, nil
}

func (r *userRepository) FindByNIP(nip string) (*entity.User, error) {
	var user entity.User

	err := r.db.
		Where("nip = ? AND is_deleted = ?", nip, false).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *entity.User) error {
	return r.db.Save(user).Error
}
