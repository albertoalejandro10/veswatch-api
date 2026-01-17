package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	binanceP2PURL = "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search"
)

// BinanceFetcher fetches USDT/VES rates from Binance P2P.
type BinanceFetcher struct {
	client *http.Client
}

// NewBinanceFetcher creates a new Binance P2P fetcher.
func NewBinanceFetcher() *BinanceFetcher {
	return &BinanceFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// binanceRequest represents the P2P search request payload.
type binanceRequest struct {
	Fiat              string   `json:"fiat"`
	Page              int      `json:"page"`
	Rows              int      `json:"rows"`
	TradeType         string   `json:"tradeType"`
	Asset             string   `json:"asset"`
	Countries         []string `json:"countries,omitempty"`
	ProMerchantAds    bool     `json:"proMerchantAds"`
	ShieldMerchantAds bool     `json:"shieldMerchantAds"`
	PublisherType     *string  `json:"publisherType,omitempty"`
	PayTypes          []string `json:"payTypes,omitempty"`
}

// binanceResponse represents the P2P search response.
type binanceResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Adv struct {
			Price         string `json:"price"`
			Asset         string `json:"asset"`
			FiatUnit      string `json:"fiatUnit"`
			TradeType     string `json:"tradeType"`
			SurplusAmount string `json:"surplusAmount"`
		} `json:"adv"`
		Advertiser struct {
			NickName        string  `json:"nickName"`
			MonthFinishRate float64 `json:"monthFinishRate"`
			PositiveRate    float64 `json:"positiveRate"`
		} `json:"advertiser"`
	} `json:"data"`
	Total int `json:"total"`
}

// Fetch retrieves the current USDT/VES rate from Binance P2P.
func (f *BinanceFetcher) Fetch() (float64, error) {
	// Build request payload
	reqBody := binanceRequest{
		Fiat:              "VES",
		Page:              1,
		Rows:              10,
		TradeType:         "BUY",
		Asset:             "USDT",
		ProMerchantAds:    false,
		ShieldMerchantAds: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", binanceP2PURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to mimic browser request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	log.Printf("Binance: Fetching P2P USDT/VES rates")

	resp, err := f.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("binance request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("binance returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var result binanceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Data) == 0 {
		return 0, fmt.Errorf("no P2P ads found for USDT/VES")
	}

	// Calculate median price from first few results for a representative rate
	var prices []float64
	for _, ad := range result.Data {
		price, err := strconv.ParseFloat(ad.Adv.Price, 64)
		if err != nil {
			log.Printf("Binance: Failed to parse price '%s': %v", ad.Adv.Price, err)
			continue
		}
		prices = append(prices, price)
	}

	if len(prices) == 0 {
		return 0, fmt.Errorf("no valid prices found")
	}

	// Use the median price for a more stable rate
	rate := median(prices)
	log.Printf("Binance: Found %d prices, median: %.2f", len(prices), rate)

	return rate, nil
}

// median calculates the median of a slice of float64.
func median(prices []float64) float64 {
	n := len(prices)
	if n == 0 {
		return 0
	}

	// Sort prices
	sorted := make([]float64, n)
	copy(sorted, prices)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}
