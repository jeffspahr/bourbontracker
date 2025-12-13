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
	storesFile   = flag.String("stores", "stores", "Path to stores file (VA ABC)")
	productsFile = flag.String("products", "products.json", "Path to products file")
	outputFile   = flag.String("output", "inventory.json", "Path to output JSON file")
	enableVA     = flag.Bool("va", true, "Enable Virginia ABC tracker")
	enableWake   = flag.Bool("wake", false, "Enable Wake County NC tracker")
)

func main() {
	flag.Parse()

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
		wakeTracker, err := wake.New()
		if err != nil {
			log.Fatalf("Failed to initialize Wake County tracker: %v", err)
		}
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

	// Write combined inventory to JSON file
	inventoryJSON, err := json.MarshalIndent(allInventory, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal inventory: %v", err)
	}

	if err := ioutil.WriteFile(*outputFile, inventoryJSON, 0644); err != nil {
		log.Fatalf("Failed to write inventory file: %v", err)
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Printf("Found %d items in stock across all trackers\n", len(allInventory))
	fmt.Fprintf(os.Stderr, "Inventory written to %s\n", *outputFile)
}
