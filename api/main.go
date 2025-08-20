package main

import (
	"log"
	"strings"

	"audio-book-ai/api/config"
	"audio-book-ai/api/database"
	"audio-book-ai/api/handlers"
	"audio-book-ai/api/routes"
	"audio-book-ai/api/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

// @title Audio Book AI API
// @version 1.0
// @description Audio Book AI REST API with Supabase authentication
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Audio Book AI API",
		ErrorHandler: handlers.ErrorHandler,
		BodyLimit:    50 * 1024 * 1024, // 50MB limit for file uploads
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.CORSOrigins, ","),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Rate limiting
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * 60, // 1 minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}))

	// Health check endpoint
	// app.Get("/health", handlers.HealthCheck)

	// Initialize database repository
	// For now, we'll use a mock repository until the database implementation is ready
	repo := database.NewMockRepository()

	// Initialize Supabase storage service
	storageService := services.NewSupabaseStorageService(cfg)

	// Initialize Redis queue service
	redisQueue, err := services.NewRedisQueueService(cfg.RedisURL, cfg.JobsPrefix)
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis queue service: %v", err)
		log.Printf("Continuing without Redis queue functionality")
		redisQueue = nil
	} else {
		log.Println("Redis queue service initialized")
	}

	// API routes
	api := app.Group("/api/v1")
	routes.SetupRoutes(api, cfg, repo, storageService, redisQueue)

	log.Printf("🚀 Server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
