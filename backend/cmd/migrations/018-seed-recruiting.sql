-- Seed Data: Recruiting (Job Postings) - FIXED
-- File: 018-seed-recruiting-fixed.sql
-- Description: Populate sample job postings and candidates

-- Engineering Jobs
-- Insert sample jobs for testing
INSERT INTO job_postings (
    id, title, department, location, employment_type, experience_level,
    salary_min, salary_max, salary_currency, description, requirements, responsibilities, benefits,
    skills, status, posted_date, remote_work, urgent_hiring, applications_count, view_count,
    created_by, created_at, updated_at
)
SELECT 
    '00000000-0000-0000-0000-000000000001'::uuid,
    'Senior Software Engineer',
    'Engineering',
    'San Francisco, CA',
    'FULL_TIME',
    'SENIOR',
    140000,
    200000,
    'USD',
    'We are seeking a talented Senior Software Engineer to join our growing team. You will work on cutting-edge technologies and help shape the future of our platform.',
    ARRAY['5+ years of experience in software development', 'Strong proficiency in Go and/or TypeScript', 'Experience with PostgreSQL and distributed systems', 'Excellent problem-solving and communication skills'],
    ARRAY['Design and implement scalable backend services', 'Collaborate with cross-functional teams', 'Mentor junior engineers', 'Participate in code reviews and architectural decisions'],
    ARRAY['Competitive salary and equity', 'Health, dental, and vision insurance', '401(k) matching', 'Flexible work arrangements', 'Professional development budget'],
    ARRAY['Go', 'TypeScript', 'PostgreSQL', 'Docker', 'Kubernetes', 'GraphQL'],
    'PUBLISHED',
    NOW() - INTERVAL '10 days',
    true,
    false,
    5,
    127,
    u.id,
    NOW() - INTERVAL '10 days',
    NOW()
FROM users u
WHERE NOT EXISTS (SELECT 1 FROM job_postings WHERE id = '00000000-0000-0000-0000-000000000001'::uuid)
LIMIT 1;

INSERT INTO job_postings (
    id, title, department, location, employment_type, experience_level,
    salary_min, salary_max, salary_currency, description, requirements, responsibilities, benefits,
    skills, status, posted_date, remote_work, urgent_hiring, applications_count, view_count,
    created_by, created_at, updated_at
)
SELECT 
    '00000000-0000-0000-0000-000000000002'::uuid,
    'Product Manager',
    'Product',
    'New York, NY',
    'FULL_TIME',
    'MID',
    120000,
    160000,
    'USD',
    'Join our product team to define and execute on our product strategy. You will work closely with engineering, design, and business stakeholders.',
    ARRAY['3+ years of product management experience', 'Strong analytical and communication skills', 'Experience with B2B SaaS products', 'Technical background preferred'],
    ARRAY['Define product roadmap and strategy', 'Gather and prioritize product requirements', 'Work with engineering to deliver features', 'Analyze metrics and user feedback'],
    ARRAY['Competitive compensation', 'Health benefits', 'Stock options', 'Remote work options', 'Learning and development opportunities'],
    ARRAY['Product Strategy', 'Agile', 'SQL', 'Analytics', 'Stakeholder Management'],
    'PUBLISHED',
    NOW() - INTERVAL '5 days',
    true,
    false,
    8,
    203,
    u.id,
    NOW() - INTERVAL '5 days',
    NOW()
FROM users u
WHERE NOT EXISTS (SELECT 1 FROM job_postings WHERE id = '00000000-0000-0000-0000-000000000002'::uuid)
LIMIT 1;

INSERT INTO job_postings (
    id, title, department, location, employment_type, experience_level,
    salary_min, salary_max, salary_currency, description, requirements, responsibilities, benefits,
    skills, status, posted_date, remote_work, urgent_hiring, applications_count, view_count,
    created_by, created_at, updated_at
)
SELECT 
    '00000000-0000-0000-0000-000000000003'::uuid,
    'UX Designer',
    'Design',
    'Remote',
    'FULL_TIME',
    'MID',
    100000,
    140000,
    'USD',
    'We are looking for a creative UX Designer to craft intuitive and delightful user experiences for our platform.',
    ARRAY['3+ years of UX design experience', 'Strong portfolio demonstrating UX process', 'Proficiency in Figma and design systems', 'Experience with user research and testing'],
    ARRAY['Create wireframes, prototypes, and high-fidelity designs', 'Conduct user research and usability testing', 'Collaborate with product and engineering teams', 'Maintain and evolve our design system'],
    ARRAY['Competitive salary', 'Full remote work', 'Health insurance', 'Home office stipend', 'Annual conference budget'],
    ARRAY['Figma', 'User Research', 'Prototyping', 'Design Systems', 'Interaction Design'],
    'PUBLISHED',
    NOW() - INTERVAL '3 days',
    true,
    false,
    12,
    156,
    u.id,
    NOW() - INTERVAL '3 days',
    NOW()
FROM users u
WHERE NOT EXISTS (SELECT 1 FROM job_postings WHERE id = '00000000-0000-0000-0000-000000000003'::uuid)
LIMIT 1;

INSERT INTO job_postings (
    id, title, department, location, employment_type, experience_level,
    salary_min, salary_max, salary_currency, description, requirements, responsibilities, benefits,
    skills, status, posted_date, remote_work, urgent_hiring, applications_count, view_count,
    created_by, created_at, updated_at
)
SELECT 
    '00000000-0000-0000-0000-000000000004'::uuid,
    'Data Scientist',
    'Engineering',
    'San Francisco, CA',
    'FULL_TIME',
    'SENIOR',
    150000,
    210000,
    'USD',
    'Join our data team to build ML models and analytics that power our product decisions and customer insights.',
    ARRAY['PhD or Masters in Computer Science, Statistics, or related field', '5+ years of experience in data science', 'Strong Python and SQL skills', 'Experience with ML frameworks (TensorFlow, PyTorch)', 'Track record of deploying ML models to production'],
    ARRAY['Build and deploy machine learning models', 'Analyze large datasets to derive insights', 'Collaborate with engineering to productionize models', 'Present findings to stakeholders'],
    ARRAY['Top-tier compensation and equity', 'Comprehensive health benefits', '401(k) with matching', 'Flexible schedule', 'Conference and education budget'],
    ARRAY['Python', 'SQL', 'Machine Learning', 'TensorFlow', 'PyTorch', 'Statistics'],
    'PUBLISHED',
    NOW() - INTERVAL '7 days',
    false,
    true,
    15,
    289,
    u.id,
    NOW() - INTERVAL '7 days',
    NOW()
FROM users u
WHERE NOT EXISTS (SELECT 1 FROM job_postings WHERE id = '00000000-0000-0000-0000-000000000004'::uuid)
LIMIT 1;

INSERT INTO job_postings (
    id, title, department, location, employment_type, experience_level,
    salary_min, salary_max, salary_currency, description, requirements, responsibilities, benefits,
    skills, status, posted_date, remote_work, urgent_hiring, applications_count, view_count,
    created_by, created_at, updated_at
)
SELECT 
    '00000000-0000-0000-0000-000000000005'::uuid,
    'DevOps Engineer',
    'Engineering',
    'Austin, TX',
    'FULL_TIME',
    'MID',
    110000,
    150000,
    'USD',
    'We need a DevOps Engineer to help us scale our infrastructure and improve our deployment processes.',
    ARRAY['3+ years of DevOps experience', 'Strong knowledge of AWS or GCP', 'Experience with Kubernetes and Docker', 'Proficiency in Infrastructure as Code (Terraform)', 'Experience with CI/CD pipelines'],
    ARRAY['Manage and scale cloud infrastructure', 'Build and maintain CI/CD pipelines', 'Implement monitoring and alerting', 'Improve deployment processes', 'Ensure security best practices'],
    ARRAY['Competitive salary', 'Health and wellness benefits', 'Stock options', 'Remote work flexibility', 'Professional development'],
    ARRAY['AWS', 'Kubernetes', 'Docker', 'Terraform', 'CI/CD', 'Monitoring'],
    'PUBLISHED',
    NOW() - INTERVAL '2 days',
    true,
    false,
    6,
    178,
    u.id,
    NOW() - INTERVAL '2 days',
    NOW()
FROM users u
WHERE NOT EXISTS (SELECT 1 FROM job_postings WHERE id = '00000000-0000-0000-0000-000000000005'::uuid)
LIMIT 1;

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
    'Full Stack Developer',
    'Engineering',
    'Austin, TX',
    'full-time',
    100000.00,
    140000.00,
    'Build full-stack applications using modern technologies.',
    ARRAY['3+ years full-stack experience', 'React/TypeScript', 'Node.js/Go backend', 'PostgreSQL/MongoDB'],
    ARRAY['Develop frontend and backend features', 'Write clean, maintainable code', 'Collaborate with designers', 'Review code'],
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
    'Marketing Manager',
    'Marketing',
    'Remote',
    'full-time',
    110000.00,
    150000.00,
    'Lead marketing strategy and campaign execution.',
    ARRAY['5+ years marketing experience', 'Digital marketing expertise', 'Analytics mindset', 'Team leadership'],
    ARRAY['Develop marketing strategy', 'Manage campaigns', 'Analyze performance', 'Lead team'],
    ARRAY['Remote work', 'Equity', 'Health insurance', 'Professional development'],
    'filled',
    CURRENT_DATE - INTERVAL '90 days',
    CURRENT_DATE - INTERVAL '45 days',
    18,
    id
FROM users WHERE role = 'admin' LIMIT 1
ON CONFLICT DO NOTHING;

-- Sample Candidates
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
FROM job_postings WHERE title = 'Senior Software Engineer' LIMIT 1
ON CONFLICT DO NOTHING;

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
FROM job_postings WHERE title = 'Senior Software Engineer' LIMIT 1
ON CONFLICT DO NOTHING;

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
FROM job_postings WHERE title = 'Product Manager' LIMIT 1
ON CONFLICT DO NOTHING;

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
    ARRAY['Figma', 'User Research', 'Design Systems', 'Prototyping']
FROM job_postings WHERE title = 'UX Designer' LIMIT 1
ON CONFLICT DO NOTHING;

-- Display summary
DO $$
DECLARE
    job_count INTEGER;
    candidate_count INTEGER;
    active_jobs INTEGER;
    published_jobs INTEGER;
BEGIN
    SELECT COUNT(*) INTO job_count FROM job_postings;
    SELECT COUNT(*) INTO candidate_count FROM candidates;
    SELECT COUNT(*) INTO active_jobs FROM job_postings WHERE status = 'active';
    SELECT COUNT(*) INTO published_jobs FROM job_postings WHERE status = 'PUBLISHED';
    
    RAISE NOTICE '================================';
    RAISE NOTICE 'Recruiting Seed Data Summary';
    RAISE NOTICE '================================';
    RAISE NOTICE 'Total job postings: %', job_count;
    RAISE NOTICE '  - Published (GraphQL): %', published_jobs;
    RAISE NOTICE '  - Active (REST): %', active_jobs;
    RAISE NOTICE '  - Closed/Filled: %', job_count - active_jobs - published_jobs;
    RAISE NOTICE 'Total candidates: %', candidate_count;
    RAISE NOTICE '================================';
END $$;