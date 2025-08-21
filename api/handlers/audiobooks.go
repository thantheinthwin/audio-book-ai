package handlers

import (
	"audio-book-ai/api/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	audiobook, err := h.repo.GetAudioBookWithDetails(context.Background(), audiobookID)
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

// CreateAudioBook creates an audio book from a completed upload session
// POST /v1/admin/audiobooks
func (h *Handler) CreateAudioBook(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("CreateAudioBook: PANIC recovered: %v", r)
		}
	}()

	log.Printf("CreateAudioBook: Request received from IP %s", c.IP())
	log.Printf("CreateAudioBook: User agent: %s", c.Get("User-Agent"))

	// Log raw request body for debugging
	body := c.Body()
	log.Printf("CreateAudioBook: Raw request body: %s", string(body))

	var req models.CreateAudioBookFromUploadRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("CreateAudioBook: Failed to parse request body: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	log.Printf("CreateAudioBook: Request parsed successfully - UploadID: %s, Title: %s, Author: %s, Language: %s, IsPublic: %v",
		req.UploadID, req.Title, req.Author, req.Language, req.IsPublic)

	if err := req.Validate(); err != nil {
		log.Printf("CreateAudioBook: Validation failed: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Validation error: %v", err),
		})
	}

	log.Printf("CreateAudioBook: Request validation passed")

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	log.Printf("CreateAudioBook: User context: %v", userCtx)
	if !ok || userCtx == nil {
		log.Printf("CreateAudioBook: User not authenticated - user context not found in context")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	log.Printf("CreateAudioBook: User authenticated - UserID: %s", userCtx.ID)
	userID := uuid.MustParse(userCtx.ID)

	// Get upload session
	log.Printf("CreateAudioBook: Looking up upload session with ID: %s", req.UploadID)
	upload, err := h.repo.GetUploadByID(context.Background(), req.UploadID)
	if err != nil {
		log.Printf("CreateAudioBook: Failed to get upload session: %v", err)
		// Log more details about the error
		if strings.Contains(err.Error(), "connection") {
			log.Printf("CreateAudioBook: Database connection error detected")
		}
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Upload session not found",
		})
	}

	log.Printf("CreateAudioBook: Upload session found - Status: %s, UserID: %s, UploadType: %s",
		upload.Status, upload.UserID, upload.UploadType)

	// Check if user owns this upload
	log.Printf("CreateAudioBook: Checking ownership - Upload UserID: %s, Request UserID: %s", upload.UserID, userID)
	if upload.UserID != userID {
		log.Printf("CreateAudioBook: Access denied - user does not own this upload")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	log.Printf("CreateAudioBook: Ownership verified")

	// Check if upload is completed
	log.Printf("CreateAudioBook: Checking upload status - Current: %s, Expected: %s", upload.Status, models.UploadStatusCompleted)
	if upload.Status != models.UploadStatusCompleted {
		log.Printf("CreateAudioBook: Upload session is not completed")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Upload session is not completed",
		})
	}

	log.Printf("CreateAudioBook: Upload status verified as completed")

	// Get upload files
	log.Printf("CreateAudioBook: Getting upload files for upload ID: %s", req.UploadID)
	uploadFiles, err := h.repo.GetUploadFiles(context.Background(), req.UploadID)
	if err != nil {
		log.Printf("CreateAudioBook: Failed to get upload files: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get upload files",
		})
	}

	log.Printf("CreateAudioBook: Found %d upload files", len(uploadFiles))

	if len(uploadFiles) == 0 {
		log.Printf("CreateAudioBook: No files found in upload session")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No files found in upload session",
		})
	}

	// Create audio book
	log.Printf("CreateAudioBook: Creating audio book with title: %s, author: %s", req.Title, req.Author)
	audiobook := &models.AudioBook{
		ID:        uuid.New(),
		Title:     req.Title,
		Author:    req.Author,
		Language:  req.Language,
		IsPublic:  req.IsPublic,
		Status:    models.StatusProcessing, // Start with processing status
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set description if provided
	if req.Description != nil && *req.Description != "" {
		audiobook.Summary = req.Description
	}

	// Set cover image URL if provided
	if req.CoverImageURL != nil && *req.CoverImageURL != "" {
		audiobook.CoverImageURL = req.CoverImageURL
	}

	// Handle different upload types
	log.Printf("CreateAudioBook: Processing upload type: %s", upload.UploadType)
	var chapters []*models.Chapter

	if upload.UploadType == models.UploadTypeSingle {
		// Single file upload - create one chapter
		log.Printf("CreateAudioBook: Processing single file upload")
		if len(uploadFiles) != 1 {
			log.Printf("CreateAudioBook: Single upload type requires exactly one file, found %d", len(uploadFiles))
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Single upload type requires exactly one file",
			})
		}

		file := uploadFiles[0]
		log.Printf("CreateAudioBook: Creating single chapter with file: %s, size: %d", file.FilePath, file.FileSize)

		chapter := &models.Chapter{
			ID:            uuid.New(),
			AudiobookID:   audiobook.ID,
			ChapterNumber: 1,
			Title:         req.Title, // Use audiobook title for single file
			FilePath:      file.FilePath,
			FileURL:       &file.FilePath, // For now, use file path as URL
			FileSizeBytes: &file.FileSize,
			MimeType:      &file.MimeType,
			CreatedAt:     time.Now(),
		}
		chapters = append(chapters, chapter)

	} else if upload.UploadType == models.UploadTypeChapters {
		// Chaptered upload
		log.Printf("CreateAudioBook: Processing chaptered upload")
		if len(uploadFiles) == 0 {
			log.Printf("CreateAudioBook: No chapter files found")
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "No chapter files found",
			})
		}

		// Sort files by chapter number
		chapterFiles := make(map[int]*models.UploadFile)
		for _, file := range uploadFiles {
			log.Printf("CreateAudioBook: Processing file: %s, chapter number: %v", file.FileName, file.ChapterNumber)
			if file.ChapterNumber == nil {
				log.Printf("CreateAudioBook: File %s missing chapter number", file.FileName)
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "All files must have chapter numbers for chaptered uploads",
				})
			}
			chapterFiles[*file.ChapterNumber] = &file
		}

		log.Printf("CreateAudioBook: Processed %d chapter files", len(chapterFiles))

		// Check if chapter 1 exists
		if chapterFiles[1] == nil {
			log.Printf("CreateAudioBook: Chapter 1 is required but not found")
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Chapter 1 is required",
			})
		}

		// Create chapter records
		for chapterNum, file := range chapterFiles {
			chapterTitle := file.FileName
			if file.ChapterTitle != nil && *file.ChapterTitle != "" {
				chapterTitle = *file.ChapterTitle
			}

			chapter := &models.Chapter{
				ID:            uuid.New(),
				AudiobookID:   audiobook.ID,
				ChapterNumber: chapterNum,
				Title:         chapterTitle,
				FilePath:      file.FilePath,
				FileURL:       &file.FilePath, // For now, use file path as URL
				FileSizeBytes: &file.FileSize,
				MimeType:      &file.MimeType,
				CreatedAt:     time.Now(),
			}
			chapters = append(chapters, chapter)
		}
	}

	// Save audio book to database first (required for foreign key constraint)
	log.Printf("CreateAudioBook: Saving audio book to database with ID: %s", audiobook.ID)
	if err := h.repo.CreateAudioBook(context.Background(), audiobook); err != nil {
		log.Printf("CreateAudioBook: Failed to create audio book in database: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create audio book",
		})
	}
	log.Printf("CreateAudioBook: Audio book saved to database successfully")

	// Save chapters to database
	log.Printf("CreateAudioBook: Creating %d chapters in database", len(chapters))
	for _, chapter := range chapters {
		if err := h.repo.CreateChapter(context.Background(), chapter); err != nil {
			log.Printf("CreateAudioBook: Failed to create chapter %d: %v", chapter.ChapterNumber, err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create chapter",
			})
		}
	}
	log.Printf("CreateAudioBook: All chapters created successfully")

	// Create transcription jobs for each chapter and enqueue them to Redis
	log.Printf("CreateAudioBook: Creating transcription jobs for %d chapters", len(chapters))
	jobsCreated := 0

	for _, chapter := range chapters {
		log.Printf("CreateAudioBook: Creating transcription job for chapter %d", chapter.ChapterNumber)
		job := &models.ProcessingJob{
			ID:          uuid.New(),
			AudiobookID: audiobook.ID,
			JobType:     models.JobTypeTranscribe,
			Status:      models.JobStatusPending,
			CreatedAt:   time.Now(),
		}

		// Save job to database
		if err := h.repo.CreateProcessingJob(context.Background(), job); err != nil {
			log.Printf("CreateAudioBook: Failed to create transcription job for chapter %d: %v", chapter.ChapterNumber, err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create transcription job",
			})
		}
		log.Printf("CreateAudioBook: Transcription job for chapter %d created in database", chapter.ChapterNumber)

		// Enqueue job to Redis if Redis service is available
		if h.redisQueue != nil {
			log.Printf("CreateAudioBook: Enqueueing transcription job for chapter %d with file path: %s", chapter.ChapterNumber, chapter.FilePath)
			if err := h.redisQueue.EnqueueTranscriptionJob(context.Background(), job, chapter.FilePath); err != nil {
				log.Printf("CreateAudioBook: Failed to enqueue transcription job for chapter %d to Redis: %v", chapter.ChapterNumber, err)
				// Continue with other jobs even if one fails to enqueue
			} else {
				log.Printf("CreateAudioBook: Transcription job for chapter %d enqueued to Redis successfully", chapter.ChapterNumber)
				jobsCreated++
			}
		} else {
			log.Printf("CreateAudioBook: Redis queue service not available, transcription job for chapter %d created in database only", chapter.ChapterNumber)
			jobsCreated++
		}
	}

	// Update upload status to indicate it's been processed
	log.Printf("CreateAudioBook: Updating upload status to processed")
	upload.Status = "processed"
	upload.UpdatedAt = time.Now()
	if err := h.repo.UpdateUpload(context.Background(), upload); err != nil {
		// Log error but don't fail the request
		log.Printf("CreateAudioBook: Failed to update upload status: %v", err)
	} else {
		log.Printf("CreateAudioBook: Upload status updated successfully")
	}

	log.Printf("CreateAudioBook: Request completed successfully - AudioBookID: %s, TranscriptionJobsCreated: %d/%d",
		audiobook.ID, jobsCreated, len(chapters))
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"audiobook_id":               audiobook.ID,
		"status":                     audiobook.Status,
		"message":                    "Audio book created successfully",
		"transcription_jobs_created": jobsCreated,
		"total_chapters":             len(chapters),
	})
}

// TriggerSummarizeAndTagJobs triggers summarize and tag jobs for an audiobook after transcription is complete
// POST /v1/admin/audiobooks/{id}/trigger-summarize-tag
func (h *Handler) TriggerSummarizeAndTagJobs(c *fiber.Ctx) error {
	log.Printf("TriggerSummarizeAndTagJobs: Request received from IP %s", c.IP())

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("TriggerSummarizeAndTagJobs: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("TriggerSummarizeAndTagJobs: Processing audiobook ID: %s", audiobookID)

	// Get audiobook to check if it exists
	_, err = h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("TriggerSummarizeAndTagJobs: Failed to get audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Get all chapters for this audiobook
	chapters, err := h.repo.GetChaptersByAudioBookID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("TriggerSummarizeAndTagJobs: Failed to get chapters: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get chapters",
		})
	}

	if len(chapters) == 0 {
		log.Printf("TriggerSummarizeAndTagJobs: No chapters found for audiobook")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No chapters found for audiobook",
		})
	}

	// Check if all chapters have been transcribed successfully
	allTranscribed := true
	var failedChapters []int

	for _, chapter := range chapters {
		// Check if chapter has a transcript
		transcript, err := h.repo.GetChapterTranscriptByChapterID(context.Background(), chapter.ID)
		if err != nil || transcript == nil {
			log.Printf("TriggerSummarizeAndTagJobs: Chapter %d has no transcript", chapter.ChapterNumber)
			allTranscribed = false
			failedChapters = append(failedChapters, chapter.ChapterNumber)
		}
	}

	if !allTranscribed {
		log.Printf("TriggerSummarizeAndTagJobs: Not all chapters are transcribed. Failed chapters: %v", failedChapters)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":           "Not all chapters are transcribed",
			"failed_chapters": failedChapters,
		})
	}

	log.Printf("TriggerSummarizeAndTagJobs: All %d chapters are transcribed, proceeding with summarize and tag jobs", len(chapters))

	// Create combined summarize and tag job
	job := &models.ProcessingJob{
		ID:          uuid.New(),
		AudiobookID: audiobookID,
		JobType:     models.JobTypeSummarize, // We'll use this for the combined job
		Status:      models.JobStatusPending,
		CreatedAt:   time.Now(),
	}

	// Save job to database
	if err := h.repo.CreateProcessingJob(context.Background(), job); err != nil {
		log.Printf("TriggerSummarizeAndTagJobs: Failed to create summarize and tag job: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create summarize and tag job",
		})
	}

	// Enqueue job to Redis if Redis service is available
	if h.redisQueue != nil {
		if err := h.redisQueue.EnqueueAIJob(context.Background(), job); err != nil {
			log.Printf("TriggerSummarizeAndTagJobs: Failed to enqueue summarize and tag job to Redis: %v", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to enqueue summarize and tag job",
			})
		}
		log.Printf("TriggerSummarizeAndTagJobs: Summarize and tag job enqueued to Redis successfully")
	} else {
		log.Printf("TriggerSummarizeAndTagJobs: Redis queue service not available, job created in database only")
	}

	log.Printf("TriggerSummarizeAndTagJobs: Successfully triggered summarize and tag jobs for audiobook %s", audiobookID)

	return c.JSON(fiber.Map{
		"audiobook_id":      audiobookID,
		"message":           "Summarize and tag jobs triggered successfully",
		"job_id":            job.ID,
		"chapters_verified": len(chapters),
	})
}

// GetJobStatus returns the status of processing jobs for an audiobook
// GET /v1/admin/audiobooks/{id}/jobs
func (h *Handler) GetJobStatus(c *fiber.Ctx) error {
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	userID := uuid.MustParse(userCtx.ID)

	// Get audiobook to check ownership
	audiobook, err := h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audio book not found",
		})
	}

	// Check if user owns this audiobook or is admin
	if audiobook.CreatedBy != userID {
		// Check if user is admin
		if userCtx.Role != "admin" {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}
	}

	// Get processing jobs
	jobs, err := h.repo.GetProcessingJobsByAudioBookID(context.Background(), audiobookID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get processing jobs",
		})
	}

	// Calculate overall progress
	totalJobs := len(jobs)
	completedJobs := 0
	failedJobs := 0

	for _, job := range jobs {
		if job.Status == models.JobStatusCompleted {
			completedJobs++
		} else if job.Status == models.JobStatusFailed {
			failedJobs++
		}
	}

	var progress float64
	if totalJobs > 0 {
		progress = float64(completedJobs) / float64(totalJobs)
	}

	// Determine overall status
	var overallStatus models.AudioBookStatus
	if failedJobs > 0 {
		overallStatus = models.StatusFailed
	} else if completedJobs == totalJobs {
		overallStatus = models.StatusCompleted
	} else {
		overallStatus = models.StatusProcessing
	}

	// Update audiobook status if needed
	if overallStatus != audiobook.Status {
		h.repo.UpdateAudioBookStatus(context.Background(), audiobookID, overallStatus)
	}

	return c.JSON(fiber.Map{
		"audiobook_id":   audiobookID,
		"jobs":           jobs,
		"overall_status": overallStatus,
		"progress":       progress,
		"total_jobs":     totalJobs,
		"completed_jobs": completedJobs,
		"failed_jobs":    failedJobs,
	})
}

// UpdateJobStatus updates the status of a processing job (called by workers)
// POST /v1/admin/jobs/{job_id}/status
func (h *Handler) UpdateJobStatus(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("job_id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid job ID",
		})
	}

	var req struct {
		Status       models.JobStatus `json:"status" validate:"required"`
		ErrorMessage *string          `json:"error_message,omitempty"`
		StartedAt    *time.Time       `json:"started_at,omitempty"`
		CompletedAt  *time.Time       `json:"completed_at,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get the job
	job, err := h.repo.GetProcessingJobByID(context.Background(), jobID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	// Update job status
	job.Status = req.Status
	job.ErrorMessage = req.ErrorMessage
	job.StartedAt = req.StartedAt
	job.CompletedAt = req.CompletedAt

	if err := h.repo.UpdateProcessingJob(context.Background(), job); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update job status",
		})
	}

	// Check and update audiobook status
	if err := h.repo.CheckAndUpdateAudioBookStatus(context.Background(), job.AudiobookID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to update audiobook status: %v\n", err)
	}

	return c.JSON(fiber.Map{
		"job_id":  jobID,
		"status":  req.Status,
		"message": "Job status updated successfully",
	})
}
