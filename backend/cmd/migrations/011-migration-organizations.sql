-- Organizations Database Migration
-- Creates tables for corporate organization structure

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    description TEXT,
    parent_id UUID REFERENCES organizations(id) ON DELETE RESTRICT,
    manager_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('division', 'department', 'team', 'unit', 'group')),
    level INTEGER NOT NULL DEFAULT 0,
    cost_center VARCHAR(50),
    location VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    employee_count INTEGER DEFAULT 0,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Organization employees junction table
CREATE TABLE IF NOT EXISTS organization_employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    role VARCHAR(100),
    is_primary BOOLEAN DEFAULT false,
    start_date DATE NOT NULL,
    end_date DATE,
    assigned_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, employee_id, end_date)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_organizations_parent_id ON organizations(parent_id);
CREATE INDEX IF NOT EXISTS idx_organizations_manager_id ON organizations(manager_id);
CREATE INDEX IF NOT EXISTS idx_organizations_code ON organizations(code);
CREATE INDEX IF NOT EXISTS idx_organizations_type ON organizations(type);
CREATE INDEX IF NOT EXISTS idx_organizations_is_active ON organizations(is_active);
CREATE INDEX IF NOT EXISTS idx_organizations_level ON organizations(level);

CREATE INDEX IF NOT EXISTS idx_org_employees_org_id ON organization_employees(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_employees_emp_id ON organization_employees(employee_id);
CREATE INDEX IF NOT EXISTS idx_org_employees_is_primary ON organization_employees(is_primary);
CREATE INDEX IF NOT EXISTS idx_org_employees_dates ON organization_employees(start_date, end_date);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();

DROP TRIGGER IF EXISTS update_org_employees_updated_at ON organization_employees;
CREATE TRIGGER update_org_employees_updated_at
    BEFORE UPDATE ON organization_employees
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();

-- Sample data for testing
INSERT INTO organizations (name, code, description, type, level, is_active, created_by)
VALUES 
    ('Executive', 'EXEC', 'Executive Leadership', 'division', 0, true, 
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('Finance', 'FIN', 'Finance Division', 'division', 0, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('Information Technology', 'IT', 'IT Division', 'division', 0, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('Human Resources', 'HR', 'Human Resources Division', 'division', 0, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('Legal', 'LEGAL', 'Legal Division', 'division', 0, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('Supply Chain', 'SCM', 'Supply Chain Management', 'division', 0, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('Learning & Development', 'L&D', 'Learning and Development', 'division', 0, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1))
ON CONFLICT (code) DO NOTHING;

-- Add departments under divisions
INSERT INTO organizations (name, code, description, parent_id, type, level, is_active, created_by)
VALUES 
    ('Accounting', 'FIN-ACCT', 'Accounting Department', 
     (SELECT id FROM organizations WHERE code = 'FIN'), 'department', 1, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('Financial Planning', 'FIN-FP', 'Financial Planning & Analysis', 
     (SELECT id FROM organizations WHERE code = 'FIN'), 'department', 1, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('Software Engineering', 'IT-SWE', 'Software Engineering Department', 
     (SELECT id FROM organizations WHERE code = 'IT'), 'department', 1, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('IT Operations', 'IT-OPS', 'IT Operations & Infrastructure', 
     (SELECT id FROM organizations WHERE code = 'IT'), 'department', 1, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('Recruiting', 'HR-REC', 'Talent Acquisition & Recruiting', 
     (SELECT id FROM organizations WHERE code = 'HR'), 'department', 1, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1)),
    ('Compensation & Benefits', 'HR-CB', 'Compensation and Benefits', 
     (SELECT id FROM organizations WHERE code = 'HR'), 'department', 1, true,
     (SELECT id FROM users WHERE role = 'admin' LIMIT 1))
ON CONFLICT (code) DO NOTHING;

-- Verification queries
SELECT 
    o.name,
    o.code,
    o.type,
    o.level,
    p.name as parent_name,
    o.is_active
FROM organizations o
LEFT JOIN organizations p ON o.parent_id = p.id
ORDER BY o.level, o.name;