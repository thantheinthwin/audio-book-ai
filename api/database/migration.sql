-- Migration script to update database schema for normalized audiobook design
-- Run this script to migrate from the old schema to the new schema

-- Step 1: Add new columns to chapters table
ALTER TABLE chapters 
ADD COLUMN IF NOT EXISTS file_path VARCHAR(500),
ADD COLUMN IF NOT EXISTS file_url VARCHAR(500),
ADD COLUMN IF NOT EXISTS file_size_bytes BIGINT,
ADD COLUMN IF NOT EXISTS mime_type VARCHAR(100);

-- Step 2: Make file_path NOT NULL after data migration
-- (We'll do this after migrating data)

-- Step 3: Remove old columns from audiobooks table
-- (We'll do this after ensuring all data is migrated)

-- Note: This migration script should be run carefully in production
-- Consider backing up data before running these changes
