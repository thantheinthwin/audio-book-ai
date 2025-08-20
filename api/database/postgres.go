package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"audio-book-ai/api/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements the Repository interface with PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(databaseURL string) (*PostgresRepository, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

// Close closes the database connection pool
func (p *PostgresRepository) Close() error {
	p.pool.Close()
	return nil
}

// Upload operations
func (p *PostgresRepository) CreateUpload(ctx context.Context, upload *models.Upload) error {
	query := `
		INSERT INTO uploads (id, user_id, upload_type, status, total_files, uploaded_files, total_size_bytes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	_, err := p.pool.Exec(ctx, query,
		upload.ID,
		upload.UserID,
		upload.UploadType,
		upload.Status,
		upload.TotalFiles,
		upload.UploadedFiles,
		upload.TotalSize,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create upload: %w", err)
	}

	upload.CreatedAt = now
	upload.UpdatedAt = now
	return nil
}

func (p *PostgresRepository) GetUploadByID(ctx context.Context, id uuid.UUID) (*models.Upload, error) {
	query := `
		SELECT id, user_id, upload_type, status, total_files, uploaded_files, total_size_bytes, created_at, updated_at
		FROM uploads
		WHERE id = $1
	`

	var upload models.Upload
	err := p.pool.QueryRow(ctx, query, id).Scan(
		&upload.ID,
		&upload.UserID,
		&upload.UploadType,
		&upload.Status,
		&upload.TotalFiles,
		&upload.UploadedFiles,
		&upload.TotalSize,
		&upload.CreatedAt,
		&upload.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get upload: %w", err)
	}

	return &upload, nil
}

func (p *PostgresRepository) GetUploadsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Upload, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM uploads WHERE user_id = $1`
	var total int
	err := p.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count uploads: %w", err)
	}

	// Get uploads with pagination
	query := `
		SELECT id, user_id, upload_type, status, total_files, uploaded_files, total_size_bytes, created_at, updated_at
		FROM uploads
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := p.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query uploads: %w", err)
	}
	defer rows.Close()

	var uploads []models.Upload
	for rows.Next() {
		var upload models.Upload
		err := rows.Scan(
			&upload.ID,
			&upload.UserID,
			&upload.UploadType,
			&upload.Status,
			&upload.TotalFiles,
			&upload.UploadedFiles,
			&upload.TotalSize,
			&upload.CreatedAt,
			&upload.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan upload: %w", err)
		}
		uploads = append(uploads, upload)
	}

	return uploads, total, nil
}

func (p *PostgresRepository) UpdateUpload(ctx context.Context, upload *models.Upload) error {
	query := `
		UPDATE uploads
		SET upload_type = $2, status = $3, total_files = $4, uploaded_files = $5, total_size_bytes = $6, updated_at = $7
		WHERE id = $1
	`

	now := time.Now()
	result, err := p.pool.Exec(ctx, query,
		upload.ID,
		upload.UploadType,
		upload.Status,
		upload.TotalFiles,
		upload.UploadedFiles,
		upload.TotalSize,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to update upload: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	upload.UpdatedAt = now
	return nil
}

func (p *PostgresRepository) DeleteUpload(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM uploads WHERE id = $1`

	result, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete upload: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Upload File operations
func (p *PostgresRepository) CreateUploadFile(ctx context.Context, uploadFile *models.UploadFile) error {
	query := `
		INSERT INTO upload_files (id, upload_id, file_name, file_size_bytes, mime_type, file_path, chapter_number, chapter_title, status, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	_, err := p.pool.Exec(ctx, query,
		uploadFile.ID,
		uploadFile.UploadID,
		uploadFile.FileName,
		uploadFile.FileSize,
		uploadFile.MimeType,
		uploadFile.FilePath,
		uploadFile.ChapterNumber,
		uploadFile.ChapterTitle,
		uploadFile.Status,
		uploadFile.Error,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create upload file: %w", err)
	}

	uploadFile.CreatedAt = now
	return nil
}

func (p *PostgresRepository) GetUploadFileByID(ctx context.Context, id uuid.UUID) (*models.UploadFile, error) {
	query := `
		SELECT id, upload_id, file_name, file_size_bytes, mime_type, file_path, chapter_number, chapter_title, status, error, created_at
		FROM upload_files
		WHERE id = $1
	`

	var uploadFile models.UploadFile
	err := p.pool.QueryRow(ctx, query, id).Scan(
		&uploadFile.ID,
		&uploadFile.UploadID,
		&uploadFile.FileName,
		&uploadFile.FileSize,
		&uploadFile.MimeType,
		&uploadFile.FilePath,
		&uploadFile.ChapterNumber,
		&uploadFile.ChapterTitle,
		&uploadFile.Status,
		&uploadFile.Error,
		&uploadFile.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get upload file: %w", err)
	}

	return &uploadFile, nil
}

func (p *PostgresRepository) GetUploadFiles(ctx context.Context, uploadID uuid.UUID) ([]models.UploadFile, error) {
	query := `
		SELECT id, upload_id, file_name, file_size_bytes, mime_type, file_path, chapter_number, chapter_title, status, error, created_at
		FROM upload_files
		WHERE upload_id = $1
		ORDER BY chapter_number NULLS LAST, created_at
	`

	rows, err := p.pool.Query(ctx, query, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to query upload files: %w", err)
	}
	defer rows.Close()

	var files []models.UploadFile
	for rows.Next() {
		var file models.UploadFile
		err := rows.Scan(
			&file.ID,
			&file.UploadID,
			&file.FileName,
			&file.FileSize,
			&file.MimeType,
			&file.FilePath,
			&file.ChapterNumber,
			&file.ChapterTitle,
			&file.Status,
			&file.Error,
			&file.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan upload file: %w", err)
		}
		files = append(files, file)
	}

	return files, nil
}

func (p *PostgresRepository) UpdateUploadFile(ctx context.Context, uploadFile *models.UploadFile) error {
	query := `
		UPDATE upload_files
		SET file_name = $2, file_size_bytes = $3, mime_type = $4, file_path = $5, chapter_number = $6, chapter_title = $7, status = $8, error = $9
		WHERE id = $1
	`

	result, err := p.pool.Exec(ctx, query,
		uploadFile.ID,
		uploadFile.FileName,
		uploadFile.FileSize,
		uploadFile.MimeType,
		uploadFile.FilePath,
		uploadFile.ChapterNumber,
		uploadFile.ChapterTitle,
		uploadFile.Status,
		uploadFile.Error,
	)

	if err != nil {
		return fmt.Errorf("failed to update upload file: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (p *PostgresRepository) DeleteUploadFile(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM upload_files WHERE id = $1`

	result, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete upload file: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (p *PostgresRepository) DeleteUploadFilesByUploadID(ctx context.Context, uploadID uuid.UUID) error {
	query := `DELETE FROM upload_files WHERE upload_id = $1`

	_, err := p.pool.Exec(ctx, query, uploadID)
	if err != nil {
		return fmt.Errorf("failed to delete upload files: %w", err)
	}

	return nil
}

func (p *PostgresRepository) GetUploadedSize(ctx context.Context, uploadID uuid.UUID) (int64, error) {
	query := `SELECT COALESCE(SUM(file_size_bytes), 0) FROM upload_files WHERE upload_id = $1`

	var totalSize int64
	err := p.pool.QueryRow(ctx, query, uploadID).Scan(&totalSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get uploaded size: %w", err)
	}

	return totalSize, nil
}

// Stub implementations for other interface methods (to be implemented as needed)
func (p *PostgresRepository) CreateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
	query := `
		INSERT INTO audiobooks (id, title, author, summary, duration_seconds, file_size_bytes, file_path, file_url, cover_image_url, language, is_public, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	now := time.Now()
	_, err := p.pool.Exec(ctx, query,
		audiobook.ID,
		audiobook.Title,
		audiobook.Author,
		audiobook.Summary,
		audiobook.DurationSeconds,
		audiobook.FileSizeBytes,
		audiobook.FilePath,
		audiobook.FileURL,
		audiobook.CoverImageURL,
		audiobook.Language,
		audiobook.IsPublic,
		audiobook.Status,
		audiobook.CreatedBy,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create audiobook: %w", err)
	}

	audiobook.CreatedAt = now
	audiobook.UpdatedAt = now
	return nil
}

func (p *PostgresRepository) GetAudioBookByID(ctx context.Context, id uuid.UUID) (*models.AudioBook, error) {
	query := `
		SELECT id, title, author, summary, duration_seconds, file_size_bytes, file_path, file_url, cover_image_url, language, is_public, status, created_by, created_at, updated_at
		FROM audiobooks
		WHERE id = $1
	`

	var audiobook models.AudioBook
	err := p.pool.QueryRow(ctx, query, id).Scan(
		&audiobook.ID,
		&audiobook.Title,
		&audiobook.Author,
		&audiobook.Summary,
		&audiobook.DurationSeconds,
		&audiobook.FileSizeBytes,
		&audiobook.FilePath,
		&audiobook.FileURL,
		&audiobook.CoverImageURL,
		&audiobook.Language,
		&audiobook.IsPublic,
		&audiobook.Status,
		&audiobook.CreatedBy,
		&audiobook.CreatedAt,
		&audiobook.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get audiobook: %w", err)
	}

	return &audiobook, nil
}

func (p *PostgresRepository) GetAudioBookWithDetails(ctx context.Context, id uuid.UUID) (*models.AudioBookWithDetails, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) UpdateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
	query := `
		UPDATE audiobooks 
		SET title = $2, author = $3, summary = $4, duration_seconds = $5, file_size_bytes = $6, 
		    file_path = $7, file_url = $8, cover_image_url = $9, language = $10, is_public = $11, 
		    status = $12, updated_at = $13
		WHERE id = $1
	`

	now := time.Now()
	result, err := p.pool.Exec(ctx, query,
		audiobook.ID,
		audiobook.Title,
		audiobook.Author,
		audiobook.Summary,
		audiobook.DurationSeconds,
		audiobook.FileSizeBytes,
		audiobook.FilePath,
		audiobook.FileURL,
		audiobook.CoverImageURL,
		audiobook.Language,
		audiobook.IsPublic,
		audiobook.Status,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to update audiobook: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	audiobook.UpdatedAt = now
	return nil
}

func (p *PostgresRepository) DeleteAudioBook(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) ListAudioBooks(ctx context.Context, limit, offset int, isPublic *bool) ([]models.AudioBook, int, error) {
	return []models.AudioBook{}, 0, nil
}

func (p *PostgresRepository) GetAudioBooksByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error) {
	return []models.AudioBook{}, 0, nil
}

func (p *PostgresRepository) UpdateAudioBookStatus(ctx context.Context, id uuid.UUID, status models.AudioBookStatus) error {
	query := `
		UPDATE audiobooks 
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	now := time.Now()
	result, err := p.pool.Exec(ctx, query, id, status, now)

	if err != nil {
		return fmt.Errorf("failed to update audiobook status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// UpdateAudioBookSummary updates the audiobook summary from AI output
func (p *PostgresRepository) UpdateAudioBookSummary(ctx context.Context, audiobookID uuid.UUID, summary string) error {
	query := `
		UPDATE audiobooks 
		SET summary = $2, updated_at = $3
		WHERE id = $1
	`

	now := time.Now()
	result, err := p.pool.Exec(ctx, query, audiobookID, summary, now)

	if err != nil {
		return fmt.Errorf("failed to update audiobook summary: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// CheckAndUpdateAudioBookStatus checks if all jobs are completed and updates audiobook status accordingly
func (p *PostgresRepository) CheckAndUpdateAudioBookStatus(ctx context.Context, audiobookID uuid.UUID) error {
	// Get all jobs for this audiobook
	jobs, err := p.GetProcessingJobsByAudioBookID(ctx, audiobookID)
	if err != nil {
		return fmt.Errorf("failed to get processing jobs: %w", err)
	}

	if len(jobs) == 0 {
		return nil // No jobs to check
	}

	// Check if all jobs are completed
	allCompleted := true
	hasFailed := false
	hasSummary := false

	for _, job := range jobs {
		if job.Status == models.JobStatusFailed {
			hasFailed = true
			break
		}
		if job.Status != models.JobStatusCompleted {
			allCompleted = false
		}
		if job.JobType == models.JobTypeSummarize && job.Status == models.JobStatusCompleted {
			hasSummary = true
		}
	}

	// Update audiobook status based on job status
	var newStatus models.AudioBookStatus
	if hasFailed {
		newStatus = models.StatusFailed
	} else if allCompleted {
		newStatus = models.StatusCompleted

		// If summary job completed, update the audiobook summary
		if hasSummary {
			summaryOutput, err := p.GetAIOutputByType(ctx, audiobookID, models.OutputTypeSummary)
			if err == nil && summaryOutput != nil {
				// Extract summary from JSON content
				var summaryData map[string]interface{}
				if err := json.Unmarshal(summaryOutput.Content, &summaryData); err == nil {
					if summaryText, ok := summaryData["summary"].(string); ok {
						p.UpdateAudioBookSummary(ctx, audiobookID, summaryText)
					}
				}
			}
		}
	} else {
		newStatus = models.StatusProcessing
	}

	return p.UpdateAudioBookStatus(ctx, audiobookID, newStatus)
}

func (p *PostgresRepository) CreateChapter(ctx context.Context, chapter *models.Chapter) error {
	query := `
		INSERT INTO chapters (id, audiobook_id, chapter_number, title, start_time_seconds, end_time_seconds, duration_seconds, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	_, err := p.pool.Exec(ctx, query,
		chapter.ID,
		chapter.AudiobookID,
		chapter.ChapterNumber,
		chapter.Title,
		chapter.StartTime,
		chapter.EndTime,
		chapter.DurationSeconds,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create chapter: %w", err)
	}

	chapter.CreatedAt = now
	return nil
}

func (p *PostgresRepository) GetChapterByID(ctx context.Context, id uuid.UUID) (*models.Chapter, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) GetChaptersByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Chapter, error) {
	return []models.Chapter{}, nil
}

func (p *PostgresRepository) GetFirstChapterByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Chapter, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) UpdateChapter(ctx context.Context, chapter *models.Chapter) error {
	return nil
}

func (p *PostgresRepository) DeleteChapter(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) DeleteChaptersByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) CreateTranscript(ctx context.Context, transcript *models.Transcript) error {
	query := `
		INSERT INTO transcripts (id, audiobook_id, content, segments, language, confidence_score, processing_time_seconds, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	_, err := p.pool.Exec(ctx, query,
		transcript.ID,
		transcript.AudiobookID,
		transcript.Content,
		transcript.Segments,
		transcript.Language,
		transcript.ConfidenceScore,
		transcript.ProcessingTimeSeconds,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create transcript: %w", err)
	}

	transcript.CreatedAt = now
	return nil
}

func (p *PostgresRepository) GetTranscriptByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Transcript, error) {
	query := `
		SELECT id, audiobook_id, content, segments, language, confidence_score, processing_time_seconds, created_at
		FROM transcripts
		WHERE audiobook_id = $1
	`

	var transcript models.Transcript
	err := p.pool.QueryRow(ctx, query, audiobookID).Scan(
		&transcript.ID,
		&transcript.AudiobookID,
		&transcript.Content,
		&transcript.Segments,
		&transcript.Language,
		&transcript.ConfidenceScore,
		&transcript.ProcessingTimeSeconds,
		&transcript.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transcript: %w", err)
	}

	return &transcript, nil
}

func (p *PostgresRepository) UpdateTranscript(ctx context.Context, transcript *models.Transcript) error {
	return nil
}

func (p *PostgresRepository) DeleteTranscript(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) CreateChapterTranscript(ctx context.Context, transcript *models.ChapterTranscript) error {
	return nil
}

func (p *PostgresRepository) GetChapterTranscriptByChapterID(ctx context.Context, chapterID uuid.UUID) (*models.ChapterTranscript, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) GetChapterTranscriptsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ChapterTranscript, error) {
	return []models.ChapterTranscript{}, nil
}

func (p *PostgresRepository) UpdateChapterTranscript(ctx context.Context, transcript *models.ChapterTranscript) error {
	return nil
}

func (p *PostgresRepository) DeleteChapterTranscript(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) DeleteChapterTranscriptsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) CreateAIOutput(ctx context.Context, output *models.AIOutput) error {
	query := `
		INSERT INTO ai_outputs (id, audiobook_id, output_type, content, model_used, processing_time_seconds, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	_, err := p.pool.Exec(ctx, query,
		output.ID,
		output.AudiobookID,
		output.OutputType,
		output.Content,
		output.ModelUsed,
		output.ProcessingTimeSeconds,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create AI output: %w", err)
	}

	output.CreatedAt = now
	return nil
}

func (p *PostgresRepository) GetAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AIOutput, error) {
	query := `
		SELECT id, audiobook_id, output_type, content, model_used, processing_time_seconds, created_at
		FROM ai_outputs
		WHERE audiobook_id = $1
		ORDER BY created_at DESC
	`

	rows, err := p.pool.Query(ctx, query, audiobookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI outputs: %w", err)
	}
	defer rows.Close()

	var outputs []models.AIOutput
	for rows.Next() {
		var output models.AIOutput
		err := rows.Scan(
			&output.ID,
			&output.AudiobookID,
			&output.OutputType,
			&output.Content,
			&output.ModelUsed,
			&output.ProcessingTimeSeconds,
			&output.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan AI output: %w", err)
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

func (p *PostgresRepository) GetAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.AIOutput, error) {
	query := `
		SELECT id, audiobook_id, output_type, content, model_used, processing_time_seconds, created_at
		FROM ai_outputs
		WHERE audiobook_id = $1 AND output_type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var output models.AIOutput
	err := p.pool.QueryRow(ctx, query, audiobookID, outputType).Scan(
		&output.ID,
		&output.AudiobookID,
		&output.OutputType,
		&output.Content,
		&output.ModelUsed,
		&output.ProcessingTimeSeconds,
		&output.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get AI output: %w", err)
	}

	return &output, nil
}

func (p *PostgresRepository) UpdateAIOutput(ctx context.Context, output *models.AIOutput) error {
	return nil
}

func (p *PostgresRepository) DeleteAIOutput(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) CreateChapterAIOutput(ctx context.Context, output *models.ChapterAIOutput) error {
	return nil
}

func (p *PostgresRepository) GetChapterAIOutputsByChapterID(ctx context.Context, chapterID uuid.UUID) ([]models.ChapterAIOutput, error) {
	return []models.ChapterAIOutput{}, nil
}

func (p *PostgresRepository) GetChapterAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ChapterAIOutput, error) {
	return []models.ChapterAIOutput{}, nil
}

func (p *PostgresRepository) GetFirstChapterAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.ChapterAIOutput, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) UpdateChapterAIOutput(ctx context.Context, output *models.ChapterAIOutput) error {
	return nil
}

func (p *PostgresRepository) DeleteChapterAIOutput(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) DeleteChapterAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) CreateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	query := `
		INSERT INTO processing_jobs (id, audiobook_id, job_type, status, redis_job_id, error_message, started_at, completed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	_, err := p.pool.Exec(ctx, query,
		job.ID,
		job.AudiobookID,
		job.JobType,
		job.Status,
		job.RedisJobID,
		job.ErrorMessage,
		job.StartedAt,
		job.CompletedAt,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create processing job: %w", err)
	}

	job.CreatedAt = now
	return nil
}

func (p *PostgresRepository) GetProcessingJobsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ProcessingJob, error) {
	query := `
		SELECT id, audiobook_id, job_type, status, redis_job_id, error_message, started_at, completed_at, created_at
		FROM processing_jobs
		WHERE audiobook_id = $1
		ORDER BY created_at DESC
	`

	rows, err := p.pool.Query(ctx, query, audiobookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get processing jobs: %w", err)
	}
	defer rows.Close()

	var jobs []models.ProcessingJob
	for rows.Next() {
		var job models.ProcessingJob
		err := rows.Scan(
			&job.ID,
			&job.AudiobookID,
			&job.JobType,
			&job.Status,
			&job.RedisJobID,
			&job.ErrorMessage,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan processing job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (p *PostgresRepository) GetPendingJobs(ctx context.Context, jobType models.JobType, limit int) ([]models.ProcessingJob, error) {
	query := `
		SELECT id, audiobook_id, job_type, status, redis_job_id, error_message, started_at, completed_at, created_at
		FROM processing_jobs
		WHERE job_type = $1 AND status = $2
		ORDER BY created_at ASC
		LIMIT $3
	`

	rows, err := p.pool.Query(ctx, query, jobType, models.JobStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []models.ProcessingJob
	for rows.Next() {
		var job models.ProcessingJob
		err := rows.Scan(
			&job.ID,
			&job.AudiobookID,
			&job.JobType,
			&job.Status,
			&job.RedisJobID,
			&job.ErrorMessage,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan processing job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (p *PostgresRepository) GetJobsByStatus(ctx context.Context, status models.JobStatus, limit int) ([]models.ProcessingJob, error) {
	query := `
		SELECT id, audiobook_id, job_type, status, redis_job_id, error_message, started_at, completed_at, created_at
		FROM processing_jobs
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := p.pool.Query(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by status: %w", err)
	}
	defer rows.Close()

	var jobs []models.ProcessingJob
	for rows.Next() {
		var job models.ProcessingJob
		err := rows.Scan(
			&job.ID,
			&job.AudiobookID,
			&job.JobType,
			&job.Status,
			&job.RedisJobID,
			&job.ErrorMessage,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan processing job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (p *PostgresRepository) GetProcessingJobByID(ctx context.Context, id uuid.UUID) (*models.ProcessingJob, error) {
	query := `
		SELECT id, audiobook_id, job_type, status, redis_job_id, error_message, started_at, completed_at, created_at
		FROM processing_jobs
		WHERE id = $1
	`

	var job models.ProcessingJob
	err := p.pool.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.AudiobookID,
		&job.JobType,
		&job.Status,
		&job.RedisJobID,
		&job.ErrorMessage,
		&job.StartedAt,
		&job.CompletedAt,
		&job.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get processing job: %w", err)
	}

	return &job, nil
}

func (p *PostgresRepository) UpdateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	query := `
		UPDATE processing_jobs 
		SET status = $3, redis_job_id = $4, error_message = $5, started_at = $6, completed_at = $7
		WHERE id = $1 AND audiobook_id = $2
	`

	result, err := p.pool.Exec(ctx, query,
		job.ID,
		job.AudiobookID,
		job.Status,
		job.RedisJobID,
		job.ErrorMessage,
		job.StartedAt,
		job.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update processing job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (p *PostgresRepository) CreateTag(ctx context.Context, tag *models.Tag) error {
	return nil
}

func (p *PostgresRepository) GetTagByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) GetTagByName(ctx context.Context, name string) (*models.Tag, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) GetTagsByCategory(ctx context.Context, category string) ([]models.Tag, error) {
	return []models.Tag{}, nil
}

func (p *PostgresRepository) UpdateTag(ctx context.Context, tag *models.Tag) error {
	return nil
}

func (p *PostgresRepository) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) ListTags(ctx context.Context, limit, offset int) ([]models.Tag, int, error) {
	return []models.Tag{}, 0, nil
}

func (p *PostgresRepository) CreateAudioBookTag(ctx context.Context, audiobookTag *models.AudioBookTag) error {
	return nil
}

func (p *PostgresRepository) GetTagsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Tag, error) {
	return []models.Tag{}, nil
}

func (p *PostgresRepository) GetAudioBooksByTagID(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error) {
	return []models.AudioBook{}, 0, nil
}

func (p *PostgresRepository) DeleteAudioBookTag(ctx context.Context, audiobookID, tagID uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) DeleteAllAudioBookTags(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) CreateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error {
	return nil
}

func (p *PostgresRepository) GetEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AudioBookEmbedding, error) {
	return []models.AudioBookEmbedding{}, nil
}

func (p *PostgresRepository) GetEmbeddingByType(ctx context.Context, audiobookID uuid.UUID, embeddingType models.EmbeddingType) (*models.AudioBookEmbedding, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) UpdateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error {
	return nil
}

func (p *PostgresRepository) DeleteAudioBookEmbedding(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) DeleteEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (p *PostgresRepository) SearchAudioBooks(ctx context.Context, query string, limit, offset int, language *string, isPublic *bool) ([]models.AudioBook, int, error) {
	return []models.AudioBook{}, 0, nil
}

func (p *PostgresRepository) SearchAudioBooksByVector(ctx context.Context, embedding []float64, embeddingType models.EmbeddingType, limit, offset int) ([]models.AudioBook, []float64, error) {
	return []models.AudioBook{}, []float64{}, nil
}

func (p *PostgresRepository) SearchAudioBooksByTags(ctx context.Context, tagNames []string, limit, offset int) ([]models.AudioBook, int, error) {
	return []models.AudioBook{}, 0, nil
}

func (p *PostgresRepository) GetAudioBookStats(ctx context.Context) (*AudioBookStats, error) {
	return &AudioBookStats{}, nil
}

func (p *PostgresRepository) GetUserAudioBookStats(ctx context.Context, userID uuid.UUID) (*UserAudioBookStats, error) {
	return &UserAudioBookStats{UserID: userID}, nil
}

func (p *PostgresRepository) CleanupOrphanedData(ctx context.Context) error {
	return nil
}
