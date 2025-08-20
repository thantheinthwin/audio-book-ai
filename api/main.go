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
	var repo database.Repository
	if cfg.DatabaseURL != "" {
		// Use PostgreSQL repository if DATABASE_URL is provided
		postgresRepo, err := database.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Printf("Warning: Failed to initialize PostgreSQL repository: %v", err)
			log.Printf("Falling back to mock repository")
			repo = database.NewMockRepository()
		} else {
			log.Println("PostgreSQL repository initialized")
			repo = postgresRepo
		}
	} else {
		// Use mock repository if no DATABASE_URL is provided
		log.Println("No DATABASE_URL provided, using mock repository")
		repo = database.NewMockRepository()
	}

	// Initialize Supabase storage service
	storageService := services.NewSupabaseStorageService(cfg)

	// Check if the storage bucket exists
	if err := storageService.CheckBucketExists(); err != nil {
		log.Printf("Warning: Supabase storage bucket check failed: %v", err)
		log.Printf("File uploads may fail. Please ensure the bucket '%s' exists in your Supabase project.", cfg.SupabaseStorageBucket)
	} else {
		log.Println("Supabase storage bucket is accessible")
	}

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

	// Uncomment the next line to run the storage example
	// ExampleStorage()

	log.Printf("🚀 Server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
