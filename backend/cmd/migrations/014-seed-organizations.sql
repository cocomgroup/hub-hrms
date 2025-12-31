-- Seed Data: Organizations
-- File: 014-seed-organizations.sql
-- Description: Populate company organizational structure

-- Note: The organizations table schema:
-- - No address fields (just location as VARCHAR)
-- - No 'company', 'subsidiary', 'branch' types (only: division, department, team, unit, group)
-- - Has 'code' field (required, unique)
-- - Has 'is_active' instead of 'status'
-- - No phone, email, website, founded_date fields

-- Top-level Divisions
INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
VALUES 
    ('Engineering', 'ENG', 'division', NULL, 'Software development and engineering', 'San Francisco, CA', 0, 15, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
VALUES 
    ('Sales', 'SALES', 'division', NULL, 'Sales and business development', 'Chicago, IL', 0, 8, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
VALUES 
    ('Marketing', 'MKT', 'division', NULL, 'Marketing and brand management', 'Los Angeles, CA', 0, 6, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
VALUES 
    ('Human Resources', 'HR', 'division', NULL, 'HR and people operations', 'New York, NY', 0, 4, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
VALUES 
    ('Finance', 'FIN', 'division', NULL, 'Finance and accounting', 'New York, NY', 0, 5, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
VALUES 
    ('IT', 'IT', 'division', NULL, 'Information technology', 'Seattle, WA', 0, 4, true)
ON CONFLICT (code) DO NOTHING;

-- Engineering Departments
INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Product Development',
    'ENG-PD',
    'department',
    id,
    'Core product development',
    'San Francisco, CA',
    1,
    8,
    true
FROM organizations WHERE code = 'ENG'
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Quality Assurance',
    'ENG-QA',
    'department',
    id,
    'Quality testing and assurance',
    'San Francisco, CA',
    1,
    4,
    true
FROM organizations WHERE code = 'ENG'
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'DevOps',
    'ENG-DO',
    'department',
    id,
    'Infrastructure and operations',
    'Remote',
    1,
    3,
    true
FROM organizations WHERE code = 'ENG'
ON CONFLICT (code) DO NOTHING;

-- Engineering Teams (under Product Development)
INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Backend Team',
    'ENG-PD-BE',
    'team',
    id,
    'Backend API development',
    'San Francisco, CA',
    2,
    4,
    true
FROM organizations WHERE code = 'ENG-PD'
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Frontend Team',
    'ENG-PD-FE',
    'team',
    id,
    'UI/UX and frontend development',
    'San Francisco, CA',
    2,
    4,
    true
FROM organizations WHERE code = 'ENG-PD'
ON CONFLICT (code) DO NOTHING;

-- Sales Departments
INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Enterprise Sales',
    'SALES-ENT',
    'department',
    id,
    'Enterprise and strategic accounts',
    'Chicago, IL',
    1,
    5,
    true
FROM organizations WHERE code = 'SALES'
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'SMB Sales',
    'SALES-SMB',
    'department',
    id,
    'Small and medium business sales',
    'Dallas, TX',
    1,
    3,
    true
FROM organizations WHERE code = 'SALES'
ON CONFLICT (code) DO NOTHING;

-- Marketing Departments
INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Digital Marketing',
    'MKT-DIG',
    'department',
    id,
    'Digital marketing and social media',
    'Los Angeles, CA',
    1,
    3,
    true
FROM organizations WHERE code = 'MKT'
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Content Marketing',
    'MKT-CNT',
    'department',
    id,
    'Content creation and strategy',
    'Los Angeles, CA',
    1,
    3,
    true
FROM organizations WHERE code = 'MKT'
ON CONFLICT (code) DO NOTHING;

-- HR Departments
INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Recruiting',
    'HR-REC',
    'department',
    id,
    'Talent acquisition',
    'New York, NY',
    1,
    2,
    true
FROM organizations WHERE code = 'HR'
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'People Operations',
    'HR-POP',
    'department',
    id,
    'HR operations and benefits',
    'New York, NY',
    1,
    2,
    true
FROM organizations WHERE code = 'HR'
ON CONFLICT (code) DO NOTHING;

-- Finance Departments
INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Accounting',
    'FIN-ACC',
    'department',
    id,
    'Financial accounting',
    'New York, NY',
    1,
    3,
    true
FROM organizations WHERE code = 'FIN'
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Financial Planning',
    'FIN-FPA',
    'department',
    id,
    'Financial planning and analysis',
    'New York, NY',
    1,
    2,
    true
FROM organizations WHERE code = 'FIN'
ON CONFLICT (code) DO NOTHING;

-- IT Departments
INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'IT Support',
    'IT-SUP',
    'department',
    id,
    'Help desk and user support',
    'Seattle, WA',
    1,
    2,
    true
FROM organizations WHERE code = 'IT'
ON CONFLICT (code) DO NOTHING;

INSERT INTO organizations (name, code, type, parent_id, description, location, level, employee_count, is_active)
SELECT 
    'Infrastructure',
    'IT-INF',
    'department',
    id,
    'IT infrastructure management',
    'Seattle, WA',
    1,
    2,
    true
FROM organizations WHERE code = 'IT'
ON CONFLICT (code) DO NOTHING;

-- Display summary
DO $$
DECLARE
    org_count INTEGER;
    division_count INTEGER;
    department_count INTEGER;
    team_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO org_count FROM organizations;
    SELECT COUNT(*) INTO division_count FROM organizations WHERE type = 'division';
    SELECT COUNT(*) INTO department_count FROM organizations WHERE type = 'department';
    SELECT COUNT(*) INTO team_count FROM organizations WHERE type = 'team';
    
    RAISE NOTICE '================================';
    RAISE NOTICE 'Organization Seed Data Summary';
    RAISE NOTICE '================================';
    RAISE NOTICE 'Total organizations: %', org_count;
    RAISE NOTICE '  - Divisions: %', division_count;
    RAISE NOTICE '  - Departments: %', department_count;
    RAISE NOTICE '  - Teams: %', team_count;
    RAISE NOTICE '================================';
    RAISE NOTICE 'Organizational Structure:';
    RAISE NOTICE '  6 Divisions (Engineering, Sales, Marketing, HR, Finance, IT)';
    RAISE NOTICE '  14 Departments';
    RAISE NOTICE '  2 Teams';
    RAISE NOTICE '================================';
END $$;