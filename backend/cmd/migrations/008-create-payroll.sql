-- Payroll System Migration
-- Supports both W2 employees and 1099 contractors

-- Employee Compensation Table
CREATE TABLE IF NOT EXISTS employee_compensation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    employment_type VARCHAR(10) NOT NULL CHECK (employment_type IN ('W2', '1099')),
    pay_type VARCHAR(20) NOT NULL CHECK (pay_type IN ('hourly', 'salary', 'commission')),
    hourly_rate DECIMAL(10, 2),
    annual_salary DECIMAL(12, 2),
    pay_frequency VARCHAR(20) NOT NULL CHECK (pay_frequency IN ('weekly', 'biweekly', 'semimonthly', 'monthly')),
    effective_date DATE NOT NULL,
    end_date DATE,
    overtime_eligible BOOLEAN DEFAULT FALSE,
    standard_hours_per_week DECIMAL(5, 2) DEFAULT 40.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_pay_rate CHECK (
        (pay_type = 'hourly' AND hourly_rate IS NOT NULL) OR
        (pay_type = 'salary' AND annual_salary IS NOT NULL) OR
        (pay_type = 'commission')
    )
);

CREATE INDEX IF NOT EXISTS idx_compensation_employee ON employee_compensation(employee_id);
CREATE INDEX IF NOT EXISTS idx_compensation_dates ON employee_compensation(effective_date, end_date);

-- W2 Tax Withholding Table
CREATE TABLE IF NOT EXISTS w2_tax_withholding (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL UNIQUE REFERENCES employees(id) ON DELETE CASCADE,
    filing_status VARCHAR(20) NOT NULL CHECK (filing_status IN ('single', 'married', 'head_of_household')),
    federal_allowances INTEGER DEFAULT 0,
    state_allowances INTEGER DEFAULT 0,
    additional_withholding DECIMAL(10, 2) DEFAULT 0.00,
    exempt_federal BOOLEAN DEFAULT FALSE,
    exempt_state BOOLEAN DEFAULT FALSE,
    exempt_fica BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tax_withholding_employee ON w2_tax_withholding(employee_id);

-- Pay Stub Earnings Table
CREATE TABLE IF NOT EXISTS pay_stub_earnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pay_stub_id UUID NOT NULL REFERENCES pay_stubs(id) ON DELETE CASCADE,
    earning_type VARCHAR(50) NOT NULL CHECK (earning_type IN ('regular', 'overtime', 'bonus', 'commission', 'contractor', 'other')),
    description TEXT NOT NULL,
    hours DECIMAL(10, 2),
    rate DECIMAL(10, 2),
    amount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_earnings_pay_stub ON pay_stub_earnings(pay_stub_id);

-- Pay Stub Deductions Table
CREATE TABLE IF NOT EXISTS pay_stub_deductions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pay_stub_id UUID NOT NULL REFERENCES pay_stubs(id) ON DELETE CASCADE,
    deduction_type VARCHAR(50) NOT NULL CHECK (deduction_type IN ('401k', 'health', 'dental', 'vision', 'life', 'fsa', 'hsa', 'other')),
    description TEXT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    employer_match DECIMAL(10, 2),
    pre_tax BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_deductions_pay_stub ON pay_stub_deductions(pay_stub_id);

-- Pay Stub Taxes Table
CREATE TABLE IF NOT EXISTS pay_stub_taxes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pay_stub_id UUID NOT NULL REFERENCES pay_stubs(id) ON DELETE CASCADE,
    tax_type VARCHAR(30) NOT NULL CHECK (tax_type IN ('federal', 'state', 'local', 'fica_ss', 'fica_medicare')),
    description TEXT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    taxable_wage DECIMAL(10, 2) NOT NULL,
    tax_rate DECIMAL(6, 4),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_taxes_pay_stub ON pay_stub_taxes(pay_stub_id);

-- Form 1099 Table (for contractors)
CREATE TABLE IF NOT EXISTS form_1099 (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    tax_year INTEGER NOT NULL,
    total_payments DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    federal_tax_withheld DECIMAL(10, 2) DEFAULT 0.00,
    state_tax_withheld DECIMAL(10, 2) DEFAULT 0.00,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'filed', 'corrected')),
    filed_date DATE,
    corrected_form_id UUID REFERENCES form_1099(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_id, tax_year)
);

CREATE INDEX IF NOT EXISTS idx_1099_employee ON form_1099(employee_id);
CREATE INDEX IF NOT EXISTS idx_1099_year ON form_1099(tax_year);
CREATE INDEX IF NOT EXISTS idx_1099_status ON form_1099(status);

-- Payroll Adjustments Table (for manual corrections)
CREATE TABLE IF NOT EXISTS payroll_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pay_stub_id UUID NOT NULL REFERENCES pay_stubs(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('earning', 'deduction', 'tax')),
    category VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    reason TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_adjustments_pay_stub ON payroll_adjustments(pay_stub_id);
CREATE INDEX IF NOT EXISTS idx_adjustments_employee ON payroll_adjustments(employee_id);

-- Update existing payroll_periods table if needed
ALTER TABLE payroll_periods ADD COLUMN IF NOT EXISTS processed_by UUID REFERENCES users(id);
ALTER TABLE payroll_periods ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP;

-- Create views for common queries

-- View: Current Employee Compensation
CREATE OR REPLACE VIEW v_current_compensation AS
SELECT 
    ec.*,
    e.first_name,
    e.last_name,
    e.email,
    e.department,
    e.position,
    e.status AS employee_status
FROM employee_compensation ec
JOIN employees e ON ec.employee_id = e.id
WHERE ec.end_date IS NULL OR ec.end_date > CURRENT_DATE
ORDER BY e.last_name, e.first_name;

-- View: Pay Stub Summary with Employee Info
CREATE OR REPLACE VIEW v_pay_stub_summary AS
SELECT 
    ps.*,
    e.first_name,
    e.last_name,
    e.email,
    e.department,
    pp.start_date AS period_start,
    pp.end_date AS period_end,
    pp.pay_date,
    ec.employment_type,
    ec.pay_type
FROM pay_stubs ps
JOIN employees e ON ps.employee_id = e.id
JOIN payroll_periods pp ON ps.payroll_period_id = pp.id
LEFT JOIN employee_compensation ec ON e.id = ec.employee_id 
    AND (ec.end_date IS NULL OR ec.end_date > pp.end_date);

-- View: 1099 Summary
CREATE OR REPLACE VIEW v_1099_summary AS
SELECT 
    f.*,
    e.first_name,
    e.last_name,
    e.email,
    e.street_address,
    e.city,
    e.state,
    e.zip_code
FROM form_1099 f
JOIN employees e ON f.employee_id = e.id;

-- Function to calculate YTD earnings
CREATE OR REPLACE FUNCTION get_ytd_earnings(p_employee_id UUID, p_year INTEGER)
RETURNS DECIMAL(12,2) AS $$
BEGIN
    RETURN (
        SELECT COALESCE(SUM(gross_pay), 0)
        FROM pay_stubs ps
        JOIN payroll_periods pp ON ps.payroll_period_id = pp.id
        WHERE ps.employee_id = p_employee_id
          AND EXTRACT(YEAR FROM pp.end_date) = p_year
    );
END;
$$ LANGUAGE plpgsql;

-- Function to calculate YTD taxes
CREATE OR REPLACE FUNCTION get_ytd_taxes(p_employee_id UUID, p_year INTEGER)
RETURNS DECIMAL(12,2) AS $$
BEGIN
    RETURN (
        SELECT COALESCE(SUM(federal_tax + state_tax + social_security + medicare), 0)
        FROM pay_stubs ps
        JOIN payroll_periods pp ON ps.payroll_period_id = pp.id
        WHERE ps.employee_id = p_employee_id
          AND EXTRACT(YEAR FROM pp.end_date) = p_year
    );
END;
$$ LANGUAGE plpgsql;

-- Trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- FIXED: Triggers for employee_compensation
DROP TRIGGER IF EXISTS update_employee_compensation_updated_at ON employee_compensation;
CREATE TRIGGER update_employee_compensation_updated_at
    BEFORE UPDATE ON employee_compensation
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();

-- FIXED: Trigger for w2_tax_withholding
DROP TRIGGER IF EXISTS update_w2_tax_withholding_updated_at ON w2_tax_withholding;
CREATE TRIGGER update_w2_tax_withholding_updated_at
    BEFORE UPDATE ON w2_tax_withholding
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();

-- FIXED: Trigger for form_1099
DROP TRIGGER IF EXISTS update_form_1099_updated_at ON form_1099;
CREATE TRIGGER update_form_1099_updated_at
    BEFORE UPDATE ON form_1099
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();

-- Sample data (optional - for testing)

-- Add sample W2 employee compensation
--INSERT INTO employee_compensation (employee_id, employment_type, pay_type, annual_salary, pay_frequency, effective_date, overtime_eligible)
--SELECT 
--    id,
--    'W2',
--    'salary',
--    75000.00,
--    'biweekly',
--    hire_date,
--    FALSE
----FROM employees 
--WHERE employment_type = 'full-time' 
--  AND NOT EXISTS (SELECT 1 FROM employee_compensation WHERE employee_id = employees.id)
--LIMIT 5;

-- Add sample 1099 contractor compensation
--INSERT INTO employee_compensation (employee_id, employment_type, pay_type, hourly_rate, pay_frequency, effective_date, overtime_eligible)
--SELECT 
--   id,
--    '1099',
--    'hourly',
--    85.00,
--    'weekly',
--    hire_date,
--    FALSE
--FROM employees 
--WHERE employment_type = 'contractor' 
--  AND NOT EXISTS (SELECT 1 FROM employee_compensation WHERE employee_id = employees.id)
--LIMIT 2;

COMMIT;