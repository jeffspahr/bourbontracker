#!/usr/bin/env python3
"""
Geocode Wake County ABC store addresses using Nominatim
"""
import json
import time
import sys
import urllib.parse
import urllib.request

def geocode_address(address):
    """Geocode an address using Nominatim API"""
    base_url = "https://nominatim.openstreetmap.org/search"
    params = {
        'q': address,
        'format': 'json',
        'limit': 1
    }

    url = f"{base_url}?{urllib.parse.urlencode(params)}"
    headers = {
        'User-Agent': 'BourbonTracker/1.0 (geocoding update script)'
    }

    try:
        req = urllib.request.Request(url, headers=headers)
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode())
            if data:
                return {
                    'lat': float(data[0]['lat']),
                    'lon': float(data[0]['lon']),
                    'display_name': data[0]['display_name']
                }
    except Exception as e:
        print(f"  ⚠️  Error geocoding {address}: {e}", file=sys.stderr)

    return None

def load_existing_coords(filename):
    """Load existing coordinates from stores.go if it exists"""
    coords = {}
    try:
        with open(filename, 'r') as f:
            content = f.read()
            # Simple parsing - look for lines like: "address": {Latitude: 35.123, Longitude: -78.456},
            import re
            pattern = r'"([^"]+)":\s*\{Latitude:\s*([\d.]+),\s*Longitude:\s*([-\d.]+)\}'
            for match in re.finditer(pattern, content):
                address, lat, lon = match.groups()
                coords[address] = {
                    'lat': float(lat),
                    'lon': float(lon),
                    'display_name': 'Existing'
                }
    except FileNotFoundError:
        pass
    return coords

def main():
    if len(sys.argv) != 3:
        print("Usage: geocode-wake-stores.py <input_json> <output_json>")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]

    # Load addresses to geocode
    with open(input_file, 'r') as f:
        addresses = json.load(f)

    # Load existing coordinates to avoid re-geocoding
    existing = load_existing_coords('pkg/nc/wake/stores.go')

    results = {}
    new_count = 0
    cached_count = 0

    print(f"   Processing {len(addresses)} addresses...")

    for i, address in enumerate(addresses, 1):
        # Check if we already have coordinates for this address
        if address in existing:
            results[address] = existing[address]
            cached_count += 1
            print(f"   [{i}/{len(addresses)}] {address[:50]:50s} → Cached", file=sys.stderr)
        else:
            # Geocode new address
            print(f"   [{i}/{len(addresses)}] {address[:50]:50s} → Geocoding...", file=sys.stderr)
            result = geocode_address(address)
            if result:
                results[address] = result
                new_count += 1
                print(f"      ✓ {result['lat']}, {result['lon']}", file=sys.stderr)
            else:
                print(f"      ✗ Failed", file=sys.stderr)

            # Rate limiting: 1 second between requests
            if i < len(addresses):
                time.sleep(1)

    # Save results
    with open(output_file, 'w') as f:
        json.dump(results, f, indent=2)

    print(f"   ✅ Geocoded: {new_count} new, {cached_count} cached, {len(results)} total", file=sys.stderr)

if __name__ == '__main__':
    main()
