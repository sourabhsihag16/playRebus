package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"backend/internal/models"
	"backend/internal/scheduler"
	"backend/internal/store"
)

// PuzzleHandler handles puzzle-related HTTP requests
type PuzzleHandler struct {
	store     *store.Store
	scheduler *scheduler.Scheduler
}

// NewPuzzleHandler creates a new puzzle handler
func NewPuzzleHandler(store *store.Store, sched *scheduler.Scheduler) *PuzzleHandler {
	return &PuzzleHandler{
		store:     store,
		scheduler: sched,
	}
}

// GetPuzzlesHandler handles GET /api/puzzles/{date}
func (h *PuzzleHandler) GetPuzzlesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

	// Validate date format
	if err := store.ValidateDate(date); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get puzzles from store
	puzzles, err := h.store.GetPuzzlesForDate(date)
	if err != nil {
		http.Error(w, fmt.Sprintf("No puzzles found for date: %s. They may not have been generated yet.", date), http.StatusNotFound)
		return
	}

	response := models.PuzzlesResponse{
		Date:    date,
		Puzzles: puzzles,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// VerifyAnswerHandler handles POST /api/puzzles/verify
func (h *PuzzleHandler) VerifyAnswerHandler(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse puzzle ID to get date and index
	date, _, err := store.ParsePuzzleID(req.PuzzleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get puzzles for the date
	puzzles, err := h.store.GetPuzzlesForDate(date)
	if err != nil {
		http.Error(w, "Puzzle not found", http.StatusNotFound)
		return
	}

	// Find the puzzle
	var puzzle *models.Puzzle
	for i := range puzzles {
		if puzzles[i].ID == req.PuzzleID {
			puzzle = &puzzles[i]
			break
		}
	}

	if puzzle == nil {
		http.Error(w, "Puzzle not found", http.StatusNotFound)
		return
	}

	// Compare answers (case-insensitive, trimmed)
	userAnswer := strings.ToLower(strings.TrimSpace(req.Answer))
	correctAnswer := strings.ToLower(strings.TrimSpace(puzzle.Answer))
	correct := userAnswer == correctAnswer

	response := models.VerifyResponse{
		Correct: correct,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// TriggerJobHandler handles POST /api/puzzles/trigger
// Triggers puzzle generation for today if puzzles don't exist
func (h *PuzzleHandler) TriggerJobHandler(w http.ResponseWriter, r *http.Request) {
	generated, err := h.scheduler.TriggerTodayGeneration()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to trigger job: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"generated": generated,
		"message":   "",
	}

	if generated {
		response["message"] = "Puzzles generated successfully for today"
	} else {
		response["message"] = "Puzzles already exist for today"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
