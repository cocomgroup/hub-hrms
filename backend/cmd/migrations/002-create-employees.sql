CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    date_of_birth DATE,
    hire_date DATE NOT NULL,
    department VARCHAR(100),
    position VARCHAR(100),
    manager_id UUID,
    employment_type VARCHAR(50),
    status VARCHAR(50) DEFAULT 'active',
    street_address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(50),
    zip_code VARCHAR(20),
    country VARCHAR(100),
    emergency_contact_name VARCHAR(200),
    emergency_contact_phone VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (manager_id) REFERENCES employees(id) ON DELETE SET NULL
);

ALTER TABLE employees 
ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id);

ALTER TABLE employees 
ADD COLUMN IF NOT EXISTS employment_type VARCHAR(50) DEFAULT 'employee';


-- Set default for existing records
UPDATE employees 
SET employment_type = 'employee' 
WHERE employment_type IS NULL;

CREATE INDEX IF NOT EXISTS idx_employees_user_id ON employees(user_id);
CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department);
CREATE INDEX IF NOT EXISTS idx_employees_manager_id ON employees(manager_id);
CREATE INDEX IF NOT EXISTS idx_employees_employment_type ON employees(employment_type);

COMMENT ON COLUMN employees.user_id IS 'Links employee to their user account (nullable if employee has no login)';
COMMENT ON COLUMN employees.employment_type IS 
'Employment classification: employee, contractor, independent contractor, 1099, etc.';