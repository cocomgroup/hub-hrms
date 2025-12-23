-- Seed Data: Users
-- File: 012-seed-users.sql
-- Description: Populate initial users with different roles

-- Note: Passwords are bcrypt hashed versions of 'password123'
-- Hash: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy

-- Admin user - skip if already exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin' OR email = 'admin@cocomgroup.com') THEN
        INSERT INTO users (username, email, password_hash, role)
        VALUES ('admin', 'admin@cocomgroup.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin');
    END IF;
END $$;

-- HR Manager (only if jane.smith exists and user doesn't)
DO $$
DECLARE
    emp_id UUID;
BEGIN
    SELECT id INTO emp_id FROM employees WHERE email = 'jane.smith@cocomgroup.com' LIMIT 1;
    
    IF emp_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM users WHERE username = 'hr.manager' OR email = 'hr.manager@company.com') THEN
        INSERT INTO users (username, email, password_hash, role, employee_id)
        VALUES ('hr.manager', 'hr.manager@cocomgroup.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'hr-manager', emp_id);
    END IF;
END $$;

-- Department Managers - create users for all managers
INSERT INTO users (username, email, password_hash, role, employee_id)
SELECT 
    LOWER(REPLACE(e.email, '@cocomgroup.com', '')),  -- username from email
    e.email,
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'manager',
    e.id
FROM employees e
WHERE e.position LIKE '%Manager%' 
  AND e.email NOT IN (SELECT email FROM users WHERE email IS NOT NULL)
  AND LOWER(REPLACE(e.email, '@cocomgroup.com', '')) NOT IN (SELECT username FROM users WHERE username IS NOT NULL)
LIMIT 10
ON CONFLICT DO NOTHING;

-- Regular employees - create users for active employees
INSERT INTO users (username, email, password_hash, role, employee_id)
SELECT 
    LOWER(REPLACE(e.email, '@cocomgroup.com', '')),  -- username from email
    e.email,
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'employee',
    e.id
FROM employees e
WHERE e.email NOT IN (SELECT email FROM users WHERE email IS NOT NULL)
  AND LOWER(REPLACE(e.email, '@cocomgroup.com', '')) NOT IN (SELECT username FROM users WHERE username IS NOT NULL)
  AND e.status = 'active'
  AND e.employment_type = 'full-time'
LIMIT 25
ON CONFLICT DO NOTHING;

-- Display created users
DO $$
DECLARE
    user_count INTEGER;
    admin_count INTEGER;
    manager_count INTEGER;
    employee_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO user_count FROM users;
    SELECT COUNT(*) INTO admin_count FROM users WHERE role = 'admin';
    SELECT COUNT(*) INTO manager_count FROM users WHERE role IN ('hr-manager', 'manager');
    SELECT COUNT(*) INTO employee_count FROM users WHERE role = 'employee';
    
    RAISE NOTICE '================================';
    RAISE NOTICE 'User Seed Data Summary';
    RAISE NOTICE '================================';
    RAISE NOTICE 'Total users: %', user_count;
    RAISE NOTICE '  - Admins: %', admin_count;
    RAISE NOTICE '  - Managers: %', manager_count;
    RAISE NOTICE '  - Employees: %', employee_count;
    RAISE NOTICE '================================';
    RAISE NOTICE 'Test Login Credentials:';
    RAISE NOTICE '  Email: admin@cocomgroup.com';
    RAISE NOTICE '  Password: password123';
    RAISE NOTICE '================================';
    RAISE NOTICE 'All users have password: password123';
    RAISE NOTICE '================================';
END $$;
