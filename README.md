# VA ABC Bourbon Tracker

[![CI](https://github.com/jeffspahr/bourbontracker/actions/workflows/main.yml/badge.svg)](https://github.com/jeffspahr/bourbontracker/actions/workflows/main.yml)
[![Release](https://img.shields.io/github/v/release/jeffspahr/bourbontracker)](https://github.com/jeffspahr/bourbontracker/releases)
[![Docker](https://img.shields.io/badge/docker-ghcr.io-blue)](https://github.com/jeffspahr/bourbontracker/pkgs/container/bourbontracker)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jeffspahr/bourbontracker)](https://github.com/jeffspahr/bourbontracker/blob/main/go.mod)
[![License](https://img.shields.io/github/license/jeffspahr/bourbontracker)](https://github.com/jeffspahr/bourbontracker/blob/main/LICENSE)

This is intended to run in Kubernetes, but it can be modified to fit your environment.  The app will output inventory data in JSON format which will be picked up by Filebeat and shipped to an Elasticsearch cluster.  I plan to include the end to end manifests to deploy this in any Kubnernetes cluster.

The end goal is to be able to visualize and alert on geolocation inventory data.

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
docker run ghcr.io/jeffspahr/bourbontracker

# Or specify a version
docker pull ghcr.io/jeffspahr/bourbontracker:1.0.1
docker run ghcr.io/jeffspahr/bourbontracker:1.0.1
```
# Architecture in Kubernetes
 <img src="bourbontracker.png">
