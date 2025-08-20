package services

import (
	"audio-book-ai/api/config"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	storage_go "github.com/supabase-community/storage-go"
)

// SupabaseStorageService handles file operations with Supabase Storage using storage-go client
type SupabaseStorageService struct {
	cfg    *config.Config
	client *storage_go.Client
}

// NewSupabaseStorageService creates a new Supabase storage service
func NewSupabaseStorageService(cfg *config.Config) *SupabaseStorageService {
	storageClient := storage_go.NewClient(
		cfg.SupabaseS3Endpoint,
		cfg.SupabaseSecretKey,
		nil,
	)

	service := &SupabaseStorageService{
		cfg:    cfg,
		client: storageClient,
	}

	return service
}

// CheckBucketExists checks if the storage bucket exists and is accessible
func (s *SupabaseStorageService) CheckBucketExists() error {
	_, err := s.client.GetBucket(s.cfg.SupabaseStorageBucket)
	if err != nil {
		return fmt.Errorf("bucket '%s' does not exist or is not accessible: %w", s.cfg.SupabaseStorageBucket, err)
	}

	fmt.Printf("Bucket '%s' exists and is accessible\n", s.cfg.SupabaseStorageBucket)
	return nil
}

// UploadFile uploads a file to Supabase Storage
func (s *SupabaseStorageService) UploadFile(file *multipart.FileHeader, uploadID string, fileName string) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create the file path in Supabase Storage
	storagePath := fmt.Sprintf("uploads/%s/%s", uploadID, fileName)

	// Upload to Supabase Storage using storage-go
	_, err = s.client.UploadFile(s.cfg.SupabaseStorageBucket, storagePath, src)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Supabase Storage: %w", err)
	}

	fmt.Printf("File uploaded successfully to Supabase Storage: %s\n", storagePath)

	// Return the public URL
	return s.GetPublicURL(storagePath), nil
}

// UploadFileFromPath uploads a local file to Supabase Storage
func (s *SupabaseStorageService) UploadFileFromPath(bucket, path, localPath string) error {
	// Open the local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Upload to Supabase Storage using storage-go
	_, err = s.client.UploadFile(bucket, path, file)
	if err != nil {
		return fmt.Errorf("failed to upload file to Supabase Storage: %w", err)
	}

	fmt.Printf("File uploaded successfully to Supabase Storage: %s/%s\n", bucket, path)
	return nil
}

// DeleteFile deletes a file from Supabase Storage
func (s *SupabaseStorageService) DeleteFile(uploadID string, fileName string) error {
	storagePath := fmt.Sprintf("uploads/%s/%s", uploadID, fileName)

	// Delete file using storage-go
	_, err := s.client.RemoveFile(s.cfg.SupabaseStorageBucket, []string{storagePath})
	if err != nil {
		return fmt.Errorf("failed to delete file from Supabase Storage: %w", err)
	}

	fmt.Printf("File deleted successfully from Supabase Storage: %s\n", storagePath)
	return nil
}

// GetPublicURL returns the public URL for a file
func (s *SupabaseStorageService) GetPublicURL(path string) string {
	// Construct the public URL manually since storage-go GetPublicUrl might not work as expected
	// Supabase public URL format: https://<project-ref>.supabase.co/storage/v1/object/public/<bucket>/<path>
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.cfg.SupabaseURL, s.cfg.SupabaseStorageBucket, path)
}

// CreateSignedURL generates a signed URL for a file
func (s *SupabaseStorageService) CreateSignedURL(path string, expiresIn time.Duration) (string, error) {
	// Convert duration to seconds for storage-go
	expireInSeconds := int(expiresIn.Seconds())

	// Create signed URL using storage-go
	result, err := s.client.CreateSignedUrl(s.cfg.SupabaseStorageBucket, path, expireInSeconds)
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %w", err)
	}

	return result.SignedURL, nil
}

// ValidateFileType checks if the file type is allowed
func (s *SupabaseStorageService) ValidateFileType(fileName string) error {
	ext := strings.ToLower(filepath.Ext(fileName))
	allowedExtensions := []string{".mp3", ".wav", ".m4a", ".aac", ".ogg", ".flac"}

	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return nil
		}
	}

	return fmt.Errorf("file type not allowed. Allowed types: %v", allowedExtensions)
}

// GetFileSize returns the size of the uploaded file
func (s *SupabaseStorageService) GetFileSize(file *multipart.FileHeader) int64 {
	return file.Size
}

// GetMimeType returns the MIME type of the file
func (s *SupabaseStorageService) GetMimeType(file *multipart.FileHeader) string {
	return file.Header.Get("Content-Type")
}

// DownloadFile downloads a file from Supabase Storage
func (s *SupabaseStorageService) DownloadFile(path string) ([]byte, error) {
	result, err := s.client.DownloadFile(s.cfg.SupabaseStorageBucket, path)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from Supabase Storage: %w", err)
	}
	return result, nil
}

// ListFiles lists all files in a bucket with optional search parameters
func (s *SupabaseStorageService) ListFiles(bucket string, path string, limit int, offset int) ([]storage_go.FileObject, error) {
	result, err := s.client.ListFiles(bucket, path, storage_go.FileSearchOptions{
		Limit:  limit,
		Offset: offset,
		SortByOptions: storage_go.SortBy{
			Column: "name",
			Order:  "asc",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files from Supabase Storage: %w", err)
	}
	return result, nil
}

// MoveFile moves a file from one location to another within the same bucket
func (s *SupabaseStorageService) MoveFile(bucket, fromPath, toPath string) error {
	_, err := s.client.MoveFile(bucket, fromPath, toPath)
	if err != nil {
		return fmt.Errorf("failed to move file in Supabase Storage: %w", err)
	}
	return nil
}

// UpdateFile replaces an existing file at the specified path
func (s *SupabaseStorageService) UpdateFile(bucket, path string, file multipart.File) error {
	_, err := s.client.UpdateFile(bucket, path, file)
	if err != nil {
		return fmt.Errorf("failed to update file in Supabase Storage: %w", err)
	}
	return nil
}
