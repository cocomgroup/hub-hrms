# hub-hrms Test Harness

Comprehensive testing infrastructure for the hub-hrms HRMS application, covering backend Go services and frontend Svelte application.

## Table of Contents

- [Overview](#overview)
- [Setup](#setup)
- [Running Tests](#running-tests)
- [Backend Tests](#backend-tests)
- [Frontend Tests](#frontend-tests)
- [Test Coverage](#test-coverage)
- [CI/CD Integration](#cicd-integration)
- [Best Practices](#best-practices)

## Overview

This test harness provides:

- **Backend Unit Tests**: Test individual functions, handlers, and services
- **Backend Integration Tests**: Test API endpoints with real database
- **Backend E2E Tests**: Test complete user workflows using Playwright
- **Frontend Unit Tests**: Test Svelte components and stores using Vitest
- **Frontend Integration Tests**: Test component interactions
- **Coverage Reports**: Comprehensive code coverage for all layers

## Setup

### Initial Setup

Run the setup script to install all testing dependencies:

```bash
chmod +x setup-test-harness.sh
./setup-test-harness.sh
```

This will:
- Install Go testing libraries (testify, sqlmock, httpmock)
- Install frontend testing libraries (Vitest, Testing Library, MSW)
- Create test directory structure
- Configure test runners

### Prerequisites

**Backend:**
- Go 1.21+
- Docker (for integration tests)

**Frontend:**
- Node.js 18+
- npm or pnpm

### Directory Structure

```
hub-hrms/
├── tests/
│   ├── unit/              # Backend unit tests
│   ├── integration/       # Backend integration tests
│   ├── e2e/              # E2E tests
│   ├── fixtures/         # Test data fixtures
│   └── mocks/            # Mock implementations
├── frontend/
│   └── src/
│       └── tests/        # Frontend tests
│           ├── components/
│           ├── stores/
│           └── setup.ts
├── scripts/
│   ├── setup-test-harness.sh
│   ├── test-backend.sh
│   ├── test-frontend.sh
│   └── test-all.sh
└── coverage/             # Coverage reports
```

## Running Tests

### Quick Start

```bash
# Run all tests (unit + frontend)
./scripts/test-all.sh

# Run all tests with coverage
./scripts/test-all.sh --coverage

# Run everything including integration and E2E
./scripts/test-all.sh --all
```

### Backend Tests

```bash
# Unit tests only
./scripts/test-backend.sh

# With verbose output
./scripts/test-backend.sh --verbose

# With coverage
./scripts/test-backend.sh --coverage

# Integration tests
./scripts/test-backend.sh --integration --coverage

# E2E tests
./scripts/test-backend.sh --e2e

# Using Go directly
go test ./... -short              # Unit tests
go test ./tests/integration/...   # Integration tests
go test ./tests/e2e/...           # E2E tests
```

### Frontend Tests

```bash
# Run all frontend tests
./scripts/test-frontend.sh

# Watch mode (for development)
./scripts/test-frontend.sh --watch

# With coverage
./scripts/test-frontend.sh --coverage

# Interactive UI
./scripts/test-frontend.sh --ui

# Using npm directly
cd frontend
npm run test              # Run once
npm run test:watch        # Watch mode
npm run test:coverage     # With coverage
npm run test:ui          # Interactive UI
```

## Backend Tests

### Unit Tests

Unit tests focus on testing individual functions and methods in isolation using mocks.

**Example: Testing a handler**

```go
func TestGetUserHandler(t *testing.T) {
    mockService := new(MockUserService)
    handler := NewUserHandler(mockService)
    
    mockService.On("GetUser", "123").Return(expectedUser, nil)
    
    req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
    rec := httptest.NewRecorder()
    
    handler.GetUser(rec, req)
    
    assert.Equal(t, http.StatusOK, rec.Code)
    mockService.AssertExpectations(t)
}
```

**Running specific tests:**

```bash
# Run tests for a specific package
go test ./internal/handlers -v

# Run a specific test
go test ./internal/handlers -run TestGetUserHandler -v

# Run with race detection
go test -race ./...
```

### Integration Tests

Integration tests run against a real database and test the full request/response cycle.

**Setup:**

Integration tests use Docker to spin up test databases:

```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d db

# Run integration tests
go test ./tests/integration/... -v

# Cleanup
docker-compose -f docker-compose.test.yml down
```

**Example: Testing API endpoints**

```go
func (suite *APIIntegrationTestSuite) TestUserCRUDOperations() {
    // Create user via API
    resp := suite.createUser(newUser)
    assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
    
    // Retrieve user
    resp = suite.getUser(userID)
    assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
    
    // Update user
    resp = suite.updateUser(userID, updates)
    assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
    
    // Delete user
    resp = suite.deleteUser(userID)
    assert.Equal(suite.T(), http.StatusNoContent, resp.StatusCode)
}
```

### E2E Tests

E2E tests use Playwright to test complete user workflows in a real browser.

**Prerequisites:**

```bash
# Install Playwright
go get github.com/playwright-community/playwright-go
go run github.com/playwright-community/playwright-go/cmd/playwright install
```

**Running E2E tests:**

```bash
# Start the application
docker-compose up -d

# Run E2E tests
go test ./tests/e2e/... -v

# Run specific test
go test ./tests/e2e/... -run TestUserRegistration -v
```

## Frontend Tests

### Component Tests

Test Svelte components using Testing Library.

**Example: Testing UserList component**

```typescript
import { render, screen, fireEvent } from '@testing-library/svelte';
import UserList from '$lib/components/UserList.svelte';

describe('UserList Component', () => {
  it('renders user list', () => {
    const users = [{ id: '1', name: 'John Doe' }];
    render(UserList, { users });
    
    expect(screen.getByText('John Doe')).toBeInTheDocument();
  });
  
  it('calls delete handler', async () => {
    const onDelete = vi.fn();
    render(UserList, { users, onDelete });
    
    await fireEvent.click(screen.getByRole('button', { name: /delete/i }));
    
    expect(onDelete).toHaveBeenCalled();
  });
});
```

### Store Tests

Test Svelte stores in isolation.

**Example: Testing userStore**

```typescript
import { get } from 'svelte/store';
import { userStore } from '$stores/userStore';

describe('userStore', () => {
  beforeEach(() => {
    userStore.reset();
  });
  
  it('loads users successfully', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ users: mockUsers })
    });
    
    await userStore.loadUsers();
    
    const state = get(userStore);
    expect(state.users).toEqual(mockUsers);
  });
});
```

### Mocking API Calls

Use MSW (Mock Service Worker) for API mocking:

```typescript
import { rest } from 'msw';
import { setupServer } from 'msw/node';

const server = setupServer(
  rest.get('/api/users', (req, res, ctx) => {
    return res(ctx.json({ users: mockUsers }));
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

## Test Coverage

### Generating Reports

```bash
# Backend coverage
./scripts/test-backend.sh --coverage
open coverage/coverage.html

# Frontend coverage
./scripts/test-frontend.sh --coverage
open frontend/coverage/index.html

# All coverage
./scripts/test-all.sh --coverage
```

### Coverage Thresholds

The test harness enforces minimum coverage thresholds:

**Backend:**
- Lines: 70%
- Functions: 70%
- Branches: 60%

**Frontend (Vitest):**
- Lines: 70%
- Functions: 70%
- Branches: 70%
- Statements: 70%

### Viewing Coverage

```bash
# Backend - View in browser
go tool cover -html=coverage/unit.out

# Frontend - View in browser
cd frontend && npm run test:coverage
# Opens frontend/coverage/index.html
```

## CI/CD Integration

### GitHub Actions

Create `.github/workflows/test.yml`:

```yaml
name: Tests

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
      
      - name: Set up Node
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Run tests
        run: |
          chmod +x scripts/test-all.sh
          ./scripts/test-all.sh --all --coverage
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/unit.out,./frontend/coverage/lcov.info
```

### GitLab CI

Create `.gitlab-ci.yml`:

```yaml
stages:
  - test

test:
  stage: test
  image: golang:1.21
  services:
    - postgres:15
  script:
    - ./scripts/setup-test-harness.sh
    - ./scripts/test-all.sh --all --coverage
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage/cobertura.xml
```

## Best Practices

### Writing Tests

1. **Follow AAA Pattern**: Arrange, Act, Assert
2. **Use descriptive test names**: `TestGetUser_WithInvalidID_ReturnsNotFound`
3. **Test one thing per test**: Keep tests focused
4. **Use table-driven tests** for multiple scenarios
5. **Mock external dependencies**: Don't call real APIs or databases in unit tests
6. **Clean up after tests**: Use `defer` or `afterEach` for cleanup

### Backend Best Practices

```go
// Good: Table-driven test
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"invalid format", "invalid", true},
        {"empty string", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Frontend Best Practices

```typescript
// Good: Use custom render function
function renderWithProviders(component, props = {}) {
  return render(component, {
    ...props,
    context: new Map([
      ['authStore', mockAuthStore],
      ['userStore', mockUserStore]
    ])
  });
}

// Good: Use data-testid for complex selectors
<button data-testid="submit-button">Submit</button>
screen.getByTestId('submit-button')
```

### Continuous Testing

```bash
# Backend watch mode (requires entr or similar)
find . -name "*.go" | entr -c go test ./... -short

# Frontend watch mode
cd frontend && npm run test:watch
```

## Troubleshooting

### Common Issues

**Backend:**

```bash
# Reset test database
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d db

# Clear Go test cache
go clean -testcache

# Update dependencies
go mod tidy
go mod download
```

**Frontend:**

```bash
# Clear node modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Vitest cache
rm -rf node_modules/.vitest
```

### Debug Mode

```bash
# Backend with verbose output
go test -v -race ./... -run TestName

# Frontend with debug
DEBUG=* npm run test -- --reporter=verbose
```

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Vitest Documentation](https://vitest.dev/)
- [Testing Library](https://testing-library.com/)
- [Playwright Go](https://playwright.dev/go/)

## Contributing

When adding new features:

1. Write tests first (TDD approach)
2. Ensure tests pass locally
3. Maintain coverage thresholds
4. Update test documentation as needed

## License

Same as hub-hrms project license.
