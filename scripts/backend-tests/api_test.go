package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"hub-hrms/backend/internal/config"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// Mock Services

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) CheckPassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

func (m *MockAuthService) GenerateToken(userID uuid.UUID, email, role string, employeeID *uuid.UUID) (string, error) {
	args := m.Called(userID, email, role, employeeID)
	return args.String(0), args.Error(1)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) List(ctx context.Context, search string, role string) ([]*models.User, error) {
	args := m.Called(ctx, search, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Test Login Handler

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockAuthService)
		expectedStatus int
		expectedBody   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(authSvc *MockAuthService) {
				authSvc.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).
					Return(&models.LoginResponse{
						Token: "test.jwt.token",
						User: models.User{
							ID:       uuid.New(),
							Email:    "test@example.com",
							Username: "testuser",
							Role:     "employee",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.LoginResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "test.jwt.token", response.Token)
				assert.Equal(t, "test@example.com", response.User.Email)
			},
		},
		{
			name: "invalid credentials",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(authSvc *MockAuthService) {
				authSvc.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).
					Return(nil, errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "invalid credentials", response["error"])
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			setupMocks:     func(authSvc *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "invalid request body", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthSvc := new(MockAuthService)
			services := &service.Services{Auth: mockAuthSvc}

			tt.setupMocks(mockAuthSvc)

			var bodyBytes []byte
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
			rec := httptest.NewRecorder()

			handler := loginHandler(services)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			tt.expectedBody(t, rec)

			mockAuthSvc.AssertExpectations(t)
		})
	}
}

// Test Auth Middleware

func TestAuthMiddleware(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}

	userID := uuid.New()
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"email":   "test@example.com",
		"role":    "employee",
	})
	validTokenString, _ := validToken.SignedString([]byte(cfg.JWTSecret))

	tests := []struct {
		name           string
		authHeader     string
		setupMocks     func(*MockAuthService)
		expectedStatus int
		checkContext   bool
	}{
		{
			name:       "valid token",
			authHeader: "Bearer " + validTokenString,
			setupMocks: func(authSvc *MockAuthService) {
				authSvc.On("ValidateToken", validTokenString).Return(validToken, nil)
			},
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			setupMocks:     func(authSvc *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name:           "invalid authorization format",
			authHeader:     "InvalidFormat token",
			setupMocks:     func(authSvc *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid.token.here",
			setupMocks: func(authSvc *MockAuthService) {
				authSvc.On("ValidateToken", "invalid.token.here").
					Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthSvc := new(MockAuthService)
			services := &service.Services{Auth: mockAuthSvc}

			tt.setupMocks(mockAuthSvc)

			// Create a test handler that checks context
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkContext {
					userIDStr, ok := r.Context().Value("user_id").(string)
					assert.True(t, ok)
					assert.Equal(t, userID.String(), userIDStr)

					claims, ok := r.Context().Value("claims").(jwt.MapClaims)
					assert.True(t, ok)
					assert.NotNil(t, claims)
				}
				w.WriteHeader(http.StatusOK)
			})

			middleware := authMiddleware(services)
			handler := middleware(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockAuthSvc.AssertExpectations(t)
		})
	}
}

// Test Helper Functions

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expectError bool
		expectedID  uuid.UUID
	}{
		{
			name: "valid user ID in context",
			setupCtx: func() context.Context {
				userID := uuid.New()
				return context.WithValue(context.Background(), "user_id", userID.String())
			},
			expectError: false,
		},
		{
			name: "user ID not in context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectError: true,
		},
		{
			name: "invalid user ID format",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), "user_id", "not-a-uuid")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			userID, err := getUserIDFromContext(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, userID)
			}
		})
	}
}

func TestGetEmployeeIDFromContext(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expectError bool
	}{
		{
			name: "employee ID in claims",
			setupCtx: func() context.Context {
				employeeID := uuid.New()
				claims := jwt.MapClaims{
					"employee_id": employeeID.String(),
				}
				return context.WithValue(context.Background(), "claims", claims)
			},
			expectError: false,
		},
		{
			name: "fallback to user ID",
			setupCtx: func() context.Context {
				userID := uuid.New()
				return context.WithValue(context.Background(), "user_id", userID.String())
			},
			expectError: false,
		},
		{
			name: "no ID in context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			employeeID, err := getEmployeeIDFromContext(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, employeeID)
			}
		})
	}
}

// Test Response Helpers

func TestRespondJSON(t *testing.T) {
	data := map[string]string{"message": "success"}
	rec := httptest.NewRecorder()

	respondJSON(rec, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

func TestRespondError(t *testing.T) {
	rec := httptest.NewRecorder()

	respondError(rec, http.StatusBadRequest, "validation error")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "validation error", response["error"])
}

// Integration Test - Full Auth Flow

func TestAuthFlow_Integration(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	services := &service.Services{Auth: mockAuthSvc}

	r := chi.NewRouter()
	RegisterAuthRoutes(r, services)

	userID := uuid.New()
	email := "test@example.com"

	// Setup mock
	mockAuthSvc.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).
		Return(&models.LoginResponse{
			Token: "test.jwt.token",
			User: models.User{
				ID:       userID,
				Email:    email,
				Username: "testuser",
				Role:     "employee",
			},
		}, nil)

	// Create login request
	loginReq := models.LoginRequest{
		Email:    email,
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response models.LoginResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "test.jwt.token", response.Token)
	assert.Equal(t, email, response.User.Email)

	mockAuthSvc.AssertExpectations(t)
}

// Benchmark Tests

func BenchmarkLoginHandler(b *testing.B) {
	mockAuthSvc := new(MockAuthService)
	services := &service.Services{Auth: mockAuthSvc}

	mockAuthSvc.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).
		Return(&models.LoginResponse{
			Token: "test.jwt.token",
			User: models.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
				Role:     "employee",
			},
		}, nil)

	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)

	handler := loginHandler(services)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkAuthMiddleware(b *testing.B) {
	mockAuthSvc := new(MockAuthService)
	services := &service.Services{Auth: mockAuthSvc}

	cfg := &config.Config{JWTSecret: "test-secret"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": uuid.New().String(),
		"email":   "test@example.com",
		"role":    "employee",
	})
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))

	mockAuthSvc.On("ValidateToken", tokenString).Return(token, nil)

	middleware := authMiddleware(services)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
