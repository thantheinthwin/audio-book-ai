package main

import (
	"context"
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
	redisConsumer, err := services.NewRedisConsumer(cfg.RedisURL, "audiobook", &services.Config{
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
	if err := redisConsumer.ConsumeJobs(ctx, "transcribe", func(message services.JobMessage) error {
		return processTranscriptionJob(worker, httpClient, cfg.APIBaseURL, message)
	}); err != nil {
		log.Fatalf("Error consuming jobs: %v", err)
	}
}

// processTranscriptionJob processes a transcription job and updates status via HTTP
func processTranscriptionJob(worker *services.Worker, httpClient *http.Client, apiBaseURL string, message services.JobMessage) error {
	// Convert JobMessage to Job model
	job := models.Job{
		ID:          message.ID,
		AudiobookID: message.AudiobookID,
		JobType:     message.JobType,
		Status:      "running",
		CreatedAt:   message.CreatedAt,
	}

	// Set file path if available
	if message.FilePath != nil {
		job.FilePath = *message.FilePath
	}

	// Process the job
	if err := worker.ProcessJob(job); err != nil {
		// Update job status to failed
		now := time.Now()
		updateJobStatus(httpClient, apiBaseURL, message.ID.String(), "failed", err.Error(), nil, &now)
		return err
	}

	// Update job status to completed
	now := time.Now()
	updateJobStatus(httpClient, apiBaseURL, message.ID.String(), "completed", "", &now, &now)
	return nil
}

// updateJobStatus sends job status update to the API
func updateJobStatus(httpClient *http.Client, apiBaseURL string, jobID string, status string, errorMessage string, startedAt, completedAt *time.Time) {
	// This would be implemented to call the API endpoint
	// For now, just log the status update
	log.Printf("Job %s status: %s", jobID, status)
	if errorMessage != "" {
		log.Printf("Job %s error: %s", jobID, errorMessage)
	}
}
