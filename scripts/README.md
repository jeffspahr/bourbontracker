# Wake County Geocoding Scripts

Scripts for updating Wake County ABC store coordinates.

## Quick Start

When Wake County adds new stores or changes addresses:

```bash
./scripts/update-wake-geocoding.sh
```

This will:
1. Scrape current addresses from wakeabc.com
2. Geocode any new addresses (using cache for existing)
3. Update `pkg/nc/wake/stores.go`
4. Show you what changed

## What Each Script Does

### `update-wake-geocoding.sh`
Main orchestrator script. Run this one.

### `scrape-wake-addresses.py`
Scrapes store addresses from Wake County ABC website by searching for multiple products.

### `geocode-wake-stores.py`
Geocodes addresses using OpenStreetMap Nominatim API. Caches existing coordinates to avoid re-geocoding.

### `generate-stores-go.py`
Generates the Go code for `pkg/nc/wake/stores.go` from geocoded data.

## When to Update

The tracker will warn you when it encounters an unknown address:

```
WARNING: No coordinates found for address: 123 New St. Raleigh, NC 27601
Run ./scripts/update-wake-geocoding.sh to update store coordinates
```

## Requirements

- Python 3
- jq (for JSON processing)

## Geocoding Notes

- Uses OpenStreetMap Nominatim API (free, no API key needed)
- Rate limit: 1 second between requests
- Caches existing coordinates to speed up updates
- Manual fallback for addresses that fail to geocode
