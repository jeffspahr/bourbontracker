# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Cask Watch** is a multi-state spirits inventory tracker that monitors rare bourbon availability across ABC stores. It features:
- Go-based tracker querying multiple state/county APIs (Virginia ABC + Wake County NC)
- JSON output consumed by an interactive Google Maps visualization
- Automated deployment to Cloudflare Pages with OAuth protection
- Docker support with multi-arch builds (amd64/arm64)

## Build & Run Commands

### Local Development

```bash
# Build the tracker
go build -o tracker ./cmd/tracker

# Run Virginia ABC only (default)
./tracker

# Run both Virginia + Wake County
./tracker -va -wake

# Run Wake County only
./tracker -va=false -wake

# Custom output location
./tracker -output my-inventory.json

# Custom product/store lists (VA ABC only)
./tracker -products custom-products.json -stores custom-stores
```

### Docker

```bash
# Build multi-arch image
docker buildx build --platform linux/amd64,linux/arm64 -t bourbontracker .

# Run and save inventory.json to current directory
docker run --rm -v $(pwd):/root ghcr.io/jeffspahr/bourbontracker:latest

# With custom flags
docker run --rm -v $(pwd):/root ghcr.io/jeffspahr/bourbontracker:latest -va -wake
```

### Web Visualization

```bash
# Serve the map locally (after running tracker)
python3 -m http.server 8000

# Open http://localhost:8000 in browser
```

### CI/CD

```bash
# Trigger Cloudflare deployment manually
gh workflow run "Deploy to Cloudflare Pages"

# Check deployment status
gh run list --workflow="Deploy to Cloudflare Pages" --limit 1
gh run watch <run-id>
```

## Architecture

### Tracker System (Go)

**Modular Plugin Architecture**: Each state/county implements a common `tracker.Tracker` interface:

```go
type Tracker interface {
    Name() string                          // e.g., "VA ABC", "NC Wake County"
    Track() ([]InventoryItem, error)       // Query inventory
    ProductCodes() []string                // Product IDs to search
    StoreCount() int                       // Number of locations
}
```

**Key Design Principles**:
- **Error Isolation**: One tracker failure doesn't stop others
- **Different APIs per Region**: VA uses REST API, Wake County uses HTML scraping
- **Unified Output**: All trackers write to common `inventory.json` format
- **Rate Limiting**: Built-in delays and exponential backoff to respect APIs

**File Structure**:
```
pkg/
├── tracker/tracker.go          # Common interface + InventoryItem struct
├── va/abc/tracker.go           # VA ABC REST API implementation
└── nc/wake/tracker.go          # Wake County HTML scraping implementation
```

**Critical Implementation Details**:

1. **Product Name Normalization** (`pkg/tracker/tracker.go:NormalizeProductName()`):
   - Standardizes product names across different sources
   - Handles variations like "Blantons" → "Blanton's"
   - Normalizes year suffixes and bottle sizes
   - Essential for consistent filtering in the UI

2. **VA ABC Tracker** (`pkg/va/abc/tracker.go`):
   - Uses numeric product codes (e.g., `018006` for Buffalo Trace)
   - Queries API: `https://www.abc.virginia.gov/webapi/inventory/mystore`
   - 250ms delay between stores, exponential backoff on errors
   - Skips stores after 5 consecutive failures
   - Returns coordinates from API

3. **Wake County Tracker** (`pkg/nc/wake/tracker.go`):
   - Uses product names (not codes) for search
   - POST requests to `https://wakeabc.com/search-results`
   - Parses HTML with goquery: `<div class="wake-product">`
   - Coordinates must be manually geocoded (see `pkg/nc/wake/geocoding.go`)
   - Product IDs prefixed with `wake-` to avoid conflicts

### Frontend (JavaScript/HTML)

**Single-File Architecture**: `index.html` contains embedded CSS, JavaScript, and map logic.

**Key Features**:
- **IP Geolocation**: Auto-selects closest region using ipapi.co
- **Region Filtering**: VA vs NC-Wake County
- **Product Filtering**: Multi-select dropdown with normalized names
- **Store Grouping**: Multiple products at same location collapsed into one marker
- **Color Coding**: Green (>10 items), Orange (6-10), Red (1-5)

**API Key Handling**:
- Google Maps API key embedded inline during deployment (not separate config.js)
- Sed replacement in GitHub Actions: `.github/workflows/deploy-cloudflare.yml:34`
- Bypasses Cloudflare Access blocking issues

**Important Functions**:
- `setDefaultRegionByLocation()`: Haversine distance calculation to nearest region
- `applyFilters()`: Combined region + product filtering logic
- `NormalizeProductName()`: Must match Go implementation for consistency

### Deployment Pipeline

**Automated Cloudflare Pages Deployment**:
1. Runs on schedule (every 6 hours) or manual trigger
2. Builds Go tracker binary
3. Executes tracker to generate fresh `inventory.json`
4. Injects Google Maps API key directly into HTML
5. Deploys to Cloudflare Pages with OAuth protection

**Key Files**:
- `.github/workflows/deploy-cloudflare.yml`: Main deployment workflow
- `.github/workflows/main.yml`: Docker build/test CI pipeline
- `cloudflare/DEPLOYMENT.md`: OAuth setup instructions

**Secrets Required** (GitHub Actions):
- `CLOUDFLARE_API_TOKEN`
- `CLOUDFLARE_ACCOUNT_ID`
- `GOOGLE_MAPS_API_KEY`

## Adding a New State/County Tracker

1. **Create tracker package**:
```bash
mkdir -p pkg/<state>/<county>
```

2. **Implement interface** in `pkg/<state>/<county>/tracker.go`:
```go
package yourtracker

import "github.com/jeffspahr/bourbontracker/pkg/tracker"

type Tracker struct {
    config tracker.Config
}

func New() (*Tracker, error) { ... }
func (t *Tracker) Name() string { return "Your Tracker" }
func (t *Tracker) Track() ([]tracker.InventoryItem, error) { ... }
func (t *Tracker) ProductCodes() []string { ... }
func (t *Tracker) StoreCount() int { ... }
```

3. **Register in `cmd/tracker/main.go`**:
```go
// Add flag
enableYourTracker = flag.Bool("yourtracker", false, "Enable your tracker")

// Add initialization
if *enableYourTracker {
    t, err := yourtracker.New()
    if err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }
    trackers = append(trackers, t)
}
```

4. **Update UI** in `index.html`:
   - Add region to `REGION_CENTERS` constant
   - Add option to `#region-select` dropdown

5. **Important**: Use `tracker.NormalizeProductName()` for all product names

## Testing

**No unit tests currently exist**. Testing is done via:
- CI pipeline container validation (`.github/workflows/main.yml:78-162`)
- Docker image tests verify binary exists and runs without errors
- Manual testing via deployments

## Git Conventions

**Commit Messages**: Use conventional commits format:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `ci:` - CI/CD changes

**Committer**: Jeff Spahr <spahrj@gmail.com>

**Never commit**:
- API keys or secrets
- `config.js` (gitignored - contains Google Maps API key)
- `inventory.json` (generated output)

## Common Gotchas

1. **Cloudflare Access blocks all resources**: External scripts fail behind OAuth. Solution: Embed API keys inline during deployment.

2. **CORS errors with inventory.json**: File must be served via HTTP server, not `file://` protocol.

3. **Product name mismatches**: UI filter won't work if Go normalization differs from frontend. Always use `NormalizeProductName()`.

4. **Wake County coordinates missing**: New stores require manual geocoding via `scripts/update-wake-geocoding.sh`.

5. **Rate limiting**: VA ABC API will return 429/503 if requests are too fast. Respect the 250ms delay.

6. **Mixed content warnings**: Always use HTTPS URLs for marker icons, not HTTP.

## Dependencies

- **Go**: 1.25
- **goquery**: HTML parsing for Wake County scraper
- **Google Maps JavaScript API**: Frontend visualization
- **Cloudflare Pages**: Hosting with OAuth protection
- **ipapi.co**: Free IP geolocation API

## Monitoring

- Cloudflare deployment runs every 6 hours (cron: `0 */6 * * *`)
- Check status: `gh run list --workflow="Deploy to Cloudflare Pages"`
- Live site: https://caskwatch.com
- GitHub Container Registry: `ghcr.io/jeffspahr/bourbontracker`
