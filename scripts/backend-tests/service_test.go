package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"hub-hrms/backend/internal/config"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)

// Mock Repositories

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, search string, role string) ([]*models.User, error) {
	args := m.Called(ctx, search, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockEmployeeRepository struct {
	mock.Mock
}

func (m *MockEmployeeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) Create(ctx context.Context, emp *models.Employee) error {
	args := m.Called(ctx, emp)
	return args.Error(0)
}

func (m *MockEmployeeRepository) Update(ctx context.Context, emp *models.Employee) error {
	args := m.Called(ctx, emp)
	return args.Error(0)
}

func (m *MockEmployeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEmployeeRepository) List(ctx context.Context) ([]*models.Employee, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Employee), args.Error(1)
}

// UserService Tests

func TestUserService_Create(t *testing.T) {
	mockRepo := new(MockUserRepository)
	repos := &repository.Repositories{User: mockRepo}
	service := NewUserService(repos)

	ctx := context.Background()
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "employee",
	}

	mockRepo.On("Create", ctx, user).Return(nil)

	err := service.Create(ctx, user)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		mockUser      *models.User
		mockError     error
		expectedError error
	}{
		{
			name:  "successfully retrieves user",
			email: "test@example.com",
			mockUser: &models.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
				Role:     "employee",
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "user not found",
			email:         "notfound@example.com",
			mockUser:      nil,
			mockError:     errors.New("not found"),
			expectedError: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			repos := &repository.Repositories{User: mockRepo}
			service := NewUserService(repos)

			ctx := context.Background()
			mockRepo.On("GetByEmail", ctx, tt.email).Return(tt.mockUser, tt.mockError)

			user, err := service.GetByEmail(ctx, tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	repos := &repository.Repositories{User: mockRepo}
	service := NewUserService(repos)

	ctx := context.Background()
	userID := uuid.New()
	expectedUser := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		Username: "testuser",
	}

	mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	user, err := service.GetByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Update(t *testing.T) {
	mockRepo := new(MockUserRepository)
	repos := &repository.Repositories{User: mockRepo}
	service := NewUserService(repos)

	ctx := context.Background()
	user := &models.User{
		ID:       uuid.New(),
		Email:    "updated@example.com",
		Username: "updateduser",
		Role:     "manager",
	}

	mockRepo.On("Update", ctx, user).Return(nil)

	err := service.Update(ctx, user)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_List(t *testing.T) {
	mockRepo := new(MockUserRepository)
	repos := &repository.Repositories{User: mockRepo}
	service := NewUserService(repos)

	ctx := context.Background()
	search := "test"
	role := "employee"
	expectedUsers := []*models.User{
		{ID: uuid.New(), Email: "test1@example.com", Username: "test1", Role: "employee"},
		{ID: uuid.New(), Email: "test2@example.com", Username: "test2", Role: "employee"},
	}

	mockRepo.On("List", ctx, search, role).Return(expectedUsers, nil)

	users, err := service.List(ctx, search, role)

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Delete(t *testing.T) {
	mockRepo := new(MockUserRepository)
	repos := &repository.Repositories{User: mockRepo}
	service := NewUserService(repos)

	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("Delete", ctx, userID).Return(nil)

	err := service.Delete(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// AuthService Tests

func TestAuthService_HashPassword(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	service := NewAuthService(&repository.Repositories{}, cfg)

	password := "testpassword123"
	hash, err := service.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Verify hash is valid bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	assert.NoError(t, err)
}

func TestAuthService_CheckPassword(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	service := NewAuthService(&repository.Repositories{}, cfg)

	password := "testpassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		hashedPassword string
		password      string
		expectError   bool
	}{
		{
			name:           "correct password",
			hashedPassword: string(hash),
			password:       password,
			expectError:    false,
		},
		{
			name:           "incorrect password",
			hashedPassword: string(hash),
			password:       "wrongpassword",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CheckPassword(tt.hashedPassword, tt.password)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	service := NewAuthService(&repository.Repositories{}, cfg)

	userID := uuid.New()
	employeeID := uuid.New()
	email := "test@example.com"
	role := "employee"

	tests := []struct {
		name       string
		employeeID *uuid.UUID
	}{
		{
			name:       "with employee ID",
			employeeID: &employeeID,
		},
		{
			name:       "without employee ID",
			employeeID: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(userID, email, role, tt.employeeID)

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Validate token structure
			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.JWTSecret), nil
			})

			assert.NoError(t, err)
			assert.True(t, parsedToken.Valid)

			claims := parsedToken.Claims.(jwt.MapClaims)
			assert.Equal(t, userID.String(), claims["user_id"])
			assert.Equal(t, email, claims["email"])
			assert.Equal(t, role, claims["role"])

			if tt.employeeID != nil {
				assert.Equal(t, tt.employeeID.String(), claims["employee_id"])
			} else {
				_, hasEmployeeID := claims["employee_id"]
				assert.False(t, hasEmployeeID)
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	service := NewAuthService(&repository.Repositories{}, cfg)

	userID := uuid.New()
	email := "test@example.com"
	role := "employee"

	tests := []struct {
		name        string
		setupToken  func() string
		expectError bool
	}{
		{
			name: "valid token",
			setupToken: func() string {
				token, _ := service.GenerateToken(userID, email, role, nil)
				return token
			},
			expectError: false,
		},
		{
			name: "expired token",
			setupToken: func() string {
				claims := jwt.MapClaims{
					"user_id": userID.String(),
					"email":   email,
					"role":    role,
					"exp":     time.Now().Add(-1 * time.Hour).Unix(),
					"iat":     time.Now().Add(-2 * time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
				return tokenString
			},
			expectError: true,
		},
		{
			name: "invalid signature",
			setupToken: func() string {
				claims := jwt.MapClaims{
					"user_id": userID.String(),
					"email":   email,
					"role":    role,
					"exp":     time.Now().Add(24 * time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("wrong-secret"))
				return tokenString
			},
			expectError: true,
		},
		{
			name: "malformed token",
			setupToken: func() string {
				return "not.a.valid.token"
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.setupToken()
			token, err := service.ValidateToken(tokenString)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.True(t, token.Valid)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}

	userID := uuid.New()
	employeeID := uuid.New()
	email := "test@example.com"
	password := "testpassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		request       *models.LoginRequest
		setupMocks    func(*MockUserRepository, *MockEmployeeRepository)
		expectError   bool
		expectedError error
	}{
		{
			name: "successful login with employee",
			request: &models.LoginRequest{
				Email:    email,
				Password: password,
			},
			setupMocks: func(userRepo *MockUserRepository, empRepo *MockEmployeeRepository) {
				userRepo.On("GetByEmail", mock.Anything, email).Return(&models.User{
					ID:           userID,
					Email:        email,
					PasswordHash: string(hashedPassword),
					Role:         "employee",
					EmployeeID:   &employeeID,
				}, nil)
				empRepo.On("GetByID", mock.Anything, employeeID).Return(&models.Employee{
					ID:        employeeID,
					FirstName: "John",
					LastName:  "Doe",
					Email:     email,
				}, nil)
			},
			expectError: false,
		},
		{
			name: "successful login without employee",
			request: &models.LoginRequest{
				Email:    email,
				Password: password,
			},
			setupMocks: func(userRepo *MockUserRepository, empRepo *MockEmployeeRepository) {
				userRepo.On("GetByEmail", mock.Anything, email).Return(&models.User{
					ID:           userID,
					Email:        email,
					PasswordHash: string(hashedPassword),
					Role:         "admin",
					EmployeeID:   nil,
				}, nil)
			},
			expectError: false,
		},
		{
			name: "user not found",
			request: &models.LoginRequest{
				Email:    "notfound@example.com",
				Password: password,
			},
			setupMocks: func(userRepo *MockUserRepository, empRepo *MockEmployeeRepository) {
				userRepo.On("GetByEmail", mock.Anything, "notfound@example.com").
					Return(nil, errors.New("not found"))
			},
			expectError:   true,
			expectedError: ErrInvalidCredentials,
		},
		{
			name: "incorrect password",
			request: &models.LoginRequest{
				Email:    email,
				Password: "wrongpassword",
			},
			setupMocks: func(userRepo *MockUserRepository, empRepo *MockEmployeeRepository) {
				userRepo.On("GetByEmail", mock.Anything, email).Return(&models.User{
					ID:           userID,
					Email:        email,
					PasswordHash: string(hashedPassword),
					Role:         "employee",
				}, nil)
			},
			expectError:   true,
			expectedError: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(MockUserRepository)
			mockEmpRepo := new(MockEmployeeRepository)
			repos := &repository.Repositories{
				User:     mockUserRepo,
				Employee: mockEmpRepo,
			}

			tt.setupMocks(mockUserRepo, mockEmpRepo)

			service := NewAuthService(repos, cfg)
			response, err := service.Login(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, email, response.User.Email)
			}

			mockUserRepo.AssertExpectations(t)
			mockEmpRepo.AssertExpectations(t)
		})
	}
}

// Benchmark Tests

func BenchmarkAuthService_HashPassword(b *testing.B) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	service := NewAuthService(&repository.Repositories{}, cfg)
	password := "testpassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.HashPassword(password)
	}
}

func BenchmarkAuthService_CheckPassword(b *testing.B) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	service := NewAuthService(&repository.Repositories{}, cfg)
	password := "testpassword123"
	hash, _ := service.HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.CheckPassword(hash, password)
	}
}

func BenchmarkAuthService_GenerateToken(b *testing.B) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	service := NewAuthService(&repository.Repositories{}, cfg)
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateToken(userID, "test@example.com", "employee", nil)
	}
}
