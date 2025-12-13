package wake

import "github.com/jeffspahr/bourbontracker/pkg/tracker"

// Store coordinates for Wake County ABC stores
var storeCoordinates = map[string]tracker.Location{
	"7200 Sandy Fork Rd. Raleigh, NC 27609":              {Latitude: 35.8719206, Longitude: -78.6232906},
	"3320 Olympia Dr. Raleigh, NC 27603":                 {Latitude: 35.7345287, Longitude: -78.6517657},
	"1793 West Williams St. Apex, NC 27502":              {Latitude: 35.7593587, Longitude: -78.8765183},
	"1940 Cinema Dr. Fuquay Varina, NC 27526":            {Latitude: 35.6018657, Longitude: -78.7531983},
	"11360 Capital Blvd. Wake Forest, NC 27587":          {Latitude: 35.9541197, Longitude: -78.5395367},
	"6301 Town Center Dr. Raleigh, NC 27614":             {Latitude: 35.870012, Longitude: -78.5776628},
	"200 New Rand Road Garner, NC 27529":                 {Latitude: 35.7017438, Longitude: -78.60232},
	"1601-61 Cross Link Rd. Raleigh, NC 27610":           {Latitude: 35.7461643, Longitude: -78.6230024},
	"7336 Creedmoor Rd. Raleigh, NC 27613":               {Latitude: 35.8867296, Longitude: -78.6793864},
	"1415 E. Williams St. Apex, NC 27617":                {Latitude: 35.7134437, Longitude: -78.8395803},
	"7911 ACC Blvd. Raleigh, NC 27617":                   {Latitude: 35.9169256, Longitude: -78.7800074},
	"100 Village Walk Dr Holly Springs, NC 27540":        {Latitude: 35.6389072, Longitude: -78.8353853},
	"1505 Banyon Pl. Wendell, NC 27571":                  {Latitude: 35.7794, Longitude: -78.3687},
	"4009 Davis Dr. Morrisville, NC 27560":               {Latitude: 35.8349365, Longitude: -78.8556205},
	"3615 SW Cary Parkway Cary, NC 27513":                {Latitude: 35.781112, Longitude: -78.8384748},
	"665 Cary Towne Blvd. Cary, NC 27511":                {Latitude: 35.7766327, Longitude: -78.766272},
	"6494 Tryon Rd. Cary, NC 27511":                      {Latitude: 35.7436372, Longitude: -78.7623834},
	"704 Money Ct. Knightdale, NC 27545":                 {Latitude: 35.7982975, Longitude: -78.4731725},
	"4501 Vineyrd Pine Ln. Rolesville, NC 27571":         {Latitude: 35.9319, Longitude: -78.4466},
	"209 S Salisbury St Raleigh, NC 27601":               {Latitude: 35.7780803, Longitude: -78.6400923},
	"4215 The Circle at North Hills Rd Raleigh, NC 27609": {Latitude: 35.8378959, Longitude: -78.642488},
	"2109-106 Avent Ferry Rd. Raleigh, NC 27606":         {Latitude: 35.7796029, Longitude: -78.6756501},
	"1420 N Ardendell Dr. Zebulon, NC 27597":             {Latitude: 35.840154, Longitude: -78.3252935},
	"6809 Davis Circle Raleigh, NC 27612":                {Latitude: 35.862667, Longitude: -78.7099899},
	"420 Woodburn Rd. Raleigh, NC 27605":                 {Latitude: 35.7899326, Longitude: -78.6586902},
	"2645 Appliance Ct. Raleigh, NC 27604":               {Latitude: 35.8125293, Longitude: -78.6018258},
}

// getStoreLocation returns the coordinates for a store address
func getStoreLocation(address string) tracker.Location {
	if loc, ok := storeCoordinates[address]; ok {
		return loc
	}
	// Return zero coordinates if not found
	return tracker.Location{Latitude: 0, Longitude: 0}
}
