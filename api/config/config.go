package config

import (
	"os"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Port        string
	Environment string

	// Supabase
	SupabaseURL            string
	SupabasePublishableKey string
	SupabaseSecretKey      string
	SupabaseJWKSURL        string
	SupabaseStorageBucket  string
	JWTAudience            string

	// Redis
	RedisURL   string
	JobsPrefix string

	// Ollama
	OllamaURL      string
	AISummaryModel string
	AIEmbedModel   string

	// Whisper
	WhisperModel    string
	WhisperLanguage string

	// CORS
	CORSOrigins []string
}

// New creates a new Config instance
func New() *Config {
	return &Config{
		// Server
		Port:        getEnv("API_PORT", "8080"),
		Environment: getEnv("NODE_ENV", "development"),

		// Supabase
		SupabaseURL:            getEnv("SUPABASE_URL", ""),
		SupabasePublishableKey: getEnv("SUPABASE_PUBLISHABLE_KEY", ""),
		SupabaseSecretKey:      getEnv("SUPABASE_SECRET_KEY", ""),
		SupabaseJWKSURL:        getEnv("SUPABASE_JWKS_URL", ""),
		SupabaseStorageBucket:  getEnv("SUPABASE_STORAGE_BUCKET", "audio"),
		JWTAudience:            getEnv("JWT_AUDIENCE", "authenticated"),

		// Redis
		RedisURL:   getEnv("REDIS_URL", "redis://redis:6379/0"),
		JobsPrefix: getEnv("JOBS_PREFIX", "audiobooks"),

		// Ollama
		OllamaURL:      getEnv("OLLAMA_URL", "http://ollama:11434"),
		AISummaryModel: getEnv("AI_SUMMARY_MODEL", "llama3.1"),
		AIEmbedModel:   getEnv("AI_EMBED_MODEL", "nomic-embed-text"),

		// Whisper
		WhisperModel:    getEnv("WHISPER_MODEL", "base"),
		WhisperLanguage: getEnv("WHISPER_LANGUAGE", "auto"),

		// CORS
		CORSOrigins: parseCORSOrigins(getEnv("CORS_ORIGIN", "http://localhost:3000")),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseCORSOrigins parses comma-separated CORS origins
func parseCORSOrigins(origins string) []string {
	if origins == "" {
		return []string{"http://localhost:3000"}
	}
	return strings.Split(origins, ",")
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
