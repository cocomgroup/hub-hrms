# HR System - Initial Setup Script (PowerShell)
# Creates the first admin user and optionally adds sample data

# Encoding: UTF-8 without BOM
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "HR System - Initial Setup" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Check if psql is available
try {
    $null = Get-Command psql -ErrorAction Stop
} catch {
    Write-Host "Error: psql command not found. Please install PostgreSQL client." -ForegroundColor Red
    Write-Host "Download from: https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
    exit 1
}

# Get database connection details
Write-Host "Enter database connection details:" -ForegroundColor Yellow
$DB_HOST = Read-Host "Host [localhost]"
if ([string]::IsNullOrWhiteSpace($DB_HOST)) { $DB_HOST = "localhost" }

$DB_PORT = Read-Host "Port [5432]"
if ([string]::IsNullOrWhiteSpace($DB_PORT)) { $DB_PORT = "5432" }

$DB_NAME = Read-Host "Database [hrapp]"
if ([string]::IsNullOrWhiteSpace($DB_NAME)) { $DB_NAME = "hrmsdb" }

$DB_USER = Read-Host "User [postgres]"
if ([string]::IsNullOrWhiteSpace($DB_USER)) { $DB_USER = "hrms_user" }

$SecurePassword = Read-Host "Password" -AsSecureString
$DB_PASSWORD = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [Runtime.InteropServices.Marshal]::SecureStringToBSTR($SecurePassword)
)

# Test connection
Write-Host ""
Write-Host "Testing database connection..." -ForegroundColor Yellow
$env:PGPASSWORD = $DB_PASSWORD

$testConnection = & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Could not connect to database. Please check your credentials." -ForegroundColor Red
    $env:PGPASSWORD = $null
    exit 1
}

Write-Host "[OK] Database connection successful" -ForegroundColor Green
Write-Host ""

# Check if admin user already exists
$checkAdmin = & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM users WHERE email = 'admin@cocomgroup.com';" 2>&1
$adminExists = [int]($checkAdmin -replace '\s','')

if ($adminExists -gt 0) {
    Write-Host "WARNING: Admin user already exists!" -ForegroundColor Yellow
    $reset = Read-Host "Do you want to reset the admin password? (y/N)"
    
    if ($reset -ne 'y' -and $reset -ne 'Y') {
        Write-Host "Skipping admin user creation." -ForegroundColor Yellow
        $env:PGPASSWORD = $null
        exit 0
    }
    
    # Reset password
    $resetSQL = "UPDATE users SET password_hash = '`$2a`$10`$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy' WHERE email = 'admin@cocomgroup.com';"
    
    $resetSQL | & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Admin password reset to 'admin123'" -ForegroundColor Green
        Write-Host ""
    } else {
        Write-Host "Error resetting password" -ForegroundColor Red
    }
} else {
    # Create admin user
    Write-Host "Creating admin user..." -ForegroundColor Yellow
    
    # Use a temporary SQL file to avoid here-string issues
    $tempSqlFile = [System.IO.Path]::GetTempFileName()
    
    $createAdminSQL = @"
BEGIN;

WITH new_employee AS (
    INSERT INTO employees (
        first_name, last_name, email, phone,
        hire_date, department, position,
        employment_type, status
    ) VALUES (
        'System', 'Administrator', 'admin@cocomgroup.com', '555-0100',
        CURRENT_DATE, 'Administration', 'System Administrator',
        'full-time', 'active'
    ) RETURNING id
),
new_user AS (
    INSERT INTO users (email, password_hash, role, employee_id)
    SELECT 'admin@company.com',
           '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
           'admin', id FROM new_employee
    RETURNING id, employee_id
)
INSERT INTO pto_balances (
    employee_id, vacation_days, sick_days, personal_days,
    accrual_rate_vacation, accrual_rate_sick, last_accrual_date
)
SELECT employee_id, 15.0, 10.0, 5.0, 1.25, 0.83, CURRENT_DATE
FROM new_user;

COMMIT;
"@

    Set-Content -Path $tempSqlFile -Value $createAdminSQL -Encoding UTF8
    
    & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $tempSqlFile 2>&1 | Out-Null
    
    Remove-Item $tempSqlFile -Force
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Admin user created successfully" -ForegroundColor Green
        Write-Host ""
    } else {
        Write-Host "Error creating admin user" -ForegroundColor Red
    }
}

# Display credentials
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Admin User Credentials" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Email:    admin@cocomgroup.com" -ForegroundColor White
Write-Host "Password: admin123" -ForegroundColor White
Write-Host ""
Write-Host "IMPORTANT: Change this password immediately after first login!" -ForegroundColor Yellow
Write-Host ""

# Ask about sample data
$addSamples = Read-Host "Do you want to add sample employees for testing? (y/N)"

if ($addSamples -eq 'y' -or $addSamples -eq 'Y') {
    Write-Host ""
    Write-Host "Creating sample employees..." -ForegroundColor Yellow
    
    $tempSampleFile = [System.IO.Path]::GetTempFileName()
    
    $sampleSQL = @"
-- Sample Employee 1: John Doe
WITH emp1 AS (
    INSERT INTO employees (
        first_name, last_name, email, phone,
        hire_date, department, position, employment_type, status
    ) VALUES (
        'John', 'Doe', 'john.doe@cocomgroup.com', '555-0101',
        CURRENT_DATE - INTERVAL '30 days', 'Engineering', 'Senior Developer',
        'full-time', 'active'
    ) RETURNING id
),
user1 AS (
    INSERT INTO users (email, password_hash, role, employee_id)
    SELECT 'john.doe@cocomgroup.com',
           '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
           'employee', id FROM emp1
    RETURNING employee_id
)
INSERT INTO pto_balances (employee_id, vacation_days, sick_days, personal_days)
SELECT employee_id, 15.0, 10.0, 5.0 FROM user1;

-- Sample Employee 2: Jane Smith
WITH emp2 AS (
    INSERT INTO employees (
        first_name, last_name, email, phone,
        hire_date, department, position, employment_type, status
    ) VALUES (
        'Jane', 'Smith', 'jane.smith@cocomgroup.com', '555-0102',
        CURRENT_DATE - INTERVAL '60 days', 'Human Resources', 'HR Manager',
        'full-time', 'active'
    ) RETURNING id
),
user2 AS (
    INSERT INTO users (email, password_hash, role, employee_id)
    SELECT 'jane.smith@cocomgroup.com',
           '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
           'manager', id FROM emp2
    RETURNING employee_id
)
INSERT INTO pto_balances (employee_id, vacation_days, sick_days, personal_days)
SELECT employee_id, 15.0, 10.0, 5.0 FROM user2;

-- Sample Employee 3: Bob Johnson (new hire)
WITH emp3 AS (
    INSERT INTO employees (
        first_name, last_name, email, phone,
        hire_date, department, position, employment_type, status
    ) VALUES (
        'Bob', 'Johnson', 'bob.johnson@cocomgroup.com', '555-0103',
        CURRENT_DATE, 'Sales', 'Sales Representative',
        'full-time', 'active'
    ) RETURNING id
),
user3 AS (
    INSERT INTO users (email, password_hash, role, employee_id)
    SELECT 'bob.johnson@cocomgroup.com',
           '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
           'employee', id FROM emp3
    RETURNING employee_id
)
INSERT INTO pto_balances (employee_id, vacation_days, sick_days, personal_days)
SELECT employee_id, 15.0, 10.0, 5.0 FROM user3;
"@

    Set-Content -Path $tempSampleFile -Value $sampleSQL -Encoding UTF8
    
    & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $tempSampleFile 2>&1 | Out-Null
    
    Remove-Item $tempSampleFile -Force
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Sample employees created:" -ForegroundColor Green
        Write-Host "  - john.doe@cocomgroup.com (password: admin123)" -ForegroundColor White
        Write-Host "  - jane.smith@cocomgroup.com (password: admin123)" -ForegroundColor White
        Write-Host "  - bob.johnson@cocomgroup.com (password: admin123)" -ForegroundColor White
        Write-Host ""
    } else {
        Write-Host "Error creating sample employees" -ForegroundColor Red
    }
}

# Verify setup
Write-Host "Verifying setup..." -ForegroundColor Yellow
$totalUsers = & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM users;" 2>&1
$userCount = [int]($totalUsers -replace '\s','')

Write-Host "[OK] Total users in system: $userCount" -ForegroundColor Green
Write-Host ""

# Display next steps
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Setup Complete!" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Start the backend server:" -ForegroundColor White
Write-Host "   cd backend" -ForegroundColor Gray
Write-Host "   go run cmd/main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Start the frontend:" -ForegroundColor White
Write-Host "   cd frontend" -ForegroundColor Gray
Write-Host "   npm run dev" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Access the application:" -ForegroundColor White
Write-Host "   http://localhost:5173" -ForegroundColor Cyan
Write-Host ""
Write-Host "4. Login with admin credentials" -ForegroundColor White
Write-Host "5. CHANGE THE DEFAULT PASSWORD!" -ForegroundColor Yellow
Write-Host ""
Write-Host "For more information, see ONBOARDING_GUIDE.md" -ForegroundColor White
Write-Host ""

# Clear password from environment
$env:PGPASSWORD = $null
$DB_PASSWORD = $null
