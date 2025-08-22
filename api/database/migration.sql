-- Migration: Add duration_seconds to upload_files and remove time fields from chapters
-- Date: 2024-01-XX

-- Step 1: Add duration_seconds column to upload_files table
ALTER TABLE upload_files 
ADD COLUMN duration_seconds INTEGER;

-- Step 2: Remove time-related columns from chapters table
ALTER TABLE chapters 
DROP COLUMN IF EXISTS start_time_seconds,
DROP COLUMN IF EXISTS end_time_seconds,
DROP COLUMN IF EXISTS duration_seconds;

-- Step 3: Add index for duration_seconds in upload_files for better query performance
CREATE INDEX IF NOT EXISTS idx_upload_files_duration_seconds ON upload_files(duration_seconds);

-- Step 4: Add comment to document the change
COMMENT ON COLUMN upload_files.duration_seconds IS 'Duration of the audio file in seconds. This field is populated during file upload and used for calculating total audiobook duration.';
