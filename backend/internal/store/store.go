package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backend/internal/database"
	"backend/internal/models"
)

// Store handles puzzle storage with PostgreSQL for metadata and Supabase S3 or file system for images
type Store struct {
	db              *database.DB
	imagesPath      string
	supabaseStorage *SupabaseStorage
	useSupabase     bool
}

// NewStore creates a new store instance with file system storage
func NewStore(db *database.DB, imagesPath string) (*Store, error) {
	// Create images directory if it doesn't exist
	if err := os.MkdirAll(imagesPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create images directory: %w", err)
	}

	return &Store{
		db:          db,
		imagesPath:  imagesPath,
		useSupabase: false,
	}, nil
}

// NewStoreWithSupabase creates a new store instance with Supabase S3 storage
func NewStoreWithSupabase(db *database.DB, bucketName, region, accessKey, secretKey, endpoint, publicURL string) (*Store, error) {
	supabaseStorage, err := NewSupabaseStorage(bucketName, region, accessKey, secretKey, endpoint, publicURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Supabase storage: %w", err)
	}

	return &Store{
		db:              db,
		supabaseStorage: supabaseStorage,
		useSupabase:     true,
	}, nil
}

// GetPuzzlesForDate returns puzzles for a specific date
func (s *Store) GetPuzzlesForDate(date string) ([]models.Puzzle, error) {
	return s.db.GetPuzzlesForDate(date)
}

// SavePuzzles saves puzzles for a date
func (s *Store) SavePuzzles(date string, puzzles []models.Puzzle) error {
	return s.db.SavePuzzles(date, puzzles)
}

// HasPuzzlesForDate checks if puzzles exist for a date
func (s *Store) HasPuzzlesForDate(date string) bool {
	exists, err := s.db.HasPuzzlesForDate(date)
	if err != nil {
		return false
	}
	return exists
}

// GetImagePath returns the full path where an image should be stored
func (s *Store) GetImagePath(date string, index int) string {
	if s.useSupabase {
		return s.supabaseStorage.GetImagePath(date, index)
	}
	filename := fmt.Sprintf("%s-%d.png", date, index)
	return filepath.Join(s.imagesPath, filename)
}

// GetImageURL returns the URL path for an image
func (s *Store) GetImageURL(date string, index int) string {
	if s.useSupabase {
		return s.supabaseStorage.GetImageURL(date, index)
	}
	return fmt.Sprintf("/api/images/%s-%d.png", date, index)
}

// SaveImage saves image data to disk or Supabase S3
func (s *Store) SaveImage(date string, index int, imageData []byte) error {
	if s.useSupabase {
		fmt.Printf("Saving image to Supabase S3: %s/%d.png\n", date, index)
		return s.supabaseStorage.SaveImage(date, index, imageData)
	}
	imagePath := s.GetImagePath(date, index)
	fmt.Println("Saving image to:", imagePath)
	return os.WriteFile(imagePath, imageData, 0644)
}

// GetTodayDate returns today's date in YYYY-MM-DD format
func GetTodayDate() string {
	return time.Now().Format("2006-01-02")
}

// ValidateDate validates date format (YYYY-MM-DD)
func ValidateDate(date string) error {
	if len(date) != 10 || date[4] != '-' || date[7] != '-' {
		return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
	}

	// Try to parse the date
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}

	return nil
}

// ParsePuzzleID extracts date and index from puzzle ID
func ParsePuzzleID(puzzleID string) (date string, index int, error error) {
	// Format: YYYY-MM-DD-index
	parts := strings.Split(puzzleID, "-")
	if len(parts) < 4 {
		return "", 0, fmt.Errorf("invalid puzzle ID format")
	}

	date = strings.Join(parts[:3], "-")
	if err := ValidateDate(date); err != nil {
		return "", 0, err
	}

	var idx int
	if _, err := fmt.Sscanf(parts[3], "%d", &idx); err != nil {
		return "", 0, fmt.Errorf("invalid index in puzzle ID")
	}

	return date, idx, nil
}
