-- Complete Database Setup Script
-- Run this after creating the database and enabling extensions

-- Enable extensions (run these first if not already done)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create admin user
-- Password: admin123
-- Note: You must generate the actual bcrypt hash using: go run scripts/hash_password.go admin123
-- Then replace the hash below
INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'admin@cocomgroup.com',
    '$2a$10$REPLACE_WITH_ACTUAL_HASH',
    'admin',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- Create sample employees
INSERT INTO employees (id, first_name, last_name, email, phone, hire_date, department, position, status, created_at, updated_at)
VALUES 
(
    gen_random_uuid(),
    'Evan',
    'Hunt',
    'evan.hunt@cocomgroup.com',
    '555-0101',
    '2020-01-15',
    'Engineering',
    'CTO',
    'active',
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'Bob',
    'Johnson',
    'bob.johnson@cocomgroup.com',
    '555-0102',
    '2021-03-20',
    'Sales',
    'Sales Representative',
    'active',
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'Jane',
    'Smith',
    'jane.smith@cocomgroup.com',
    '555-0103',
    '2021-06-10',
    'Human Resources',
    'HR Manager',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- Fix foreign key constraint (if already created with wrong reference)
ALTER TABLE IF EXISTS onboarding_workflows 
DROP CONSTRAINT IF EXISTS onboarding_workflows_created_by_fkey;

ALTER TABLE IF EXISTS onboarding_workflows 
ADD CONSTRAINT onboarding_workflows_created_by_fkey 
FOREIGN KEY (created_by) REFERENCES users(id);

-- Verify setup
SELECT 'Users:' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 'Employees:', COUNT(*) FROM employees;

-- Show admin user
SELECT id, email, role FROM users WHERE email = 'admin@cocomgroup.com';

-- Show employees
SELECT id, first_name, last_name, email, position FROM employees;