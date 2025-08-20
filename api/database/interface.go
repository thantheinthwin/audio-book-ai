package database

import (
	"context"

	"audio-book-ai/api/models"

	"github.com/google/uuid"
)

// Repository defines the interface for all database operations
type Repository interface {
	// AudioBook operations
	CreateAudioBook(ctx context.Context, audiobook *models.AudioBook) error
	GetAudioBookByID(ctx context.Context, id uuid.UUID) (*models.AudioBook, error)
	GetAudioBookWithDetails(ctx context.Context, id uuid.UUID) (*models.AudioBookWithDetails, error)
	UpdateAudioBook(ctx context.Context, audiobook *models.AudioBook) error
	DeleteAudioBook(ctx context.Context, id uuid.UUID) error
	ListAudioBooks(ctx context.Context, limit, offset int, isPublic *bool) ([]models.AudioBook, int, error)
	GetAudioBooksByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error)
	UpdateAudioBookStatus(ctx context.Context, id uuid.UUID, status models.AudioBookStatus) error
	CheckAndUpdateAudioBookStatus(ctx context.Context, audiobookID uuid.UUID) error

	// Chapter operations
	CreateChapter(ctx context.Context, chapter *models.Chapter) error
	GetChapterByID(ctx context.Context, id uuid.UUID) (*models.Chapter, error)
	GetChaptersByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Chapter, error)
	GetFirstChapterByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Chapter, error)
	UpdateChapter(ctx context.Context, chapter *models.Chapter) error
	DeleteChapter(ctx context.Context, id uuid.UUID) error
	DeleteChaptersByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error

	// Transcript operations
	CreateTranscript(ctx context.Context, transcript *models.Transcript) error
	GetTranscriptByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Transcript, error)
	UpdateTranscript(ctx context.Context, transcript *models.Transcript) error
	DeleteTranscript(ctx context.Context, id uuid.UUID) error

	// Chapter Transcript operations
	CreateChapterTranscript(ctx context.Context, transcript *models.ChapterTranscript) error
	GetChapterTranscriptByChapterID(ctx context.Context, chapterID uuid.UUID) (*models.ChapterTranscript, error)
	GetChapterTranscriptsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ChapterTranscript, error)
	UpdateChapterTranscript(ctx context.Context, transcript *models.ChapterTranscript) error
	DeleteChapterTranscript(ctx context.Context, id uuid.UUID) error
	DeleteChapterTranscriptsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error

	// AI Output operations
	CreateAIOutput(ctx context.Context, output *models.AIOutput) error
	GetAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AIOutput, error)
	GetAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.AIOutput, error)
	UpdateAIOutput(ctx context.Context, output *models.AIOutput) error
	DeleteAIOutput(ctx context.Context, id uuid.UUID) error

	// Chapter AI Output operations
	CreateChapterAIOutput(ctx context.Context, output *models.ChapterAIOutput) error
	GetChapterAIOutputsByChapterID(ctx context.Context, chapterID uuid.UUID) ([]models.ChapterAIOutput, error)
	GetChapterAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ChapterAIOutput, error)
	GetFirstChapterAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.ChapterAIOutput, error)
	UpdateChapterAIOutput(ctx context.Context, output *models.ChapterAIOutput) error
	DeleteChapterAIOutput(ctx context.Context, id uuid.UUID) error
	DeleteChapterAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error

	// Processing Job operations
	CreateProcessingJob(ctx context.Context, job *models.ProcessingJob) error
	GetProcessingJobsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ProcessingJob, error)
	GetProcessingJobByID(ctx context.Context, id uuid.UUID) (*models.ProcessingJob, error)
	UpdateProcessingJob(ctx context.Context, job *models.ProcessingJob) error
	GetPendingJobs(ctx context.Context, jobType models.JobType, limit int) ([]models.ProcessingJob, error)
	GetJobsByStatus(ctx context.Context, status models.JobStatus, limit int) ([]models.ProcessingJob, error)

	// Tag operations
	CreateTag(ctx context.Context, tag *models.Tag) error
	GetTagByID(ctx context.Context, id uuid.UUID) (*models.Tag, error)
	GetTagByName(ctx context.Context, name string) (*models.Tag, error)
	GetTagsByCategory(ctx context.Context, category string) ([]models.Tag, error)
	UpdateTag(ctx context.Context, tag *models.Tag) error
	DeleteTag(ctx context.Context, id uuid.UUID) error
	ListTags(ctx context.Context, limit, offset int) ([]models.Tag, int, error)

	// AudioBook Tag operations
	CreateAudioBookTag(ctx context.Context, audiobookTag *models.AudioBookTag) error
	GetTagsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Tag, error)
	GetAudioBooksByTagID(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error)
	DeleteAudioBookTag(ctx context.Context, audiobookID, tagID uuid.UUID) error
	DeleteAllAudioBookTags(ctx context.Context, audiobookID uuid.UUID) error

	// AudioBook Embedding operations
	CreateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error
	GetEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AudioBookEmbedding, error)
	GetEmbeddingByType(ctx context.Context, audiobookID uuid.UUID, embeddingType models.EmbeddingType) (*models.AudioBookEmbedding, error)
	UpdateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error
	DeleteAudioBookEmbedding(ctx context.Context, id uuid.UUID) error
	DeleteEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error

	// Upload operations
	CreateUpload(ctx context.Context, upload *models.Upload) error
	GetUploadByID(ctx context.Context, id uuid.UUID) (*models.Upload, error)
	GetUploadsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Upload, int, error)
	UpdateUpload(ctx context.Context, upload *models.Upload) error
	DeleteUpload(ctx context.Context, id uuid.UUID) error

	// Upload File operations
	CreateUploadFile(ctx context.Context, uploadFile *models.UploadFile) error
	GetUploadFileByID(ctx context.Context, id uuid.UUID) (*models.UploadFile, error)
	GetUploadFiles(ctx context.Context, uploadID uuid.UUID) ([]models.UploadFile, error)
	UpdateUploadFile(ctx context.Context, uploadFile *models.UploadFile) error
	DeleteUploadFile(ctx context.Context, id uuid.UUID) error
	DeleteUploadFilesByUploadID(ctx context.Context, uploadID uuid.UUID) error
	GetUploadedSize(ctx context.Context, uploadID uuid.UUID) (int64, error)

	// Search operations
	SearchAudioBooks(ctx context.Context, query string, limit, offset int, language *string, isPublic *bool) ([]models.AudioBook, int, error)
	SearchAudioBooksByVector(ctx context.Context, embedding []float64, embeddingType models.EmbeddingType, limit, offset int) ([]models.AudioBook, []float64, error)
	SearchAudioBooksByTags(ctx context.Context, tagNames []string, limit, offset int) ([]models.AudioBook, int, error)

	// Utility operations
	GetAudioBookStats(ctx context.Context) (*AudioBookStats, error)
	GetUserAudioBookStats(ctx context.Context, userID uuid.UUID) (*UserAudioBookStats, error)
	CleanupOrphanedData(ctx context.Context) error
}

// AudioBookStats represents statistics about audio books
type AudioBookStats struct {
	TotalAudioBooks      int   `json:"total_audiobooks"`
	CompletedAudioBooks  int   `json:"completed_audiobooks"`
	PendingAudioBooks    int   `json:"pending_audiobooks"`
	ProcessingAudioBooks int   `json:"processing_audiobooks"`
	FailedAudioBooks     int   `json:"failed_audiobooks"`
	TotalDuration        int   `json:"total_duration_seconds"`
	TotalFileSize        int64 `json:"total_file_size_bytes"`
}

// UserAudioBookStats represents statistics about a user's audio books
type UserAudioBookStats struct {
	UserID               uuid.UUID `json:"user_id"`
	TotalAudioBooks      int       `json:"total_audiobooks"`
	CompletedAudioBooks  int       `json:"completed_audiobooks"`
	PendingAudioBooks    int       `json:"pending_audiobooks"`
	ProcessingAudioBooks int       `json:"processing_audiobooks"`
	FailedAudioBooks     int       `json:"failed_audiobooks"`
	TotalDuration        int       `json:"total_duration_seconds"`
	TotalFileSize        int64     `json:"total_file_size_bytes"`
}

// SearchFilters represents filters for audio book search
type SearchFilters struct {
	Language    *string                 `json:"language,omitempty"`
	IsPublic    *bool                   `json:"is_public,omitempty"`
	Status      *models.AudioBookStatus `json:"status,omitempty"`
	Author      *string                 `json:"author,omitempty"`
	Tags        []string                `json:"tags,omitempty"`
	MinDuration *int                    `json:"min_duration_seconds,omitempty"`
	MaxDuration *int                    `json:"max_duration_seconds,omitempty"`
}

// SearchOptions represents options for audio book search
type SearchOptions struct {
	Limit     int            `json:"limit"`
	Offset    int            `json:"offset"`
	SortBy    string         `json:"sort_by"`    // title, author, created_at, duration
	SortOrder string         `json:"sort_order"` // asc, desc
	Filters   *SearchFilters `json:"filters,omitempty"`
}
