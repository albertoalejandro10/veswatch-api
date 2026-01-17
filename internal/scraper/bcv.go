// Package scraper provides exchange rate scrapers for BCV and Binance.
//
// Legal Disclaimer:
// "VESWatch provides reference exchange rates obtained from public sources.
// This information is not official financial advice."
package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	bcvURL = "https://www.bcv.org.ve/"
)

// BCVScraper scrapes the official USD rate from BCV website using Colly.
type BCVScraper struct {
	collector *colly.Collector
}

// NewBCVScraper creates a new BCV scraper instance.
func NewBCVScraper() *BCVScraper {
	c := colly.NewCollector(
		colly.AllowedDomains("www.bcv.org.ve", "bcv.org.ve"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	// Set timeouts
	c.SetRequestTimeout(30 * time.Second)

	return &BCVScraper{
		collector: c,
	}
}

// Fetch scrapes the current USD rate from BCV website.
func (s *BCVScraper) Fetch() (float64, error) {
	var rate float64
	var scrapeErr error

	// Clone collector for thread safety
	c := s.collector.Clone()

	// Track if we found the rate
	found := false

	// Primary selector: USD section in the exchange rates area
	// The BCV website shows exchange rates in a specific section
	c.OnHTML("#dolar", func(e *colly.HTMLElement) {
		// Try to find the rate value within the USD section
		rateStr := e.ChildText("strong")
		if rateStr == "" {
			rateStr = e.Text
		}

		parsed, err := parseVESRate(rateStr)
		if err == nil && parsed > 0 {
			rate = parsed
			found = true
			log.Printf("BCV: Found USD rate using #dolar selector: %.4f", rate)
		}
	})

	// Fallback selector: Look for the exchange rate in the recuadrotsmc section
	c.OnHTML(".recuadrotsmc .centmark", func(e *colly.HTMLElement) {
		if found {
			return
		}

		// Check if this is the USD section
		parent := e.DOM.Parent()
		if parent.Find("#dolar").Length() > 0 || strings.Contains(e.Text, "USD") {
			rateStr := e.ChildText("strong")
			if rateStr == "" {
				rateStr = e.Text
			}

			parsed, err := parseVESRate(rateStr)
			if err == nil && parsed > 0 {
				rate = parsed
				found = true
				log.Printf("BCV: Found USD rate using fallback selector: %.4f", rate)
			}
		}
	})

	// Another fallback: Look for any strong element with a rate pattern near USD text
	c.OnHTML("div.col-sm-6.col-xs-6.centmark", func(e *colly.HTMLElement) {
		if found {
			return
		}

		rateStr := e.ChildText("strong")
		parsed, err := parseVESRate(rateStr)
		if err == nil && parsed > 0 {
			rate = parsed
			found = true
			log.Printf("BCV: Found USD rate using col-sm-6 selector: %.4f", rate)
		}
	})

	// Generic fallback: Look for rate patterns in the page
	c.OnHTML("strong", func(e *colly.HTMLElement) {
		if found {
			return
		}

		parsed, err := parseVESRate(e.Text)
		if err == nil && parsed > 20 && parsed < 200 {
			// Reasonable USD/VES rate range check
			rate = parsed
			found = true
			log.Printf("BCV: Found USD rate using strong tag fallback: %.4f", rate)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		scrapeErr = fmt.Errorf("BCV request failed: %w (status: %d)", err, r.StatusCode)
		log.Printf("BCV scrape error: %v", scrapeErr)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("BCV: Scraping %s", r.URL.String())
	})

	// Visit the BCV website
	if err := c.Visit(bcvURL); err != nil {
		return 0, fmt.Errorf("failed to visit BCV: %w", err)
	}

	if scrapeErr != nil {
		return 0, scrapeErr
	}

	if !found || rate == 0 {
		return 0, fmt.Errorf("BCV: USD rate not found on page")
	}

	return rate, nil
}

// parseVESRate extracts a float rate from a Venezuelan formatted string.
// Venezuelan format uses comma as decimal separator (e.g., "45,82")
func parseVESRate(s string) (float64, error) {
	// Clean the string
	s = strings.TrimSpace(s)

	// Remove any currency symbols, spaces, and non-numeric chars except comma and dot
	re := regexp.MustCompile(`[^\d,.]`)
	s = re.ReplaceAllString(s, "")

	if s == "" {
		return 0, fmt.Errorf("empty rate string")
	}

	// Venezuelan format: comma as decimal separator
	// Replace comma with dot for parsing
	s = strings.Replace(s, ",", ".", -1)

	// Handle multiple dots (thousand separators)
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		// Assume last part is decimals, rest are thousand separators
		s = strings.Join(parts[:len(parts)-1], "") + "." + parts[len(parts)-1]
	}

	rate, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse rate '%s': %w", s, err)
	}

	return rate, nil
}
