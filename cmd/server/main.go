// VESWatch API Server
//
// Legal Disclaimer:
// "VESWatch provides reference exchange rates obtained from public sources.
// This information is not official financial advice."

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httphandlers "github.com/veswatch/api/internal/http"
	"github.com/veswatch/api/internal/rates"
	"github.com/veswatch/api/internal/scheduler"
	"github.com/veswatch/api/internal/scraper"
)

func main() {
	log.Println("Starting VESWatch API Server...")

	// Initialize scrapers
	bcvScraper := scraper.NewBCVScraper()
	binanceFetcher := scraper.NewBinanceFetcher()

	// Initialize rates service
	ratesService := rates.NewService(bcvScraper, binanceFetcher)

	// Initialize scheduler
	sched := scheduler.New(ratesService)
	sched.Start()

	// Initialize HTTP handlers
	handler := httphandlers.NewHandler(ratesService)

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configure HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop scheduler
	sched.Stop()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
