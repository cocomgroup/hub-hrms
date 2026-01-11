package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"hub-hrms/backend/internal/config"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
	"hub-hrms/backend/tests/fixtures"
	"hub-hrms/backend/tests/helpers"
	"hub-hrms/backend/tests/mocks"
)

func TestAuthService_Login_Success(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	repos, mockRepos := helpers.NewMockRepositories()
	testUser := fixtures.NewUser()

	mockRepos.User.On("GetByEmail", mock.Anything, testUser.Email).Return(testUser, nil)
	mockRepos.Employee.On("GetByEmail", mock.Anything, testUser.Email).Return(fixtures.NewEmployee(), nil)

	authService := service.NewAuthService(repos, cfg)

	request := &models.LoginRequest{
		Email:    testUser.Email,
		Password: "password123", // matches fixture password
	}

	// Act
	response, err := authService.Login(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, testUser.Email, response.User.Email)

	mockRepos.User.AssertExpectations(t)
	mockRepos.Employee.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	repos, mockRepos := helpers.NewMockRepositories()
	testUser := fixtures.NewUser()

	mockRepos.User.On("GetByEmail", mock.Anything, testUser.Email).Return(testUser, nil)

	authService := service.NewAuthService(repos, cfg)

	request := &models.LoginRequest{
		Email:    testUser.Email,
		Password: "wrong-password",
	}

	// Act
	response, err := authService.Login(context.Background(), request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid credentials")

	mockRepos.User.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	repos, mockRepos := helpers.NewMockRepositories()

	mockRepos.User.On("GetByEmail", mock.Anything, "nonexistent@example.com").
		Return(nil, errors.New("user not found"))

	authService := service.NewAuthService(repos, cfg)

	request := &models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	// Act
	response, err := authService.Login(context.Background(), request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)

	mockRepos.User.AssertExpectations(t)
}

func TestAuthService_Register_Success(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	repos, mockRepos := helpers.NewMockRepositories()

	mockRepos.User.On("GetByEmail", mock.Anything, "newuser@example.com").
		Return(nil, errors.New("user not found"))
	mockRepos.User.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
		Return(nil)

	authService := service.NewAuthService(repos, cfg)

	request := &models.RegisterRequest{
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	// Act
	user, err := authService.Register(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newuser@example.com", user.Email)
	assert.NotEmpty(t, user.ID)

	mockRepos.User.AssertExpectations(t)
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	repos, mockRepos := helpers.NewMockRepositories()
	existingUser := fixtures.NewUser()

	mockRepos.User.On("GetByEmail", mock.Anything, existingUser.Email).
		Return(existingUser, nil)

	authService := service.NewAuthService(repos, cfg)

	request := &models.RegisterRequest{
		Email:     existingUser.Email,
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	// Act
	user, err := authService.Register(context.Background(), request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "already exists")

	mockRepos.User.AssertExpectations(t)
}
