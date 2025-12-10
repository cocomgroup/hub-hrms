#!/bin/bash

# HR System - Initial Setup Script
# Creates the first admin user and optionally adds sample data

set -e

echo "=========================================="
echo "HR System - Initial Setup"
echo "=========================================="
echo ""

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "Error: psql command not found. Please install PostgreSQL client."
    exit 1
fi

# Get database connection details
echo "Enter database connection details:"
read -p "Host [localhost]: " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "Port [5432]: " DB_PORT
DB_PORT=${DB_PORT:-5432}

read -p "Database [hrapp]: " DB_NAME
DB_NAME=${DB_NAME:-hrapp}

read -p "User [postgres]: " DB_USER
DB_USER=${DB_USER:-postgres}

read -sp "Password: " DB_PASSWORD
echo ""

# Test connection
echo ""
echo "Testing database connection..."
export PGPASSWORD="$DB_PASSWORD"

if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "Error: Could not connect to database. Please check your credentials."
    exit 1
fi

echo "✓ Database connection successful"
echo ""

# Check if admin user already exists
ADMIN_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT COUNT(*) FROM users WHERE email = 'admin@company.com';" | xargs)

if [ "$ADMIN_EXISTS" -gt "0" ]; then
    echo "⚠ Warning: Admin user already exists!"
    read -p "Do you want to reset the admin password? (y/N): " RESET
    if [[ ! $RESET =~ ^[Yy]$ ]]; then
        echo "Skipping admin user creation."
        exit 0
    fi
    
    # Reset password
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
UPDATE users
SET password_hash = '\$2a\$10\$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'
WHERE email = 'admin@company.com';
EOF
    
    echo "✓ Admin password reset to 'admin123'"
    echo ""
else
    # Create admin user
    echo "Creating admin user..."
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << 'EOF'
BEGIN;

WITH new_employee AS (
    INSERT INTO employees (
        first_name, last_name, email, phone,
        hire_date, department, position,
        employment_type, status
    ) VALUES (
        'System', 'Administrator', 'admin@company.com', '555-0100',
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
EOF

    echo "✓ Admin user created successfully"
    echo ""
fi

# Display credentials
echo "=========================================="
echo "Admin User Credentials"
echo "=========================================="
echo "Email:    admin@company.com"
echo "Password: admin123"
echo ""
echo "⚠ IMPORTANT: Change this password immediately after first login!"
echo ""

# Ask about sample data
read -p "Do you want to add sample employees for testing? (y/N): " ADD_SAMPLES
if [[ $ADD_SAMPLES =~ ^[Yy]$ ]]; then
    echo ""
    echo "Creating sample employees..."
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << 'EOF'
-- Sample Employee 1: John Doe
WITH emp1 AS (
    INSERT INTO employees (
        first_name, last_name, email, phone,
        hire_date, department, position, employment_type, status
    ) VALUES (
        'John', 'Doe', 'john.doe@company.com', '555-0101',
        CURRENT_DATE - INTERVAL '30 days', 'Engineering', 'Senior Developer',
        'full-time', 'active'
    ) RETURNING id
),
user1 AS (
    INSERT INTO users (email, password_hash, role, employee_id)
    SELECT 'john.doe@company.com',
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
        'Jane', 'Smith', 'jane.smith@company.com', '555-0102',
        CURRENT_DATE - INTERVAL '60 days', 'Human Resources', 'HR Manager',
        'full-time', 'active'
    ) RETURNING id
),
user2 AS (
    INSERT INTO users (email, password_hash, role, employee_id)
    SELECT 'jane.smith@company.com',
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
        'Bob', 'Johnson', 'bob.johnson@company.com', '555-0103',
        CURRENT_DATE, 'Sales', 'Sales Representative',
        'full-time', 'active'
    ) RETURNING id
),
user3 AS (
    INSERT INTO users (email, password_hash, role, employee_id)
    SELECT 'bob.johnson@company.com',
           '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
           'employee', id FROM emp3
    RETURNING employee_id
)
INSERT INTO pto_balances (employee_id, vacation_days, sick_days, personal_days)
SELECT employee_id, 15.0, 10.0, 5.0 FROM user3;
EOF

    echo "✓ Sample employees created:"
    echo "  - john.doe@company.com (password: admin123)"
    echo "  - jane.smith@company.com (password: admin123)"
    echo "  - bob.johnson@company.com (password: admin123)"
    echo ""
fi

# Verify setup
echo "Verifying setup..."
TOTAL_USERS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT COUNT(*) FROM users;" | xargs)

echo "✓ Total users in system: $TOTAL_USERS"
echo ""

# Display next steps
echo "=========================================="
echo "Setup Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Start the backend server:"
echo "   cd backend && go run cmd/main.go"
echo ""
echo "2. Start the frontend:"
echo "   cd frontend && npm run dev"
echo ""
echo "3. Access the application:"
echo "   http://localhost:5173"
echo ""
echo "4. Login with admin credentials"
echo "5. CHANGE THE DEFAULT PASSWORD!"
echo ""
echo "For more information, see ONBOARDING_GUIDE.md"
echo ""

unset PGPASSWORD
