# Test Harness Quick Start Guide

## 🚀 5-Minute Setup

### 1. Initial Setup (One Time)

```bash
# Clone your repo and navigate to it
cd hub-hrms

# Copy test harness files to your project
# (copy all files from this test harness package)

# Run setup script
chmod +x scripts/*.sh
./scripts/setup-test-harness.sh
```

### 2. Run Tests

```bash
# Quick test (unit tests only)
./scripts/test-all.sh

# Full test suite (recommended before commits)
./scripts/test-all.sh --all --coverage

# Watch mode for development
cd frontend && npm run test:watch
```

## 📋 Common Commands

### Backend

```bash
# Unit tests
go test ./... -short

# With coverage
go test -coverprofile=coverage.out ./... -short
go tool cover -html=coverage.out

# Specific package
go test ./internal/handlers -v

# Integration tests
./scripts/test-backend.sh --integration
```

### Frontend

```bash
cd frontend

# Run tests once
npm run test

# Watch mode
npm run test:watch

# Coverage
npm run test:coverage

# Interactive UI
npm run test:ui
```

### Full Suite

```bash
# Everything with coverage
./scripts/test-all.sh --all --coverage

# Just unit + frontend (fast)
./scripts/test-all.sh

# Only integration tests
./scripts/test-all.sh --integration

# Only E2E tests
./scripts/test-all.sh --e2e
```

## 📁 Project Structure After Setup

```
hub-hrms/
├── scripts/
│   ├── setup-test-harness.sh     # Initial setup
│   ├── test-backend.sh            # Backend test runner
│   ├── test-frontend.sh           # Frontend test runner
│   └── test-all.sh                # Master test runner
├── tests/
│   ├── unit/                      # Backend unit tests
│   ├── integration/               # Backend integration tests
│   └── e2e/                       # E2E tests
├── frontend/
│   ├── vitest.config.ts           # Vitest configuration
│   └── src/
│       └── tests/
│           ├── setup.ts           # Test setup
│           ├── components/        # Component tests
│           └── stores/            # Store tests
├── coverage/                      # Coverage reports
├── docker-compose.test.yml        # Test environment
└── .github/
    └── workflows/
        └── test.yml               # CI/CD pipeline
```

## 🔧 Writing Your First Test

### Backend Handler Test

```go
// internal/handlers/user_handler_test.go
package handlers

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
    // Arrange
    mockService := new(MockUserService)
    handler := NewUserHandler(mockService)
    mockService.On("GetUser", "123").Return(expectedUser, nil)
    
    // Act
    result, err := handler.GetUser("123")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.ID, result.ID)
    mockService.AssertExpectations(t)
}
```

### Frontend Component Test

```typescript
// frontend/src/tests/components/Button.test.ts
import { render, screen, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import Button from '$lib/components/Button.svelte';

describe('Button Component', () => {
  it('calls onClick when clicked', async () => {
    const onClick = vi.fn();
    render(Button, { onClick, label: 'Click me' });
    
    await fireEvent.click(screen.getByText('Click me'));
    
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});
```

## 🐛 Troubleshooting

### Tests Won't Run

```bash
# Clear caches
go clean -testcache
rm -rf node_modules/.vitest

# Reinstall dependencies
go mod download
cd frontend && npm ci
```

### Database Issues

```bash
# Reset test database
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d db
```

### Port Conflicts

```bash
# Check what's using port 5432
lsof -i :5432

# Use different port in docker-compose.test.yml
ports:
  - "5433:5432"  # Changed from 5432:5432
```

## 📊 Coverage Thresholds

**Minimum Requirements:**
- Backend: 70% coverage
- Frontend: 70% coverage
- Integration: 60% coverage

**View Coverage:**
```bash
# Backend
open coverage/coverage.html

# Frontend
open frontend/coverage/index.html
```

## 🔄 CI/CD Integration

The test harness includes GitHub Actions workflow:

```bash
# Copy to your project
cp github-actions-test.yml .github/workflows/test.yml

# Commit and push
git add .github/workflows/test.yml
git commit -m "Add test CI/CD pipeline"
git push
```

Tests will automatically run on:
- Every push to main/develop
- Every pull request
- Manual workflow dispatch

## 🎯 Best Practices

1. **Run tests before committing**
   ```bash
   ./scripts/test-all.sh --all
   ```

2. **Use watch mode during development**
   ```bash
   cd frontend && npm run test:watch
   ```

3. **Write tests first (TDD)**
   - Write test
   - See it fail
   - Implement feature
   - See it pass

4. **Keep tests fast**
   - Mock external dependencies
   - Use in-memory databases for unit tests
   - Run integration tests separately

5. **Test one thing at a time**
   - Each test should verify a single behavior
   - Use descriptive test names

## 📚 Additional Resources

- **Full Documentation**: See TEST_HARNESS_README.md
- **Backend Testing**: See handlers_test.go and services_test.go examples
- **Frontend Testing**: See UserList.test.ts and stores.test.ts examples
- **Integration Tests**: See integration_test.go example
- **E2E Tests**: See e2e_test.go example

## 🆘 Getting Help

1. Check TEST_HARNESS_README.md for detailed documentation
2. Look at example test files in the test harness
3. Run with verbose mode: `go test -v` or `npm run test -- --reporter=verbose`
4. Check test output for specific error messages

## ✅ Checklist

Before pushing code:
- [ ] All tests pass: `./scripts/test-all.sh --all`
- [ ] Coverage meets thresholds
- [ ] New features have tests
- [ ] Tests are documented
- [ ] No failing tests are committed

## 🎉 You're Ready!

Your test harness is set up and ready to use. Start by running:

```bash
./scripts/test-all.sh
```

Happy testing! 🧪
