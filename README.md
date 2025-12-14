# Cask Watch

<p align="center">
  <img src="logo.svg" alt="Cask Watch Logo" width="400"/>
</p>

[![CI](https://github.com/jeffspahr/bourbontracker/actions/workflows/main.yml/badge.svg)](https://github.com/jeffspahr/bourbontracker/actions/workflows/main.yml)
[![Release](https://img.shields.io/github/v/release/jeffspahr/bourbontracker)](https://github.com/jeffspahr/bourbontracker/releases)
[![Docker](https://img.shields.io/badge/docker-ghcr.io-blue)](https://github.com/jeffspahr/bourbontracker/pkgs/container/bourbontracker)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jeffspahr/bourbontracker)](https://github.com/jeffspahr/bourbontracker/blob/main/go.mod)
[![License](https://img.shields.io/github/license/jeffspahr/bourbontracker)](https://github.com/jeffspahr/bourbontracker/blob/main/LICENSE)

Track rare bourbon and spirits availability across multiple states' ABC stores with real-time inventory monitoring, smart caching, and interactive map visualization.

> **v2.0.0** - Now tracking **48,850+ items** across Virginia ABC and North Carolina with intelligent caching and listing type filters!

## Features

- 🗺️ **Interactive Google Maps visualization** - See spirits inventory on a color-coded map with geocoded locations
- 🌎 **Multi-state support** - Virginia ABC (390 stores) + North Carolina Wake County (15 stores)
- 🧠 **Smart caching** - Intelligent request optimization reduces API calls by 80% on scheduled runs
- 🏷️ **Listing type filters** - Filter NC products by Limited, Allocation, Listed, Barrel, Christmas
- 📊 **Comprehensive tracking** - 48,850+ items including rare allocations (Pappy, Blanton's, Buffalo Trace, etc.)
- ⚡ **Optimized performance** - Conservative rate limiting prevents 429 errors
- 🔒 **Secure** - API keys stored in gitignored config files
- 🐳 **Containerized** - Docker image with multi-arch support (amd64/arm64)
- 🔧 **Modular architecture** - Easy to add new states/counties

## Visualization Options

### Option 1: Google Maps (Simple, Recommended)

Interactive map showing spirits locations and quantities. Perfect for local use or simple deployments.

**Quick Start:**
```bash
# Copy config template and add your Google Maps API key
cp config.example.js config.js
# Edit config.js and add your API key

# Build the tracker
go build -o tracker ./cmd/tracker

# Run tracker to generate inventory data (Virginia ABC only)
./tracker

# Run with Wake County NC included
./tracker -va -wake

# Start local web server
python3 -m http.server 8000

# Open http://localhost:8000 in your browser
```

See [MAP_USAGE.md](MAP_USAGE.md) for detailed setup instructions.

**Screenshots:**

<p align="center">
  <img src="screenshots/map-overview.png" alt="Map Overview" width="800"/>
  <br/>
  <em>Interactive Google Maps view showing spirits inventory across Virginia ABC stores</em>
</p>

<p align="center">
  <img src="screenshots/product-filter.png" alt="Product Filter" width="800"/>
  <br/>
  <em>Product filter dropdown with multi-select capability</em>
</p>

<p align="center">
  <img src="screenshots/store-detail.png" alt="Store Detail" width="600"/>
  <br/>
  <em>Store marker showing product details and quantities</em>
</p>

### Option 2: Elasticsearch + Kibana (Advanced)

Full ELK stack deployment for historical data, time-series analysis, and alerting. Ideal for production Kubernetes environments.

The app outputs inventory data in JSON format which is picked up by Filebeat and shipped to an Elasticsearch cluster. End-to-end Kubernetes manifests are included in the `k8s/` directory.

See Architecture section below for details.

---

Inspired by https://github.com/misfitlabs/pappytracker

# Convert Product List from a Python Dictionary to JSON
```python3 dict2json.py |jq > products.json```
Tested on python3.

# Generate List of Stores
```go run generateStoreList.go```

# Run the Tracker

## Command Line Options
```bash
# Virginia ABC only (default)
./tracker

# Virginia ABC + Wake County NC
./tracker -va -wake

# Wake County NC only
./tracker -va=false -wake

# Custom output file
./tracker -output my-inventory.json

# Custom product list (for VA ABC)
./tracker -products my-products.json

# Custom store list (for VA ABC)
./tracker -stores my-stores
```

## Supported Regions

### Virginia ABC (`-va`)
- **Stores**: 390 across Virginia
- **Method**: REST API at `abc.virginia.gov`
- **Product IDs**: Numeric codes (e.g., `018006` for Buffalo Trace)
- **Products Tracked**: ~48 curated rare/allocated spirits
- **Coordinates**: Yes (latitude/longitude for each store)

### Wake County NC (`-wake`)
- **Stores**: 15 across Wake County
- **Method**: HTML parsing via web scraping at `wakeabc.com`
- **Product Search**: NC Codes from state warehouse (e.g., `18006` for Buffalo Trace)
- **Products Tracked**: All 3,167 products from NC ABC warehouse catalog
- **Listing Types**: Limited, Allocation, Listed, Barrel, Christmas
- **Coordinates**: Yes (geocoded latitude/longitude for all 15 stores)
- **Smart Caching**:
  - "Listed" products: Update every 24 hours
  - Limited/Allocation/Barrel/Christmas: Update hourly
  - Result: 80% reduction in API requests on scheduled runs

### Performance Stats
- **Total Items**: 48,850+ tracked across both states
- **Fresh Deployment**: ~36 minutes (full scan of all products)
- **Scheduled Runs**: ~3-5 minutes (with smart caching)
- **Rate Limiting**: Conservative (3 concurrent, 1 second delay) to prevent 429 errors

See [ARCHITECTURE.md](ARCHITECTURE.md) for details on adding new states/counties.

## Run using Docker
```bash
# Pull the latest version
docker pull ghcr.io/jeffspahr/bourbontracker:latest

# Run and save inventory.json to current directory
docker run --rm -v $(pwd):/root ghcr.io/jeffspahr/bourbontracker

# Or specify a version
docker pull ghcr.io/jeffspahr/bourbontracker:1.0.1
docker run --rm -v $(pwd):/root ghcr.io/jeffspahr/bourbontracker:1.0.1
```

**Note:** The `-v $(pwd):/root` flag mounts the current directory so `inventory.json` is written to your host machine for use with the Google Maps visualization.
# Architecture in Kubernetes
 <img src="bourbontracker.png">
