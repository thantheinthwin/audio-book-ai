package models

import (
	"time"

	"github.com/google/uuid"
)

// Job represents a processing job
type Job struct {
	ID           uuid.UUID  `json:"id"`
	AudiobookID  uuid.UUID  `json:"audiobook_id"`
	ChapterID    *uuid.UUID `json:"chapter_id,omitempty"`
	JobType      string     `json:"job_type"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
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
	JobTypeEmbed     = "embed"
	JobTypeSummarize = "summarize"
)
