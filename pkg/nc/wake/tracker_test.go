package wake

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type stubTransport func(*http.Request) (*http.Response, error)

func (f stubTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTrackSelection(t *testing.T) {
	for _, tc := range []struct {
		name  string
		set   bool
		codes []string
		want  map[string]int
	}{
		{"default all", false, nil, map[string]int{"A": 1, "B": 1}},
		{"explicit nil", true, nil, map[string]int{}},
		{"explicit empty", true, []string{}, map[string]int{}},
		{"subset", true, []string{"B"}, map[string]int{"B": 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			calls := map[string]int{}
			tr := &Tracker{products: map[string]NCProduct{"A": {}, "B": {}}, client: &http.Client{Transport: stubTransport(func(r *http.Request) (*http.Response, error) {
				if err := r.ParseForm(); err != nil {
					return nil, err
				}
				code := r.Form.Get("productSearch")
				mu.Lock()
				calls[code]++
				mu.Unlock()
				return response(200, noMatchHTML), nil
			})}}
			if tc.set {
				tr.SetProductsToTrack(tc.codes)
			}
			items, err := tr.Track()
			if err != nil || len(items) != 0 {
				t.Fatalf("items=%v err=%v", items, err)
			}
			mu.Lock()
			defer mu.Unlock()
			if !reflect.DeepEqual(calls, tc.want) {
				t.Fatalf("calls=%v want=%v", calls, tc.want)
			}
		})
	}
}

func TestTrackRejectsUnknownSelection(t *testing.T) {
	tr := &Tracker{products: map[string]NCProduct{"A": {}}}
	tr.SetProductsToTrack([]string{"missing"})
	if _, err := tr.Track(); err == nil {
		t.Fatal("unknown selection succeeded")
	}
}

func TestTrackRejectsPartialRefresh(t *testing.T) {
	tr := &Tracker{products: map[string]NCProduct{"00026": {}, "failed": {}}, client: &http.Client{Transport: stubTransport(func(r *http.Request) (*http.Response, error) {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		if r.Form.Get("productSearch") == "failed" {
			return response(503, "unavailable"), nil
		}
		data, err := os.ReadFile("testdata/in-stock.html")
		if err != nil {
			return nil, err
		}
		return response(200, string(data)), nil
	})}}
	items, err := tr.Track()
	if err == nil || items != nil || !strings.Contains(err.Error(), "1 of 2 products failed") {
		t.Fatalf("partial refresh accepted: items=%v err=%v", items, err)
	}
}

func TestParseVerifiedResponses(t *testing.T) {
	for _, tc := range []struct {
		file, code string
		want       int
	}{
		{"in-stock.html", "00026", 16},
		{"out-of-stock.html", "19730", 0},
		{"no-match.html", "17536", 0},
	} {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			items, err := (&Tracker{}).parseSearchResults(tc.code, NCProduct{ListingType: "Listed"}, string(data))
			if err != nil || len(items) != tc.want {
				t.Fatalf("items=%d want=%d err=%v", len(items), tc.want, err)
			}
			for _, item := range items {
				if item.ProductID != tc.code || item.Quantity <= 0 || item.ListingType != "Listed" {
					t.Fatalf("unexpected item: %+v", item)
				}
			}
		})
	}
}

func TestParseRejectsUnknownOrIncompleteResponses(t *testing.T) {
	for name, body := range map[string]string{
		"maintenance":          "<html><body>Temporarily unavailable</body></html>",
		"empty container":      `<div id="productSearchResults"></div>`,
		"unscoped no match":    `<p>Sorry, your search did not return any results.</p>`,
		"missing inventory":    productHTML("00026", ""),
		"wrong product":        productHTML("99999", `<p class="out-of-stock">All Locations Out of Stock</p>`),
		"changed out of stock": productHTML("00026", `<p class="out-of-stock">Service error</p>`),
		"bad quantity":         productHTML("00026", `<div class="inventory-collapse"><ul><li><span class="address">7200 Sandy Fork Rd.</span><span class="quantity">Unavailable</span></li></ul></div>`),
		"missing address":      productHTML("00026", `<div class="inventory-collapse"><ul><li><span class="quantity">1 in stock</span></li></ul></div>`),
		"partial parse":        productHTML("00026", `<div class="inventory-collapse"><ul><li><span class="address">7200 Sandy Fork Rd.</span><span class="quantity">1 in stock</span></li><li>Unavailable</li></ul></div>`),
	} {
		t.Run(name, func(t *testing.T) {
			items, err := (&Tracker{}).parseSearchResults("00026", NCProduct{}, body)
			if err == nil || items != nil {
				t.Fatalf("unsafe response accepted: items=%v err=%v", items, err)
			}
		})
	}
}

const noMatchHTML = `<div id="productSearchResults"><p>Sorry, your search did not return any results.</p></div>`

func productHTML(code, inventory string) string {
	return fmt.Sprintf(`<div id="productSearchResults"><div class="wake-product"><h4>Fixture</h4><small>PLU: %s</small>%s</div></div>`, code, inventory)
}

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}
