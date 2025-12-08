-- Initial Admin User Setup Script
-- Run this after your database migrations are complete

-- Step 1: Create the admin employee record
INSERT INTO employees (
    id,
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
    gen_random_uuid(),
    'Admin',
    'User',
    'admin@cocomgroup.com',
    '555-0100',
    CURRENT_DATE,
    'Administration',
    'System Administrator',
    'full-time',
    'active'
) RETURNING id;

-- Step 2: Create the user account linked to the employee
-- Note: Replace 'EMPLOYEE_ID_FROM_STEP_1' with the actual UUID returned above
-- Password below is 'admin123' - CHANGE THIS IMMEDIATELY!
INSERT INTO users (
    email,
    password_hash,
    role,
    employee_id
) VALUES (
    'admin@cocomgroup.com',
    '$2a$10$rQYDzJ7U0qKXLJvHKJGJ9O8nxK9qY6M5f5tD3qJ5L5J5qJ5qJ5qJ5', -- admin123
    'admin',
    'EMPLOYEE_ID_FROM_STEP_1' -- Replace with actual UUID
);

-- Step 3: Create initial PTO balance for admin
INSERT INTO pto_balances (
    employee_id,
    vacation_days,
    sick_days,
    personal_days,
    accrual_rate_vacation,
    accrual_rate_sick,
    last_accrual_date
) VALUES (
    'EMPLOYEE_ID_FROM_STEP_1', -- Replace with actual UUID
    15.0,
    10.0,
    5.0,
    1.25, -- 15 days per year = 1.25 per month
    0.83, -- 10 days per year
    CURRENT_DATE
);

-- Verify the user was created
SELECT u.id, u.email, u.role, e.first_name, e.last_name, e.position
FROM users u
JOIN employees e ON u.employee_id = e.id
WHERE u.email = 'admin@cocomgroup.com';
