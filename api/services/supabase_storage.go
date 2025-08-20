package services

import (
	"audio-book-ai/api/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// SupabaseStorageService handles file operations with Supabase Storage
type SupabaseStorageService struct {
	cfg *config.Config
}

// NewSupabaseStorageService creates a new Supabase storage service
func NewSupabaseStorageService(cfg *config.Config) *SupabaseStorageService {
	return &SupabaseStorageService{
		cfg: cfg,
	}
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

	// Create the request body
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// Add the file to the form
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	
	// Copy file content to the form
	if _, err := io.Copy(part, src); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}
	
	writer.Close()

	// Create HTTP request to Supabase Storage API
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.cfg.SupabaseURL, s.cfg.SupabaseStorageBucket, storagePath)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+s.cfg.SupabaseSecretKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("storage API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to get the file URL
	var uploadResp struct {
		Key string `json:"Key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Return the public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.cfg.SupabaseURL, s.cfg.SupabaseStorageBucket, storagePath)
	return publicURL, nil
}

// DeleteFile deletes a file from Supabase Storage
func (s *SupabaseStorageService) DeleteFile(uploadID string, fileName string) error {
	storagePath := fmt.Sprintf("uploads/%s/%s", uploadID, fileName)
	
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.cfg.SupabaseURL, s.cfg.SupabaseStorageBucket, storagePath)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.cfg.SupabaseSecretKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetFileURL returns the public URL for a file
func (s *SupabaseStorageService) GetFileURL(uploadID string, fileName string) string {
	storagePath := fmt.Sprintf("uploads/%s/%s", uploadID, fileName)
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.cfg.SupabaseURL, s.cfg.SupabaseStorageBucket, storagePath)
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
