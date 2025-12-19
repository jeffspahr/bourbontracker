package alerts

import (
	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

// ComparisonResult holds detected changes between inventory snapshots
type ComparisonResult struct {
	NewItems        []tracker.InventoryItem // Product appeared at store for first time
	RemovedItems    []tracker.InventoryItem // Product disappeared from store
	QuantityChanges []QuantityChange        // Quantity increased/decreased
}

// QuantityChange represents a change in quantity for an existing product-store combination
type QuantityChange struct {
	Item        tracker.InventoryItem
	OldQuantity int
	NewQuantity int
	Delta       int // NewQuantity - OldQuantity
}

// Subscriber represents a user subscription configuration
type Subscriber struct {
	ID          string      `json:"id"`
	Email       string      `json:"email"`
	Enabled     bool        `json:"enabled"`
	Preferences Preferences `json:"preferences"`
}

// Preferences defines user filtering preferences
type Preferences struct {
	States       []string `json:"states"`        // Filter by state codes (VA, NC)
	Counties     []string `json:"counties"`      // Filter by county names (NC only)
	ListingTypes []string `json:"listing_types"` // Filter by listing type (Allocation, Limited, etc.)
	Products     []string `json:"products"`      // Product name patterns (supports "*" wildcard)
	ProductIDs   []string `json:"product_ids"`   // Explicit product code filters
	MinQuantity  int      `json:"min_quantity"`  // Minimum quantity to trigger alert
	AlertOn      AlertOn  `json:"alert_on"`      // What changes trigger alerts
}

// AlertOn defines what types of changes trigger alerts
type AlertOn struct {
	NewProductAtStore bool `json:"new_product_at_store"` // Alert when product appears at new store
	QuantityIncrease  bool `json:"quantity_increase"`    // Alert when quantity increases
}

// Config represents the top-level subscriptions configuration
type Config struct {
	Version     string       `json:"version"`
	SMTP        SMTPConfig   `json:"smtp"`
	Subscribers []Subscriber `json:"subscribers"`
}

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Enabled   bool   `json:"enabled"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
}
