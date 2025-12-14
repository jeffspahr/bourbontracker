package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// NCProduct represents a product from NC ABC warehouse
type NCProduct struct {
	NCCode      string `json:"nc_code"`
	BrandName   string `json:"brand_name"`
	ListingType string `json:"listing_type"`
	Size        string `json:"size"`
	Available   int    `json:"total_available"`
	Supplier    string `json:"supplier,omitempty"`
}

var (
	outputFile = flag.String("output", "nc-products.json", "Output JSON file")
	listingType = flag.String("listing", "", "Filter by listing type (Limited, Allocation, Listed, Barrel, Christmas)")
	minCases   = flag.Int("min-cases", 0, "Minimum cases available to include")
)

func main() {
	flag.Parse()

	fmt.Fprintf(os.Stderr, "Fetching NC ABC warehouse stock data...\n")

	resp, err := http.Get("https://abc2.nc.gov/StoresBoards/Stocks")
	if err != nil {
		log.Fatalf("Failed to fetch stock page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	products, err := parseStockPage(string(body))
	if err != nil {
		log.Fatalf("Failed to parse stock page: %v", err)
	}

	// Apply filters
	filtered := filterProducts(products)

	fmt.Fprintf(os.Stderr, "Found %d products (filtered from %d total)\n", len(filtered), len(products))

	// Write to JSON
	data, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(*outputFile, data, 0644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Wrote %d products to %s\n", len(filtered), *outputFile)

	// Print summary by listing type
	typeCount := make(map[string]int)
	for _, p := range filtered {
		typeCount[p.ListingType]++
	}

	fmt.Fprintf(os.Stderr, "\nBreakdown by Listing Type:\n")
	for listType, count := range typeCount {
		fmt.Fprintf(os.Stderr, "  %s: %d\n", listType, count)
	}
}

func parseStockPage(html string) ([]NCProduct, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var products []NCProduct

	// Find the stock table and parse rows
	doc.Find("table tbody tr").Each(func(i int, row *goquery.Selection) {
		cols := row.Find("td")
		if cols.Length() < 7 {
			return // Skip rows without enough columns
		}

		// Extract data from columns
		ncCode := strings.TrimSpace(cols.Eq(0).Text())
		brandName := strings.TrimSpace(cols.Eq(1).Text())
		listingType := strings.TrimSpace(cols.Eq(2).Text())
		availableText := strings.TrimSpace(cols.Eq(3).Text())
		size := strings.TrimSpace(cols.Eq(4).Text())
		supplier := strings.TrimSpace(cols.Eq(6).Text())

		// Parse available quantity
		available := 0
		if availableText != "" {
			available, _ = strconv.Atoi(strings.ReplaceAll(availableText, ",", ""))
		}

		// Skip empty rows
		if ncCode == "" || brandName == "" {
			return
		}

		product := NCProduct{
			NCCode:      ncCode,
			BrandName:   brandName,
			ListingType: listingType,
			Size:        size,
			Available:   available,
			Supplier:    supplier,
		}

		products = append(products, product)
	})

	return products, nil
}

func filterProducts(products []NCProduct) []NCProduct {
	var filtered []NCProduct

	for _, p := range products {
		// Filter by listing type if specified
		if *listingType != "" && !strings.EqualFold(p.ListingType, *listingType) {
			continue
		}

		// Filter by minimum cases
		if p.Available < *minCases {
			continue
		}

		filtered = append(filtered, p)
	}

	return filtered
}
