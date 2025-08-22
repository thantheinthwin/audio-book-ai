package test

import (
	"audio-book-ai/api/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCreateAudioBook tests the CreateAudioBook handler
func TestCreateAudioBook(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func(*MockRepository, *map[string]interface{})
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "successful create audiobook",
			requestBody: map[string]interface{}{},
			setupMock: func(mockRepo *MockRepository, requestBody *map[string]interface{}) {
				uploadID := uuid.New()
				// Use a fixed user ID that matches what will be created in the test context
				userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

				// Set up the request body with the correct upload ID
				(*requestBody)["upload_id"] = uploadID.String()
				(*requestBody)["title"] = "Test Audiobook"
				(*requestBody)["author"] = "Test Author"
				(*requestBody)["language"] = "en"
				(*requestBody)["is_public"] = true
				(*requestBody)["price"] = 19.99
				(*requestBody)["description"] = "Test description"

				upload := &models.Upload{
					ID:     uploadID,
					UserID: userID,
					Status: models.UploadStatusCompleted,
				}

				mockFiles := []models.UploadFile{
					{ID: uuid.New(), UploadID: uploadID, FileName: "chapter1.mp3", FilePath: "/path/to/chapter1.mp3"},
				}

				mockRepo.On("GetUploadByID", mock.Anything, uploadID).Return(upload, nil)
				mockRepo.On("GetUploadFiles", mock.Anything, uploadID).Return(mockFiles, nil)
				mockRepo.On("CreateAudioBook", mock.Anything, mock.AnythingOfType("*models.AudioBook")).Return(nil)
				mockRepo.On("CreateProcessingJob", mock.Anything, mock.AnythingOfType("*models.ProcessingJob")).Return(nil)
				mockRepo.On("UpdateUpload", mock.Anything, mock.AnythingOfType("*models.Upload")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedData:   true,
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"title": "Test Audiobook",
				// Missing upload_id, author and language
			},
			setupMock: func(mockRepo *MockRepository, requestBody *map[string]interface{}) {
				// No mock setup needed for validation error
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:        "database error",
			requestBody: map[string]interface{}{},
			setupMock: func(mockRepo *MockRepository, requestBody *map[string]interface{}) {
				uploadID := uuid.New()
				// Use a fixed user ID that matches what will be created in the test context
				userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

				// Set up the request body with the correct upload ID
				(*requestBody)["upload_id"] = uploadID.String()
				(*requestBody)["title"] = "Test Audiobook"
				(*requestBody)["author"] = "Test Author"
				(*requestBody)["language"] = "en"

				upload := &models.Upload{
					ID:     uploadID,
					UserID: userID,
					Status: models.UploadStatusCompleted,
				}

				mockFiles := []models.UploadFile{
					{ID: uuid.New(), UploadID: uploadID, FileName: "chapter1.mp3", FilePath: "/path/to/chapter1.mp3"},
				}

				mockRepo.On("GetUploadByID", mock.Anything, uploadID).Return(upload, nil)
				mockRepo.On("GetUploadFiles", mock.Anything, uploadID).Return(mockFiles, nil)
				mockRepo.On("CreateAudioBook", mock.Anything, mock.AnythingOfType("*models.AudioBook")).Return(fmt.Errorf("database error"))
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
			tt.setupMock(mockRepo, &tt.requestBody)

			// Create request body
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/audiobooks", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := &models.UserContext{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Email: "test@example.com",
				Role:  "admin",
			}
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Post("/audiobooks", handler.CreateAudioBook)

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
				assert.Contains(t, response, "audiobook_id")
			} else {
				assert.Contains(t, response, "error")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetAudioBooks tests the GetAudioBooks handler
func TestGetAudioBooks(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name: "successful get audiobooks",
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobooks := []models.AudioBook{*createTestAudioBook()}
				mockRepo.On("GetAudioBooksByUser", mock.Anything, userID, 20, 0).Return(audiobooks, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name: "database error",
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				mockRepo.On("GetAudioBooksByUser", mock.Anything, userID, 20, 0).Return([]models.AudioBook{}, 0, fmt.Errorf("database error"))
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
			req := httptest.NewRequest("GET", "/audiobooks", nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Get("/audiobooks", handler.GetAudioBooks)

			// Execute request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedData {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response, "data")
				assert.Contains(t, response, "pagination")
			} else {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
