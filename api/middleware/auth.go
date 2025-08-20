package middleware

import (
	"audio-book-ai/api/config"
	"audio-book-ai/api/models"
	"audio-book-ai/api/services"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware handles Supabase JWT authentication
func AuthMiddleware(cfg *config.Config) fiber.Handler {
	authService := services.NewSupabaseAuthService(cfg)

	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")

		// Extract token from header
		tokenString, err := authService.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Validate token using Supabase auth service
		userContext, err := authService.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Store user in context
		c.Locals("user", userContext)
		return c.Next()
	}
}

// OptionalAuthMiddleware provides optional authentication
func OptionalAuthMiddleware(cfg *config.Config) fiber.Handler {
	authService := services.NewSupabaseAuthService(cfg)

	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			return c.Next()
		}

		// Try to extract and validate token
		tokenString, err := authService.ExtractTokenFromHeader(authHeader)
		if err != nil {
			// Invalid header format, continue without authentication
			return c.Next()
		}

		// Try to validate token, but don't fail if it doesn't work
		userContext, err := authService.ValidateToken(tokenString)
		if err != nil {
			// Invalid token, continue without authentication
			return c.Next()
		}

		// Store user in context if validation succeeded
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

// RequireAdmin middleware checks if user has admin role
func RequireAdmin() fiber.Handler {
	return RequireRole(models.RoleAdmin)
}

// RequireUser middleware checks if user has user role (or admin)
func RequireUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.UserContext)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Allow both user and admin roles
		if user.Role != models.RoleUser && user.Role != models.RoleAdmin {
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
