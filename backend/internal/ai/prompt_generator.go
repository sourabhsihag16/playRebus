package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// PromptGenerator handles generation of rebus puzzle prompts using Claude API
type PromptGenerator struct {
	claudeAPIKey string
	client       *http.Client
	promptCache  map[string][]RebusPrompt
	cacheMutex   sync.Mutex
}

// NewPromptGenerator creates a new prompt generator
func NewPromptGenerator(claudeAPIKey string) *PromptGenerator {
	return &PromptGenerator{
		claudeAPIKey: claudeAPIKey,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		promptCache: make(map[string][]RebusPrompt),
	}
}

// GetPromptsFromClaude fetches 5 rebus puzzle prompts from Claude API
func (pg *PromptGenerator) GetPromptsFromClaude(date string) ([]RebusPrompt, error) {
	// Check cache first
	pg.cacheMutex.Lock()
	if prompts, exists := pg.promptCache[date]; exists {
		pg.cacheMutex.Unlock()
		fmt.Printf("Using cached prompts for date: %s\n", date)
		return prompts, nil
	}
	pg.cacheMutex.Unlock()

	fmt.Printf("Calling Claude API to generate 5 rebus puzzle prompts for date: %s\n", date)

	// Claude API endpoint
	claudeURL := "https://api.anthropic.com/v1/messages"

	// Enhanced system prompt for better, more descriptive prompts with common phrases
	systemPrompt := `You are an expert at creating rebus puzzles. A rebus puzzle uses pictures, words, or symbols arranged to represent a word or phrase.

IMPORTANT GUIDELINES:
- Use VERY COMMON and FAMILIAR phrases that most people know (e.g., "break the ice", "piece of cake", "once upon a time", "home sweet home", "time flies", "raining cats and dogs")
- Avoid obscure or uncommon phrases
- Make the puzzles fun and engaging for all ages
- The visual description should be clear and detailed enough for image generation

Return the response as a JSON array with this exact format:
[
  {
    "prompt": "Detailed description of what the rebus puzzle image should show (describe visual elements clearly)",
    "answer": "the correct answer (common phrase or word)",
    "hint": "a helpful hint that guides without giving away the answer"
  },
  ... (4 more puzzles)
]`

	// Enhanced user prompt with more specific instructions
	userPrompt := fmt.Sprintf(`Generate exactly 5 different rebus puzzle prompts for date %s. 

For each puzzle, you must provide:

1. PROMPT: A very detailed and descriptive prompt for generating the rebus puzzle image. This should clearly describe:
   - What visual elements should appear in the image
   - How words, pictures, or symbols should be arranged
   - The layout and composition of the rebus puzzle
   - Be specific about colors, positions, and relationships between elements
   - Remember: The image will have a BLACK background with WHITE elements

2. ANSWER: The correct answer must be a VERY COMMON phrase or word that most people would recognize. Examples:
   - Common idioms: "break the ice", "piece of cake", "once upon a time"
   - Common phrases: "home sweet home", "time flies", "raining cats and dogs"
   - Common words: "butterfly", "sunshine", "rainbow"
   - Avoid obscure references, technical terms, or niche knowledge

3. HINT: A helpful hint that guides the solver without revealing the answer directly. Make it encouraging and fun.

Requirements:
- All 5 puzzles should be creative and varied
- Use different types of rebus puzzles (word combinations, picture-word mixes, symbol arrangements)
- Ensure answers are appropriate for all ages
- Make sure the phrases are VERY COMMON and easily recognizable
- The visual descriptions should be rich and detailed for better image generation`, date)

	requestPayload := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 3000, // Increased for more descriptive prompts
		"system":     systemPrompt,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Claude request: %w", err)
	}

	// Make request to Claude API
	req, err := http.NewRequest("POST", claudeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Claude request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", pg.claudeAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Log request details (without exposing full API key)
	if len(pg.claudeAPIKey) > 0 {
		previewLen := 8
		if len(pg.claudeAPIKey) < previewLen {
			previewLen = len(pg.claudeAPIKey)
		}
		apiKeyPreview := pg.claudeAPIKey[:previewLen] + "..."
		fmt.Println("Making Claude API request with API key:", apiKeyPreview)
	}

	resp, err := pg.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		// Provide helpful error message for 401
		if resp.StatusCode == http.StatusUnauthorized {
			if pg.claudeAPIKey == "" {
				return nil, fmt.Errorf("authentication failed: CLAUDE_API_KEY environment variable is not set or is empty")
			}
			return nil, fmt.Errorf("authentication failed: invalid or expired API key. Status %d: %s", resp.StatusCode, string(body))
		}

		return nil, fmt.Errorf("claude API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Claude response
	var claudeResponse struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&claudeResponse); err != nil {
		return nil, fmt.Errorf("failed to decode Claude response: %w", err)
	}

	if len(claudeResponse.Content) == 0 {
		return nil, fmt.Errorf("no content in Claude response")
	}

	// Extract JSON from Claude's text response
	responseText := claudeResponse.Content[0].Text
	// Try to extract JSON array from the response (it might be wrapped in markdown code blocks)
	jsonStart := strings.Index(responseText, "[")
	jsonEnd := strings.LastIndex(responseText, "]")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("could not find JSON array in Claude response")
	}
	jsonText := responseText[jsonStart : jsonEnd+1]
	fmt.Println("Extracted JSON from Claude response")

	// Parse the prompts
	var prompts []RebusPrompt
	if err := json.Unmarshal([]byte(jsonText), &prompts); err != nil {
		return nil, fmt.Errorf("failed to parse prompts from Claude response: %w", err)
	}

	if len(prompts) != 5 {
		return nil, fmt.Errorf("expected 5 prompts, got %d", len(prompts))
	}

	// Cache the prompts
	pg.cacheMutex.Lock()
	pg.promptCache[date] = prompts
	pg.cacheMutex.Unlock()
	fmt.Printf("Prompts cached for date: %s\n", date)

	return prompts, nil
}

