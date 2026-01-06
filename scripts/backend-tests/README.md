# hub-hrms Backend Test Suite

Comprehensive test files for all hub-hrms backend packages.

## 📦 Test Files Included

### 1. `config_test.go`
Tests for the `internal/config` package:
- Configuration loading from environment variables
- Default value handling
- Database URL generation
- Environment variable parsing

**Place in:** `backend/internal/config/config_test.go`

### 2. `models_test.go`
Tests for the `internal/models` package:
- JSON serialization/deserialization
- Model validation
- Password hash exclusion from JSON
- Benchmark tests for performance

**Place in:** `backend/internal/models/models_test.go`

### 3. `service_test.go`
Comprehensive tests for the `internal/service` package:
- UserService CRUD operations
- AuthService authentication logic
- Password hashing and verification
- JWT token generation and validation
- Login flow with mocked repositories
- Benchmark tests for critical operations

**Place in:** `backend/internal/service/service_test.go`

### 4. `api_test.go`
Tests for the `internal/api` package:
- HTTP handler testing
- Authentication middleware
- Request/response validation
- Context management
- Integration tests for auth flow
- Benchmark tests for handlers

**Place in:** `backend/internal/api/api_test.go`

## 🚀 Installation

### Step 1: Copy Test Files

```powershell
# Copy each test file to its respective package directory

# Config tests
Copy-Item config_test.go backend\internal\config\

# Models tests
Copy-Item models_test.go backend\internal\models\

# Service tests
Copy-Item service_test.go backend\internal\service\

# API tests
Copy-Item api_test.go backend\internal\api\
```

### Step 2: Install Testing Dependencies

```powershell
cd backend

# Install testify for assertions and mocks
go get -u github.com/stretchr/testify/assert
go get -u github.com/stretchr/testify/mock
go get -u github.com/stretchr/testify/suite

# Install other dependencies if needed
go mod tidy
```

## 🧪 Running Tests

### Run All Tests

```powershell
# From backend directory
cd backend

# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Package Tests

```powershell
# Test config package
go test ./internal/config -v

# Test models package
go test ./internal/models -v

# Test service package
go test ./internal/service -v

# Test API handlers
go test ./internal/api -v
```

### Run Specific Tests

```powershell
# Run a specific test
go test ./internal/service -run TestAuthService_Login -v

# Run tests matching a pattern
go test ./internal/service -run TestUserService -v
```

### Run with Coverage

```powershell
# Generate coverage for specific package
go test ./internal/service -coverprofile=service-coverage.out
go tool cover -func=service-coverage.out

# View HTML coverage report
go tool cover -html=service-coverage.out
```

## 📊 Test Coverage

Target coverage thresholds:
- **Config:** 90%+ (simple configuration loading)
- **Models:** 85%+ (data structures and validation)
- **Service:** 80%+ (business logic)
- **API:** 75%+ (HTTP handlers)

Check current coverage:

```powershell
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | Select-String "total:"
```

## 🔧 Test Structure

### Unit Tests
Each test file includes comprehensive unit tests with:
- Multiple test cases using table-driven tests
- Mock dependencies using testify/mock
- Assertions using testify/assert
- Edge case and error handling tests

### Mock Objects
The test files include mock implementations:
- `MockUserRepository` - Mocks database operations
- `MockEmployeeRepository` - Mocks employee database operations
- `MockAuthService` - Mocks authentication service
- `MockUserService` - Mocks user service

### Benchmark Tests
Performance benchmarks for critical operations:
- Password hashing
- JWT token generation
- JSON serialization
- Handler processing

Run benchmarks:
```powershell
go test -bench=. -benchmem ./internal/service
go test -bench=. -benchmem ./internal/api
```

## 📝 Writing Additional Tests

### Example: Testing a New Service Method

```go
func TestMyService_NewMethod(t *testing.T) {
    // Setup
    mockRepo := new(MockUserRepository)
    service := NewMyService(mockRepo)
    
    // Configure mock
    mockRepo.On("SomeMethod", mock.Anything).Return(expectedResult, nil)
    
    // Execute
    result, err := service.NewMethod(context.Background())
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
    mockRepo.AssertExpectations(t)
}
```

### Example: Testing an API Handler

```go
func TestMyHandler(t *testing.T) {
    // Setup
    mockService := new(MockMyService)
    services := &service.Services{MyService: mockService}
    
    // Configure mock
    mockService.On("DoSomething", mock.Anything).Return(result, nil)
    
    // Create request
    req := httptest.NewRequest(http.MethodGet, "/endpoint", nil)
    rec := httptest.NewRecorder()
    
    // Execute
    handler := myHandler(services)
    handler.ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, rec.Code)
    mockService.AssertExpectations(t)
}
```

## 🐛 Troubleshooting

### Issue: Import Errors

```powershell
# Make sure go.mod is in the backend directory
cd backend
go mod tidy

# If you see import errors, update module path in test files
# Change: "hub-hrms/backend/internal/..."
# To match your actual module path in go.mod
```

### Issue: Mock Method Not Found

If you see "method not found on mock" errors:
1. Check that the mock implements the interface correctly
2. Ensure all required methods are mocked
3. Use `mock.Anything` for parameters you don't care about

### Issue: Tests Failing Due to Missing Dependencies

```powershell
# Install all testing dependencies
go get -u github.com/stretchr/testify/assert
go get -u github.com/stretchr/testify/mock
go get -u github.com/stretchr/testify/suite
go get -u github.com/golang-jwt/jwt/v5
go get -u github.com/google/uuid
go get -u golang.org/x/crypto/bcrypt
go get -u github.com/go-chi/chi/v5

# Update all dependencies
go mod tidy
```

## 📈 Continuous Integration

### GitHub Actions Example

```yaml
name: Backend Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        working-directory: ./backend
        run: |
          go test -v -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./backend/coverage.out
```

## 🎯 Best Practices

1. **Always use table-driven tests** for multiple scenarios
2. **Mock external dependencies** to keep tests fast and reliable
3. **Test error cases** as thoroughly as success cases
4. **Use meaningful test names** that describe what's being tested
5. **Keep tests independent** - no test should depend on another
6. **Clean up resources** in defer statements or teardown functions
7. **Use subtests** (`t.Run`) for better organization
8. **Check coverage** regularly and aim for >80%

## 📚 Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Go Test Coverage](https://go.dev/blog/cover)

## 🤝 Contributing

When adding new features to hub-hrms:
1. Write tests first (TDD)
2. Ensure all tests pass
3. Maintain or improve coverage
4. Add benchmark tests for performance-critical code
5. Update this README if adding new test patterns

## ✅ Quick Checklist

Before committing:
- [ ] All tests pass: `go test ./...`
- [ ] Coverage >75%: `go test -cover ./...`
- [ ] No race conditions: `go test -race ./...`
- [ ] Benchmarks run: `go test -bench=.`
- [ ] Code formatted: `go fmt ./...`
- [ ] Imports organized: `goimports -w .`

## 🎉 You're Ready!

Run your tests:

```powershell
cd backend
go test -v ./...
```

Happy testing! 🧪
