package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUser(id string) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserService) CreateUser(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserService) UpdateUser(id string, user *User) error {
	args := m.Called(id, user)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(page, limit int) ([]*User, int, error) {
	args := m.Called(page, limit)
	return args.Get(0).([]*User), args.Int(1), args.Error(2)
}

// TestGetUserHandler tests the GetUser handler
func TestGetUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockUser       *User
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful user retrieval",
			userID: "123",
			mockUser: &User{
				ID:        "123",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not found",
			userID:         "999",
			mockUser:       nil,
			mockError:      ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "user not found",
		},
		{
			name:           "invalid user ID",
			userID:         "",
			mockUser:       nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			handler := NewUserHandler(mockService)

			if tt.userID != "" && tt.mockError != ErrUserNotFound {
				mockService.On("GetUser", tt.userID).Return(tt.mockUser, tt.mockError)
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/users/"+tt.userID, nil)
			rec := httptest.NewRecorder()

			// Execute
			handler.GetUser(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response User
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockUser.ID, response.ID)
				assert.Equal(t, tt.mockUser.Email, response.Email)
			} else if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// TestCreateUserHandler tests the CreateUser handler
func TestCreateUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful user creation",
			requestBody: map[string]string{
				"email":      "new@example.com",
				"first_name": "New",
				"last_name":  "User",
				"password":   "SecurePass123!",
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "duplicate email",
			requestBody: map[string]string{
				"email":      "existing@example.com",
				"first_name": "Test",
				"last_name":  "User",
				"password":   "SecurePass123!",
			},
			mockError:      ErrDuplicateEmail,
			expectedStatus: http.StatusConflict,
			expectedBody:   "email already exists",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name: "missing required fields",
			requestBody: map[string]string{
				"email": "incomplete@example.com",
			},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "missing required fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			handler := NewUserHandler(mockService)

			// Setup mock expectation for valid requests
			if tt.expectedStatus == http.StatusCreated || tt.mockError != nil {
				mockService.On("CreateUser", mock.AnythingOfType("*User")).Return(tt.mockError)
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute
			handler.CreateUser(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}

			if tt.expectedStatus == http.StatusCreated || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

// TestListUsersHandler tests the ListUsers handler
func TestListUsersHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockUsers      []*User
		mockTotal      int
		mockError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "successful list with pagination",
			queryParams: "page=1&limit=10",
			mockUsers: []*User{
				{ID: "1", Email: "user1@example.com"},
				{ID: "2", Email: "user2@example.com"},
			},
			mockTotal:      2,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "empty result set",
			queryParams:    "page=10&limit=10",
			mockUsers:      []*User{},
			mockTotal:      0,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "invalid pagination parameters",
			queryParams:    "page=-1&limit=10",
			mockUsers:      nil,
			mockTotal:      0,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			handler := NewUserHandler(mockService)

			if tt.expectedStatus == http.StatusOK {
				mockService.On("ListUsers", mock.AnythingOfType("int"), mock.AnythingOfType("int")).
					Return(tt.mockUsers, tt.mockTotal, tt.mockError)
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/users?"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			// Execute
			handler.ListUsers(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Users []*User `json:"users"`
					Total int     `json:"total"`
				}
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(response.Users))
				assert.Equal(t, tt.mockTotal, response.Total)
			}

			if tt.expectedStatus == http.StatusOK {
				mockService.AssertExpectations(t)
			}
		})
	}
}
