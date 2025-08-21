package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"audio-book-ai/transcriber/models"
)

// Worker handles the transcription job processing
type Worker struct {
	dbService    *DatabaseService
	revAIService *RevAIService
	config       *Config
	apiBaseURL   string
	httpClient   *http.Client
}

// Config holds worker configuration
type Config struct {
	MaxConcurrentJobs int
	JobPollInterval   int
	JobTimeout        int
	APIBaseURL        string
}

// NewWorker creates a new worker
func NewWorker(dbService *DatabaseService, revAIService *RevAIService, config *Config) *Worker {
	return &Worker{
		dbService:    dbService,
		revAIService: revAIService,
		config:       config,
		apiBaseURL:   config.APIBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessJob processes a single transcription job
func (w *Worker) ProcessJob(job models.Job) error {
	log.Printf("Processing transcription job %s for audiobook %s", job.ID, job.AudiobookID)

	// Update job status to running
	if err := w.dbService.UpdateJobStatus(job.ID, models.JobStatusRunning, nil); err != nil {
		return err
	}

	// Check if file exists (assuming local file path for now)
	if _, err := os.Stat(job.FilePath); os.IsNotExist(err) {
		errorMsg := fmt.Sprintf("Audio file not found: %s", job.FilePath)
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return fmt.Errorf("audio file not found: %s", job.FilePath)
	}

	// Transcribe audio
	transcript, err := w.transcribeAudio(job.FilePath)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to transcribe audio: %v", err)
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return err
	}

	// Set audiobook ID, file path, and processing time
	transcript.AudiobookID = job.AudiobookID
	transcript.FilePath = job.FilePath

	// Save transcript to database
	if err := w.dbService.SaveTranscript(transcript); err != nil {
		errorMsg := fmt.Sprintf("Failed to save transcript: %v", err)
		w.dbService.UpdateJobStatus(job.ID, models.JobStatusFailed, &errorMsg)
		return err
	}

	// Update job status to completed
	if err := w.dbService.UpdateJobStatus(job.ID, models.JobStatusCompleted, nil); err != nil {
		return err
	}

	// Check if all chapters are transcribed and trigger summarize/tag jobs if so
	if err := w.checkAndTriggerSummarizeTagJobs(job.AudiobookID.String()); err != nil {
		log.Printf("Warning: Failed to check/trigger summarize and tag jobs: %v", err)
		// Don't fail the transcription job if this fails
	}

	log.Printf("Successfully processed transcription job %s", job.ID)
	return nil
}

// transcribeAudio transcribes an audio file using Rev.ai
func (w *Worker) transcribeAudio(filePath string) (*models.Transcript, error) {
	startTime := time.Now()

	// Submit job to Rev.ai
	jobID, err := w.revAIService.SubmitJob(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to submit to Rev.ai: %v", err)
	}

	log.Printf("Submitted job to Rev.ai: %s", jobID)

	// Wait for job completion
	log.Printf("Waiting for Rev.ai job %s to complete (timeout: %d seconds)", jobID, w.config.JobTimeout/5)
	_, err = w.revAIService.WaitForJobCompletion(jobID, w.config.JobTimeout/5) // 5-second intervals
	if err != nil {
		log.Printf("Rev.ai job %s failed to complete: %v", jobID, err)
		return nil, fmt.Errorf("failed to wait for job completion: %v", err)
	}
	log.Printf("Rev.ai job %s completed successfully", jobID)

	// Get the transcript
	transcript, err := w.revAIService.GetTranscript(jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript: %v", err)
	}

	// Process the transcript
	processedTranscript := w.revAIService.ProcessTranscript(transcript)
	processedTranscript.ProcessingTimeSeconds = int(time.Since(startTime).Seconds())

	return processedTranscript, nil
}

// checkAndTriggerSummarizeTagJobs checks if all chapters are transcribed and triggers summarize/tag jobs
func (w *Worker) checkAndTriggerSummarizeTagJobs(audiobookID string) error {
	// Check if all chapters have transcripts
	allTranscribed, err := w.dbService.AreAllChaptersTranscribed(audiobookID)
	if err != nil {
		return fmt.Errorf("failed to check chapter transcription status: %v", err)
	}

	if !allTranscribed {
		log.Printf("Not all chapters are transcribed for audiobook %s, skipping summarize/tag trigger", audiobookID)
		return nil
	}

	log.Printf("All chapters transcribed for audiobook %s, triggering summarize and tag jobs", audiobookID)

	// Call the webhook to trigger summarize and tag jobs
	url := fmt.Sprintf("%s/v1/admin/audiobooks/%s/trigger-summarize-tag", w.apiBaseURL, audiobookID)
	log.Printf("Calling webhook URL: %s", url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call webhook: %v", err)
	}
	defer resp.Body.Close()

	// Read response body for better error reporting
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Webhook call failed with status: %d, response: %s", resp.StatusCode, string(body))
		return fmt.Errorf("webhook call failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully triggered summarize and tag jobs for audiobook %s, response: %s", audiobookID, string(body))
	return nil
}

// Run starts the main worker loop
func (w *Worker) Run() {
	log.Println("Starting Rev.ai Transcriber Worker")

	for {
		// Get pending jobs
		pendingJobs, err := w.dbService.GetPendingJobs(w.config.MaxConcurrentJobs)
		if err != nil {
			log.Printf("Error getting pending jobs: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		if len(pendingJobs) > 0 {
			log.Printf("Found %d pending transcription jobs", len(pendingJobs))

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
