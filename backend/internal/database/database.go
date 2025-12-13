package database

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"

	_ "github.com/lib/pq"
)

// DB wraps the database connection
type DB struct {
	*sql.DB
}

// NewDB creates a new database connection
func NewDB(connectionString string) (*DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &DB{DB: db}

	// Initialize schema
	if err := database.InitSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// InitSchema creates the necessary tables if they don't exist
func (db *DB) InitSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS puzzles (
		id VARCHAR(50) PRIMARY KEY,
		date VARCHAR(10) NOT NULL,
		index_num INTEGER NOT NULL,
		image_url VARCHAR(255) NOT NULL,
		image_path VARCHAR(500) NOT NULL,
		answer VARCHAR(255) NOT NULL,
		hint TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(date, index_num)
	);

	CREATE INDEX IF NOT EXISTS idx_puzzles_date ON puzzles(date);
	CREATE INDEX IF NOT EXISTS idx_puzzles_id ON puzzles(id);
	`

	_, err := db.Exec(query)
	return err
}

// GetPuzzlesForDate retrieves all puzzles for a specific date
func (db *DB) GetPuzzlesForDate(date string) ([]models.Puzzle, error) {
	query := `
		SELECT id, date, index_num, image_url, image_path, answer, hint
		FROM puzzles
		WHERE date = $1
		ORDER BY index_num ASC
	`

	rows, err := db.Query(query, date)
	if err != nil {
		return nil, fmt.Errorf("failed to query puzzles: %w", err)
	}
	defer rows.Close()

	var puzzles []models.Puzzle
	for rows.Next() {
		var p models.Puzzle
		var indexNum int
		if err := rows.Scan(&p.ID, &p.Date, &indexNum, &p.ImageURL, &p.ImagePath, &p.Answer, &p.Hint); err != nil {
			return nil, fmt.Errorf("failed to scan puzzle: %w", err)
		}
		p.Index = indexNum
		puzzles = append(puzzles, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating puzzles: %w", err)
	}

	return puzzles, nil
}

// SavePuzzle saves a single puzzle to the database
func (db *DB) SavePuzzle(puzzle *models.Puzzle) error {
	query := `
		INSERT INTO puzzles (id, date, index_num, image_url, image_path, answer, hint, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) 
		DO UPDATE SET 
			image_url = EXCLUDED.image_url,
			image_path = EXCLUDED.image_path,
			answer = EXCLUDED.answer,
			hint = EXCLUDED.hint
	`

	_, err := db.Exec(query,
		puzzle.ID,
		puzzle.Date,
		puzzle.Index,
		puzzle.ImageURL,
		puzzle.ImagePath,
		puzzle.Answer,
		puzzle.Hint,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save puzzle: %w", err)
	}

	return nil
}

// SavePuzzles saves multiple puzzles for a date (transactional)
func (db *DB) SavePuzzles(date string, puzzles []models.Puzzle) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing puzzles for this date
	deleteQuery := `DELETE FROM puzzles WHERE date = $1`
	if _, err := tx.Exec(deleteQuery, date); err != nil {
		return fmt.Errorf("failed to delete existing puzzles: %w", err)
	}

	// Insert new puzzles
	insertQuery := `
		INSERT INTO puzzles (id, date, index_num, image_url, image_path, answer, hint, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, puzzle := range puzzles {
		_, err := stmt.Exec(
			puzzle.ID,
			puzzle.Date,
			puzzle.Index,
			puzzle.ImageURL,
			puzzle.ImagePath,
			puzzle.Answer,
			puzzle.Hint,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert puzzle %s: %w", puzzle.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// HasPuzzlesForDate checks if puzzles exist for a date
func (db *DB) HasPuzzlesForDate(date string) (bool, error) {
	query := `SELECT COUNT(*) FROM puzzles WHERE date = $1`
	var count int
	err := db.QueryRow(query, date).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check puzzles: %w", err)
	}
	return count > 0, nil
}

// GetPuzzleByID retrieves a puzzle by its ID
func (db *DB) GetPuzzleByID(id string) (*models.Puzzle, error) {
	query := `
		SELECT id, date, index_num, image_url, image_path, answer, hint
		FROM puzzles
		WHERE id = $1
	`

	var p models.Puzzle
	var indexNum int
	err := db.QueryRow(query, id).Scan(&p.ID, &p.Date, &indexNum, &p.ImageURL, &p.ImagePath, &p.Answer, &p.Hint)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("puzzle not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get puzzle: %w", err)
	}

	p.Index = indexNum
	return &p, nil
}
