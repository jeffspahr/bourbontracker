package tracker

import "time"

// InventoryItem represents a single bourbon product at a specific store location
type InventoryItem struct {
	Timestamp   time.Time `json:"@timestamp"`
	ProductName string    `json:"bt.productName"`
	ProductID   string    `json:"bt.productId"`
	Location    Location  `json:"geo.location"`
	Quantity    int       `json:"bt.quantity"`
	StoreID     string    `json:"bt.storeId"`
	StoreURL    string    `json:"bt.storeurl"`
	State       string    `json:"bt.state"`   // VA, NC, etc.
	County      string    `json:"bt.county"`  // For NC counties
}

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

// Tracker is the interface that all state/county trackers must implement
type Tracker interface {
	// Name returns the tracker name (e.g., "VA ABC", "NC Wake County")
	Name() string

	// Track queries the inventory and returns all items with quantity > 0
	Track() ([]InventoryItem, error)

	// ProductCodes returns the list of product codes this tracker should search for
	ProductCodes() []string

	// StoreCount returns the number of stores this tracker queries
	StoreCount() int
}

// Config holds common configuration for all trackers
type Config struct {
	// BaseDelay is the delay between requests to avoid rate limiting
	BaseDelay time.Duration

	// MaxRetries is the maximum number of retries per store
	MaxRetries int

	// Timeout is the HTTP request timeout
	Timeout time.Duration
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		BaseDelay:  250 * time.Millisecond,
		MaxRetries: 5,
		Timeout:    30 * time.Second,
	}
}
