package abc

import (
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

type stubTransport func(*http.Request) (*http.Response, error)

func (f stubTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestExhaustedStoreRetriesRejectPartialInventory(t *testing.T) {
	for _, status := range []int{400, 429, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := map[string]int{}
			stubHTTP(t, func(r *http.Request) (*http.Response, error) {
				store := r.URL.Query().Get("storeNumbers")
				calls[store]++
				if store == "2" {
					return response(status, "unavailable"), nil
				}
				return response(200, `{"products":[{"productId":"P","storeInfo":{"quantity":7,"storeId":1}}]}`), nil
			})
			tr := testTracker([]string{"1", "2", "3"})
			items, err := tr.Track()
			if err == nil || items != nil || calls["2"] != 5 || calls["3"] != 0 {
				t.Fatalf("incomplete inventory accepted: items=%v err=%v calls=%v", items, err, calls)
			}
		})
	}
}

func TestStoreRecoversWithinRetryBudget(t *testing.T) {
	calls := 0
	stubHTTP(t, func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return response(503, "unavailable"), nil
		}
		return response(200, `{"products":[]}`), nil
	})
	items, err := testTracker([]string{"1"}).Track()
	if err != nil || len(items) != 0 || calls != 2 {
		t.Fatalf("items=%v err=%v calls=%d", items, err, calls)
	}
}

func TestNonexistentStoreDoesNotBlockActiveStores(t *testing.T) {
	body, err := os.ReadFile("testdata/nonexistent-store.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, stores := range [][]string{{"325", "1", "3"}, {"1", "325", "3"}, {"1", "3", "325"}} {
		t.Run(strings.Join(stores, ","), func(t *testing.T) {
			calls := map[string]int{}
			stubHTTP(t, func(r *http.Request) (*http.Response, error) {
				store := r.URL.Query().Get("storeNumbers")
				calls[store]++
				if store == "325" {
					return response(400, string(body)), nil
				}
				return response(200, `{"products":[{"productId":"P","storeInfo":{"quantity":7}}]}`), nil
			})
			items, err := testTracker(stores).Track()
			if err != nil || len(items) != 2 || !reflect.DeepEqual(calls, map[string]int{"1": 1, "325": 1, "3": 1}) {
				t.Fatalf("items=%v err=%v calls=%v", items, err, calls)
			}
			for _, item := range items {
				if item.StoreID == "325" {
					t.Fatal("nonexistent store produced inventory")
				}
			}
		})
	}
}

func TestNonexistentStoreResponseMustBeExact(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		want       bool
	}{
		{"verified response", "No Store exists for store number &#39;325&#39;", 400, true},
		{"plain text", "No Store exists for store number '325'", 400, true},
		{"whitespace", "\nNo Store exists for store number &#39;325&#39;\r\n", 400, true},
		{"different store", "No Store exists for store number &#39;32&#39;", 400, false},
		{"missing store", "No Store exists", 400, false},
		{"unknown 400", "Invalid product code", 400, false},
		{"empty 400", "", 400, false},
		{"HTML page", "<html>No Store exists for store number &#39;325&#39;</html>", 400, false},
		{"extra message", "No Store exists for store number &#39;325&#39;: service unavailable", 400, false},
		{"server failure", "No Store exists for store number &#39;325&#39;", 503, false},
		{"rate limited", "No Store exists for store number &#39;325&#39;", 429, false},
		{"not found", "No Store exists for store number &#39;325&#39;", 404, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isNonexistentStore(tc.status, []byte(tc.body), "325"); got != tc.want {
				t.Fatalf("recognized=%v want=%v", got, tc.want)
			}
			if !tc.want {
				stubHTTP(t, func(r *http.Request) (*http.Response, error) { return response(tc.status, tc.body), nil })
				if items, err := testTracker([]string{"325"}).Track(); err == nil || items != nil {
					t.Fatalf("unverified store response accepted: items=%v err=%v", items, err)
				}
			}
		})
	}
}

func TestAllStoresMissingRejectsEmptySnapshot(t *testing.T) {
	stubHTTP(t, func(r *http.Request) (*http.Response, error) {
		return response(400, "No Store exists for store number '"+r.URL.Query().Get("storeNumbers")+"'"), nil
	})
	items, err := testTracker([]string{"1", "2"}).Track()
	if err == nil || items != nil || !strings.Contains(err.Error(), "no stores returned inventory") {
		t.Fatalf("unverified empty inventory accepted: items=%v err=%v", items, err)
	}
}

func TestNonexistentStoreAndConfirmedZeroStockCanSucceed(t *testing.T) {
	stubHTTP(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("storeNumbers") == "325" {
			return response(400, "No Store exists for store number &#39;325&#39;"), nil
		}
		return response(200, `{"products":[]}`), nil
	})
	items, err := testTracker([]string{"325", "32"}).Track()
	if err != nil || len(items) != 0 {
		t.Fatalf("confirmed zero stock rejected: items=%v err=%v", items, err)
	}
}

func TestRejectsUnrecognizedSuccessBody(t *testing.T) {
	for _, body := range []string{`{}`, `{"products":null}`, `<html>Maintenance</html>`} {
		t.Run(body, func(t *testing.T) {
			stubHTTP(t, func(r *http.Request) (*http.Response, error) { return response(200, body), nil })
			if _, err := testTracker([]string{"1"}).Track(); err == nil {
				t.Fatal("invalid response accepted")
			}
		})
	}
}

func stubHTTP(t *testing.T, f stubTransport) {
	t.Helper()
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	http.DefaultTransport = f
}

func testTracker(stores []string) *Tracker {
	// Zero delay/backoff makes the real retry loop fast in offline tests.
	return &Tracker{config: tracker.Config{MaxRetries: 5}, stores: stores, products: map[string]string{"P": "Fixture"}, storeRetries: map[int]int{}}
}

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}
}
