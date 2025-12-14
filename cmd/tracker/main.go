package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jeffspahr/bourbontracker/pkg/nc/wake"
	"github.com/jeffspahr/bourbontracker/pkg/tracker"
	vaabc "github.com/jeffspahr/bourbontracker/pkg/va/abc"
)

var (
	storesFile     = flag.String("stores", "stores", "Path to stores file (VA ABC)")
	productsFile   = flag.String("products", "products.json", "Path to products file (VA ABC)")
	ncProductsFile = flag.String("nc-products", "nc-products.json", "Path to NC products file (Wake County)")
	outputFile     = flag.String("output", "inventory.json", "Path to output JSON file")
	enableVA       = flag.Bool("va", true, "Enable Virginia ABC tracker")
	enableWake     = flag.Bool("wake", false, "Enable Wake County NC tracker")
)

func main() {
	flag.Parse()

	// Load existing inventory to determine what needs updating
	existingInventory := loadExistingInventory(*outputFile)

	var allInventory []tracker.InventoryItem
	var trackers []tracker.Tracker

	// Initialize enabled trackers
	if *enableVA {
		va, err := vaabc.New(*storesFile, *productsFile)
		if err != nil {
			log.Fatalf("Failed to initialize VA ABC tracker: %v", err)
		}
		trackers = append(trackers, va)
	}

	if *enableWake {
		wakeTracker, err := wake.New(*ncProductsFile)
		if err != nil {
			log.Fatalf("Failed to initialize Wake County tracker: %v", err)
		}

		// Determine which products need updating based on age and listing type
		productsToUpdate := getProductsNeedingUpdate(wakeTracker, existingInventory)
		wakeTracker.SetProductsToTrack(productsToUpdate)

		trackers = append(trackers, wakeTracker)
	}

	if len(trackers) == 0 {
		log.Fatal("No trackers enabled. Use -va or -wake flags.")
	}

	// Run each tracker
	for _, t := range trackers {
		fmt.Fprintf(os.Stderr, "Running %s tracker...\n", t.Name())
		fmt.Fprintf(os.Stderr, "  Stores: %d\n", t.StoreCount())
		fmt.Fprintf(os.Stderr, "  Products: %d\n", len(t.ProductCodes()))

		startTime := time.Now()
		items, err := t.Track()
		duration := time.Since(startTime)

		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s tracker failed: %v\n", t.Name(), err)
			// Continue with other trackers instead of failing completely
			continue
		}

		fmt.Fprintf(os.Stderr, "  Completed in %v\n", duration)
		fmt.Fprintf(os.Stderr, "  Found %d items\n", len(items))

		allInventory = append(allInventory, items...)
	}

	// Merge new inventory with existing inventory (keep fresh NC data)
	finalInventory := mergeInventory(existingInventory, allInventory)

	// Write combined inventory to JSON file
	inventoryJSON, err := json.MarshalIndent(finalInventory, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal inventory: %v", err)
	}

	if err := ioutil.WriteFile(*outputFile, inventoryJSON, 0644); err != nil {
		log.Fatalf("Failed to write inventory file: %v", err)
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Printf("Found %d items in stock across all trackers\n", len(finalInventory))
	fmt.Fprintf(os.Stderr, "Inventory written to %s\n", *outputFile)
}

// loadExistingInventory loads the existing inventory file
func loadExistingInventory(filename string) []tracker.InventoryItem {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		// File doesn't exist or can't be read, return empty inventory
		fmt.Fprintf(os.Stderr, "No existing inventory found, will create fresh data\n")
		return []tracker.InventoryItem{}
	}

	var inventory []tracker.InventoryItem
	if err := json.Unmarshal(data, &inventory); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse existing inventory: %v\n", err)
		return []tracker.InventoryItem{}
	}

	return inventory
}

// getProductsNeedingUpdate determines which NC products need updating
func getProductsNeedingUpdate(wakeTracker *wake.Tracker, existingInventory []tracker.InventoryItem) []string {
	// Build map of product ID -> latest timestamp
	productTimestamps := make(map[string]time.Time)
	productListingTypes := make(map[string]string)

	for _, item := range existingInventory {
		// Only consider NC Wake County items
		if item.State != "NC" || item.County != "Wake" {
			continue
		}

		productID := item.ProductID
		if ts, exists := productTimestamps[productID]; !exists || item.Timestamp.After(ts) {
			productTimestamps[productID] = item.Timestamp
			productListingTypes[productID] = item.ListingType
		}
	}

	// Determine which products need updating
	var productsToUpdate []string
	now := time.Now()

	// Load NC products to get all product codes
	allProducts := wakeTracker.ProductCodes()

	for _, ncCode := range allProducts {
		timestamp, exists := productTimestamps[ncCode]
		if !exists {
			// Never tracked before, needs update
			productsToUpdate = append(productsToUpdate, ncCode)
			continue
		}

		listingType := productListingTypes[ncCode]
		age := now.Sub(timestamp)

		// "Listed" products: update every 24 hours
		// Other types (Limited, Allocation, Barrel, Christmas): update every hour
		needsUpdate := false
		if listingType == "Listed" {
			needsUpdate = age > 24*time.Hour
		} else {
			needsUpdate = age > 1*time.Hour
		}

		if needsUpdate {
			productsToUpdate = append(productsToUpdate, ncCode)
		}
	}

	return productsToUpdate
}

// mergeInventory merges new inventory with existing, replacing old NC data with new
func mergeInventory(existing, new []tracker.InventoryItem) []tracker.InventoryItem {
	// Create set of product IDs that were updated
	updatedProducts := make(map[string]bool)
	for _, item := range new {
		if item.State == "NC" && item.County == "Wake" {
			updatedProducts[item.ProductID] = true
		}
	}

	// Keep existing items that weren't updated
	var merged []tracker.InventoryItem
	for _, item := range existing {
		// Skip NC items that were updated
		if item.State == "NC" && item.County == "Wake" && updatedProducts[item.ProductID] {
			continue
		}
		// Keep VA items and NC items that weren't updated
		merged = append(merged, item)
	}

	// Add all new items
	merged = append(merged, new...)

	return merged
}
