-- Timesheet System Database Schema
-- Migration: Add timesheet tables

-- Create simplified timesheets table first (weekly submissions)
CREATE TABLE IF NOT EXISTS timesheets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    start_date DATE NOT NULL, -- Monday of the week
    end_date DATE NOT NULL, -- Sunday of the week
    status VARCHAR(20) NOT NULL DEFAULT 'draft', -- draft, submitted, approved, rejected
    total_hours DECIMAL(6,2) NOT NULL DEFAULT 0,
    regular_hours DECIMAL(6,2) NOT NULL DEFAULT 0,
    pto_hours DECIMAL(6,2) NOT NULL DEFAULT 0,
    holiday_hours DECIMAL(6,2) NOT NULL DEFAULT 0,
    submitted_at TIMESTAMP WITH TIME ZONE,
    approved_at TIMESTAMP WITH TIME ZONE,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_employee_week UNIQUE(employee_id, start_date, end_date),
    CONSTRAINT check_week_dates CHECK (end_date = start_date + INTERVAL '6 days')
);

-- Create time_entries table (references timesheets and projects)
CREATE TABLE IF NOT EXISTS time_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    entry_date DATE NOT NULL,
    hours DECIMAL(5,2) NOT NULL CHECK (hours >= 0 AND hours <= 24),
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    entry_type VARCHAR(20) NOT NULL DEFAULT 'regular', -- regular, pto, holiday
    notes TEXT,
    timesheet_id UUID REFERENCES timesheets(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_employee_date UNIQUE(employee_id, entry_date)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_time_entries_employee_date ON time_entries(employee_id, entry_date);
CREATE INDEX IF NOT EXISTS idx_time_entries_project ON time_entries(project_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_timesheet ON time_entries(timesheet_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_type ON time_entries(entry_type);

CREATE INDEX IF NOT EXISTS idx_timesheets_employee ON timesheets(employee_id);
CREATE INDEX IF NOT EXISTS idx_timesheets_dates ON timesheets(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_timesheets_status ON timesheets(status);
CREATE INDEX IF NOT EXISTS idx_timesheets_approved_by ON timesheets(approved_by);

-- Function to update timestamps
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for timestamp updates
DROP TRIGGER IF EXISTS trigger_time_entries_timestamp ON time_entries;
CREATE TRIGGER trigger_time_entries_timestamp
    BEFORE UPDATE ON time_entries
    FOR EACH ROW
    EXECUTE PROCEDURE update_timestamp();

DROP TRIGGER IF EXISTS trigger_timesheets_timestamp ON timesheets;
CREATE TRIGGER trigger_timesheets_timestamp
    BEFORE UPDATE ON timesheets
    FOR EACH ROW
    EXECUTE PROCEDURE update_timestamp();

-- Drop old timesheet_periods table (no longer needed)
DROP TABLE IF EXISTS timesheet_periods CASCADE;

-- Drop time_entry_projects table (simplified - project is on time_entry now)
DROP TABLE IF EXISTS time_entry_projects CASCADE;

-- Comments for documentation
COMMENT ON TABLE time_entries IS 'Daily time entries with hours worked per day';
COMMENT ON TABLE timesheets IS 'Weekly timesheet submissions for manager approval';

COMMENT ON COLUMN time_entries.hours IS 'Hours worked on this day (0-24)';
COMMENT ON COLUMN time_entries.entry_type IS 'Type: regular, pto, or holiday';
COMMENT ON COLUMN time_entries.timesheet_id IS 'Link to weekly timesheet when submitted';

COMMENT ON COLUMN timesheets.start_date IS 'Monday of the week';
COMMENT ON COLUMN timesheets.end_date IS 'Sunday of the week (start_date + 6 days)';
COMMENT ON COLUMN timesheets.status IS 'Status: draft, submitted, approved, rejected';