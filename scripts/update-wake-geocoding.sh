#!/bin/bash
set -e

echo "🔍 Wake County ABC Store Geocoding Update Script"
echo "=================================================="
echo ""

# Check for required tools
if ! command -v python3 &> /dev/null; then
    echo "❌ Error: python3 is required but not installed"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "❌ Error: jq is required but not installed"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
TEMP_DIR="$ROOT_DIR/.geocoding-temp"

mkdir -p "$TEMP_DIR"

echo "📥 Step 1: Scraping current Wake County addresses..."
python3 "$SCRIPT_DIR/scrape-wake-addresses.py" > "$TEMP_DIR/current-addresses.json"

ADDR_COUNT=$(jq 'length' "$TEMP_DIR/current-addresses.json")
echo "   Found $ADDR_COUNT unique store addresses"
echo ""

echo "🗺️  Step 2: Geocoding addresses..."
python3 "$SCRIPT_DIR/geocode-wake-stores.py" "$TEMP_DIR/current-addresses.json" "$TEMP_DIR/geocoded.json"
echo ""

echo "📝 Step 3: Generating Go code..."
python3 "$SCRIPT_DIR/generate-stores-go.py" "$TEMP_DIR/geocoded.json" "$ROOT_DIR/pkg/nc/wake/stores.go"
echo ""

echo "🔍 Step 4: Checking for changes..."
if git diff --quiet pkg/nc/wake/stores.go; then
    echo "   ✅ No changes - all stores already up to date!"
else
    echo "   📊 Changes detected:"
    git diff --stat pkg/nc/wake/stores.go
    echo ""
    echo "   To see full diff: git diff pkg/nc/wake/stores.go"
    echo "   To commit: git add pkg/nc/wake/stores.go && git commit -m 'chore(wake): update store coordinates'"
fi

# Cleanup
rm -rf "$TEMP_DIR"

echo ""
echo "✅ Done!"
