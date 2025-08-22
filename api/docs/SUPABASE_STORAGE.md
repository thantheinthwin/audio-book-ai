# Supabase Storage Service

This service provides a clean interface for uploading files to Supabase Storage using the AWS S3 SDK v2.

## Features

- ✅ Upload files from multipart form data (for web uploads)
- ✅ Upload files from local file system
- ✅ Generate public URLs
- ✅ Generate signed URLs with expiration
- ✅ Delete files
- ✅ File type validation
- ✅ Bucket existence checking
- ✅ Proper error handling
- ✅ Uses AWS S3 SDK v2 with S3-compatible endpoint: `/storage/v1/s3`

## Configuration

Set these environment variables:

```bash
# Basic Supabase configuration
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_SECRET_KEY=your-service-role-key
SUPABASE_STORAGE_BUCKET=your-bucket-name

# S3-compatible storage configuration (required for AWS S3 SDK)
SUPABASE_S3_ENDPOINT=https://your-project-ref.supabase.co/storage/v1/s3
SUPABASE_S3_REGION=us-east-1
SUPABASE_S3_ACCESS_KEY_ID=supabase
SUPABASE_S3_SECRET_KEY=your-service-role-key
```

## Usage

### Basic Setup

```go
import (
    "audio-book-ai/api/config"
    "audio-book-ai/api/services"
)

// Initialize configuration
cfg := config.New()

// Create storage service
storageService := services.NewSupabaseStorageService(cfg)

// Check if bucket exists
if err := storageService.CheckBucketExists(); err != nil {
    log.Printf("Bucket check failed: %v", err)
}
```

### Upload File from Local Path

```go
// Upload a local file to Supabase Storage
err := storageService.UploadFileFromPath(
    "your-bucket",           // bucket name
    "avatars/user123.png",   // storage path
    "./local/avatar.png",    // local file path
)
if err != nil {
    log.Printf("Upload failed: %v", err)
}
```

### Upload File from Multipart Form

```go
// For web uploads (multipart form data)
fileURL, err := storageService.UploadFile(
    fileHeader,              // *multipart.FileHeader
    "upload-session-id",     // upload ID
    "filename.png",          // filename
)
if err != nil {
    log.Printf("Upload failed: %v", err)
}
```

### Generate URLs

```go
// Get public URL
publicURL := storageService.GetPublicURL("avatars/user123.png")

// Create signed URL (expires in 1 hour)
signedURL, err := storageService.CreateSignedURL(
    "avatars/user123.png",
    1 * time.Hour,
)
```

### Delete Files

```go
// Delete a file
err := storageService.DeleteFile("upload-id", "filename.png")
if err != nil {
    log.Printf("Delete failed: %v", err)
}
```

### Validate File Types

```go
// Check if file type is allowed
err := storageService.ValidateFileType("audio.mp3")
if err != nil {
    log.Printf("Invalid file type: %v", err)
}
```

## Working Example

Run the example to test the storage functionality:

```bash
# Set your environment variables
export SUPABASE_URL="https://your-project-ref.supabase.co"
export SUPABASE_SECRET_KEY="your-service-role-key"
export SUPABASE_STORAGE_BUCKET="your-bucket-name"

# Run the example
cd api
go run example_storage.go
```

The example will:

1. ✅ Check if the bucket exists
2. 📤 Upload `avatar.png` to `avatars/users/123/avatar.png`
3. 🌐 Generate a public URL
4. 🔐 Create a signed URL (expires in 1 hour)
5. 🎉 Print all URLs

## Supported File Types

The service automatically detects content types for:

- Images: `.png`, `.jpg`, `.jpeg`, `.gif`
- Audio: `.mp3`, `.wav`, `.m4a`, `.aac`, `.ogg`, `.flac`

## Error Handling

All functions return proper errors that can be wrapped and logged:

```go
err := storageService.UploadFileFromPath(bucket, path, localPath)
if err != nil {
    return fmt.Errorf("failed to upload file: %w", err)
}
```

## Integration with Web API

The service is already integrated with the web API handlers. When users upload files through the web interface, the service automatically:

1. Validates file type
2. Uploads to Supabase Storage using AWS S3 SDK
3. Returns the public URL
4. Stores the URL in the database

## Technical Details

### AWS S3 SDK Configuration

The service uses AWS S3 SDK v2 with custom configuration for Supabase:

- **Endpoint**: `https://<project-ref>.supabase.co/storage/v1/s3`
- **Region**: `us-east-1` (Supabase default)
- **Access Key**: `supabase` (Supabase uses this as the access key)
- **Secret Key**: Your Supabase service role key
- **Path Style**: Enabled (required for Supabase S3)

### AWS SDK Dependencies

```go
require (
    github.com/aws/aws-sdk-go-v2 v1.24.1
    github.com/aws/aws-sdk-go-v2/config v1.26.6
    github.com/aws/aws-sdk-go-v2/credentials v1.16.16
    github.com/aws/aws-sdk-go-v2/service/s3 v1.48.1
)
```

### S3 Operations Used

- `HeadBucket` - Check if bucket exists
- `PutObject` - Upload files
- `DeleteObject` - Delete files
- `PresignGetObject` - Generate signed URLs

## Troubleshooting

### Common Issues

1. **Bucket doesn't exist**: Create the bucket in your Supabase dashboard
2. **Permission denied**: Ensure you're using the service role key, not the anon key
3. **Invalid endpoint**: Make sure your Supabase URL is correct
4. **File too large**: Check Supabase storage limits
5. **AWS SDK errors**: Ensure all AWS SDK dependencies are properly installed

### Debug Mode

The service logs detailed information when uploading:

```
Supabase Storage Service initialized:
  URL: https://your-project-ref.supabase.co
  S3 Endpoint: https://your-project-ref.supabase.co/storage/v1/s3
  Bucket: your-bucket-name
  Secret Key: eyJhbGciOi...
File uploaded successfully to Supabase Storage: avatars/users/123/avatar.png
```

## API Reference

### Methods

- `NewSupabaseStorageService(cfg *config.Config) *SupabaseStorageService`
- `CheckBucketExists() error`
- `UploadFile(file *multipart.FileHeader, uploadID, fileName string) (string, error)`
- `UploadFileFromPath(bucket, path, localPath string) error`
- `DeleteFile(uploadID, fileName string) error`
- `GetPublicURL(path string) string`
- `CreateSignedURL(path string, expiresIn time.Duration) (string, error)`
- `ValidateFileType(fileName string) error`
- `GetFileSize(file *multipart.FileHeader) int64`
- `GetMimeType(file *multipart.FileHeader) string`
