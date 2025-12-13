package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ImageGenerator handles generation of rebus puzzle images using Replicate
type ImageGenerator struct {
	replicateAPIKey string
	client          *http.Client
	// Model to use for image generation (default: flux-schnell for speed)
	model string
	// Cached model version to avoid fetching it every time
	modelVersion string
}

// NewImageGenerator creates a new image generator using Replicate
func NewImageGenerator(replicateAPIKey string) *ImageGenerator {
	return &ImageGenerator{
		replicateAPIKey: replicateAPIKey,
		// Using flux-schnell for fast image generation, you can change this to other models
		// Popular options: "black-forest-labs/flux-schnell", "stability-ai/sdxl", "stability-ai/stable-diffusion"
		model: "black-forest-labs/flux-1.1-pro",
		client: &http.Client{
			Timeout: 120 * time.Second, // Increased timeout for image generation
		},
	}
}

// GenerateImageFromPrompt generates an image from a prompt using Replicate API
// The image will have a black background with white elements, showing only the puzzle question (no hints or answers)
func (ig *ImageGenerator) GenerateImageFromPrompt(prompt string) ([]byte, error) {
	// Enhance the prompt to specify black background, white elements, and no hints/answers
	enhancedPrompt := fmt.Sprintf(`Create a rebus puzzle image with the following specifications:
- Background: Pure black (#000000)
- All visual elements: White (#FFFFFF) or light colors that contrast well with black
- Content: Show ONLY the rebus puzzle question/visual elements
- DO NOT include any hints, answers, or text explanations in the image
- The image should be clean and clear, focusing solely on the puzzle elements

Rebus puzzle description: %s

Style: Modern, clean, minimalist rebus puzzle design with black background and white/light colored elements.`, prompt)

	// Step 1: Create a prediction
	prediction, err := ig.createPrediction(enhancedPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to create prediction: %w", err)
	}

	// Step 2: Poll for completion
	imageURL, err := ig.pollPrediction(prediction.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prediction result: %w", err)
	}

	// Step 3: Download the image
	imageData, err := ig.downloadImage(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	return imageData, nil
}

// createPrediction creates a new prediction on Replicate
func (ig *ImageGenerator) createPrediction(prompt string) (*ReplicatePrediction, error) {
	// Get the model version (will fetch if not cached)
	modelVersion, err := ig.getModelVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get model version: %w", err)
	}

	requestPayload := map[string]interface{}{
		"version": modelVersion,
		"input": map[string]interface{}{
			"prompt": prompt,
			"width":  800,
			"height": 600,
		},
	}

	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prediction request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.replicate.com/v1/predictions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", ig.replicateAPIKey))

	// Print curl command for debugging
	fmt.Printf("CURL Request for creating prediction:\n")
	fmt.Printf("curl -X POST https://api.replicate.com/v1/predictions \\\n")
	fmt.Printf("  -H \"Content-Type: application/json\" \\\n")
	fmt.Printf("  -H \"Authorization: Token %s\" \\\n", ig.replicateAPIKey)
	fmt.Printf("  -d '%s'\n\n", string(jsonData))

	resp, err := ig.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Replicate API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("replicate API returned status %d: %s", resp.StatusCode, string(body))
	}

	var prediction ReplicatePrediction
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return nil, fmt.Errorf("failed to decode prediction response: %w", err)
	}

	return &prediction, nil
}

// pollPrediction polls the prediction until it's completed
func (ig *ImageGenerator) pollPrediction(predictionID string) (string, error) {
	pollURL := fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", predictionID)
	maxAttempts := 60 // Maximum 5 minutes (60 * 5 seconds)
	attempt := 0

	for attempt < maxAttempts {
		req, err := http.NewRequest("GET", pollURL, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create poll request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Token %s", ig.replicateAPIKey))

		// Print curl command for polling (only on first attempt)
		if attempt == 0 {
			fmt.Printf("CURL Request for polling prediction:\n")
			fmt.Printf("curl -X GET %s \\\n", pollURL)
			fmt.Printf("  -H \"Authorization: Token %s\"\n\n", ig.replicateAPIKey)
		}

		resp, err := ig.client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to poll prediction: %w", err)
		}

		// Read the response body first to handle it properly
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read poll response: %w", err)
		}

		// Parse response with flexible output handling
		var prediction ReplicatePrediction
		if err := json.Unmarshal(bodyBytes, &prediction); err != nil {
			return "", fmt.Errorf("failed to decode poll response: %w, body: %s", err, string(bodyBytes))
		}

		switch prediction.Status {
		case "succeeded":
			// Extract image URL from output - handle both string and array formats
			imageURL, err := ig.extractImageURL(prediction.OutputRaw)
			if err != nil {
				return "", fmt.Errorf("prediction succeeded but failed to extract image URL: %w", err)
			}
			return imageURL, nil
		case "failed":
			errorMsg := "unknown error"
			if prediction.Error != nil {
				errorMsg = *prediction.Error
			}
			return "", fmt.Errorf("prediction failed: %s", errorMsg)
		case "canceled":
			return "", fmt.Errorf("prediction was canceled")
		case "starting", "processing":
			// Still processing, wait and retry
			time.Sleep(5 * time.Second)
			attempt++
			continue
		default:
			return "", fmt.Errorf("unknown prediction status: %s", prediction.Status)
		}
	}

	return "", fmt.Errorf("prediction timed out after %d attempts", maxAttempts)
}

// downloadImage downloads an image from a URL
func (ig *ImageGenerator) downloadImage(imageURL string) ([]byte, error) {
	fmt.Printf("CURL Request for downloading image:\n")
	fmt.Printf("curl -X GET \"%s\" -o image.png\n\n", imageURL)
	fmt.Println("Downloading image from URL:", imageURL)

	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return imageData, nil
}

// getModelVersion returns the model version ID for the current model
// It fetches the latest version from Replicate API if not cached
func (ig *ImageGenerator) getModelVersion() (string, error) {
	// If we already have a cached version, use it
	if ig.modelVersion != "" {
		return ig.modelVersion, nil
	}

	// Fetch the latest version for the model
	version, err := ig.fetchLatestModelVersion()
	if err != nil {
		return "", fmt.Errorf("failed to fetch model version: %w", err)
	}

	// Cache it for future use
	ig.modelVersion = version
	return version, nil
}

// fetchLatestModelVersion fetches the latest version ID for the configured model
func (ig *ImageGenerator) fetchLatestModelVersion() (string, error) {
	modelURL := fmt.Sprintf("https://api.replicate.com/v1/models/%s", ig.model)

	req, err := http.NewRequest("GET", modelURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", ig.replicateAPIKey))

	resp, err := ig.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch model info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch model info: status %d: %s", resp.StatusCode, string(body))
	}

	var modelInfo struct {
		LatestVersion struct {
			ID string `json:"id"`
		} `json:"latest_version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelInfo); err != nil {
		return "", fmt.Errorf("failed to decode model info: %w", err)
	}

	if modelInfo.LatestVersion.ID == "" {
		return "", fmt.Errorf("no latest version found for model %s", ig.model)
	}

	return modelInfo.LatestVersion.ID, nil
}

// ReplicatePrediction represents a Replicate API prediction response
type ReplicatePrediction struct {
	ID        string          `json:"id"`
	Status    string          `json:"status"`
	OutputRaw json.RawMessage `json:"output"` // Use RawMessage to handle different types
	Error     *string         `json:"error"`
	Created   string          `json:"created_at"`
	URLs      struct {
		Get    string `json:"get"`
		Cancel string `json:"cancel"`
	} `json:"urls"`
}

// extractImageURL extracts the image URL from the output field
// Output can be: a string (single image), an array of strings (multiple images), or null
func (ig *ImageGenerator) extractImageURL(outputRaw json.RawMessage) (string, error) {
	if len(outputRaw) == 0 {
		return "", fmt.Errorf("output is empty")
	}

	// Try to unmarshal as a string first (single image)
	var urlString string
	if err := json.Unmarshal(outputRaw, &urlString); err == nil {
		if urlString != "" {
			return urlString, nil
		}
	}

	// Try to unmarshal as an array of strings (multiple images)
	var urlArray []string
	if err := json.Unmarshal(outputRaw, &urlArray); err == nil {
		if len(urlArray) > 0 && urlArray[0] != "" {
			return urlArray[0], nil
		}
	}

	// Try as array of interface{} (in case of mixed types)
	var urlInterfaceArray []interface{}
	if err := json.Unmarshal(outputRaw, &urlInterfaceArray); err == nil {
		if len(urlInterfaceArray) > 0 {
			if url, ok := urlInterfaceArray[0].(string); ok && url != "" {
				return url, nil
			}
		}
	}

	return "", fmt.Errorf("could not extract image URL from output: %s", string(outputRaw))
}
