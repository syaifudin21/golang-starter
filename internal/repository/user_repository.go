package repository

import (
	"exam/internal/model"

	"gorm.io/gorm"
)

// UserRepository defines methods for interacting with user data in the database.
type UserRepository interface {
	CreateUser(user *model.User) error
	GetUserByEmail(email string) (*model.User, error)
	GetUserByID(userID uint) (*model.User, error)
	GetUserByUUID(userUUID string) (*model.User, error)
	UpdateUser(user *model.User) error
	ListAllUsers() ([]model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found, return nil user and nil error
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found, return nil user and nil error
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByUUID(userUUID string) (*model.User, error) {
	var user model.User
	err := r.db.Where("uuid = ?", userUUID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found, return nil user and nil error
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) ListAllUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.Find(&users).Error
	return users, err
}
