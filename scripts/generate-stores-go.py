#!/usr/bin/env python3
"""
Generate pkg/nc/wake/stores.go from geocoded addresses
"""
import json
import sys

HEADER = '''package wake

import "github.com/jeffspahr/bourbontracker/pkg/tracker"

// Store coordinates for Wake County ABC stores
var storeCoordinates = map[string]tracker.Location{
'''

FOOTER = '''}

// getStoreLocation returns the coordinates for a store address
func getStoreLocation(address string) (tracker.Location, bool) {
	if loc, ok := storeCoordinates[address]; ok {
		return loc, true
	}
	// Return zero coordinates if not found
	return tracker.Location{Latitude: 0, Longitude: 0}, false
}
'''

def main():
    if len(sys.argv) != 3:
        print("Usage: generate-stores-go.py <input_json> <output_go>")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]

    # Load geocoded data
    with open(input_file, 'r') as f:
        stores = json.load(f)

    # Generate Go code
    lines = [HEADER]

    # Sort addresses for consistent output
    for address in sorted(stores.keys()):
        coords = stores[address]
        lat = coords['lat']
        lon = coords['lon']

        # Escape quotes in address
        escaped_addr = address.replace('"', '\\"')

        line = f'\t"{escaped_addr}": {{Latitude: {lat}, Longitude: {lon}}},\n'
        lines.append(line)

    lines.append(FOOTER)

    # Write output
    with open(output_file, 'w') as f:
        f.writelines(lines)

    print(f"   ✅ Generated {output_file} with {len(stores)} stores")

if __name__ == '__main__':
    main()
