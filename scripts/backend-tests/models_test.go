package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_JSONSerialization(t *testing.T) {
	userID := uuid.New()
	employeeID := uuid.New()
	now := time.Now()

	user := User{
		ID:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "secrethash",
		Role:         "employee",
		EmployeeID:   &employeeID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Test JSON marshaling
	data, err := json.Marshal(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Password hash should not be in JSON
	jsonStr := string(data)
	assert.NotContains(t, jsonStr, "secrethash")
	assert.NotContains(t, jsonStr, "password_hash")

	// Other fields should be present
	assert.Contains(t, jsonStr, "testuser")
	assert.Contains(t, jsonStr, "test@example.com")
	assert.Contains(t, jsonStr, "employee")

	// Test JSON unmarshaling
	var decoded User
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, decoded.ID)
	assert.Equal(t, user.Username, decoded.Username)
	assert.Equal(t, user.Email, decoded.Email)
	assert.Equal(t, user.Role, decoded.Role)
}

func TestUser_NilEmployeeID(t *testing.T) {
	user := User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         "admin",
		EmployeeID:   nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	data, err := json.Marshal(user)
	assert.NoError(t, err)

	var decoded User
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Nil(t, decoded.EmployeeID)
}

func TestLoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request LoginRequest
		isValid bool
	}{
		{
			name: "valid login request",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			isValid: true,
		},
		{
			name: "empty email",
			request: LoginRequest{
				Email:    "",
				Password: "password123",
			},
			isValid: false,
		},
		{
			name: "empty password",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "",
			},
			isValid: false,
		},
		{
			name: "both empty",
			request: LoginRequest{
				Email:    "",
				Password: "",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.request.Email == "" || tt.request.Password == ""
			assert.Equal(t, !tt.isValid, isEmpty)
		})
	}
}

func TestLoginResponse_JSONSerialization(t *testing.T) {
	userID := uuid.New()
	employeeID := uuid.New()

	response := LoginResponse{
		Token: "jwt.token.here",
		User: User{
			ID:         userID,
			Username:   "testuser",
			Email:      "test@example.com",
			Role:       "employee",
			EmployeeID: &employeeID,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		Employee: &Employee{
			ID:        employeeID,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "jwt.token.here")
	assert.Contains(t, jsonStr, "testuser")
	assert.Contains(t, jsonStr, "John")

	// Test JSON unmarshaling
	var decoded LoginResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, response.Token, decoded.Token)
	assert.Equal(t, response.User.ID, decoded.User.ID)
	assert.NotNil(t, decoded.Employee)
	assert.Equal(t, "John", decoded.Employee.FirstName)
}

func TestLoginResponse_NilEmployee(t *testing.T) {
	response := LoginResponse{
		Token: "jwt.token.here",
		User: User{
			ID:       uuid.New(),
			Username: "admin",
			Email:    "admin@example.com",
			Role:     "admin",
		},
		Employee: nil,
	}

	data, err := json.Marshal(response)
	assert.NoError(t, err)

	var decoded LoginResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Nil(t, decoded.Employee)
}

// Benchmark tests

func BenchmarkUser_JSONMarshal(b *testing.B) {
	user := User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      "employee",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(user)
	}
}

func BenchmarkUser_JSONUnmarshal(b *testing.B) {
	user := User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      "employee",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, _ := json.Marshal(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded User
		_ = json.Unmarshal(data, &decoded)
	}
}

func BenchmarkLoginResponse_JSONMarshal(b *testing.B) {
	response := LoginResponse{
		Token: "jwt.token.here",
		User: User{
			ID:        uuid.New(),
			Username:  "testuser",
			Email:     "test@example.com",
			Role:      "employee",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(response)
	}
}
