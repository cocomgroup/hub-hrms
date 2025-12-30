-- Timesheet System Database Schema
-- Migration: Add timesheet tables

-- Time entries table (for clock in/out and manual entries)
CREATE TABLE IF NOT EXISTS time_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    entry_date DATE NOT NULL,
    clock_in TIMESTAMP,
    clock_out TIMESTAMP,
    break_duration INTEGER DEFAULT 0, -- in minutes
    notes TEXT,
    entry_type VARCHAR(20) DEFAULT 'regular', -- regular, overtime, pto, sick, holiday
    status VARCHAR(20) DEFAULT 'draft', -- draft, submitted, approved, rejected
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP,
    rejection_reason TEXT,
    total_hours DECIMAL(5,2), -- calculated field
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_clock_times CHECK (clock_out IS NULL OR clock_out > clock_in),
    CONSTRAINT check_break_duration CHECK (break_duration >= 0)
);

ALTER TABLE time_entries
ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS rejection_reason TEXT;

CREATE INDEX IF NOT EXISTS idx_time_entries_status 
ON time_entries(status);

CREATE INDEX IF NOT EXISTS idx_time_entries_approved_by 
ON time_entries(approved_by);

-- Timesheet periods table (weekly or bi-weekly periods)
CREATE TABLE IF NOT EXISTS timesheet_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'draft', -- draft, submitted, approved, rejected
    total_regular_hours DECIMAL(6,2) DEFAULT 0,
    total_overtime_hours DECIMAL(6,2) DEFAULT 0,
    total_pto_hours DECIMAL(6,2) DEFAULT 0,
    submitted_at TIMESTAMP,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP,
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_period_dates CHECK (end_date > start_date),
    CONSTRAINT unique_employee_period UNIQUE(employee_id, start_date, end_date)
);

-- Projects table (for project-based time tracking)
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    client_name VARCHAR(200),
    start_date DATE,
    end_date DATE,
    budget_hours DECIMAL(8,2),
    status VARCHAR(20) DEFAULT 'active', -- active, on_hold, completed, cancelled
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Time entry projects (many-to-many: time entries can be split across projects)
CREATE TABLE IF NOT EXISTS time_entry_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    time_entry_id UUID NOT NULL REFERENCES time_entries(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    hours DECIMAL(5,2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_hours_positive CHECK (hours > 0),
    CONSTRAINT unique_entry_project UNIQUE(time_entry_id, project_id)
);

-- Overtime rules table
CREATE TABLE IF NOT EXISTS overtime_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    daily_threshold DECIMAL(4,2) DEFAULT 8.0, -- hours per day before overtime
    weekly_threshold DECIMAL(5,2) DEFAULT 40.0, -- hours per week before overtime
    overtime_multiplier DECIMAL(3,2) DEFAULT 1.5, -- 1.5x for overtime
    double_time_threshold DECIMAL(5,2), -- hours before double time (optional)
    double_time_multiplier DECIMAL(3,2) DEFAULT 2.0,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Employee overtime assignments
CREATE TABLE IF NOT EXISTS employee_overtime_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    overtime_rule_id UUID NOT NULL REFERENCES overtime_rules(id) ON DELETE CASCADE,
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT unique_employee_rule UNIQUE(employee_id, overtime_rule_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_time_entries_employee ON time_entries(employee_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_date ON time_entries(entry_date);
CREATE INDEX IF NOT EXISTS idx_time_entries_status ON time_entries(status);
CREATE INDEX IF NOT EXISTS idx_timesheet_periods_employee ON timesheet_periods(employee_id);
CREATE INDEX IF NOT EXISTS idx_timesheet_periods_dates ON timesheet_periods(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_timesheet_periods_status ON timesheet_periods(status);
CREATE INDEX IF NOT EXISTS idx_time_entry_projects_entry ON time_entry_projects(time_entry_id);
CREATE INDEX IF NOT EXISTS idx_time_entry_projects_project ON time_entry_projects(project_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_code ON projects(code);

-- Function to calculate total hours for time entry
CREATE OR REPLACE FUNCTION calculate_time_entry_hours()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.clock_in IS NOT NULL AND NEW.clock_out IS NOT NULL THEN
        -- ✅ FIX: Cast to NUMERIC before ROUND to avoid double precision error
        NEW.total_hours = ROUND(
            (
                (EXTRACT(EPOCH FROM (NEW.clock_out - NEW.clock_in)) / 3600.0) - 
                (COALESCE(NEW.break_duration, 0) / 60.0)
            )::NUMERIC,  -- ✅ Cast to NUMERIC
            2
        );
    END IF;
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-calculate hours
DROP TRIGGER IF EXISTS trigger_calculate_time_entry_hours ON time_entries;
CREATE TRIGGER trigger_calculate_time_entry_hours
    BEFORE INSERT OR UPDATE ON time_entries
    FOR EACH ROW
    EXECUTE PROCEDURE calculate_time_entry_hours();

-- Function to update timesheet period totals
DROP FUNCTION IF EXISTS update_timesheet_period_totals() CASCADE;

CREATE OR REPLACE FUNCTION update_timesheet_period_totals()
RETURNS TRIGGER AS $$
DECLARE
    period_record RECORD;
    target_employee_id UUID;
    target_entry_date DATE;
BEGIN
    -- Determine which record to use based on operation
    IF TG_OP = 'DELETE' THEN
        -- For DELETE, use OLD record
        target_employee_id := OLD.employee_id;
        target_entry_date := OLD.entry_date;
    ELSE
        -- For INSERT/UPDATE, use NEW record
        target_employee_id := NEW.employee_id;
        target_entry_date := NEW.entry_date;
    END IF;

    -- Find the period this entry belongs to
    SELECT *
    INTO period_record
    FROM timesheet_periods
    WHERE employee_id = target_employee_id
      AND target_entry_date BETWEEN start_date AND end_date
    LIMIT 1;

    -- If no period found, skip (return appropriate record)
    IF NOT FOUND THEN
        IF TG_OP = 'DELETE' THEN
            RETURN OLD;
        ELSE
            RETURN NEW;
        END IF;
    END IF;

    -- Recalculate totals for this period
    UPDATE timesheet_periods
    SET 
        total_regular_hours = COALESCE((
            SELECT SUM(total_hours)
            FROM time_entries
            WHERE employee_id = period_record.employee_id
              AND entry_date BETWEEN period_record.start_date AND period_record.end_date
              AND entry_type = 'regular'
              AND status IN ('draft', 'submitted', 'approved')
        ), 0),
        total_overtime_hours = COALESCE((
            SELECT SUM(total_hours)
            FROM time_entries
            WHERE employee_id = period_record.employee_id
              AND entry_date BETWEEN period_record.start_date AND period_record.end_date
              AND entry_type = 'overtime'
              AND status IN ('draft', 'submitted', 'approved')
        ), 0),
        total_pto_hours = COALESCE((
            SELECT SUM(total_hours)
            FROM time_entries
            WHERE employee_id = period_record.employee_id
              AND entry_date BETWEEN period_record.start_date AND period_record.end_date
              AND entry_type IN ('pto', 'sick', 'holiday')
              AND status IN ('draft', 'submitted', 'approved')
        ), 0),
        updated_at = NOW()
    WHERE id = period_record.id;

    -- Return appropriate record
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update period totals
DROP TRIGGER IF EXISTS trigger_update_period_totals ON time_entries;
CREATE TRIGGER trigger_update_period_totals
    AFTER INSERT OR UPDATE OR DELETE ON time_entries
    FOR EACH ROW
    EXECUTE PROCEDURE update_timesheet_period_totals();

-- Insert default overtime rule
INSERT INTO overtime_rules (name, daily_threshold, weekly_threshold, overtime_multiplier, double_time_multiplier)
VALUES ('Standard US Overtime', 8.0, 40.0, 1.5, 2.0)
ON CONFLICT DO NOTHING;

-- Comments for documentation
COMMENT ON TABLE time_entries IS 'Individual time entries for clock in/out and manual time tracking';
COMMENT ON TABLE timesheet_periods IS 'Timesheet periods (weekly/bi-weekly) for approval workflow';
COMMENT ON TABLE projects IS 'Projects for project-based time tracking';
COMMENT ON TABLE time_entry_projects IS 'Links time entries to projects with hour allocations';
COMMENT ON TABLE overtime_rules IS 'Overtime calculation rules';
COMMENT ON TABLE employee_overtime_rules IS 'Assigns overtime rules to employees';