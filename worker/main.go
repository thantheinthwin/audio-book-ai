package main

import (
	"log"

	"audio-book-ai/worker/config"
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

	// Initialize worker configuration
	workerConfig := &services.Config{
		MaxConcurrentJobs: cfg.MaxConcurrentJobs,
		JobPollInterval:   cfg.JobPollInterval,
		JobTimeout:        cfg.JobTimeout,
	}

	// Initialize worker
	worker := services.NewWorker(dbService, geminiService, workerConfig)

	// Start the worker
	worker.Run()
}
