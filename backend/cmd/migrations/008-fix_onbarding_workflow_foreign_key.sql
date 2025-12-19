-- Fix foreign key constraint for onboarding_workflows.created_by
-- This changes it from referencing employees(id) to users(id)

-- Step 1: Drop the existing constraint
ALTER TABLE onboarding_workflows 
DROP CONSTRAINT IF EXISTS onboarding_workflows_created_by_fkey;

-- Step 2: Add the corrected constraint
ALTER TABLE onboarding_workflows 
ADD CONSTRAINT onboarding_workflows_created_by_fkey 
FOREIGN KEY (created_by) REFERENCES users(id);

-- Verify the change
SELECT 
    conname as constraint_name,
    conrelid::regclass as table_name,
    confrelid::regclass as referenced_table
FROM pg_constraint
WHERE conname = 'onboarding_workflows_created_by_fkey';
