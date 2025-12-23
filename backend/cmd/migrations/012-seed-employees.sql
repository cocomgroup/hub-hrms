-- Seed Data: Employees
-- File: 013-seed-employees.sql
-- Description: Populate sample employees across departments

-- CEO
INSERT INTO employees (first_name, last_name, email, phone, hire_date, department, position, employment_type, status, city, state, country)
VALUES 
    ('John', 'CEO', 'john.ceo@company.com', '555-0001', '2020-01-15', 'Executive', 'Chief Executive Officer', 'full-time', 'active', 'New York', 'NY', 'USA')
ON CONFLICT (email) DO NOTHING;

-- HR Department
INSERT INTO employees (first_name, last_name, email, phone, hire_date, department, position, employment_type, status, city, state, country)
VALUES 
    ('Jane', 'Smith', 'jane.smith@company.com', '555-0101', '2020-03-01', 'Human Resources', 'HR Manager', 'full-time', 'active', 'New York', 'NY', 'USA'),
    ('Sarah', 'Johnson', 'sarah.johnson@company.com', '555-0102', '2021-06-15', 'Human Resources', 'HR Specialist', 'full-time', 'active', 'New York', 'NY', 'USA'),
    ('Michael', 'Brown', 'michael.brown@company.com', '555-0103', '2022-01-10', 'Human Resources', 'Recruiter', 'full-time', 'active', 'New York', 'NY', 'USA')
ON CONFLICT (email) DO NOTHING;

-- Engineering Department
INSERT INTO employees (first_name, last_name, email, phone, hire_date, department, position, employment_type, status, city, state, country)
VALUES 
    ('David', 'Wilson', 'david.wilson@company.com', '555-0201', '2019-05-20', 'Engineering', 'Engineering Manager', 'full-time', 'active', 'San Francisco', 'CA', 'USA'),
    ('Emily', 'Davis', 'emily.davis@company.com', '555-0202', '2021-02-15', 'Engineering', 'Senior Software Engineer', 'full-time', 'active', 'San Francisco', 'CA', 'USA'),
    ('James', 'Miller', 'james.miller@company.com', '555-0203', '2021-08-01', 'Engineering', 'Software Engineer', 'full-time', 'active', 'San Francisco', 'CA', 'USA'),
    ('Jennifer', 'Garcia', 'jennifer.garcia@company.com', '555-0204', '2022-03-10', 'Engineering', 'Software Engineer', 'full-time', 'active', 'Austin', 'TX', 'USA'),
    ('Robert', 'Martinez', 'robert.martinez@company.com', '555-0205', '2022-09-01', 'Engineering', 'Junior Software Engineer', 'full-time', 'active', 'Austin', 'TX', 'USA')
ON CONFLICT (email) DO NOTHING;

-- Sales Department
INSERT INTO employees (first_name, last_name, email, phone, hire_date, department, position, employment_type, status, city, state, country)
VALUES 
    ('Lisa', 'Anderson', 'lisa.anderson@company.com', '555-0301', '2020-07-01', 'Sales', 'Sales Manager', 'full-time', 'active', 'Chicago', 'IL', 'USA'),
    ('William', 'Thomas', 'william.thomas@company.com', '555-0302', '2021-04-15', 'Sales', 'Account Executive', 'full-time', 'active', 'Chicago', 'IL', 'USA'),
    ('Mary', 'Taylor', 'mary.taylor@company.com', '555-0303', '2021-11-01', 'Sales', 'Account Executive', 'full-time', 'active', 'Chicago', 'IL', 'USA'),
    ('Daniel', 'Moore', 'daniel.moore@company.com', '555-0304', '2022-05-20', 'Sales', 'Sales Representative', 'full-time', 'active', 'Dallas', 'TX', 'USA')
ON CONFLICT (email) DO NOTHING;

-- Marketing Department
INSERT INTO employees (first_name, last_name, email, phone, hire_date, department, position, employment_type, status, city, state, country)
VALUES 
    ('Patricia', 'Jackson', 'patricia.jackson@company.com', '555-0401', '2020-09-15', 'Marketing', 'Marketing Manager', 'full-time', 'active', 'Los Angeles', 'CA', 'USA'),
    ('Christopher', 'White', 'christopher.white@company.com', '555-0402', '2021-07-01', 'Marketing', 'Marketing Specialist', 'full-time', 'active', 'Los Angeles', 'CA', 'USA'),
    ('Jessica', 'Harris', 'jessica.harris@company.com', '555-0403', '2022-02-15', 'Marketing', 'Content Creator', 'full-time', 'active', 'Los Angeles', 'CA', 'USA')
ON CONFLICT (email) DO NOTHING;

-- Finance Department
INSERT INTO employees (first_name, last_name, email, phone, hire_date, department, position, employment_type, status, city, state, country)
VALUES 
    ('Charles', 'Martin', 'charles.martin@company.com', '555-0501', '2019-11-01', 'Finance', 'Finance Manager', 'full-time', 'active', 'New York', 'NY', 'USA'),
    ('Nancy', 'Thompson', 'nancy.thompson@company.com', '555-0502', '2021-03-20', 'Finance', 'Accountant', 'full-time', 'active', 'New York', 'NY', 'USA'),
    ('Thomas', 'Lee', 'thomas.lee@company.com', '555-0503', '2022-01-05', 'Finance', 'Financial Analyst', 'full-time', 'active', 'New York', 'NY', 'USA')
ON CONFLICT (email) DO NOTHING;

-- IT Department
INSERT INTO employees (first_name, last_name, email, phone, hire_date, department, position, employment_type, status, city, state, country)
VALUES 
    ('Kevin', 'Walker', 'kevin.walker@company.com', '555-0601', '2020-04-10', 'IT', 'IT Manager', 'full-time', 'active', 'Seattle', 'WA', 'USA'),
    ('Karen', 'Hall', 'karen.hall@company.com', '555-0602', '2021-09-15', 'IT', 'System Administrator', 'full-time', 'active', 'Seattle', 'WA', 'USA'),
    ('Steven', 'Allen', 'steven.allen@company.com', '555-0603', '2022-06-01', 'IT', 'Help Desk Technician', 'full-time', 'active', 'Seattle', 'WA', 'USA')
ON CONFLICT (email) DO NOTHING;

-- Contractors (1099)
INSERT INTO employees (first_name, last_name, email, phone, hire_date, department, position, employment_type, status, city, state, country)
VALUES 
    ('Alex', 'Consultant', 'alex.consultant@company.com', '555-0701', '2022-01-15', 'Consulting', 'Business Consultant', 'contractor', 'active', 'Remote', '', 'USA'),
    ('Sam', 'Designer', 'sam.designer@company.com', '555-0702', '2022-04-01', 'Design', 'Freelance Designer', 'contractor', 'active', 'Remote', '', 'USA'),
    ('Jordan', 'Writer', 'jordan.writer@company.com', '555-0703', '2022-07-10', 'Marketing', 'Content Writer', 'contractor', 'active', 'Remote', '', 'USA')
ON CONFLICT (email) DO NOTHING;

-- Set up manager relationships
UPDATE employees 
SET manager_id = (SELECT id FROM employees WHERE email = 'john.ceo@company.com')
WHERE position LIKE '%Manager%' AND email != 'john.ceo@company.com';

UPDATE employees 
SET manager_id = (SELECT id FROM employees WHERE email = 'jane.smith@company.com')
WHERE department = 'Human Resources' AND position NOT LIKE '%Manager%';

UPDATE employees 
SET manager_id = (SELECT id FROM employees WHERE email = 'david.wilson@company.com')
WHERE department = 'Engineering' AND position NOT LIKE '%Manager%';

UPDATE employees 
SET manager_id = (SELECT id FROM employees WHERE email = 'lisa.anderson@company.com')
WHERE department = 'Sales' AND position NOT LIKE '%Manager%';

UPDATE employees 
SET manager_id = (SELECT id FROM employees WHERE email = 'patricia.jackson@company.com')
WHERE department = 'Marketing' AND position NOT LIKE '%Manager%';

UPDATE employees 
SET manager_id = (SELECT id FROM employees WHERE email = 'charles.martin@company.com')
WHERE department = 'Finance' AND position NOT LIKE '%Manager%';

UPDATE employees 
SET manager_id = (SELECT id FROM employees WHERE email = 'kevin.walker@company.com')
WHERE department = 'IT' AND position NOT LIKE '%Manager%';

-- Display summary
DO $$
DECLARE
    emp_count INTEGER;
    dept_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO emp_count FROM employees;
    SELECT COUNT(DISTINCT department) INTO dept_count FROM employees;
    RAISE NOTICE 'Total employees created: %', emp_count;
    RAISE NOTICE 'Departments: %', dept_count;
END $$;
