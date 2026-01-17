package rates

import (
	"log"
)

// Scraper defines the interface for exchange rate scrapers.
type Scraper interface {
	Fetch() (float64, error)
}

// Service manages exchange rate fetching and storage.
type Service struct {
	store          *RateStore
	bcvScraper     Scraper
	binanceFetcher Scraper
}

// NewService creates a new rate service.
func NewService(bcvScraper, binanceFetcher Scraper) *Service {
	return &Service{
		store:          NewRateStore(),
		bcvScraper:     bcvScraper,
		binanceFetcher: binanceFetcher,
	}
}

// FetchBCV scrapes the BCV rate and updates the store.
// If scraping fails, the previous value is retained.
func (s *Service) FetchBCV() error {
	rate, err := s.bcvScraper.Fetch()
	if err != nil {
		log.Printf("BCV fetch error (keeping previous value): %v", err)
		return err
	}

	s.store.SetBCV(rate)
	log.Printf("BCV rate updated: %.2f", rate)
	return nil
}

// FetchBinance fetches the Binance P2P rate and updates the store.
// If fetching fails, the previous value is retained.
func (s *Service) FetchBinance() error {
	rate, err := s.binanceFetcher.Fetch()
	if err != nil {
		log.Printf("Binance fetch error (keeping previous value): %v", err)
		return err
	}

	s.store.SetBinance(rate)
	log.Printf("Binance rate updated: %.2f", rate)
	return nil
}

// GetRates returns the current rate data.
func (s *Service) GetRates() RateData {
	return s.store.GetRateData()
}

// Initialize performs the initial data fetch on startup.
func (s *Service) Initialize() {
	log.Println("Initializing rate data...")

	// Fetch Binance first (more reliable)
	if err := s.FetchBinance(); err != nil {
		log.Printf("Initial Binance fetch failed: %v", err)
	}

	// Attempt BCV fetch
	if err := s.FetchBCV(); err != nil {
		log.Printf("Initial BCV fetch failed: %v", err)
	}

	log.Println("Rate data initialization complete")
}
