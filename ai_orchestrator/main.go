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
	"strings"
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

type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbeddingResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

type AIOrchestrator struct {
	db           *sql.DB
	redisClient  *redis.Client
	ollamaURL    string
	summaryModel string
	embedModel   string
}

func NewAIOrchestrator() (*AIOrchestrator, error) {
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

	// Use small models for faster processing
	summaryModel := os.Getenv("AI_SUMMARY_MODEL")
	if summaryModel == "" {
		summaryModel = "llama2:7b" // Small model
	}

	embedModel := os.Getenv("AI_EMBED_MODEL")
	if embedModel == "" {
		embedModel = "nomic-embed-text" // Small embedding model
	}

	return &AIOrchestrator{
		db:           db,
		redisClient:  redisClient,
		ollamaURL:    ollamaURL,
		summaryModel: summaryModel,
		embedModel:   embedModel,
	}, nil
}

func (ao *AIOrchestrator) getPendingJobs(jobType string) ([]Job, error) {
	query := `
		SELECT pj.id, pj.audiobook_id, pj.job_type, pj.status
		FROM processing_jobs pj
		WHERE pj.job_type = $1 
		AND pj.status = 'pending'
		ORDER BY pj.created_at ASC
		LIMIT 5
	`

	rows, err := ao.db.Query(query, jobType)
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

func (ao *AIOrchestrator) updateJobStatus(jobID uuid.UUID, status string, errorMessage *string) error {
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

	_, err := ao.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update job status: %v", err)
	}

	log.Printf("Updated job %s status to %s", jobID, status)
	return nil
}

func (ao *AIOrchestrator) getTranscript(audiobookID uuid.UUID) (*Transcript, error) {
	query := `
		SELECT content
		FROM transcripts
		WHERE audiobook_id = $1
	`

	var transcript Transcript
	err := ao.db.QueryRow(query, audiobookID).Scan(&transcript.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript: %v", err)
	}

	return &transcript, nil
}

func (ao *AIOrchestrator) generateTags(text string) ([]string, error) {
	// Create a simple prompt for tag generation
	prompt := fmt.Sprintf(`Generate 5-8 relevant tags for this audiobook content. Return only the tags separated by commas, no explanations:

%s

Tags:`, text[:1000]) // Limit text length for faster processing

	// Prepare Ollama request
	req := OllamaRequest{
		Model:  ao.summaryModel,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.2, // Lower temperature for more consistent output
			"top_p":       0.8,
			"max_tokens":  100, // Limit output length
		},
	}

	// Convert request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make request to Ollama
	resp, err := http.Post(ao.ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %v", err)
	}

	// Parse tags from response
	tagsStr := strings.TrimSpace(ollamaResp.Response)
	tags := strings.Split(tagsStr, ",")

	// Clean up tags
	var cleanTags []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}

	return cleanTags, nil
}

func (ao *AIOrchestrator) generateEmbeddings(texts []string) ([][]float64, error) {
	// Prepare embedding request
	req := EmbeddingRequest{
		Model: ao.embedModel,
		Input: texts,
	}

	// Convert request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make request to Ollama
	resp, err := http.Post(ao.ollamaURL+"/api/embeddings", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var embedResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %v", err)
	}

	return embedResp.Embeddings, nil
}

func (ao *AIOrchestrator) saveAIOutput(audiobookID uuid.UUID, outputType string, content interface{}) error {
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

	_, err = ao.db.Exec(query, uuid.New(), audiobookID, outputType, contentJSON, ao.summaryModel, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save AI output: %v", err)
	}

	log.Printf("Saved %s output for audiobook %s", outputType, audiobookID)
	return nil
}

func (ao *AIOrchestrator) saveEmbedding(audiobookID uuid.UUID, embeddingType string, embedding []float64) error {
	query := `
		INSERT INTO audiobook_embeddings (
			id, audiobook_id, embedding, embedding_type, created_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (audiobook_id, embedding_type) DO UPDATE SET
			embedding = EXCLUDED.embedding,
			created_at = EXCLUDED.created_at
	`

	_, err := ao.db.Exec(query, uuid.New(), audiobookID, embedding, embeddingType, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save embedding: %v", err)
	}

	log.Printf("Saved %s embedding for audiobook %s", embeddingType, audiobookID)
	return nil
}

func (ao *AIOrchestrator) processTagJob(job Job) error {
	log.Printf("Processing tagging job %s for audiobook %s", job.ID, job.AudiobookID)

	// Update job status to running
	if err := ao.updateJobStatus(job.ID, "running", nil); err != nil {
		return err
	}

	// Get transcript
	transcript, err := ao.getTranscript(job.AudiobookID)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get transcript: %v", err)
		ao.updateJobStatus(job.ID, "failed", &errorMsg)
		return err
	}

	// Generate tags
	tags, err := ao.generateTags(transcript.Content)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to generate tags: %v", err)
		ao.updateJobStatus(job.ID, "failed", &errorMsg)
		return err
	}

	// Save tags
	tagsData := map[string]interface{}{
		"tags":  tags,
		"count": len(tags),
	}

	if err := ao.saveAIOutput(job.AudiobookID, "tags", tagsData); err != nil {
		errorMsg := fmt.Sprintf("Failed to save tags: %v", err)
		ao.updateJobStatus(job.ID, "failed", &errorMsg)
		return err
	}

	// Update job status to completed
	if err := ao.updateJobStatus(job.ID, "completed", nil); err != nil {
		return err
	}

	log.Printf("Successfully processed tagging job %s", job.ID)
	return nil
}

func (ao *AIOrchestrator) processEmbedJob(job Job) error {
	log.Printf("Processing embedding job %s for audiobook %s", job.ID, job.AudiobookID)

	// Update job status to running
	if err := ao.updateJobStatus(job.ID, "running", nil); err != nil {
		return err
	}

	// Get transcript
	transcript, err := ao.getTranscript(job.AudiobookID)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get transcript: %v", err)
		ao.updateJobStatus(job.ID, "failed", &errorMsg)
		return err
	}

	// Generate embeddings for different content types
	texts := []string{
		transcript.Content[:1000], // Truncate for faster processing
	}

	embeddings, err := ao.generateEmbeddings(texts)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to generate embeddings: %v", err)
		ao.updateJobStatus(job.ID, "failed", &errorMsg)
		return err
	}

	// Save transcript embedding
	if len(embeddings) > 0 {
		if err := ao.saveEmbedding(job.AudiobookID, "transcript", embeddings[0]); err != nil {
			errorMsg := fmt.Sprintf("Failed to save embedding: %v", err)
			ao.updateJobStatus(job.ID, "failed", &errorMsg)
			return err
		}
	}

	// Update job status to completed
	if err := ao.updateJobStatus(job.ID, "completed", nil); err != nil {
		return err
	}

	log.Printf("Successfully processed embedding job %s", job.ID)
	return nil
}

func (ao *AIOrchestrator) run() {
	log.Println("Starting AI Orchestrator")

	for {
		// Process tagging jobs
		tagJobs, err := ao.getPendingJobs("tag")
		if err != nil {
			log.Printf("Error getting pending tag jobs: %v", err)
		} else if len(tagJobs) > 0 {
			log.Printf("Found %d pending tagging jobs", len(tagJobs))
			for _, job := range tagJobs {
				if err := ao.processTagJob(job); err != nil {
					log.Printf("Error processing tag job %s: %v", job.ID, err)
				}
			}
		}

		// Process embedding jobs
		embedJobs, err := ao.getPendingJobs("embed")
		if err != nil {
			log.Printf("Error getting pending embed jobs: %v", err)
		} else if len(embedJobs) > 0 {
			log.Printf("Found %d pending embedding jobs", len(embedJobs))
			for _, job := range embedJobs {
				if err := ao.processEmbedJob(job); err != nil {
					log.Printf("Error processing embed job %s: %v", job.ID, err)
				}
			}
		}

		// If no jobs, wait a bit
		if len(tagJobs) == 0 && len(embedJobs) == 0 {
			time.Sleep(10 * time.Second)
		}

		// Small delay between iterations
		time.Sleep(5 * time.Second)
	}
}

func main() {
	orchestrator, err := NewAIOrchestrator()
	if err != nil {
		log.Fatalf("Failed to create AI orchestrator: %v", err)
	}
	defer orchestrator.db.Close()
	defer orchestrator.redisClient.Close()

	orchestrator.run()
}
