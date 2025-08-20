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
	// Auth routes
	auth := app.Group("/auth", middleware.AuthMiddleware(cfg))
	SetupAuthRoutes(auth, cfg)

	// Protected routes (authentication required)
	protected := app.Group("/", middleware.AuthMiddleware(cfg))
	SetupProtectedRoutes(protected, cfg, h)

	// Admin routes (admin authentication required)
	admin := app.Group("/admin", middleware.AuthMiddleware(cfg), middleware.RequireAdmin())
	SetupAdminRoutes(admin, cfg, h)

	// Public routes (no authentication required)
	public := app.Group("/")
	SetupPublicRoutes(public, cfg)
}

// SetupAuthRoutes configures authentication routes
func SetupAuthRoutes(router fiber.Router, cfg *config.Config) {
	router.Post("/validate", handlers.ValidateToken)
	router.Get("/me", handlers.Me)
	router.Get("/health", handlers.HealthCheck)
}

// SetupProtectedRoutes configures protected routes (requires user or admin role)
func SetupProtectedRoutes(router fiber.Router, cfg *config.Config, h *handlers.Handler) {
	// Apply user role middleware
	router.Use(middleware.RequireUser())

	// User profile
	router.Get("/profile", handlers.GetProfile)
	router.Put("/profile", handlers.UpdateProfile)
	router.Delete("/profile", handlers.DeleteProfile)

	// Audio books (user operations)
	router.Get("/audiobooks", handlers.GetAudioBooks)
	router.Get("/audiobooks/:id", handlers.GetAudioBook)

	// Library
	router.Get("/library", handlers.GetLibrary)
	router.Post("/library/:audiobookId", handlers.AddToLibrary)
	router.Delete("/library/:audiobookId", handlers.RemoveFromLibrary)

	// Playlists
	router.Get("/playlists", handlers.GetPlaylists)
	router.Post("/playlists", handlers.CreatePlaylist)
	router.Get("/playlists/:id", handlers.GetPlaylist)
	router.Put("/playlists/:id", handlers.UpdatePlaylist)
	router.Delete("/playlists/:id", handlers.DeletePlaylist)
	router.Post("/playlists/:id/items", handlers.AddToPlaylist)
	router.Delete("/playlists/:id/items/:audiobookId", handlers.RemoveFromPlaylist)

	// Progress
	router.Get("/progress/:audiobookId", handlers.GetProgress)
	router.Put("/progress/:audiobookId", handlers.UpdateProgress)

	// Bookmarks
	router.Get("/bookmarks/:audiobookId", handlers.GetBookmarks)
	router.Post("/bookmarks/:audiobookId", handlers.CreateBookmark)
	router.Put("/bookmarks/:id", handlers.UpdateBookmark)
	router.Delete("/bookmarks/:id", handlers.DeleteBookmark)
}

// SetupAdminRoutes configures admin-only routes
func SetupAdminRoutes(router fiber.Router, cfg *config.Config, h *handlers.Handler) {
	// Upload operations
	router.Post("/uploads", h.CreateUpload)
	router.Post("/uploads/:id/files", h.UploadFile)
	router.Get("/uploads/:id/progress", h.GetUploadProgress)
	router.Get("/uploads/:id", h.GetUploadDetails)
	router.Delete("/uploads/:id", h.DeleteUpload)

	// Audio books (admin operations)
	router.Post("/audiobooks", h.CreateAudioBook)
	router.Put("/audiobooks/:id", handlers.UpdateAudioBook)
	router.Delete("/audiobooks/:id", handlers.DeleteAudioBook)
	router.Get("/audiobooks/:id/jobs", h.GetJobStatus)
}

// SetupPublicRoutes configures public routes (no authentication required)
func SetupPublicRoutes(router fiber.Router, cfg *config.Config) {
	// Public audio books
	router.Get("/audiobooks", handlers.GetPublicAudioBooks)
	router.Get("/audiobooks/:id", handlers.GetPublicAudioBook)
}
