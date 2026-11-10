package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jeffspahr/bourbontracker/pkg/nc/wake"
	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

func TestLoadExistingInventoryUsesActiveFilename(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory-nc.json")
	want := []tracker.InventoryItem{{ProductID: "A", State: "NC", County: "Wake", Quantity: 7}}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".previous", data, 0600); err != nil {
		t.Fatal(err)
	}
	if got := loadExistingInventory(path); len(got) != 0 {
		t.Fatalf("unexpected cache source: %v", got)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if got := loadExistingInventory(path); !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestProductFreshnessSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "products.json")
	writeFile(t, path, `[{"nc_code":"A","brand_name":"Fixture","listing_type":"Listed"}]`)
	tr, err := wake.New(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, listing   string
		age             time.Duration
		present, update bool
	}{
		{"fresh listed", "Listed", 2 * time.Hour, true, false},
		{"old listed", "Listed", 25 * time.Hour, true, true},
		{"fresh allocation", "Allocation", 30 * time.Minute, true, false},
		{"old allocation", "Allocation", 2 * time.Hour, true, true},
		{"absent", "Listed", 0, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var existing []tracker.InventoryItem
			if tc.present {
				existing = []tracker.InventoryItem{{ProductID: "A", State: "NC", County: "Wake", ListingType: tc.listing, Timestamp: time.Now().Add(-tc.age)}}
			}
			if got := getProductsNeedingUpdate(tr, existing); (len(got) > 0) != tc.update {
				t.Fatalf("selection=%v expected update=%v", got, tc.update)
			}
		})
	}
}

func TestMergeInventoryUsesSuccessfulSelectionIncludingZeroStock(t *testing.T) {
	old := []tracker.InventoryItem{
		{ProductID: "A", State: "NC", County: "Wake", Quantity: 7},
		{ProductID: "B", State: "NC", County: "Wake", Quantity: 4},
		{ProductID: "C", State: "NC", County: "Wake", Quantity: 9},
		{ProductID: "A", State: "VA", Quantity: 3},
		{ProductID: "A", State: "NC", County: "Other", Quantity: 5},
	}
	newItems := []tracker.InventoryItem{{ProductID: "A", State: "NC", County: "Wake", Quantity: 2}}
	want := []tracker.InventoryItem{old[2], old[3], old[4], newItems[0]}
	if got := mergeInventory(old, newItems, []string{"A", "B"}); !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
	if got := mergeInventory(old, nil, nil); !reflect.DeepEqual(got, old) {
		t.Fatalf("empty selection changed cache: %v", got)
	}
	if got := mergeInventory(old[:1], nil, []string{"A"}); got == nil || len(got) != 0 {
		t.Fatalf("zero stock must clear old row: %v", got)
	}
}

func TestMergeInventoryDoesNotAliasInputs(t *testing.T) {
	existing := []tracker.InventoryItem{{ProductID: "A", State: "NC", County: "Wake", Quantity: 7}}
	incoming := []tracker.InventoryItem{{ProductID: "B", State: "NC", County: "Wake", Quantity: 2}}

	merged := mergeInventory(existing, incoming, []string{"B"})
	if len(merged) != 2 {
		t.Fatalf("merged inventory has %d items, want 2", len(merged))
	}
	merged[0].Quantity = 99
	merged[1].Quantity = 98
	if existing[0].Quantity != 7 || incoming[0].Quantity != 2 {
		t.Fatalf("merged inventory aliases its inputs: existing=%v incoming=%v", existing, incoming)
	}

	if empty := mergeInventory(nil, nil, nil); empty == nil || len(empty) != 0 {
		t.Fatalf("empty merge must return an empty array, got %v", empty)
	}
}

func TestWriteEmptyInventoryAsArray(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.json")
	if err := writeInventory(inventoryOutput{path: path}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "[]" {
		t.Fatalf("empty inventory=%s", data)
	}
}

type stubTransport func(*http.Request) (*http.Response, error)

func (f stubTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestWakeRefreshCLI(t *testing.T) {
	if mode := os.Getenv("TRACKER_TEST_FAILURE_MODE"); mode != "" {
		*enableVA, *enableWake = false, true
		*outputNCFile = os.Getenv("TRACKER_TEST_CACHE")
		*ncProductsFile = os.Getenv("TRACKER_TEST_PRODUCTS")
		http.DefaultTransport = stubTransport(func(r *http.Request) (*http.Response, error) {
			status := 503
			if mode == "unknown-html" {
				status = 200
			}
			body := "<html>Maintenance</html>"
			if mode == "zero-stock" {
				status = 200
				body = `<div id="productSearchResults"><p>Sorry, your search did not return any results.</p></div>`
			}
			return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}, nil
		})
		os.Args = []string{"tracker"}
		main()
		return
	}
	for _, mode := range []string{"http-error", "unknown-html", "zero-stock", "fresh-cache"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			cache, products := filepath.Join(dir, "inventory-nc.json"), filepath.Join(dir, "products.json")
			original := `[{"bt.productId":"00026","bt.state":"NC","bt.county":"Wake","bt.quantity":7,"@timestamp":"2020-01-01T00:00:00Z"}]`
			if mode == "fresh-cache" {
				original = strings.Replace(original, "2020-01-01T00:00:00Z", time.Now().UTC().Format(time.RFC3339), 1)
			}
			writeFile(t, cache, original)
			writeFile(t, products, `[{"nc_code":"00026","brand_name":"Fixture","listing_type":"Listed"}]`)
			cmd := exec.Command(os.Args[0], "-test.run=^TestWakeRefreshCLI$")
			cmd.Env = append(os.Environ(), "TRACKER_TEST_FAILURE_MODE="+mode, "TRACKER_TEST_CACHE="+cache, "TRACKER_TEST_PRODUCTS="+products)
			output, err := cmd.CombinedOutput()
			failed := mode == "http-error" || mode == "unknown-html"
			if failed && (err == nil || !strings.Contains(string(output), "cached inventory unchanged")) {
				t.Fatalf("expected failed CLI refresh: err=%v output=%s", err, output)
			} else if !failed && err != nil {
				t.Fatalf("expected successful CLI refresh: err=%v output=%s", err, output)
			}
			data, err := os.ReadFile(cache)
			if err != nil {
				t.Fatal(err)
			}
			if failed && string(data) != original {
				t.Fatalf("cache overwritten on failure: %s err=%v", data, err)
			}
			if mode == "zero-stock" && string(data) != "[]" {
				t.Fatalf("successful zero stock retained cached rows: %s", data)
			}
			if mode == "fresh-cache" {
				var before, after []tracker.InventoryItem
				if err := json.Unmarshal([]byte(original), &before); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(data, &after); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(before, after) || !strings.Contains(string(output), "No products need updating") {
					t.Fatalf("fresh cache not preserved: data=%s output=%s", data, output)
				}
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestNonexistentVAStoreDoesNotBlockWakeCLI(t *testing.T) {
	if dir := os.Getenv("TRACKER_TEST_MISSING_VA_STORE"); dir != "" {
		*enableVA, *enableWake = true, true
		*storesFile, *productsFile = filepath.Join(dir, "stores"), filepath.Join(dir, "products.json")
		*ncProductsFile = filepath.Join(dir, "nc-products.json")
		*outputVAFile, *outputNCFile = filepath.Join(dir, "inventory-va.json"), filepath.Join(dir, "inventory-nc.json")
		http.DefaultTransport = stubTransport(func(r *http.Request) (*http.Response, error) {
			status, body := 200, `<div id="productSearchResults"><p>Sorry, your search did not return any results.</p></div>`
			if r.URL.Host == "www.abc.virginia.gov" {
				if r.URL.Query().Get("storeNumbers") == "325" {
					status, body = 400, "No Store exists for store number &#39;325&#39;"
				} else {
					body = `{"products":[{"productId":"P","storeInfo":{"quantity":7,"storeId":32}}]}`
				}
			}
			return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}, nil
		})
		os.Args = []string{"tracker"}
		main()
		return
	}

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "stores"), "325\n32\n")
	writeFile(t, filepath.Join(dir, "products.json"), `{"P":"Fixture"}`)
	writeFile(t, filepath.Join(dir, "nc-products.json"), `[{"nc_code":"00026","brand_name":"Fixture","listing_type":"Listed"}]`)
	writeFile(t, filepath.Join(dir, "inventory-va.json"), `[{"bt.productId":"P","bt.state":"VA","bt.storeId":"325","bt.quantity":5}]`)
	writeFile(t, filepath.Join(dir, "inventory-nc.json"), `[{"bt.productId":"00026","bt.state":"NC","bt.county":"Wake","bt.quantity":7,"@timestamp":"2020-01-01T00:00:00Z"}]`)
	cmd := exec.Command(os.Args[0], "-test.run=^TestNonexistentVAStoreDoesNotBlockWakeCLI$")
	cmd.Env = append(os.Environ(), "TRACKER_TEST_MISSING_VA_STORE="+dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("refresh failed: err=%v output=%s", err, output)
	}
	vaData, err := os.ReadFile(filepath.Join(dir, "inventory-va.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vaItems []tracker.InventoryItem
	if err := json.Unmarshal(vaData, &vaItems); err != nil {
		t.Fatal(err)
	}
	if len(vaItems) != 1 || vaItems[0].StoreID != "32" || vaItems[0].Quantity != 7 {
		t.Fatalf("VA snapshot retained a missing store or lost active inventory: %s", vaData)
	}
	ncData, err := os.ReadFile(filepath.Join(dir, "inventory-nc.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(ncData) != "[]" {
		t.Fatalf("Wake refresh did not run and clear zero stock: %s output=%s", ncData, output)
	}
}
