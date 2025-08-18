package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the worker service
type Config struct {
	// Database
	DatabaseURL string

	// Redis
	RedisURL   string
	JobsPrefix string

	// Gemini API
	GeminiAPIKey string
	GeminiURL    string
	GeminiModel  string

	// Processing
	MaxConcurrentJobs int
	JobPollInterval   int // seconds
	JobTimeout        int // seconds
}

// New creates a new Config instance
func New() *Config {
	return &Config{
		// Database
		DatabaseURL: getEnv("DATABASE_URL", ""),

		// Redis
		RedisURL:   getEnv("REDIS_URL", "redis://redis:6379/0"),
		JobsPrefix: getEnv("JOBS_PREFIX", "audiobooks"),

		// Gemini API
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		GeminiURL:    getEnv("GEMINI_URL", "https://generativelanguage.googleapis.com/v1beta"),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-2.0-flash-exp"),

		// Processing
		MaxConcurrentJobs: getEnvAsInt("MAX_CONCURRENT_JOBS", 5),
		JobPollInterval:   getEnvAsInt("JOB_POLL_INTERVAL", 5),
		JobTimeout:        getEnvAsInt("JOB_TIMEOUT", 300), // 5 minutes
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := parseInt(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// parseInt is a helper function to parse string to int
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.GeminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is required")
	}
	return nil
}
