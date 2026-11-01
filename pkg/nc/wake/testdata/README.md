These result-container excerpts were captured with five read-only POST searches
at https://wakeabc.com/search-results on 2026-11-01. Page headers, scripts, and
footers are omitted; no full inventory scan was performed.

- `in-stock.html`: productSearch=00026, HTTP 200, positive store quantities.
- `out-of-stock.html`: productSearch=19730, HTTP 200, explicit
  `All Locations Out of Stock` marker and an empty inventory list.
- `no-match.html`: productSearch=17536, HTTP 200, explicit no-match paragraph.
  The deliberately unmatched query 999999999 produced the same result shape.
- A fifth query, 27090, returned positive inventory and is not needed as a fixture.

Unknown HTML, missing containers, incomplete product rows, and transport errors
must remain errors rather than being interpreted as zero stock.
