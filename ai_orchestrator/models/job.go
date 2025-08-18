package models

import (
	"time"

	"github.com/google/uuid"
)

// Job represents a processing job
type Job struct {
	ID          uuid.UUID  `json:"id"`
	AudiobookID uuid.UUID  `json:"audiobook_id"`
	JobType     string     `json:"job_type"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
}

// Job status constants
const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)

// Job type constants
const (
	JobTypeTag   = "tag"
	JobTypeEmbed = "embed"
)
