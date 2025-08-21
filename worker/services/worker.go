package services

import (
	"fmt"
	"log"
	"time"

	"audio-book-ai/worker/models"

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
	case models.JobTypeEmbed:
		output, err = w.processEmbedding(job.AudiobookID, transcript)
	case models.JobTypeSummarize:
		// For combined jobs, use the dedicated method
		return w.ProcessSummarizeJob(job)
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

	// Save the output
	if err := w.dbService.SaveAIOutput(output); err != nil {
		errorMsg := fmt.Sprintf("Failed to save output: %v", err)
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
func (w *Worker) processEmbedding(audiobookID uuid.UUID, transcript string) (*models.AIOutput, error) {
	embedding, err := w.geminiService.GenerateEmbedding(transcript)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %v", err)
	}

	return &models.AIOutput{
		AudiobookID: audiobookID,
		OutputType:  models.OutputTypeEmbedding,
		Content:     map[string]interface{}{"embedding": embedding},
		ModelUsed:   w.geminiService.model,
		CreatedAt:   time.Now(),
	}, nil
}

// ProcessSummarizeJob processes a summarize job (includes both summary and tags)
func (w *Worker) ProcessSummarizeJob(job models.Job) error {
	log.Printf("Processing summarize job %s for audiobook %s", job.ID, job.AudiobookID)

	// Update job status to running
	if err := w.dbService.UpdateJobStatus(job.ID, models.JobStatusRunning, nil); err != nil {
		return err
	}

	// Try to get chapter transcripts first
	transcripts, err := w.dbService.GetChapterTranscripts(job.AudiobookID)
	if err != nil {
		// If chapter transcripts fail, try to get a single transcript
		log.Printf("Failed to get chapter transcripts, trying single transcript: %v", err)
		singleTranscript, err := w.dbService.GetTranscript(job.AudiobookID)
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to get any transcripts: %v", err)
			w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
			return err
		}
		// Use the single transcript
		transcripts = []models.ChapterTranscript{
			{
				Content: singleTranscript,
			},
		}
	}

	if len(transcripts) == 0 {
		errorMsg := "No transcripts found for audiobook"
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return fmt.Errorf("no transcripts found for audiobook")
	}

	// Combine all transcripts
	var combinedTranscript string
	for i, transcript := range transcripts {
		if i > 0 {
			combinedTranscript += "\n\n"
		}
		combinedTranscript += transcript.Content
	}

	// Process combined summary and tags
	startTime := time.Now()
	summaryAndTags, err := w.geminiService.GenerateSummaryAndTags(combinedTranscript)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to generate summary and tags: %v", err)
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return err
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
		errorMsg := fmt.Sprintf("Failed to save summary output: %v", err)
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return err
	}

	// Update job status to completed
	if err := w.dbService.UpdateJobStatus(job.ID, models.JobStatusCompleted, nil); err != nil {
		return err
	}

	log.Printf("Successfully processed summarize job %s", job.ID)
	return nil
}

// Run starts the main worker loop
func (w *Worker) Run() {
	log.Println("Starting Gemini AI Processing Worker")

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
