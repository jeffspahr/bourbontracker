package wake

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

// NCProduct represents a product from NC ABC warehouse
type NCProduct struct {
	NCCode      string `json:"nc_code"`
	BrandName   string `json:"brand_name"`
	ListingType string `json:"listing_type"`
	Size        string `json:"size"`
	Available   int    `json:"total_available"`
	Supplier    string `json:"supplier,omitempty"`
}

// Tracker implements the tracker.Tracker interface for Wake County, NC ABC
type Tracker struct {
	config          tracker.Config
	products        map[string]NCProduct // map of NC code to product info
	productsToTrack map[string]bool      // specific products to track (nil = track all)
	client          *http.Client
}

// New creates a new Wake County ABC tracker
func New(productsFile string) (*Tracker, error) {
	// Load NC products from JSON
	file, err := os.Open(productsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open products file: %w", err)
	}
	defer file.Close()

	var productList []NCProduct
	if err := json.NewDecoder(file).Decode(&productList); err != nil {
		return nil, fmt.Errorf("failed to parse products file: %w", err)
	}

	// Create map for quick lookup
	products := make(map[string]NCProduct)
	for _, p := range productList {
		products[p.NCCode] = p
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Tracker{
		config:          tracker.DefaultConfig(),
		products:        products,
		productsToTrack: nil, // nil means track all products
		client:          client,
	}, nil
}

// SetProductsToTrack sets specific products to track (by NC Code)
// If nil or empty, all products will be tracked
func (t *Tracker) SetProductsToTrack(ncCodes []string) {
	if len(ncCodes) == 0 {
		t.productsToTrack = nil
		return
	}
	t.productsToTrack = make(map[string]bool)
	for _, code := range ncCodes {
		t.productsToTrack[code] = true
	}
}

// Name returns the tracker name
func (t *Tracker) Name() string {
	return "NC Wake County ABC"
}

// ProductCodes returns the list of product codes
func (t *Tracker) ProductCodes() []string {
	codes := make([]string, 0, len(t.products))
	for code := range t.products {
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
	type result struct {
		items []tracker.InventoryItem
		err   error
	}

	// Count how many products we'll actually search
	productsToSearch := 0
	for ncCode := range t.products {
		if t.productsToTrack == nil || t.productsToTrack[ncCode] {
			productsToSearch++
		}
	}

	if productsToSearch == 0 {
		fmt.Fprintf(log.Writer(), "  No products need updating (all data is fresh)\n")
		return []tracker.InventoryItem{}, nil
	}

	fmt.Fprintf(log.Writer(), "  Searching %d/%d products\n", productsToSearch, len(t.products))

	// Create buffered channel for results
	results := make(chan result, productsToSearch)

	// Limit concurrent requests to avoid overwhelming the server
	const maxConcurrent = 15
	semaphore := make(chan struct{}, maxConcurrent)

	// Search by NC Code for each product concurrently
	for ncCode, product := range t.products {
		// Skip if we have a specific product list and this product isn't in it
		if t.productsToTrack != nil && !t.productsToTrack[ncCode] {
			continue
		}

		ncCode := ncCode // capture for goroutine
		product := product

		semaphore <- struct{}{} // acquire semaphore

		go func() {
			defer func() { <-semaphore }() // release semaphore

			// Rate limiting: 200ms delay per request
			time.Sleep(200 * time.Millisecond)

			items, err := t.searchProduct(ncCode, product)
			if err != nil {
				fmt.Fprintf(log.Writer(), "  ERROR searching %s: %v\n", ncCode, err)
			} else if len(items) > 0 {
				fmt.Fprintf(log.Writer(), "  Found %d items for %s (%s)\n", len(items), ncCode, product.BrandName)
			}

			results <- result{items: items, err: err}
		}()
	}

	// Collect results
	var allItems []tracker.InventoryItem
	for i := 0; i < productsToSearch; i++ {
		res := <-results
		if res.err == nil {
			allItems = append(allItems, res.items...)
		}
	}

	return allItems, nil
}

// searchProduct searches for a specific product by NC Code and parses results
func (t *Tracker) searchProduct(ncCode string, product NCProduct) ([]tracker.InventoryItem, error) {
	// Try searching by NC Code first
	formData := url.Values{}
	formData.Set("productSearch", ncCode)

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

	return t.parseSearchResults(ncCode, product, string(body))
}

// parseSearchResults extracts inventory items from HTML
func (t *Tracker) parseSearchResults(ncCode string, product NCProduct, html string) ([]tracker.InventoryItem, error) {
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

				// Create search URL with NC Code
				searchURL := fmt.Sprintf("https://wakeabc.com/search-our-inventory/?productSearch=%s",
					url.QueryEscape(ncCode))

				// Create inventory item
				item := tracker.InventoryItem{
					Timestamp:   now,
					ProductName: tracker.NormalizeProductName(productName),
					ProductID:   ncCode, // Use NC Code as product ID
					Location:    location,
					Quantity:    quantity,
					StoreID:     fmt.Sprintf("wake-%s", sanitizeStoreID(address)),
					StoreURL:    searchURL,
					State:       "NC",
					County:      "Wake",
					ListingType: product.ListingType, // Add listing type
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
