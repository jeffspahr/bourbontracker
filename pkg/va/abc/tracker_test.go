package abc

import (
	"io"
	"net/http"
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
