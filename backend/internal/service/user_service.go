package service

import (
	"context"
	"fmt"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"

	"github.com/google/uuid"
)

// UserService handles user management operations
type UserService interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	List(ctx context.Context, filters map[string]interface{}) ([]*models.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ResetPassword(ctx context.Context, userID uuid.UUID, newPassword string) error
}

type userService struct {
	repos *repository.Repositories
}

func NewUserService(repos *repository.Repositories) UserService {
	return &userService{repos: repos}
}

// Create creates a new user
func (s *userService) Create(ctx context.Context, user *models.User) error {
	// Check if user with email already exists
	existing, err := s.repos.User.GetByEmail(ctx, user.Email)
	if err == nil && existing != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	return s.repos.User.Create(ctx, user)
}

// GetByID retrieves a user by ID
func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repos.User.GetByID(ctx, id)
}

// GetByEmail retrieves a user by email
func (s *userService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repos.User.GetByEmail(ctx, email)
}

// Update updates a user
func (s *userService) Update(ctx context.Context, user *models.User) error {
	// Verify user exists
	existing, err := s.repos.User.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// If email changed, check for duplicates
	if existing.Email != user.Email {
		duplicate, err := s.repos.User.GetByEmail(ctx, user.Email)
		if err == nil && duplicate != nil && duplicate.ID != user.ID {
			return fmt.Errorf("email %s is already in use", user.Email)
		}
	}

	return s.repos.User.Update(ctx, user)
}

// List retrieves all users with optional filters
func (s *userService) List(ctx context.Context, filters map[string]interface{}) ([]*models.User, error) {
	// TODO: Implement List in repository
	return []*models.User{}, nil
}

// Delete deletes a user
func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement Delete in repository
	return fmt.Errorf("delete not yet implemented")
}

// ResetPassword resets a user's password
func (s *userService) ResetPassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error {
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.PasswordHash = newPasswordHash
	return s.repos.User.Update(ctx, user)
}