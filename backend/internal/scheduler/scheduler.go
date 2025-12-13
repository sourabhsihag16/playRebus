package scheduler

import (
	"fmt"
	"log"
	"time"

	"backend/internal/ai"
	"backend/internal/models"
	"backend/internal/store"
)

// Scheduler handles daily batch jobs for puzzle generation
type Scheduler struct {
	store     *store.Store
	generator ai.AIGenerator
	hour      int
	minute    int
	stopChan  chan struct{}
	running   bool
}

// NewScheduler creates a new scheduler
func NewScheduler(store *store.Store, generator ai.AIGenerator, hour, minute int) *Scheduler {
	return &Scheduler{
		store:     store,
		generator: generator,
		hour:      hour,
		minute:    minute,
		stopChan:  make(chan struct{}),
		running:   false,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	if s.running {
		log.Println("Scheduler is already running")
		return
	}

	s.running = true
	log.Printf("Scheduler started. Will run daily at %02d:%02d", s.hour, s.minute)

	// Run immediately if it's past the scheduled time today
	s.checkAndRunIfNeeded()

	// Then schedule for daily execution
	go s.run()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if !s.running {
		return
	}

	s.running = false
	close(s.stopChan)
	log.Println("Scheduler stopped")
}

// run runs the scheduler loop
func (s *Scheduler) run() {
	for {
		// Calculate next run time
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), s.hour, s.minute, 0, 0, now.Location())

		// If the time has passed today, schedule for tomorrow
		if nextRun.Before(now) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		duration := nextRun.Sub(now)
		log.Printf("Next batch job scheduled for: %s (in %v)", nextRun.Format("2006-01-02 15:04:05"), duration)

		// Wait until next run time or stop signal
		select {
		case <-time.After(duration):
			s.generatePuzzlesForToday()
		case <-s.stopChan:
			return
		}
	}
}

// checkAndRunIfNeeded checks if we should run the job now
func (s *Scheduler) checkAndRunIfNeeded() {
	now := time.Now()
	today := store.GetTodayDate()

	// Check if puzzles already exist for today
	if s.store.HasPuzzlesForDate(today) {
		log.Printf("Puzzles already exist for today (%s), skipping generation", today)
		return
	}

	// Check if it's past the scheduled time
	scheduledTime := time.Date(now.Year(), now.Month(), now.Day(), s.hour, s.minute, 0, 0, now.Location())
	if now.After(scheduledTime) {
		log.Printf("Scheduled time (%02d:%02d) has passed, generating puzzles for today", s.hour, s.minute)
		s.generatePuzzlesForToday()
	}
}

// generatePuzzlesForToday generates puzzles for today's date
func (s *Scheduler) generatePuzzlesForToday() {
	today := store.GetTodayDate()
	log.Printf("Starting batch job to generate puzzles for %s", today)

	// Check if puzzles already exist
	if s.store.HasPuzzlesForDate(today) {
		log.Printf("Puzzles already exist for %s, skipping", today)
		return
	}

	if !s.store.HasPuzzlesForDate(today) {
		log.Printf("Puzzles do not exist for %s, not generating", today)
		return
	}
	// Generate all 5 puzzles at once using Claude API
	// This will first call Claude to get 5 prompts, then generate images for each
	puzzlePointers, err := s.generator.GenerateRebusPuzzles(today, s.store)
	if err != nil {
		log.Printf("Error generating puzzles: %v", err)
		return
	}

	// Convert pointers to values
	puzzles := make([]models.Puzzle, len(puzzlePointers))
	for i, p := range puzzlePointers {
		if p != nil {
			puzzles[i] = *p
		}
	}

	// Save puzzles to store
	if err := s.store.SavePuzzles(today, puzzles); err != nil {
		log.Printf("Error saving puzzles: %v", err)
		return
	}

	log.Printf("Successfully generated and saved 5 puzzles for %s", today)
}

// TriggerManualGeneration manually triggers puzzle generation for a specific date
func (s *Scheduler) TriggerManualGeneration(date string) error {
	if err := store.ValidateDate(date); err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}

	// Check if puzzles already exist
	if s.store.HasPuzzlesForDate(date) {
		return fmt.Errorf("puzzles already exist for date: %s", date)
	}

	// Generate all 5 puzzles at once using Claude API
	// This will first call Claude to get 5 prompts, then generate images for each
	puzzlePointers, err := s.generator.GenerateRebusPuzzles(date, s.store)
	if err != nil {
		return fmt.Errorf("failed to generate puzzles: %w", err)
	}

	// Convert pointers to values
	puzzles := make([]models.Puzzle, len(puzzlePointers))
	for i, p := range puzzlePointers {
		if p != nil {
			puzzles[i] = *p
		}
	}

	// Save puzzles
	return s.store.SavePuzzles(date, puzzles)
}

// TriggerTodayGeneration triggers puzzle generation for today if puzzles don't exist
// Returns true if puzzles were generated, false if they already existed
func (s *Scheduler) TriggerTodayGeneration() (bool, error) {
	today := store.GetTodayDate()

	// Check if puzzles already exist
	if s.store.HasPuzzlesForDate(today) {
		log.Printf("Puzzles already exist for today (%s)", today)
		return false, nil
	}

	// Generate puzzles for today
	s.generatePuzzlesForToday()

	// Verify that puzzles were created
	if !s.store.HasPuzzlesForDate(today) {
		return false, fmt.Errorf("failed to generate puzzles for today")
	}

	return true, nil
}
