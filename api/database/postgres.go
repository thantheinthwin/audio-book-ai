package database

import (
	"context"
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
	return nil
}

func (p *PostgresRepository) GetAudioBookByID(ctx context.Context, id uuid.UUID) (*models.AudioBook, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) GetAudioBookWithDetails(ctx context.Context, id uuid.UUID) (*models.AudioBookWithDetails, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) UpdateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
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
	return nil
}

func (p *PostgresRepository) CreateChapter(ctx context.Context, chapter *models.Chapter) error {
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
	return nil
}

func (p *PostgresRepository) GetTranscriptByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Transcript, error) {
	return nil, ErrNotFound
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
	return nil
}

func (p *PostgresRepository) GetAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AIOutput, error) {
	return []models.AIOutput{}, nil
}

func (p *PostgresRepository) GetAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.AIOutput, error) {
	return nil, ErrNotFound
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
	return nil
}

func (p *PostgresRepository) GetProcessingJobsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ProcessingJob, error) {
	return []models.ProcessingJob{}, nil
}

func (p *PostgresRepository) GetProcessingJobByID(ctx context.Context, id uuid.UUID) (*models.ProcessingJob, error) {
	return nil, ErrNotFound
}

func (p *PostgresRepository) UpdateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	return nil
}

func (p *PostgresRepository) GetPendingJobs(ctx context.Context, jobType models.JobType, limit int) ([]models.ProcessingJob, error) {
	return []models.ProcessingJob{}, nil
}

func (p *PostgresRepository) GetJobsByStatus(ctx context.Context, status models.JobStatus, limit int) ([]models.ProcessingJob, error) {
	return []models.ProcessingJob{}, nil
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
