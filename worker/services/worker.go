package services

import (
	"fmt"
	"log"
	"time"

	"audio-book-ai/worker/models"
)

// Worker handles AI processing job execution
type Worker struct {
	dbService     *DatabaseService
	geminiService *GeminiService
	config        *Config
}

// Config holds worker configuration
type Config struct {
	MaxConcurrentJobs int
	JobPollInterval   int
	JobTimeout        int
}

// NewWorker creates a new worker
func NewWorker(dbService *DatabaseService, geminiService *GeminiService, config *Config) *Worker {
	return &Worker{
		dbService:     dbService,
		geminiService: geminiService,
		config:        config,
	}
}

// ProcessJob processes a single AI processing job
func (w *Worker) ProcessJob(job models.Job) error {
	log.Printf("Processing %s job %s for audiobook %s (retry %d/%d)", job.JobType, job.ID, job.AudiobookID, job.RetryCount, job.MaxRetries)

	// Check if we've exceeded max retries
	if job.RetryCount >= job.MaxRetries {
		return fmt.Errorf("max retries exceeded for job %s", job.ID)
	}

	// Process the summarize job
	err := w.ProcessSummarizeJob(job)
	if err != nil {
		return err
	}

	log.Printf("Successfully processed %s job %s", job.JobType, job.ID)
	return nil
}

// ProcessSummarizeJob processes a summarize job (includes both summary and tags)
func (w *Worker) ProcessSummarizeJob(job models.Job) error {
	log.Printf("Processing summarize job %s for audiobook %s", job.ID, job.AudiobookID)

	// Try to get chapter 1 transcript first
	chapter1Transcript, err := w.dbService.GetChapter1Transcript(job.AudiobookID.String())
	if err != nil {
		return fmt.Errorf("failed to get chapter 1 transcript: %v", err)
	}

	if chapter1Transcript == nil {
		return fmt.Errorf("chapter 1 transcript not found for audiobook")
	}

	// Use chapter 1 transcript for summary and tags generation
	transcriptContent := chapter1Transcript.Content

	// Process summary and tags from chapter 1
	startTime := time.Now()
	summaryAndTags, err := w.geminiService.GenerateSummaryAndTags(transcriptContent)
	if err != nil {
		return fmt.Errorf("failed to generate summary and tags: %v", err)
	}

	// Save summary output
	summaryOutput := &models.AIOutput{
		AudiobookID:           job.AudiobookID,
		OutputType:            models.OutputTypeSummary,
		Content:               map[string]interface{}{"summary": summaryAndTags.Summary, "tags": summaryAndTags.Tags},
		ModelUsed:             w.geminiService.model,
		CreatedAt:             time.Now(),
		ProcessingTimeSeconds: int(time.Since(startTime).Seconds()),
	}

	if err := w.dbService.SaveAIOutput(summaryOutput); err != nil {
		return fmt.Errorf("failed to save summary output: %v", err)
	}

	log.Printf("Successfully processed summarize job %s", job.ID)
	return nil
}
