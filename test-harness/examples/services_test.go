package services

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UserServiceTestSuite is a test suite for UserService
type UserServiceTestSuite struct {
	suite.Suite
	service *UserService
	mock    sqlmock.Sqlmock
}

// SetupTest runs before each test
func (suite *UserServiceTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	assert.NoError(suite.T(), err)
	
	suite.mock = mock
	suite.service = NewUserService(db)
}

// TestGetUser tests the GetUser method
func (suite *UserServiceTestSuite) TestGetUser() {
	userID := "123"
	expectedUser := &User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "employee",
		CreatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "created_at"}).
		AddRow(expectedUser.ID, expectedUser.Email, expectedUser.FirstName, 
			expectedUser.LastName, expectedUser.Role, expectedUser.CreatedAt)

	suite.mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
		WithArgs(userID).
		WillReturnRows(rows)

	// Execute
	user, err := suite.service.GetUser(userID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), expectedUser.ID, user.ID)
	assert.Equal(suite.T(), expectedUser.Email, user.Email)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestGetUserNotFound tests GetUser when user doesn't exist
func (suite *UserServiceTestSuite) TestGetUserNotFound() {
	userID := "999"

	suite.mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	// Execute
	user, err := suite.service.GetUser(userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Equal(suite.T(), ErrUserNotFound, err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestCreateUser tests the CreateUser method
func (suite *UserServiceTestSuite) TestCreateUser() {
	newUser := &User{
		Email:     "new@example.com",
		FirstName: "New",
		LastName:  "User",
		Password:  "SecurePass123!",
		Role:      "employee",
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), newUser.Email, newUser.FirstName, 
			newUser.LastName, sqlmock.AnyArg(), newUser.Role, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()

	// Execute
	err := suite.service.CreateUser(newUser)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), newUser.ID)
	assert.NotEmpty(suite.T(), newUser.CreatedAt)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestCreateUserDuplicateEmail tests CreateUser with duplicate email
func (suite *UserServiceTestSuite) TestCreateUserDuplicateEmail() {
	newUser := &User{
		Email:     "existing@example.com",
		FirstName: "Test",
		LastName:  "User",
		Password:  "SecurePass123!",
	}

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), newUser.Email, newUser.FirstName, 
			newUser.LastName, sqlmock.AnyArg(), newUser.Role, sqlmock.AnyArg()).
		WillReturnError(errors.New("UNIQUE constraint failed: users.email"))
	suite.mock.ExpectRollback()

	// Execute
	err := suite.service.CreateUser(newUser)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrDuplicateEmail, err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestUpdateUser tests the UpdateUser method
func (suite *UserServiceTestSuite) TestUpdateUser() {
	userID := "123"
	updatedUser := &User{
		FirstName: "Updated",
		LastName:  "Name",
		Email:     "updated@example.com",
	}

	suite.mock.ExpectExec("UPDATE users SET").
		WithArgs(updatedUser.FirstName, updatedUser.LastName, 
			updatedUser.Email, sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	err := suite.service.UpdateUser(userID, updatedUser)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestUpdateUserNotFound tests UpdateUser when user doesn't exist
func (suite *UserServiceTestSuite) TestUpdateUserNotFound() {
	userID := "999"
	updatedUser := &User{
		FirstName: "Updated",
		LastName:  "Name",
	}

	suite.mock.ExpectExec("UPDATE users SET").
		WithArgs(updatedUser.FirstName, updatedUser.LastName, 
			sqlmock.AnyArg(), sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute
	err := suite.service.UpdateUser(userID, updatedUser)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrUserNotFound, err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestDeleteUser tests the DeleteUser method
func (suite *UserServiceTestSuite) TestDeleteUser() {
	userID := "123"

	suite.mock.ExpectExec("DELETE FROM users WHERE id = ?").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	err := suite.service.DeleteUser(userID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestListUsers tests the ListUsers method
func (suite *UserServiceTestSuite) TestListUsers() {
	page := 1
	limit := 10
	offset := 0

	// Mock count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(countRows)

	// Mock users query
	rows := sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "created_at"}).
		AddRow("1", "user1@example.com", "User", "One", "employee", time.Now()).
		AddRow("2", "user2@example.com", "User", "Two", "manager", time.Now())

	suite.mock.ExpectQuery("SELECT (.+) FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?").
		WithArgs(limit, offset).
		WillReturnRows(rows)

	// Execute
	users, total, err := suite.service.ListUsers(page, limit)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), users, 2)
	assert.Equal(suite.T(), 25, total)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestValidateUserData tests user validation
func (suite *UserServiceTestSuite) TestValidateUserData() {
	tests := []struct {
		name      string
		user      *User
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid user",
			user: &User{
				Email:     "valid@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "SecurePass123!",
			},
			expectErr: false,
		},
		{
			name: "invalid email",
			user: &User{
				Email:     "invalid-email",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "SecurePass123!",
			},
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name: "weak password",
			user: &User{
				Email:     "valid@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "weak",
			},
			expectErr: true,
			errMsg:    "password must be at least 8 characters",
		},
		{
			name: "missing required fields",
			user: &User{
				Email: "valid@example.com",
			},
			expectErr: true,
			errMsg:    "first name and last name are required",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := suite.service.ValidateUserData(tt.user)
			
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Run the test suite
func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
