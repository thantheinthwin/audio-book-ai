package handlers

import (
	"audio-book-ai/api/models"
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetAudioBooks returns a list of audiobooks for the authenticated user
// GET /audiobooks
func (h *Handler) GetAudioBooks(c *fiber.Ctx) error {
	log.Printf("GetAudioBooks: Request received from IP %s", c.IP())
	log.Printf("GetAudioBooks: User agent: %s", c.Get("User-Agent"))

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("GetAudioBooks: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	log.Printf("GetAudioBooks: User authenticated - UserID: %s", userCtx.ID)
	userID := uuid.MustParse(userCtx.ID)

	// Parse query parameters
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Cap at 100 items per page
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	log.Printf("GetAudioBooks: Fetching audiobooks for user %s with limit=%d, offset=%d", userID, limit, offset)

	// Get audiobooks from database
	audiobooks, total, err := h.repo.ListAudioBooks(context.Background(), limit, offset, nil)
	if err != nil {
		log.Printf("GetAudioBooks: Failed to fetch audiobooks: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch audiobooks",
		})
	}

	log.Printf("GetAudioBooks: Found %d audiobooks (total: %d)", len(audiobooks), total)

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	currentPage := (offset / limit) + 1

	return c.JSON(fiber.Map{
		"data": audiobooks,
		"pagination": fiber.Map{
			"total":        total,
			"limit":        limit,
			"offset":       offset,
			"current_page": currentPage,
			"total_pages":  totalPages,
		},
	})
}

// GetAudioBook returns a specific audiobook by ID
// GET /audiobooks/:id
func (h *Handler) GetAudioBook(c *fiber.Ctx) error {
	log.Printf("GetAudioBook: Request received from IP %s", c.IP())
	log.Printf("GetAudioBook: User agent: %s", c.Get("User-Agent"))

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("GetAudioBook: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("GetAudioBook: Fetching audiobook with ID: %s", audiobookID)

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("GetAudioBook: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("GetAudioBook: User authenticated - UserID: %s", userID)

	// Get audiobook with details from database
	audiobook, err := h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("GetAudioBook: Failed to fetch audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Check if user owns this audiobook or if it's public
	if audiobook.CreatedBy != userID && !audiobook.IsPublic {
		log.Printf("GetAudioBook: Access denied - user does not own audiobook and it's not public")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	log.Printf("GetAudioBook: Successfully retrieved audiobook: %s", audiobook.Title)

	return c.JSON(fiber.Map{
		"data": audiobook,
	})
}

// UpdateAudioBook updates an existing audiobook
// PUT /audiobooks/:id
func (h *Handler) UpdateAudioBook(c *fiber.Ctx) error {
	log.Printf("UpdateAudioBook: Request received from IP %s", c.IP())
	log.Printf("UpdateAudioBook: User agent: %s", c.Get("User-Agent"))

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("UpdateAudioBook: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("UpdateAudioBook: Updating audiobook with ID: %s", audiobookID)

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("UpdateAudioBook: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("UpdateAudioBook: User authenticated - UserID: %s", userID)

	// Get existing audiobook
	existingAudiobook, err := h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("UpdateAudioBook: Failed to fetch audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Check if user owns this audiobook
	if existingAudiobook.CreatedBy != userID {
		log.Printf("UpdateAudioBook: Access denied - user does not own audiobook")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Parse request body
	var req models.UpdateAudioBookRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("UpdateAudioBook: Failed to parse request body: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		log.Printf("UpdateAudioBook: Validation failed: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
		})
	}

	// Update fields if provided
	if req.Title != nil {
		existingAudiobook.Title = *req.Title
	}
	if req.Author != nil {
		existingAudiobook.Author = *req.Author
	}
	if req.Language != nil {
		existingAudiobook.Language = *req.Language
	}
	if req.IsPublic != nil {
		existingAudiobook.IsPublic = *req.IsPublic
	}
	if req.CoverImageURL != nil {
		existingAudiobook.CoverImageURL = req.CoverImageURL
	}

	// Update in database
	if err := h.repo.UpdateAudioBook(context.Background(), existingAudiobook); err != nil {
		log.Printf("UpdateAudioBook: Failed to update audiobook: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update audiobook",
		})
	}

	log.Printf("UpdateAudioBook: Successfully updated audiobook: %s", existingAudiobook.Title)

	return c.JSON(fiber.Map{
		"data": existingAudiobook,
	})
}

// DeleteAudioBook deletes an audiobook
// DELETE /audiobooks/:id
func (h *Handler) DeleteAudioBook(c *fiber.Ctx) error {
	log.Printf("DeleteAudioBook: Request received from IP %s", c.IP())
	log.Printf("DeleteAudioBook: User agent: %s", c.Get("User-Agent"))

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("DeleteAudioBook: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("DeleteAudioBook: Deleting audiobook with ID: %s", audiobookID)

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("DeleteAudioBook: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("DeleteAudioBook: User authenticated - UserID: %s", userID)

	// Get existing audiobook
	existingAudiobook, err := h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("DeleteAudioBook: Failed to fetch audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Check if user owns this audiobook
	if existingAudiobook.CreatedBy != userID {
		log.Printf("DeleteAudioBook: Access denied - user does not own audiobook")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Delete from database
	if err := h.repo.DeleteAudioBook(context.Background(), audiobookID); err != nil {
		log.Printf("DeleteAudioBook: Failed to delete audiobook: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete audiobook",
		})
	}

	log.Printf("DeleteAudioBook: Successfully deleted audiobook: %s", existingAudiobook.Title)

	return c.JSON(fiber.Map{
		"message": "Audiobook deleted successfully",
	})
}

// GetPublicAudioBooks returns a list of public audiobooks
// GET /audiobooks (public route)
func (h *Handler) GetPublicAudioBooks(c *fiber.Ctx) error {
	log.Printf("GetPublicAudioBooks: Request received from IP %s", c.IP())
	log.Printf("GetPublicAudioBooks: User agent: %s", c.Get("User-Agent"))

	// Parse query parameters
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Cap at 100 items per page
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	log.Printf("GetPublicAudioBooks: Fetching public audiobooks with limit=%d, offset=%d", limit, offset)

	// Set isPublic to true to only get public audiobooks
	isPublic := true

	// Get public audiobooks from database
	audiobooks, total, err := h.repo.ListAudioBooks(context.Background(), limit, offset, &isPublic)
	if err != nil {
		log.Printf("GetPublicAudioBooks: Failed to fetch audiobooks: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch audiobooks",
		})
	}

	log.Printf("GetPublicAudioBooks: Found %d public audiobooks (total: %d)", len(audiobooks), total)

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	currentPage := (offset / limit) + 1

	return c.JSON(fiber.Map{
		"data": audiobooks,
		"pagination": fiber.Map{
			"total":        total,
			"limit":        limit,
			"offset":       offset,
			"current_page": currentPage,
			"total_pages":  totalPages,
		},
	})
}

// GetPublicAudioBook returns a specific public audiobook by ID
// GET /audiobooks/:id (public route)
func (h *Handler) GetPublicAudioBook(c *fiber.Ctx) error {
	log.Printf("GetPublicAudioBook: Request received from IP %s", c.IP())
	log.Printf("GetPublicAudioBook: User agent: %s", c.Get("User-Agent"))

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("GetPublicAudioBook: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("GetPublicAudioBook: Fetching public audiobook with ID: %s", audiobookID)

	// Get audiobook with details from database
	audiobook, err := h.repo.GetAudioBookWithDetails(context.Background(), audiobookID)
	if err != nil {
		log.Printf("GetPublicAudioBook: Failed to fetch audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Check if audiobook is public
	if !audiobook.IsPublic {
		log.Printf("GetPublicAudioBook: Access denied - audiobook is not public")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	log.Printf("GetPublicAudioBook: Successfully retrieved public audiobook: %s", audiobook.Title)

	return c.JSON(fiber.Map{
		"data": audiobook,
	})
}
