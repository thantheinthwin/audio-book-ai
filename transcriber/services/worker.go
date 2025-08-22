package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"audio-book-ai/transcriber/models"

	"github.com/google/uuid"
)

// Worker handles the transcription job processing
type Worker struct {
	dbService      *DatabaseService
	revAIService   *RevAIService
	config         *Config
	apiBaseURL     string
	internalAPIKey string
	httpClient     *http.Client
}

// Config holds worker configuration
type Config struct {
	MaxConcurrentJobs int
	JobPollInterval   int
	JobTimeout        int
	APIBaseURL        string
	InternalAPIKey    string
}

// NewWorker creates a new worker
func NewWorker(dbService *DatabaseService, revAIService *RevAIService, config *Config) *Worker {
	return &Worker{
		dbService:      dbService,
		revAIService:   revAIService,
		config:         config,
		apiBaseURL:     config.APIBaseURL,
		internalAPIKey: config.InternalAPIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessJob processes a single transcription job
func (w *Worker) ProcessJob(job models.Job) error {
	log.Printf("Processing transcription job %s for audiobook %s (retry %d/%d)", job.ID, job.AudiobookID, job.RetryCount, job.MaxRetries)

	// Check if we've exceeded max retries
	if job.RetryCount >= job.MaxRetries {
		return fmt.Errorf("max retries exceeded for job %s", job.ID)
	}

	// Check if file exists (assuming local file path for now)
	if _, err := os.Stat(job.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("audio file not found: %s", job.FilePath)
	}

	// Transcribe audio
	transcript, err := w.transcribeAudio(job.FilePath)
	if err != nil {
		return fmt.Errorf("failed to transcribe audio: %v", err)
	}

	// Set audiobook ID, file path, and processing time
	transcript.AudiobookID = job.AudiobookID
	transcript.FilePath = job.FilePath

	// Save transcript to database
	if err := w.dbService.SaveTranscript(transcript); err != nil {
		return fmt.Errorf("failed to save transcript: %v", err)
	}

	// Check if this is chapter 1 and trigger summarize/tag jobs immediately
	if err := w.checkAndTriggerSummarizeTagJobsForChapter1(job.AudiobookID.String(), job.ChapterID); err != nil {
		log.Printf("Warning: Failed to check/trigger summarize and tag jobs for chapter 1: %v", err)
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

// checkAndTriggerSummarizeTagJobsForChapter1 checks if the current chapter is chapter 1 and triggers summarize/tag jobs immediately
func (w *Worker) checkAndTriggerSummarizeTagJobsForChapter1(audiobookID string, chapterID *uuid.UUID) error {
	if chapterID == nil {
		log.Printf("Chapter ID is nil, skipping summarize/tag trigger for audiobook %s", audiobookID)
		return nil
	}

	// Check if this is chapter 1
	isChapter1, err := w.dbService.IsChapter1(*chapterID)
	if err != nil {
		return fmt.Errorf("failed to check if chapter is chapter 1: %v", err)
	}

	if !isChapter1 {
		log.Printf("Chapter %s is not chapter 1 for audiobook %s, skipping summarize/tag trigger", *chapterID, audiobookID)
		return nil
	}

	log.Printf("Chapter 1 transcribed for audiobook %s, triggering summarize and tag jobs immediately", audiobookID)

	// Call the webhook to trigger summarize and tag jobs
	url := fmt.Sprintf("%s/api/v1/internal/audiobooks/%s/trigger-summarize-tag", w.apiBaseURL, audiobookID)
	log.Printf("Calling webhook URL: %s", url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-API-Key", w.internalAPIKey)

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

	log.Printf("Successfully triggered summarize and tag jobs for audiobook %s after chapter 1 transcription, response: %s", audiobookID, string(body))
	return nil
}
