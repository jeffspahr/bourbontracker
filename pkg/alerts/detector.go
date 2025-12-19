package alerts

import (
	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

// DetectChanges compares previous and current inventory to find new items,
// removed items, and quantity changes
func DetectChanges(previous, current []tracker.InventoryItem) *ComparisonResult {
	// Build maps keyed by (ProductID + StoreID)
	prevMap := make(map[string]tracker.InventoryItem)
	currMap := make(map[string]tracker.InventoryItem)

	for _, item := range previous {
		key := makeInventoryKey(item.ProductID, item.StoreID)
		prevMap[key] = item
	}

	for _, item := range current {
		key := makeInventoryKey(item.ProductID, item.StoreID)
		currMap[key] = item
	}

	result := &ComparisonResult{
		NewItems:        []tracker.InventoryItem{},
		RemovedItems:    []tracker.InventoryItem{},
		QuantityChanges: []QuantityChange{},
	}

	// Find new items (exists in current, not in previous)
	for key, item := range currMap {
		if prevItem, exists := prevMap[key]; !exists {
			result.NewItems = append(result.NewItems, item)
		} else {
			// Check for quantity changes
			if item.Quantity != prevItem.Quantity {
				result.QuantityChanges = append(result.QuantityChanges, QuantityChange{
					Item:        item,
					OldQuantity: prevItem.Quantity,
					NewQuantity: item.Quantity,
					Delta:       item.Quantity - prevItem.Quantity,
				})
			}
		}
	}

	// Find removed items (exists in previous, not in current)
	for key, item := range prevMap {
		if _, exists := currMap[key]; !exists {
			result.RemovedItems = append(result.RemovedItems, item)
		}
	}

	return result
}

// makeInventoryKey creates a unique key for product-store combinations
func makeInventoryKey(productID, storeID string) string {
	return productID + "-" + storeID
}
