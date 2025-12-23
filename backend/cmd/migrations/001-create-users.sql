-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID,
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

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_employee_id ON users(employee_id);

-- Insert admin user
INSERT INTO users (username, email, password_hash, first_name, last_name, role, is_active)
SELECT 'admin', 'admin@cocomgroup.com', '$2a$10$r9W05gPN.BTleDt6yOghVulZ5uShQtk/sqHBswfG9ydJ8E3ZyCriG', 'Admin', 'User', 'admin', true
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE username = 'admin'
);

COMMIT;
