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

	// Admin routes (admin authentication required)
	admin := app.Group("/admin", middleware.AuthMiddleware(cfg), middleware.RequireAdmin())
	SetupAdminRoutes(admin, cfg)

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
func SetupProtectedRoutes(router fiber.Router, cfg *config.Config) {
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
func SetupAdminRoutes(router fiber.Router, cfg *config.Config) {
	// Audio books (admin operations)
	router.Post("/audiobooks", handlers.CreateAudioBook)
	router.Put("/audiobooks/:id", handlers.UpdateAudioBook)
	router.Delete("/audiobooks/:id", handlers.DeleteAudioBook)

	// User management
	router.Get("/users", handlers.GetUsers)
	router.Get("/users/:id", handlers.GetUser)
	router.Put("/users/:id", handlers.UpdateUser)
	router.Delete("/users/:id", handlers.DeleteUser)
}

// SetupPublicRoutes configures public routes (no authentication required)
func SetupPublicRoutes(router fiber.Router, cfg *config.Config) {
	// Public audio books
	router.Get("/audiobooks", handlers.GetPublicAudioBooks)
	router.Get("/audiobooks/:id", handlers.GetPublicAudioBook)
}
