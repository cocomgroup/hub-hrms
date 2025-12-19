-- HRMS Database Cleanup Script
-- Use this if you need to start fresh and drop all existing tables
-- WARNING: This will delete ALL data!

-- Disable foreign key checks temporarily
SET session_replication_role = 'replica';

-- Drop all tables in reverse dependency order
DROP TABLE IF EXISTS workflow_approvals CASCADE;
DROP TABLE IF EXISTS workflow_instances CASCADE;
DROP TABLE IF EXISTS workflow_steps CASCADE;
DROP TABLE IF EXISTS workflows CASCADE;

DROP TABLE IF EXISTS interviews CASCADE;
DROP TABLE IF EXISTS applications CASCADE;
DROP TABLE IF EXISTS candidates CASCADE;
DROP TABLE IF EXISTS job_postings CASCADE;

DROP TABLE IF EXISTS payroll_records CASCADE;
DROP TABLE IF EXISTS payroll_runs CASCADE;

DROP TABLE IF EXISTS timesheet_entries CASCADE;
DROP TABLE IF EXISTS timesheets CASCADE;

DROP TABLE IF EXISTS employee_benefits CASCADE;
DROP TABLE IF EXISTS benefits_plans CASCADE;

DROP TABLE IF EXISTS pto_balances CASCADE;
DROP TABLE IF EXISTS pto_requests CASCADE;
DROP TABLE IF EXISTS pto_policies CASCADE;

DROP TABLE IF EXISTS users CASCADE;

-- Re-enable foreign key checks
SET session_replication_role = 'origin';

-- Verify cleanup
SELECT 
    'All tables dropped successfully!' as status,
    COUNT(*) as remaining_tables
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_type = 'BASE TABLE';
