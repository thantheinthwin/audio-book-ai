# Complete Audiobook Creation Flow

This document describes the complete audiobook creation flow that has been implemented, including transcription, auto-tagging, and summary generation using Redis queues.

## Overview

The audiobook creation flow consists of the following steps:

1. **Upload Session Creation** - Initialize an upload session for file uploads
2. **File Upload** - Upload audio files to Supabase Storage
3. **Audiobook Creation** - Create audiobook record and enqueue processing jobs
4. **Background Processing** - Process jobs using Redis queues:
   - **Transcription** - Convert audio to text using Rev.ai
   - **Summarization** - Generate AI-powered summaries using Gemini (triggered after chapter 1 transcription)
   - **Tagging** - Auto-generate relevant tags using Gemini (triggered after chapter 1 transcription)
   - **Embedding** - Create vector embeddings for search
5. **Status Updates** - Monitor job progress and update audiobook status

## Architecture

### Services

1. **API Service** (`api/`) - Main API server handling uploads and audiobook creation
2. **Transcriber Service** (`transcriber/`) - Handles audio transcription using Rev.ai
3. **Worker Service** (`worker/`) - Handles AI processing (summarization, tagging, embeddings)
4. **Redis** - Message queue for job processing
5. **PostgreSQL** - Database for storing audiobook data and job status

### Database Schema

The system uses the following main tables:

- `audiobooks` - Main audiobook records
- `chapters` - Chapter information for multi-chapter audiobooks
- `transcripts` - Transcribed text content
- `ai_outputs` - AI-generated content (summaries, tags, embeddings)
- `processing_jobs` - Background job tracking
- `uploads` - Upload session management
- `upload_files` - Individual file uploads

## API Endpoints

### Upload Management

```
POST /api/v1/admin/uploads
POST /api/v1/admin/uploads/{id}/files
GET /api/v1/admin/uploads/{id}/progress
GET /api/v1/admin/uploads/{id}
DELETE /api/v1/admin/uploads/{id}
```

### Audiobook Management

```
POST /api/v1/admin/audiobooks
GET /api/v1/admin/audiobooks/{id}/jobs
PUT /api/v1/admin/audiobooks/{id}
DELETE /api/v1/admin/audiobooks/{id}
```

### Job Management

```
POST /api/v1/admin/jobs/{job_id}/status
```

## Complete Flow Implementation

### 1. Upload Session Creation

```bash
curl -X POST http://localhost:8080/api/v1/admin/uploads \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "upload_type": "single",
    "total_files": 1,
    "total_size_bytes": 1024000
  }'
```

Response:

```json
{
  "upload_id": "uuid",
  "status": "pending",
  "message": "Upload session created successfully"
}
```

### 2. File Upload

```bash
curl -X POST http://localhost:8080/api/v1/admin/uploads/{upload_id}/files \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@audio.mp3"
```

Response:

```json
{
  "file_id": "uuid",
  "upload_id": "uuid",
  "file_name": "audio.mp3",
  "file_size_bytes": 1024000,
  "uploaded_at": "2024-01-01T00:00:00Z"
}
```

### 3. Audiobook Creation

```bash
curl -X POST http://localhost:8080/api/v1/admin/audiobooks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "upload_id": "uuid",
    "title": "My Audio Book",
    "author": "Author Name",
    "description": "Book description",
    "language": "en",
    "is_public": false
  }'
```

Response:

```json
{
  "audiobook_id": "uuid",
  "status": "processing",
  "message": "Audio book created successfully",
  "jobs_created": 4,
  "total_jobs": 4
}
```

### 4. Job Processing

When an audiobook is created, the following jobs are automatically enqueued:

1. **Transcription Job** - Sent to transcriber service
2. **Summarization Job** - Sent to worker service
3. **Tagging Job** - Sent to worker service
4. **Embedding Job** - Sent to worker service

### 5. Job Status Monitoring

```bash
curl -X GET http://localhost:8080/api/v1/admin/audiobooks/{audiobook_id}/jobs \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Response:

```json
{
  "audiobook_id": "uuid",
  "jobs": [
    {
      "id": "uuid",
      "audiobook_id": "uuid",
      "job_type": "transcribe",
      "status": "completed",
      "started_at": "2024-01-01T00:00:00Z",
      "completed_at": "2024-01-01T00:05:00Z"
    }
  ],
  "overall_status": "processing",
  "progress": 0.25,
  "total_jobs": 4,
  "completed_jobs": 1,
  "failed_jobs": 0
}
```

## Background Services

### Transcriber Service

The transcriber service consumes transcription jobs from Redis and processes them using Rev.ai:

```bash
cd transcriber
go run main.go
```

Configuration:

- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `REV_AI_API_KEY` - Rev.ai API key
- `API_BASE_URL` - API service URL for status updates

### Worker Service

The worker service consumes AI processing jobs from Redis and processes them using Gemini:

```bash
cd worker
go run main.go
```

Configuration:

- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `GEMINI_API_KEY` - Gemini API key
- `API_BASE_URL` - API service URL for status updates

## Redis Queue Structure

The system uses Redis sorted sets for job queues:

- `audiobook:queue:transcribe` - Transcription jobs
- `audiobook:queue:summarize` - Summarization jobs
- `audiobook:queue:tag` - Tagging jobs
- `audiobook:queue:embed` - Embedding jobs
- `audiobook:processing:*` - Jobs currently being processed
- `audiobook:failed:*` - Failed jobs

## Job Processing Flow

1. **Job Creation**: When an audiobook is created, jobs are enqueued in Redis
2. **Job Consumption**: Services consume jobs from their respective queues
3. **Job Processing**: Jobs are processed using external APIs (Rev.ai, Gemini)
4. **Status Updates**: Job status is updated via HTTP calls to the API
5. **Audiobook Status Update**: When all jobs complete, audiobook status is updated
6. **Summary Update**: If summarization completes, the audiobook summary is updated

## Error Handling

- **Retry Logic**: Failed jobs are retried up to 3 times with exponential backoff
- **Dead Letter Queue**: Permanently failed jobs are moved to failed queues
- **Status Tracking**: All job statuses are tracked in the database
- **Graceful Degradation**: If Redis is unavailable, jobs are still created in the database

## Testing

Use the provided test script to verify the complete flow:

```bash
go run test_upload_flow.go
```

This script will:

1. Create an upload session
2. Upload a test audio file
3. Create an audiobook
4. Monitor job progress
5. Report completion status

## Environment Variables

### API Service

```
DATABASE_URL=postgresql://...
REDIS_URL=redis://localhost:6379/0
SUPABASE_URL=...
SUPABASE_ANON_KEY=...
SUPABASE_SERVICE_ROLE_KEY=...
```

### Transcriber Service

```
DATABASE_URL=postgresql://...
REDIS_URL=redis://localhost:6379/0
REV_AI_API_KEY=...
API_BASE_URL=http://localhost:8080
```

### Worker Service

```
DATABASE_URL=postgresql://...
REDIS_URL=redis://localhost:6379/0
GEMINI_API_KEY=...
API_BASE_URL=http://localhost:8080
```

## Deployment

1. **Start Redis**: `docker run -d -p 6379:6379 redis:alpine`
2. **Start PostgreSQL**: Use your preferred method
3. **Start API Service**: `cd api && go run main.go`
4. **Start Transcriber**: `cd transcriber && go run main.go`
5. **Start Worker**: `cd worker && go run main.go`

## Monitoring

- **Job Status**: Monitor via `/api/v1/admin/audiobooks/{id}/jobs`
- **Queue Stats**: Check Redis queue sizes
- **Service Logs**: Monitor service logs for errors
- **Database**: Check processing_jobs table for job status

## Future Enhancements

1. **Real-time Updates**: WebSocket support for real-time job status updates
2. **Job Prioritization**: Priority-based job processing
3. **Batch Processing**: Support for processing multiple audiobooks
4. **Advanced Error Handling**: More sophisticated retry and error recovery
5. **Metrics and Monitoring**: Prometheus metrics and Grafana dashboards
6. **Scaling**: Horizontal scaling of worker services
