-- Migration: Create recruiting tables
-- Description: Tables for job postings, candidates, interviews, and recruiting workflow

-- Job Postings Table
CREATE TABLE IF NOT EXISTS job_postings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    department VARCHAR(100) NOT NULL,
    location VARCHAR(255) NOT NULL,
    employment_type VARCHAR(50) NOT NULL CHECK (employment_type IN ('full-time', 'part-time', 'contract', 'internship')),
    salary_min DECIMAL(12,2),
    salary_max DECIMAL(12,2),
    description TEXT NOT NULL,
    requirements TEXT[] DEFAULT '{}',
    responsibilities TEXT[] DEFAULT '{}',
    benefits TEXT[] DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'closed', 'filled')),
    posted_date TIMESTAMP,
    closed_date TIMESTAMP,
    applications_count INTEGER DEFAULT 0,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_job_postings_status ON job_postings(status);
CREATE INDEX IF NOT EXISTS idx_job_postings_department ON job_postings(department);
CREATE INDEX IF NOT EXISTS idx_job_postings_created_at ON job_postings(created_at DESC);

-- Candidates Table
CREATE TABLE IF NOT EXISTS candidates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_posting_id UUID NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    resume_url TEXT,
    cover_letter TEXT,
    linkedin_url TEXT,
    portfolio_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'screening', 'interview', 'offered', 'rejected', 'hired')),
    score INTEGER CHECK (score >= 0 AND score <= 100),
    ai_summary TEXT,
    strengths TEXT[] DEFAULT '{}',
    weaknesses TEXT[] DEFAULT '{}',
    experience_years INTEGER,
    skills TEXT[] DEFAULT '{}',
    applied_date TIMESTAMP NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_candidates_job_posting ON candidates(job_posting_id);
CREATE INDEX IF NOT EXISTS idx_candidates_status ON candidates(status);
CREATE INDEX IF NOT EXISTS idx_candidates_email ON candidates(email);
CREATE INDEX IF NOT EXISTS idx_candidates_applied_date ON candidates(applied_date DESC);
CREATE INDEX IF NOT EXISTS idx_candidates_score ON candidates(score DESC);

-- Interviews Table
CREATE TABLE IF NOT EXISTS interviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    interviewer_id UUID NOT NULL REFERENCES users(id),
    scheduled_at TIMESTAMP NOT NULL,
    duration INTEGER NOT NULL DEFAULT 60, -- minutes
    interview_type VARCHAR(20) NOT NULL CHECK (interview_type IN ('phone', 'video', 'onsite', 'technical', 'behavioral')),
    location TEXT,
    meeting_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'completed', 'cancelled', 'no_show', 'rescheduled')),
    feedback TEXT,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_interviews_candidate ON interviews(candidate_id);
CREATE INDEX IF NOT EXISTS idx_interviews_interviewer ON interviews(interviewer_id);
CREATE INDEX IF NOT EXISTS idx_interviews_scheduled_at ON interviews(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_interviews_status ON interviews(status);

-- Job Board Postings Table
CREATE TABLE IF NOT EXISTS job_board_postings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_posting_id UUID NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    board_name VARCHAR(50) NOT NULL, -- linkedin, indeed, glassdoor, ziprecruiter, monster
    external_id VARCHAR(255), -- ID on the external platform
    posted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired', 'removed')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_job_board_postings_job ON job_board_postings(job_posting_id);
CREATE INDEX IF NOT EXISTS idx_job_board_postings_board ON job_board_postings(board_name);
CREATE INDEX IF NOT EXISTS idx_job_board_postings_status ON job_board_postings(status);

-- Candidate Emails Table
CREATE TABLE IF NOT EXISTS candidate_emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    sent_by UUID NOT NULL REFERENCES users(id),
    subject VARCHAR(500) NOT NULL,
    body TEXT NOT NULL,
    email_type VARCHAR(50) NOT NULL DEFAULT 'custom', -- screening, interview, offer, rejection, custom
    sent_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_candidate_emails_candidate ON candidate_emails(candidate_id);
CREATE INDEX IF NOT EXISTS idx_candidate_emails_sent_by ON candidate_emails(sent_by);
CREATE INDEX IF NOT EXISTS idx_candidate_emails_sent_at ON candidate_emails(sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_candidate_emails_type ON candidate_emails(email_type);

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_recruiting_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop and recreate trigger for job_postings
DROP TRIGGER IF EXISTS trigger_job_postings_updated_at ON job_postings;
CREATE TRIGGER trigger_job_postings_updated_at
    BEFORE UPDATE ON job_postings
    FOR EACH ROW
    EXECUTE PROCEDURE update_recruiting_updated_at();

-- Drop and recreate trigger for candidates
DROP TRIGGER IF EXISTS trigger_candidates_updated_at ON candidates;
CREATE TRIGGER trigger_candidates_updated_at
    BEFORE UPDATE ON candidates
    FOR EACH ROW
    EXECUTE PROCEDURE update_recruiting_updated_at();

-- Drop and recreate trigger for interviews
DROP TRIGGER IF EXISTS trigger_interviews_updated_at ON interviews;
CREATE TRIGGER trigger_interviews_updated_at
    BEFORE UPDATE ON interviews
    FOR EACH ROW
    EXECUTE PROCEDURE update_recruiting_updated_at();

-- Drop and recreate trigger for job_board_postings
DROP TRIGGER IF EXISTS trigger_job_board_postings_updated_at ON job_board_postings;
CREATE TRIGGER trigger_job_board_postings_updated_at
    BEFORE UPDATE ON job_board_postings
    FOR EACH ROW
    EXECUTE PROCEDURE update_recruiting_updated_at();

-- FIXED: Drop and recreate views
DROP VIEW IF EXISTS recruiting_pipeline_summary CASCADE;
CREATE VIEW recruiting_pipeline_summary AS
SELECT 
    jp.id as job_id,
    jp.title as job_title,
    jp.department,
    jp.status as job_status,
    COUNT(c.id) as total_candidates,
    COUNT(CASE WHEN c.status = 'new' THEN 1 END) as new_candidates,
    COUNT(CASE WHEN c.status = 'screening' THEN 1 END) as screening_candidates,
    COUNT(CASE WHEN c.status = 'interview' THEN 1 END) as interview_candidates,
    COUNT(CASE WHEN c.status = 'offered' THEN 1 END) as offered_candidates,
    COUNT(CASE WHEN c.status = 'hired' THEN 1 END) as hired_candidates,
    AVG(c.score) as avg_candidate_score
FROM job_postings jp
LEFT JOIN candidates c ON jp.id = c.job_posting_id
GROUP BY jp.id, jp.title, jp.department, jp.status;

DROP VIEW IF EXISTS candidate_interview_history CASCADE;
CREATE VIEW candidate_interview_history AS
SELECT 
    c.id as candidate_id,
    c.first_name,
    c.last_name,
    c.email,
    c.status as candidate_status,
    jp.title as job_title,
    i.id as interview_id,
    i.scheduled_at,
    i.interview_type,
    i.status as interview_status,
    i.rating,
    u.email as interviewer_email
FROM candidates c
JOIN job_postings jp ON c.job_posting_id = jp.id
LEFT JOIN interviews i ON c.id = i.candidate_id
LEFT JOIN users u ON i.interviewer_id = u.id
ORDER BY c.applied_date DESC, i.scheduled_at DESC;

-- Grant permissions (commented out - uncomment if using dedicated app role)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON job_postings TO hrms_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON candidates TO hrms_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON interviews TO hrms_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON job_board_postings TO hrms_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON candidate_emails TO hrms_user;
-- GRANT SELECT ON recruiting_pipeline_summary TO hrms_user;
-- GRANT SELECT ON candidate_interview_history TO hrms_user;

-- Add comments
COMMENT ON TABLE job_postings IS 'Job openings and their details';
COMMENT ON TABLE candidates IS 'Job applicants and their information';
COMMENT ON TABLE interviews IS 'Scheduled interviews for candidates';
COMMENT ON TABLE job_board_postings IS 'Tracking of job postings on external job boards';
COMMENT ON TABLE candidate_emails IS 'Communication history with candidates';
COMMENT ON VIEW recruiting_pipeline_summary IS 'Summary of recruiting pipeline by job';
COMMENT ON VIEW candidate_interview_history IS 'Complete interview history for all candidates';