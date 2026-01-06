# Test Harness Installation Instructions

## 📦 What's Included

This test harness provides comprehensive testing infrastructure for your hub-hrms application:

- ✅ Backend unit tests (Go + Testify)
- ✅ Backend integration tests (with real database)
- ✅ Backend E2E tests (Playwright)
- ✅ Frontend unit tests (Vitest + Testing Library)
- ✅ Frontend component tests
- ✅ Store/state management tests
- ✅ Coverage reporting
- ✅ CI/CD GitHub Actions workflow
- ✅ Docker test environment

## 📋 Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Git

## 🚀 Installation Steps

### Step 1: Copy Files to Your Project

```bash
# Navigate to your hub-hrms project root
cd /path/to/hub-hrms

# Create scripts directory if it doesn't exist
mkdir -p scripts

# Copy test scripts
cp /path/to/test-harness/scripts/* scripts/
chmod +x scripts/*.sh

# Copy frontend test files
cp /path/to/test-harness/frontend/vitest.config.ts frontend/
cp /path/to/test-harness/frontend/test-setup.ts frontend/src/tests/setup.ts
cp /path/to/test-harness/frontend/package.json frontend/package.json

# Copy Docker compose for tests
cp /path/to/test-harness/docker-compose.test.yml .

# Copy GitHub Actions workflow
mkdir -p .github/workflows
cp /path/to/test-harness/.github/workflows/test.yml .github/workflows/

# Copy documentation
cp /path/to/test-harness/TEST_HARNESS_README.md .
cp /path/to/test-harness/QUICK_START.md .
```

### Step 2: Install Dependencies

```bash
# Run the setup script
./scripts/setup-test-harness.sh
```

This will:
- Install Go testing libraries
- Install Node.js testing dependencies
- Create test directory structure
- Set up test configurations

### Step 3: Directory Structure

After installation, your project should have:

```
hub-hrms/
├── scripts/
│   ├── setup-test-harness.sh
│   ├── test-backend.sh
│   ├── test-frontend.sh
│   └── test-all.sh
├── tests/
│   ├── unit/
│   ├── integration/
│   ├── e2e/
│   ├── fixtures/
│   └── mocks/
├── frontend/
│   ├── vitest.config.ts
│   └── src/
│       └── tests/
│           └── setup.ts
├── docker-compose.test.yml
├── .github/
│   └── workflows/
│       └── test.yml
├── TEST_HARNESS_README.md
└── QUICK_START.md
```

### Step 4: Verify Installation

```bash
# Test backend
go test ./... -short

# Test frontend
cd frontend && npm test

# Run full test suite
./scripts/test-all.sh
```

## 📝 Writing Your First Tests

### Backend Test Example

Create `internal/handlers/user_handler_test.go`:

```go
package handlers

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestGetUser_Success(t *testing.T) {
    // See examples/handlers_test.go for full implementation
    mockService := new(MockUserService)
    handler := NewUserHandler(mockService)
    
    expectedUser := &User{ID: "123", Email: "test@example.com"}
    mockService.On("GetUser", "123").Return(expectedUser, nil)
    
    result, err := handler.GetUser("123")
    
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.ID, result.ID)
    mockService.AssertExpectations(t)
}
```

### Frontend Test Example

Create `frontend/src/lib/components/Button.test.ts`:

```typescript
import { render, screen, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import Button from './Button.svelte';

describe('Button Component', () => {
  it('renders with correct label', () => {
    render(Button, { label: 'Click me' });
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('calls onClick when clicked', async () => {
    const onClick = vi.fn();
    render(Button, { label: 'Click me', onClick });
    
    await fireEvent.click(screen.getByText('Click me'));
    
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});
```

## 🔧 Configuration

### Update package.json (Frontend)

Merge the test scripts from `frontend/package.json` with your existing package.json:

```json
{
  "scripts": {
    "test": "vitest run",
    "test:watch": "vitest",
    "test:ui": "vitest --ui",
    "test:coverage": "vitest run --coverage"
  }
}
```

### Update go.mod (Backend)

Add testing dependencies:

```bash
go get -u github.com/stretchr/testify
go get -u github.com/DATA-DOG/go-sqlmock
go get -u github.com/jarcoal/httpmock
```

## 🎯 Next Steps

1. **Read the Documentation**
   - Start with `QUICK_START.md` for immediate usage
   - Read `TEST_HARNESS_README.md` for comprehensive guide

2. **Review Example Tests**
   - Check `examples/handlers_test.go` for backend unit tests
   - Check `examples/services_test.go` for service layer tests
   - Check `examples/integration_test.go` for integration tests
   - Check `frontend/UserList.test.ts` for component tests
   - Check `frontend/stores.test.ts` for store tests

3. **Write Tests for Existing Code**
   - Start with critical business logic
   - Add tests for new features
   - Aim for 70%+ coverage

4. **Set Up CI/CD**
   - Push `.github/workflows/test.yml` to your repo
   - Configure Codecov or similar for coverage reporting
   - Set up branch protection requiring passing tests

## 🐛 Troubleshooting

### Issue: Go dependencies not installing

```bash
go clean -modcache
go mod download
go get -u github.com/stretchr/testify
```

### Issue: Frontend tests not running

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run test
```

### Issue: Port conflicts in Docker

Edit `docker-compose.test.yml` and change ports:

```yaml
ports:
  - "5433:5432"  # Changed from 5432
```

### Issue: Permission denied on scripts

```bash
chmod +x scripts/*.sh
```

## 📞 Support

- **Documentation**: See TEST_HARNESS_README.md
- **Examples**: Check the `examples/` directory
- **Issues**: Create an issue in your project repository

## ✅ Verification Checklist

After installation:

- [ ] All scripts are executable: `ls -l scripts/`
- [ ] Backend tests run: `go test ./... -short`
- [ ] Frontend tests run: `cd frontend && npm test`
- [ ] Coverage reports generate: `./scripts/test-all.sh --coverage`
- [ ] Docker compose works: `docker-compose -f docker-compose.test.yml up -d`
- [ ] GitHub Actions workflow is present: `.github/workflows/test.yml`

## 🎉 You're Ready!

Your test harness is installed and ready to use. Start testing:

```bash
./scripts/test-all.sh
```

For daily development, use watch mode:

```bash
cd frontend && npm run test:watch
```

Happy testing! 🧪
