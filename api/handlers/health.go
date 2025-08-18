package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// HealthCheck handles health check requests
func HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "audio-book-ai-api",
		"version": "1.0.0",
	})
}

