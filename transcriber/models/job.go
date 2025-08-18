package models

import (
	"time"

	"github.com/google/uuid"
)

// Job represents a transcription job
type Job struct {
	ID          uuid.UUID `json:"id"`
	AudiobookID uuid.UUID `json:"audiobook_id"`
	JobType     string    `json:"job_type"`
	Status      string    `json:"status"`
	FilePath    string    `json:"file_path"`
	Language    string    `json:"language"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
}

// JobStatus represents the possible job statuses
const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)

// JobType represents the possible job types
const (
	JobTypeTranscribe = "transcribe"
)
