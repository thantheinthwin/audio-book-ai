package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"audio-book-ai/worker/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// DatabaseService handles database operations
type DatabaseService struct {
	db *sql.DB
}

// NewDatabaseService creates a new database service
func NewDatabaseService(dbURL string) (*DatabaseService, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &DatabaseService{db: db}, nil
}

// Close closes the database connection
func (d *DatabaseService) Close() error {
	return d.db.Close()
}

// GetPendingJobs retrieves pending AI processing jobs from database
func (d *DatabaseService) GetPendingJobs(limit int) ([]models.Job, error) {
	query := `
		SELECT id, audiobook_id, job_type, status, created_at, started_at, completed_at, error_message
		FROM processing_jobs
		WHERE job_type IN ($1, $2, $3)
		AND status = $4
		ORDER BY created_at ASC
		LIMIT $5
	`

	rows, err := d.db.Query(query,
		models.JobTypeSummarize,
		models.JobTypeTag,
		models.JobTypeEmbed,
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
			&job.ID, &job.AudiobookID, &job.JobType, &job.Status,
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

	_, err := d.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update job status: %v", err)
	}

	log.Printf("Updated job %s status to %s", jobID, status)
	return nil
}

// GetTranscript retrieves the transcript for an audiobook
func (d *DatabaseService) GetTranscript(audiobookID uuid.UUID) (string, error) {
	query := `
		SELECT content
		FROM transcripts
		WHERE audiobook_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var content string
	err := d.db.QueryRow(query, audiobookID).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no transcript found for audiobook %s", audiobookID)
		}
		return "", fmt.Errorf("failed to get transcript: %v", err)
	}

	return content, nil
}

// SaveAIOutput saves AI processing output to the database
func (d *DatabaseService) SaveAIOutput(output *models.AIOutput) error {
	contentJSON, err := json.Marshal(output.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %v", err)
	}

	query := `
		INSERT INTO ai_outputs (
			id, audiobook_id, output_type, content, model_used, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (audiobook_id, output_type) DO UPDATE SET
			content = EXCLUDED.content,
			model_used = EXCLUDED.model_used,
			created_at = EXCLUDED.created_at
	`

	_, err = d.db.Exec(query,
		uuid.New(),
		output.AudiobookID,
		output.OutputType,
		contentJSON,
		output.ModelUsed,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save AI output: %v", err)
	}

	log.Printf("Saved %s output for audiobook %s", output.OutputType, output.AudiobookID)
	return nil
}
