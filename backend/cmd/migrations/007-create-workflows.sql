-- Onboarding Workflow System Database Schema
-- Safe migration that handles existing tables

-- Main workflow table
CREATE TABLE IF NOT EXISTS onboarding_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    template_name VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'completed', 'cancelled', 'on-hold'
    current_stage VARCHAR(100) NOT NULL DEFAULT 'pre-boarding',
    progress_percentage INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expected_completion TIMESTAMP WITH TIME ZONE,
    actual_completion TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES employees(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Workflow steps/tasks
CREATE TABLE IF NOT EXISTS workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,
    step_name VARCHAR(255) NOT NULL,
    step_type VARCHAR(50) NOT NULL, -- 'manual', 'integration', 'agent', 'approval', 'document'
    stage VARCHAR(100) NOT NULL, -- 'pre-boarding', 'day-1', 'week-1', 'month-1'
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'in-progress', 'completed', 'failed', 'skipped', 'blocked'
    description TEXT,
    dependencies JSONB DEFAULT '[]', -- Array of step IDs that must complete first
    assigned_to UUID REFERENCES employees(id),
    integration_type VARCHAR(50), -- 'docusign', 'background-check', 'doc-search', 'email'
    integration_config JSONB, -- Configuration for integration
    due_date TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    completed_by UUID REFERENCES employees(id),
    metadata JSONB DEFAULT '{}', -- Flexible data storage
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE workflow_steps 
ADD COLUMN IF NOT EXISTS required BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS auto_trigger BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS assigned_role VARCHAR(50),
ADD COLUMN IF NOT EXISTS due_days INTEGER;

-- Add template management columns if they don't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'workflow_steps' AND column_name = 'required') THEN
        ALTER TABLE workflow_steps ADD COLUMN required BOOLEAN DEFAULT true;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'workflow_steps' AND column_name = 'auto_trigger') THEN
        ALTER TABLE workflow_steps ADD COLUMN auto_trigger BOOLEAN DEFAULT false;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'workflow_steps' AND column_name = 'assigned_role') THEN
        ALTER TABLE workflow_steps ADD COLUMN assigned_role VARCHAR(50);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'workflow_steps' AND column_name = 'due_days') THEN
        ALTER TABLE workflow_steps ADD COLUMN due_days INTEGER;
    END IF;
END $$;

-- Integration tracking
CREATE TABLE IF NOT EXISTS workflow_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    integration_type VARCHAR(50) NOT NULL, -- 'docusign', 'background-check', 'doc-search'
    external_id VARCHAR(255), -- ID from external system (e.g., DocuSign envelope ID)
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'in-progress', 'completed', 'failed'
    request_payload JSONB,
    response_payload JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Exception tracking
CREATE TABLE IF NOT EXISTS workflow_exceptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    exception_type VARCHAR(100) NOT NULL, -- 'integration_failure', 'timeout', 'validation_error', 'manual_intervention'
    severity VARCHAR(20) NOT NULL, -- 'low', 'medium', 'high', 'critical'
    title VARCHAR(255) NOT NULL,
    description TEXT,
    resolution_status VARCHAR(50) NOT NULL DEFAULT 'open', -- 'open', 'in-progress', 'resolved', 'dismissed'
    assigned_to UUID REFERENCES employees(id),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID REFERENCES employees(id),
    resolution_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Agent activity logs
CREATE TABLE IF NOT EXISTS workflow_agent_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    agent_type VARCHAR(50) NOT NULL, -- 'progress', 'exception', 'document', 'recommendation'
    action VARCHAR(100) NOT NULL,
    input_data JSONB,
    output_data JSONB,
    execution_time_ms INTEGER,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Document tracking for workflow
CREATE TABLE IF NOT EXISTS workflow_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    document_name VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL, -- 'i9', 'w4', 'offer-letter', 'handbook', 'policy', etc.
    s3_key VARCHAR(500), -- S3 object key
    file_type VARCHAR(20), -- 'pdf', 'docx', 'gdoc'
    file_size INTEGER, -- bytes
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'uploaded', 'signed', 'approved', 'rejected'
    uploaded_by UUID REFERENCES employees(id),
    uploaded_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Workflow comments/notes
CREATE TABLE IF NOT EXISTS workflow_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES employees(id),
    comment TEXT NOT NULL,
    is_internal BOOLEAN DEFAULT false, -- Internal notes vs visible to employee
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_workflows_employee ON onboarding_workflows(employee_id);
CREATE INDEX IF NOT EXISTS idx_workflows_status ON onboarding_workflows(status);
CREATE INDEX IF NOT EXISTS idx_workflows_template ON onboarding_workflows(template_name);
CREATE INDEX IF NOT EXISTS idx_workflows_created_by ON onboarding_workflows(created_by);

CREATE INDEX IF NOT EXISTS idx_workflow_steps_workflow ON workflow_steps(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_workflow_order ON workflow_steps(workflow_id, step_order);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_status ON workflow_steps(status);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_assigned ON workflow_steps(assigned_to);

CREATE INDEX IF NOT EXISTS idx_workflow_integrations_workflow ON workflow_integrations(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_integrations_step ON workflow_integrations(step_id);

CREATE INDEX IF NOT EXISTS idx_workflow_exceptions_workflow ON workflow_exceptions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_exceptions_status ON workflow_exceptions(resolution_status);

CREATE INDEX IF NOT EXISTS idx_workflow_documents_workflow ON workflow_documents(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_comments_workflow ON workflow_comments(workflow_id);

-- Function to update workflow progress
CREATE OR REPLACE FUNCTION update_workflow_progress()
RETURNS TRIGGER AS $$
DECLARE
    total_steps INTEGER;
    completed_steps INTEGER;
    progress INTEGER;
BEGIN
    -- Count total and completed steps for the workflow
    SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'completed')
    INTO total_steps, completed_steps
    FROM workflow_steps
    WHERE workflow_id = NEW.workflow_id;
    
    -- Calculate progress percentage
    IF total_steps > 0 THEN
        progress := ROUND((completed_steps::DECIMAL / total_steps) * 100);
    ELSE
        progress := 0;
    END IF;
    
    -- Update workflow progress
    UPDATE onboarding_workflows
    SET progress_percentage = progress,
        updated_at = NOW()
    WHERE id = NEW.workflow_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update progress when step status changes
DROP TRIGGER IF EXISTS trigger_update_workflow_progress ON workflow_steps;
CREATE TRIGGER trigger_update_workflow_progress
AFTER INSERT OR UPDATE OF status ON workflow_steps
FOR EACH ROW
EXECUTE PROCEDURE update_workflow_progress();

-- Function to check if step dependencies are met
CREATE OR REPLACE FUNCTION check_step_dependencies(step_uuid UUID)
RETURNS BOOLEAN AS $$
DECLARE
    deps JSONB;
    dep_id TEXT;
    dep_status TEXT;
BEGIN
    -- Get dependencies for the step
    SELECT dependencies INTO deps
    FROM workflow_steps
    WHERE id = step_uuid;
    
    -- If no dependencies, return true
    IF deps IS NULL OR jsonb_array_length(deps) = 0 THEN
        RETURN TRUE;
    END IF;
    
    -- Check each dependency
    FOR dep_id IN SELECT jsonb_array_elements_text(deps)
    LOOP
        SELECT status INTO dep_status
        FROM workflow_steps
        WHERE id = dep_id::UUID;
        
        -- If any dependency is not completed, return false
        IF dep_status != 'completed' THEN
            RETURN FALSE;
        END IF;
    END LOOP;
    
    -- All dependencies met
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Workflows table stores workflow templates/definitions
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    workflow_type VARCHAR(50),  -- 'onboarding', 'offboarding', 'performance_review', etc.
    status VARCHAR(20) DEFAULT 'active',  -- 'active', 'inactive', 'draft'
    created_by UUID REFERENCES employees(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_workflows_type ON workflows(workflow_type);
CREATE INDEX IF NOT EXISTS idx_workflows_status ON workflows(status);
CREATE INDEX IF NOT EXISTS idx_workflows_created_by ON workflows(created_by);

-- Drop the incorrect foreign key constraint
ALTER TABLE workflows DROP CONSTRAINT IF EXISTS workflows_created_by_fkey;

-- Add correct foreign key to users table
ALTER TABLE workflows 
ADD CONSTRAINT workflows_created_by_fkey 
FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

-- Make stage nullable
ALTER TABLE workflow_steps ALTER COLUMN stage DROP NOT NULL;

-- Set default for stage
ALTER TABLE workflow_steps ALTER COLUMN stage SET DEFAULT 'pending';

-- Drop the existing foreign key constraint
ALTER TABLE workflow_steps DROP CONSTRAINT IF EXISTS workflow_steps_workflow_id_fkey;

-- Don't add it back - we need flexibility to reference either table
-- The application logic will ensure referential integrity

-- Add indexes for performance (without constraints)
CREATE INDEX IF NOT EXISTS idx_workflow_steps_workflow_id ON workflow_steps(workflow_id);

-- Note: workflow_steps table already exists and can be used for both:
-- - Template step definitions (when workflow_id references workflows table)
-- - Instance step tracking (when workflow_id references onboarding_workflows table)
-- The step_type, status, and other fields make it clear which type it is.

-- Add comment for clarity
COMMENT ON COLUMN workflow_steps.workflow_id IS 'References either workflows.id (for templates) or onboarding_workflows.id (for instances)';
COMMENT ON TABLE workflow_steps IS 'Used for both template step definitions and workflow instance steps';
COMMENT ON COLUMN workflow_steps.stage IS 'Workflow stage (only used for workflow instances, optional for templates)';
COMMENT ON COLUMN workflows.created_by IS 'User ID (from users table) who created this workflow template';
COMMENT ON TABLE workflows IS 'Workflow templates/definitions that can be instantiated';
COMMENT ON TABLE onboarding_workflows IS 'Active workflow instances for specific employees';
COMMENT ON TABLE onboarding_workflows IS 'Workflow instances for employee onboarding';
COMMENT ON TABLE workflow_steps IS 'Individual steps/tasks in a workflow';
COMMENT ON COLUMN workflow_steps.required IS 'Whether this step must be completed to proceed';
COMMENT ON COLUMN workflow_steps.auto_trigger IS 'Whether this step starts automatically';
COMMENT ON COLUMN workflow_steps.assigned_role IS 'Role responsible for this step (hr, manager, it, employee, etc.)';
COMMENT ON COLUMN workflow_steps.due_days IS 'Number of days to complete this step';