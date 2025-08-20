package services

import (
	"audio-book-ai/api/config"
	"audio-book-ai/api/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SupabaseAuthService handles Supabase authentication
type SupabaseAuthService struct {
	cfg *config.Config
}

// NewSupabaseAuthService creates a new Supabase authentication service
func NewSupabaseAuthService(cfg *config.Config) *SupabaseAuthService {
	return &SupabaseAuthService{
		cfg: cfg,
	}
}

// SupabaseUserResponse represents the response from Supabase user endpoint
type SupabaseUserResponse struct {
	ID            string                 `json:"id"`
	Aud           string                 `json:"aud"`
	Role          string                 `json:"role"`
	Email         string                 `json:"email"`
	AppMetadata   map[string]interface{} `json:"app_metadata"`
	UserMetadata  map[string]interface{} `json:"user_metadata"`
	EmailVerified bool                   `json:"email_verified"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// ValidateToken validates a Supabase JWT token by calling the Supabase user endpoint
func (s *SupabaseAuthService) ValidateToken(tokenString string) (*models.UserContext, error) {
	// Create HTTP client
	client := &http.Client{Timeout: 10 * time.Second}

	// Create request to Supabase user endpoint
	req, err := http.NewRequest("GET", s.cfg.SupabaseURL+"/auth/v1/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("apikey", s.cfg.SupabasePublishableKey)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Supabase API: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Supabase API returned status %d", resp.StatusCode)
	}

	// Parse response
	var userResp SupabaseUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract role from user_metadata first, then fallback to app_metadata
	role := models.RoleUser
	if userMetadata, ok := userResp.UserMetadata["role"].(string); ok && userMetadata != "" {
		role = userMetadata
	} else if appMetadata, ok := userResp.AppMetadata["role"].(string); ok && appMetadata != "" {
		role = appMetadata
	}

	// Create user context
	userContext := &models.UserContext{
		ID:    userResp.ID,
		Email: userResp.Email,
		Aud:   userResp.Aud,
		Role:  role,
		Token: tokenString,
	}

	return userContext, nil
}

// ExtractTokenFromHeader extracts the token from the Authorization header
func (s *SupabaseAuthService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("bearer token required")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("empty token")
	}

	return token, nil
}
