package handlers

import (
	"net/http"
	"path/filepath"
	"strings"
)

// ImageHandler handles image serving
type ImageHandler struct {
	imagesPath      string
	useSupabase     bool
	supabaseBaseURL string
}

// NewImageHandler creates a new image handler for file system storage
func NewImageHandler(imagesPath string) *ImageHandler {
	return &ImageHandler{
		imagesPath:  imagesPath,
		useSupabase: false,
	}
}

// NewImageHandlerWithSupabase creates a new image handler that redirects to Supabase URLs
func NewImageHandlerWithSupabase(supabaseBaseURL string) *ImageHandler {
	return &ImageHandler{
		useSupabase:     true,
		supabaseBaseURL: supabaseBaseURL,
	}
}

// ServeImage handles GET /api/images/{filename}
func (h *ImageHandler) ServeImage(w http.ResponseWriter, r *http.Request) {
	// Get filename from URL path
	filename := strings.TrimPrefix(r.URL.Path, "/api/images/")

	// Security: prevent directory traversal
	if strings.Contains(filename, "..") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// If using Supabase, redirect to the Supabase public URL
	if h.useSupabase {
		// Extract date and index from filename (format: YYYY-MM-DD-index.png)
		// Parse to get date folder structure
		parts := strings.Split(filename, "-")
		if len(parts) >= 4 {
			// Format: YYYY-MM-DD-index.png
			date := strings.Join(parts[:3], "-")
			indexWithExt := parts[3]
			index := strings.TrimSuffix(indexWithExt, filepath.Ext(indexWithExt))
			
			// Redirect to Supabase URL: {baseURL}/{date}/{index}.png
			supabaseURL := h.supabaseBaseURL + "/" + date + "/" + index + ".png"
			http.Redirect(w, r, supabaseURL, http.StatusMovedPermanently)
			return
		}
		// Fallback: try to construct URL directly
		supabaseURL := h.supabaseBaseURL + "/" + filename
		http.Redirect(w, r, supabaseURL, http.StatusMovedPermanently)
		return
	}

	// File system serving (legacy)
	// Security: prevent directory traversal
	if strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Construct full path
	imagePath := filepath.Join(h.imagesPath, filename)

	// Set CORS headers to allow frontend to load images
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle OPTIONS preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set content type
	ext := filepath.Ext(filename)
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Serve file
	http.ServeFile(w, r, imagePath)
}
