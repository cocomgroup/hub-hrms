# hub-hrms Test Harness

Complete testing infrastructure for your HRMS application with backend (Go) and frontend (Svelte) test suites.

## 📚 Documentation Index

Start here to get up and running quickly:

1. **[INSTALLATION.md](INSTALLATION.md)** - Step-by-step installation guide
2. **[QUICK_START.md](QUICK_START.md)** - 5-minute quick start guide
3. **[TEST_HARNESS_README.md](TEST_HARNESS_README.md)** - Complete documentation

## 🎯 Quick Overview

### What You Get

- ✅ **Backend Testing**
  - Unit tests with mocks (Testify, Sqlmock)
  - Integration tests with real database
  - E2E tests with Playwright
  - Coverage reporting

- ✅ **Frontend Testing**
  - Component tests (Testing Library)
  - Store/state tests (Vitest)
  - Mock API calls (MSW)
  - Coverage reporting

- ✅ **Automation**
  - Test runner scripts
  - Docker test environment
  - GitHub Actions CI/CD
  - Coverage thresholds

## 🚀 Quick Commands

```bash
# Install (one time)
./scripts/setup-test-harness.sh

# Run all tests
./scripts/test-all.sh

# Run with coverage
./scripts/test-all.sh --coverage

# Run everything (unit + integration + E2E)
./scripts/test-all.sh --all --coverage

# Watch mode for development
cd frontend && npm run test:watch
```

## 📁 Package Contents

```
test-harness/
├── INSTALLATION.md              # Installation guide
├── QUICK_START.md              # Quick start guide
├── TEST_HARNESS_README.md      # Full documentation
│
├── scripts/
│   ├── setup-test-harness.sh   # Install dependencies
│   ├── test-backend.sh         # Backend test runner
│   ├── test-frontend.sh        # Frontend test runner
│   └── test-all.sh            # Master test runner
│
├── examples/
│   ├── handlers_test.go        # Handler test examples
│   ├── services_test.go        # Service test examples
│   ├── integration_test.go     # Integration test examples
│   └── e2e_test.go            # E2E test examples
│
├── frontend/
│   ├── vitest.config.ts        # Vitest configuration
│   ├── test-setup.ts           # Test setup file
│   ├── UserList.test.ts        # Component test example
│   ├── stores.test.ts          # Store test example
│   └── package.json            # NPM scripts
│
├── .github/
│   └── workflows/
│       └── test.yml           # GitHub Actions workflow
│
└── docker-compose.test.yml     # Test environment
```

## 🏃 Getting Started

### 1. Install

```bash
cd /path/to/hub-hrms
cp -r /path/to/test-harness/* .
./scripts/setup-test-harness.sh
```

### 2. Run Tests

```bash
# Quick test
./scripts/test-all.sh

# Full test suite
./scripts/test-all.sh --all --coverage
```

### 3. Write Your First Test

**Backend (Go):**
```go
func TestGetUser(t *testing.T) {
    // See examples/handlers_test.go
    mockService := new(MockUserService)
    mockService.On("GetUser", "123").Return(user, nil)
    
    result, err := service.GetUser("123")
    
    assert.NoError(t, err)
    assert.Equal(t, user.ID, result.ID)
}
```

**Frontend (TypeScript):**
```typescript
// See examples in frontend/UserList.test.ts
describe('Button', () => {
  it('calls onClick when clicked', async () => {
    const onClick = vi.fn();
    render(Button, { onClick });
    
    await fireEvent.click(screen.getByRole('button'));
    
    expect(onClick).toHaveBeenCalled();
  });
});
```

## 📊 Test Coverage

The test harness enforces coverage thresholds:
- Backend: 70% coverage required
- Frontend: 70% coverage required
- Integration: 60% coverage required

View coverage reports:
```bash
# Generate reports
./scripts/test-all.sh --coverage

# View backend
open coverage/coverage.html

# View frontend
open frontend/coverage/index.html
```

## 🔄 CI/CD Integration

GitHub Actions workflow included:
- Runs on push to main/develop
- Runs on pull requests
- Generates coverage reports
- Comments on PRs with coverage info

```bash
# Setup CI/CD
cp .github/workflows/test.yml .github/workflows/
git add .github/workflows/test.yml
git commit -m "Add test CI/CD"
git push
```

## 🧪 Test Types

### Backend

1. **Unit Tests** - Test individual functions with mocks
   ```bash
   go test ./... -short
   ```

2. **Integration Tests** - Test with real database
   ```bash
   ./scripts/test-backend.sh --integration
   ```

3. **E2E Tests** - Test complete workflows
   ```bash
   ./scripts/test-backend.sh --e2e
   ```

### Frontend

1. **Component Tests** - Test Svelte components
   ```bash
   npm run test
   ```

2. **Store Tests** - Test state management
   ```bash
   npm run test -- stores
   ```

3. **Watch Mode** - Continuous testing during development
   ```bash
   npm run test:watch
   ```

## 📖 Documentation Guide

**For New Users:**
1. Read [INSTALLATION.md](INSTALLATION.md) first
2. Follow [QUICK_START.md](QUICK_START.md) for basics
3. Reference [TEST_HARNESS_README.md](TEST_HARNESS_README.md) as needed

**For Developers:**
1. Check example test files in `examples/`
2. Review test scripts in `scripts/`
3. Understand test setup in `frontend/test-setup.ts`

**For DevOps:**
1. Review `.github/workflows/test.yml`
2. Configure `docker-compose.test.yml`
3. Set up coverage reporting services

## 🎯 Best Practices

1. **Write Tests First** (TDD approach)
2. **Keep Tests Fast** - Mock external dependencies
3. **Test One Thing** - Each test should verify one behavior
4. **Use Descriptive Names** - `TestGetUser_WithInvalidID_ReturnsError`
5. **Run Tests Often** - Use watch mode during development
6. **Maintain Coverage** - Keep above 70% threshold

## 🛠️ Common Tasks

### Add a New Test

**Backend:**
```bash
# Create test file
touch internal/handlers/new_feature_test.go

# Write test (see examples/handlers_test.go)
# Run test
go test ./internal/handlers -run TestNewFeature -v
```

**Frontend:**
```bash
# Create test file
touch frontend/src/lib/components/NewComponent.test.ts

# Write test (see frontend/UserList.test.ts)
# Run test
npm run test -- NewComponent
```

### Debug Failing Tests

**Backend:**
```bash
# Verbose output
go test -v ./internal/handlers

# Run specific test
go test -run TestGetUser -v

# With race detection
go test -race ./...
```

**Frontend:**
```bash
# Verbose output
npm run test -- --reporter=verbose

# Run specific test
npm run test -- UserList

# Watch mode for debugging
npm run test:watch
```

### Update Coverage Thresholds

Edit configuration files:
- Backend: Add to test scripts
- Frontend: Edit `vitest.config.ts`

```typescript
// vitest.config.ts
coverage: {
  thresholds: {
    lines: 80,      // Changed from 70
    functions: 80,
    branches: 75,
    statements: 80
  }
}
```

## 🐛 Troubleshooting

### Tests Won't Run
```bash
go clean -testcache
rm -rf node_modules/.vitest
npm ci
```

### Database Connection Issues
```bash
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d db
```

### Port Conflicts
Edit `docker-compose.test.yml` to use different ports.

## 📞 Support

- **Documentation Issues**: Check TEST_HARNESS_README.md
- **Test Examples**: Review files in `examples/` directory
- **Configuration**: See `vitest.config.ts` and test scripts
- **CI/CD**: Review `.github/workflows/test.yml`

## 🎉 Success Metrics

Your test harness is working correctly when:

- ✅ All scripts run without errors
- ✅ Coverage reports generate successfully
- ✅ Tests run in CI/CD pipeline
- ✅ Coverage thresholds are met
- ✅ Tests run quickly (<5 min for unit tests)
- ✅ Watch mode works for development

## 🚦 Next Steps

1. **Install the test harness** - Follow INSTALLATION.md
2. **Run initial tests** - `./scripts/test-all.sh`
3. **Review examples** - Check `examples/` directory
4. **Write your first test** - Pick a handler or component
5. **Set up CI/CD** - Configure GitHub Actions
6. **Maintain tests** - Keep coverage above thresholds

## 📝 Version

Test Harness v1.0.0 for hub-hrms
- Backend: Go 1.21+ with Testify, Sqlmock
- Frontend: Vitest, Testing Library, MSW
- E2E: Playwright Go
- CI/CD: GitHub Actions

---

**Ready to start testing?**

```bash
./scripts/setup-test-harness.sh
./scripts/test-all.sh
```

Happy testing! 🧪
