package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jeffspahr/bourbontracker/pkg/alerts"
	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

var (
	previousVAFile    = flag.String("previous-va", "", "Path to previous VA inventory JSON")
	previousNCFile    = flag.String("previous-nc", "", "Path to previous NC inventory JSON")
	currentVAFile     = flag.String("current-va", "", "Path to current VA inventory JSON")
	currentNCFile     = flag.String("current-nc", "", "Path to current NC inventory JSON")
	subscriptionsFile = flag.String("subscriptions", "", "Path to subscriptions config file")
	dryRun            = flag.Bool("dry-run", false, "Print email preview instead of sending")
)

func main() {
	flag.Parse()

	// Load inventories
	previousVA := loadInventory(*previousVAFile)
	previousNC := loadInventory(*previousNCFile)
	currentVA := loadInventory(*currentVAFile)
	currentNC := loadInventory(*currentNCFile)

	// Combine VA and NC inventories
	previous := append(previousVA, previousNC...)
	current := append(currentVA, currentNC...)

	// Skip alerts if no previous inventory (avoid spam on first run)
	if len(previous) == 0 {
		log.Println("No previous inventory - skipping alerts on first run")
		return
	}

	// Detect changes
	changes := alerts.DetectChanges(previous, current)

	log.Printf("Detected %d new items, %d removed items, %d quantity changes",
		len(changes.NewItems), len(changes.RemovedItems), len(changes.QuantityChanges))

	if len(changes.NewItems) == 0 {
		log.Println("No new items detected - no alerts to send")
		return
	}

	// Load subscriptions config
	if *subscriptionsFile == "" {
		log.Println("No subscriptions file specified (-subscriptions flag)")
		log.Println("Use -subscriptions to enable multi-user alerts")
		return
	}

	config, err := alerts.LoadConfig(*subscriptionsFile)
	if err != nil {
		log.Fatalf("Failed to load subscriptions config: %v", err)
	}

	subscribers := alerts.GetEnabledSubscribers(config)
	log.Printf("Loaded %d enabled subscriber(s)", len(subscribers))

	if len(subscribers) == 0 {
		log.Println("No enabled subscribers - no alerts to send")
		return
	}

	// Filter changes for each subscriber based on their preferences
	itemsPerSubscriber := make(map[string][]tracker.InventoryItem)
	totalMatches := 0

	for _, sub := range subscribers {
		filtered := alerts.FilterForSubscriber(changes.NewItems, sub.Preferences)
		if len(filtered) > 0 {
			itemsPerSubscriber[sub.ID] = filtered
			totalMatches += len(filtered)
			log.Printf("Subscriber %s: %d matching items", sub.ID, len(filtered))
		} else {
			log.Printf("Subscriber %s: no matches", sub.ID)
		}
	}

	if totalMatches == 0 {
		log.Println("No items matched any subscriber preferences - no alerts to send")
		return
	}

	// Dry run mode: print previews instead of sending
	if *dryRun {
		log.Println("\n=== DRY RUN MODE - Email Previews ===")
		for _, sub := range subscribers {
			items, ok := itemsPerSubscriber[sub.ID]
			if !ok || len(items) == 0 {
				continue
			}

			fmt.Printf("\n--- Email for %s (%s) ---\n", sub.ID, sub.Email)
			fmt.Printf("Subject: Cask Watch Alert: %d New Allocation Item%s\n",
				len(items), pluralize(len(items)))
			fmt.Printf("Items:\n")
			for _, item := range items {
				fmt.Printf("  - %s (%s, %d bottles)\n", item.ProductName, item.StoreID, item.Quantity)
			}
		}
		return
	}

	// Create mailer
	mailer, err := alerts.NewMailer(config.SMTP.FromEmail, config.SMTP.FromName)
	if err != nil {
		log.Fatalf("Failed to initialize mailer: %v", err)
	}

	// Send alerts
	if err := mailer.SendAlertBatch(subscribers, itemsPerSubscriber); err != nil {
		log.Printf("Warning: Some emails failed to send: %v", err)
		os.Exit(1)
	}

	log.Println("Alert batch complete")
}

// loadInventory loads inventory from a JSON file
func loadInventory(filePath string) []tracker.InventoryItem {
	if filePath == "" {
		return []tracker.InventoryItem{}
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		// File might not exist (first run or tracker failure)
		log.Printf("Warning: Could not read %s: %v", filePath, err)
		return []tracker.InventoryItem{}
	}

	var items []tracker.InventoryItem
	if err := json.Unmarshal(data, &items); err != nil {
		log.Printf("Warning: Could not parse %s: %v", filePath, err)
		return []tracker.InventoryItem{}
	}

	return items
}

// pluralize returns "s" if count != 1, otherwise empty string
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
