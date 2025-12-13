# VA ABC Bourbon Tracker

[![CI](https://github.com/jeffspahr/bourbontracker/actions/workflows/main.yml/badge.svg)](https://github.com/jeffspahr/bourbontracker/actions/workflows/main.yml)
[![Release](https://img.shields.io/github/v/release/jeffspahr/bourbontracker)](https://github.com/jeffspahr/bourbontracker/releases)
[![Docker](https://img.shields.io/badge/docker-ghcr.io-blue)](https://github.com/jeffspahr/bourbontracker/pkgs/container/bourbontracker)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jeffspahr/bourbontracker)](https://github.com/jeffspahr/bourbontracker/blob/main/go.mod)
[![License](https://img.shields.io/github/license/jeffspahr/bourbontracker)](https://github.com/jeffspahr/bourbontracker/blob/main/LICENSE)

Track rare bourbon availability across Virginia ABC stores with real-time inventory monitoring and interactive map visualization.

## Features

- 🗺️ **Interactive Google Maps visualization** - See bourbon inventory on a color-coded map
- 📊 **47 tracked products** - Rare and allocated bourbons (Pappy, Blanton's, Buffalo Trace, etc.)
- 🏪 **391 stores** - Complete coverage of Virginia ABC locations
- ⚡ **Real-time data** - Query live inventory via Virginia ABC API
- 🔒 **Secure** - API keys stored in gitignored config files
- 🐳 **Containerized** - Docker image with multi-arch support (amd64/arm64)

## Visualization Options

### Option 1: Google Maps (Simple, Recommended)

Interactive map showing bourbon locations and quantities. Perfect for local use or simple deployments.

**Quick Start:**
```bash
# Copy config template and add your Google Maps API key
cp config.example.js config.js
# Edit config.js and add your API key

# Run tracker to generate inventory data
go build tracker.go
./tracker

# Start local web server
python3 -m http.server 8000

# Open http://localhost:8000/map.html in your browser
```

See [MAP_USAGE.md](MAP_USAGE.md) for detailed setup instructions.

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
## Without building
```go run tracker.go```

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
