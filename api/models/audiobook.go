package models

import (
	"encoding/json"
	"fmt"
	"time"

	"audio-book-ai/api/utils"

	"github.com/google/uuid"
)

// AudioBookStatus represents the processing status of an audio book
type AudioBookStatus string

const (
	StatusPending    AudioBookStatus = "pending"
	StatusProcessing AudioBookStatus = "processing"
	StatusCompleted  AudioBookStatus = "completed"
	StatusFailed     AudioBookStatus = "failed"
)

// JobType represents the type of processing job
type JobType string

const (
	JobTypeTranscribe JobType = "transcribe"
	JobTypeEmbed      JobType = "embed"
	JobTypeSummarize  JobType = "summarize"
)

// JobStatus represents the status of a processing job
type JobStatus string

const (
	JobStatusIdle      JobStatus = "idle"
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// OutputType represents the type of AI output
type OutputType string

const (
	OutputTypeSummary   OutputType = "summary"
	OutputTypeTags      OutputType = "tags"
	OutputTypeEmbedding OutputType = "embedding"
)

// EmbeddingType represents the type of embedding
type EmbeddingType string

const (
	EmbeddingTypeTitle       EmbeddingType = "title"
	EmbeddingTypeDescription EmbeddingType = "description"
	EmbeddingTypeSummary     EmbeddingType = "summary"
	EmbeddingTypeTranscript  EmbeddingType = "transcript"
)

// AudioBook represents an audio book in the system
type AudioBook struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	Title           string          `json:"title" db:"title" validate:"required,min=1,max=255"`
	Author          string          `json:"author" db:"author" validate:"required,min=1,max=255"`
	Summary         *string         `json:"summary,omitempty" db:"summary"`
	Tags            []*string       `json:"tags,omitempty" db:"tags"`
	DurationSeconds *int            `json:"duration_seconds,omitempty" db:"duration_seconds"`
	CoverImageURL   *string         `json:"cover_image_url,omitempty" db:"cover_image_url"`
	Language        string          `json:"language" db:"language" validate:"required,len=2"`
	IsPublic        bool            `json:"is_public" db:"is_public"`
	Status          AudioBookStatus `json:"status" db:"status" validate:"required"`
	CreatedBy       uuid.UUID       `json:"created_by" db:"created_by" validate:"required"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// Chapter represents a chapter within an audio book
type Chapter struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	AudiobookID     uuid.UUID  `json:"audiobook_id" db:"audiobook_id" validate:"required"`
	UploadFileID    *uuid.UUID `json:"upload_file_id,omitempty" db:"upload_file_id"`
	ChapterNumber   int        `json:"chapter_number" db:"chapter_number" validate:"required,min=1"`
	Title           string     `json:"title" db:"title" validate:"required,min=1,max=255"`
	FilePath        string     `json:"file_path" db:"file_path" validate:"required"`
	FileURL         *string    `json:"file_url,omitempty" db:"file_url"`
	FileSizeBytes   *int64     `json:"file_size_bytes,omitempty" db:"file_size_bytes"`
	MimeType        *string    `json:"mime_type,omitempty" db:"mime_type"`
	StartTime       *int       `json:"start_time_seconds,omitempty" db:"start_time_seconds"`
	EndTime         *int       `json:"end_time_seconds,omitempty" db:"end_time_seconds"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" db:"duration_seconds"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	Content         *string    `json:"content,omitempty" db:"content"`
}

// Transcript represents the transcription of an audio book
type Transcript struct {
	ID                    uuid.UUID        `json:"id" db:"id"`
	AudiobookID           uuid.UUID        `json:"audiobook_id" db:"audiobook_id" validate:"required"`
	Content               string           `json:"content" db:"content" validate:"required"`
	Segments              *json.RawMessage `json:"segments,omitempty" db:"segments"`
	Language              *string          `json:"language,omitempty" db:"language"`
	ConfidenceScore       *float64         `json:"confidence_score,omitempty" db:"confidence_score"`
	ProcessingTimeSeconds *int             `json:"processing_time_seconds,omitempty" db:"processing_time_seconds"`
	CreatedAt             time.Time        `json:"created_at" db:"created_at"`
}

// ChapterTranscript represents the transcription of a specific chapter
type ChapterTranscript struct {
	ID                    uuid.UUID        `json:"id" db:"id"`
	ChapterID             uuid.UUID        `json:"chapter_id" db:"chapter_id" validate:"required"`
	AudiobookID           uuid.UUID        `json:"audiobook_id" db:"audiobook_id" validate:"required"`
	Content               string           `json:"content" db:"content" validate:"required"`
	Segments              *json.RawMessage `json:"segments,omitempty" db:"segments"`
	Language              *string          `json:"language,omitempty" db:"language"`
	ConfidenceScore       *float64         `json:"confidence_score,omitempty" db:"confidence_score"`
	ProcessingTimeSeconds *int             `json:"processing_time_seconds,omitempty" db:"processing_time_seconds"`
	CreatedAt             time.Time        `json:"created_at" db:"created_at"`
}

// AIOutput represents AI-generated content for an audio book
type AIOutput struct {
	ID                    uuid.UUID       `json:"id" db:"id"`
	AudiobookID           uuid.UUID       `json:"audiobook_id" db:"audiobook_id" validate:"required"`
	OutputType            OutputType      `json:"output_type" db:"output_type" validate:"required"`
	Content               json.RawMessage `json:"content" db:"content" validate:"required"`
	ModelUsed             *string         `json:"model_used,omitempty" db:"model_used"`
	ProcessingTimeSeconds *int            `json:"processing_time_seconds,omitempty" db:"processing_time_seconds"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
}

// ChapterAIOutput represents AI-generated content for a specific chapter
type ChapterAIOutput struct {
	ID                    uuid.UUID       `json:"id" db:"id"`
	ChapterID             uuid.UUID       `json:"chapter_id" db:"chapter_id" validate:"required"`
	AudiobookID           uuid.UUID       `json:"audiobook_id" db:"audiobook_id" validate:"required"`
	OutputType            OutputType      `json:"output_type" db:"output_type" validate:"required"`
	Content               json.RawMessage `json:"content" db:"content" validate:"required"`
	ModelUsed             *string         `json:"model_used,omitempty" db:"model_used"`
	ProcessingTimeSeconds *int            `json:"processing_time_seconds,omitempty" db:"processing_time_seconds"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
}

// ProcessingJob represents a background processing job
type ProcessingJob struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	AudiobookID  uuid.UUID  `json:"audiobook_id" db:"audiobook_id" validate:"required"`
	ChapterID    *uuid.UUID `json:"chapter_id,omitempty" db:"chapter_id"`
	JobType      JobType    `json:"job_type" db:"job_type" validate:"required"`
	Status       JobStatus  `json:"status" db:"status" validate:"required"`
	RedisJobID   *string    `json:"redis_job_id,omitempty" db:"redis_job_id"`
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"`
	StartedAt    *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// Tag represents a tag that can be applied to audio books
type Tag struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Category  *string   `json:"category,omitempty" db:"category"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AudioBookTag represents the relationship between audio books and tags
type AudioBookTag struct {
	AudiobookID     uuid.UUID `json:"audiobook_id" db:"audiobook_id" validate:"required"`
	TagID           uuid.UUID `json:"tag_id" db:"tag_id" validate:"required"`
	ConfidenceScore *float64  `json:"confidence_score,omitempty" db:"confidence_score"`
	IsAIGenerated   bool      `json:"is_ai_generated" db:"is_ai_generated"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// AudioBookEmbedding represents vector embeddings for semantic search
type AudioBookEmbedding struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	AudiobookID   uuid.UUID     `json:"audiobook_id" db:"audiobook_id" validate:"required"`
	Embedding     []float64     `json:"embedding" db:"embedding" validate:"required"`
	EmbeddingType EmbeddingType `json:"embedding_type" db:"embedding_type" validate:"required"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
}

// Request/Response models

// CreateAudioBookRequest represents the request to create a new audio book
type CreateAudioBookRequest struct {
	Title         string  `json:"title" validate:"required,min=1,max=255"`
	Author        string  `json:"author" validate:"required,min=1,max=255"`
	Language      string  `json:"language" validate:"required,len=2"`
	IsPublic      bool    `json:"is_public"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
}

// UpdateAudioBookRequest represents the request to update an audio book
type UpdateAudioBookRequest struct {
	Title         *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Author        *string `json:"author,omitempty" validate:"omitempty,min=1,max=255"`
	Language      *string `json:"language,omitempty" validate:"omitempty,len=2"`
	IsPublic      *bool   `json:"is_public,omitempty"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
}

// AudioBookWithDetails represents an audio book with all related data
type AudioBookWithDetails struct {
	AudioBook
	Chapters       []Chapter       `json:"chapters,omitempty"`
	AIOutputs      []AIOutput      `json:"ai_outputs,omitempty"`
	ProcessingJobs []ProcessingJob `json:"processing_jobs,omitempty"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query    string  `json:"query" validate:"required,min=1"`
	Limit    int     `json:"limit" validate:"min=1,max=100"`
	Offset   int     `json:"offset" validate:"min=0"`
	Language *string `json:"language,omitempty"`
	IsPublic *bool   `json:"is_public,omitempty"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	AudioBooks []AudioBook `json:"audiobooks"`
	Total      int         `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}

// JobStatusResponse represents the status of processing jobs
type JobStatusResponse struct {
	AudiobookID   uuid.UUID       `json:"audiobook_id"`
	Jobs          []ProcessingJob `json:"jobs"`
	OverallStatus AudioBookStatus `json:"overall_status"`
	Progress      float64         `json:"progress"` // 0.0 to 1.0
	EstimatedTime *int            `json:"estimated_time_seconds,omitempty"`
}

// Helper methods

// IsCompleted returns true if the audio book processing is completed
func (ab *AudioBook) IsCompleted() bool {
	return ab.Status == StatusCompleted
}

// IsFailed returns true if the audio book processing has failed
func (ab *AudioBook) IsFailed() bool {
	return ab.Status == StatusFailed
}

// IsProcessing returns true if the audio book is currently being processed
func (ab *AudioBook) IsProcessing() bool {
	return ab.Status == StatusProcessing
}

// IsPending returns true if the audio book is pending processing
func (ab *AudioBook) IsPending() bool {
	return ab.Status == StatusPending
}

// GetDurationFormatted returns the duration in a human-readable format
func (ab *AudioBook) GetDurationFormatted() string {
	if ab.DurationSeconds == nil {
		return "Unknown"
	}

	seconds := *ab.DurationSeconds
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	}
	return fmt.Sprintf("%dm %ds", minutes, secs)
}

// GetFileSizeFormatted returns the file size in a human-readable format
// Note: File size is now stored in chapters, not in the audiobook itself
func (ab *AudioBook) GetFileSizeFormatted() string {
	return "See chapters for file sizes"
}

// IsFirstChapter returns true if this is the first chapter
func (c *Chapter) IsFirstChapter() bool {
	return c.ChapterNumber == 1
}

// GetDurationFormatted returns the chapter duration in a human-readable format
func (c *Chapter) GetDurationFormatted() string {
	if c.DurationSeconds == nil {
		return "Unknown"
	}

	seconds := *c.DurationSeconds
	minutes := seconds / 60
	secs := seconds % 60

	return fmt.Sprintf("%dm %ds", minutes, secs)
}

// Validate validates the audio book struct
func (ab *AudioBook) Validate() error {
	return utils.GetValidator().Struct(ab)
}

// Validate validates the chapter struct
func (c *Chapter) Validate() error {
	return utils.GetValidator().Struct(c)
}

// Validate validates the transcript struct
func (t *Transcript) Validate() error {
	return utils.GetValidator().Struct(t)
}

// Validate validates the chapter transcript struct
func (ct *ChapterTranscript) Validate() error {
	return utils.GetValidator().Struct(ct)
}

// Validate validates the AI output struct
func (ao *AIOutput) Validate() error {
	return utils.GetValidator().Struct(ao)
}

// Validate validates the chapter AI output struct
func (cao *ChapterAIOutput) Validate() error {
	return utils.GetValidator().Struct(cao)
}

// Validate validates the processing job struct
func (pj *ProcessingJob) Validate() error {
	return utils.GetValidator().Struct(pj)
}

// Validate validates the tag struct
func (tag *Tag) Validate() error {
	return utils.GetValidator().Struct(tag)
}

// Validate validates the audio book tag struct
func (abt *AudioBookTag) Validate() error {
	return utils.GetValidator().Struct(abt)
}

// Validate validates the audio book embedding struct
func (abe *AudioBookEmbedding) Validate() error {
	return utils.GetValidator().Struct(abe)
}

// Validate validates the create audio book request
func (car *CreateAudioBookRequest) Validate() error {
	return utils.GetValidator().Struct(car)
}

// Validate validates the update audio book request
func (uar *UpdateAudioBookRequest) Validate() error {
	return utils.GetValidator().Struct(uar)
}

// Validate validates the search request
func (sr *SearchRequest) Validate() error {
	return utils.GetValidator().Struct(sr)
}
