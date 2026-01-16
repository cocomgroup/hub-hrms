-- Migration: Create Benefits tables
-- File: migrations/007_create_benefits_tables.up.sql

-- Benefit Plans table
CREATE TABLE IF NOT EXISTS benefit_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    category VARCHAR(50) NOT NULL CHECK (category IN ('health', 'dental', 'vision', 'life', 'disability', 'retirement', 'fsa', 'hsa', 'commuter', 'wellness', 'other')),
    plan_type VARCHAR(50) CHECK (plan_type IN ('hmo', 'ppo', 'epo', 'pos', 'hdhp', 'traditional')),
    provider VARCHAR(200) NOT NULL,
    description TEXT,
    employee_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    employer_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    deductible_single DECIMAL(10,2) DEFAULT 0,
    deductible_family DECIMAL(10,2) DEFAULT 0,
    out_of_pocket_max_single DECIMAL(10,2) DEFAULT 0,
    out_of_pocket_max_family DECIMAL(10,2) DEFAULT 0,
    copay_primary_care DECIMAL(10,2) DEFAULT 0,
    copay_specialist DECIMAL(10,2) DEFAULT 0,
    copay_emergency DECIMAL(10,2) DEFAULT 0,
    coinsurance_rate DECIMAL(5,2) DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT true,
    enrollment_start_date DATE NOT NULL,
    enrollment_end_date DATE NOT NULL,
    effective_date DATE NOT NULL,
    termination_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (enrollment_end_date >= enrollment_start_date),
    CHECK (employee_cost >= 0),
    CHECK (employer_cost >= 0)
);

-- Benefit Enrollments table
CREATE TABLE IF NOT EXISTS benefit_enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES benefit_plans(id),
    coverage_level VARCHAR(50) NOT NULL CHECK (coverage_level IN ('employee', 'employee_spouse', 'employee_child', 'family')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'pending', 'cancelled', 'expired')),
    enrollment_date DATE NOT NULL,
    effective_date DATE NOT NULL,
    termination_date DATE,
    employee_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    employer_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    total_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    payroll_deduction DECIMAL(10,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (employee_cost >= 0),
    CHECK (employer_cost >= 0),
    CHECK (total_cost >= 0),
    CHECK (payroll_deduction >= 0)
);

-- Benefit Dependents table
CREATE TABLE IF NOT EXISTS benefit_dependents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enrollment_id UUID NOT NULL REFERENCES benefit_enrollments(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    relationship VARCHAR(50) NOT NULL CHECK (relationship IN ('spouse', 'child', 'domestic_partner', 'parent', 'other')),
    date_of_birth DATE NOT NULL,
    ssn VARCHAR(11),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Benefit Claims table (optional - for tracking claims)
CREATE TABLE IF NOT EXISTS benefit_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enrollment_id UUID NOT NULL REFERENCES benefit_enrollments(id),
    employee_id UUID NOT NULL REFERENCES employees(id),
    plan_id UUID NOT NULL REFERENCES benefit_plans(id),
    claim_number VARCHAR(100) UNIQUE NOT NULL,
    service_date DATE NOT NULL,
    provider VARCHAR(200),
    service_type VARCHAR(100),
    claim_amount DECIMAL(10,2) NOT NULL,
    approved_amount DECIMAL(10,2) DEFAULT 0,
    paid_amount DECIMAL(10,2) DEFAULT 0,
    employee_portion DECIMAL(10,2) DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'submitted' CHECK (status IN ('submitted', 'processing', 'approved', 'denied', 'paid')),
    submitted_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_date TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Enrollment Periods table (for open enrollment tracking)
CREATE TABLE IF NOT EXISTS enrollment_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    plan_year INT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (end_date >= start_date)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_benefit_plans_category ON benefit_plans(category);
CREATE INDEX IF NOT EXISTS idx_benefit_plans_active ON benefit_plans(active);
CREATE INDEX IF NOT EXISTS idx_benefit_enrollments_employee ON benefit_enrollments(employee_id);
CREATE INDEX IF NOT EXISTS idx_benefit_enrollments_plan ON benefit_enrollments(plan_id);
CREATE INDEX IF NOT EXISTS idx_benefit_enrollments_status ON benefit_enrollments(status);
CREATE INDEX IF NOT EXISTS idx_benefit_dependents_enrollment ON benefit_dependents(enrollment_id);
CREATE INDEX IF NOT EXISTS idx_benefit_claims_employee ON benefit_claims(employee_id);
CREATE INDEX IF NOT EXISTS idx_benefit_claims_enrollment ON benefit_claims(enrollment_id);
CREATE INDEX IF NOT EXISTS idx_benefit_claims_status ON benefit_claims(status);

-- Insert sample benefit plans
INSERT INTO benefit_plans (
    name, category, plan_type, provider, description,
    employee_cost, employer_cost,
    deductible_single, deductible_family,
    out_of_pocket_max_single, out_of_pocket_max_family,
    copay_primary_care, copay_specialist, copay_emergency,
    coinsurance_rate, active,
    enrollment_start_date, enrollment_end_date, effective_date
) VALUES
    -- Health Insurance Plans
    ('Blue Cross HMO', 'health', 'hmo', 'Blue Cross Blue Shield',
     'Comprehensive HMO plan with low out-of-pocket costs',
     2400.00, 9600.00, 1000.00, 2000.00, 3000.00, 6000.00,
     25.00, 50.00, 150.00, 20.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    ('Blue Cross PPO', 'health', 'ppo', 'Blue Cross Blue Shield',
     'Flexible PPO plan with broader network access',
     3000.00, 11000.00, 1500.00, 3000.00, 4000.00, 8000.00,
     30.00, 60.00, 200.00, 20.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    ('High Deductible Health Plan', 'health', 'hdhp', 'Aetna',
     'Lower premium plan with HSA eligibility',
     1800.00, 8200.00, 3000.00, 6000.00, 6000.00, 12000.00,
     0.00, 0.00, 0.00, 20.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    -- Dental Insurance
    ('Delta Dental PPO', 'dental', 'ppo', 'Delta Dental',
     'Comprehensive dental coverage',
     360.00, 240.00, 50.00, 150.00, 1500.00, 3000.00,
     0.00, 0.00, 0.00, 20.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    -- Vision Insurance
    ('VSP Vision Care', 'vision', 'traditional', 'VSP',
     'Comprehensive vision coverage',
     120.00, 80.00, 0.00, 0.00, 0.00, 0.00,
     10.00, 0.00, 0.00, 0.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    -- Life Insurance
    ('Basic Life Insurance', 'life', 'traditional', 'MetLife',
     'Company-paid life insurance (1x salary)',
     0.00, 200.00, 0.00, 0.00, 0.00, 0.00,
     0.00, 0.00, 0.00, 0.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    ('Supplemental Life Insurance', 'life', 'traditional', 'MetLife',
     'Additional voluntary life insurance',
     180.00, 0.00, 0.00, 0.00, 0.00, 0.00,
     0.00, 0.00, 0.00, 0.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    -- Disability
    ('Short-Term Disability', 'disability', 'traditional', 'Prudential',
     'Short-term disability coverage (60% salary)',
     240.00, 160.00, 0.00, 0.00, 0.00, 0.00,
     0.00, 0.00, 0.00, 0.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    ('Long-Term Disability', 'disability', 'traditional', 'Prudential',
     'Long-term disability coverage (60% salary)',
     300.00, 200.00, 0.00, 0.00, 0.00, 0.00,
     0.00, 0.00, 0.00, 0.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    -- FSA/HSA
    ('Health Savings Account', 'hsa', 'traditional', 'HealthEquity',
     'HSA for HDHP participants',
     0.00, 500.00, 0.00, 0.00, 0.00, 0.00,
     0.00, 0.00, 0.00, 0.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    ('Flexible Spending Account', 'fsa', 'traditional', 'WageWorks',
     'Healthcare FSA',
     0.00, 0.00, 0.00, 0.00, 0.00, 0.00,
     0.00, 0.00, 0.00, 0.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01'),
    
    -- 401(k)
    ('401(k) Retirement Plan', 'retirement', 'traditional', 'Fidelity',
     'Company 401(k) with 4% match',
     0.00, 0.00, 0.00, 0.00, 0.00, 0.00,
     0.00, 0.00, 0.00, 0.00, true,
     '2025-11-01', '2025-11-30', '2026-01-01')
ON CONFLICT DO NOTHING;

-- Insert current enrollment period
INSERT INTO enrollment_periods (
    name, description, start_date, end_date, plan_year, active
) VALUES (
    '2026 Open Enrollment',
    'Annual open enrollment for 2026 benefit year',
    '2025-11-01',
    '2025-11-30',
    2026,
    true
) ON CONFLICT DO NOTHING;
