-- Migration: Add retry_count and max_retries to processing_jobs table
-- Date: 2024-01-XX
-- Description: Add retry tracking fields to processing_jobs table for better job monitoring

-- Add retry_count column with default value 0
ALTER TABLE processing_jobs 
ADD COLUMN IF NOT EXISTS retry_count INTEGER DEFAULT 0;

-- Add max_retries column with default value 3
ALTER TABLE processing_jobs 
ADD COLUMN IF NOT EXISTS max_retries INTEGER DEFAULT 3;

-- Update existing records to have default values
UPDATE processing_jobs 
SET retry_count = 0, max_retries = 3 
WHERE retry_count IS NULL OR max_retries IS NULL;

-- Add comments to document the new columns
COMMENT ON COLUMN processing_jobs.retry_count IS 'Number of times this job has been retried';
COMMENT ON COLUMN processing_jobs.max_retries IS 'Maximum number of retries allowed for this job';
