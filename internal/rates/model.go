// Package rates provides data models and services for exchange rate management.
//
// Legal Disclaimer:
// "VESWatch provides reference exchange rates obtained from public sources.
// This information is not official financial advice."
package rates

import (
	"sync"
	"time"
)

// RateData represents the current exchange rate information.
type RateData struct {
	BCV       float64   `json:"bcv"`
	Binance   float64   `json:"binance"`
	Breach    float64   `json:"breach"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// RateStore provides thread-safe storage for rate data.
type RateStore struct {
	mu      sync.RWMutex
	bcv     float64
	binance float64
	bcvTime time.Time
	binTime time.Time
}

// NewRateStore creates a new RateStore instance.
func NewRateStore() *RateStore {
	return &RateStore{}
}

// SetBCV updates the BCV rate value.
func (s *RateStore) SetBCV(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bcv = rate
	s.bcvTime = time.Now()
}

// SetBinance updates the Binance rate value.
func (s *RateStore) SetBinance(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.binance = rate
	s.binTime = time.Now()
}

// GetBCV returns the current BCV rate.
func (s *RateStore) GetBCV() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bcv
}

// GetBinance returns the current Binance rate.
func (s *RateStore) GetBinance() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.binance
}

// GetRateData returns the complete rate data with breach calculation.
func (s *RateStore) GetRateData() RateData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var breach float64
	if s.bcv > 0 {
		breach = ((s.binance - s.bcv) / s.bcv) * 100
	}

	// Use the most recent update time
	updatedAt := s.bcvTime
	if s.binTime.After(s.bcvTime) {
		updatedAt = s.binTime
	}

	// Round breach to 2 decimal places
	breach = float64(int(breach*100)) / 100

	return RateData{
		BCV:       s.bcv,
		Binance:   s.binance,
		Breach:    breach,
		UpdatedAt: updatedAt,
	}
}
