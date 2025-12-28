-- Migration: Create projects and project_members tables
-- File: migrations/011_create_projects.sql
-- This version matches the actual database schema including code, client_name, and budget_hours

-- Create projects table with all columns
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50),  -- Made nullable to allow creation without code
    description TEXT,
    client_name VARCHAR(255),  -- Client/customer name
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    priority VARCHAR(50) NOT NULL DEFAULT 'medium',
    budget_hours DECIMAL(10, 2),  -- Budgeted hours for project
    manager_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    budget DECIMAL(15, 2),  -- Budget in currency
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on manager_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_projects_manager_id ON projects(manager_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_created_by ON projects(created_by);
CREATE INDEX IF NOT EXISTS idx_projects_code ON projects(code);

-- Create project_members table
CREATE TABLE IF NOT EXISTS project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    role VARCHAR(100) DEFAULT 'member',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, employee_id)
);

-- Create indexes for project_members
CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members(project_id);
CREATE INDEX IF NOT EXISTS idx_project_members_employee_id ON project_members(employee_id);

-- Add trigger to update updated_at
CREATE OR REPLACE FUNCTION update_projects_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_projects_updated_at ON projects;
CREATE TRIGGER trigger_update_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE PROCEDURE update_projects_updated_at();

-- Add comments
COMMENT ON TABLE projects IS 'Company projects';
COMMENT ON TABLE project_members IS 'Employees assigned to projects';
COMMENT ON COLUMN projects.code IS 'Project code/identifier (e.g., PROJ-001)';
COMMENT ON COLUMN projects.client_name IS 'Client or customer name';
COMMENT ON COLUMN projects.status IS 'Project status: active, on-hold, completed, archived';
COMMENT ON COLUMN projects.priority IS 'Project priority: low, medium, high, critical';
COMMENT ON COLUMN projects.budget_hours IS 'Budgeted hours for the project';
COMMENT ON COLUMN projects.budget IS 'Project budget in currency';