# Audio Book Upload Flow

This document describes the complete upload flow for creating audio books with Supabase storage and Redis queue integration.

## Overview

The upload flow consists of four main steps:

1. **Create Upload Session** - Initialize an upload session for file uploads
2. **Upload Files** - Upload audio files to Supabase Storage
3. **Monitor Progress** - Track upload progress and status
4. **Create Audio Book** - Create audio book record and enqueue processing jobs

## API Endpoints

### 1. Create Upload Session

```
POST /api/v1/admin/uploads
Content-Type: application/json

{
  "upload_type": "single" | "chapters",
  "total_files": 1,
  "total_size_bytes": 0
}
```

**Response:**

```json
{
  "upload_id": "uuid",
  "status": "pending",
  "message": "Upload session created successfully"
}
```

### 2. Upload File

```
POST /api/v1/admin/uploads/{upload_id}/files
Content-Type: multipart/form-data

Form fields:
- file: audio file (mp3, wav, m4a, aac, ogg, flac)
- chapter_number: (optional) chapter number for chaptered uploads
- chapter_title: (optional) chapter title
```

**Response:**

```json
{
  "file_id": "uuid",
  "upload_id": "uuid",
  "file_name": "audio.mp3",
  "file_size_bytes": 1024000,
  "uploaded_at": "2024-01-01T00:00:00Z",
  "chapter_number": 1,
  "chapter_title": "Chapter 1"
}
```

### 3. Get Upload Progress

```
GET /api/v1/admin/uploads/{upload_id}/progress
```

**Response:**

```json
{
  "upload_id": "uuid",
  "status": "uploading",
  "total_files": 1,
  "uploaded_files": 1,
  "progress": 1.0,
  "total_size_bytes": 1024000,
  "uploaded_size_bytes": 1024000
}
```

### 4. Create Audio Book

```
POST /api/v1/admin/audiobooks
Content-Type: application/json

{
  "upload_id": "uuid",
  "title": "Audio Book Title",
  "author": "Author Name",
  "language": "en",
  "is_public": false,
  "description": "Optional description",
  "cover_image_url": "https://example.com/cover.jpg"
}
```

**Response:**

```json
{
  "audiobook_id": "uuid",
  "status": "pending",
  "message": "Audio book created successfully",
  "jobs_created": 4,
  "total_jobs": 4
}
```

## Architecture

### File Storage

- Files are uploaded to **Supabase Storage** in the `audio` bucket
- File path structure: `uploads/{upload_id}/{filename}`
- Files are validated for type (audio formats only) and size (max 500MB)

### Database

- Upload sessions and file metadata are stored in the database
- Audio book records are created with file references
- Processing jobs are tracked in the database

### Queue System

- **Redis** is used for job queuing
- Four job types are created:
  - `transcribe` - Audio transcription
  - `summarize` - Content summarization
  - `tag` - Content tagging
  - `embed` - Vector embeddings

## Environment Variables

Required environment variables:

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_PUBLISHABLE_KEY=your-publishable-key
SUPABASE_SECRET_KEY=your-secret-key
SUPABASE_STORAGE_BUCKET=audio

# Redis Configuration
REDIS_URL=redis://localhost:6379/0
JOBS_PREFIX=audiobooks

# API Configuration
API_PORT=8080
CORS_ORIGIN=http://localhost:3000
```

## Error Handling

### Upload Errors

- File type validation (only audio formats allowed)
- File size limits (500MB max)
- Storage upload failures
- Database transaction failures

### Queue Errors

- Redis connection failures (graceful degradation)
- Job enqueue failures (logged but don't fail the request)
- Retry mechanism for failed jobs

## Testing

To test the upload flow:

1. Start the API server:

```bash
cd api
go run main.go
```

2. Use curl or Postman to test the endpoints:

```bash
# Create upload session
curl -X POST http://localhost:8080/api/v1/admin/uploads \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"upload_type":"single","total_files":1,"total_size_bytes":0}'

# Upload file (replace UPLOAD_ID and TOKEN)
curl -X POST http://localhost:8080/api/v1/admin/uploads/UPLOAD_ID/files \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@audio.mp3"

# Create audio book (replace UPLOAD_ID and TOKEN)
curl -X POST http://localhost:8080/api/v1/admin/audiobooks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"upload_id":"UPLOAD_ID","title":"Test Book","author":"Test Author","language":"en","is_public":false}'
```

## Security

- All endpoints require authentication via JWT tokens
- Admin role required for upload and audio book creation
- File type validation prevents malicious uploads
- File size limits prevent abuse
- User ownership validation for all operations

## Monitoring

- Upload progress tracking
- Job status monitoring
- Error logging and alerting
- Queue statistics via Redis
- Database transaction monitoring
