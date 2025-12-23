-- Seed Data: PTO and Benefits
-- File: 016-seed-pto-benefits.sql
-- Description: Populate PTO balances and benefit enrollments

-- Note: Schema differences from original:
-- - pto_balances: No accrual_rate fields, has 'year' field, UNIQUE(employee_id, year)
-- - pto_requests: No requested_date (uses created_at), no approved_date/approved_by (uses reviewed_at/reviewed_by)
-- - benefit_plans: No UNIQUE constraint on (name, enrollment_start_date)

-- PTO Balances for all active employees (for current year)
INSERT INTO pto_balances (employee_id, vacation_days, sick_days, personal_days, year)
SELECT 
    id,
    CASE 
        WHEN EXTRACT(YEAR FROM AGE(CURRENT_DATE, hire_date)) >= 5 THEN 20.0
        WHEN EXTRACT(YEAR FROM AGE(CURRENT_DATE, hire_date)) >= 2 THEN 15.0
        ELSE 10.0
    END,
    10.0,
    5.0,
    EXTRACT(YEAR FROM CURRENT_DATE)::INT
FROM employees 
WHERE status = 'active' 
  AND employment_type = 'full-time'
ON CONFLICT (employee_id, year) DO NOTHING;

-- Sample PTO Requests
INSERT INTO pto_requests (employee_id, pto_type, start_date, end_date, days_requested, reason, status)
SELECT 
    e.id,
    'vacation',
    CURRENT_DATE + INTERVAL '30 days',
    CURRENT_DATE + INTERVAL '37 days',
    7.0,
    'Family vacation',
    'pending'
FROM employees e
WHERE e.email = 'emily.davis@company.com';

INSERT INTO pto_requests (employee_id, pto_type, start_date, end_date, days_requested, reason, status, reviewed_by, reviewed_at)
SELECT 
    e.id,
    'sick',
    CURRENT_DATE - INTERVAL '5 days',
    CURRENT_DATE - INTERVAL '3 days',
    2.0,
    'Flu recovery',
    'approved',
    e.manager_id,
    CURRENT_DATE - INTERVAL '5 days'
FROM employees e
WHERE e.email = 'james.miller@company.com'
  AND e.manager_id IS NOT NULL;

-- Sample additional PTO requests
INSERT INTO pto_requests (employee_id, pto_type, start_date, end_date, days_requested, reason, status)
SELECT 
    e.id,
    'personal',
    CURRENT_DATE + INTERVAL '15 days',
    CURRENT_DATE + INTERVAL '15 days',
    1.0,
    'Personal appointment',
    'approved'
FROM employees e
WHERE e.email = 'david.wilson@company.com';

INSERT INTO pto_requests (employee_id, pto_type, start_date, end_date, days_requested, reason, status)
SELECT 
    e.id,
    'vacation',
    CURRENT_DATE + INTERVAL '60 days',
    CURRENT_DATE + INTERVAL '74 days',
    10.0,
    'Summer vacation',
    'pending'
FROM employees e
WHERE e.department = 'Engineering'
  AND e.status = 'active'
LIMIT 1;

-- Benefit Plans (no unique constraint, check before insert)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM benefit_plans WHERE name = 'Premium Health Plan' AND enrollment_start_date = '2025-11-01') THEN
        INSERT INTO benefit_plans (name, category, plan_type, provider, description, employee_cost, employer_cost, deductible_single, deductible_family, out_of_pocket_max_single, out_of_pocket_max_family, copay_primary_care, copay_specialist, enrollment_start_date, enrollment_end_date, effective_date)
        VALUES ('Premium Health Plan', 'health', 'ppo', 'Blue Cross Blue Shield', 'Comprehensive PPO health coverage', 150.00, 450.00, 1000.00, 2000.00, 3000.00, 6000.00, 20.00, 40.00, '2025-11-01', '2025-11-30', '2026-01-01');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM benefit_plans WHERE name = 'Basic Health Plan' AND enrollment_start_date = '2025-11-01') THEN
        INSERT INTO benefit_plans (name, category, plan_type, provider, description, employee_cost, employer_cost, deductible_single, deductible_family, out_of_pocket_max_single, out_of_pocket_max_family, copay_primary_care, copay_specialist, enrollment_start_date, enrollment_end_date, effective_date)
        VALUES ('Basic Health Plan', 'health', 'hmo', 'Kaiser Permanente', 'Affordable HMO health coverage', 75.00, 300.00, 2000.00, 4000.00, 5000.00, 10000.00, 30.00, 50.00, '2025-11-01', '2025-11-30', '2026-01-01');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM benefit_plans WHERE name = 'Dental Plan' AND enrollment_start_date = '2025-11-01') THEN
        INSERT INTO benefit_plans (name, category, plan_type, provider, description, employee_cost, employer_cost, deductible_single, deductible_family, out_of_pocket_max_single, out_of_pocket_max_family, enrollment_start_date, enrollment_end_date, effective_date)
        VALUES ('Dental Plan', 'dental', 'traditional', 'Delta Dental', 'Comprehensive dental coverage', 25.00, 50.00, 50.00, 150.00, 1500.00, 3000.00, '2025-11-01', '2025-11-30', '2026-01-01');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM benefit_plans WHERE name = 'Vision Plan' AND enrollment_start_date = '2025-11-01') THEN
        INSERT INTO benefit_plans (name, category, plan_type, provider, description, employee_cost, employer_cost, enrollment_start_date, enrollment_end_date, effective_date)
        VALUES ('Vision Plan', 'vision', 'traditional', 'VSP', 'Vision care coverage', 10.00, 15.00, '2025-11-01', '2025-11-30', '2026-01-01');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM benefit_plans WHERE name = '401(k) Plan' AND enrollment_start_date = '2025-11-01') THEN
        INSERT INTO benefit_plans (name, category, plan_type, provider, description, employee_cost, employer_cost, enrollment_start_date, enrollment_end_date, effective_date)
        VALUES ('401(k) Plan', 'retirement', 'traditional', 'Fidelity', 'Company 401(k) with 4% match', 0.00, 0.00, '2025-11-01', '2025-11-30', '2026-01-01');
    END IF;
END $$;

-- Benefit Enrollments for employees (Premium Health)
INSERT INTO benefit_enrollments (employee_id, plan_id, coverage_level, status, enrollment_date, effective_date, employee_cost, employer_cost, total_cost, payroll_deduction)
SELECT 
    e.id,
    bp.id,
    'employee',
    'active',
    '2025-11-15',
    '2026-01-01',
    bp.employee_cost,
    bp.employer_cost,
    bp.employee_cost + bp.employer_cost,
    bp.employee_cost / 2  -- Biweekly
FROM employees e
CROSS JOIN benefit_plans bp
WHERE e.status = 'active'
  AND e.employment_type = 'full-time'
  AND bp.category = 'health'
  AND bp.name = 'Premium Health Plan'
  AND e.id IN (SELECT id FROM employees WHERE status = 'active' AND employment_type = 'full-time' ORDER BY hire_date LIMIT 10)
  AND NOT EXISTS (
      SELECT 1 FROM benefit_enrollments be 
      WHERE be.employee_id = e.id AND be.plan_id = bp.id
  );

-- Benefit Enrollments for employees (Dental)
INSERT INTO benefit_enrollments (employee_id, plan_id, coverage_level, status, enrollment_date, effective_date, employee_cost, employer_cost, total_cost, payroll_deduction)
SELECT 
    e.id,
    bp.id,
    'employee',
    'active',
    '2025-11-15',
    '2026-01-01',
    bp.employee_cost,
    bp.employer_cost,
    bp.employee_cost + bp.employer_cost,
    bp.employee_cost / 2
FROM employees e
CROSS JOIN benefit_plans bp
WHERE e.status = 'active'
  AND e.employment_type = 'full-time'
  AND bp.category = 'dental'
  AND e.id IN (SELECT id FROM employees WHERE status = 'active' AND employment_type = 'full-time' ORDER BY hire_date LIMIT 15)
  AND NOT EXISTS (
      SELECT 1 FROM benefit_enrollments be 
      WHERE be.employee_id = e.id AND be.plan_id = bp.id
  );

-- Benefit Enrollments for employees (Vision)
INSERT INTO benefit_enrollments (employee_id, plan_id, coverage_level, status, enrollment_date, effective_date, employee_cost, employer_cost, total_cost, payroll_deduction)
SELECT 
    e.id,
    bp.id,
    'employee',
    'active',
    '2025-11-15',
    '2026-01-01',
    bp.employee_cost,
    bp.employer_cost,
    bp.employee_cost + bp.employer_cost,
    bp.employee_cost / 2
FROM employees e
CROSS JOIN benefit_plans bp
WHERE e.status = 'active'
  AND e.employment_type = 'full-time'
  AND bp.category = 'vision'
  AND e.id IN (SELECT id FROM employees WHERE status = 'active' AND employment_type = 'full-time' ORDER BY hire_date LIMIT 12)
  AND NOT EXISTS (
      SELECT 1 FROM benefit_enrollments be 
      WHERE be.employee_id = e.id AND be.plan_id = bp.id
  );

-- Enrollment Period (no unique constraint, check before insert)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM enrollment_periods WHERE name = '2026 Open Enrollment') THEN
        INSERT INTO enrollment_periods (name, description, start_date, end_date, plan_year, active)
        VALUES ('2026 Open Enrollment', 'Annual open enrollment for 2026 benefit year', '2025-11-01', '2025-11-30', 2026, true);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM enrollment_periods WHERE name = '2025 Open Enrollment') THEN
        INSERT INTO enrollment_periods (name, description, start_date, end_date, plan_year, active)
        VALUES ('2025 Open Enrollment', '2025 benefit year enrollment period', '2024-11-01', '2024-11-30', 2025, false);
    END IF;
END $$;

-- Display summary
DO $$
DECLARE
    pto_count INTEGER;
    pto_request_count INTEGER;
    benefit_plan_count INTEGER;
    enrollment_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO pto_count FROM pto_balances;
    SELECT COUNT(*) INTO pto_request_count FROM pto_requests;
    SELECT COUNT(*) INTO benefit_plan_count FROM benefit_plans;
    SELECT COUNT(*) INTO enrollment_count FROM benefit_enrollments;
    
    RAISE NOTICE '================================';
    RAISE NOTICE 'PTO & Benefits Seed Data Summary';
    RAISE NOTICE '================================';
    RAISE NOTICE 'PTO balances: %', pto_count;
    RAISE NOTICE 'PTO requests: %', pto_request_count;
    RAISE NOTICE 'Benefit plans: %', benefit_plan_count;
    RAISE NOTICE 'Benefit enrollments: %', enrollment_count;
    RAISE NOTICE '================================';
    RAISE NOTICE 'Benefits Available:';
    RAISE NOTICE '  - Premium Health (PPO)';
    RAISE NOTICE '  - Basic Health (HMO)';
    RAISE NOTICE '  - Dental';
    RAISE NOTICE '  - Vision';
    RAISE NOTICE '  - 401(k)';
    RAISE NOTICE '================================';
END $$;