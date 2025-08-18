package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"audio-book-ai/ai_orchestrator/models"

	"github.com/google/uuid"
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
	log.Printf("Processing %s job %s for audiobook %s", job.JobType, job.ID, job.AudiobookID)

	// Update job status to running
	if err := w.dbService.UpdateJobStatus(job.ID, models.JobStatusRunning, nil); err != nil {
		return err
	}

	// Get transcript for the audiobook
	transcript, err := w.dbService.GetTranscript(job.AudiobookID)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get transcript: %v", err)
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return err
	}

	// Process based on job type
	var output *models.AIOutput
	switch job.JobType {
	case models.JobTypeTag:
		output, err = w.processTags(job.AudiobookID, transcript)
	case models.JobTypeEmbed:
		output, err = w.processEmbedding(job.AudiobookID, transcript)
	default:
		errorMsg := fmt.Sprintf("Unknown job type: %s", job.JobType)
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return fmt.Errorf("unknown job type: %s", job.JobType)
	}

	if err != nil {
		errorMsg := fmt.Sprintf("Failed to process %s: %v", job.JobType, err)
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return err
	}

	// Save AI output to database
	if err := w.dbService.SaveAIOutput(output); err != nil {
		errorMsg := fmt.Sprintf("Failed to save AI output: %v", err)
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

// processTags processes tagging for an audiobook
func (w *Worker) processTags(audiobookID uuid.UUID, transcript *models.Transcript) (*models.AIOutput, error) {
	// Generate tags using Gemini
	tags, err := w.geminiService.GenerateTags(transcript.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tags: %v", err)
	}

	// Create AI output
	content, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %v", err)
	}

	output := &models.AIOutput{
		ID:          uuid.New(),
		AudiobookID: audiobookID,
		OutputType:  "tags",
		Content:     content,
		ModelUsed:   w.geminiService.model,
		CreatedAt:   time.Now(),
	}

	return output, nil
}

// processEmbedding processes embedding for an audiobook
func (w *Worker) processEmbedding(audiobookID uuid.UUID, transcript *models.Transcript) (*models.AIOutput, error) {
	// Generate embedding using Gemini
	embedding, err := w.geminiService.GenerateEmbedding(transcript.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %v", err)
	}

	// Create AI output
	content, err := json.Marshal(embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding: %v", err)
	}

	output := &models.AIOutput{
		ID:          uuid.New(),
		AudiobookID: audiobookID,
		OutputType:  "embedding",
		Content:     content,
		ModelUsed:   w.geminiService.model,
		CreatedAt:   time.Now(),
	}

	return output, nil
}

// Run starts the main worker loop
func (w *Worker) Run() {
	log.Println("Starting AI Orchestrator Worker")

	for {
		// Get pending jobs
		pendingJobs, err := w.dbService.GetPendingJobs(w.config.MaxConcurrentJobs)
		if err != nil {
			log.Printf("Error getting pending jobs: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		if len(pendingJobs) > 0 {
			log.Printf("Found %d pending AI processing jobs", len(pendingJobs))

			for _, job := range pendingJobs {
				if err := w.ProcessJob(job); err != nil {
					log.Printf("Error processing job %s: %v", job.ID, err)
				}
			}
		} else {
			// No jobs, wait a bit
			time.Sleep(time.Duration(w.config.JobPollInterval) * time.Second)
		}

		// Small delay between iterations
		time.Sleep(5 * time.Second)
	}
}
