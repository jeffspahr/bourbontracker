package wake

import (
	"errors"

	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

// Tracker implements the tracker.Tracker interface for Wake County, NC ABC
type Tracker struct {
	config tracker.Config
}

// New creates a new Wake County ABC tracker
func New() (*Tracker, error) {
	return &Tracker{
		config: tracker.DefaultConfig(),
	}, nil
}

// Name returns the tracker name
func (t *Tracker) Name() string {
	return "NC Wake County ABC"
}

// ProductCodes returns the list of product codes (placeholder)
func (t *Tracker) ProductCodes() []string {
	// TODO: Implement product codes for Wake County
	// Wake County may use different product identification than VA
	return []string{}
}

// StoreCount returns the number of stores (placeholder)
func (t *Tracker) StoreCount() int {
	// TODO: Determine how Wake County organizes stores
	// May be a single county-wide inventory or multiple locations
	return 0
}

// Track queries Wake County inventory and returns items
func (t *Tracker) Track() ([]tracker.InventoryItem, error) {
	// TODO: Implement Wake County tracking
	// Options:
	// 1. Web scraping https://wakeabc.com/search-results
	// 2. API integration if one becomes available
	// 3. Browser automation for JavaScript-heavy sites
	//
	// Wake County structure is different from VA:
	// - May not have individual store numbers
	// - Products may be searched by name, not code
	// - Inventory may be county-wide, not per-store
	//
	// Implementation notes:
	// - Search endpoint: https://wakeabc.com/?s={product_name}
	// - Parse HTML results for product availability
	// - Extract store locations if available
	// - Map to common InventoryItem format

	return nil, errors.New("Wake County tracker not yet implemented")
}

// TODO: Helper functions for Wake County implementation
// - func (t *Tracker) searchProduct(name string) ([]tracker.InventoryItem, error)
// - func (t *Tracker) parseSearchResults(html string) ([]tracker.InventoryItem, error)
// - func (t *Tracker) getStoreLocations() (map[string]tracker.Location, error)
