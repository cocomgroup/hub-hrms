-- PTO (Paid Time Off) Management System Migration
-- Creates tables for PTO balances and requests

-- ============================================================================
-- PTO Balances Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS pto_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    vacation_days DECIMAL(5, 2) NOT NULL DEFAULT 0,
    sick_days DECIMAL(5, 2) NOT NULL DEFAULT 0,
    personal_days DECIMAL(5, 2) NOT NULL DEFAULT 0,
    year INT NOT NULL DEFAULT EXTRACT(YEAR FROM CURRENT_DATE),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(employee_id, year)
);

CREATE INDEX idx_pto_balances_employee ON pto_balances(employee_id);
CREATE INDEX idx_pto_balances_year ON pto_balances(year);

COMMENT ON TABLE pto_balances IS 'Tracks available PTO days for each employee per year';
COMMENT ON COLUMN pto_balances.vacation_days IS 'Available vacation days';
COMMENT ON COLUMN pto_balances.sick_days IS 'Available sick leave days';
COMMENT ON COLUMN pto_balances.personal_days IS 'Available personal days';

-- ============================================================================
-- PTO Requests Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS pto_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    pto_type VARCHAR(20) NOT NULL CHECK (pto_type IN ('vacation', 'sick', 'personal')),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    days_requested DECIMAL(5, 2) NOT NULL,
    reason TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'denied', 'cancelled')),
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP,
    review_notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_date_range CHECK (end_date >= start_date),
    CONSTRAINT valid_days CHECK (days_requested > 0)
);

CREATE INDEX idx_pto_requests_employee ON pto_requests(employee_id);
CREATE INDEX idx_pto_requests_status ON pto_requests(status);
CREATE INDEX idx_pto_requests_dates ON pto_requests(start_date, end_date);
CREATE INDEX idx_pto_requests_type ON pto_requests(pto_type);

COMMENT ON TABLE pto_requests IS 'Stores PTO requests from employees';
COMMENT ON COLUMN pto_requests.pto_type IS 'Type of PTO: vacation, sick, or personal';
COMMENT ON COLUMN pto_requests.status IS 'Request status: pending, approved, denied, or cancelled';
COMMENT ON COLUMN pto_requests.days_requested IS 'Number of days requested (can be decimal for half-days)';

-- ============================================================================
-- Seed Data: Create default PTO balances for existing employees
-- ============================================================================

-- Default PTO allocation per year:
-- - 15 vacation days
-- - 10 sick days
-- - 5 personal days

INSERT INTO pto_balances (employee_id, vacation_days, sick_days, personal_days, year)
SELECT 
    id,
    15.0,  -- vacation days
    10.0,  -- sick days
    5.0,   -- personal days
    EXTRACT(YEAR FROM CURRENT_DATE)::INT
FROM employees
WHERE NOT EXISTS (
    SELECT 1 
    FROM pto_balances 
    WHERE pto_balances.employee_id = employees.id 
    AND pto_balances.year = EXTRACT(YEAR FROM CURRENT_DATE)
)
ON CONFLICT (employee_id, year) DO NOTHING;

-- ============================================================================
-- Trigger: Auto-update updated_at timestamp
-- ============================================================================

CREATE OR REPLACE FUNCTION update_pto_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_pto_balances_updated_at
    BEFORE UPDATE ON pto_balances
    FOR EACH ROW
    EXECUTE FUNCTION update_pto_timestamp();

CREATE TRIGGER trigger_pto_requests_updated_at
    BEFORE UPDATE ON pto_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_pto_timestamp();

-- ============================================================================
-- Sample Data (Optional - for testing)
-- ============================================================================

-- Uncomment to insert sample PTO requests
/*
INSERT INTO pto_requests (employee_id, pto_type, start_date, end_date, days_requested, reason, status)
SELECT 
    e.id,
    'vacation',
    CURRENT_DATE + INTERVAL '30 days',
    CURRENT_DATE + INTERVAL '34 days',
    5.0,
    'Family vacation',
    'pending'
FROM employees e
LIMIT 1;
*/

-- ============================================================================
-- Views for convenience
-- ============================================================================

CREATE OR REPLACE VIEW pto_requests_with_employee AS
SELECT 
    pr.id,
    pr.employee_id,
    CONCAT(e.first_name, ' ', e.last_name) as employee_name,
    e.department,
    pr.pto_type,
    pr.start_date,
    pr.end_date,
    pr.days_requested,
    pr.reason,
    pr.status,
    pr.reviewed_by,
    CASE 
        WHEN pr.reviewed_by IS NOT NULL 
        THEN CONCAT(u.first_name, ' ', u.last_name)
        ELSE NULL
    END as reviewer_name,
    pr.reviewed_at,
    pr.review_notes,
    pr.created_at,
    pr.updated_at
FROM pto_requests pr
JOIN employees e ON e.id = pr.employee_id
LEFT JOIN users u ON u.id = pr.reviewed_by;

COMMENT ON VIEW pto_requests_with_employee IS 'PTO requests with employee and reviewer details';

-- ============================================================================
-- Query Examples
-- ============================================================================

-- Get employee's current PTO balance
-- SELECT * FROM pto_balances WHERE employee_id = 'xxx' AND year = EXTRACT(YEAR FROM CURRENT_DATE);

-- Get all pending PTO requests
-- SELECT * FROM pto_requests_with_employee WHERE status = 'pending';

-- Get PTO requests for a date range
-- SELECT * FROM pto_requests WHERE start_date >= '2024-01-01' AND end_date <= '2024-12-31';

-- Get employee's PTO history
-- SELECT * FROM pto_requests WHERE employee_id = 'xxx' ORDER BY created_at DESC;
