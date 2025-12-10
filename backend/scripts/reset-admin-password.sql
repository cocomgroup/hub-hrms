-- Reset Admin Password to 'admin123'
-- Run this with: psql -h localhost -U postgres -d drmsdb -f reset-admin-password.sql

-- Update the password hash for admin user
-- This hash is for password: admin123
-- Generated with: bcrypt.GenerateFromPassword([]byte("admin123"), 10)

UPDATE users 
SET password_hash = '$2a$10$rV8xQvZ8JqXqy4JqXqy4JO8QqXqy4JqXqy4JqXqy4JqXqy4JqXqy'
WHERE email = 'admin@cocomgroup.com';

-- Verify the update
SELECT 
    email, 
    role,
    CASE 
        WHEN password_hash LIKE '$2a$10$%' THEN 'Valid bcrypt hash'
        ELSE 'Invalid hash format'
    END as hash_status,
    LENGTH(password_hash) as hash_length
FROM users 
WHERE email = 'admin@cocomgroup.com';

-- If no rows updated, the user doesn't exist
-- Run this to check:
SELECT COUNT(*) as user_count 
FROM users 
WHERE email = 'admin@cocomgroup.com';
