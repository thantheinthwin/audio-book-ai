package test

import (
	"audio-book-ai/api/handlers"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestValidateToken tests the ValidateToken handler
func TestValidateToken(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedData   bool
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedData:   false,
		},
		{
			name:           "invalid authorization header format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
			expectedData:   false,
		},
		{
			name:           "invalid bearer token",
			authHeader:     "Bearer invalid_token",
			expectedStatus: http.StatusUnauthorized,
			expectedData:   false,
		},
		// Note: Valid token tests would require actual Supabase setup
		// In a real test environment, you would mock the Supabase auth service
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := createTestApp()

			// Create request
			req := httptest.NewRequest("POST", "/auth/validate", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			req.Header.Set("Content-Type", "application/json")

			app.Post("/auth/validate", handlers.ValidateToken)

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
				assert.Contains(t, response, "user")
				assert.Contains(t, response, "message")
			} else {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			}
		})
	}
}

// TestMe tests the Me handler
func TestMe(t *testing.T) {
	tests := []struct {
		name           string
		setupAuth      bool
		expectedStatus int
		expectedData   bool
	}{
		{
			name:           "successful get user profile",
			setupAuth:      true,
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:           "unauthorized access",
			setupAuth:      false,
			expectedStatus: http.StatusUnauthorized,
			expectedData:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := createTestApp()

			// Create request
			req := httptest.NewRequest("GET", "/auth/me", nil)
			req.Header.Set("Content-Type", "application/json")

			// Set user context if needed
			if tt.setupAuth {
				userCtx := createTestUserContext()
				app.Use(func(c *fiber.Ctx) error {
					c.Locals("user", userCtx)
					return c.Next()
				})
			}

			app.Get("/auth/me", handlers.Me)

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
				assert.Contains(t, response, "user")
				assert.Contains(t, response, "message")
			} else {
				assert.Contains(t, response, "error")
			}
		})
	}
}

// TestHealthCheck tests the HealthCheck handler
func TestHealthCheck(t *testing.T) {
	// Setup
	app := createTestApp()

	// Create request
	req := httptest.NewRequest("GET", "/auth/health", nil)
	req.Header.Set("Content-Type", "application/json")

	app.Get("/auth/health", handlers.HealthCheck)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assertions
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "audio-book-ai-api", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
}
