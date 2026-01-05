-- ================================================
-- BANK INFORMATION TABLE
-- For storing employee/vendor banking details
-- ================================================

CREATE TABLE IF NOT EXISTS bank_information (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- Account Holder Information
    account_holder_name VARCHAR(255) NOT NULL,
    
    -- Bank Details
    bank_name VARCHAR(255) NOT NULL,
    account_type VARCHAR(20) NOT NULL CHECK (account_type IN ('checking', 'savings')),
    
    -- Encrypted sensitive data (use encryption at application level)
    account_number_encrypted TEXT NOT NULL,
    routing_number_encrypted TEXT NOT NULL,
    
    -- Last 4 digits for display (unencrypted for convenience)
    account_number_last4 VARCHAR(4),
    
    -- International transfers
    swift_code VARCHAR(11),
    iban VARCHAR(34),
    
    -- Bank Address
    bank_address TEXT,
    bank_city VARCHAR(100),
    bank_state VARCHAR(2),
    bank_zip VARCHAR(10),
    bank_country VARCHAR(2) DEFAULT 'US',
    
    -- Status and Verification
    is_primary BOOLEAN DEFAULT true,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'inactive', 'suspended')),
    verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP,
    verified_by UUID REFERENCES users(id),
    
    -- Audit fields
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- Ensure only one primary account per employee (partial unique index)
CREATE UNIQUE INDEX IF NOT EXISTS unique_primary_per_employee 
ON bank_information(employee_id) 
WHERE is_primary = true;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_bank_info_employee_id 
ON bank_information(employee_id);

CREATE INDEX IF NOT EXISTS idx_bank_info_status 
ON bank_information(status);

CREATE INDEX IF NOT EXISTS idx_bank_info_verified 
ON bank_information(verified);

-- Comments for documentation
COMMENT ON TABLE bank_information IS 
'Employee and vendor banking information for direct deposit payments';

COMMENT ON COLUMN bank_information.account_number_encrypted IS 
'Encrypted account number - must be encrypted/decrypted at application level';

COMMENT ON COLUMN bank_information.routing_number_encrypted IS 
'Encrypted routing number - must be encrypted/decrypted at application level';

COMMENT ON COLUMN bank_information.account_number_last4 IS 
'Last 4 digits of account number for display purposes (unencrypted)';

COMMENT ON COLUMN bank_information.status IS 
'Status: pending (awaiting verification), active (verified and in use), inactive (not in use), suspended (temporarily disabled)';

-- Trigger for updated_at
DROP TRIGGER IF EXISTS update_bank_information_updated_at ON bank_information;
CREATE TRIGGER update_bank_information_updated_at
    BEFORE UPDATE ON bank_information
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();

-- ================================================
-- BANK VERIFICATION LOG
-- Track verification attempts and history
-- ================================================

CREATE TABLE IF NOT EXISTS bank_verification_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_info_id UUID NOT NULL REFERENCES bank_information(id) ON DELETE CASCADE,
    verification_method VARCHAR(50) NOT NULL CHECK (verification_method IN ('micro-deposit', 'instant', 'manual', 'third-party')),
    verification_status VARCHAR(20) NOT NULL CHECK (verification_status IN ('initiated', 'pending', 'success', 'failed', 'expired')),
    verification_code VARCHAR(100),
    attempt_count INTEGER DEFAULT 1,
    verified_at TIMESTAMP,
    verified_by UUID REFERENCES users(id),
    failure_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_bank_verification_bank_info_id 
ON bank_verification_log(bank_info_id);

CREATE INDEX IF NOT EXISTS idx_bank_verification_status 
ON bank_verification_log(verification_status);

-- Comments
COMMENT ON TABLE bank_verification_log IS 
'Log of bank account verification attempts and results';

COMMENT ON COLUMN bank_verification_log.verification_method IS 
'Method: micro-deposit (small test deposits), instant (3rd party API), manual (HR verification), third-party (service like Plaid)';

-- ================================================
-- PAYMENT HISTORY (Optional - for tracking)
-- ================================================

CREATE TABLE IF NOT EXISTS payment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    bank_info_id UUID NOT NULL REFERENCES bank_information(id) ON DELETE RESTRICT,
    payment_type VARCHAR(50) NOT NULL CHECK (payment_type IN ('payroll', 'bonus', 'reimbursement', 'commission', 'advance')),
    amount DECIMAL(12, 2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) DEFAULT 'USD',
    payment_date DATE NOT NULL,
    payment_status VARCHAR(20) DEFAULT 'pending' CHECK (payment_status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    transaction_id VARCHAR(255),
    failure_reason TEXT,
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_payment_history_employee_id 
ON payment_history(employee_id);

CREATE INDEX IF NOT EXISTS idx_payment_history_bank_info_id 
ON payment_history(bank_info_id);

CREATE INDEX IF NOT EXISTS idx_payment_history_status 
ON payment_history(payment_status);

CREATE INDEX IF NOT EXISTS idx_payment_history_date 
ON payment_history(payment_date);

-- Comments
COMMENT ON TABLE payment_history IS 
'History of payments made to employee bank accounts';

-- ================================================
-- SAMPLE DATA (Optional - for testing)
-- ================================================

-- Note: In production, account numbers and routing numbers should be encrypted
-- This is just sample structure

/*
INSERT INTO bank_information (
    employee_id,
    account_holder_name,
    bank_name,
    account_type,
    account_number_encrypted,
    routing_number_encrypted,
    account_number_last4,
    bank_country,
    status,
    verified
)
SELECT 
    id,
    first_name || ' ' || last_name,
    'Sample Bank',
    'checking',
    'ENCRYPTED_ACCOUNT_NUMBER_HERE',
    'ENCRYPTED_ROUTING_NUMBER_HERE',
    '1234',
    'US',
    'pending',
    false
FROM employees
WHERE employment_type = 'employee'
LIMIT 3;
*/

-- ================================================
-- VERIFICATION QUERIES
-- ================================================

-- Check bank information
SELECT 
    e.first_name,
    e.last_name,
    bi.bank_name,
    bi.account_type,
    bi.account_number_last4,
    bi.status,
    bi.verified
FROM bank_information bi
JOIN employees e ON bi.employee_id = e.id
ORDER BY bi.created_at DESC;

-- Check primary accounts per employee
SELECT 
    employee_id,
    COUNT(*) as total_accounts,
    COUNT(*) FILTER (WHERE is_primary = true) as primary_accounts
FROM bank_information
GROUP BY employee_id
HAVING COUNT(*) FILTER (WHERE is_primary = true) > 1; -- Should return no rows

-- Check verification status
SELECT 
    verified,
    status,
    COUNT(*) as count
FROM bank_information
GROUP BY verified, status;