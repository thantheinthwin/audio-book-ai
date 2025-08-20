package handlers

import (
	"audio-book-ai/api/database"
	"audio-book-ai/api/models"
	"audio-book-ai/api/services"
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler struct holds all handlers with their dependencies
type Handler struct {
	repo       database.Repository
	storage    *services.SupabaseStorageService
	redisQueue *services.RedisQueueService
}

// NewHandler creates a new handler instance with dependencies
func NewHandler(repo database.Repository, storage *services.SupabaseStorageService, redisQueue *services.RedisQueueService) *Handler {
	return &Handler{
		repo:       repo,
		storage:    storage,
		redisQueue: redisQueue,
	}
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
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	userID := uuid.MustParse(userCtx.ID)

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

	// Create processing jobs and enqueue them to Redis
	jobTypes := []models.JobType{
		models.JobTypeTranscribe,
		models.JobTypeSummarize,
		models.JobTypeTag,
		models.JobTypeEmbed,
	}

	jobsCreated := 0
	for _, jobType := range jobTypes {
		job := &models.ProcessingJob{
			ID:          uuid.New(),
			AudiobookID: audiobook.ID,
			JobType:     jobType,
			Status:      models.JobStatusPending,
			CreatedAt:   time.Now(),
		}

		// Save job to database
		if err := h.repo.CreateProcessingJob(context.Background(), job); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create processing job",
			})
		}

		// Enqueue job to Redis if Redis service is available
		if h.redisQueue != nil {
			var enqueueErr error
			if jobType == models.JobTypeTranscribe {
				// For transcription jobs, pass the file path
				enqueueErr = h.redisQueue.EnqueueTranscriptionJob(context.Background(), job, audiobook.FilePath)
			} else {
				// For other AI jobs, don't pass file path
				enqueueErr = h.redisQueue.EnqueueAIJob(context.Background(), job)
			}

			if enqueueErr != nil {
				log.Printf("Failed to enqueue job %s to Redis: %v", jobType, enqueueErr)
				// Continue with other jobs even if one fails to enqueue
			} else {
				jobsCreated++
			}
		} else {
			log.Printf("Redis queue service not available, job %s created in database only", jobType)
			jobsCreated++
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
		"jobs_created": jobsCreated,
		"total_jobs":   len(jobTypes),
	})
}

// CreateUpload creates a new upload session
// POST /v1/admin/uploads
func (h *Handler) CreateUpload(c *fiber.Ctx) error {
	log.Printf("CreateUpload: Request received from IP %s", c.IP())
	log.Printf("CreateUpload: User agent: %s", c.Get("User-Agent"))

	// Log raw request body for debugging
	body := c.Body()
	log.Printf("CreateUpload: Raw request body: %s", string(body))

	var req models.CreateUploadRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("CreateUpload: Failed to parse request body: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	log.Printf("CreateUpload: Request parsed successfully - UploadType: %s, TotalFiles: %d, TotalSize: %d",
		req.UploadType, req.TotalFiles, req.TotalSize)

	// Log individual field values for debugging validation
	log.Printf("CreateUpload: Field values - UploadType: '%v', TotalFiles: %v, TotalSize: %v",
		req.UploadType, req.TotalFiles, req.TotalSize)

	if err := req.Validate(); err != nil {
		log.Printf("CreateUpload: Validation failed: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Validation error: %v", err),
		})
	}

	log.Printf("CreateUpload: Request validation passed")

	// Get user ID from context (set by auth middleware)
	userCtx, ok := c.Locals("user").(*models.UserContext)
	log.Printf("CreateUpload: User context: %v", userCtx)
	if !ok || userCtx == nil {
		log.Printf("CreateUpload: User not authenticated - user context not found in context")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	log.Printf("CreateUpload: User authenticated - UserID: %s", userCtx.ID)

	// Handle case where TotalSize is 0 or not provided
	totalSize := req.TotalSize
	if totalSize <= 0 {
		log.Printf("CreateUpload: TotalSize is 0 or not provided, will be calculated during file uploads")
		totalSize = 0 // Will be updated as files are uploaded
	}

	upload := &models.Upload{
		ID:            uuid.New(),
		UserID:        uuid.MustParse(userCtx.ID),
		UploadType:    req.UploadType,
		Status:        models.UploadStatusPending,
		TotalFiles:    req.TotalFiles,
		UploadedFiles: 0,
		TotalSize:     totalSize,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	log.Printf("CreateUpload: Creating upload session - UploadID: %s", upload.ID)

	if err := h.repo.CreateUpload(context.Background(), upload); err != nil {
		log.Printf("CreateUpload: Failed to create upload in database: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create upload session",
		})
	}

	log.Printf("CreateUpload: Upload session created successfully - UploadID: %s, Status: %s",
		upload.ID, upload.Status)

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
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	userID := uuid.MustParse(userCtx.ID)

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

	// Validate file type
	if err := h.storage.ValidateFileType(file.Filename); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Generate unique filename
	fileExt := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), fileExt)

	// Upload file to Supabase Storage
	fileURL, err := h.storage.UploadFile(file, uploadID.String(), fileName)
	if err != nil {
		log.Printf("UploadFile: Failed to upload file to Supabase Storage: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload file to storage",
		})
	}

	// Create upload file record
	uploadFile := &models.UploadFile{
		ID:            uuid.New(),
		UploadID:      uploadID,
		FileName:      file.Filename,
		FileSize:      file.Size,
		MimeType:      file.Header.Get("Content-Type"),
		FilePath:      fileURL, // Store the Supabase Storage URL
		ChapterNumber: chapterNumber,
		ChapterTitle:  &chapterTitle,
		Status:        models.UploadStatusCompleted,
		CreatedAt:     time.Now(),
	}

	if err := h.repo.CreateUploadFile(context.Background(), uploadFile); err != nil {
		// Clean up file from Supabase Storage if database insert fails
		if err := h.storage.DeleteFile(uploadID.String(), fileName); err != nil {
			log.Printf("UploadFile: Failed to delete file from storage after database error: %v", err)
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file metadata",
		})
	}

	// Update upload session
	upload.UploadedFiles++
	upload.TotalSize += file.Size // Add file size to total
	upload.UpdatedAt = time.Now()
	log.Printf("UploadFile: Updated upload session - UploadedFiles: %d/%d, TotalSize: %d bytes",
		upload.UploadedFiles, upload.TotalFiles, upload.TotalSize)

	if upload.UploadedFiles >= upload.TotalFiles {
		upload.Status = models.UploadStatusCompleted
		log.Printf("UploadFile: Upload session completed")
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
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	userID := uuid.MustParse(userCtx.ID)

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
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	userID := uuid.MustParse(userCtx.ID)

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
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	userID := uuid.MustParse(userCtx.ID)

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

	// Delete files from Supabase Storage
	for _, file := range files {
		// Extract filename from the file path (which is now a URL)
		fileName := filepath.Base(file.FilePath)
		if err := h.storage.DeleteFile(uploadID.String(), fileName); err != nil {
			// Log error but continue with deletion
			log.Printf("Failed to delete file %s from storage: %v", file.FilePath, err)
		}
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
