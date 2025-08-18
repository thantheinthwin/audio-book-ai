package middleware

import (
	"fmt"
	"strings"

	"audio-book-ai/api/config"
	"audio-book-ai/api/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/supabase-community/gotrue-go"
)

// AuthMiddleware handles Supabase JWT authentication
func AuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Bearer token required",
			})
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Initialize Supabase client with token
		client := gotrue.New(cfg.SupabaseURL, cfg.SupabasePublishableKey).WithToken(tokenString)

		// Verify token with Supabase
		user, err := client.GetUser()
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Parse JWT to get additional claims
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.SupabasePublishableKey), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token format",
			})
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Create user context
			userContext := &models.UserContext{
				ID:    user.User.ID.String(),
				Email: user.User.Email,
				Aud:   claims["aud"].(string),
				Role:  claims["role"].(string),
				Token: tokenString,
			}

			// Store user in context
			c.Locals("user", userContext)
			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}
}

// OptionalAuthMiddleware provides optional authentication
func OptionalAuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// No token provided, continue without authentication
			return c.Next()
		}

		// Try to authenticate, but don't fail if it doesn't work
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		client := gotrue.New(cfg.SupabaseURL, cfg.SupabasePublishableKey).WithToken(tokenString)

		user, err := client.GetUser()
		if err != nil {
			// Continue without authentication
			return c.Next()
		}

		// If we get here, we have a valid user
		userContext := &models.UserContext{
			ID:    user.User.ID.String(),
			Email: user.User.Email,
			Token: tokenString,
		}

		c.Locals("user", userContext)
		return c.Next()
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.UserContext)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		if user.Role != requiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

// GetUserFromContext extracts user from Fiber context
func GetUserFromContext(c *fiber.Ctx) *models.UserContext {
	if user, ok := c.Locals("user").(*models.UserContext); ok {
		return user
	}
	return nil
}
