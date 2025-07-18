package service

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/acheevo/test/internal/user/domain"
	"github.com/acheevo/test/internal/user/repository"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(id)
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(email)
}

// GetAll retrieves all users
func (s *UserService) GetAll() ([]domain.User, error) {
	return s.userRepo.GetAll()
}

// Create creates a new user
func (s *UserService) Create(email, password, name string, role domain.UserRole) (*domain.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:       uuid.New(),
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
		Role:     role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Update updates a user
func (s *UserService) Update(user *domain.User) error {
	return s.userRepo.Update(user)
}

// Delete deletes a user
func (s *UserService) Delete(id uuid.UUID) error {
	return s.userRepo.Delete(id)
}
