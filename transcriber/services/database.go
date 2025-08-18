package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"audio-book-ai/transcriber/models"

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

// GetPendingJobs retrieves pending transcription jobs from database
func (d *DatabaseService) GetPendingJobs(limit int) ([]models.Job, error) {
	query := `
		SELECT pj.id, pj.audiobook_id, pj.job_type, pj.status, ab.file_path, ab.language,
		       pj.created_at, pj.started_at, pj.completed_at, pj.error_message
		FROM processing_jobs pj
		JOIN audiobooks ab ON pj.audiobook_id = ab.id
		WHERE pj.job_type = $1 
		AND pj.status = $2
		ORDER BY pj.created_at ASC
		LIMIT $3
	`

	rows, err := d.pool.Query(context.Background(), query, models.JobTypeTranscribe, models.JobStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending jobs: %v", err)
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var job models.Job
		if err := rows.Scan(
			&job.ID, &job.AudiobookID, &job.JobType, &job.Status, &job.FilePath, &job.Language,
			&job.CreatedAt, &job.StartedAt, &job.CompletedAt, &job.ErrorMessage,
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

// SaveTranscript saves the transcript to the database
func (d *DatabaseService) SaveTranscript(transcript *models.Transcript) error {
	segmentsJSON, err := json.Marshal(transcript.Segments)
	if err != nil {
		return fmt.Errorf("failed to marshal segments: %v", err)
	}

	query := `
		INSERT INTO transcripts (
			id, audiobook_id, content, segments, language, 
			confidence_score, processing_time_seconds, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (audiobook_id) DO UPDATE SET
			content = EXCLUDED.content,
			segments = EXCLUDED.segments,
			language = EXCLUDED.language,
			confidence_score = EXCLUDED.confidence_score,
			processing_time_seconds = EXCLUDED.processing_time_seconds,
			created_at = EXCLUDED.created_at
	`

	_, err = d.pool.Exec(context.Background(), query,
		uuid.New(),
		transcript.AudiobookID,
		transcript.Content,
		segmentsJSON,
		transcript.Language,
		transcript.ConfidenceScore,
		transcript.ProcessingTimeSeconds,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save transcript: %v", err)
	}

	log.Printf("Saved transcript for audiobook %s", transcript.AudiobookID)
	return nil
}
