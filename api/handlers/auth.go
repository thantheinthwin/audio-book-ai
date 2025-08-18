package handlers

import (
	"audio-book-ai/api/config"
	"audio-book-ai/api/models"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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

	// Get token from header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(AuthResponse{
			Error: "Authorization header required",
		})
	}

	// Check if it's a Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(AuthResponse{
			Error: "Bearer token required",
		})
	}

	// Extract token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.SupabasePublishableKey), nil
	})

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(AuthResponse{
			Error: "Invalid token format",
		})
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Extract user information from claims
		userID, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)

		// Create user object
		user := &models.User{
			ID:         userID,
			Email:      email,
			IsActive:   true,
			IsVerified: true, // Supabase handles email verification
		}

		return c.JSON(AuthResponse{
			User:    user,
			Message: "Token is valid",
		})
	}

	return c.Status(fiber.StatusUnauthorized).JSON(AuthResponse{
		Error: "Invalid token claims",
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
		IsActive:   true,
		IsVerified: true,
	}

	return c.JSON(AuthResponse{
		User:    profile,
		Message: "Profile retrieved successfully",
	})
}
