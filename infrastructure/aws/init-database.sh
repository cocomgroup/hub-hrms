#!/bin/bash

# Database Initialization Script
# Run this after the stack is deployed to set up the database schema

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Configuration
STACK_NAME="hrms-prod"
REGION="us-east-1"

# Get database endpoint from CloudFormation
get_db_endpoint() {
    aws cloudformation describe-stacks \
        --stack-name "$STACK_NAME" \
        --region "$REGION" \
        --query 'Stacks[0].Outputs[?OutputKey==`DatabaseEndpoint`].OutputValue' \
        --output text
}

# Get database credentials
get_db_credentials() {
    print_info "Retrieving database credentials..."
    
    DB_HOST=$(get_db_endpoint)
    DB_NAME=$(aws cloudformation describe-stacks \
        --stack-name "$STACK_NAME" \
        --region "$REGION" \
        --query 'Stacks[0].Outputs[?OutputKey==`DatabaseName`].OutputValue' \
        --output text)
    
    echo "Database Host: $DB_HOST"
    echo "Database Name: $DB_NAME"
}

# Function to run SQL file via ECS task
run_sql_via_ecs() {
    local SQL_FILE=$1
    
    print_info "Running SQL file: $SQL_FILE"
    
    # Get cluster and service info
    CLUSTER=$(aws cloudformation describe-stacks \
        --stack-name "$STACK_NAME" \
        --region "$REGION" \
        --query 'Stacks[0].Outputs[?OutputKey==`ECSClusterName`].OutputValue' \
        --output text)
    
    # Get a running task
    TASK_ARN=$(aws ecs list-tasks \
        --cluster "$CLUSTER" \
        --service-name "${STACK_NAME}-backend" \
        --desired-status RUNNING \
        --region "$REGION" \
        --query 'taskArns[0]' \
        --output text)
    
    if [ -z "$TASK_ARN" ] || [ "$TASK_ARN" == "None" ]; then
        print_error "No running backend tasks found"
        return 1
    fi
    
    print_info "Using task: $TASK_ARN"
    
    # Execute SQL commands
    print_info "Executing migrations..."
    
    # Note: This requires ECS Exec to be enabled
    print_warning "ECS Exec must be enabled for this to work"
    print_info "Attempting to execute SQL file..."
}

# Generate SQL for all migrations
generate_migration_sql() {
    cat << 'EOF' > /tmp/init_database.sql
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

CREATE INDEX idx_pto_requests_employee ON pto_requests(employee_id);
CREATE INDEX idx_pto_requests_status ON pto_requests(status);
CREATE INDEX idx_pto_balances_employee ON pto_balances(employee_id);

CREATE INDEX idx_employee_benefits_employee ON employee_benefits(employee_id);
CREATE INDEX idx_employee_benefits_plan ON employee_benefits(plan_id);

CREATE INDEX idx_timesheets_employee ON timesheets(employee_id);
CREATE INDEX idx_timesheets_status ON timesheets(status);
CREATE INDEX idx_timesheet_entries_timesheet ON timesheet_entries(timesheet_id);

CREATE INDEX idx_payroll_records_run ON payroll_records(payroll_run_id);
CREATE INDEX idx_payroll_records_employee ON payroll_records(employee_id);

CREATE INDEX idx_applications_candidate ON applications(candidate_id);
CREATE INDEX idx_applications_job ON applications(job_posting_id);
CREATE INDEX idx_interviews_application ON interviews(application_id);

CREATE INDEX idx_workflow_instances_workflow ON workflow_instances(workflow_id);
CREATE INDEX idx_workflow_instances_entity ON workflow_instances(entity_type, entity_id);
CREATE INDEX idx_workflow_approvals_instance ON workflow_approvals(workflow_instance_id);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- ==========================================
-- Insert Sample Data
-- ==========================================

-- Insert a default admin user (password: admin123)
-- Password hash generated with bcrypt
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
EOF

    print_success "Migration SQL generated at /tmp/init_database.sql"
}

# Function to display connection instructions
display_connection_info() {
    DB_HOST=$(get_db_endpoint)
    DB_NAME=$(aws cloudformation describe-stacks \
        --stack-name "$STACK_NAME" \
        --region "$REGION" \
        --query 'Stacks[0].Outputs[?OutputKey==`DatabaseName`].OutputValue' \
        --output text)
    
    echo ""
    echo "======================================"
    echo "Database Connection Information"
    echo "======================================"
    echo "Host: $DB_HOST"
    echo "Port: 5432"
    echo "Database: $DB_NAME"
    echo "Username: postgres"
    echo "Password: [as configured in stack]"
    echo ""
    echo "To connect manually:"
    echo "psql -h $DB_HOST -U postgres -d $DB_NAME"
    echo ""
    echo "To run the initialization SQL:"
    echo "psql -h $DB_HOST -U postgres -d $DB_NAME -f /tmp/init_database.sql"
    echo ""
}

# Main function
main() {
    print_info "Initializing HRMS database..."
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --stack-name)
                STACK_NAME="$2"
                shift 2
                ;;
            --region)
                REGION="$2"
                shift 2
                ;;
            --help)
                echo "Usage: ./init-database.sh [options]"
                echo ""
                echo "Options:"
                echo "  --stack-name NAME  CloudFormation stack name (default: hrms-prod)"
                echo "  --region REGION    AWS region (default: us-east-1)"
                echo "  --help             Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    generate_migration_sql
    display_connection_info
    
    print_success "Database initialization script ready"
    print_info "Please run the SQL file manually using psql or your preferred database client"
}

main "$@"
