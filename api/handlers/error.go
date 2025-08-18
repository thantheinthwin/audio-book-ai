package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler handles application errors
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default error
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Log error
	log.Printf("Error: %v", err)

	// Return error response
	return c.Status(code).JSON(fiber.Map{
		"error":   message,
		"code":    code,
		"success": false,
	})
}

