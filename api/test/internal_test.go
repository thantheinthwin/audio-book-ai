package test

import (
	"audio-book-ai/api/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestInternalTriggerSummarizeAndTagJobs tests the internal trigger endpoint
func TestInternalTriggerSummarizeAndTagJobs(t *testing.T) {
	tests := []struct {
		name           string
		audiobookID    string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "successful internal trigger jobs",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.New()
				audiobook := createTestAudioBook()
				audiobook.ID = audiobookID
				audiobook.Status = models.StatusCompleted
				
				mockRepo.On("GetAudioBookByID", mock.Anything, audiobookID).Return(audiobook, nil)
				mockRepo.On("CreateProcessingJob", mock.Anything, mock.AnythingOfType("*models.ProcessingJob")).Return(nil).Times(2)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "invalid audiobook ID",
			audiobookID: "invalid-uuid",
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:        "audiobook not found",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.New()
				mockRepo.On("GetAudioBookByID", mock.Anything, audiobookID).Return(nil, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedData:   false,
		},
		{
			name:        "audiobook not ready for processing",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.New()
				audiobook := createTestAudioBook()
				audiobook.ID = audiobookID
				audiobook.Status = models.StatusPending // Not completed yet
				
				mockRepo.On("GetAudioBookByID", mock.Anything, audiobookID).Return(audiobook, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:        "database error creating jobs",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.New()
				audiobook := createTestAudioBook()
				audiobook.ID = audiobookID
				audiobook.Status = models.StatusCompleted
				
				mockRepo.On("GetAudioBookByID", mock.Anything, audiobookID).Return(audiobook, nil)
				mockRepo.On("CreateProcessingJob", mock.Anything, mock.AnythingOfType("*models.ProcessingJob")).Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedData:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := createTestApp()
			handler, mockRepo := createTestHandler()
			tt.setupMock(mockRepo)

			// Create request
			req := httptest.NewRequest("POST", "/internal/audiobooks/"+tt.audiobookID+"/trigger-summarize-tag", nil)
			req.Header.Set("Content-Type", "application/json")
			// Note: In real implementation, this would require API key authentication

			app.Post("/internal/audiobooks/:id/trigger-summarize-tag", handler.TriggerSummarizeAndTagJobs)

			// Execute request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectedData {
				assert.Contains(t, response, "message")
			} else {
				assert.Contains(t, response, "error")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestInternalUpdateJobStatus tests the internal job status update endpoint
func TestInternalUpdateJobStatus(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		requestBody    map[string]interface{}
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:  "successful internal job status update",
			jobID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"status": "completed",
				"result": map[string]interface{}{
					"output": "Job completed successfully",
				},
			},
			setupMock: func(mockRepo *MockRepository) {
				jobID := uuid.New()
				job := &models.ProcessingJob{
					ID:           jobID,
					AudiobookID:  uuid.New(),
					JobType:      models.JobTypeSummarize,
					Status:       models.JobStatusRunning,
					CreatedAt:    time.Now(),
					
				}
				
				mockRepo.On("GetProcessingJobByID", mock.Anything, jobID).Return(job, nil)
				mockRepo.On("UpdateProcessingJob", mock.Anything, mock.AnythingOfType("*models.ProcessingJob")).Return(nil)
				mockRepo.On("CheckAndUpdateAudioBookStatus", mock.Anything, job.AudiobookID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:  "invalid job ID",
			jobID: "invalid-uuid",
			requestBody: map[string]interface{}{
				"status": "completed",
			},
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:  "job not found",
			jobID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"status": "completed",
			},
			setupMock: func(mockRepo *MockRepository) {
				jobID := uuid.New()
				mockRepo.On("GetProcessingJobByID", mock.Anything, jobID).Return(nil, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedData:   false,
		},
		{
			name:  "invalid status transition",
			jobID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"status": "running",
			},
			setupMock: func(mockRepo *MockRepository) {
				jobID := uuid.New()
				job := &models.ProcessingJob{
					ID:           jobID,
					AudiobookID:  uuid.New(),
					JobType:      models.JobTypeSummarize,
					Status:       models.JobStatusCompleted, // Already completed
				}
				
				mockRepo.On("GetProcessingJobByID", mock.Anything, jobID).Return(job, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:  "job status update with error result",
			jobID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"status": "failed",
				"error":  "Processing failed due to invalid input",
			},
			setupMock: func(mockRepo *MockRepository) {
				jobID := uuid.New()
				job := &models.ProcessingJob{
					ID:           jobID,
					AudiobookID:  uuid.New(),
					JobType:      models.JobTypeSummarize,
					Status:       models.JobStatusRunning,
					CreatedAt:    time.Now(),
					
				}
				
				mockRepo.On("GetProcessingJobByID", mock.Anything, jobID).Return(job, nil)
				mockRepo.On("UpdateProcessingJob", mock.Anything, mock.AnythingOfType("*models.ProcessingJob")).Return(nil)
				mockRepo.On("CheckAndUpdateAudioBookStatus", mock.Anything, job.AudiobookID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:  "database error updating job",
			jobID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"status": "completed",
			},
			setupMock: func(mockRepo *MockRepository) {
				jobID := uuid.New()
				job := &models.ProcessingJob{
					ID:           jobID,
					AudiobookID:  uuid.New(),
					JobType:      models.JobTypeSummarize,
					Status:       models.JobStatusRunning,
				}
				
				mockRepo.On("GetProcessingJobByID", mock.Anything, jobID).Return(job, nil)
				mockRepo.On("UpdateProcessingJob", mock.Anything, mock.AnythingOfType("*models.ProcessingJob")).Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedData:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := createTestApp()
			handler, mockRepo := createTestHandler()
			tt.setupMock(mockRepo)

			// Create request body
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/internal/jobs/"+tt.jobID+"/status", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			// Note: In real implementation, this would require API key authentication

			app.Post("/internal/jobs/:job_id/status", handler.UpdateJobStatus)

			// Execute request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectedData {
				assert.Contains(t, response, "message")
			} else {
				assert.Contains(t, response, "error")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestInternalAPIKeyMiddleware tests the internal API key middleware behavior
// Note: This is a conceptual test - actual middleware testing would require the middleware setup
func TestInternalAPIKeyMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "missing API key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid API key",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid API key",
			apiKey:         "valid-internal-key",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := createTestApp()

			// Mock middleware behavior
			app.Use(func(c *fiber.Ctx) error {
				apiKey := c.Get("X-API-Key")
				if apiKey == "" {
					return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
						"error": "API key required",
					})
				}
				if apiKey != "valid-internal-key" {
					return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
						"error": "Invalid API key",
					})
				}
				return c.Next()
			})

			app.Get("/internal/test", func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"message": "success"})
			})

			// Create request
			req := httptest.NewRequest("GET", "/internal/test", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			// Execute request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, response, "message")
				assert.Equal(t, "success", response["message"])
			} else {
				assert.Contains(t, response, "error")
			}
		})
	}
}
