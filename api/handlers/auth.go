package handlers

import (
	"audio-book-ai/api/config"
	"audio-book-ai/api/models"
	"audio-book-ai/api/services"

	"github.com/gofiber/fiber/v2"
)

// AuthResponse represents the standard auth response format
type AuthResponse struct {
	User    *models.User `json:"user,omitempty"`
	Message string       `json:"message"`
	Error   string       `json:"error,omitempty"`
}

// ValidateToken validates a Supabase JWT token and returns user info
func ValidateToken(c *fiber.Ctx) error {
	cfg := config.New()
	authService := services.NewSupabaseAuthService(cfg)

	// Get token from header
	authHeader := c.Get("Authorization")

	// Extract token from header
	tokenString, err := authService.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(AuthResponse{
			Error: err.Error(),
		})
	}

	// Validate token using Supabase auth service
	userContext, err := authService.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(AuthResponse{
			Error: "Invalid token",
		})
	}

	// Create user object from context
	user := &models.User{
		ID:         userContext.ID,
		Email:      userContext.Email,
		Role:       userContext.Role,
		IsActive:   true,
		IsVerified: true, // Supabase handles email verification
	}

	return c.JSON(AuthResponse{
		User:    user,
		Message: "Token is valid",
	})
}

// Me returns the current user's information
func Me(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.UserContext)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(AuthResponse{
			Error: "Authentication required",
		})
	}

	// Create user profile from context
	profile := &models.User{
		ID:         user.ID,
		Email:      user.Email,
		Role:       user.Role,
		IsActive:   true,
		IsVerified: true,
	}

	return c.JSON(AuthResponse{
		User:    profile,
		Message: "Profile retrieved successfully",
	})
}
