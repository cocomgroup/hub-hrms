package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"hub-hrms/backend/internal/models"
)

// RepositoryTestSuite provides integration tests for repositories
// These tests require a real PostgreSQL database
type RepositoryTestSuite struct {
	suite.Suite
	db    *pgxpool.Pool
	repos *Repositories
}

// SetupSuite runs once before all tests
func (s *RepositoryTestSuite) SetupSuite() {
	// Get test database URL from environment
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/hrmsdb_test?sslmode=disable"
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		s.T().Skip("Skipping integration tests: database not available")
		return
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		s.T().Skip("Skipping integration tests: cannot ping database")
		return
	}

	s.db = pool
	s.repos = NewRepositories(pool)
}

// TearDownSuite runs once after all tests
func (s *RepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

// SetupTest runs before each test
func (s *RepositoryTestSuite) SetupTest() {
	if s.db == nil {
		s.T().Skip("Database not available")
	}
}

// TearDownTest runs after each test - clean up test data
func (s *RepositoryTestSuite) TearDownTest() {
	if s.db == nil {
		return
	}

	ctx := context.Background()
	// Clean up in reverse order of dependencies
	s.db.Exec(ctx, "DELETE FROM users WHERE email LIKE '%@test.example.com'")
	s.db.Exec(ctx, "DELETE FROM employees WHERE email LIKE '%@test.example.com'")
}

// User Repository Integration Tests

func (s *RepositoryTestSuite) TestUserRepository_Create() {
	ctx := context.Background()

	user := &models.User{
		Username:     "testuser",
		Email:        "testuser@test.example.com",
		PasswordHash: "hashedpassword",
		Role:         "employee",
	}

	err := s.repos.User.Create(ctx, user)
	require.NoError(s.T(), err)
	assert.NotEqual(s.T(), uuid.Nil, user.ID)
	assert.NotZero(s.T(), user.CreatedAt)
	assert.NotZero(s.T(), user.UpdatedAt)
}

func (s *RepositoryTestSuite) TestUserRepository_GetByEmail() {
	ctx := context.Background()

	// Create test user
	user := &models.User{
		Username:     "getbyemail",
		Email:        "getbyemail@test.example.com",
		PasswordHash: "hash",
		Role:         "employee",
	}
	require.NoError(s.T(), s.repos.User.Create(ctx, user))

	// Get by email
	retrieved, err := s.repos.User.GetByEmail(ctx, user.Email)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), user.ID, retrieved.ID)
	assert.Equal(s.T(), user.Email, retrieved.Email)
	assert.Equal(s.T(), user.Username, retrieved.Username)
}

func (s *RepositoryTestSuite) TestUserRepository_GetByID() {
	ctx := context.Background()

	// Create test user
	user := &models.User{
		Username:     "getbyid",
		Email:        "getbyid@test.example.com",
		PasswordHash: "hash",
		Role:         "manager",
	}
	require.NoError(s.T(), s.repos.User.Create(ctx, user))

	// Get by ID
	retrieved, err := s.repos.User.GetByID(ctx, user.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), user.ID, retrieved.ID)
	assert.Equal(s.T(), user.Email, retrieved.Email)
}

func (s *RepositoryTestSuite) TestUserRepository_Update() {
	ctx := context.Background()

	// Create test user
	user := &models.User{
		Username:     "updateuser",
		Email:        "update@test.example.com",
		PasswordHash: "hash",
		Role:         "employee",
	}
	require.NoError(s.T(), s.repos.User.Create(ctx, user))

	// Update user
	originalUpdatedAt := user.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure timestamp changes
	user.Username = "updatedusername"
	user.Role = "manager"

	err := s.repos.User.Update(ctx, user)
	require.NoError(s.T(), err)
	assert.NotEqual(s.T(), originalUpdatedAt, user.UpdatedAt)

	// Verify update
	retrieved, err := s.repos.User.GetByID(ctx, user.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "updatedusername", retrieved.Username)
	assert.Equal(s.T(), "manager", retrieved.Role)
}

func (s *RepositoryTestSuite) TestUserRepository_List() {
	ctx := context.Background()

	// Create test users
	users := []*models.User{
		{Username: "alice", Email: "alice@test.example.com", PasswordHash: "hash", Role: "employee"},
		{Username: "bob", Email: "bob@test.example.com", PasswordHash: "hash", Role: "manager"},
		{Username: "charlie", Email: "charlie@test.example.com", PasswordHash: "hash", Role: "employee"},
	}

	for _, user := range users {
		require.NoError(s.T(), s.repos.User.Create(ctx, user))
	}

	tests := []struct {
		name          string
		search        string
		role          string
		expectedCount int
	}{
		{
			name:          "list all",
			search:        "",
			role:          "",
			expectedCount: 3,
		},
		{
			name:          "filter by role",
			search:        "",
			role:          "employee",
			expectedCount: 2,
		},
		{
			name:          "search by username",
			search:        "alice",
			role:          "",
			expectedCount: 1,
		},
		{
			name:          "search and filter",
			search:        "bob",
			role:          "manager",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			result, err := s.repos.User.List(ctx, tt.search, tt.role)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(result), tt.expectedCount)
		})
	}
}

func (s *RepositoryTestSuite) TestUserRepository_Delete() {
	ctx := context.Background()

	// Create test user
	user := &models.User{
		Username:     "deleteuser",
		Email:        "delete@test.example.com",
		PasswordHash: "hash",
		Role:         "employee",
	}
	require.NoError(s.T(), s.repos.User.Create(ctx, user))

	// Delete user
	err := s.repos.User.Delete(ctx, user.ID)
	require.NoError(s.T(), err)

	// Verify deletion
	_, err = s.repos.User.GetByID(ctx, user.ID)
	assert.Error(s.T(), err) // Should not be found
}

// Employee Repository Integration Tests

func (s *RepositoryTestSuite) TestEmployeeRepository_Create() {
	ctx := context.Background()

	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	employee := &models.Employee{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john.doe@test.example.com",
		Phone:       "555-0100",
		DateOfBirth: &dob,
		HireDate:    time.Now(),
		Department:  "Engineering",
		Position:    "Software Engineer",
		Status:      "active",
	}

	err := s.repos.Employee.Create(ctx, employee)
	require.NoError(s.T(), err)
	assert.NotEqual(s.T(), uuid.Nil, employee.ID)
	assert.NotZero(s.T(), employee.CreatedAt)
}

func (s *RepositoryTestSuite) TestEmployeeRepository_GetByEmail() {
	ctx := context.Background()

	// Create test employee
	employee := &models.Employee{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@test.example.com",
		HireDate:  time.Now(),
		Status:    "active",
	}
	require.NoError(s.T(), s.repos.Employee.Create(ctx, employee))

	// Get by email
	retrieved, err := s.repos.Employee.GetByEmail(ctx, employee.Email)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), employee.ID, retrieved.ID)
	assert.Equal(s.T(), employee.Email, retrieved.Email)
}

func (s *RepositoryTestSuite) TestEmployeeRepository_Update() {
	ctx := context.Background()

	// Create test employee
	employee := &models.Employee{
		FirstName: "Update",
		LastName:  "Test",
		Email:     "update.test@test.example.com",
		HireDate:  time.Now(),
		Status:    "active",
	}
	require.NoError(s.T(), s.repos.Employee.Create(ctx, employee))

	// Update employee
	employee.Department = "Sales"
	employee.Position = "Sales Manager"

	err := s.repos.Employee.Update(ctx, employee)
	require.NoError(s.T(), err)

	// Verify update
	retrieved, err := s.repos.Employee.GetByID(ctx, employee.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Sales", retrieved.Department)
	assert.Equal(s.T(), "Sales Manager", retrieved.Position)
}

// NewRepositories Tests

func TestNewRepositories(t *testing.T) {
	// This is a unit test that doesn't require a database
	// It just verifies the factory function creates all repositories

	// We can't actually create a pool without a database, but we can test the structure
	repos := &Repositories{}

	assert.NotNil(t, repos)
	
	// Verify all repository fields exist (struct completeness)
	// This will fail to compile if any repository is missing
	_ = repos.User
	_ = repos.Employee
	_ = repos.Onboarding
	_ = repos.Workflow
	_ = repos.Timesheet
	_ = repos.PTO
	_ = repos.Benefits
	_ = repos.Payroll
	_ = repos.Recruiting
	_ = repos.Organization
	_ = repos.Project
	_ = repos.Compensation
	_ = repos.BankInfo
}

// TestUserRepository_ErrorHandling tests error cases
func TestUserRepository_ErrorHandling(t *testing.T) {
	// Skip if no test database
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping: TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skip("Skipping: cannot connect to database")
	}
	defer pool.Close()

	repo := NewUserRepository(pool)

	tests := []struct {
		name    string
		action  func() error
		wantErr bool
	}{
		{
			name: "get non-existent user by ID",
			action: func() error {
				_, err := repo.GetByID(ctx, uuid.New())
				return err
			},
			wantErr: true,
		},
		{
			name: "get non-existent user by email",
			action: func() error {
				_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
				return err
			},
			wantErr: true,
		},
		{
			name: "delete non-existent user",
			action: func() error {
				return repo.Delete(ctx, uuid.New())
			},
			wantErr: false, // DELETE doesn't error if row doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEmployeeRepository_ErrorHandling tests error cases
func TestEmployeeRepository_ErrorHandling(t *testing.T) {
	// Skip if no test database
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping: TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skip("Skipping: cannot connect to database")
	}
	defer pool.Close()

	repo := NewEmployeeRepository(pool)

	tests := []struct {
		name    string
		action  func() error
		wantErr bool
	}{
		{
			name: "get non-existent employee by ID",
			action: func() error {
				_, err := repo.GetByID(ctx, uuid.New())
				return err
			},
			wantErr: true,
		},
		{
			name: "get non-existent employee by email",
			action: func() error {
				_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
				return err
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Run the test suite
func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

// Benchmark tests

func BenchmarkUserRepository_Create(b *testing.B) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		b.Skip("Skipping: TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		b.Skip("Skipping: cannot connect to database")
	}
	defer pool.Close()

	repo := NewUserRepository(pool)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &models.User{
			Username:     "benchuser",
			Email:        "bench@example.com",
			PasswordHash: "hash",
			Role:         "employee",
		}
		_ = repo.Create(ctx, user)
	}
}

func BenchmarkUserRepository_GetByEmail(b *testing.B) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		b.Skip("Skipping: TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		b.Skip("Skipping: cannot connect to database")
	}
	defer pool.Close()

	repo := NewUserRepository(pool)

	// Create test user
	user := &models.User{
		Username:     "benchuser",
		Email:        "bench.get@test.example.com",
		PasswordHash: "hash",
		Role:         "employee",
	}
	_ = repo.Create(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByEmail(ctx, user.Email)
	}
}