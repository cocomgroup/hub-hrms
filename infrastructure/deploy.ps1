# HR Workflow System - Complete Stack Deployment Script
# Windows PowerShell

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "HR Workflow System - Stack Deployment" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$DB_NAME = "hub_hrms"
$DB_USER = "postgres"
$BACKEND_PORT = 8080
$FRONTEND_PORT = 5173
$PROJECT_ROOT = $PSScriptRoot

# Check if running as Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "⚠️  Warning: Not running as Administrator. Some operations may fail." -ForegroundColor Yellow
    Write-Host ""
}

# Function to check if a command exists
function Test-Command($command) {
    try {
        Get-Command $command -ErrorAction Stop | Out-Null
        return $true
    } catch {
        return $false
    }
}

# Step 1: Check Prerequisites
Write-Host "Step 1: Checking Prerequisites..." -ForegroundColor Green
Write-Host "--------------------------------" -ForegroundColor Green

$prereqsMet = $true

# Check PostgreSQL
if (Test-Command "psql") {
    Write-Host "✓ PostgreSQL found" -ForegroundColor Green
} else {
    Write-Host "✗ PostgreSQL not found" -ForegroundColor Red
    Write-Host "  Install from: https://www.postgresql.org/download/" -ForegroundColor Yellow
    $prereqsMet = $false
}

# Check Go
if (Test-Command "go") {
    $goVersion = go version
    Write-Host "✓ Go found: $goVersion" -ForegroundColor Green
} else {
    Write-Host "✗ Go not found" -ForegroundColor Red
    Write-Host "  Install from: https://go.dev/dl/" -ForegroundColor Yellow
    $prereqsMet = $false
}

# Check Node.js
if (Test-Command "node") {
    $nodeVersion = node --version
    Write-Host "✓ Node.js found: $nodeVersion" -ForegroundColor Green
} else {
    Write-Host "✗ Node.js not found" -ForegroundColor Red
    Write-Host "  Install from: https://nodejs.org/" -ForegroundColor Yellow
    $prereqsMet = $false
}

# Check npm
if (Test-Command "npm") {
    $npmVersion = npm --version
    Write-Host "✓ npm found: v$npmVersion" -ForegroundColor Green
} else {
    Write-Host "✗ npm not found" -ForegroundColor Red
    $prereqsMet = $false
}

if (-not $prereqsMet) {
    Write-Host ""
    Write-Host "❌ Prerequisites not met. Please install missing components." -ForegroundColor Red
    exit 1
}

Write-Host ""

# Step 2: Database Setup
Write-Host "Step 2: Database Setup..." -ForegroundColor Green
Write-Host "------------------------" -ForegroundColor Green

# Check if database exists
$dbExists = psql -U $DB_USER -lqt | Select-String -Pattern $DB_NAME -Quiet

if ($dbExists) {
    Write-Host "⚠️  Database '$DB_NAME' already exists" -ForegroundColor Yellow
    $response = Read-Host "Do you want to recreate it? (yes/no)"
    
    if ($response -eq "yes") {
        Write-Host "Dropping existing database..." -ForegroundColor Yellow
        psql -U $DB_USER -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>&1 | Out-Null
        Write-Host "✓ Database dropped" -ForegroundColor Green
    } else {
        Write-Host "Using existing database" -ForegroundColor Yellow
    }
}

if (-not $dbExists -or $response -eq "yes") {
    Write-Host "Creating database '$DB_NAME'..." -ForegroundColor Cyan
    psql -U $DB_USER -c "CREATE DATABASE $DB_NAME;" 2>&1 | Out-Null
    Write-Host "✓ Database created" -ForegroundColor Green
}

# Enable extensions
Write-Host "Enabling PostgreSQL extensions..." -ForegroundColor Cyan
psql -U $DB_USER -d $DB_NAME -c "CREATE EXTENSION IF NOT EXISTS `"uuid-ossp`"; CREATE EXTENSION IF NOT EXISTS `"pgcrypto`";" 2>&1 | Out-Null
Write-Host "✓ Extensions enabled" -ForegroundColor Green

Write-Host ""

# Step 3: Backend Setup
Write-Host "Step 3: Backend Setup..." -ForegroundColor Green
Write-Host "-----------------------" -ForegroundColor Green

Set-Location "$PROJECT_ROOT\backend"

# Check for .env file
if (-not (Test-Path ".env")) {
    Write-Host "Creating .env file..." -ForegroundColor Cyan
    @"
DB_HOST=localhost
DB_PORT=5432
DB_USER=$DB_USER
DB_PASSWORD=
DB_NAME=$DB_NAME
JWT_SECRET=$(New-Guid)
PORT=$BACKEND_PORT
"@ | Out-File -FilePath ".env" -Encoding utf8
    Write-Host "✓ .env file created" -ForegroundColor Green
} else {
    Write-Host "✓ .env file exists" -ForegroundColor Green
}

# Download Go dependencies
Write-Host "Downloading Go dependencies..." -ForegroundColor Cyan
go mod download
Write-Host "✓ Dependencies downloaded" -ForegroundColor Green

# Run migrations
Write-Host "Running database migrations..." -ForegroundColor Cyan
go run cmd/main.go migrate
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Migrations completed" -ForegroundColor Green
} else {
    Write-Host "✗ Migration failed" -ForegroundColor Red
    exit 1
}

# Generate password hash for admin
Write-Host "Generating admin password..." -ForegroundColor Cyan
$adminHash = go run scripts/hash_password.go admin123 | Select-String -Pattern '\$2a\$10\$.*'
if ($adminHash) {
    Write-Host "✓ Admin password generated" -ForegroundColor Green
    
    # Insert admin user
    Write-Host "Creating admin user..." -ForegroundColor Cyan
    $adminId = [guid]::NewGuid().ToString()
    $insertAdminSQL = @"
INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
VALUES ('$adminId', 'admin@cocomgroup.com', '$adminHash', 'admin', NOW(), NOW())
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;
"@
    
    $insertAdminSQL | psql -U $DB_USER -d $DB_NAME 2>&1 | Out-Null
    Write-Host "✓ Admin user created" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to generate admin password" -ForegroundColor Red
}

# Insert sample employees
Write-Host "Creating sample employees..." -ForegroundColor Cyan
$insertEmployeesSQL = @"
INSERT INTO employees (first_name, last_name, email, phone, hire_date, department, position, status, created_at, updated_at)
VALUES 
('Evan', 'Hunt', 'evan.hunt@cocomgroup.com', '555-0101', '2020-01-15', 'Engineering', 'CTO', 'active', NOW(), NOW()),
('Bob', 'Johnson', 'bob.johnson@cocomgroup.com', '555-0102', '2021-03-20', 'Sales', 'Sales Representative', 'active', NOW(), NOW()),
('Jane', 'Smith', 'jane.smith@cocomgroup.com', '555-0103', '2021-06-10', 'Human Resources', 'HR Manager', 'active', NOW(), NOW())
ON CONFLICT (email) DO NOTHING;
"@

$insertEmployeesSQL | psql -U $DB_USER -d $DB_NAME 2>&1 | Out-Null
Write-Host "✓ Sample employees created" -ForegroundColor Green

Write-Host ""

# Step 4: Frontend Setup
Write-Host "Step 4: Frontend Setup..." -ForegroundColor Green
Write-Host "------------------------" -ForegroundColor Green

Set-Location "$PROJECT_ROOT\frontend"

# Check if node_modules exists
if (-not (Test-Path "node_modules")) {
    Write-Host "Installing npm packages..." -ForegroundColor Cyan
    npm install
    Write-Host "✓ npm packages installed" -ForegroundColor Green
} else {
    Write-Host "✓ node_modules exists" -ForegroundColor Green
}

Write-Host ""

# Step 5: Start Services
Write-Host "Step 5: Starting Services..." -ForegroundColor Green
Write-Host "---------------------------" -ForegroundColor Green

# Start backend in new window
Write-Host "Starting backend server on port $BACKEND_PORT..." -ForegroundColor Cyan
Set-Location "$PROJECT_ROOT\backend"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PROJECT_ROOT\backend'; Write-Host 'Backend Server' -ForegroundColor Green; go run cmd/main.go"
Write-Host "✓ Backend started in new window" -ForegroundColor Green

# Wait a moment for backend to start
Start-Sleep -Seconds 3

# Start frontend in new window
Write-Host "Starting frontend dev server on port $FRONTEND_PORT..." -ForegroundColor Cyan
Set-Location "$PROJECT_ROOT\frontend"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PROJECT_ROOT\frontend'; Write-Host 'Frontend Dev Server' -ForegroundColor Green; npm run dev"
Write-Host "✓ Frontend started in new window" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✅ Deployment Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "🌐 Application URLs:" -ForegroundColor Yellow
Write-Host "   Frontend: http://localhost:$FRONTEND_PORT" -ForegroundColor White
Write-Host "   Backend:  http://localhost:$BACKEND_PORT" -ForegroundColor White
Write-Host ""
Write-Host "👤 Default Admin Credentials:" -ForegroundColor Yellow
Write-Host "   Email:    admin@cocomgroup.com" -ForegroundColor White
Write-Host "   Password: admin123" -ForegroundColor White
Write-Host ""
Write-Host "📊 Sample Employees:" -ForegroundColor Yellow
Write-Host "   - Evan Hunt (CTO)" -ForegroundColor White
Write-Host "   - Bob Johnson (Sales Representative)" -ForegroundColor White
Write-Host "   - Jane Smith (HR Manager)" -ForegroundColor White
Write-Host ""
Write-Host "🔧 Management Commands:" -ForegroundColor Yellow
Write-Host "   View logs: Check the opened PowerShell windows" -ForegroundColor White
Write-Host "   Stop services: Close the PowerShell windows" -ForegroundColor White
Write-Host "   Database: psql -U $DB_USER -d $DB_NAME" -ForegroundColor White
Write-Host ""
Write-Host "Press any key to open the application in your browser..." -ForegroundColor Cyan
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

Start-Process "http://localhost:$FRONTEND_PORT"

Write-Host ""
Write-Host "✨ Enjoy your HR Workflow System!" -ForegroundColor Green