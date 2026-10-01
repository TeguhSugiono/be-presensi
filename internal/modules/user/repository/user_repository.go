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

func (r *userRepository) FindByEmail(email string) (*entity.User, error) {
	var user entity.User

	err := r.db.
		Where("email = ? AND is_deleted = ?", email, false).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
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
