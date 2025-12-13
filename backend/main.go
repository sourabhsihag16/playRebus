package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"backend/internal/ai"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handlers"
	"backend/internal/scheduler"
	"backend/internal/store"
	"fmt"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
	}
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL database")

	// Initialize store - use Supabase if configured, otherwise use file system
	var storeInstance *store.Store
	if cfg.SupabaseS3Bucket != "" && cfg.SupabaseS3AccessKey != "" && cfg.SupabaseS3SecretKey != "" {
		log.Println("Using Supabase S3 storage for images")
		storeInstance, err = store.NewStoreWithSupabase(
			db,
			cfg.SupabaseS3Bucket,
			cfg.SupabaseS3Region,
			cfg.SupabaseS3AccessKey,
			cfg.SupabaseS3SecretKey,
			cfg.SupabaseS3Endpoint,
			cfg.SupabaseS3PublicURL,
		)
		if err != nil {
			log.Fatalf("Failed to initialize Supabase store: %v", err)
		}
	} else {
		log.Println("Using local file system storage for images")
		storeInstance, err = store.NewStore(db, cfg.ImagesPath)
		if err != nil {
			log.Fatalf("Failed to initialize store: %v", err)
		}
	}

	// Initialize AI generator - always use real generator with Claude API and Replicate
	var aiGenerator ai.AIGenerator
	if cfg.ClaudeAPIKey == "" || cfg.ReplicateAPIKey == "" {
		log.Println("WARNING: Missing API keys. Please set CLAUDE_API_KEY and REPLICATE_API_KEY")
		log.Println("The generator will attempt to use Claude API and Replicate but may fail if keys are not set")
	}

	log.Println("Using real AI generator with Claude API and Replicate")
	aiGenerator = ai.NewRealAIGenerator(cfg.ClaudeAPIKey, cfg.ReplicateAPIKey, cfg.Environment)

	// Initialize scheduler
	sched := scheduler.NewScheduler(storeInstance, aiGenerator, cfg.BatchJobHour, cfg.BatchJobMinute)
	sched.Start()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		sched.Stop()
		os.Exit(0)
	}()

	// Initialize handlers
	puzzleHandler := handlers.NewPuzzleHandler(storeInstance, sched)
	var imageHandler *handlers.ImageHandler
	if cfg.SupabaseS3Bucket != "" && cfg.SupabaseS3PublicURL != "" {
		imageHandler = handlers.NewImageHandlerWithSupabase(cfg.SupabaseS3PublicURL)
	} else {
		imageHandler = handlers.NewImageHandler(cfg.ImagesPath)
	}

	// Setup router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/puzzles/{date}", puzzleHandler.GetPuzzlesHandler).Methods("GET")
	api.HandleFunc("/puzzles/verify", puzzleHandler.VerifyAnswerHandler).Methods("POST")
	api.HandleFunc("/puzzles/trigger", puzzleHandler.TriggerJobHandler).Methods("POST")
	api.HandleFunc("/images/{filename}", imageHandler.ServeImage).Methods("GET")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// CORS middleware
	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins(cfg.AllowedOrigins),
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type"}),
	)(r)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      corsHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("API endpoints:")
	log.Printf("  GET  /api/puzzles/{date} - Get puzzles for a date")
	log.Printf("  POST /api/puzzles/verify - Verify an answer")
	log.Printf("  POST /api/puzzles/trigger - Trigger puzzle generation for today")
	log.Printf("  GET  /api/images/{filename} - Get puzzle image")
	log.Printf("Batch job scheduled to run daily at %02d:%02d", cfg.BatchJobHour, cfg.BatchJobMinute)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

