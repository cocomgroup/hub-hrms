-- ========================================
-- INITIAL ADMIN USER SETUP
-- ========================================
-- This script creates the first admin user for your HR system
-- Run this ONCE after completing database migrations

BEGIN;

-- Create admin employee
WITH new_employee AS (
    INSERT INTO employees (
        first_name,
        last_name,
        email,
        phone,
        hire_date,
        department,
        position,
        employment_type,
        status
    ) VALUES (
        'System',
        'Administrator',
        'admin@company.com',
        '555-0100',
        CURRENT_DATE,
        'Administration',
        'System Administrator',
        'full-time',
        'active'
    ) RETURNING id
),

-- Create user account
new_user AS (
    INSERT INTO users (
        email,
        password_hash,
        role,
        employee_id
    )
    SELECT
        'admin@company.com',
        -- Password: 'admin123' (CHANGE THIS IMMEDIATELY!)
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
        'admin',
        id
    FROM new_employee
    RETURNING id, employee_id
)

-- Create PTO balance
INSERT INTO pto_balances (
    employee_id,
    vacation_days,
    sick_days,
    personal_days,
    accrual_rate_vacation,
    accrual_rate_sick,
    last_accrual_date
)
SELECT
    employee_id,
    15.0,
    10.0,
    5.0,
    1.25,
    0.83,
    CURRENT_DATE
FROM new_user;

COMMIT;

-- Verify the setup
SELECT 
    u.email as "Email",
    u.role as "Role",
    e.first_name || ' ' || e.last_name as "Name",
    e.position as "Position",
    'admin123' as "Default Password (CHANGE THIS!)"
FROM users u
JOIN employees e ON u.employee_id = e.id
WHERE u.email = 'admin@company.com';

-- ========================================
-- NEXT STEPS:
-- 1. Login with: admin@company.com / admin123
-- 2. Change the password immediately
-- 3. Use the web interface to add more employees
-- ========================================
