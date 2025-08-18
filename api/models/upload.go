package models

import (
	"fmt"
	"time"

	"audio-book-ai/api/utils"

	"github.com/google/uuid"
)

// UploadType represents the type of upload
type UploadType string

const (
	UploadTypeSingle   UploadType = "single"   // Single audio file
	UploadTypeChapters UploadType = "chapters" // Multiple chapter files
)

// UploadStatus represents the status of an upload
type UploadStatus string

const (
	UploadStatusPending   UploadStatus = "pending"
	UploadStatusUploading UploadStatus = "uploading"
	UploadStatusCompleted UploadStatus = "completed"
	UploadStatusFailed    UploadStatus = "failed"
)

// Upload represents an upload session
type Upload struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	UserID        uuid.UUID    `json:"user_id" db:"user_id" validate:"required"`
	UploadType    UploadType   `json:"upload_type" db:"upload_type" validate:"required"`
	Status        UploadStatus `json:"status" db:"status" validate:"required"`
	TotalFiles    int          `json:"total_files" db:"total_files"`
	UploadedFiles int          `json:"uploaded_files" db:"uploaded_files"`
	TotalSize     int64        `json:"total_size_bytes" db:"total_size_bytes"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
}

// UploadFile represents a file being uploaded
type UploadFile struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	UploadID      uuid.UUID    `json:"upload_id" db:"upload_id" validate:"required"`
	FileName      string       `json:"file_name" db:"file_name" validate:"required"`
	FileSize      int64        `json:"file_size_bytes" db:"file_size_bytes" validate:"required"`
	MimeType      string       `json:"mime_type" db:"mime_type" validate:"required"`
	FilePath      string       `json:"file_path" db:"file_path" validate:"required"`
	ChapterNumber *int         `json:"chapter_number,omitempty" db:"chapter_number"`
	ChapterTitle  *string      `json:"chapter_title,omitempty" db:"chapter_title"`
	Status        UploadStatus `json:"status" db:"status" validate:"required"`
	Error         *string      `json:"error,omitempty" db:"error"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
}

// Request/Response models

// CreateUploadRequest represents the request to start an upload
type CreateUploadRequest struct {
	UploadType UploadType `json:"upload_type" validate:"required"`
	TotalFiles int        `json:"total_files" validate:"required,min=1,max=100"`
	TotalSize  int64      `json:"total_size_bytes" validate:"required,min=1"`
}

// UploadFileRequest represents the request to upload a file
type UploadFileRequest struct {
	UploadID      uuid.UUID `json:"upload_id" validate:"required"`
	FileName      string    `json:"file_name" validate:"required"`
	FileSize      int64     `json:"file_size_bytes" validate:"required,min=1"`
	MimeType      string    `json:"mime_type" validate:"required"`
	ChapterNumber *int      `json:"chapter_number,omitempty"`
	ChapterTitle  *string   `json:"chapter_title,omitempty"`
}

// CreateAudioBookFromUploadRequest represents the request to create an audio book from upload
type CreateAudioBookFromUploadRequest struct {
	UploadID      uuid.UUID `json:"upload_id" validate:"required"`
	Title         string    `json:"title" validate:"required,min=1,max=255"`
	Author        string    `json:"author" validate:"required,min=1,max=255"`
	Description   *string   `json:"description,omitempty"`
	Language      string    `json:"language" validate:"required,len=2"`
	IsPublic      bool      `json:"is_public"`
	CoverImageURL *string   `json:"cover_image_url,omitempty"`
}

// UploadProgressResponse represents the progress of an upload
type UploadProgressResponse struct {
	UploadID      uuid.UUID    `json:"upload_id"`
	Status        UploadStatus `json:"status"`
	TotalFiles    int          `json:"total_files"`
	UploadedFiles int          `json:"uploaded_files"`
	Progress      float64      `json:"progress"` // 0.0 to 1.0
	TotalSize     int64        `json:"total_size_bytes"`
	UploadedSize  int64        `json:"uploaded_size_bytes"`
	EstimatedTime *int         `json:"estimated_time_seconds,omitempty"`
}

// UploadFileInfo represents information about an uploaded file
type UploadFileInfo struct {
	ID            uuid.UUID    `json:"id"`
	FileName      string       `json:"file_name"`
	FileSize      int64        `json:"file_size_bytes"`
	MimeType      string       `json:"mime_type"`
	ChapterNumber *int         `json:"chapter_number,omitempty"`
	ChapterTitle  *string      `json:"chapter_title,omitempty"`
	Status        UploadStatus `json:"status"`
	Error         *string      `json:"error,omitempty"`
	UploadedAt    time.Time    `json:"uploaded_at"`
}

// UploadDetailsResponse represents detailed information about an upload
type UploadDetailsResponse struct {
	Upload
	Files []UploadFileInfo `json:"files"`
}

// FileUploadResponse represents the response after uploading a file
type FileUploadResponse struct {
	FileID        uuid.UUID `json:"file_id"`
	UploadID      uuid.UUID `json:"upload_id"`
	FileName      string    `json:"file_name"`
	FileSize      int64     `json:"file_size_bytes"`
	UploadedAt    time.Time `json:"uploaded_at"`
	ChapterNumber *int      `json:"chapter_number,omitempty"`
	ChapterTitle  *string   `json:"chapter_title,omitempty"`
}

// Helper methods

// GetProgress returns the upload progress as a percentage
func (u *Upload) GetProgress() float64 {
	if u.TotalFiles == 0 {
		return 0.0
	}
	return float64(u.UploadedFiles) / float64(u.TotalFiles)
}

// IsCompleted returns true if the upload is completed
func (u *Upload) IsCompleted() bool {
	return u.Status == UploadStatusCompleted
}

// IsFailed returns true if the upload has failed
func (u *Upload) IsFailed() bool {
	return u.Status == UploadStatusFailed
}

// IsUploading returns true if the upload is in progress
func (u *Upload) IsUploading() bool {
	return u.Status == UploadStatusUploading
}

// IsPending returns true if the upload is pending
func (u *Upload) IsPending() bool {
	return u.Status == UploadStatusPending
}

// GetFileSizeFormatted returns the file size in a human-readable format
func (uf *UploadFile) GetFileSizeFormatted() string {
	bytes := uf.FileSize
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// IsChapterFile returns true if this is a chapter file
func (uf *UploadFile) IsChapterFile() bool {
	return uf.ChapterNumber != nil
}

// Validate validates the upload struct
func (u *Upload) Validate() error {
	return utils.GetValidator().Struct(u)
}

// Validate validates the upload file struct
func (uf *UploadFile) Validate() error {
	return utils.GetValidator().Struct(uf)
}

// Validate validates the create upload request
func (cur *CreateUploadRequest) Validate() error {
	return utils.GetValidator().Struct(cur)
}

// Validate validates the upload file request
func (ufr *UploadFileRequest) Validate() error {
	return utils.GetValidator().Struct(ufr)
}

// Validate validates the create audio book from upload request
func (cabur *CreateAudioBookFromUploadRequest) Validate() error {
	return utils.GetValidator().Struct(cabur)
}
