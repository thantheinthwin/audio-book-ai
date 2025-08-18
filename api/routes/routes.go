package routes

import (
	"audio-book-ai/api/config"
	"audio-book-ai/api/handlers"
	"audio-book-ai/api/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all API routes
func SetupRoutes(app fiber.Router, cfg *config.Config) {
	// Auth routes (no authentication required)
	auth := app.Group("/auth")
	SetupAuthRoutes(auth, cfg)

	// Protected routes (authentication required)
	protected := app.Group("/", middleware.AuthMiddleware(cfg))
	SetupProtectedRoutes(protected, cfg)

	// Optional auth routes (authentication optional)
	optional := app.Group("/", middleware.OptionalAuthMiddleware(cfg))
	SetupOptionalRoutes(optional, cfg)
}

// SetupAuthRoutes configures authentication routes
func SetupAuthRoutes(router fiber.Router, cfg *config.Config) {
	router.Post("/validate", handlers.ValidateToken)
	router.Get("/me", handlers.Me)
	router.Get("/health", handlers.HealthCheck)
}

// SetupProtectedRoutes configures protected routes
func SetupProtectedRoutes(router fiber.Router, cfg *config.Config) {
	// User profile
	router.Get("/profile", handlers.GetProfile)
	router.Put("/profile", handlers.UpdateProfile)
	router.Delete("/profile", handlers.DeleteProfile)

	// Audio books
	router.Get("/audiobooks", handlers.GetAudioBooks)
	router.Post("/audiobooks", handlers.CreateAudioBook)
	router.Get("/audiobooks/:id", handlers.GetAudioBook)
	router.Put("/audiobooks/:id", handlers.UpdateAudioBook)
	router.Delete("/audiobooks/:id", handlers.DeleteAudioBook)

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

// SetupOptionalRoutes configures optional auth routes
func SetupOptionalRoutes(router fiber.Router, cfg *config.Config) {
	// Public audio books
	router.Get("/public/audiobooks", handlers.GetPublicAudioBooks)
	router.Get("/public/audiobooks/:id", handlers.GetPublicAudioBook)
}
