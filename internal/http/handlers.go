// Package http provides HTTP handlers for the VESWatch API.
//
// Legal Disclaimer:
// "VESWatch provides reference exchange rates obtained from public sources.
// This information is not official financial advice."
package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/veswatch/api/internal/rates"
)

// RateProvider defines the interface for getting rate data.
type RateProvider interface {
	GetRates() rates.RateData
}

// Handler handles HTTP requests for the API.
type Handler struct {
	rateProvider RateProvider
}

// NewHandler creates a new HTTP handler.
func NewHandler(provider RateProvider) *Handler {
	return &Handler{
		rateProvider: provider,
	}
}

// Routes returns the HTTP router with all endpoints registered.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", h.handleHealth)

	// Main rates endpoint
	mux.HandleFunc("GET /rates", h.handleRates)

	// Root endpoint (redirect to rates)
	mux.HandleFunc("GET /", h.handleRoot)

	// Apply middleware
	return h.withMiddleware(mux)
}

// withMiddleware applies common middleware to all routes.
func (h *Handler) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers for frontend access
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Log request
		log.Printf("HTTP: %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

// handleHealth returns a simple health check response.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleRates returns the current exchange rates.
func (h *Handler) handleRates(w http.ResponseWriter, r *http.Request) {
	rateData := h.rateProvider.GetRates()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(rateData); err != nil {
		log.Printf("HTTP: Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleRoot redirects to the rates endpoint.
func (h *Handler) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"name":       "VESWatch API",
		"version":    "1.0.0",
		"endpoints":  "/rates",
		"disclaimer": "VESWatch provides reference exchange rates obtained from public sources. This information is not official financial advice.",
	})
}
