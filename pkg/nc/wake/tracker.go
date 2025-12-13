package wake

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

// Tracker implements the tracker.Tracker interface for Wake County, NC ABC
type Tracker struct {
	config       tracker.Config
	productNames map[string]string // map of product code to name
	client       *http.Client
}

// New creates a new Wake County ABC tracker
func New() (*Tracker, error) {
	// Product names to search for (subset of VA ABC products)
	// Wake County uses product names, not codes
	productNames := map[string]string{
		"blantons":          "Blanton's",
		"buffalo-trace":     "Buffalo Trace",
		"pappy-23":          "Pappy Van Winkle 23",
		"pappy-20":          "Pappy Van Winkle 20",
		"pappy-15":          "Pappy Van Winkle 15",
		"weller-12":         "Weller 12",
		"weller-antique":    "Weller Antique",
		"weller-full-proof": "Weller Full Proof",
		"weller-cypb":       "Weller C.Y.P.B",
		"eagle-rare":        "Eagle Rare",
		"eh-taylor":         "E.H. Taylor",
		"stagg-jr":          "Stagg Jr",
		"george-stagg":      "George T. Stagg",
		"elmer-lee":         "Elmer T. Lee",
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Tracker{
		config:       tracker.DefaultConfig(),
		productNames: productNames,
		client:       client,
	}, nil
}

// Name returns the tracker name
func (t *Tracker) Name() string {
	return "NC Wake County ABC"
}

// ProductCodes returns the list of product codes
func (t *Tracker) ProductCodes() []string {
	codes := make([]string, 0, len(t.productNames))
	for code := range t.productNames {
		codes = append(codes, code)
	}
	return codes
}

// StoreCount returns the number of stores (15 Wake County ABC stores)
func (t *Tracker) StoreCount() int {
	return 15 // Wake County has 15 ABC store locations
}

// Track queries Wake County inventory and returns items
func (t *Tracker) Track() ([]tracker.InventoryItem, error) {
	var allItems []tracker.InventoryItem

	for code, productName := range t.productNames {
		fmt.Fprintf(log.Writer(), "  Searching for: %s\n", productName)

		items, err := t.searchProduct(code, productName)
		if err != nil {
			fmt.Fprintf(log.Writer(), "  ERROR searching %s: %v\n", productName, err)
			continue
		}

		allItems = append(allItems, items...)

		// Rate limiting: 500ms delay between searches
		time.Sleep(500 * time.Millisecond)
	}

	return allItems, nil
}

// searchProduct searches for a specific product and parses results
func (t *Tracker) searchProduct(productCode, productName string) ([]tracker.InventoryItem, error) {
	// POST form data
	formData := url.Values{}
	formData.Set("productSearch", productName)

	req, err := http.NewRequest("POST", "https://wakeabc.com/search-results", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to mimic browser
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://wakeabc.com/search-our-inventory/")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search product: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read and parse HTML
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return t.parseSearchResults(productCode, string(body))
}

// parseSearchResults extracts inventory items from HTML
func (t *Tracker) parseSearchResults(productCode, html string) ([]tracker.InventoryItem, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var items []tracker.InventoryItem
	now := time.Now()

	// Find all product divs
	doc.Find("div.wake-product").Each(func(i int, s *goquery.Selection) {
		// Extract product name
		productName := strings.TrimSpace(s.Find("h4").Text())

		// Extract PLU (Wake County's product code)
		pluText := s.Find("small").Text()
		plu := extractPLU(pluText)

		// Check if out of stock
		outOfStock := s.Find("p.out-of-stock").Length() > 0
		if outOfStock {
			// Skip out of stock items
			return
		}

		// Extract store inventory
		s.Find("div.inventory-collapse ul li").Each(func(j int, store *goquery.Selection) {
			// Extract address
			addressHTML, _ := store.Find("span.address").Html()
			address := parseAddress(addressHTML)

			// Extract quantity
			quantityText := store.Find("span.quantity").Text()
			quantity := extractQuantity(quantityText)

			if quantity > 0 {
				// Get store coordinates
				location, found := getStoreLocation(address)
				if !found {
					fmt.Fprintf(log.Writer(), "  WARNING: No coordinates found for address: %s\n", address)
					fmt.Fprintf(log.Writer(), "  Run ./scripts/update-wake-geocoding.sh to update store coordinates\n")
				}

				// Create inventory item
				item := tracker.InventoryItem{
					Timestamp:   now,
					ProductName: productName,
					ProductID:   fmt.Sprintf("wake-%s", plu), // Prefix with "wake-" to differentiate from VA codes
					Location:    location,
					Quantity:    quantity,
					StoreID:     fmt.Sprintf("wake-%s", sanitizeStoreID(address)),
					StoreURL:    "https://wakeabc.com/search-results",
					State:       "NC",
					County:      "Wake",
				}
				items = append(items, item)
			}
		})
	})

	return items, nil
}

// extractPLU extracts PLU number from text like "PLU: 18010"
func extractPLU(text string) string {
	re := regexp.MustCompile(`PLU:\s*(\d+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return "unknown"
}

// extractQuantity extracts quantity from text like "24 in stock"
func extractQuantity(text string) int {
	re := regexp.MustCompile(`(\d+)\s+in stock`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		qty, _ := strconv.Atoi(matches[1])
		return qty
	}
	return 0
}

// parseAddress extracts clean address from HTML like "7200 Sandy Fork Rd.<br/>Raleigh, NC 27609"
func parseAddress(htmlAddress string) string {
	// Replace <br/> with space
	address := strings.ReplaceAll(htmlAddress, "<br/>", " ")
	address = strings.ReplaceAll(address, "<br>", " ")
	// Remove extra spaces
	address = strings.Join(strings.Fields(address), " ")
	return address
}

// sanitizeStoreID creates a store ID from address
func sanitizeStoreID(address string) string {
	// Extract just the street address (before city)
	parts := strings.Split(address, ",")
	if len(parts) > 0 {
		street := strings.TrimSpace(parts[0])
		// Remove special characters
		street = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(street, "-")
		street = strings.Trim(street, "-")
		street = strings.ToLower(street)
		return street
	}
	return "unknown"
}
