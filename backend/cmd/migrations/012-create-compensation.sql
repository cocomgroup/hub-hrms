-- Compensation Management Tables
-- Migration: Add compensation plans and bonuses tables

-- Compensation Plans Table
CREATE TABLE IF NOT EXISTS compensation_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    compensation_type VARCHAR(20) NOT NULL CHECK (compensation_type IN ('salary', 'hourly', 'contract')),
    base_amount DECIMAL(12,2) NOT NULL CHECK (base_amount >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    pay_frequency VARCHAR(20) NOT NULL CHECK (pay_frequency IN ('hourly', 'weekly', 'biweekly', 'bimonthly', 'monthly', 'annually')),
    effective_date DATE NOT NULL,
    end_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'pending', 'expired')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT check_end_date CHECK (end_date IS NULL OR end_date > effective_date)
);

-- Bonuses Table
CREATE TABLE IF NOT EXISTS bonuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    bonus_type VARCHAR(20) NOT NULL CHECK (bonus_type IN ('monthly', 'quarterly', 'annual', 'performance', 'signing', 'retention', 'referral', 'spot', 'holiday')),
    amount DECIMAL(12,2) NOT NULL CHECK (amount >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    description TEXT NOT NULL,
    payment_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'paid', 'cancelled')),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT chk_bonus_approval CHECK (
        (status IN ('approved', 'paid') AND approved_by IS NOT NULL AND approved_at IS NOT NULL) 
        OR status IN ('pending', 'cancelled')
    ),
    
    -- Ensure payment date is set when paid
    CONSTRAINT chk_bonus_payment CHECK (
        (status = 'paid' AND paid_at IS NOT NULL) 
        OR status IN ('pending', 'approved', 'cancelled')
    )
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_compensation_plans_employee ON compensation_plans(employee_id);
CREATE INDEX IF NOT EXISTS idx_compensation_plans_status ON compensation_plans(status);
CREATE INDEX IF NOT EXISTS idx_compensation_plans_effective_date ON compensation_plans(effective_date);

CREATE INDEX IF NOT EXISTS idx_bonuses_employee ON bonuses(employee_id);
CREATE INDEX IF NOT EXISTS idx_bonuses_status ON bonuses(status);
CREATE INDEX IF NOT EXISTS idx_bonuses_payment_date ON bonuses(payment_date);
CREATE INDEX IF NOT EXISTS idx_bonuses_bonus_type ON bonuses(bonus_type);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_compensation_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_compensation_plans_updated_at ON compensation_plans;
CREATE TRIGGER trigger_compensation_plans_updated_at
    BEFORE UPDATE ON compensation_plans
    FOR EACH ROW
    EXECUTE PROCEDURE update_compensation_updated_at();

DROP TRIGGER IF EXISTS trigger_bonuses_updated_at ON bonuses;
CREATE TRIGGER trigger_bonuses_updated_at
    BEFORE UPDATE ON bonuses
    FOR EACH ROW
    EXECUTE PROCEDURE update_compensation_updated_at();

-- Comments for documentation
COMMENT ON TABLE compensation_plans IS 'Employee compensation plans including salary, hourly, and contract arrangements';
COMMENT ON TABLE bonuses IS 'Employee bonuses including monthly, quarterly, annual, performance, signing, and retention bonuses';

COMMENT ON COLUMN compensation_plans.compensation_type IS 'Type of compensation: salary, hourly, or contract';
COMMENT ON COLUMN compensation_plans.base_amount IS 'Base compensation amount in the specified currency';
COMMENT ON COLUMN compensation_plans.pay_frequency IS 'How often the employee is paid: hourly, weekly, biweekly, monthly, or annually';
COMMENT ON COLUMN compensation_plans.status IS 'Current status of the plan: active, pending, or expired';

COMMENT ON COLUMN bonuses.bonus_type IS 'Type of bonus: monthly, quarterly, annual, performance, signing, or retention';
COMMENT ON COLUMN bonuses.status IS 'Workflow status: pending (awaiting approval), approved (ready to pay), paid (completed), cancelled';
COMMENT ON COLUMN bonuses.approved_by IS 'User who approved the bonus';
COMMENT ON COLUMN bonuses.paid_at IS 'When the bonus was actually paid';