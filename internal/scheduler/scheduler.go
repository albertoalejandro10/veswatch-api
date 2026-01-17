// Package scheduler provides scheduling functionality for rate fetching.
package scheduler

import (
	"log"
	"sync"
	"time"
)

// RateService defines the interface for rate fetching operations.
type RateService interface {
	Initialize()
	FetchBCV() error
	FetchBinance() error
}

// Scheduler manages timed jobs for fetching exchange rates.
type Scheduler struct {
	service RateService
	stop    chan struct{}
	wg      sync.WaitGroup
}

// New creates a new scheduler instance.
func New(service RateService) *Scheduler {
	return &Scheduler{
		service: service,
		stop:    make(chan struct{}),
	}
}

// Start begins the scheduler jobs.
func (s *Scheduler) Start() {
	log.Println("Scheduler: Starting...")

	// Initialize data on startup
	s.service.Initialize()

	// Start Binance refresh job (every 5 minutes)
	s.wg.Add(1)
	go s.binanceJob()

	// Start BCV daily job
	s.wg.Add(1)
	go s.bcvDailyJob()

	log.Println("Scheduler: All jobs started")
}

// Stop gracefully stops all scheduler jobs.
func (s *Scheduler) Stop() {
	log.Println("Scheduler: Stopping...")
	close(s.stop)
	s.wg.Wait()
	log.Println("Scheduler: Stopped")
}

// binanceJob refreshes Binance rates every 5 minutes.
func (s *Scheduler) binanceJob() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("Scheduler: Binance refresh job started (every 5 minutes)")

	for {
		select {
		case <-s.stop:
			log.Println("Scheduler: Binance job stopped")
			return
		case <-ticker.C:
			log.Println("Scheduler: Refreshing Binance rate")
			if err := s.service.FetchBinance(); err != nil {
				log.Printf("Scheduler: Binance refresh failed: %v", err)
			}
		}
	}
}

// bcvDailyJob scrapes BCV once per day on weekdays.
func (s *Scheduler) bcvDailyJob() {
	defer s.wg.Done()

	log.Println("Scheduler: BCV daily job started")

	for {
		// Calculate time until next BCV update (11:00 AM Venezuela time)
		nextRun := s.nextBCVRunTime()
		waitDuration := time.Until(nextRun)

		log.Printf("Scheduler: Next BCV scrape scheduled for %s (in %s)",
			nextRun.Format(time.RFC3339), waitDuration.Round(time.Minute))

		select {
		case <-s.stop:
			log.Println("Scheduler: BCV job stopped")
			return
		case <-time.After(waitDuration):
			if s.isWeekday() {
				log.Println("Scheduler: Running BCV daily scrape")
				if err := s.service.FetchBCV(); err != nil {
					log.Printf("Scheduler: BCV daily scrape failed: %v", err)
				}
			} else {
				log.Println("Scheduler: Skipping BCV scrape (weekend)")
			}
		}
	}
}

// nextBCVRunTime calculates the next time to run the BCV scraper.
// BCV typically updates around 11:00 AM Venezuela time (UTC-4).
func (s *Scheduler) nextBCVRunTime() time.Time {
	// Venezuela timezone (UTC-4)
	loc := time.FixedZone("VET", -4*60*60)
	now := time.Now().In(loc)

	// Target time: 11:30 AM (giving BCV time to update)
	targetHour := 11
	targetMinute := 30

	next := time.Date(now.Year(), now.Month(), now.Day(),
		targetHour, targetMinute, 0, 0, loc)

	// If we've passed today's target time, schedule for tomorrow
	if now.After(next) {
		next = next.Add(24 * time.Hour)
	}

	// Skip to Monday if next run falls on weekend
	for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
		next = next.Add(24 * time.Hour)
	}

	return next
}

// isWeekday returns true if today is a weekday.
func (s *Scheduler) isWeekday() bool {
	weekday := time.Now().Weekday()
	return weekday != time.Saturday && weekday != time.Sunday
}
