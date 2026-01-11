# Testing Structure - Hub HRMS Backend

## Directory Structure

```
backend/
├── tests/
│   ├── unit/               # Unit tests (fast, isolated)
│   │   ├── auth_service_test.go
│   │   ├── applicant_service_test.go
│   │   └── ...
│   ├── integration/        # Integration tests (slower, with DB)
│   │   └── ...
│   ├── mocks/             # Mock implementations
│   │   ├── repositories.go
│   │   └── ...
│   ├── fixtures/          # Test data fixtures
│   │   ├── models.go
│   │   └── ...
│   └── helpers/           # Test helper functions
│       ├── repository_helpers.go
│       └── ...
└── internal/
    └── service/
        └── service_test.go    # Legacy tests (being migrated)
```

## Test Types

### Unit Tests (`tests/unit/`)
**Purpose:** Test individual components in isolation
**Speed:** Fast (<1s per test)
**Dependencies:** Mocked

**Example:**
```go
func TestApplicantService_Create_Success(t *testing.T) {
    mockRepo := &mocks.MockApplicantRepository{}
    repos := &repository.Repositories{Applicant: mockRepo}
    
    mockRepo.On("Create", mock.Anything, testApplicant).Return(nil)
    
    service := service.NewApplicantService(repos)
    err := service.Create(context.Background(), testApplicant)
    
    assert.NoError(t, err)
}
```

### Integration Tests (`tests/integration/`)
**Purpose:** Test components working together
**Speed:** Slower (with real DB)
**Dependencies:** Real database, external services

**Example:**
```go
func TestApplicantWorkflow_EndToEnd(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Test complete workflow
}
```

## Using Mocks

### 1. Import Mocks
```go
import "hub-hrms/backend/tests/mocks"
```

### 2. Create Mock Repository
```go
mockApplicantRepo := &mocks.MockApplicantRepository{}
```

### 3. Set Expectations
```go
mockApplicantRepo.On("Create", mock.Anything, testApplicant).
    Return(nil)
```

### 4. Assert Expectations Met
```go
mockApplicantRepo.AssertExpectations(t)
```

## Using Fixtures

### 1. Import Fixtures
```go
import "hub-hrms/backend/tests/fixtures"
```

### 2. Create Test Data
```go
testUser := fixtures.NewUser()
testEmployee := fixtures.NewEmployee()
testApplicant := fixtures.NewApplicant()
```

### 3. Customize Fixtures
```go
testUser := fixtures.NewUser()
testUser.Role = "admin"
testUser.Email = "custom@example.com"
```

## Using Helpers

### 1. Create Mock Repositories
```go
import "hub-hrms/backend/tests/helpers"

repos, mockRepos := helpers.NewMockRepositories()

// Access individual mocks
mockRepos.User.On("GetByEmail", ...).Return(...)
mockRepos.Employee.On("GetByID", ...).Return(...)
```

### 2. Create Minimal Mock Setup
```go
repos := helpers.NewMinimalMockRepositories("user", "employee")
```

## Running Tests

### Run All Tests
```bash
cd backend
go test ./...
```

### Run Unit Tests Only
```bash
go test ./tests/unit/...
```

### Run Integration Tests Only
```bash
go test ./tests/integration/...
```

### Run Specific Test
```bash
go test ./tests/unit/ -run TestApplicantService_Create_Success
```

### Run with Coverage
```bash
go test ./... -cover
```

### Run with Verbose Output
```bash
go test ./... -v
```

### Run with Coverage Report
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Naming Conventions

### Test Function Names
```
Test<ComponentName>_<MethodName>_<Scenario>

Examples:
- TestApplicantService_Create_Success
- TestApplicantService_Create_RepositoryError
- TestAuthService_Login_InvalidCredentials
```

### Mock Repository Methods
```
Mock<EntityName>Repository

Examples:
- MockUserRepository
- MockEmployeeRepository
- MockApplicantRepository
```

### Fixture Functions
```
New<EntityName>()
New<EntityName>With<Attribute>()

Examples:
- NewUser()
- NewAdminUser()
- NewApplicantWithScore()
```

## Writing a New Test

### 1. Create Test File
```bash
# For unit test
touch tests/unit/my_service_test.go

# For integration test
touch tests/integration/my_workflow_test.go
```

### 2. Write Test
```go
package unit

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "hub-hrms/backend/tests/fixtures"
    "hub-hrms/backend/tests/helpers"
    "hub-hrms/backend/tests/mocks"
)

func TestMyService_MyMethod_Success(t *testing.T) {
    // Arrange
    repos, mockRepos := helpers.NewMockRepositories()
    testData := fixtures.NewApplicant()
    
    mockRepos.Applicant.On("Create", mock.Anything, testData).
        Return(nil)
    
    service := service.NewMyService(repos)
    
    // Act
    err := service.MyMethod(context.Background(), testData)
    
    // Assert
    assert.NoError(t, err)
    mockRepos.Applicant.AssertExpectations(t)
}
```

### 3. Run Test
```bash
go test ./tests/unit/ -run TestMyService_MyMethod_Success -v
```

## Adding New Mocks

### 1. Add to `tests/mocks/repositories.go`
```go
type MockNewEntityRepository struct {
    mock.Mock
}

func (m *MockNewEntityRepository) Create(ctx context.Context, entity *models.NewEntity) error {
    args := m.Called(ctx, entity)
    return args.Error(0)
}

// Add all interface methods...
```

### 2. Add to Helper
```go
// In tests/helpers/repository_helpers.go
func NewMockRepositories() (*repository.Repositories, *RepositoryMocks) {
    mocks := &RepositoryMocks{
        User:      &mocks.MockUserRepository{},
        Employee:  &mocks.MockEmployeeRepository{},
        NewEntity: &mocks.MockNewEntityRepository{}, // Add this
    }

    repos := &repository.Repositories{
        User:      mocks.User,
        Employee:  mocks.Employee,
        NewEntity: mocks.NewEntity, // Add this
        // ... rest
    }

    return repos, mocks
}
```

## Adding New Fixtures

### 1. Add to `tests/fixtures/models.go`
```go
func NewMyEntity() *models.MyEntity {
    return &models.MyEntity{
        ID:        uuid.New(),
        Name:      "Test Entity",
        Status:    "active",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

func NewMyEntityWithStatus(status string) *models.MyEntity {
    entity := NewMyEntity()
    entity.Status = status
    return entity
}
```

## Test Organization Best Practices

### 1. **Arrange, Act, Assert** Pattern
```go
func TestMyMethod(t *testing.T) {
    // Arrange - Set up test data and mocks
    testData := fixtures.NewApplicant()
    mockRepo.On("Create", mock.Anything, testData).Return(nil)
    
    // Act - Call the method being tested
    result, err := service.MyMethod(context.Background(), testData)
    
    // Assert - Verify the results
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 2. **Table-Driven Tests**
```go
func TestMyMethod_MultipleCases(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    string
        expectError bool
    }{
        {"valid input", "test", "TEST", false},
        {"empty input", "", "", true},
        {"special chars", "test!", "TEST!", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := MyMethod(tt.input)
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### 3. **Test Helpers for Common Setup**
```go
func setupAuthTest(t *testing.T) (*service.AuthService, *mocks.RepositoryMocks) {
    repos, mockRepos := helpers.NewMockRepositories()
    cfg := &config.Config{JWTSecret: "test-secret"}
    authService := service.NewAuthService(repos, cfg)
    return authService, mockRepos
}

func TestLogin(t *testing.T) {
    authService, mockRepos := setupAuthTest(t)
    // ... rest of test
}
```

## Common Testing Patterns

### Testing Error Cases
```go
func TestMyMethod_ErrorHandling(t *testing.T) {
    mockRepo.On("Create", mock.Anything, mock.Anything).
        Return(errors.New("database error"))
    
    err := service.Create(context.Background(), testData)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "database error")
}
```

### Testing with Context Timeout
```go
func TestMyMethod_ContextTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()
    
    time.Sleep(2 * time.Millisecond) // Force timeout
    
    err := service.MyMethod(ctx, testData)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "context deadline exceeded")
}
```

### Testing Concurrent Operations
```go
func TestMyMethod_Concurrent(t *testing.T) {
    var wg sync.WaitGroup
    errors := make(chan error, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            if err := service.MyMethod(context.Background(), testData); err != nil {
                errors <- err
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        assert.NoError(t, err)
    }
}
```

## Migration from Old Tests

The old tests in `internal/service/service_test.go` are being migrated to the new structure:

1. Extract mock definitions → `tests/mocks/`
2. Extract test data → `tests/fixtures/`
3. Rewrite tests using helpers → `tests/unit/`
4. Keep running old tests until migration complete

## Fixed Issues

### ✅ Line 546 in service_test.go
**Problem:** Missing `Applicant` field in Repositories struct
**Solution:** Added `Applicant: nil` to all Repositories initializations

**Before:**
```go
repos := &repository.Repositories{
    User:     mockUserRepo,
    Employee: mockEmpRepo,
}
```

**After:**
```go
repos := &repository.Repositories{
    User:            mockUserRepo,
    Employee:        mockEmpRepo,
    Applicant:       nil,
    Onboarding:      nil,
    // ... all other fields
}
```

## Summary

- ✅ **Unit tests** - tests/unit/
- ✅ **Integration tests** - tests/integration/
- ✅ **Mocks** - tests/mocks/
- ✅ **Fixtures** - tests/fixtures/
- ✅ **Helpers** - tests/helpers/
- ✅ **service_test.go line 546 fixed**
- ✅ **Clean separation of concerns**
- ✅ **Reusable test components**

Use the new structure for all new tests! 🎯
