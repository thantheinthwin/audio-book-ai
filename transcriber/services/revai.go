package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"audio-book-ai/transcriber/models"
)

// RevAIService handles Rev.ai API interactions
type RevAIService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewRevAIService creates a new Rev.ai service
func NewRevAIService(apiKey, baseURL string) *RevAIService {
	return &RevAIService{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SubmitJob submits an audio file to Rev.ai for transcription
func (r *RevAIService) SubmitJob(filePath string) (string, error) {
	jobData := models.RevAIJob{
		MediaURL: filePath, // This should be a publicly accessible URL
		Metadata: "Audio Book AI Transcription",
	}

	jsonData, err := json.Marshal(jobData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job data: %v", err)
	}

	req, err := http.NewRequest("POST", r.baseURL+"/jobs", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to submit job to Rev.ai: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Rev.ai API error: %d - %s", resp.StatusCode, string(body))
	}

	var jobResp models.RevAIJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return jobResp.ID, nil
}

// WaitForJobCompletion waits for a Rev.ai job to complete
func (r *RevAIService) WaitForJobCompletion(jobID string, maxRetries int) (*models.RevAITranscript, error) {
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("GET", r.baseURL+"/jobs/"+jobID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+r.apiKey)

		resp, err := r.client.Do(req)
		if err != nil {
			log.Printf("Failed to check job status: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to get job status: %d", resp.StatusCode)
			time.Sleep(5 * time.Second)
			continue
		}

		var transcript models.RevAITranscript
		if err := json.NewDecoder(resp.Body).Decode(&transcript); err != nil {
			log.Printf("Failed to decode response: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Job %s status: %s, waiting...", jobID, transcript.Status)

		// Check for various completion statuses
		if transcript.Status == "completed" || transcript.Status == "transcribed" || transcript.Status == "done" {
			log.Printf("Job %s completed successfully with status: %s", jobID, transcript.Status)
			return &transcript, nil
		} else if transcript.Status == "failed" || transcript.Status == "error" {
			log.Printf("Job %s failed with status: %s", jobID, transcript.Status)
			return nil, fmt.Errorf("Rev.ai job failed with status: %s", transcript.Status)
		} else if transcript.Status == "canceled" || transcript.Status == "cancelled" {
			log.Printf("Job %s was canceled with status: %s", jobID, transcript.Status)
			return nil, fmt.Errorf("Rev.ai job was canceled with status: %s", transcript.Status)
		}

		// Log the current status for debugging
		log.Printf("Job %s current status: %s, retry %d/%d", jobID, transcript.Status, i+1, maxRetries)
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("job did not complete within timeout")
}

// GetTranscript retrieves the transcript from Rev.ai
func (r *RevAIService) GetTranscript(jobID string) (*models.RevAITranscript, error) {
	req, err := http.NewRequest("GET", r.baseURL+"/jobs/"+jobID+"/transcript", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Accept", "application/vnd.rev.transcript.v1.0+json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get transcript: %d - %s", resp.StatusCode, string(body))
	}

	var transcript models.RevAITranscript
	if err := json.NewDecoder(resp.Body).Decode(&transcript); err != nil {
		return nil, fmt.Errorf("failed to decode transcript: %v", err)
	}

	return &transcript, nil
}

// ProcessTranscript processes the Rev.ai transcript into our format
func (r *RevAIService) ProcessTranscript(revTranscript *models.RevAITranscript) *models.Transcript {
	var content string
	var segments []models.Segment
	var totalConfidence float64
	var confidenceCount int

	for _, monologue := range revTranscript.Monologues {
		for _, element := range monologue.Elements {
			if element.Type == "text" {
				content += element.Value + " "

				segment := models.Segment{
					Start:      element.StartTs,
					End:        element.EndTs,
					Text:       element.Value,
					Confidence: element.Confidence,
					Speaker:    monologue.Speaker,
				}
				segments = append(segments, segment)

				totalConfidence += element.Confidence
				confidenceCount++
			}
		}
	}

	avgConfidence := 0.95 // Default confidence
	if confidenceCount > 0 {
		avgConfidence = totalConfidence / float64(confidenceCount)
	}

	return &models.Transcript{
		Content:               content,
		Segments:              segments,
		Language:              "en", // Rev.ai will detect language
		ConfidenceScore:       avgConfidence,
		ProcessingTimeSeconds: 0, // Will be set by caller
	}
}
