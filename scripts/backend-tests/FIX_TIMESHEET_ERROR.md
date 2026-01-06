# Fix for timesheet_service.go Line 301

## The Error
```
internal\service\timesheet_service.go:301:26: non-constant format string in call to fmt.Errorf
```

## Current Code (WRONG)
```go
return nil, fmt.Errorf("description: %w",errorMsg)
```

## Problem
1. `errorMsg` is a variable, not an error type
2. Using `%w` requires an error type for error wrapping
3. `errorMsg` is a string, so you should use `%s`

## Fixed Code (CORRECT)

### Option 1: Use %s for string
```go
return nil, fmt.Errorf("description: %s", errorMsg)
```

### Option 2: Use errors.New
```go
return nil, errors.New(errorMsg)
```

### Option 3: Better error handling (RECOMMENDED)
```go
if len(processedEntries) == 0 {
    if len(errors) > 0 {
        return nil, fmt.Errorf("no entries were processed successfully. Errors: %v", errors)
    }
    return nil, errors.New("no entries were processed successfully")
}
```

## How to Fix

Open `backend/internal/service/timesheet_service.go` and replace line 301:

**Before:**
```go
return nil, fmt.Errorf("description: %w",errorMsg)
```

**After:**
```go
return nil, fmt.Errorf("description: %s", errorMsg)
```

Or use the recommended version above.

## Quick Fix Command (PowerShell)

```powershell
# Navigate to backend
cd backend

# Option 1: Manual edit
code internal\service\timesheet_service.go
# Go to line 301 and change %w to %s

# Option 2: Automated fix (if you have sed or similar)
# Replace the line programmatically

# Then test it compiles
go build ./...
```

## Why This Happens

The `%w` format verb in `fmt.Errorf` is specifically for **error wrapping** (introduced in Go 1.13). It expects an error type, not a string.

**Valid uses of %w:**
```go
return fmt.Errorf("failed to process: %w", err)  // err is error type
```

**Invalid uses of %w:**
```go
return fmt.Errorf("failed: %w", errorMsg)  // errorMsg is string - ERROR!
```

**For strings, use %s:**
```go
return fmt.Errorf("failed: %s", errorMsg)  // CORRECT
```

## After Fixing

Test that everything compiles:

```powershell
cd backend
go build ./...

# Then run tests
go test ./... -short
```
