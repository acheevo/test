package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/acheevo/test/internal/shared/database"
	"github.com/acheevo/test/internal/user/domain"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	db *database.Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.Database) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *domain.User) error {
	return r.db.DB.Create(user).Error
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.DB.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.DB.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// GetAll retrieves all users
func (r *UserRepository) GetAll() ([]domain.User, error) {
	var users []domain.User
	err := r.db.DB.Find(&users).Error
	return users, err
}

// Update updates a user
func (r *UserRepository) Update(user *domain.User) error {
	return r.db.DB.Save(user).Error
}

// Delete deletes a user
func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.DB.Delete(&domain.User{}, id).Error
}
