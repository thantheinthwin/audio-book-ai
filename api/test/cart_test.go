package test

import (
	"audio-book-ai/api/models"
	"bytes"
	"context"
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

// Cart mock methods for MockRepository
func (m *MockRepository) AddToCart(ctx context.Context, userID, audiobookID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID, audiobookID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockRepository) RemoveFromCart(ctx context.Context, userID, audiobookID uuid.UUID) error {
	args := m.Called(ctx, userID, audiobookID)
	return args.Error(0)
}

func (m *MockRepository) GetCart(ctx context.Context, userID uuid.UUID) ([]models.AudioBook, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.AudioBook), args.Error(1)
}

func (m *MockRepository) IsInCart(ctx context.Context, userID, audiobookID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, audiobookID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) Checkout(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRepository) GetPurchaseHistory(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.PurchaseHistoryResponse, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).(*models.PurchaseHistoryResponse), args.Error(1)
}

func (m *MockRepository) IsAudioBookPurchased(ctx context.Context, userID, audiobookID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, audiobookID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// TestAddToCart tests the AddToCart handler
func TestAddToCart(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name: "successful add to cart",
			requestBody: map[string]interface{}{
				"audiobook_id": uuid.New().String(),
			},
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobookID := uuid.New()
				cartItemID := uuid.New()
				mockRepo.On("AddToCart", mock.Anything, userID, audiobookID).Return(cartItemID, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "missing audiobook ID",
			requestBody: map[string]interface{}{},
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name: "invalid audiobook ID format",
			requestBody: map[string]interface{}{
				"audiobook_id": "invalid-uuid",
			},
			setupMock: func(mockRepo *MockRepository) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedData:   false,
		},
		{
			name: "database error",
			requestBody: map[string]interface{}{
				"audiobook_id": uuid.New().String(),
			},
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobookID := uuid.New()
				mockRepo.On("AddToCart", mock.Anything, userID, audiobookID).Return(uuid.Nil, fmt.Errorf("database error"))
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
			req := httptest.NewRequest("POST", "/cart", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Post("/cart", handler.AddToCart)

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

// TestRemoveFromCart tests the RemoveFromCart handler
func TestRemoveFromCart(t *testing.T) {
	tests := []struct {
		name           string
		audiobookID    string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "successful remove from cart",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobookID := uuid.New()
				mockRepo.On("RemoveFromCart", mock.Anything, userID, audiobookID).Return(nil)
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
			name:        "database error",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobookID := uuid.New()
				mockRepo.On("RemoveFromCart", mock.Anything, userID, audiobookID).Return(fmt.Errorf("database error"))
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
			req := httptest.NewRequest("DELETE", "/cart/"+tt.audiobookID, nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Delete("/cart/:audiobookId", handler.RemoveFromCart)

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

// TestGetCart tests the GetCart handler
func TestGetCart(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name: "successful get cart",
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobooks := []models.AudioBook{*createTestAudioBook()}
				mockRepo.On("GetCart", mock.Anything, userID).Return(audiobooks, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name: "empty cart",
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				mockRepo.On("GetCart", mock.Anything, userID).Return([]models.AudioBook{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name: "database error",
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				mockRepo.On("GetCart", mock.Anything, userID).Return([]models.AudioBook{}, fmt.Errorf("database error"))
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
			req := httptest.NewRequest("GET", "/cart", nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Get("/cart", handler.GetCart)

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
				assert.Contains(t, response, "data")
			} else {
				assert.Contains(t, response, "error")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestIsInCart tests the IsInCart handler
func TestIsInCart(t *testing.T) {
	tests := []struct {
		name           string
		audiobookID    string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "audiobook in cart",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobookID := uuid.New()
				mockRepo.On("IsInCart", mock.Anything, userID, audiobookID).Return(true, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "audiobook not in cart",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobookID := uuid.New()
				mockRepo.On("IsInCart", mock.Anything, userID, audiobookID).Return(false, nil)
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
			name:        "database error",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobookID := uuid.New()
				mockRepo.On("IsInCart", mock.Anything, userID, audiobookID).Return(false, fmt.Errorf("database error"))
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
			req := httptest.NewRequest("GET", "/cart/"+tt.audiobookID+"/check", nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Get("/cart/:audiobookId/check", handler.IsInCart)

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
				assert.Contains(t, response, "in_cart")
			} else {
				assert.Contains(t, response, "error")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestCheckout tests the Checkout handler
func TestCheckout(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name: "successful checkout",
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				mockRepo.On("Checkout", mock.Anything, userID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name: "database error",
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				mockRepo.On("Checkout", mock.Anything, userID).Return(fmt.Errorf("database error"))
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
			req := httptest.NewRequest("POST", "/checkout", nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Post("/checkout", handler.Checkout)

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

// TestGetPurchaseHistory tests the GetPurchaseHistory handler
func TestGetPurchaseHistory(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name: "successful get purchase history",
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				response := &models.PurchaseHistoryResponse{
					Purchases:  []models.PurchasedAudioBookWithDetails{},
					TotalItems: 0,
					TotalSpent: 0.0,
				}
				mockRepo.On("GetPurchaseHistory", mock.Anything, userID, 20, 0).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name: "database error",
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				mockRepo.On("GetPurchaseHistory", mock.Anything, userID, 20, 0).Return(nil, fmt.Errorf("database error"))
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
			req := httptest.NewRequest("GET", "/purchases", nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Get("/purchases", handler.GetPurchaseHistory)

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
				assert.Contains(t, response, "data")
				assert.Contains(t, response, "pagination")
			} else {
				assert.Contains(t, response, "error")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestIsAudioBookPurchased tests the IsAudioBookPurchased handler
func TestIsAudioBookPurchased(t *testing.T) {
	tests := []struct {
		name           string
		audiobookID    string
		setupMock      func(*MockRepository)
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "audiobook purchased",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobookID := uuid.New()
				mockRepo.On("IsAudioBookPurchased", mock.Anything, userID, audiobookID).Return(true, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "audiobook not purchased",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobookID := uuid.New()
				mockRepo.On("IsAudioBookPurchased", mock.Anything, userID, audiobookID).Return(false, nil)
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
			name:        "database error",
			audiobookID: uuid.New().String(),
			setupMock: func(mockRepo *MockRepository) {
				userID := uuid.New()
				audiobookID := uuid.New()
				mockRepo.On("IsAudioBookPurchased", mock.Anything, userID, audiobookID).Return(false, fmt.Errorf("database error"))
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
			req := httptest.NewRequest("GET", "/audiobooks/"+tt.audiobookID+"/purchased", nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context
			userCtx := createTestUserContext()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user", userCtx)
				return c.Next()
			})

			app.Get("/audiobooks/:audiobookId/purchased", handler.IsAudioBookPurchased)

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
				assert.Contains(t, response, "purchased")
			} else {
				assert.Contains(t, response, "error")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
