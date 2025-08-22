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

// TestUpdateJobStatus tests the UpdateJobStatus handler
func TestUpdateJobStatus(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		requestBody    map[string]interface{}
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:  "successful job status update",
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
			name:  "invalid status",
			jobID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"status": "invalid_status",
			},
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed for validation error
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:  "database error",
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
					CreatedAt:    time.Now(),
					
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
			req := httptest.NewRequest("POST", "/jobs/"+tt.jobID+"/status", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Post("/jobs/:job_id/status", handler.UpdateJobStatus)

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

// TestRetryJob tests the RetryJob handler
func TestRetryJob(t *testing.T) {
	tests := []struct {
		name           string
		audiobookID    string
		jobID          string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "successful job retry",
			audiobookID: uuid.New().String(),
			jobID:       uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				jobID := uuid.New()
				audiobookID := uuid.New()
				
				job := &models.ProcessingJob{
					ID:           jobID,
					AudiobookID:  audiobookID,
					JobType:      models.JobTypeSummarize,
					Status:       models.JobStatusFailed,
					CreatedAt:    time.Now(),
					
				}
				
				mockRepo.On("GetProcessingJobByID", mock.Anything, jobID).Return(job, nil)
				mockRepo.On("UpdateProcessingJob", mock.Anything, mock.AnythingOfType("*models.ProcessingJob")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "invalid audiobook ID",
			audiobookID: "invalid-uuid",
			jobID:       uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:        "invalid job ID",
			audiobookID: uuid.New().String(),
			jobID:       "invalid-uuid",
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:        "job not found",
			audiobookID: uuid.New().String(),
			jobID:       uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				jobID := uuid.New()
				mockRepo.On("GetProcessingJobByID", mock.Anything, jobID).Return(nil, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedData:   false,
		},
		{
			name:        "job audiobook mismatch",
			audiobookID: uuid.New().String(),
			jobID:       uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				jobID := uuid.New()
				differentAudiobookID := uuid.New()
				
				job := &models.ProcessingJob{
					ID:           jobID,
					AudiobookID:  differentAudiobookID, // Different audiobook ID
					JobType:      models.JobTypeSummarize,
					Status:       models.JobStatusFailed,
				}
				
				mockRepo.On("GetProcessingJobByID", mock.Anything, jobID).Return(job, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:        "job not in failed status",
			audiobookID: uuid.New().String(),
			jobID:       uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				jobID := uuid.New()
				audiobookID := uuid.New()
				
				job := &models.ProcessingJob{
					ID:           jobID,
					AudiobookID:  audiobookID,
					JobType:      models.JobTypeSummarize,
					Status:       models.JobStatusCompleted, // Not failed
				}
				
				mockRepo.On("GetProcessingJobByID", mock.Anything, jobID).Return(job, nil)
			},
			expectedStatus: http.StatusBadRequest,
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
			req := httptest.NewRequest("POST", "/audiobooks/"+tt.audiobookID+"/jobs/"+tt.jobID+"/retry", nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Post("/audiobooks/:id/jobs/:job_id/retry", handler.RetryJob)

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

// TestRetryAllFailedJobs tests the RetryAllFailedJobs handler
func TestRetryAllFailedJobs(t *testing.T) {
	tests := []struct {
		name           string
		audiobookID    string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "successful retry all failed jobs",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.New()
				
				failedJob1 := &models.ProcessingJob{
					ID:           uuid.New(),
					AudiobookID:  audiobookID,
					JobType:      models.JobTypeSummarize,
					Status:       models.JobStatusFailed,
					CreatedAt:    time.Now(),
					
				}
				
				failedJob2 := &models.ProcessingJob{
					ID:           uuid.New(),
					AudiobookID:  audiobookID,
					JobType:      models.JobTypeEmbed,
					Status:       models.JobStatusFailed,
					CreatedAt:    time.Now(),
					
				}
				
				failedJobs := []models.ProcessingJob{*failedJob1, *failedJob2}
				
				mockRepo.On("GetProcessingJobsByAudioBookID", mock.Anything, audiobookID).Return(failedJobs, nil)
				mockRepo.On("UpdateProcessingJob", mock.Anything, mock.AnythingOfType("*models.ProcessingJob")).Return(nil).Times(2)
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
			name:        "no failed jobs found",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.New()
				
				// Return jobs but none are failed
				completedJob := &models.ProcessingJob{
					ID:           uuid.New(),
					AudiobookID:  audiobookID,
					JobType:      models.JobTypeSummarize,
					Status:       models.JobStatusCompleted,
				}
				
				jobs := []models.ProcessingJob{*completedJob}
				
				mockRepo.On("GetProcessingJobsByAudioBookID", mock.Anything, audiobookID).Return(jobs, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "database error getting jobs",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.New()
				mockRepo.On("GetProcessingJobsByAudioBookID", mock.Anything, audiobookID).Return([]models.ProcessingJob{}, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedData:   false,
		},
		{
			name:        "database error updating job",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.New()
				
				failedJob := &models.ProcessingJob{
					ID:           uuid.New(),
					AudiobookID:  audiobookID,
					JobType:      models.JobTypeSummarize,
					Status:       models.JobStatusFailed,
				}
				
				failedJobs := []models.ProcessingJob{*failedJob}
				
				mockRepo.On("GetProcessingJobsByAudioBookID", mock.Anything, audiobookID).Return(failedJobs, nil)
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

			// Create request
			req := httptest.NewRequest("POST", "/audiobooks/"+tt.audiobookID+"/retry-all", nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Post("/audiobooks/:id/retry-all", handler.RetryAllFailedJobs)

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
