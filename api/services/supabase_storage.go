package services

import (
	"audio-book-ai/api/config"
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SupabaseStorageService handles file operations with Supabase Storage using AWS S3 SDK
type SupabaseStorageService struct {
	cfg    *config.Config
	client *s3.Client
}

// NewSupabaseStorageService creates a new Supabase storage service
func NewSupabaseStorageService(cfg *config.Config) *SupabaseStorageService {

	s3Endpoint, s3Region, s3AccessKeyID, s3SecretKey := cfg.SupabaseS3Endpoint, cfg.SupabaseS3Region, cfg.SupabaseS3AccessKeyID, cfg.SupabaseS3SecretKey

	// Create custom AWS configuration for Supabase S3
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               s3Endpoint,
					SigningRegion:     s3Region,
					HostnameImmutable: true,
				}, nil
			},
		)),
		awsconfig.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     s3AccessKeyID,
				SecretAccessKey: s3SecretKey,
			},
		}),
		awsconfig.WithRegion(s3Region),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to load AWS config: %v", err))
	}

	// Create S3 client with path-style addressing
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for Supabase S3
	})

	service := &SupabaseStorageService{
		cfg:    cfg,
		client: s3Client,
	}

	return service
}

// CheckBucketExists checks if the storage bucket exists and is accessible
func (s *SupabaseStorageService) CheckBucketExists() error {
	ctx := context.Background()

	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.cfg.SupabaseStorageBucket),
	})
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

	// Upload to Supabase Storage
	ctx := context.Background()
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.SupabaseStorageBucket),
		Key:         aws.String(storagePath),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
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

	// Determine content type based on file extension
	contentType := "application/octet-stream"
	if ext := filepath.Ext(localPath); ext != "" {
		switch strings.ToLower(ext) {
		case ".png":
			contentType = "image/png"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".gif":
			contentType = "image/gif"
		case ".mp3":
			contentType = "audio/mpeg"
		case ".wav":
			contentType = "audio/wav"
		case ".m4a":
			contentType = "audio/mp4"
		case ".aac":
			contentType = "audio/aac"
		case ".ogg":
			contentType = "audio/ogg"
		case ".flac":
			contentType = "audio/flac"
		}
	}

	// Upload to Supabase Storage
	ctx := context.Background()
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(path),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to Supabase Storage: %w", err)
	}

	fmt.Printf("File uploaded successfully to Supabase Storage: %s/%s\n", bucket, path)
	return nil
}

// DeleteFile deletes a file from Supabase Storage
func (s *SupabaseStorageService) DeleteFile(uploadID string, fileName string) error {
	storagePath := fmt.Sprintf("uploads/%s/%s", uploadID, fileName)

	ctx := context.Background()
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.SupabaseStorageBucket),
		Key:    aws.String(storagePath),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from Supabase Storage: %w", err)
	}

	fmt.Printf("File deleted successfully from Supabase Storage: %s\n", storagePath)
	return nil
}

// GetPublicURL returns the public URL for a file
func (s *SupabaseStorageService) GetPublicURL(path string) string {
	// Supabase public URL format: https://<project-ref>.supabase.co/storage/v1/object/public/<bucket>/<path>
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.cfg.SupabaseURL, s.cfg.SupabaseStorageBucket, path)
}

// CreateSignedURL generates a signed URL for a file
func (s *SupabaseStorageService) CreateSignedURL(path string, expiresIn time.Duration) (string, error) {
	ctx := context.Background()

	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.SupabaseStorageBucket),
		Key:    aws.String(path),
	}, s3.WithPresignExpires(expiresIn))

	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %w", err)
	}

	return request.URL, nil
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
