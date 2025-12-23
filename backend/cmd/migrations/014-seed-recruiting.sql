-- Seed Data: Recruiting (Job Postings)
-- File: 015-seed-recruiting.sql
-- Description: Populate sample job postings and candidates

-- Note: 
-- - job_postings.status values: 'draft', 'active', 'closed', 'filled' (NOT 'open')
-- - Table is 'candidates' not 'applications'

-- Engineering Jobs
INSERT INTO job_postings (title, department, location, employment_type, salary_min, salary_max, description, requirements, responsibilities, benefits, status, posted_date, applications_count, created_by)
SELECT 
    'Senior Software Engineer',
    'Engineering',
    'San Francisco, CA (Hybrid)',
    'full-time',
    120000.00,
    160000.00,
    'We are seeking an experienced Senior Software Engineer to join our growing team.',
    ARRAY['5+ years of software development experience', 'Proficiency in Go and TypeScript', 'Experience with cloud platforms (AWS/Azure)', 'Strong problem-solving skills'],
    ARRAY['Design and implement scalable backend services', 'Mentor junior engineers', 'Collaborate with product team', 'Code reviews and documentation'],
    ARRAY['Health insurance', '401(k) matching', 'Flexible work hours', 'Remote work options'],
    'active',
    CURRENT_DATE - INTERVAL '10 days',
    5,
    id
FROM users WHERE role = 'admin' LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO job_postings (title, department, location, employment_type, salary_min, salary_max, description, requirements, responsibilities, benefits, status, posted_date, applications_count, created_by)
SELECT 
    'Frontend Developer',
    'Engineering',
    'Remote',
    'full-time',
    90000.00,
    130000.00,
    'Join our frontend team to build beautiful, responsive user interfaces.',
    ARRAY['3+ years of frontend development', 'Expert in React/Vue/Svelte', 'Strong CSS skills', 'UI/UX design sense'],
    ARRAY['Build responsive web applications', 'Implement modern UI designs', 'Optimize performance', 'Work with design team'],
    ARRAY['Remote work', 'Flexible hours', 'Learning budget', 'Health insurance'],
    'active',
    CURRENT_DATE - INTERVAL '5 days',
    3,
    id
FROM users WHERE role = 'admin' LIMIT 1
ON CONFLICT DO NOTHING;

-- Sales Jobs
INSERT INTO job_postings (title, department, location, employment_type, salary_min, salary_max, description, requirements, responsibilities, benefits, status, posted_date, applications_count, created_by)
SELECT 
    'Account Executive',
    'Sales',
    'Chicago, IL',
    'full-time',
    70000.00,
    110000.00,
    'Seeking a motivated Account Executive to drive business growth.',
    ARRAY['2+ years of B2B sales experience', 'Strong communication skills', 'Track record of meeting quotas', 'CRM experience (Salesforce)'],
    ARRAY['Manage sales pipeline', 'Client relationship management', 'Conduct product demos', 'Negotiate contracts'],
    ARRAY['Commission structure', 'Health insurance', 'Car allowance', 'Professional development'],
    'active',
    CURRENT_DATE - INTERVAL '15 days',
    8,
    id
FROM users WHERE role = 'admin' LIMIT 1
ON CONFLICT DO NOTHING;

-- Marketing Jobs
INSERT INTO job_postings (title, department, location, employment_type, salary_min, salary_max, description, requirements, responsibilities, benefits, status, posted_date, applications_count, created_by)
SELECT 
    'Content Marketing Manager',
    'Marketing',
    'Los Angeles, CA',
    'full-time',
    75000.00,
    105000.00,
    'Lead our content marketing strategy and team.',
    ARRAY['4+ years in content marketing', 'SEO/SEM expertise', 'Strong writing skills', 'Analytics experience'],
    ARRAY['Develop content strategy', 'Manage content calendar', 'Lead content team', 'Analyze campaign performance'],
    ARRAY['Creative environment', 'Flexible schedule', 'Health benefits', '401(k) matching'],
    'active',
    CURRENT_DATE - INTERVAL '7 days',
    4,
    id
FROM users WHERE role = 'admin' LIMIT 1
ON CONFLICT DO NOTHING;

-- HR Jobs
INSERT INTO job_postings (title, department, location, employment_type, salary_min, salary_max, description, requirements, responsibilities, benefits, status, posted_date, applications_count, created_by)
SELECT 
    'HR Generalist',
    'Human Resources',
    'New York, NY',
    'full-time',
    60000.00,
    80000.00,
    'Support HR operations and employee relations.',
    ARRAY['2+ years HR experience', 'Knowledge of employment law', 'HRIS experience', 'Strong interpersonal skills'],
    ARRAY['Employee onboarding', 'Benefits administration', 'Policy compliance', 'Employee relations'],
    ARRAY['Comprehensive benefits', 'Professional development', 'Work-life balance', 'Collaborative environment'],
    'active',
    CURRENT_DATE - INTERVAL '3 days',
    6,
    id
FROM users WHERE role = 'admin' LIMIT 1
ON CONFLICT DO NOTHING;

-- IT Jobs
INSERT INTO job_postings (title, department, location, employment_type, salary_min, salary_max, description, requirements, responsibilities, benefits, status, posted_date, applications_count, created_by)
SELECT 
    'System Administrator',
    'IT',
    'Seattle, WA',
    'full-time',
    65000.00,
    90000.00,
    'Manage and maintain our IT infrastructure and systems.',
    ARRAY['3+ years system administration', 'Linux/Windows expertise', 'Network configuration', 'Security best practices'],
    ARRAY['Maintain servers and networks', 'User support', 'Security monitoring', 'System updates'],
    ARRAY['Health insurance', 'Learning budget', 'Flexible schedule', 'Remote work'],
    'active',
    CURRENT_DATE - INTERVAL '12 days',
    7,
    id
FROM users WHERE role = 'admin' LIMIT 1
ON CONFLICT DO NOTHING;

-- Closed/Filled Jobs
INSERT INTO job_postings (title, department, location, employment_type, salary_min, salary_max, description, requirements, responsibilities, benefits, status, posted_date, closed_date, applications_count, created_by)
SELECT 
    'DevOps Engineer',
    'Engineering',
    'Austin, TX',
    'full-time',
    100000.00,
    140000.00,
    'Manage infrastructure and deployment pipelines.',
    ARRAY['3+ years DevOps experience', 'Kubernetes/Docker', 'CI/CD pipelines', 'AWS/Azure'],
    ARRAY['Manage cloud infrastructure', 'Automate deployments', 'Monitor systems', 'Improve reliability'],
    ARRAY['Remote work', 'Stock options', 'Health insurance', 'Unlimited PTO'],
    'closed',
    CURRENT_DATE - INTERVAL '60 days',
    CURRENT_DATE - INTERVAL '30 days',
    12,
    id
FROM users WHERE role = 'admin' LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO job_postings (title, department, location, employment_type, salary_min, salary_max, description, requirements, responsibilities, benefits, status, posted_date, closed_date, applications_count, created_by)
SELECT 
    'Product Manager',
    'Product',
    'Remote',
    'full-time',
    110000.00,
    150000.00,
    'Lead product strategy and roadmap development.',
    ARRAY['5+ years product management', 'Technical background', 'Agile/Scrum experience', 'Analytical mindset'],
    ARRAY['Define product roadmap', 'Work with engineering', 'Conduct user research', 'Track metrics'],
    ARRAY['Remote work', 'Equity', 'Health insurance', 'Professional development'],
    'filled',
    CURRENT_DATE - INTERVAL '90 days',
    CURRENT_DATE - INTERVAL '45 days',
    18,
    id
FROM users WHERE role = 'admin' LIMIT 1
ON CONFLICT DO NOTHING;

-- Sample Candidates (using 'candidates' table, not 'applications')
INSERT INTO candidates (job_posting_id, first_name, last_name, email, phone, resume_url, cover_letter, status, applied_date)
SELECT 
    id,
    'Alice',
    'Johnson',
    'alice.johnson@email.com',
    '555-1001',
    'https://storage.example.com/resumes/alice-johnson.pdf',
    'I am excited to apply for the Senior Software Engineer position...',
    'screening',
    CURRENT_DATE - INTERVAL '8 days'
FROM job_postings WHERE title = 'Senior Software Engineer' LIMIT 1;

INSERT INTO candidates (job_posting_id, first_name, last_name, email, phone, resume_url, cover_letter, status, applied_date)
SELECT 
    id,
    'Bob',
    'Smith',
    'bob.smith@email.com',
    '555-1002',
    'https://storage.example.com/resumes/bob-smith.pdf',
    'With 6 years of experience in Go development...',
    'interview',
    CURRENT_DATE - INTERVAL '7 days'
FROM job_postings WHERE title = 'Senior Software Engineer' LIMIT 1;

INSERT INTO candidates (job_posting_id, first_name, last_name, email, phone, resume_url, status, applied_date)
SELECT 
    id,
    'Carol',
    'Williams',
    'carol.williams@email.com',
    '555-1003',
    'https://storage.example.com/resumes/carol-williams.pdf',
    'new',
    CURRENT_DATE - INTERVAL '2 days'
FROM job_postings WHERE title = 'Frontend Developer' LIMIT 1;

INSERT INTO candidates (job_posting_id, first_name, last_name, email, phone, resume_url, status, applied_date)
SELECT 
    id,
    'David',
    'Brown',
    'david.brown@email.com',
    '555-1004',
    'https://storage.example.com/resumes/david-brown.pdf',
    'offered',
    CURRENT_DATE - INTERVAL '25 days'
FROM job_postings WHERE title = 'Account Executive' LIMIT 1;

INSERT INTO candidates (job_posting_id, first_name, last_name, email, phone, resume_url, status, applied_date, score, experience_years, skills)
SELECT 
    id,
    'Emma',
    'Davis',
    'emma.davis@email.com',
    '555-1005',
    'https://storage.example.com/resumes/emma-davis.pdf',
    'screening',
    CURRENT_DATE - INTERVAL '5 days',
    85,
    7,
    ARRAY['React', 'TypeScript', 'CSS', 'Node.js']
FROM job_postings WHERE title = 'Frontend Developer' LIMIT 1;

INSERT INTO candidates (job_posting_id, first_name, last_name, email, phone, resume_url, status, applied_date, score, experience_years, skills)
SELECT 
    id,
    'Frank',
    'Miller',
    'frank.miller@email.com',
    '555-1006',
    'https://storage.example.com/resumes/frank-miller.pdf',
    'interview',
    CURRENT_DATE - INTERVAL '10 days',
    90,
    4,
    ARRAY['B2B Sales', 'Salesforce', 'Negotiation', 'Account Management']
FROM job_postings WHERE title = 'Account Executive' LIMIT 1;

-- Display summary
DO $$
DECLARE
    job_count INTEGER;
    candidate_count INTEGER;
    active_jobs INTEGER;
BEGIN
    SELECT COUNT(*) INTO job_count FROM job_postings;
    SELECT COUNT(*) INTO candidate_count FROM candidates;
    SELECT COUNT(*) INTO active_jobs FROM job_postings WHERE status = 'active';
    
    RAISE NOTICE '================================';
    RAISE NOTICE 'Recruiting Seed Data Summary';
    RAISE NOTICE '================================';
    RAISE NOTICE 'Total job postings: %', job_count;
    RAISE NOTICE '  - Active: %', active_jobs;
    RAISE NOTICE '  - Closed/Filled: %', job_count - active_jobs;
    RAISE NOTICE 'Total candidates: %', candidate_count;
    RAISE NOTICE '================================';
    RAISE NOTICE 'Job Categories:';
    RAISE NOTICE '  - Engineering (3)';
    RAISE NOTICE '  - Sales (1)';
    RAISE NOTICE '  - Marketing (1)';
    RAISE NOTICE '  - HR (1)';
    RAISE NOTICE '  - IT (1)';
    RAISE NOTICE '  - Product (1)';
    RAISE NOTICE '================================';
END $$;