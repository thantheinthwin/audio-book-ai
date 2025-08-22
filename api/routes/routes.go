package routes

import (
	"audio-book-ai/api/config"
	"audio-book-ai/api/database"
	"audio-book-ai/api/handlers"
	"audio-book-ai/api/middleware"
	"audio-book-ai/api/services"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all API routes
func SetupRoutes(app fiber.Router, cfg *config.Config, repo database.Repository, storage *services.SupabaseStorageService, redisQueue *services.RedisQueueService) {
	// Create handler instance
	h := handlers.NewHandler(repo, storage, redisQueue)

	// Internal routes (API key authentication required, for service-to-service communication)
	internal := app.Group("/internal", middleware.InternalAPIKeyMiddleware(cfg))
	SetupInternalRoutes(internal, cfg, h)

	// Auth routes
	auth := app.Group("/auth", middleware.AuthMiddleware(cfg))
	SetupAuthRoutes(auth, cfg)

	// Protected routes (authentication required)
	protected := app.Group("/user", middleware.AuthMiddleware(cfg))
	SetupProtectedRoutes(protected, cfg, h)

	// Admin routes (admin authentication required)
	admin := app.Group("/admin", middleware.AuthMiddleware(cfg), middleware.RequireAdmin())
	SetupAdminRoutes(admin, cfg, h)
}

// SetupAuthRoutes configures authentication routes
func SetupAuthRoutes(router fiber.Router, cfg *config.Config) {
	router.Post("/validate", handlers.ValidateToken)
	router.Get("/me", handlers.Me)
	router.Get("/health", handlers.HealthCheck)
}

// SetupProtectedRoutes configures protected routes (requires user or admin role)
func SetupProtectedRoutes(router fiber.Router, cfg *config.Config, h *handlers.Handler) {
	// Apply user role middleware (allows both user and admin roles)
	router.Use(middleware.RequireUser())

	// User profile
	router.Get("/profile", handlers.GetProfile)
	router.Put("/profile", handlers.UpdateProfile)
	router.Delete("/profile", handlers.DeleteProfile)

	// Audio books (user operations)
	router.Get("/audiobooks", h.GetAudioBooks)
	router.Get("/audiobooks/:id", h.GetAudioBook)

	// Progress
	router.Get("/progress/:audiobookId", handlers.GetProgress)
	router.Put("/progress/:audiobookId", handlers.UpdateProgress)
}

// SetupAdminRoutes configures admin-only routes
func SetupAdminRoutes(router fiber.Router, cfg *config.Config, h *handlers.Handler) {
	// Upload operations
	router.Post("/uploads", h.CreateUpload)
	router.Post("/uploads/:id/files", h.UploadFile)
	router.Post("/uploads/:id/files/batch", h.UploadFilesBatch)
	router.Post("/uploads/:id/files/:file_id/retry", h.RetryFailedUpload)
	router.Get("/uploads/:id/progress", h.GetUploadProgress)
	router.Get("/uploads/:id", h.GetUploadDetails)
	router.Delete("/uploads/:id", h.DeleteUpload)

	// Audio books (admin operations)
	router.Post("/audiobooks", h.CreateAudioBook)
	router.Put("/audiobooks/:id", h.UpdateAudioBook)
	router.Delete("/audiobooks/:id", h.DeleteAudioBook)
	router.Get("/audiobooks/:id/jobs", h.GetJobStatus)
	router.Post("/audiobooks/:id/trigger-summarize-tag", h.TriggerSummarizeAndTagJobs)

	// Job management
	router.Post("/jobs/:job_id/status", h.UpdateJobStatus)
}

// SetupInternalRoutes configures internal service-to-service routes (API key authentication required)
func SetupInternalRoutes(router fiber.Router, cfg *config.Config, h *handlers.Handler) {
	// Internal webhook for triggering summarize and tag jobs
	router.Post("/audiobooks/:id/trigger-summarize-tag", h.TriggerSummarizeAndTagJobs)

	// Internal job status updates
	router.Post("/jobs/:job_id/status", h.UpdateJobStatus)
}
