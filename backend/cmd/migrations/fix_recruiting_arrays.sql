-- Fix recruiting array columns for PostgreSQL compatibility
-- Run this migration to fix the array parsing issue

-- First, check if columns are text type (they should be text[] array type)
DO $$ 
BEGIN
    -- Update strengths column to be array type if it's text
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'candidates' 
        AND column_name = 'strengths' 
        AND data_type != 'ARRAY'
    ) THEN
        -- Convert existing data to proper array format
        ALTER TABLE candidates 
        ALTER COLUMN strengths TYPE text[] 
        USING CASE 
            WHEN strengths IS NULL OR strengths = '' THEN ARRAY[]::text[]
            WHEN strengths LIKE '{%}' THEN strengths::text[]
            ELSE ARRAY[strengths]::text[]
        END;
    END IF;

    -- Update weaknesses column to be array type if it's text
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'candidates' 
        AND column_name = 'weaknesses' 
        AND data_type != 'ARRAY'
    ) THEN
        -- Convert existing data to proper array format
        ALTER TABLE candidates 
        ALTER COLUMN weaknesses TYPE text[] 
        USING CASE 
            WHEN weaknesses IS NULL OR weaknesses = '' THEN ARRAY[]::text[]
            WHEN weaknesses LIKE '{%}' THEN weaknesses::text[]
            ELSE ARRAY[weaknesses]::text[]
        END;
    END IF;

    -- Update skills column to be array type if it's text
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'candidates' 
        AND column_name = 'skills' 
        AND data_type != 'ARRAY'
    ) THEN
        -- Convert existing data to proper array format
        ALTER TABLE candidates 
        ALTER COLUMN skills TYPE text[] 
        USING CASE 
            WHEN skills IS NULL OR skills = '' THEN ARRAY[]::text[]
            WHEN skills LIKE '{%}' THEN skills::text[]
            ELSE ARRAY[skills]::text[]
        END;
    END IF;
END $$;

-- Set defaults for array columns
ALTER TABLE candidates 
ALTER COLUMN strengths SET DEFAULT ARRAY[]::text[];

ALTER TABLE candidates 
ALTER COLUMN weaknesses SET DEFAULT ARRAY[]::text[];

ALTER TABLE candidates 
ALTER COLUMN skills SET DEFAULT ARRAY[]::text[];

-- Update any NULL values to empty arrays
UPDATE candidates 
SET strengths = ARRAY[]::text[] 
WHERE strengths IS NULL;

UPDATE candidates 
SET weaknesses = ARRAY[]::text[] 
WHERE weaknesses IS NULL;

UPDATE candidates 
SET skills = ARRAY[]::text[] 
WHERE skills IS NULL;

-- Verify the changes
SELECT 
    column_name, 
    data_type, 
    is_nullable 
FROM information_schema.columns 
WHERE table_name = 'candidates' 
AND column_name IN ('strengths', 'weaknesses', 'skills');