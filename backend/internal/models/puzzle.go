package models

// Puzzle represents a rebus puzzle with image, answer, and hint
type Puzzle struct {
	ID        string `json:"id"`       // Unique identifier: "YYYY-MM-DD-index"
	ImageURL  string `json:"imageUrl"` // URL to puzzle image (relative or absolute)
	ImagePath string `json:"-"`        // Local file path to the stored image
	Answer    string `json:"answer"`   // Correct answer (lowercase)
	Hint      string `json:"hint"`     // Hint for the puzzle
	Date      string `json:"date"`     // Date in YYYY-MM-DD format
	Index     int    `json:"index"`    // Puzzle number (0-4)
}

// VerifyRequest represents a request to verify an answer
type VerifyRequest struct {
	PuzzleID string `json:"puzzleId"`
	Answer   string `json:"answer"`
}

// VerifyResponse represents the response to a verification request
type VerifyResponse struct {
	Correct bool `json:"correct"`
}

// PuzzlesResponse represents the response containing puzzles for a date
type PuzzlesResponse struct {
	Date    string   `json:"date"`
	Puzzles []Puzzle `json:"puzzles"`
}
