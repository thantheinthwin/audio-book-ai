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

	"audio-book-ai/worker/config"
	"audio-book-ai/worker/models"
	"audio-book-ai/worker/services"

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

	// Initialize Gemini service
	geminiService := services.NewGeminiService(cfg.GeminiAPIKey, cfg.GeminiURL, cfg.GeminiModel)

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
	worker := services.NewWorker(dbService, geminiService, workerConfig)

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
		log.Println("Received shutdown signal, stopping worker...")
		cancel()
	}()

	// Start consuming AI processing jobs from Redis concurrently
	log.Println("Starting AI processing worker...")

	// Create channels for error handling
	errChan := make(chan error, 3)

	// Start summarize consumer
	go func() {
		if err := redisConsumer.ConsumeJobs(ctx, "summarize", func(message services.JobMessage) error {
			return processAIJob(worker, httpClient, cfg.APIBaseURL, message, "summarize")
		}); err != nil {
			errChan <- fmt.Errorf("Error consuming summarize jobs: %v", err)
		}
	}()

	// Start tag consumer
	go func() {
		if err := redisConsumer.ConsumeJobs(ctx, "tag", func(message services.JobMessage) error {
			return processAIJob(worker, httpClient, cfg.APIBaseURL, message, "tag")
		}); err != nil {
			errChan <- fmt.Errorf("Error consuming tag jobs: %v", err)
		}
	}()

	// Start embed consumer
	go func() {
		if err := redisConsumer.ConsumeJobs(ctx, "embed", func(message services.JobMessage) error {
			return processAIJob(worker, httpClient, cfg.APIBaseURL, message, "embed")
		}); err != nil {
			errChan <- fmt.Errorf("Error consuming embed jobs: %v", err)
		}
	}()

	// Wait for any consumer to fail
	select {
	case err := <-errChan:
		log.Fatalf("Consumer error: %v", err)
	case <-ctx.Done():
		log.Println("Worker shutting down...")
	}
}

// processAIJob processes an AI job and updates status via HTTP
func processAIJob(worker *services.Worker, httpClient *http.Client, apiBaseURL string, message services.JobMessage, jobType string) error {
	// Convert JobMessage to Job model
	job := models.Job{
		ID:          message.ID,
		AudiobookID: message.AudiobookID,
		JobType:     jobType,
		Status:      "running",
		CreatedAt:   message.CreatedAt,
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

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling job status update: %v", err)
		return
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/admin/jobs/%s/status", apiBaseURL, jobID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating job status update request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

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
