CREATE TABLE IF NOT EXISTS onboarding_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    task_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending',
    due_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    assigned_to UUID,
    documents_required BOOLEAN DEFAULT false,
    document_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to) REFERENCES employees(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_onboarding_employee_id ON onboarding_tasks(employee_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_status ON onboarding_tasks(status);