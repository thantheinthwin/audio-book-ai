package test

import (
	"audio-book-ai/api/database"
	"audio-book-ai/api/handlers"
	"audio-book-ai/api/models"
	"audio-book-ai/api/services"
	"bytes"
	"context"
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

// MockRepository is a mock implementation of the database.Repository interface
type MockRepository struct {
	mock.Mock
}

// Mock methods for AudioBook operations
func (m *MockRepository) GetAudioBooksByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.AudioBook), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetAudioBookByID(ctx context.Context, id uuid.UUID) (*models.AudioBook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AudioBook), args.Error(1)
}

func (m *MockRepository) GetAudioBookWithDetails(ctx context.Context, id uuid.UUID) (*models.AudioBookWithDetails, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AudioBookWithDetails), args.Error(1)
}

func (m *MockRepository) UpdateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
	args := m.Called(ctx, audiobook)
	return args.Error(0)
}

func (m *MockRepository) DeleteAudioBook(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListAudioBooks(ctx context.Context, limit, offset int, isPublic *bool) ([]models.AudioBook, int, error) {
	args := m.Called(ctx, limit, offset, isPublic)
	return args.Get(0).([]models.AudioBook), args.Int(1), args.Error(2)
}

// Mock methods for other required interface methods (implement as needed)
func (m *MockRepository) CreateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
	args := m.Called(ctx, audiobook)
	return args.Error(0)
}

func (m *MockRepository) UpdateAudioBookStatus(ctx context.Context, id uuid.UUID, status models.AudioBookStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockRepository) CheckAndUpdateAudioBookStatus(ctx context.Context, audiobookID uuid.UUID) error {
	args := m.Called(ctx, audiobookID)
	return args.Error(0)
}

// Mock methods for Chapter operations
func (m *MockRepository) CreateChapter(ctx context.Context, chapter *models.Chapter) error {
	args := m.Called(ctx, chapter)
	return args.Error(0)
}

func (m *MockRepository) GetChapterByID(ctx context.Context, id uuid.UUID) (*models.Chapter, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chapter), args.Error(1)
}

func (m *MockRepository) GetChaptersByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Chapter, error) {
	args := m.Called(ctx, audiobookID)
	return args.Get(0).([]models.Chapter), args.Error(1)
}

func (m *MockRepository) GetFirstChapterByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Chapter, error) {
	args := m.Called(ctx, audiobookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chapter), args.Error(1)
}

func (m *MockRepository) UpdateChapter(ctx context.Context, chapter *models.Chapter) error {
	args := m.Called(ctx, chapter)
	return args.Error(0)
}

func (m *MockRepository) DeleteChapter(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteChaptersByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	args := m.Called(ctx, audiobookID)
	return args.Error(0)
}

// Mock methods for Upload operations
func (m *MockRepository) CreateUpload(ctx context.Context, upload *models.Upload) error {
	args := m.Called(ctx, upload)
	return args.Error(0)
}

func (m *MockRepository) GetUploadByID(ctx context.Context, id uuid.UUID) (*models.Upload, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Upload), args.Error(1)
}

func (m *MockRepository) GetUploadsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Upload, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.Upload), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateUpload(ctx context.Context, upload *models.Upload) error {
	args := m.Called(ctx, upload)
	return args.Error(0)
}

func (m *MockRepository) DeleteUpload(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Mock methods for Upload File operations
func (m *MockRepository) CreateUploadFile(ctx context.Context, uploadFile *models.UploadFile) error {
	args := m.Called(ctx, uploadFile)
	return args.Error(0)
}

func (m *MockRepository) GetUploadFileByID(ctx context.Context, id uuid.UUID) (*models.UploadFile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UploadFile), args.Error(1)
}

func (m *MockRepository) GetUploadFiles(ctx context.Context, uploadID uuid.UUID) ([]models.UploadFile, error) {
	args := m.Called(ctx, uploadID)
	return args.Get(0).([]models.UploadFile), args.Error(1)
}

func (m *MockRepository) UpdateUploadFile(ctx context.Context, uploadFile *models.UploadFile) error {
	args := m.Called(ctx, uploadFile)
	return args.Error(0)
}

func (m *MockRepository) DeleteUploadFile(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteUploadFilesByUploadID(ctx context.Context, uploadID uuid.UUID) error {
	args := m.Called(ctx, uploadID)
	return args.Error(0)
}

func (m *MockRepository) GetUploadedSize(ctx context.Context, uploadID uuid.UUID) (int64, error) {
	args := m.Called(ctx, uploadID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) GetFailedUploadFiles(ctx context.Context, uploadID uuid.UUID) ([]models.UploadFile, error) {
	args := m.Called(ctx, uploadID)
	return args.Get(0).([]models.UploadFile), args.Error(1)
}

func (m *MockRepository) GetRetryingUploadFiles(ctx context.Context, uploadID uuid.UUID) ([]models.UploadFile, error) {
	args := m.Called(ctx, uploadID)
	return args.Get(0).([]models.UploadFile), args.Error(1)
}

func (m *MockRepository) IncrementUploadFileRetryCount(ctx context.Context, fileID uuid.UUID) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *MockRepository) ResetUploadFileRetryCount(ctx context.Context, fileID uuid.UUID) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

// Mock methods for Processing Job operations
func (m *MockRepository) CreateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockRepository) GetProcessingJobsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ProcessingJob, error) {
	args := m.Called(ctx, audiobookID)
	return args.Get(0).([]models.ProcessingJob), args.Error(1)
}

func (m *MockRepository) GetProcessingJobByID(ctx context.Context, id uuid.UUID) (*models.ProcessingJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProcessingJob), args.Error(1)
}

func (m *MockRepository) UpdateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockRepository) GetPendingJobs(ctx context.Context, jobType models.JobType, limit int) ([]models.ProcessingJob, error) {
	args := m.Called(ctx, jobType, limit)
	return args.Get(0).([]models.ProcessingJob), args.Error(1)
}

func (m *MockRepository) GetJobsByStatus(ctx context.Context, status models.JobStatus, limit int) ([]models.ProcessingJob, error) {
	args := m.Called(ctx, status, limit)
	return args.Get(0).([]models.ProcessingJob), args.Error(1)
}

// Mock methods for other required interface methods (stubs)
func (m *MockRepository) CreateTranscript(ctx context.Context, transcript *models.Transcript) error {
	return nil
}

func (m *MockRepository) GetTranscriptByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Transcript, error) {
	return nil, nil
}

func (m *MockRepository) UpdateTranscript(ctx context.Context, transcript *models.Transcript) error {
	return nil
}

func (m *MockRepository) DeleteTranscript(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateChapterTranscript(ctx context.Context, transcript *models.ChapterTranscript) error {
	return nil
}

func (m *MockRepository) GetChapterTranscriptByChapterID(ctx context.Context, chapterID uuid.UUID) (*models.ChapterTranscript, error) {
	return nil, nil
}

func (m *MockRepository) GetChapterTranscriptsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ChapterTranscript, error) {
	return nil, nil
}

func (m *MockRepository) UpdateChapterTranscript(ctx context.Context, transcript *models.ChapterTranscript) error {
	return nil
}

func (m *MockRepository) DeleteChapterTranscript(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) DeleteChapterTranscriptsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateAIOutput(ctx context.Context, output *models.AIOutput) error {
	return nil
}

func (m *MockRepository) GetAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AIOutput, error) {
	return nil, nil
}

func (m *MockRepository) GetAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.AIOutput, error) {
	return nil, nil
}

func (m *MockRepository) UpdateAIOutput(ctx context.Context, output *models.AIOutput) error {
	return nil
}

func (m *MockRepository) DeleteAIOutput(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateChapterAIOutput(ctx context.Context, output *models.ChapterAIOutput) error {
	return nil
}

func (m *MockRepository) GetChapterAIOutputsByChapterID(ctx context.Context, chapterID uuid.UUID) ([]models.ChapterAIOutput, error) {
	return nil, nil
}

func (m *MockRepository) GetChapterAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ChapterAIOutput, error) {
	return nil, nil
}

func (m *MockRepository) GetFirstChapterAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.ChapterAIOutput, error) {
	return nil, nil
}

func (m *MockRepository) UpdateChapterAIOutput(ctx context.Context, output *models.ChapterAIOutput) error {
	return nil
}

func (m *MockRepository) DeleteChapterAIOutput(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) DeleteChapterAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateTag(ctx context.Context, tag *models.Tag) error {
	return nil
}

func (m *MockRepository) GetTagByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	return nil, nil
}

func (m *MockRepository) GetTagByName(ctx context.Context, name string) (*models.Tag, error) {
	return nil, nil
}

func (m *MockRepository) GetTagsByCategory(ctx context.Context, category string) ([]models.Tag, error) {
	return nil, nil
}

func (m *MockRepository) UpdateTag(ctx context.Context, tag *models.Tag) error {
	return nil
}

func (m *MockRepository) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) ListTags(ctx context.Context, limit, offset int) ([]models.Tag, int, error) {
	return nil, 0, nil
}

func (m *MockRepository) CreateAudioBookTag(ctx context.Context, audiobookTag *models.AudioBookTag) error {
	return nil
}

func (m *MockRepository) GetTagsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Tag, error) {
	return nil, nil
}

func (m *MockRepository) GetAudioBooksByTagID(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error) {
	return nil, 0, nil
}

func (m *MockRepository) DeleteAudioBookTag(ctx context.Context, audiobookID, tagID uuid.UUID) error {
	return nil
}

func (m *MockRepository) DeleteAllAudioBookTags(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error {
	return nil
}

func (m *MockRepository) GetEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AudioBookEmbedding, error) {
	return nil, nil
}

func (m *MockRepository) GetEmbeddingByType(ctx context.Context, audiobookID uuid.UUID, embeddingType models.EmbeddingType) (*models.AudioBookEmbedding, error) {
	return nil, nil
}

func (m *MockRepository) UpdateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error {
	return nil
}

func (m *MockRepository) DeleteAudioBookEmbedding(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) DeleteEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (m *MockRepository) SearchAudioBooks(ctx context.Context, query string, limit, offset int, language *string, isPublic *bool) ([]models.AudioBook, int, error) {
	return nil, 0, nil
}

func (m *MockRepository) SearchAudioBooksByVector(ctx context.Context, embedding []float64, embeddingType models.EmbeddingType, limit, offset int) ([]models.AudioBook, []float64, error) {
	return nil, nil, nil
}

func (m *MockRepository) SearchAudioBooksByTags(ctx context.Context, tagNames []string, limit, offset int) ([]models.AudioBook, int, error) {
	return nil, 0, nil
}

func (m *MockRepository) GetAudioBookStats(ctx context.Context) (*database.AudioBookStats, error) {
	return nil, nil
}

func (m *MockRepository) GetUserAudioBookStats(ctx context.Context, userID uuid.UUID) (*database.UserAudioBookStats, error) {
	return nil, nil
}

func (m *MockRepository) CleanupOrphanedData(ctx context.Context) error {
	return nil
}

// Test helper functions
func createTestHandler() (*handlers.Handler, *MockRepository) {
	mockRepo := new(MockRepository)
	mockStorage := &services.SupabaseStorageService{}
	mockRedisQueue := &services.RedisQueueService{}

	handler := handlers.NewHandler(mockRepo, mockStorage, mockRedisQueue)
	return handler, mockRepo
}

func createTestApp() *fiber.App {
	app := fiber.New()
	return app
}

func createTestUserContext() *models.UserContext {
	return &models.UserContext{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Role:  "admin",
	}
}

func createTestAudioBook() *models.AudioBook {
	return &models.AudioBook{
		ID:        uuid.New(),
		Title:     "Test Audio Book",
		Author:    "Test Author",
		Language:  "en",
		IsPublic:  false,
		Status:    models.StatusCompleted,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestAudioBookWithDetails() *models.AudioBookWithDetails {
	audiobook := createTestAudioBook()
	return &models.AudioBookWithDetails{
		AudioBook: *audiobook,
		Chapters:  []models.Chapter{},
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
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetAudioBook tests the GetAudioBook handler
func TestGetAudioBook(t *testing.T) {
	tests := []struct {
		name           string
		audiobookID    string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "successful get audiobook",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.MustParse(uuid.New().String())
				audiobook := createTestAudioBookWithDetails()
				audiobook.CreatedBy = uuid.MustParse(createTestUserContext().ID)
				mockRepo.On("GetAudioBookWithDetails", mock.Anything, audiobookID).Return(audiobook, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "invalid audiobook ID",
			audiobookID: "invalid-uuid",
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed for invalid UUID
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:        "audiobook not found",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.MustParse(uuid.New().String())
				mockRepo.On("GetAudioBookWithDetails", mock.Anything, audiobookID).Return(nil, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusNotFound,
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
			req := httptest.NewRequest("GET", "/audiobooks/"+tt.audiobookID, nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Get("/audiobooks/:id", handler.GetAudioBook)

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
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUpdateAudioBook tests the UpdateAudioBook handler
func TestUpdateAudioBook(t *testing.T) {
	tests := []struct {
		name           string
		audiobookID    string
		requestBody    map[string]interface{}
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "successful update audiobook",
			audiobookID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"title": "Updated Title",
			},
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.MustParse(uuid.New().String())
				existingAudiobook := createTestAudioBook()
				existingAudiobook.CreatedBy = uuid.MustParse(createTestUserContext().ID)
				updatedAudiobook := *existingAudiobook
				updatedAudiobook.Title = "Updated Title"

				mockRepo.On("GetAudioBookByID", mock.Anything, audiobookID).Return(existingAudiobook, nil)
				mockRepo.On("UpdateAudioBook", mock.Anything, mock.AnythingOfType("*models.AudioBook")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "invalid audiobook ID",
			audiobookID: "invalid-uuid",
			requestBody: map[string]interface{}{
				"title": "Updated Title",
			},
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed for invalid UUID
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:        "audiobook not found",
			audiobookID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"title": "Updated Title",
			},
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.MustParse(uuid.New().String())
				mockRepo.On("GetAudioBookByID", mock.Anything, audiobookID).Return(nil, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusNotFound,
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
			req := httptest.NewRequest("PUT", "/audiobooks/"+tt.audiobookID, bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Put("/audiobooks/:id", handler.UpdateAudioBook)

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
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestDeleteAudioBook tests the DeleteAudioBook handler
func TestDeleteAudioBook(t *testing.T) {
	tests := []struct {
		name           string
		audiobookID    string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "successful delete audiobook",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.MustParse(uuid.New().String())
				existingAudiobook := createTestAudioBook()
				existingAudiobook.CreatedBy = uuid.MustParse(createTestUserContext().ID)

				mockRepo.On("GetAudioBookByID", mock.Anything, audiobookID).Return(existingAudiobook, nil)
				mockRepo.On("DeleteAudioBook", mock.Anything, audiobookID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "invalid audiobook ID",
			audiobookID: "invalid-uuid",
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed for invalid UUID
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:        "audiobook not found",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.MustParse(uuid.New().String())
				mockRepo.On("GetAudioBookByID", mock.Anything, audiobookID).Return(nil, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusNotFound,
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
			req := httptest.NewRequest("DELETE", "/audiobooks/"+tt.audiobookID, nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Delete("/audiobooks/:id", handler.DeleteAudioBook)

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
				assert.Contains(t, response, "message")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetPublicAudioBooks tests the GetPublicAudioBooks handler
func TestGetPublicAudioBooks(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name: "successful get public audiobooks",
			setupMock: func(mockRepo *MockRepository) {
				audiobooks := []models.AudioBook{*createTestAudioBook()}
				isPublic := true
				mockRepo.On("ListAudioBooks", mock.Anything, 20, 0, &isPublic).Return(audiobooks, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name: "database error",
			setupMock: func(mockRepo *MockRepository) {
				isPublic := true
				mockRepo.On("ListAudioBooks", mock.Anything, 20, 0, &isPublic).Return([]models.AudioBook{}, 0, fmt.Errorf("database error"))
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

			app.Get("/audiobooks", handler.GetPublicAudioBooks)

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
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetPublicAudioBook tests the GetPublicAudioBook handler
func TestGetPublicAudioBook(t *testing.T) {
	tests := []struct {
		name           string
		audiobookID    string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "successful get public audiobook",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.MustParse(uuid.New().String())
				audiobook := createTestAudioBookWithDetails()
				audiobook.IsPublic = true
				mockRepo.On("GetAudioBookWithDetails", mock.Anything, audiobookID).Return(audiobook, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "invalid audiobook ID",
			audiobookID: "invalid-uuid",
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed for invalid UUID
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name:        "audiobook not found",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.MustParse(uuid.New().String())
				mockRepo.On("GetAudioBookWithDetails", mock.Anything, audiobookID).Return(nil, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedData:   false,
		},
		{
			name:        "audiobook not public",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				audiobookID := uuid.MustParse(uuid.New().String())
				audiobook := createTestAudioBookWithDetails()
				audiobook.IsPublic = false
				mockRepo.On("GetAudioBookWithDetails", mock.Anything, audiobookID).Return(audiobook, nil)
			},
			expectedStatus: http.StatusForbidden,
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
			req := httptest.NewRequest("GET", "/audiobooks/"+tt.audiobookID, nil)
			req.Header.Set("Content-Type", "application/json")

			app.Get("/audiobooks/:id", handler.GetPublicAudioBook)

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
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
