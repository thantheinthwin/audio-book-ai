package middleware

import (
	"fmt"
	"strings"

	"audio-book-ai/api/config"
	"audio-book-ai/api/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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

		// Parse and validate JWT token
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
			// Extract user information from claims
			userID, _ := claims["sub"].(string)
			email, _ := claims["email"].(string)
			aud, _ := claims["aud"].(string)
			
			// Get role from claims, default to "user" if not present
			role, _ := claims["role"].(string)
			if role == "" {
				role = models.RoleUser
			}

			// Create user context
			userContext := &models.UserContext{
				ID:    userID,
				Email: email,
				Aud:   aud,
				Role:  role,
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

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.SupabasePublishableKey), nil
		})

		if err != nil {
			// Continue without authentication
			return c.Next()
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Extract user information from claims
			userID, _ := claims["sub"].(string)
			email, _ := claims["email"].(string)
			aud, _ := claims["aud"].(string)
			
			// Get role from claims, default to "user" if not present
			role, _ := claims["role"].(string)
			if role == "" {
				role = models.RoleUser
			}

			// Create user context
			userContext := &models.UserContext{
				ID:    userID,
				Email: email,
				Aud:   aud,
				Role:  role,
				Token: tokenString,
			}

			c.Locals("user", userContext)
		}

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
