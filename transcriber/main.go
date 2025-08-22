package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"audio-book-ai/transcriber/config"
	"audio-book-ai/transcriber/models"
	"audio-book-ai/transcriber/services"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize database service
	dbService, err := services.NewDatabaseService(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create database service: %v", err)
	}
	defer dbService.Close()

	// Initialize Rev.ai service
	revAIService := services.NewRevAIService(cfg.RevAIAPIKey, cfg.RevAIURL)

	// Initialize Redis consumer
	redisConsumer, err := services.NewRedisConsumer(cfg.RedisURL, "audiobooks", &services.Config{
		MaxConcurrentJobs: cfg.MaxConcurrentJobs,
		JobPollInterval:   cfg.JobPollInterval,
		JobTimeout:        cfg.JobTimeout,
	})
	if err != nil {
		log.Fatalf("Failed to create Redis consumer: %v", err)
	}
	defer redisConsumer.Close()

	// Initialize worker configuration
	workerConfig := &services.Config{
		MaxConcurrentJobs: cfg.MaxConcurrentJobs,
		JobPollInterval:   cfg.JobPollInterval,
		JobTimeout:        cfg.JobTimeout,
		APIBaseURL:        cfg.APIBaseURL,
		InternalAPIKey:    cfg.InternalAPIKey,
	}

	// Initialize worker
	worker := services.NewWorker(dbService, revAIService, workerConfig)

	// Create HTTP client for status updates
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping transcriber...")
		cancel()
	}()

	// Start consuming transcription jobs from Redis
	log.Println("Starting transcriber service...")
	if err := redisConsumer.ConsumeJobs(ctx, "transcribe", func(message services.JobMessage, currentRetryCount int) error {
		return processTranscriptionJob(worker, httpClient, cfg.APIBaseURL, cfg.InternalAPIKey, message, currentRetryCount)
	}); err != nil {
		log.Fatalf("Error consuming jobs: %v", err)
	}
}

// processTranscriptionJob processes a transcription job and updates status via HTTP
func processTranscriptionJob(worker *services.Worker, httpClient *http.Client, apiBaseURL string, internalAPIKey string, message services.JobMessage, currentRetryCount int) error {
	// Convert JobMessage to Job model
	job := models.Job{
		ID:          message.ID,
		AudiobookID: message.AudiobookID,
		JobType:     message.JobType,
		Status:      "running",
		RetryCount:  message.RetryCount,
		MaxRetries:  message.MaxRetries,
		CreatedAt:   message.CreatedAt,
	}

	// Set file path if available
	if message.FilePath != nil {
		job.FilePath = *message.FilePath
	} else {
		return fmt.Errorf("no file path provided in job message")
	}

	// fmt.Println("currentRetryCount", currentRetryCount)

	// Update job status to running
	now := time.Now()
	updateJobStatus(httpClient, apiBaseURL, internalAPIKey, message.ID.String(), "running", "", &now, nil, 0)

	// Process the job
	if err := worker.ProcessJob(job); err != nil {
		// Update job status to failed and pass the incremented retry count
		now := time.Now()
		fmt.Println("incremented retryCount", currentRetryCount)
		updateJobStatus(httpClient, apiBaseURL, internalAPIKey, message.ID.String(), "failed", err.Error(), nil, &now, currentRetryCount)
		return err
	}

	// Update job status to completed
	now = time.Now()
	updateJobStatus(httpClient, apiBaseURL, internalAPIKey, message.ID.String(), "completed", "", nil, &now, 0)

	log.Printf("Transcription job %s completed successfully for audiobook %s", message.ID, message.AudiobookID)
	return nil
}

// updateJobStatus sends job status update to the API
func updateJobStatus(httpClient *http.Client, apiBaseURL string, internalAPIKey string, jobID string, status string, errorMessage string, startedAt, completedAt *time.Time, retryCount int) {
	// Build the request payload
	payload := map[string]interface{}{
		"status": status,
	}

	if errorMessage != "" {
		payload["error_message"] = errorMessage
	}
	if startedAt != nil {
		payload["started_at"] = startedAt.Format(time.RFC3339)
	}
	if completedAt != nil {
		payload["completed_at"] = completedAt.Format(time.RFC3339)
	}
	if retryCount > 0 {
		payload["retry_count"] = retryCount
	}

	fmt.Println("retryCount", retryCount)
	fmt.Println("payload retryCount", payload["retry_count"])

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling job status update: %v", err)
		return
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/internal/jobs/%s/status", apiBaseURL, jobID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating job status update request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-API-Key", internalAPIKey)

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Error sending job status update: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Job status update failed with status: %d", resp.StatusCode)
	} else {
		log.Printf("Job %s status updated successfully: %s", jobID, status)
	}
}
