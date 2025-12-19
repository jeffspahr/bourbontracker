package alerts

import (
	"strings"

	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

// FilterForSubscriber filters inventory items based on subscriber preferences
func FilterForSubscriber(items []tracker.InventoryItem, prefs Preferences) []tracker.InventoryItem {
	var filtered []tracker.InventoryItem

	for _, item := range items {
		if matchesPreferences(item, prefs) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// matchesPreferences checks if an item matches all subscriber preferences
func matchesPreferences(item tracker.InventoryItem, prefs Preferences) bool {
	// State filter
	if len(prefs.States) > 0 && !contains(prefs.States, item.State) {
		return false
	}

	// County filter (NC only)
	if len(prefs.Counties) > 0 && !contains(prefs.Counties, item.County) {
		return false
	}

	// Listing type filter (NC only - VA items don't have this field)
	if len(prefs.ListingTypes) > 0 {
		// For NC items, check if listing type matches
		if item.State == "NC" {
			if item.ListingType == "" || !contains(prefs.ListingTypes, item.ListingType) {
				return false
			}
		}
		// VA items pass through if they have listing type filter
		// (VA doesn't have listing types, so we can't filter on them)
	}

	// Product name filter
	if len(prefs.Products) > 0 && !matchesProductFilter(item.ProductName, prefs.Products) {
		return false
	}

	// Product ID filter
	if len(prefs.ProductIDs) > 0 && !matchesProductIDFilter(item.ProductID, prefs.ProductIDs) {
		return false
	}

	// Quantity threshold
	if item.Quantity < prefs.MinQuantity {
		return false
	}

	return true
}

// contains checks if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// matchesProductFilter checks if a product name matches any of the filter patterns
func matchesProductFilter(productName string, filters []string) bool {
	// Normalize the product name using tracker's normalization function
	normalizedProduct := tracker.NormalizeProductName(productName)

	for _, filter := range filters {
		// Wildcard matches everything
		if filter == "*" {
			return true
		}

		// Normalize the filter for consistent matching
		normalizedFilter := tracker.NormalizeProductName(filter)

		// Check if the product name contains the filter (substring match)
		if strings.Contains(strings.ToLower(normalizedProduct), strings.ToLower(normalizedFilter)) {
			return true
		}
	}

	return false
}

// matchesProductIDFilter checks if a product ID matches any of the filter IDs
func matchesProductIDFilter(productID string, filterIDs []string) bool {
	for _, filterID := range filterIDs {
		// Direct match
		if productID == filterID {
			return true
		}

		// Handle NC product IDs with "wake-" prefix
		// NC product IDs are often prefixed with "wake-"
		if strings.HasPrefix(productID, "wake-") {
			unprefixed := strings.TrimPrefix(productID, "wake-")
			if unprefixed == filterID {
				return true
			}
		}

		// Also check if filterID with wake- prefix matches
		if "wake-"+filterID == productID {
			return true
		}
	}

	return false
}
