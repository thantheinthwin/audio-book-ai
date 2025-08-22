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

// ProcessJobWithDBUpdates processes a job and updates status directly to database (for Run() method)
func (w *Worker) ProcessJobWithDBUpdates(job models.Job) error {
	log.Printf("Processing %s job %s for audiobook %s (retry %d/%d)", job.JobType, job.ID, job.AudiobookID, job.RetryCount, job.MaxRetries)

	// Check if we've exceeded max retries
	if job.RetryCount >= job.MaxRetries {
		errorMsg := fmt.Sprintf("Job failed after %d retries (max: %d)", job.RetryCount, job.MaxRetries)
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return fmt.Errorf("max retries exceeded for job %s", job.ID)
	}

	// Update job status to running
	if err := w.dbService.UpdateJobStatus(job.ID, models.JobStatusRunning, nil); err != nil {
		return err
	}

	// Process the summarize job
	err := w.ProcessSummarizeJob(job)
	if err != nil {
		// Increment retry count before updating status
		if incrementErr := w.dbService.IncrementRetryCount(job.ID); incrementErr != nil {
			log.Printf("Warning: Failed to increment retry count for job %s: %v", job.ID, incrementErr)
		}

		errorMsg := fmt.Sprintf("Failed to process summarize job: %v", err)
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return err
	}

	// Update job status to completed
	if err := w.dbService.UpdateJobStatus(job.ID, models.JobStatusCompleted, nil); err != nil {
		return err
	}

	log.Printf("Successfully processed %s job %s", job.JobType, job.ID)
	return nil
}

// processEmbedding processes an embedding job
// func (w *Worker) processEmbedding(audiobookID uuid.UUID, transcript string) (*models.AIOutput, error) {
// 	embedding, err := w.geminiService.GenerateEmbedding(transcript)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to generate embedding: %v", err)
// 	}

// 	return &models.AIOutput{
// 		AudiobookID: audiobookID,
// 		OutputType:  models.OutputTypeEmbedding,
// 		Content:     map[string]interface{}{"embedding": embedding},
// 		ModelUsed:   w.geminiService.model,
// 		CreatedAt:   time.Now(),
// 	}, nil
// }

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

// // Run starts the main worker loop
// func (w *Worker) Run() {
// 	log.Println("Starting Gemini AI Processing Worker")

// 	for {
// 		// Get pending jobs
// 		pendingJobs, err := w.dbService.GetPendingJobs(w.config.MaxConcurrentJobs)
// 		if err != nil {
// 			log.Printf("Error getting pending jobs: %v", err)
// 			time.Sleep(30 * time.Second)
// 			continue
// 		}

// 		if len(pendingJobs) > 0 {
// 			log.Printf("Found %d pending AI processing jobs", len(pendingJobs))

// 			for _, job := range pendingJobs {
// 				if err := w.ProcessJobWithDBUpdates(job); err != nil {
// 					log.Printf("Error processing job %s: %v", job.ID, err)
// 				}
// 			}
// 		} else {
// 			// No jobs, wait a bit
// 			time.Sleep(time.Duration(w.config.JobPollInterval) * time.Second)
// 		}

// 		// Small delay between iterations
// 		time.Sleep(5 * time.Second)
// 	}
// }
