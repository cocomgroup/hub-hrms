# Database Initialization Script (PowerShell)
# Run this after the stack is deployed to set up the database schema

[CmdletBinding()]
param(
    [string]$StackName = "hrms-prod",
    [string]$Region = "us-east-1",
    [switch]$Help
)

# Function to display help
function Show-Help {
    Write-Host @"
HRMS Database Initialization Script

Usage: .\init-database.ps1 [options]

Options:
  -StackName <name>  CloudFormation stack name (default: hrms-prod)
  -Region <region>   AWS region (default: us-east-1)
  -Help              Show this help message

Examples:
  .\init-database.ps1
  .\init-database.ps1 -StackName my-hrms -Region us-west-2
"@
    exit 0
}

if ($Help) {
    Show-Help
}

# Colors for output
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Function to get database endpoint from CloudFormation
function Get-DatabaseEndpoint {
    $endpoint = aws cloudformation describe-stacks `
        --stack-name $StackName `
        --region $Region `
        --query 'Stacks[0].Outputs[?OutputKey==`DatabaseEndpoint`].OutputValue' `
        --output text 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to get database endpoint. Make sure the stack is deployed."
        exit 1
    }
    
    return $endpoint.Trim()
}

# Function to get database name
function Get-DatabaseName {
    $dbName = aws cloudformation describe-stacks `
        --stack-name $StackName `
        --region $Region `
        --query 'Stacks[0].Outputs[?OutputKey==`DatabaseName`].OutputValue' `
        --output text 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to get database name"
        exit 1
    }
    
    return $dbName.Trim()
}

# Function to get database credentials
function Get-DatabaseCredentials {
    Write-Info "Retrieving database credentials..."
    
    $dbHost = Get-DatabaseEndpoint
    $dbName = Get-DatabaseName
    
    Write-Host ""
    Write-Host "Database Host: $dbHost" -ForegroundColor Cyan
    Write-Host "Database Name: $dbName" -ForegroundColor Cyan
    Write-Host ""
}

# Function to generate SQL for all migrations
function New-MigrationSQL {
    Write-Info "Generating database migration SQL..."
    
    $sqlContent = @'
-- HRMS Database Initialization Script
-- This combines all migration files

-- ==========================================
-- PTO Tables
-- ==========================================

CREATE TABLE IF NOT EXISTS pto_policies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    days_per_year INTEGER NOT NULL,
    carryover_days INTEGER DEFAULT 0,
    accrual_rate DECIMAL(5,2),
    effective_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pto_requests (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL,
    policy_id INTEGER REFERENCES pto_policies(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    days_requested DECIMAL(5,2) NOT NULL,
    reason TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    approved_by INTEGER,
    approved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pto_balances (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL UNIQUE,
    policy_id INTEGER REFERENCES pto_policies(id),
    days_available DECIMAL(5,2) DEFAULT 0,
    days_used DECIMAL(5,2) DEFAULT 0,
    days_pending DECIMAL(5,2) DEFAULT 0,
    year INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==========================================
-- Benefits Tables
-- ==========================================

CREATE TABLE IF NOT EXISTS benefits_plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    description TEXT,
    provider VARCHAR(255),
    monthly_cost DECIMAL(10,2),
    employer_contribution DECIMAL(10,2),
    employee_contribution DECIMAL(10,2),
    coverage_details JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS employee_benefits (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL,
    plan_id INTEGER REFERENCES benefits_plans(id),
    enrollment_date DATE NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    status VARCHAR(50) DEFAULT 'active',
    dependents JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==========================================
-- Timesheet Tables
-- ==========================================

CREATE TABLE IF NOT EXISTS timesheets (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL,
    week_start_date DATE NOT NULL,
    week_end_date DATE NOT NULL,
    total_hours DECIMAL(5,2) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'draft',
    submitted_at TIMESTAMP,
    approved_by INTEGER,
    approved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_id, week_start_date)
);

CREATE TABLE IF NOT EXISTS timesheet_entries (
    id SERIAL PRIMARY KEY,
    timesheet_id INTEGER REFERENCES timesheets(id) ON DELETE CASCADE,
    work_date DATE NOT NULL,
    hours DECIMAL(5,2) NOT NULL,
    project_code VARCHAR(100),
    task_description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==========================================
-- Payroll Tables
-- ==========================================

CREATE TABLE IF NOT EXISTS payroll_runs (
    id SERIAL PRIMARY KEY,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    pay_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',
    total_gross DECIMAL(12,2),
    total_net DECIMAL(12,2),
    total_taxes DECIMAL(12,2),
    total_deductions DECIMAL(12,2),
    processed_by INTEGER,
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS payroll_records (
    id SERIAL PRIMARY KEY,
    payroll_run_id INTEGER REFERENCES payroll_runs(id) ON DELETE CASCADE,
    employee_id INTEGER NOT NULL,
    hours_worked DECIMAL(6,2),
    regular_pay DECIMAL(10,2),
    overtime_pay DECIMAL(10,2),
    bonus_pay DECIMAL(10,2),
    gross_pay DECIMAL(10,2),
    federal_tax DECIMAL(10,2),
    state_tax DECIMAL(10,2),
    social_security DECIMAL(10,2),
    medicare DECIMAL(10,2),
    deductions DECIMAL(10,2),
    net_pay DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==========================================
-- Recruiting Tables
-- ==========================================

CREATE TABLE IF NOT EXISTS job_postings (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    department VARCHAR(100),
    location VARCHAR(255),
    employment_type VARCHAR(50),
    description TEXT,
    requirements TEXT,
    salary_range VARCHAR(100),
    status VARCHAR(50) DEFAULT 'draft',
    posted_date DATE,
    closing_date DATE,
    created_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS candidates (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(50),
    resume_url VARCHAR(500),
    linkedin_url VARCHAR(500),
    status VARCHAR(50) DEFAULT 'new',
    source VARCHAR(100),
    applied_date DATE DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS applications (
    id SERIAL PRIMARY KEY,
    candidate_id INTEGER REFERENCES candidates(id) ON DELETE CASCADE,
    job_posting_id INTEGER REFERENCES job_postings(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'applied',
    cover_letter TEXT,
    applied_date DATE DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(candidate_id, job_posting_id)
);

CREATE TABLE IF NOT EXISTS interviews (
    id SERIAL PRIMARY KEY,
    application_id INTEGER REFERENCES applications(id) ON DELETE CASCADE,
    interview_type VARCHAR(100),
    scheduled_date TIMESTAMP,
    duration_minutes INTEGER,
    interviewer_id INTEGER,
    location VARCHAR(255),
    notes TEXT,
    status VARCHAR(50) DEFAULT 'scheduled',
    feedback TEXT,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==========================================
-- Workflow Tables
-- ==========================================

CREATE TABLE IF NOT EXISTS workflows (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    workflow_type VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflow_steps (
    id SERIAL PRIMARY KEY,
    workflow_id INTEGER REFERENCES workflows(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,
    step_name VARCHAR(255) NOT NULL,
    approver_role VARCHAR(100),
    is_required BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflow_instances (
    id SERIAL PRIMARY KEY,
    workflow_id INTEGER REFERENCES workflows(id),
    entity_type VARCHAR(100) NOT NULL,
    entity_id INTEGER NOT NULL,
    current_step INTEGER,
    status VARCHAR(50) DEFAULT 'pending',
    initiated_by INTEGER,
    initiated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflow_approvals (
    id SERIAL PRIMARY KEY,
    workflow_instance_id INTEGER REFERENCES workflow_instances(id) ON DELETE CASCADE,
    step_id INTEGER REFERENCES workflow_steps(id),
    approver_id INTEGER NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    comments TEXT,
    approved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==========================================
-- User and Authentication Tables
-- ==========================================

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) DEFAULT 'employee',
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==========================================
-- Indexes for Performance
-- ==========================================

CREATE INDEX IF NOT EXISTS idx_pto_requests_employee ON pto_requests(employee_id);
CREATE INDEX IF NOT EXISTS idx_pto_requests_status ON pto_requests(status);
CREATE INDEX IF NOT EXISTS idx_pto_balances_employee ON pto_balances(employee_id);

CREATE INDEX IF NOT EXISTS idx_employee_benefits_employee ON employee_benefits(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_benefits_plan ON employee_benefits(plan_id);

CREATE INDEX IF NOT EXISTS idx_timesheets_employee ON timesheets(employee_id);
CREATE INDEX IF NOT EXISTS idx_timesheets_status ON timesheets(status);
CREATE INDEX IF NOT EXISTS idx_timesheet_entries_timesheet ON timesheet_entries(timesheet_id);

CREATE INDEX IF NOT EXISTS idx_payroll_records_run ON payroll_records(payroll_run_id);
CREATE INDEX IF NOT EXISTS idx_payroll_records_employee ON payroll_records(employee_id);

CREATE INDEX IF NOT EXISTS idx_applications_candidate ON applications(candidate_id);
CREATE INDEX IF NOT EXISTS idx_applications_job ON applications(job_posting_id);
CREATE INDEX IF NOT EXISTS idx_interviews_application ON interviews(application_id);

CREATE INDEX IF NOT EXISTS idx_workflow_instances_workflow ON workflow_instances(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_instances_entity ON workflow_instances(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_workflow_approvals_instance ON workflow_approvals(workflow_instance_id);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- ==========================================
-- Insert Sample Data
-- ==========================================

-- Insert a default admin user (password: admin123)
-- You should change this password immediately after first login!
INSERT INTO users (username, email, password_hash, first_name, last_name, role, is_active)
VALUES (
    'admin',
    'admin@example.com',
    '$2a$10$rQ7YXVJK5xGJ3K5jYx8b5.QZ4Y8K5K5K5K5K5K5K5K5K5K5K5K5K5',
    'Admin',
    'User',
    'admin',
    true
) ON CONFLICT (username) DO NOTHING;

-- Insert default PTO policy
INSERT INTO pto_policies (name, description, days_per_year, carryover_days, accrual_rate, effective_date)
VALUES (
    'Standard PTO',
    'Standard paid time off policy',
    20,
    5,
    1.67,
    CURRENT_DATE
) ON CONFLICT DO NOTHING;

-- ==========================================
-- Completion
-- ==========================================

SELECT 'Database initialization completed successfully!' as status;
'@

    $outputPath = Join-Path $env:TEMP "init_database.sql"
    $sqlContent | Out-File -FilePath $outputPath -Encoding UTF8
    
    Write-Success "Migration SQL generated at: $outputPath"
    return $outputPath
}

# Function to display connection information
function Show-ConnectionInfo {
    $dbHost = Get-DatabaseEndpoint
    $dbName = Get-DatabaseName
    
    Write-Host ""
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host "Database Connection Information" -ForegroundColor Cyan
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host "Host: $dbHost" -ForegroundColor White
    Write-Host "Port: 5432" -ForegroundColor White
    Write-Host "Database: $dbName" -ForegroundColor White
    Write-Host "Username: postgres" -ForegroundColor White
    Write-Host "Password: [as configured in stack]" -ForegroundColor White
    Write-Host ""
    Write-Host "To connect manually using psql:" -ForegroundColor Yellow
    Write-Host "  psql -h $dbHost -U postgres -d $dbName" -ForegroundColor White
    Write-Host ""
    Write-Host "To run the initialization SQL:" -ForegroundColor Yellow
    Write-Host "  psql -h $dbHost -U postgres -d $dbName -f `"$outputPath`"" -ForegroundColor White
    Write-Host ""
    Write-Host "Or use pgAdmin, DBeaver, or any PostgreSQL client" -ForegroundColor Yellow
    Write-Host ""
}

# Function to create a PowerShell connection script
function New-ConnectionScript {
    param([string]$DbHost, [string]$DbName, [string]$SqlFile)
    
    $scriptPath = Join-Path $env:TEMP "connect_and_init.ps1"
    
    $scriptContent = @"
# Quick Database Connection and Initialization Script
# This script helps you connect to the database and run the initialization

`$DBHost = "$DbHost"
`$DBName = "$DbName"
`$DBUser = "postgres"
`$SQLFile = "$SqlFile"

Write-Host "Connecting to database..." -ForegroundColor Green
Write-Host "Host: `$DBHost" -ForegroundColor Cyan
Write-Host "Database: `$DBName" -ForegroundColor Cyan
Write-Host ""

# Check if psql is available
try {
    `$null = Get-Command psql -ErrorAction Stop
    Write-Host "PostgreSQL client (psql) found!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Running initialization SQL..." -ForegroundColor Yellow
    
    # Prompt for password
    `$env:PGPASSWORD = Read-Host "Enter database password" -AsSecureString | ConvertFrom-SecureString
    
    # Run the SQL file
    psql -h `$DBHost -U `$DBUser -d `$DBName -f `$SQLFile
    
    if (`$LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "Database initialized successfully!" -ForegroundColor Green
    } else {
        Write-Host ""
        Write-Host "Failed to initialize database. Check the output above for errors." -ForegroundColor Red
    }
} catch {
    Write-Host "PostgreSQL client (psql) not found in PATH" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install PostgreSQL client tools or use a GUI tool like:" -ForegroundColor Yellow
    Write-Host "  - pgAdmin: https://www.pgadmin.org/" -ForegroundColor White
    Write-Host "  - DBeaver: https://dbeaver.io/" -ForegroundColor White
    Write-Host "  - Azure Data Studio: https://docs.microsoft.com/sql/azure-data-studio/" -ForegroundColor White
    Write-Host ""
    Write-Host "Connection details:" -ForegroundColor Yellow
    Write-Host "  Host: `$DBHost" -ForegroundColor White
    Write-Host "  Port: 5432" -ForegroundColor White
    Write-Host "  Database: `$DBName" -ForegroundColor White
    Write-Host "  Username: postgres" -ForegroundColor White
    Write-Host ""
    Write-Host "SQL file location: `$SQLFile" -ForegroundColor White
}

Write-Host ""
Write-Host "Press any key to exit..."
`$null = `$Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
"@

    $scriptContent | Out-File -FilePath $scriptPath -Encoding UTF8
    
    Write-Info "Created connection helper script: $scriptPath"
    return $scriptPath
}

# Main function
function Main {
    Write-Info "Initializing HRMS database..."
    Write-Info "Stack Name: $StackName"
    Write-Info "Region: $Region"
    Write-Host ""
    
    try {
        Get-DatabaseCredentials
        
        $sqlFile = New-MigrationSQL
        $dbHost = Get-DatabaseEndpoint
        $dbName = Get-DatabaseName
        
        Show-ConnectionInfo
        
        $connectScript = New-ConnectionScript -DbHost $dbHost -DbName $dbName -SqlFile $sqlFile
        
        Write-Host "======================================" -ForegroundColor Green
        Write-Host "Next Steps" -ForegroundColor Green
        Write-Host "======================================" -ForegroundColor Green
        Write-Host ""
        Write-Host "1. SQL file has been generated and saved" -ForegroundColor White
        Write-Host "2. Use one of these methods to initialize the database:" -ForegroundColor White
        Write-Host ""
        Write-Host "   Option A - Use psql command line:" -ForegroundColor Yellow
        Write-Host "     psql -h $dbHost -U postgres -d $dbName -f `"$sqlFile`"" -ForegroundColor White
        Write-Host ""
        Write-Host "   Option B - Run the helper script:" -ForegroundColor Yellow
        Write-Host "     & `"$connectScript`"" -ForegroundColor White
        Write-Host ""
        Write-Host "   Option C - Use a GUI tool (pgAdmin, DBeaver, etc.):" -ForegroundColor Yellow
        Write-Host "     - Connect using the credentials above" -ForegroundColor White
        Write-Host "     - Open and execute the SQL file: $sqlFile" -ForegroundColor White
        Write-Host ""
        Write-Host "3. After initialization, create additional admin users as needed" -ForegroundColor White
        Write-Host "4. Change the default admin password immediately!" -ForegroundColor Red
        Write-Host ""
        
        Write-Success "Database initialization files ready"
        
    } catch {
        Write-Error "Failed to initialize database: $_"
        exit 1
    }
}

# Run main function
Main
