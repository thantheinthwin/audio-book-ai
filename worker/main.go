package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Job struct {
	ID          uuid.UUID `json:"id"`
	AudiobookID uuid.UUID `json:"audiobook_id"`
	JobType     string    `json:"job_type"`
	Status      string    `json:"status"`
}

type Transcript struct {
	Content string `json:"content"`
}

type AIOutput struct {
	ID          uuid.UUID       `json:"id"`
	AudiobookID uuid.UUID       `json:"audiobook_id"`
	OutputType  string          `json:"output_type"`
	Content     json.RawMessage `json:"content"`
	ModelUsed   string          `json:"model_used"`
	CreatedAt   time.Time       `json:"created_at"`
}

type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options"`
}

type OllamaResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type Worker struct {
	db          *sql.DB
	redisClient *redis.Client
	ollamaURL   string
	model       string
}

func NewWorker() (*Worker, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Redis connection
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379/0"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	redisClient := redis.NewClient(opt)

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %v", err)
	}

	// Ollama configuration
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://ollama:11434"
	}

	// Use small model for faster processing
	model := os.Getenv("AI_SUMMARY_MODEL")
	if model == "" {
		model = "llama2:7b" // Small model
	}

	return &Worker{
		db:          db,
		redisClient: redisClient,
		ollamaURL:   ollamaURL,
		model:       model,
	}, nil
}

func (w *Worker) getPendingJobs() ([]Job, error) {
	query := `
		SELECT pj.id, pj.audiobook_id, pj.job_type, pj.status
		FROM processing_jobs pj
		WHERE pj.job_type = 'summarize' 
		AND pj.status = 'pending'
		ORDER BY pj.created_at ASC
		LIMIT 5
	`

	rows, err := w.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending jobs: %v", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.ID, &job.AudiobookID, &job.JobType, &job.Status); err != nil {
			return nil, fmt.Errorf("failed to scan job: %v", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (w *Worker) updateJobStatus(jobID uuid.UUID, status string, errorMessage *string) error {
	var query string
	var args []interface{}

	if status == "running" {
		query = `
			UPDATE processing_jobs 
			SET status = $1, started_at = NOW()
			WHERE id = $2
		`
		args = []interface{}{status, jobID}
	} else if status == "completed" || status == "failed" {
		query = `
			UPDATE processing_jobs 
			SET status = $1, completed_at = NOW(), error_message = $2
			WHERE id = $3
		`
		args = []interface{}{status, errorMessage, jobID}
	}

	_, err := w.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update job status: %v", err)
	}

	log.Printf("Updated job %s status to %s", jobID, status)
	return nil
}

func (w *Worker) getTranscript(audiobookID uuid.UUID) (*Transcript, error) {
	query := `
		SELECT content
		FROM transcripts
		WHERE audiobook_id = $1
	`

	var transcript Transcript
	err := w.db.QueryRow(query, audiobookID).Scan(&transcript.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript: %v", err)
	}

	return &transcript, nil
}

func (w *Worker) generateSummary(text string) (string, error) {
	// Create a simple prompt for summarization
	prompt := fmt.Sprintf(`Please provide a brief summary of the following text in 2-3 sentences:

%s

Summary:`, text)

	// Prepare Ollama request
	req := OllamaRequest{
		Model:  w.model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.3, // Lower temperature for more focused output
			"top_p":       0.9,
			"max_tokens":  200, // Limit output length
		},
	}

	// Convert request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make request to Ollama
	resp, err := http.Post(w.ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to make request to Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode Ollama response: %v", err)
	}

	return ollamaResp.Response, nil
}

func (w *Worker) saveAIOutput(audiobookID uuid.UUID, outputType string, content interface{}) error {
	contentJSON, err := json.Marshal(content)
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

	_, err = w.db.Exec(query, uuid.New(), audiobookID, outputType, contentJSON, w.model, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save AI output: %v", err)
	}

	log.Printf("Saved %s output for audiobook %s", outputType, audiobookID)
	return nil
}

func (w *Worker) processJob(job Job) error {
	log.Printf("Processing summarization job %s for audiobook %s", job.ID, job.AudiobookID)

	// Update job status to running
	if err := w.updateJobStatus(job.ID, "running", nil); err != nil {
		return err
	}

	// Get transcript
	transcript, err := w.getTranscript(job.AudiobookID)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get transcript: %v", err)
		w.updateJobStatus(job.ID, "failed", &errorMsg)
		return err
	}

	// Generate summary
	summary, err := w.generateSummary(transcript.Content)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to generate summary: %v", err)
		w.updateJobStatus(job.ID, "failed", &errorMsg)
		return err
	}

	// Save summary
	summaryData := map[string]interface{}{
		"summary": summary,
		"length":  len(transcript.Content),
	}

	if err := w.saveAIOutput(job.AudiobookID, "summary", summaryData); err != nil {
		errorMsg := fmt.Sprintf("Failed to save summary: %v", err)
		w.updateJobStatus(job.ID, "failed", &errorMsg)
		return err
	}

	// Update job status to completed
	if err := w.updateJobStatus(job.ID, "completed", nil); err != nil {
		return err
	}

	log.Printf("Successfully processed summarization job %s", job.ID)
	return nil
}

func (w *Worker) run() {
	log.Println("Starting Summarization Worker")

	for {
		// Get pending jobs
		jobs, err := w.getPendingJobs()
		if err != nil {
			log.Printf("Error getting pending jobs: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		if len(jobs) > 0 {
			log.Printf("Found %d pending summarization jobs", len(jobs))

			for _, job := range jobs {
				if err := w.processJob(job); err != nil {
					log.Printf("Error processing job %s: %v", job.ID, err)
				}
			}
		} else {
			// No jobs, wait a bit
			time.Sleep(10 * time.Second)
		}

		// Small delay between iterations
		time.Sleep(5 * time.Second)
	}
}

func main() {
	worker, err := NewWorker()
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.db.Close()
	defer worker.redisClient.Close()

	worker.run()
}
