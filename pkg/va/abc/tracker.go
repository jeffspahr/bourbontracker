package abc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

// Tracker implements the tracker.Tracker interface for Virginia ABC
type Tracker struct {
	config       tracker.Config
	stores       []string
	products     map[string]string
	storeRetries map[int]int
	waitTime     int
}

// payloadIn represents the Virginia ABC API response
type payloadIn struct {
	Products []struct {
		ProductID string `json:"productId"`
		StoreInfo struct {
			Distance    interface{} `json:"distance"`
			Latitude    float64     `json:"latitude"`
			Longitude   float64     `json:"longitude"`
			PhoneNumber struct {
				AreaCode             string `json:"AreaCode"`
				FormattedPhoneNumber string `json:"FormattedPhoneNumber"`
				LineNumber           string `json:"LineNumber"`
				Prefix               string `json:"Prefix"`
			} `json:"phoneNumber"`
			Quantity int `json:"quantity"`
			StoreID  int `json:"storeId"`
		} `json:"storeInfo"`
	} `json:"products"`
	URL string `json:"url"`
}

// New creates a new Virginia ABC tracker
func New(storesFile, productsFile string) (*Tracker, error) {
	t := &Tracker{
		config:       tracker.DefaultConfig(),
		storeRetries: make(map[int]int),
		waitTime:     1,
	}

	// Load stores
	if err := t.loadStores(storesFile); err != nil {
		return nil, fmt.Errorf("failed to load stores: %w", err)
	}

	// Load products
	if err := t.loadProducts(productsFile); err != nil {
		return nil, fmt.Errorf("failed to load products: %w", err)
	}

	return t, nil
}

// loadStores reads the store list from a file
func (t *Tracker) loadStores(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t.stores = append(t.stores, scanner.Text())
	}

	return scanner.Err()
}

// loadProducts reads the product list from a JSON file
func (t *Tracker) loadProducts(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(&t.products)
}

// Name returns the tracker name
func (t *Tracker) Name() string {
	return "VA ABC"
}

// ProductCodes returns the list of product codes
func (t *Tracker) ProductCodes() []string {
	codes := make([]string, 0, len(t.products))
	for code := range t.products {
		codes = append(codes, code)
	}
	return codes
}

// StoreCount returns the number of stores
func (t *Tracker) StoreCount() int {
	return len(t.stores)
}

// Track queries all stores and returns inventory items
func (t *Tracker) Track() ([]tracker.InventoryItem, error) {
	var inventory []tracker.InventoryItem

	// Create comma-delimited product list for query string
	productListString := ""
	for key := range t.products {
		if productListString == "" {
			productListString = key
		} else {
			productListString += "," + key
		}
	}

	for h := 0; h < len(t.stores); h++ {
		// Sleep before each request except the first
		if h > 0 {
			time.Sleep(t.config.BaseDelay)
		}

		client := &http.Client{Timeout: t.config.Timeout}
		req, err := http.NewRequest("GET", "https://www.abc.virginia.gov/webapi/inventory/mystore", nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("Content-type", "application/json")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Add("Referer", "https://www.abc.virginia.gov/")

		q := req.URL.Query()
		q.Add("storeNumbers", t.stores[h])
		q.Add("productCodes", productListString)
		req.URL.RawQuery = q.Encode()

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Handle non-200 responses
		if resp.StatusCode != 200 {
			t.storeRetries[h]++
			fmt.Fprintf(os.Stderr, "Got HTTP %d for store %s (retry %d, backoff %ds)\n",
				resp.StatusCode, t.stores[h], t.storeRetries[h], t.waitTime)

			// Skip stores that consistently fail
			if t.storeRetries[h] >= t.config.MaxRetries {
				fmt.Fprintf(os.Stderr, "Skipping store %s after %d failed attempts\n",
					t.stores[h], t.storeRetries[h])
				t.waitTime = 1 // Reset backoff
				continue
			}

			time.Sleep(time.Duration(t.waitTime) * time.Second)
			t.waitTime = t.waitTime * 2
			if t.waitTime > 512 {
				t.waitTime = 512 // Cap backoff
			}
			h-- // Retry this store
			continue
		}

		// Gradually reduce backoff on success
		if resp.StatusCode == 200 && t.waitTime > 1 {
			t.waitTime = t.waitTime / 2
		}

		// Parse response
		var pIn payloadIn
		if err := json.Unmarshal(body, &pIn); err != nil {
			return nil, fmt.Errorf("failed to parse response for store %s: %w", t.stores[h], err)
		}

		// Convert to common inventory format
		for i := range pIn.Products {
			if pIn.Products[i].StoreInfo.Quantity <= 0 {
				continue // Skip items with no quantity
			}

			storeID, _ := strconv.Atoi(t.stores[h])
			item := tracker.InventoryItem{
				Timestamp:   time.Now(),
				ProductName: tracker.NormalizeProductName(t.products[pIn.Products[i].ProductID]),
				ProductID:   pIn.Products[i].ProductID,
				Location: tracker.Location{
					Latitude:  pIn.Products[i].StoreInfo.Latitude,
					Longitude: pIn.Products[i].StoreInfo.Longitude,
				},
				Quantity: pIn.Products[i].StoreInfo.Quantity,
				StoreID:  strconv.Itoa(storeID),
				StoreURL: "https://www.abc.virginia.gov/" + pIn.URL,
				State:    "VA",
				County:   "",
			}
			inventory = append(inventory, item)
		}
	}

	return inventory, nil
}
