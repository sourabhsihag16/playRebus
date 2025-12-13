package ai

import (
	"fmt"
	"strings"

	"backend/internal/models"
	"backend/internal/store"
)

// RebusPrompt represents a prompt for generating a rebus puzzle
type RebusPrompt struct {
	Prompt string `json:"prompt"` // Prompt for image generation
	Answer string `json:"answer"` // Correct answer
	Hint   string `json:"hint"`   // Hint for the puzzle
}

// AIGenerator interface for generating rebus puzzles
type AIGenerator interface {
	GenerateRebusPuzzle(date string, index int, imageStore *store.Store) (*models.Puzzle, error)
	GenerateRebusPuzzles(date string, imageStore *store.Store) ([]*models.Puzzle, error)
}

// RealAIGenerator implements AIGenerator using Claude API and image generation service
type RealAIGenerator struct {
	promptGenerator *PromptGenerator
	imageGenerator  *ImageGenerator
	environment     string
}

// NewRealAIGenerator creates a new AI generator
func NewRealAIGenerator(claudeAPIKey, replicateAPIKey, environment string) *RealAIGenerator {
	return &RealAIGenerator{
		promptGenerator: NewPromptGenerator(claudeAPIKey),
		imageGenerator:  NewImageGenerator(replicateAPIKey),
		environment:     environment,
	}
}

// GenerateRebusPuzzle generates a single rebus puzzle using AI service
func (g *RealAIGenerator) GenerateRebusPuzzle(date string, index int, imageStore *store.Store) (*models.Puzzle, error) {
	// Get prompts from Claude (will use cache if already fetched)
	prompts, err := g.promptGenerator.GetPromptsFromClaude(date)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompts: %w", err)
	}

	if index >= len(prompts) {
		return nil, fmt.Errorf("index %d out of range (max %d)", index, len(prompts)-1)
	}

	prompt := prompts[index]

	// Generate image from prompt (with black background, white elements, no hints/answers)
	imageData, err := g.imageGenerator.GenerateImageFromPrompt(prompt.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}

	// Save image locally
	if err := imageStore.SaveImage(date, index, imageData); err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	// Create puzzle
	puzzleID := fmt.Sprintf("%s-%d", date, index)
	puzzle := &models.Puzzle{
		ID:        puzzleID,
		ImageURL:  imageStore.GetImageURL(date, index),
		ImagePath: imageStore.GetImagePath(date, index),
		Answer:    strings.ToLower(strings.TrimSpace(prompt.Answer)),
		Hint:      prompt.Hint,
		Date:      date,
		Index:     index,
	}

	return puzzle, nil
}

// GenerateRebusPuzzles generates all 5 rebus puzzles for a date
func (g *RealAIGenerator) GenerateRebusPuzzles(date string, imageStore *store.Store) ([]*models.Puzzle, error) {
	fmt.Printf("Starting to generate 5 rebus puzzles for date: %s\n", date)

	// Step 1: Get all prompts from Claude API
	prompts, err := g.promptGenerator.GetPromptsFromClaude(date)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompts from Claude: %w", err)
	}

	fmt.Printf("Successfully received %d prompts from Claude API\n", len(prompts))

	puzzles := make([]*models.Puzzle, len(prompts))

	// Step 2: Generate images for each prompt one by one
	for i, prompt := range prompts {
		fmt.Printf("Generating image %d/5 for puzzle with answer: %s\n", i+1, prompt.Answer)
		// Generate image from prompt (with black background, white elements, no hints/answers)
		imageData, err := g.imageGenerator.GenerateImageFromPrompt(prompt.Prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to generate image for puzzle %d: %w", i, err)
		}
		fmt.Printf("Successfully generated image %d/5\n", i+1)

		// Save image locally
		if err := imageStore.SaveImage(date, i, imageData); err != nil {
			return nil, fmt.Errorf("failed to save image for puzzle %d: %w", i, err)
		}

		// Create puzzle
		puzzleID := fmt.Sprintf("%s-%d", date, i)
		puzzles[i] = &models.Puzzle{
			ID:        puzzleID,
			ImageURL:  imageStore.GetImageURL(date, i),
			ImagePath: imageStore.GetImagePath(date, i),
			Answer:    strings.ToLower(strings.TrimSpace(prompt.Answer)),
			Hint:      prompt.Hint,
			Date:      date,
			Index:     i,
		}
	}

	fmt.Printf("Successfully generated all 5 rebus puzzles for date: %s\n", date)
	return puzzles, nil
}
