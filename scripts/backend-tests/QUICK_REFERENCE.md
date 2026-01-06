# hub-hrms Backend Tests - Quick Reference Card

## 🚀 Installation (3 Steps)

```powershell
# 1. Extract
Expand-Archive backend-tests.zip

# 2. Install
cd backend-tests
.\install-tests.ps1 -BackendPath ..\backend

# 3. Fix & Test
# Fix timesheet_service.go line 301: change %w to %s
cd ..\backend
go test ./...
```

## 📋 Common Commands

### Run Tests
```powershell
go test ./...                          # All tests
go test -v ./...                       # Verbose
go test ./internal/service -v          # Specific package
go test -run TestAuthService_Login -v  # Specific test
```

### Coverage
```powershell
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
go tool cover -func=coverage.out | Select-String "total:"
```

### Benchmarks
```powershell
go test -bench=. -benchmem ./internal/service
go test -bench=. -benchmem ./internal/api
```

## 🐛 Quick Fixes

### Compilation Error
**File:** `backend/internal/service/timesheet_service.go`  
**Line:** 301  
**Change:** `%w` → `%s`

```go
// WRONG
return nil, fmt.Errorf("description: %w", errorMsg)

// CORRECT
return nil, fmt.Errorf("description: %s", errorMsg)
```

### Import Errors
```powershell
cd backend
go mod tidy
go get -u github.com/stretchr/testify/assert
go get -u github.com/stretchr/testify/mock
```

### Race Detection Error (Windows)
```powershell
# Don't use -race flag on Windows without CGO
go test ./...  # Instead of: go test -race ./...
```

## 📊 What Gets Tested

| Package | Tests | Coverage |
|---------|-------|----------|
| config | 15+ | 90%+ |
| models | 10+ | 85%+ |
| service | 30+ | 80%+ |
| api | 25+ | 75%+ |

## 📁 File Locations

```
backend/
├── internal/
│   ├── config/
│   │   └── config_test.go      ← Tests config loading
│   ├── models/
│   │   └── models_test.go      ← Tests models & JSON
│   ├── service/
│   │   └── service_test.go     ← Tests auth & user service
│   └── api/
│       └── api_test.go         ← Tests HTTP handlers
```

## ✅ Verification

```powershell
# Should show tests running, not [no test files]
go test ./... -v

# Check coverage
go test -cover ./...

# Expected output:
# ok    hub-hrms/backend/internal/config    0.123s  coverage: 92.0%
# ok    hub-hrms/backend/internal/models    0.089s  coverage: 87.5%
# ok    hub-hrms/backend/internal/service   0.234s  coverage: 81.3%
# ok    hub-hrms/backend/internal/api       0.156s  coverage: 76.8%
```

## 🎯 Success Criteria

- ✅ `go test ./...` passes
- ✅ No `[no test files]` messages
- ✅ Coverage > 75% overall
- ✅ No compilation errors
- ✅ All imports resolved

## 📚 Documentation

- **IMPLEMENTATION_SUMMARY.md** - Complete overview
- **README.md** - Detailed guide
- **FIX_TIMESHEET_ERROR.md** - Error fix

## 🆘 Help

**Tests won't run?**
1. Check you're in `backend/` directory
2. Run `go mod tidy`
3. Verify Go 1.21+ installed

**Import errors?**
- Check module path in go.mod matches test imports

**Coverage low?**
- Tests are for core packages only
- Add tests for other packages as needed

## ⚡ Pro Tips

```powershell
# Clean cache if tests behave oddly
go clean -testcache

# Format before committing
go fmt ./...

# Run specific test pattern
go test -run User ./...

# Skip long tests
go test -short ./...

# Parallel execution
go test -parallel 4 ./...

# JSON output for CI/CD
go test -json ./...
```

## 🔗 Quick Links

- Go Testing: https://golang.org/pkg/testing/
- Testify: https://github.com/stretchr/testify
- Coverage: https://go.dev/blog/cover

---
**Created:** 2026-01-06  
**Version:** 1.0  
**Coverage:** config, models, service, api
