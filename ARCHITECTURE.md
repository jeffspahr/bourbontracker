# Bourbon Tracker Architecture

## Overview

The bourbon tracker has been refactored into a modular, extensible architecture that supports multiple states and counties, each with their own inventory APIs.

## Project Structure

```
bourbontracker/
├── cmd/
│   └── tracker/
│       └── main.go          # Main entry point - orchestrates all trackers
├── pkg/
│   ├── tracker/
│   │   └── tracker.go       # Common tracker interface and types
│   ├── va/
│   │   └── abc/
│   │       └── tracker.go   # Virginia ABC implementation
│   └── nc/
│       └── wake/
│           └── tracker.go   # Wake County, NC implementation (placeholder)
├── stores                   # VA ABC store list
├── products.json            # Product codes to track
├── inventory.json           # Combined output from all trackers
└── map.html                 # Google Maps visualization
```

## Core Interface

All trackers implement the `tracker.Tracker` interface:

```go
type Tracker interface {
    Name() string                          // e.g., "VA ABC", "NC Wake County"
    Track() ([]InventoryItem, error)       // Query inventory
    ProductCodes() []string                // Product IDs to search
    StoreCount() int                       // Number of locations
}
```

## Inventory Item Format

All trackers output to a common format:

```go
type InventoryItem struct {
    Timestamp   time.Time  `json:"@timestamp"`
    ProductName string     `json:"bt.productName"`
    ProductID   string     `json:"bt.productId"`
    Location    Location   `json:"geo.location"`
    Quantity    int        `json:"bt.quantity"`
    StoreID     string     `json:"bt.storeId"`
    StoreURL    string     `json:"bt.storeurl"`
    State       string     `json:"bt.state"`    // VA, NC, etc.
    County      string     `json:"bt.county"`   // For NC counties
}
```

## State/County Implementations

### Virginia ABC (`pkg/va/abc`)

**API:** REST API at `https://www.abc.virginia.gov/webapi/inventory/mystore`

**Features:**
- Centralized statewide inventory system
- 390 stores across Virginia
- Product code-based search
- Per-store inventory quantities
- Geographic coordinates for each store

**Implementation:**
- HTTP requests with User-Agent headers
- 250ms rate limiting between requests
- Exponential backoff for failed requests
- Skip stores after 5 failed attempts

### Wake County, NC (`pkg/nc/wake`)

**Status:** Placeholder - Not yet implemented

**Website:** https://wakeabc.com/search-results

**Implementation Options:**
1. Web scraping (no public API available)
2. Form submission automation
3. HTML parsing for results

**Differences from VA:**
- County-level system (not statewide)
- Product name search (not codes)
- May not have per-store inventory
- Requires different data extraction approach

### Future Counties

Additional NC counties can be added under `pkg/nc/<county>/`:
- Durham County
- Orange County
- Chatham County
- etc.

Each county may have completely different inventory systems.

## Running Trackers

### Command Line Flags

```bash
./tracker \
  -va              # Enable VA ABC (default: true)
  -wake            # Enable Wake County NC (default: false)
  -stores FILE     # VA ABC stores file (default: "stores")
  -products FILE   # Products file (default: "products.json")
  -output FILE     # Output JSON (default: "inventory.json")
```

### Examples

**Virginia only (default):**
```bash
./tracker
```

**Virginia + Wake County:**
```bash
./tracker -va -wake
```

**Wake County only:**
```bash
./tracker -va=false -wake
```

## Adding a New Tracker

1. **Create package directory:**
   ```bash
   mkdir -p pkg/<state>/<county or region>
   ```

2. **Implement the interface:**
   ```go
   package yourtracker

   import "github.com/jeffspahr/bourbontracker/pkg/tracker"

   type Tracker struct {
       config tracker.Config
       // Add fields for your specific implementation
   }

   func New(...) (*Tracker, error) { ... }
   func (t *Tracker) Name() string { ... }
   func (t *Tracker) Track() ([]tracker.InventoryItem, error) { ... }
   func (t *Tracker) ProductCodes() []string { ... }
   func (t *Tracker) StoreCount() int { ... }
   ```

3. **Register in main.go:**
   ```go
   // Add flag
   enableYourTracker = flag.Bool("yourtracker", false, "Enable your tracker")

   // Add initialization
   if *enableYourTracker {
       t, err := yourtracker.New(...)
       trackers = append(trackers, t)
   }
   ```

4. **Test:**
   ```bash
   go build -o tracker ./cmd/tracker
   ./tracker -yourtracker
   ```

## Design Principles

1. **Modularity:** Each state/county is independent
2. **Common Interface:** All trackers produce the same output format
3. **Extensibility:** Easy to add new regions
4. **Flexibility:** Each tracker can use different APIs/methods
5. **Error Isolation:** One tracker failure doesn't stop others
6. **Backward Compatibility:** Maintains existing inventory.json format

## Rate Limiting & Best Practices

### Common Configuration

```go
type Config struct {
    BaseDelay  time.Duration  // Delay between requests (250ms)
    MaxRetries int            // Max retries per store (5)
    Timeout    time.Duration  // HTTP timeout (30s)
}
```

### Guidelines

- **Respect APIs:** Always add delays between requests
- **User-Agent:** Use browser-like headers to avoid blocking
- **Retry Logic:** Implement exponential backoff
- **Error Handling:** Skip problematic stores, don't fail entirely
- **Logging:** Output progress to stderr, results to stdout

## Output Format

All trackers write to a single `inventory.json` file:

```json
[
  {
    "@timestamp": "2025-12-13T10:00:00Z",
    "bt.productName": "Pappy Van Winkle 23yr",
    "bt.productId": "021030",
    "geo.location": {"lat": 37.5407, "lon": -77.4364},
    "bt.quantity": 2,
    "bt.storeId": "42",
    "bt.storeurl": "https://...",
    "bt.state": "VA",
    "bt.county": ""
  },
  {
    "@timestamp": "2025-12-13T10:05:00Z",
    "bt.productName": "Blanton's Single Barrel",
    "bt.productId": "000485",
    "geo.location": {"lat": 35.7796, "lon": -78.6382},
    "bt.quantity": 5,
    "bt.storeId": "wake-001",
    "bt.storeurl": "https://wakeabc.com/...",
    "bt.state": "NC",
    "bt.county": "Wake"
  }
]
```

## Future Enhancements

- Configuration file (YAML/JSON) for enabling trackers
- Parallel tracker execution for faster runs
- Database storage for historical tracking
- API server mode for real-time queries
- Webhook notifications for new inventory
- Product name normalization across different systems
