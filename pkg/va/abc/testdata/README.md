# Virginia ABC response provenance

`nonexistent-store.txt` is the complete response body from a read-only GET on
2026-11-05 to:

https://www.abc.virginia.gov/webapi/inventory/mystore?storeNumbers=325&productCodes=018006

The request used `Accept: application/json` and returned HTTP 400. The official
store page at https://www.abc.virginia.gov/stores/325 also returned HTTP 404.
A control request with store 32 and the same product returned HTTP 200 and a
valid products array. Store 325 is removed from the checked-in `stores` list.

Only the exact HTTP 400 response naming the requested store is recognized as a
nonexistent store. Other status codes, changed text, HTML pages, and responses
naming another store must still fail the refresh. A run where no store returns
valid inventory must also fail, rather than publishing an empty snapshot.
