-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


CREATE TABLE users (
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

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_employee_id ON users(employee_id);

-- Insert admin user
INSERT INTO users (username, email, password_hash, first_name, last_name, role, is_active)
VALUES (
    'admin',
    'admin@cocomgroup.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMye1J1mPGp7.EaHMgxvvvKLQYYaGZhcAXq',
    'Admin',
    'User',
    'admin',
    true
);

COMMIT;
