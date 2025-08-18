package handlers

import (
	"audio-book-ai/api/database"
	"audio-book-ai/api/models"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler struct holds all handlers with their dependencies
type Handler struct {
	repo database.Repository
}

// NewHandler creates a new handler instance with dependencies
func NewHandler(repo database.Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateAudioBook creates an audio book from a completed upload session
// POST /v1/admin/audiobooks
func (h *Handler) CreateAudioBook(c *fiber.Ctx) error {
	var req models.CreateAudioBookFromUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Validation error: %v", err),
		})
	}

	// Get user ID from context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get upload session
	upload, err := h.repo.GetUploadByID(context.Background(), req.UploadID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Upload session not found",
		})
	}

	// Check if user owns this upload
	if upload.UserID != userID {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Check if upload is completed
	if upload.Status != models.UploadStatusCompleted {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Upload session is not completed",
		})
	}

	// Get upload files
	uploadFiles, err := h.repo.GetUploadFiles(context.Background(), req.UploadID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get upload files",
		})
	}

	if len(uploadFiles) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No files found in upload session",
		})
	}

	// Create audio book
	audiobook := &models.AudioBook{
		ID:          uuid.New(),
		Title:       req.Title,
		Author:      req.Author,
		Description: req.Description,
		Language:    req.Language,
		IsPublic:    req.IsPublic,
		Status:      models.StatusPending,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Handle different upload types
	if upload.UploadType == models.UploadTypeSingle {
		// Single file upload
		if len(uploadFiles) != 1 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Single upload type requires exactly one file",
			})
		}

		file := uploadFiles[0]
		audiobook.FilePath = file.FilePath
		audiobook.FileSizeBytes = &file.FileSize
		audiobook.DurationSeconds = nil

	} else if upload.UploadType == models.UploadTypeChapters {
		// Chaptered upload
		if len(uploadFiles) == 0 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "No chapter files found",
			})
		}

		// Sort files by chapter number
		chapterFiles := make(map[int]*models.UploadFile)
		var totalSize int64
		for _, file := range uploadFiles {
			if file.ChapterNumber == nil {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "All files must have chapter numbers for chaptered uploads",
				})
			}
			chapterFiles[*file.ChapterNumber] = &file
			totalSize += file.FileSize
		}

		// Use the first chapter file as the main file path
		firstChapter := chapterFiles[1]
		if firstChapter == nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Chapter 1 is required",
			})
		}

		audiobook.FilePath = firstChapter.FilePath
		audiobook.FileSizeBytes = &totalSize
		audiobook.DurationSeconds = nil

		// Create chapter records
		chapters := make([]*models.Chapter, 0, len(chapterFiles))
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
				CreatedAt:     time.Now(),
			}
			chapters = append(chapters, chapter)
		}

		// Save chapters to database
		for _, chapter := range chapters {
			if err := h.repo.CreateChapter(context.Background(), chapter); err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create chapter",
				})
			}
		}
	}

	// Save audio book to database
	if err := h.repo.CreateAudioBook(context.Background(), audiobook); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create audio book",
		})
	}

	// Create processing jobs
	jobTypes := []models.JobType{
		models.JobTypeTranscribe,
		models.JobTypeSummarize,
		models.JobTypeTag,
		models.JobTypeEmbed,
	}

	for _, jobType := range jobTypes {
		job := &models.ProcessingJob{
			ID:          uuid.New(),
			AudiobookID: audiobook.ID,
			JobType:     jobType,
			Status:      models.JobStatusPending,
			CreatedAt:   time.Now(),
		}

		if err := h.repo.CreateProcessingJob(context.Background(), job); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create processing job",
			})
		}
	}

	// Update upload status to indicate it's been processed
	upload.Status = "processed"
	upload.UpdatedAt = time.Now()
	if err := h.repo.UpdateUpload(context.Background(), upload); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to update upload status: %v\n", err)
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"audiobook_id": audiobook.ID,
		"status":       audiobook.Status,
		"message":      "Audio book created successfully",
		"jobs_created": len(jobTypes),
	})
}

// CreateUpload creates a new upload session
// POST /v1/admin/uploads
func (h *Handler) CreateUpload(c *fiber.Ctx) error {
	var req models.CreateUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Validation error: %v", err),
		})
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	upload := &models.Upload{
		ID:            uuid.New(),
		UserID:        userID,
		UploadType:    req.UploadType,
		Status:        models.UploadStatusPending,
		TotalFiles:    req.TotalFiles,
		UploadedFiles: 0,
		TotalSize:     req.TotalSize,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.repo.CreateUpload(context.Background(), upload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create upload session",
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"upload_id": upload.ID,
		"status":    upload.Status,
		"message":   "Upload session created successfully",
	})
}

// UploadFile uploads a file to an existing upload session
// POST /v1/admin/uploads/:id/files
func (h *Handler) UploadFile(c *fiber.Ctx) error {
	uploadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid upload ID",
		})
	}

	// Get user ID from context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get upload session
	upload, err := h.repo.GetUploadByID(context.Background(), uploadID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Upload session not found",
		})
	}

	// Check if user owns this upload
	if upload.UserID != userID {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided",
		})
	}

	// Parse metadata
	chapterNumberStr := c.FormValue("chapter_number")
	chapterTitle := c.FormValue("chapter_title")

	var chapterNumber *int
	if chapterNumberStr != "" {
		num, err := strconv.Atoi(chapterNumberStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid chapter number",
			})
		}
		chapterNumber = &num
	}

	// Validate file size
	if file.Size > 500*1024*1024 { // 500MB limit
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "File too large (max 500MB)",
		})
	}

	// Create upload directory
	uploadDir := fmt.Sprintf("uploads/%s", uploadID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create upload directory",
		})
	}

	// Generate unique filename
	fileExt := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), fileExt)
	filePath := filepath.Join(uploadDir, fileName)

	// Save file
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	// Create upload file record
	uploadFile := &models.UploadFile{
		ID:            uuid.New(),
		UploadID:      uploadID,
		FileName:      file.Filename,
		FileSize:      file.Size,
		MimeType:      file.Header.Get("Content-Type"),
		FilePath:      filePath,
		ChapterNumber: chapterNumber,
		ChapterTitle:  &chapterTitle,
		Status:        models.UploadStatusCompleted,
		CreatedAt:     time.Now(),
	}

	if err := h.repo.CreateUploadFile(context.Background(), uploadFile); err != nil {
		// Clean up file if database insert fails
		os.Remove(filePath)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file metadata",
		})
	}

	// Update upload session
	upload.UploadedFiles++
	upload.UpdatedAt = time.Now()
	if upload.UploadedFiles >= upload.TotalFiles {
		upload.Status = models.UploadStatusCompleted
	} else {
		upload.Status = models.UploadStatusUploading
	}

	if err := h.repo.UpdateUpload(context.Background(), upload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update upload session",
		})
	}

	return c.Status(http.StatusCreated).JSON(models.FileUploadResponse{
		FileID:        uploadFile.ID,
		UploadID:      uploadID,
		FileName:      uploadFile.FileName,
		FileSize:      uploadFile.FileSize,
		UploadedAt:    uploadFile.CreatedAt,
		ChapterNumber: uploadFile.ChapterNumber,
		ChapterTitle:  uploadFile.ChapterTitle,
	})
}

// GetUploadProgress returns the progress of an upload session
// GET /v1/admin/uploads/:id/progress
func (h *Handler) GetUploadProgress(c *fiber.Ctx) error {
	uploadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid upload ID",
		})
	}

	// Get user ID from context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get upload session
	upload, err := h.repo.GetUploadByID(context.Background(), uploadID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Upload session not found",
		})
	}

	// Check if user owns this upload
	if upload.UserID != userID {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Calculate uploaded size
	uploadedSize, err := h.repo.GetUploadedSize(context.Background(), uploadID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate upload progress",
		})
	}

	progress := upload.GetProgress()

	return c.JSON(models.UploadProgressResponse{
		UploadID:      uploadID,
		Status:        upload.Status,
		TotalFiles:    upload.TotalFiles,
		UploadedFiles: upload.UploadedFiles,
		Progress:      progress,
		TotalSize:     upload.TotalSize,
		UploadedSize:  uploadedSize,
	})
}

// GetUploadDetails returns detailed information about an upload session
// GET /v1/admin/uploads/:id
func (h *Handler) GetUploadDetails(c *fiber.Ctx) error {
	uploadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid upload ID",
		})
	}

	// Get user ID from context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get upload session
	upload, err := h.repo.GetUploadByID(context.Background(), uploadID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Upload session not found",
		})
	}

	// Check if user owns this upload
	if upload.UserID != userID {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Get upload files
	files, err := h.repo.GetUploadFiles(context.Background(), uploadID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get upload files",
		})
	}

	// Convert to response format
	fileInfos := make([]models.UploadFileInfo, len(files))
	for i, file := range files {
		fileInfos[i] = models.UploadFileInfo{
			ID:            file.ID,
			FileName:      file.FileName,
			FileSize:      file.FileSize,
			MimeType:      file.MimeType,
			ChapterNumber: file.ChapterNumber,
			ChapterTitle:  file.ChapterTitle,
			Status:        file.Status,
			Error:         file.Error,
			UploadedAt:    file.CreatedAt,
		}
	}

	return c.JSON(models.UploadDetailsResponse{
		Upload: *upload,
		Files:  fileInfos,
	})
}

// DeleteUpload deletes an upload session and all associated files
// DELETE /v1/admin/uploads/:id
func (h *Handler) DeleteUpload(c *fiber.Ctx) error {
	uploadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid upload ID",
		})
	}

	// Get user ID from context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get upload session
	upload, err := h.repo.GetUploadByID(context.Background(), uploadID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Upload session not found",
		})
	}

	// Check if user owns this upload
	if upload.UserID != userID {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Get upload files to delete from disk
	files, err := h.repo.GetUploadFiles(context.Background(), uploadID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get upload files",
		})
	}

	// Delete files from disk
	for _, file := range files {
		if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
			// Log error but continue with deletion
			fmt.Printf("Failed to delete file %s: %v\n", file.FilePath, err)
		}
	}

	// Delete upload directory
	uploadDir := fmt.Sprintf("uploads/%s", uploadID)
	if err := os.RemoveAll(uploadDir); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Failed to delete upload directory %s: %v\n", uploadDir, err)
	}

	// Delete from database
	if err := h.repo.DeleteUpload(context.Background(), uploadID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete upload session",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Upload session deleted successfully",
	})
}

// GetJobStatus returns the status of processing jobs for an audio book
// GET /v1/admin/audiobooks/:id/jobs
func (h *Handler) GetJobStatus(c *fiber.Ctx) error {
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audio book ID",
		})
	}

	// Get user ID from context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get audio book to check ownership
	audiobook, err := h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audio book not found",
		})
	}

	if audiobook.CreatedBy != userID {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Get processing jobs
	jobs, err := h.repo.GetProcessingJobsByAudioBookID(context.Background(), audiobookID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get processing jobs",
		})
	}

	// Calculate overall status and progress
	completedJobs := 0
	failedJobs := 0
	for _, job := range jobs {
		if job.Status == models.JobStatusCompleted {
			completedJobs++
		} else if job.Status == models.JobStatusFailed {
			failedJobs++
		}
	}

	var overallStatus models.AudioBookStatus
	var progress float64

	if len(jobs) == 0 {
		overallStatus = models.StatusPending
		progress = 0.0
	} else if failedJobs > 0 {
		overallStatus = models.StatusFailed
		progress = float64(completedJobs) / float64(len(jobs))
	} else if completedJobs == len(jobs) {
		overallStatus = models.StatusCompleted
		progress = 1.0
	} else {
		overallStatus = models.StatusProcessing
		progress = float64(completedJobs) / float64(len(jobs))
	}

	return c.JSON(models.JobStatusResponse{
		AudiobookID:   audiobookID,
		Jobs:          jobs,
		OverallStatus: overallStatus,
		Progress:      progress,
	})
}
