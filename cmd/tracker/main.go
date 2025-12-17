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
	outputVAFile   = flag.String("output-va", "inventory-va.json", "Path to VA output JSON file")
	outputNCFile   = flag.String("output-nc", "inventory-nc.json", "Path to NC output JSON file")
	enableVA       = flag.Bool("va", true, "Enable Virginia ABC tracker")
	enableWake     = flag.Bool("wake", false, "Enable Wake County NC tracker")
)

func main() {
	flag.Parse()

	var vaInventory []tracker.InventoryItem
	var ncInventory []tracker.InventoryItem

	// Run VA tracker
	if *enableVA {
		va, err := vaabc.New(*storesFile, *productsFile)
		if err != nil {
			log.Fatalf("Failed to initialize VA ABC tracker: %v", err)
		}

		fmt.Fprintf(os.Stderr, "Running %s tracker...\n", va.Name())
		fmt.Fprintf(os.Stderr, "  Stores: %d\n", va.StoreCount())
		fmt.Fprintf(os.Stderr, "  Products: %d\n", len(va.ProductCodes()))

		startTime := time.Now()
		items, err := va.Track()
		duration := time.Since(startTime)

		if err != nil {
			log.Fatalf("ERROR: %s tracker failed: %v\n", va.Name(), err)
		}

		fmt.Fprintf(os.Stderr, "  Completed in %v\n", duration)
		fmt.Fprintf(os.Stderr, "  Found %d items\n", len(items))

		vaInventory = items

		// Write VA inventory
		inventoryJSON, err := json.MarshalIndent(vaInventory, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal VA inventory: %v", err)
		}

		if err := ioutil.WriteFile(*outputVAFile, inventoryJSON, 0644); err != nil {
			log.Fatalf("Failed to write VA inventory file: %v", err)
		}

		fmt.Fprintf(os.Stderr, "  Written to %s\n", *outputVAFile)
	}

	// Run Wake County tracker
	if *enableWake {
		// Load existing NC inventory for caching
		existingNCInventory := loadExistingInventory(*outputNCFile)

		wakeTracker, err := wake.New(*ncProductsFile)
		if err != nil {
			log.Fatalf("Failed to initialize Wake County tracker: %v", err)
		}

		// Determine which products need updating based on age and listing type
		productsToUpdate := getProductsNeedingUpdate(wakeTracker, existingNCInventory)
		wakeTracker.SetProductsToTrack(productsToUpdate)

		fmt.Fprintf(os.Stderr, "Running %s tracker...\n", wakeTracker.Name())
		fmt.Fprintf(os.Stderr, "  Stores: %d\n", wakeTracker.StoreCount())
		fmt.Fprintf(os.Stderr, "  Products: %d (updating %d)\n", len(wakeTracker.ProductCodes()), len(productsToUpdate))

		startTime := time.Now()
		items, err := wakeTracker.Track()
		duration := time.Since(startTime)

		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s tracker failed: %v\n", wakeTracker.Name(), err)
			// Use existing inventory if tracking fails
			ncInventory = existingNCInventory
		} else {
			fmt.Fprintf(os.Stderr, "  Completed in %v\n", duration)
			fmt.Fprintf(os.Stderr, "  Found %d items\n", len(items))

			// Merge new NC data with existing NC data
			ncInventory = mergeInventory(existingNCInventory, items)
		}

		// Write NC inventory
		inventoryJSON, err := json.MarshalIndent(ncInventory, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal NC inventory: %v", err)
		}

		if err := ioutil.WriteFile(*outputNCFile, inventoryJSON, 0644); err != nil {
			log.Fatalf("Failed to write NC inventory file: %v", err)
		}

		fmt.Fprintf(os.Stderr, "  Written to %s\n", *outputNCFile)
	}

	if !*enableVA && !*enableWake {
		log.Fatal("No trackers enabled. Use -va or -wake flags.")
	}

	fmt.Fprintf(os.Stderr, "\n")
	totalItems := len(vaInventory) + len(ncInventory)
	fmt.Printf("Found %d items in stock across all trackers\n", totalItems)
	if *enableVA {
		fmt.Fprintf(os.Stderr, "  VA: %d items\n", len(vaInventory))
	}
	if *enableWake {
		fmt.Fprintf(os.Stderr, "  NC: %d items\n", len(ncInventory))
	}
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
