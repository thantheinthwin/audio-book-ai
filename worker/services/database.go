package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"audio-book-ai/worker/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseService handles database operations
type DatabaseService struct {
	pool *pgxpool.Pool
}

// NewDatabaseService creates a new database service
func NewDatabaseService(dbURL string) (*DatabaseService, error) {
	// First, test the connection with pgx.Connect to handle early fallback gracefully
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	// Test database connection
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Now create the connection pool
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %v", err)
	}

	// Test the pool connection as well
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping connection pool: %v", err)
	}

	fmt.Println("Database service initialized successfully")
	return &DatabaseService{pool: pool}, nil
}

// Close closes the database connection
func (d *DatabaseService) Close() error {
	d.pool.Close()
	return nil
}

// GetPendingJobs retrieves pending AI processing jobs from database
func (d *DatabaseService) GetPendingJobs(limit int) ([]models.Job, error) {
	query := `
		SELECT id, audiobook_id, chapter_id, job_type, status, retry_count, max_retries, created_at, started_at, completed_at, error_message
		FROM processing_jobs
		WHERE job_type IN ($1, $2)
		AND status = $3
		ORDER BY created_at ASC
		LIMIT $4
	`

	rows, err := d.pool.Query(context.Background(), query,
		models.JobTypeEmbed,
		models.JobTypeSummarize,
		models.JobStatusPending,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending jobs: %v", err)
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var job models.Job
		if err := rows.Scan(
			&job.ID, &job.AudiobookID, &job.ChapterID, &job.JobType, &job.Status,
			&job.RetryCount, &job.MaxRetries, &job.CreatedAt, &job.StartedAt, &job.CompletedAt, &job.ErrorMessage,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job: %v", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// UpdateJobStatus updates the status of a job in the database
func (d *DatabaseService) UpdateJobStatus(jobID uuid.UUID, status string, errorMessage *string) error {
	var query string
	var args []interface{}

	if status == models.JobStatusRunning {
		query = `
			UPDATE processing_jobs 
			SET status = $1, started_at = NOW()
			WHERE id = $2
		`
		args = []interface{}{status, jobID}
	} else if status == models.JobStatusCompleted || status == models.JobStatusFailed {
		query = `
			UPDATE processing_jobs 
			SET status = $1, completed_at = NOW(), error_message = $2
			WHERE id = $3
		`
		args = []interface{}{status, errorMessage, jobID}
	}

	_, err := d.pool.Exec(context.Background(), query, args...)
	if err != nil {
		return fmt.Errorf("failed to update job status: %v", err)
	}

	log.Printf("Updated job %s status to %s", jobID, status)
	return nil
}

// SaveAIOutput saves AI processing output to the database
func (d *DatabaseService) SaveAIOutput(output *models.AIOutput) error {
	// First, try to delete any existing output for this audiobook and output type
	deleteQuery := `
		DELETE FROM ai_outputs 
		WHERE audiobook_id = $1 AND output_type = $2
	`
	_, err := d.pool.Exec(context.Background(), deleteQuery, output.AudiobookID, output.OutputType)
	if err != nil {
		return fmt.Errorf("failed to delete existing AI output: %v", err)
	}

	// Then insert the new output
	// Pass content directly to PostgreSQL - pgx will handle JSONB conversion automatically
	insertQuery := `
		INSERT INTO ai_outputs (
			id, audiobook_id, output_type, content, model_used, created_at, processing_time_seconds
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = d.pool.Exec(context.Background(), insertQuery,
		uuid.New(),
		output.AudiobookID,
		output.OutputType,
		output.Content, // Pass content directly without marshaling
		output.ModelUsed,
		time.Now(),
		output.ProcessingTimeSeconds,
	)

	if err != nil {
		return fmt.Errorf("failed to save AI output: %v", err)
	}

	log.Printf("Saved %s output for audiobook %s", output.OutputType, output.AudiobookID)
	return nil
}

// GetChapterTranscripts retrieves all chapter transcripts for an audiobook
func (d *DatabaseService) GetChapterTranscripts(audiobookID uuid.UUID) ([]models.ChapterTranscript, error) {
	query := `
		SELECT ct.id, ct.chapter_id, ct.audiobook_id, ct.content, ct.segments, 
		       ct.language, ct.confidence_score, ct.processing_time_seconds, ct.created_at
		FROM chapter_transcripts ct
		JOIN chapters c ON ct.chapter_id = c.id
		WHERE ct.audiobook_id = $1
		ORDER BY c.chapter_number ASC
	`

	rows, err := d.pool.Query(context.Background(), query, audiobookID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chapter transcripts: %v", err)
	}
	defer rows.Close()

	var transcripts []models.ChapterTranscript
	for rows.Next() {
		var transcript models.ChapterTranscript
		var segmentsJSON []byte

		if err := rows.Scan(
			&transcript.ID, &transcript.ChapterID, &transcript.AudiobookID, &transcript.Content,
			&segmentsJSON, &transcript.Language, &transcript.ConfidenceScore,
			&transcript.ProcessingTimeSeconds, &transcript.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan chapter transcript: %v", err)
		}

		// Parse segments JSON into the proper slice of Segment structs
		if segmentsJSON != nil {
			if err := json.Unmarshal(segmentsJSON, &transcript.Segments); err != nil {
				return nil, fmt.Errorf("failed to unmarshal segments for transcript %s: %v", transcript.ID, err)
			}
		}

		transcripts = append(transcripts, transcript)
	}

	return transcripts, nil
}

// GetAllTags retrieves all available tags from the database
func (d *DatabaseService) GetAllTags() ([]string, error) {
	query := `
		SELECT name
		FROM tags
		ORDER BY name ASC
	`

	rows, err := d.pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %v", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tagName string
		if err := rows.Scan(&tagName); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %v", err)
		}
		tags = append(tags, tagName)
	}

	return tags, nil
}

// GetChapter1Transcript gets the transcript for chapter 1 of an audiobook
func (d *DatabaseService) GetChapter1Transcript(audiobookID string) (*models.ChapterTranscript, error) {
	query := `
		SELECT ct.id, ct.chapter_id, ct.audiobook_id, ct.content, ct.segments, 
		       ct.language, ct.confidence_score, ct.processing_time_seconds, ct.created_at
		FROM chapter_transcripts ct
		JOIN chapters c ON ct.chapter_id = c.id
		WHERE c.audiobook_id = $1 AND c.chapter_number = 1
		LIMIT 1
	`

	var transcript models.ChapterTranscript
	var segmentsJSON []byte

	err := d.pool.QueryRow(context.Background(), query, audiobookID).Scan(
		&transcript.ID,
		&transcript.ChapterID,
		&transcript.AudiobookID,
		&transcript.Content,
		&segmentsJSON,
		&transcript.Language,
		&transcript.ConfidenceScore,
		&transcript.ProcessingTimeSeconds,
		&transcript.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No transcript found
		}
		return nil, fmt.Errorf("failed to get chapter 1 transcript: %v", err)
	}

	// Parse segments JSON into the proper slice of Segment structs
	if segmentsJSON != nil {
		if err := json.Unmarshal(segmentsJSON, &transcript.Segments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal segments: %v", err)
		}
	}

	return &transcript, nil
}

// IncrementRetryCount increments the retry count for a processing job
func (d *DatabaseService) IncrementRetryCount(jobID uuid.UUID) error {
	query := `
		UPDATE processing_jobs 
		SET retry_count = retry_count + 1
		WHERE id = $1
	`

	result, err := d.pool.Exec(context.Background(), query, jobID)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found: %s", jobID)
	}

	log.Printf("Incremented retry count for job %s", jobID)
	return nil
}

// UpdateAudioBookSummaryAndTags updates the audiobook summary and tags directly
func (d *DatabaseService) UpdateAudioBookSummaryAndTags(audiobookID uuid.UUID, summary string, tags []string) error {
	query := `
		UPDATE audiobooks 
		SET summary = $2, tags = $3, updated_at = $4
		WHERE id = $1
	`

	now := time.Now()
	result, err := d.pool.Exec(context.Background(), query, audiobookID, summary, tags, now)

	if err != nil {
		return fmt.Errorf("failed to update audiobook summary and tags: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("audiobook not found or no changes made")
	}

	log.Printf("Updated audiobook %s with summary and tags", audiobookID)
	return nil
}
