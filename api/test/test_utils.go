package test

import (
	"audio-book-ai/api/database"
	"audio-book-ai/api/handlers"
	"audio-book-ai/api/models"
	"audio-book-ai/api/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MockRepository is the centralized mock repository for all tests
type MockRepository = database.TestifyMockRepository

// Test helper functions
func createTestHandler() (*handlers.Handler, *database.TestifyMockRepository) {
	mockRepo := new(database.TestifyMockRepository)
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
