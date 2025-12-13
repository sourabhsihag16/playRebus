package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Port            string
	StoragePath     string
	ImagesPath      string
	DatabaseURL     string // PostgreSQL connection string
	AIAPIKey        string
	AIAPIURL        string
	ClaudeAPIKey    string // Claude API key for generating prompts
	ReplicateAPIKey string // Replicate API key for image generation
	Environment     string // "local" or "production"
	BatchJobHour    int    // Hour of day to run batch job (0-23)
	BatchJobMinute  int    // Minute of hour to run batch job (0-59)
	AllowedOrigins  []string
	// Supabase S3 Configuration
	SupabaseS3Bucket    string // S3 bucket name
	SupabaseS3Region    string // S3 region
	SupabaseS3AccessKey string // S3 access key
	SupabaseS3SecretKey string // S3 secret key
	SupabaseS3Endpoint  string // S3 endpoint URL (Supabase storage endpoint)
	SupabaseS3PublicURL string // Public URL base for accessing images
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./storage"
	}

	imagesPath := os.Getenv("IMAGES_PATH")
	if imagesPath == "" {
		imagesPath = "./storage/images"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Default PostgreSQL connection string
		// Format: postgres://user:password@host:port/dbname?sslmode=disable
		databaseURL = "postgres://postgres:postgres@localhost:5432/rebus_puzzles?sslmode=disable"
	}

	// Default batch job time: 6:00 AM
	batchHour := 6
	batchMinute := 0

	// You can override with environment variables if needed
	// BATCH_JOB_HOUR and BATCH_JOB_MINUTE

	allowedOrigins := []string{
		"http://localhost:5173",
		"http://localhost:3000",
	}

	// Allow additional origins from environment
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		// Could parse comma-separated origins here
	}

	// Get environment type (local or production)
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "local" // Default to local
	}

	return &Config{
		Port:            port,
		StoragePath:     storagePath,
		ImagesPath:      imagesPath,
		DatabaseURL:     databaseURL,
		AIAPIKey:        os.Getenv("AI_API_KEY"),
		AIAPIURL:        os.Getenv("AI_API_URL"),
		ClaudeAPIKey:    os.Getenv("CLAUDE_API_KEY"),
		ReplicateAPIKey: os.Getenv("REPLICATE_API_KEY"),
		Environment:     environment,
		BatchJobHour:    batchHour,
		BatchJobMinute:  batchMinute,
		AllowedOrigins:  allowedOrigins,
		// Supabase S3 Configuration
		SupabaseS3Bucket:    os.Getenv("SUPABASE_S3_BUCKET"),
		SupabaseS3Region:    os.Getenv("SUPABASE_S3_REGION"),
		SupabaseS3AccessKey: os.Getenv("SUPABASE_S3_ACCESS_KEY"),
		SupabaseS3SecretKey: os.Getenv("SUPABASE_S3_SECRET_KEY"),
		SupabaseS3Endpoint:  os.Getenv("SUPABASE_S3_ENDPOINT"),
		SupabaseS3PublicURL: os.Getenv("SUPABASE_S3_PUBLIC_URL"),
	}
}
