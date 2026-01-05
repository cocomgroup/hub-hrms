-- ============================================================================
-- FINAL CORRECTED MIGRATION: Workflow Engine + New Hire Onboarding
-- ============================================================================
-- This migration properly separates:
-- 1. Generic Workflow Engine (for ALL lifecycle events)
-- 2. Specialized New Hire Onboarding (AI-powered with tasks)
-- ============================================================================

BEGIN;

-- ============================================================================
-- PART 1: WORKFLOW ENGINE (Generic System)
-- ============================================================================

-- Workflow templates (definitions/blueprints for any lifecycle event)
CREATE TABLE IF NOT EXISTS workflow_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    workflow_type VARCHAR(50) NOT NULL,  -- 'onboarding', 'offboarding', 'performance', 'leave', 'vendor'
    status VARCHAR(20) DEFAULT 'active',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT chk_workflow_template_type CHECK (workflow_type IN ('onboarding', 'offboarding', 'performance', 'leave', 'vendor', 'other'))
);

-- Employee workflows (active workflow instances)
CREATE TABLE IF NOT EXISTS employee_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    template_id UUID REFERENCES workflow_templates(id) ON DELETE SET NULL,
    template_name VARCHAR(100) NOT NULL,
    workflow_type VARCHAR(50) NOT NULL DEFAULT 'onboarding',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    current_stage VARCHAR(100) NOT NULL DEFAULT 'pre-boarding',
    progress_percentage INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expected_completion TIMESTAMP WITH TIME ZONE,
    actual_completion TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT chk_employee_workflow_type CHECK (workflow_type IN ('onboarding', 'offboarding', 'performance', 'leave', 'vendor', 'other'))
);

-- Workflow steps (for both templates and instances)
CREATE TABLE IF NOT EXISTS workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL,
    step_order INTEGER NOT NULL,
    step_name VARCHAR(255) NOT NULL,
    step_type VARCHAR(50) NOT NULL,
    stage VARCHAR(100) DEFAULT 'pending',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    description TEXT,
    dependencies JSONB DEFAULT '[]',
    assigned_to UUID REFERENCES employees(id) ON DELETE SET NULL,
    integration_type VARCHAR(50),
    integration_config JSONB,
    due_date TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    completed_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',
    required BOOLEAN DEFAULT true,
    auto_trigger BOOLEAN DEFAULT false,
    assigned_role VARCHAR(50),
    due_days INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Workflow integrations
CREATE TABLE IF NOT EXISTS workflow_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL,
    step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    integration_type VARCHAR(50) NOT NULL,
    external_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    request_payload JSONB,
    response_payload JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Workflow exceptions
CREATE TABLE IF NOT EXISTS workflow_exceptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    exception_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    resolution_status VARCHAR(50) NOT NULL DEFAULT 'open',
    assigned_to UUID REFERENCES employees(id) ON DELETE SET NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    resolution_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Workflow documents
CREATE TABLE IF NOT EXISTS workflow_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    document_name VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    s3_key VARCHAR(500),
    file_type VARCHAR(20),
    file_size INTEGER,
    status VARCHAR(50) DEFAULT 'pending',
    uploaded_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Workflow comments
CREATE TABLE IF NOT EXISTS workflow_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    is_internal BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- PART 2: NEW HIRE ONBOARDING (Specialized, AI-powered)
-- ============================================================================

-- New hire onboarding workflows
CREATE TABLE IF NOT EXISTS new_hire_onboardings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- Links to generic workflow engine (optional)
    workflow_template_id UUID REFERENCES workflow_templates(id) ON DELETE SET NULL,
    employee_workflow_id UUID REFERENCES employee_workflows(id) ON DELETE SET NULL,
    
    start_date DATE NOT NULL,
    expected_completion_date DATE,
    actual_completion_date DATE,
    status VARCHAR(50) NOT NULL DEFAULT 'not_started',
    overall_progress INTEGER DEFAULT 0,
    assigned_buddy_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    assigned_manager_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    
    CONSTRAINT valid_progress CHECK (overall_progress >= 0 AND overall_progress <= 100),
    CONSTRAINT valid_onboarding_status CHECK (status IN ('not_started', 'in_progress', 'completed', 'overdue'))
);

-- Onboarding tasks (AI-assisted)
CREATE TABLE IF NOT EXISTS onboarding_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES new_hire_onboardings(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    priority VARCHAR(50) DEFAULT 'medium',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    assigned_to UUID REFERENCES employees(id) ON DELETE SET NULL,
    due_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    completed_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    order_index INTEGER DEFAULT 0,
    is_mandatory BOOLEAN DEFAULT true,
    estimated_hours DECIMAL(5,2),
    actual_hours DECIMAL(5,2),
    dependencies JSONB,
    attachments JSONB,
    ai_generated BOOLEAN DEFAULT false,
    ai_suggestions TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT valid_task_priority CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT valid_task_status CHECK (status IN ('pending', 'in_progress', 'completed', 'blocked', 'skipped'))
);

-- Onboarding interactions (AI chat)
CREATE TABLE IF NOT EXISTS onboarding_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES new_hire_onboardings(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    interaction_type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    ai_response TEXT,
    sentiment VARCHAR(50),
    requires_action BOOLEAN DEFAULT false,
    action_taken BOOLEAN DEFAULT false,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT valid_interaction_type CHECK (interaction_type IN 
        ('chat', 'reminder', 'suggestion', 'check_in', 'escalation')),
    CONSTRAINT valid_sentiment CHECK (sentiment IN 
        ('positive', 'neutral', 'negative', 'concerned') OR sentiment IS NULL)
);

-- Onboarding milestones
CREATE TABLE IF NOT EXISTS onboarding_milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES new_hire_onboardings(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    target_date DATE,
    completed_date DATE,
    status VARCHAR(50) DEFAULT 'pending',
    celebration_sent BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT valid_milestone_status CHECK (status IN ('pending', 'completed', 'missed'))
);

-- Onboarding checklist templates
CREATE TABLE IF NOT EXISTS onboarding_checklist_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    department VARCHAR(100),
    role_type VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Onboarding checklist template items
CREATE TABLE IF NOT EXISTS onboarding_checklist_template_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES onboarding_checklist_templates(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    priority VARCHAR(50) DEFAULT 'medium',
    day_offset INTEGER DEFAULT 0,
    is_mandatory BOOLEAN DEFAULT true,
    estimated_hours DECIMAL(5,2),
    order_index INTEGER DEFAULT 0,
    assigned_to_role VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Onboarding metrics
CREATE TABLE IF NOT EXISTS onboarding_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES new_hire_onboardings(id) ON DELETE CASCADE,
    metric_date DATE NOT NULL,
    tasks_completed INTEGER DEFAULT 0,
    tasks_pending INTEGER DEFAULT 0,
    tasks_overdue INTEGER DEFAULT 0,
    engagement_score INTEGER,
    ai_interactions INTEGER DEFAULT 0,
    average_response_time INTEGER,
    manager_check_ins INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- PART 3: INDEXES
-- ============================================================================

-- Workflow engine indexes
CREATE INDEX IF NOT EXISTS idx_workflow_templates_type ON workflow_templates(workflow_type);
CREATE INDEX IF NOT EXISTS idx_employee_workflows_employee ON employee_workflows(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_workflows_type ON employee_workflows(workflow_type);
CREATE INDEX IF NOT EXISTS idx_employee_workflows_status ON employee_workflows(status);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_workflow ON workflow_steps(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_workflow_order ON workflow_steps(workflow_id, step_order);

-- Onboarding indexes
CREATE INDEX IF NOT EXISTS idx_new_hire_onboardings_employee ON new_hire_onboardings(employee_id);
CREATE INDEX IF NOT EXISTS idx_new_hire_onboardings_status ON new_hire_onboardings(status);
CREATE INDEX IF NOT EXISTS idx_new_hire_onboardings_start_date ON new_hire_onboardings(start_date);
CREATE INDEX IF NOT EXISTS idx_new_hire_onboardings_workflow_template ON new_hire_onboardings(workflow_template_id);
CREATE INDEX IF NOT EXISTS idx_new_hire_onboardings_employee_workflow ON new_hire_onboardings(employee_workflow_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_tasks_workflow ON onboarding_tasks(workflow_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_tasks_status ON onboarding_tasks(status);
CREATE INDEX IF NOT EXISTS idx_onboarding_tasks_assigned ON onboarding_tasks(assigned_to);
CREATE INDEX IF NOT EXISTS idx_onboarding_tasks_due_date ON onboarding_tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_onboarding_interactions_workflow ON onboarding_interactions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_milestones_workflow ON onboarding_milestones(workflow_id);
CREATE INDEX IF NOT EXISTS idx_template_items_template ON onboarding_checklist_template_items(template_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_metrics_workflow ON onboarding_metrics(workflow_id);

-- ============================================================================
-- PART 4: TRIGGERS AND FUNCTIONS
-- ============================================================================

-- Update employee workflow progress
CREATE OR REPLACE FUNCTION update_employee_workflow_progress()
RETURNS TRIGGER AS $$
DECLARE
    total_steps INTEGER;
    completed_steps INTEGER;
    progress INTEGER;
BEGIN
    SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'completed')
    INTO total_steps, completed_steps
    FROM workflow_steps
    WHERE workflow_id = NEW.workflow_id
    AND stage IS NOT NULL;
    
    IF total_steps > 0 THEN
        progress := ROUND((completed_steps::DECIMAL / total_steps) * 100);
    ELSE
        progress := 0;
    END IF;
    
    UPDATE employee_workflows
    SET progress_percentage = progress,
        updated_at = NOW()
    WHERE id = NEW.workflow_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_employee_workflow_progress ON workflow_steps;
CREATE TRIGGER trigger_update_employee_workflow_progress
AFTER INSERT OR UPDATE OF status ON workflow_steps
FOR EACH ROW
WHEN (NEW.stage IS NOT NULL)
EXECUTE FUNCTION update_employee_workflow_progress();

-- Update onboarding progress
CREATE OR REPLACE FUNCTION update_onboarding_progress()
RETURNS TRIGGER AS $$
DECLARE
    total_tasks INTEGER;
    completed_tasks INTEGER;
    progress INTEGER;
BEGIN
    SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'completed')
    INTO total_tasks, completed_tasks
    FROM onboarding_tasks
    WHERE workflow_id = NEW.workflow_id;
    
    IF total_tasks > 0 THEN
        progress := ROUND((completed_tasks::DECIMAL / total_tasks) * 100);
    ELSE
        progress := 0;
    END IF;
    
    UPDATE new_hire_onboardings
    SET 
        overall_progress = progress,
        status = CASE
            WHEN progress = 100 THEN 'completed'
            WHEN progress > 0 THEN 'in_progress'
            ELSE 'not_started'
        END,
        updated_at = NOW()
    WHERE id = NEW.workflow_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_onboarding_progress ON onboarding_tasks;
CREATE TRIGGER trigger_update_onboarding_progress
AFTER INSERT OR UPDATE OF status ON onboarding_tasks
FOR EACH ROW
EXECUTE FUNCTION update_onboarding_progress();

-- ============================================================================
-- PART 5: DEFAULT DATA
-- ============================================================================

-- Insert default onboarding templates
INSERT INTO onboarding_checklist_templates (id, name, description, is_active)
VALUES (
    'a1111111-1111-1111-1111-111111111111',
    'Standard New Hire Onboarding',
    'General onboarding checklist for all new employees',
    true
) ON CONFLICT (id) DO NOTHING;

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
('a1111111-1111-1111-1111-111111111111', '90-Day Review', 'Formal performance review and feedback session', 'administrative', 'critical', 90, 'manager', 19, 2.0)
ON CONFLICT DO NOTHING;

-- Developer-specific template
INSERT INTO onboarding_checklist_templates (id, name, description, department, is_active)
VALUES (
    'b2222222-2222-2222-2222-222222222222',
    'Software Developer Onboarding',
    'Technical onboarding for software developers',
    'Engineering',
    true
) ON CONFLICT (id) DO NOTHING;

INSERT INTO onboarding_checklist_template_items (template_id, title, description, category, priority, day_offset, assigned_to_role, order_index, estimated_hours) VALUES
('b2222222-2222-2222-2222-222222222222', 'GitHub Access and SSH Setup', 'Configure Git, GitHub access, and SSH keys', 'access', 'critical', 0, 'it', 1, 1.0),
('b2222222-2222-2222-2222-222222222222', 'Clone Repositories', 'Clone main project repositories', 'access', 'high', 1, 'new_hire', 2, 0.5),
('b2222222-2222-2222-2222-222222222222', 'Local Development Setup', 'Set up local dev environment (Docker, DBs, etc)', 'access', 'high', 1, 'new_hire', 3, 4.0),
('b2222222-2222-2222-2222-222222222222', 'Code Style Guide Review', 'Read and understand coding standards', 'documentation', 'medium', 2, 'new_hire', 4, 1.0),
('b2222222-2222-2222-2222-222222222222', 'CI/CD Pipeline Overview', 'Learn deployment and testing processes', 'training', 'high', 3, 'new_hire', 5, 2.0),
('b2222222-2222-2222-2222-222222222222', 'Complete First Code Review', 'Submit PR and complete code review process', 'training', 'high', 7, 'new_hire', 6, 4.0),
('b2222222-2222-2222-2222-222222222222', 'Attend Architecture Review', 'Participate in system architecture discussion', 'training', 'medium', 14, 'new_hire', 7, 2.0)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- PART 6: VIEWS
-- ============================================================================

CREATE OR REPLACE VIEW onboarding_workflow_summary AS
SELECT 
    nho.id,
    nho.employee_id,
    e.first_name || ' ' || e.last_name as employee_name,
    e.email,
    nho.start_date,
    nho.expected_completion_date,
    nho.status,
    nho.overall_progress,
    COALESCE(buddy.first_name || ' ' || buddy.last_name, '') as buddy_name,
    COALESCE(mgr.first_name || ' ' || mgr.last_name, '') as manager_name,
    COUNT(DISTINCT ot.id) as total_tasks,
    COUNT(DISTINCT CASE WHEN ot.status = 'completed' THEN ot.id END) as completed_tasks,
    COUNT(DISTINCT CASE WHEN ot.status = 'pending' THEN ot.id END) as pending_tasks,
    COUNT(DISTINCT CASE WHEN ot.due_date < CURRENT_DATE AND ot.status != 'completed' THEN ot.id END) as overdue_tasks,
    nho.created_at,
    nho.updated_at
FROM new_hire_onboardings nho
JOIN employees e ON e.id = nho.employee_id
LEFT JOIN employees buddy ON buddy.id = nho.assigned_buddy_id
LEFT JOIN employees mgr ON mgr.id = nho.assigned_manager_id
LEFT JOIN onboarding_tasks ot ON ot.workflow_id = nho.id
GROUP BY nho.id, e.first_name, e.last_name, e.email, buddy.first_name, buddy.last_name, mgr.first_name, mgr.last_name;

COMMIT;

-- ============================================================================
-- VERIFICATION
-- ============================================================================

SELECT 'Migration completed successfully!' as status;

-- Show created tables
SELECT 
    table_name,
    (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = t.table_name) as column_count
FROM information_schema.tables t
WHERE table_schema = 'public'
AND (table_name LIKE 'workflow%' OR table_name LIKE 'employee_workflow%' OR table_name LIKE '%onboarding%')
ORDER BY table_name;