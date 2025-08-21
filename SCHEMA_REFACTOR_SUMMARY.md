# Database Schema Refactoring Summary

## Overview

Refactored the database schema to normalize the relationship between audiobooks and chapters, removing redundant file-related fields from the audiobooks table and moving them to the chapters table where they belong.

## Changes Made

### 1. Database Schema (`api/database/schema.sql`)

#### Audiobooks Table

**Removed fields:**

- `file_size_bytes BIGINT`
- `file_path VARCHAR(500) NOT NULL`
- `file_url VARCHAR(500)`

**Remaining fields:**

- `id UUID PRIMARY KEY`
- `title VARCHAR(255) NOT NULL`
- `author VARCHAR(255) NOT NULL`
- `summary TEXT`
- `duration_seconds INTEGER`
- `cover_image_url VARCHAR(500)`
- `language VARCHAR(2) NOT NULL`
- `is_public BOOLEAN DEFAULT false`
- `status VARCHAR(20) NOT NULL DEFAULT 'pending'`
- `created_by UUID NOT NULL`
- `created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()`
- `updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()`

#### Chapters Table

**Added fields:**

- `file_path VARCHAR(500) NOT NULL`
- `file_url VARCHAR(500)`
- `file_size_bytes BIGINT`
- `mime_type VARCHAR(100)`

**Existing fields:**

- `id UUID PRIMARY KEY`
- `audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE`
- `chapter_number INTEGER NOT NULL`
- `title VARCHAR(255) NOT NULL`
- `start_time_seconds INTEGER`
- `end_time_seconds INTEGER`
- `duration_seconds INTEGER`
- `created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()`
- `UNIQUE(audiobook_id, chapter_number)`

### 2. Models (`api/models/audiobook.go`)

#### AudioBook Struct

**Removed fields:**

- `FileSizeBytes *int64`
- `FilePath string`
- `FileURL *string`

**Updated method:**

- `GetFileSizeFormatted()` now returns "See chapters for file sizes"

#### Chapter Struct

**Added fields:**

- `FilePath string`
- `FileURL *string`
- `FileSizeBytes *int64`
- `MimeType *string`

### 3. Database Operations (`api/database/postgres.go`)

#### Updated Functions:

- `CreateAudioBook()` - Removed file-related fields from INSERT
- `GetAudioBookByID()` - Removed file-related fields from SELECT
- `UpdateAudioBook()` - Removed file-related fields from UPDATE
- `CreateChapter()` - Added file-related fields to INSERT
- `GetChaptersByAudioBookID()` - Added file-related fields to SELECT
- `GetFirstChapterByAudioBookID()` - Added file-related fields to SELECT

### 4. Handlers (`api/handlers/handlers.go`)

#### CreateAudioBook Handler

**Major changes:**

- Now creates chapters for both single and chaptered uploads
- Single file uploads create one chapter with chapter_number = 1
- Chaptered uploads create multiple chapters based on upload files
- File paths, sizes, and URLs are now stored in chapters
- Redis job enqueueing uses the first chapter's file path for transcription

**Logic flow:**

1. Parse request and validate
2. Get upload session and verify ownership
3. Get upload files
4. Create audiobook record (without file info)
5. Create chapter records (with file info)
6. Create processing jobs
7. Enqueue jobs to Redis using first chapter's file path

### 5. Migration Script (`api/database/migration.sql`)

Created a migration script to:

- Add new columns to chapters table
- Prepare for removing old columns from audiobooks table
- Handle data migration safely

## Benefits of This Refactoring

### 1. **Proper Normalization**

- File-related data is now stored where it logically belongs (in chapters)
- Eliminates data redundancy between audiobooks and chapters
- Follows database normalization principles

### 2. **Consistency**

- All audiobooks now have chapters (single file = 1 chapter)
- File information is consistently stored in the same place
- No ambiguity about which file path/size to use

### 3. **Scalability**

- Easy to add more chapters to existing audiobooks
- Better support for chaptered content
- Cleaner separation of concerns

### 4. **Maintainability**

- Clearer data model
- Easier to understand relationships
- More predictable data access patterns

## Breaking Changes

### API Changes

- AudioBook responses no longer include `file_size_bytes`, `file_path`, `file_url`
- Chapter responses now include `file_path`, `file_url`, `file_size_bytes`, `mime_type`
- All audiobooks now have at least one chapter

### Database Changes

- Existing audiobooks will need data migration
- New schema requires chapters for all audiobooks
- File information moved from audiobooks to chapters

## Migration Notes

### For Existing Data

1. Run the migration script to add new columns
2. Migrate existing file data from audiobooks to chapters
3. Create chapter records for single-file audiobooks
4. Remove old columns from audiobooks table

### For New Deployments

- Use the updated schema directly
- No migration needed for fresh installations

## Testing Recommendations

1. **Unit Tests**: Update all tests to work with new schema
2. **Integration Tests**: Test the complete audiobook creation flow
3. **Migration Tests**: Test data migration for existing databases
4. **API Tests**: Verify API responses match new structure
5. **End-to-End Tests**: Test complete user workflows

## Next Steps

1. Update all tests to match new schema
2. Create data migration scripts for production
3. Update API documentation
4. Update frontend code to work with new response structure
5. Test thoroughly in staging environment
