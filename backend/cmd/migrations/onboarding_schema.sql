-- ============================================================================
-- New Hire Onboarding Workflow - Database Schema
-- ============================================================================

-- STEP 1: Create onboarding_workflows table
-- ============================================================================
CREATE TABLE IF NOT EXISTS onboarding_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    expected_completion_date DATE,
    actual_completion_date DATE,
    status VARCHAR(50) NOT NULL DEFAULT 'not_started', -- not_started, in_progress, completed, overdue
    overall_progress INTEGER DEFAULT 0, -- 0-100 percentage
    assigned_buddy_id UUID REFERENCES employees(id),
    assigned_manager_id UUID REFERENCES employees(id),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by UUID REFERENCES employees(id),
    
    CONSTRAINT valid_progress CHECK (overall_progress >= 0 AND overall_progress <= 100),
    CONSTRAINT valid_status CHECK (status IN ('not_started', 'in_progress', 'completed', 'overdue'))
);

CREATE INDEX idx_onboarding_workflows_employee ON onboarding_workflows(employee_id);
CREATE INDEX idx_onboarding_workflows_status ON onboarding_workflows(status);
CREATE INDEX idx_onboarding_workflows_start_date ON onboarding_workflows(start_date);


-- STEP 2: Create onboarding_tasks table
-- ============================================================================
CREATE TABLE IF NOT EXISTS onboarding_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100), -- documentation, equipment, training, access, administrative, social
    priority VARCHAR(50) DEFAULT 'medium', -- low, medium, high, critical
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, in_progress, completed, blocked, skipped
    assigned_to UUID REFERENCES employees(id), -- Who needs to complete this
    due_date DATE,
    completed_at TIMESTAMP,
    completed_by UUID REFERENCES employees(id),
    order_index INTEGER DEFAULT 0, -- For sorting/ordering tasks
    is_mandatory BOOLEAN DEFAULT true,
    estimated_hours DECIMAL(5,2),
    actual_hours DECIMAL(5,2),
    dependencies JSONB, -- Array of task IDs that must be completed first
    attachments JSONB, -- Array of file metadata
    ai_generated BOOLEAN DEFAULT false, -- Track if task was created by AI
    ai_suggestions TEXT, -- AI recommendations for this task
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'in_progress', 'completed', 'blocked', 'skipped'))
);

CREATE INDEX idx_onboarding_tasks_workflow ON onboarding_tasks(workflow_id);
CREATE INDEX idx_onboarding_tasks_status ON onboarding_tasks(status);
CREATE INDEX idx_onboarding_tasks_assigned ON onboarding_tasks(assigned_to);
CREATE INDEX idx_onboarding_tasks_due_date ON onboarding_tasks(due_date);


-- STEP 3: Create onboarding_checklist_items table (predefined templates)
-- ============================================================================
CREATE TABLE IF NOT EXISTS onboarding_checklist_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    department VARCHAR(100), -- Specific to department or NULL for all
    role_type VARCHAR(100), -- Specific to role or NULL for all
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS onboarding_checklist_template_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES onboarding_checklist_templates(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    priority VARCHAR(50) DEFAULT 'medium',
    day_offset INTEGER DEFAULT 0, -- Days from start date
    is_mandatory BOOLEAN DEFAULT true,
    estimated_hours DECIMAL(5,2),
    order_index INTEGER DEFAULT 0,
    assigned_to_role VARCHAR(100), -- 'new_hire', 'manager', 'hr', 'it', 'buddy'
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_template_items_template ON onboarding_checklist_template_items(template_id);


-- STEP 4: Create onboarding_interactions table (AI agent interactions)
-- ============================================================================
CREATE TABLE IF NOT EXISTS onboarding_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id),
    interaction_type VARCHAR(50) NOT NULL, -- chat, reminder, suggestion, check_in, escalation
    message TEXT NOT NULL,
    ai_response TEXT,
    sentiment VARCHAR(50), -- positive, neutral, negative, concerned
    requires_action BOOLEAN DEFAULT false,
    action_taken BOOLEAN DEFAULT false,
    metadata JSONB, -- Additional context (task_ids, links, etc)
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT valid_interaction_type CHECK (interaction_type IN 
        ('chat', 'reminder', 'suggestion', 'check_in', 'escalation')),
    CONSTRAINT valid_sentiment CHECK (sentiment IN 
        ('positive', 'neutral', 'negative', 'concerned') OR sentiment IS NULL)
);

CREATE INDEX idx_onboarding_interactions_workflow ON onboarding_interactions(workflow_id);
CREATE INDEX idx_onboarding_interactions_employee ON onboarding_interactions(employee_id);
CREATE INDEX idx_onboarding_interactions_created ON onboarding_interactions(created_at);


-- STEP 5: Create onboarding_milestones table
-- ============================================================================
CREATE TABLE IF NOT EXISTS onboarding_milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    target_date DATE,
    completed_date DATE,
    status VARCHAR(50) DEFAULT 'pending', -- pending, completed, missed
    celebration_sent BOOLEAN DEFAULT false, -- Track if we sent congratulations
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT valid_milestone_status CHECK (status IN ('pending', 'completed', 'missed'))
);

CREATE INDEX idx_onboarding_milestones_workflow ON onboarding_milestones(workflow_id);


-- STEP 6: Create onboarding_metrics table (for tracking and analytics)
-- ============================================================================
CREATE TABLE IF NOT EXISTS onboarding_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    metric_date DATE NOT NULL,
    tasks_completed INTEGER DEFAULT 0,
    tasks_pending INTEGER DEFAULT 0,
    tasks_overdue INTEGER DEFAULT 0,
    engagement_score INTEGER, -- 0-100 based on activity
    ai_interactions INTEGER DEFAULT 0,
    average_response_time INTEGER, -- Hours
    manager_check_ins INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_onboarding_metrics_workflow ON onboarding_metrics(workflow_id);
CREATE INDEX idx_onboarding_metrics_date ON onboarding_metrics(metric_date);


-- STEP 7: Insert default checklist templates
-- ============================================================================

-- General Onboarding Template
INSERT INTO onboarding_checklist_templates (id, name, description, is_active)
VALUES (
    'a1111111-1111-1111-1111-111111111111',
    'Standard New Hire Onboarding',
    'General onboarding checklist for all new employees',
    true
);

-- Insert template items
INSERT INTO onboarding_checklist_template_items (template_id, title, description, category, priority, day_offset, assigned_to_role, order_index, estimated_hours) VALUES
-- Day 1
('a1111111-1111-1111-1111-111111111111', 'Complete I-9 and W-4 Forms', 'Submit required tax and employment verification forms', 'administrative', 'critical', 0, 'new_hire', 1, 0.5),
('a1111111-1111-1111-1111-111111111111', 'Sign Employment Agreement', 'Review and sign employment contract and NDA', 'administrative', 'critical', 0, 'new_hire', 2, 0.5),
('a1111111-1111-1111-1111-111111111111', 'Set Up Workstation', 'Receive laptop, monitor, and other equipment', 'equipment', 'high', 0, 'it', 3, 1.0),
('a1111111-1111-1111-1111-111111111111', 'Create Email and System Accounts', 'Set up email, Slack, and system access', 'access', 'critical', 0, 'it', 4, 1.0),
('a1111111-1111-1111-1111-111111111111', 'Welcome Meeting with Manager', 'Initial meeting to discuss role, expectations, and goals', 'social', 'high', 0, 'manager', 5, 1.0),
('a1111111-1111-1111-1111-111111111111', 'Company Overview Presentation', 'Attend new hire orientation session', 'training', 'high', 0, 'hr', 6, 2.0),

-- Week 1
('a1111111-1111-1111-1111-111111111111', 'Complete Security Training', 'Cybersecurity awareness and best practices', 'training', 'critical', 1, 'new_hire', 7, 1.5),
('a1111111-1111-1111-1111-111111111111', 'Read Employee Handbook', 'Review company policies and procedures', 'documentation', 'high', 2, 'new_hire', 8, 2.0),
('a1111111-1111-1111-1111-111111111111', 'Meet Your Onboarding Buddy', 'Introduction to assigned mentor/buddy', 'social', 'high', 2, 'buddy', 9, 0.5),
('a1111111-1111-1111-1111-111111111111', 'Team Introduction Meeting', 'Meet team members and learn team dynamics', 'social', 'medium', 3, 'manager', 10, 1.0),
('a1111111-1111-1111-1111-111111111111', 'Set Up Development Environment', 'Install and configure necessary tools and software', 'access', 'high', 3, 'new_hire', 11, 3.0),

-- Week 2
('a1111111-1111-1111-1111-111111111111', 'Complete Benefits Enrollment', 'Choose health insurance and retirement plans', 'administrative', 'high', 7, 'new_hire', 12, 1.0),
('a1111111-1111-1111-1111-111111111111', 'Review Codebase/Systems Architecture', 'Understand system architecture and code structure', 'training', 'high', 7, 'new_hire', 13, 4.0),
('a1111111-1111-1111-1111-111111111111', 'First Small Project Assignment', 'Complete first coding task or project', 'training', 'medium', 10, 'new_hire', 14, 8.0),

-- Week 3-4
('a1111111-1111-1111-1111-111111111111', 'Attend Department Meetings', 'Participate in regular team meetings', 'social', 'medium', 14, 'new_hire', 15, 2.0),
('a1111111-1111-1111-1111-111111111111', 'Complete Compliance Training', 'Required compliance and ethics training', 'training', 'critical', 14, 'new_hire', 16, 2.0),
('a1111111-1111-1111-1111-111111111111', '30-Day Check-in with Manager', 'Review progress, address concerns, set goals', 'social', 'high', 30, 'manager', 17, 1.0),

-- Month 2-3
('a1111111-1111-1111-1111-111111111111', 'Complete First Major Project', 'Finish and deploy first significant contribution', 'training', 'high', 60, 'new_hire', 18, 40.0),
('a1111111-1111-1111-1111-111111111111', '90-Day Review', 'Formal performance review and feedback session', 'administrative', 'critical', 90, 'manager', 19, 2.0);


-- Developer-Specific Template
INSERT INTO onboarding_checklist_templates (id, name, description, department, is_active)
VALUES (
    'b2222222-2222-2222-2222-222222222222',
    'Software Developer Onboarding',
    'Technical onboarding for software developers',
    'Engineering',
    true
);

INSERT INTO onboarding_checklist_template_items (template_id, title, description, category, priority, day_offset, assigned_to_role, order_index, estimated_hours) VALUES
('b2222222-2222-2222-2222-222222222222', 'GitHub Access and SSH Setup', 'Configure Git, GitHub access, and SSH keys', 'access', 'critical', 0, 'it', 1, 1.0),
('b2222222-2222-2222-2222-222222222222', 'Clone Repositories', 'Clone main project repositories', 'access', 'high', 1, 'new_hire', 2, 0.5),
('b2222222-2222-2222-2222-222222222222', 'Local Development Setup', 'Set up local dev environment (Docker, DBs, etc)', 'access', 'high', 1, 'new_hire', 3, 4.0),
('b2222222-2222-2222-2222-222222222222', 'Code Style Guide Review', 'Read and understand coding standards', 'documentation', 'medium', 2, 'new_hire', 4, 1.0),
('b2222222-2222-2222-2222-222222222222', 'CI/CD Pipeline Overview', 'Learn deployment and testing processes', 'training', 'high', 3, 'new_hire', 5, 2.0),
('b2222222-2222-2222-2222-222222222222', 'Complete First Code Review', 'Submit PR and complete code review process', 'training', 'high', 7, 'new_hire', 6, 4.0),
('b2222222-2222-2222-2222-222222222222', 'Attend Architecture Review', 'Participate in system architecture discussion', 'training', 'medium', 14, 'new_hire', 7, 2.0);


-- STEP 8: Create views for common queries
-- ============================================================================

CREATE OR REPLACE VIEW onboarding_workflow_summary AS
SELECT 
    ow.id,
    ow.employee_id,
    e.first_name || ' ' || e.last_name as employee_name,
    e.email,
    ow.start_date,
    ow.expected_completion_date,
    ow.status,
    ow.overall_progress,
    buddy.first_name || ' ' || buddy.last_name as buddy_name,
    mgr.first_name || ' ' || mgr.last_name as manager_name,
    COUNT(DISTINCT ot.id) as total_tasks,
    COUNT(DISTINCT CASE WHEN ot.status = 'completed' THEN ot.id END) as completed_tasks,
    COUNT(DISTINCT CASE WHEN ot.status = 'pending' THEN ot.id END) as pending_tasks,
    COUNT(DISTINCT CASE WHEN ot.due_date < CURRENT_DATE AND ot.status != 'completed' THEN ot.id END) as overdue_tasks,
    ow.created_at,
    ow.updated_at
FROM onboarding_workflows ow
JOIN employees e ON e.id = ow.employee_id
LEFT JOIN employees buddy ON buddy.id = ow.assigned_buddy_id
LEFT JOIN employees mgr ON mgr.id = ow.assigned_manager_id
LEFT JOIN onboarding_tasks ot ON ot.workflow_id = ow.id
GROUP BY ow.id, e.first_name, e.last_name, e.email, buddy.first_name, buddy.last_name, mgr.first_name, mgr.last_name;


-- STEP 9: Create trigger to update workflow progress
-- ============================================================================

CREATE OR REPLACE FUNCTION update_workflow_progress()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE onboarding_workflows
    SET 
        overall_progress = (
            SELECT COALESCE(
                ROUND(
                    (COUNT(*) FILTER (WHERE status = 'completed')::DECIMAL / 
                    NULLIF(COUNT(*), 0)) * 100
                ),
                0
            )
            FROM onboarding_tasks
            WHERE workflow_id = NEW.workflow_id
        ),
        status = CASE
            WHEN (SELECT COUNT(*) FROM onboarding_tasks WHERE workflow_id = NEW.workflow_id AND status = 'completed') = 
                 (SELECT COUNT(*) FROM onboarding_tasks WHERE workflow_id = NEW.workflow_id)
            THEN 'completed'
            WHEN (SELECT COUNT(*) FROM onboarding_tasks WHERE workflow_id = NEW.workflow_id AND status != 'pending') > 0
            THEN 'in_progress'
            ELSE 'not_started'
        END,
        updated_at = NOW()
    WHERE id = NEW.workflow_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_workflow_progress
AFTER INSERT OR UPDATE ON onboarding_tasks
FOR EACH ROW
EXECUTE FUNCTION update_workflow_progress();


-- STEP 10: Verification queries
-- ============================================================================

-- Check tables created
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name LIKE 'onboarding%'
ORDER BY table_name;

-- Check indexes
SELECT tablename, indexname 
FROM pg_indexes 
WHERE tablename LIKE 'onboarding%'
ORDER BY tablename, indexname;

-- Check template items
SELECT 
    t.name,
    COUNT(ti.id) as item_count
FROM onboarding_checklist_templates t
LEFT JOIN onboarding_checklist_template_items ti ON ti.template_id = t.id
GROUP BY t.id, t.name;
