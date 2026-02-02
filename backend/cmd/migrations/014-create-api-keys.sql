-- Hub-HRMS API Keys Table
-- Run this on your Hub-HRMS PostgreSQL database

-- Enable pgcrypto extension for password hashing
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- API Keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash TEXT NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_by_user_id UUID REFERENCES users(id),
    metadata JSONB DEFAULT '{}'::JSONB,
    
    CONSTRAINT api_keys_name_unique UNIQUE (name)
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_api_keys_expires ON api_keys(expires_at) WHERE expires_at IS NOT NULL;

-- API Key usage logs (simplified - not partitioned for easier compatibility)
CREATE TABLE IF NOT EXISTS api_key_usage_logs (
    id BIGSERIAL PRIMARY KEY,
    api_key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    endpoint VARCHAR(500),
    method VARCHAR(10),
    status_code INTEGER,
    request_ip INET,
    user_agent TEXT,
    request_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    response_time_ms INTEGER
);

-- Indexes for usage logs
CREATE INDEX IF NOT EXISTS idx_api_key_usage_api_key ON api_key_usage_logs(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_timestamp ON api_key_usage_logs(request_timestamp);

-- Function to verify API key
CREATE OR REPLACE FUNCTION verify_api_key(key_value TEXT)
RETURNS TABLE(
    api_key_id UUID,
    name VARCHAR(255),
    scopes TEXT[],
    is_valid BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ak.id,
        ak.name,
        ak.scopes,
        (ak.is_active AND (ak.expires_at IS NULL OR ak.expires_at > NOW()))::BOOLEAN as is_valid
    FROM api_keys ak
    WHERE ak.key_hash = crypt(key_value, ak.key_hash)
    LIMIT 1;
    
    -- Update last_used_at
    UPDATE api_keys
    SET last_used_at = NOW()
    WHERE id = (
        SELECT id FROM api_keys
        WHERE key_hash = crypt(key_value, key_hash)
        LIMIT 1
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to log API key usage
CREATE OR REPLACE FUNCTION log_api_key_usage(
    p_api_key_id UUID,
    p_endpoint VARCHAR(500),
    p_method VARCHAR(10),
    p_status_code INTEGER,
    p_request_ip INET,
    p_user_agent TEXT,
    p_response_time_ms INTEGER
) RETURNS VOID AS $$
BEGIN
    INSERT INTO api_key_usage_logs (
        api_key_id,
        endpoint,
        method,
        status_code,
        request_ip,
        user_agent,
        response_time_ms
    ) VALUES (
        p_api_key_id,
        p_endpoint,
        p_method,
        p_status_code,
        p_request_ip,
        p_user_agent,
        p_response_time_ms
    );
END;
$$ LANGUAGE plpgsql;

-- Create database user if needed (skip if it already exists or you use different auth)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'hub_hrms_api_user') THEN
        CREATE ROLE hub_hrms_api_user LOGIN PASSWORD 'change_this_password';
    END IF;
END
$$;

-- Grant necessary permissions
GRANT SELECT ON api_keys TO hub_hrms_api_user;
GRANT SELECT, INSERT ON api_key_usage_logs TO hub_hrms_api_user;
GRANT USAGE ON SEQUENCE api_key_usage_logs_id_seq TO hub_hrms_api_user;
GRANT EXECUTE ON FUNCTION verify_api_key(TEXT) TO hub_hrms_api_user;
GRANT EXECUTE ON FUNCTION log_api_key_usage(UUID, VARCHAR, VARCHAR, INTEGER, INET, TEXT, INTEGER) TO hub_hrms_api_user;

-- Create initial API key for HR Recruiting
-- Replace the key value with your generated key
INSERT INTO api_keys (
    key_hash,
    name,
    description,
    scopes,
    expires_at
) VALUES (
    crypt('(vgV5gby6$sq*GOIw&VAStz_MEOyMr]K@%bA/)YaY&k3FL;t{!bx[ZPQ@Dn-=:C^', gen_salt('bf')),
    'hr-recruiting-app',
    'Public recruiting portal API access',
    ARRAY['jobs:read', 'jobs:write', 'applications:read', 'applications:write', 'candidates:read'],
    NOW() + INTERVAL '1 year'
) ON CONFLICT (name) DO UPDATE
SET 
    key_hash = EXCLUDED.key_hash,
    scopes = EXCLUDED.scopes,
    expires_at = EXCLUDED.expires_at,
    updated_at = NOW();

-- View to monitor API key usage
CREATE OR REPLACE VIEW api_key_usage_summary AS
SELECT 
    ak.name,
    ak.is_active,
    ak.scopes,
    ak.created_at,
    ak.expires_at,
    ak.last_used_at,
    COUNT(aul.id) as total_requests,
    COUNT(CASE WHEN aul.status_code >= 200 AND aul.status_code < 300 THEN 1 END) as successful_requests,
    COUNT(CASE WHEN aul.status_code >= 400 THEN 1 END) as failed_requests,
    AVG(aul.response_time_ms)::INTEGER as avg_response_time_ms,
    MAX(aul.request_timestamp) as last_request_at
FROM api_keys ak
LEFT JOIN api_key_usage_logs aul ON ak.id = aul.api_key_id
    AND aul.request_timestamp > NOW() - INTERVAL '30 days'
GROUP BY ak.id, ak.name, ak.is_active, ak.scopes, ak.created_at, ak.expires_at, ak.last_used_at
ORDER BY ak.created_at DESC;

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_api_keys_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS api_keys_updated_at_trigger ON api_keys;
CREATE TRIGGER api_keys_updated_at_trigger
    BEFORE UPDATE ON api_keys
    FOR EACH ROW
    EXECUTE PROCEDURE update_api_keys_updated_at();

-- Example queries:

-- Check API key validity
-- SELECT * FROM verify_api_key('your-api-key-here');

-- View usage statistics
-- SELECT * FROM api_key_usage_summary;

-- Revoke an API key
-- UPDATE api_keys SET is_active = FALSE WHERE name = 'hr-recruiting-app';

-- Rotate API key (generate new one with same scopes)
/*
INSERT INTO api_keys (
    key_hash,
    name,
    description,
    scopes,
    expires_at
) VALUES (
    crypt('NEW_KEY_HERE', gen_salt('bf')),
    'hr-recruiting-app-v2',
    'Public recruiting portal API access (rotated)',
    ARRAY['jobs:read', 'jobs:write', 'applications:read', 'applications:write', 'candidates:read'],
    NOW() + INTERVAL '1 year'
);
*/