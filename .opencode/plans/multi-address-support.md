# Implementation Plan: Multi-Address (Multiple Endpoint) Support

Target repo: `/projects/elasticsearch` (`github.com/disaster37/elasticsearch/v9`).
Feature: replace the single-endpoint `Config.URL` with `Config.Addresses []string` (multiple endpoints).

---

## 1. Summary

Today the client supports exactly one endpoint via `Config.URL`, which `New` passes to resty
through `c.SetBaseURL(cfg.URL)` (a single base URL). This plan **removes `Config.URL` entirely** and
replaces it with `Config.Addresses []string` (a list of node URLs). Requests are distributed across
the addresses in **round-robin** order, with **failover on retry** — matching the official
`github.com/elastic/go-elasticsearch` client's connection-pool behavior, but implemented idiomatically
for this resty-based client. A single endpoint is expressed as a one-element slice
(`Addresses: []string{"https://node:9200"}`).

**How the official client does it (verified against `go-elasticsearch` / `elastic-transport-go` source):**

- `elasticsearch.Config.Addresses []string` — a list of node URLs.
- Each address is parsed with `url.Parse(strings.TrimRight(addr, "/"))` (trailing slashes trimmed);
  a `*url.URL` becomes one `Connection`.
- A connection pool holds one `Connection` per address. The default `roundRobinSelector` returns
  the next connection via `atomic.AddUint64(&curr, 1) % uint64(len(conns))` (counter seeded to
  `math.MaxUint64` so the **first** pick is index 0).
- `Perform` loops `for i := 0; i <= maxRetries; i++`: each attempt calls `pool.Next()`, rewrites the
  request's scheme/host to the selected address (and prepends any address path), executes, and on a
  network error or a retryable status calls `pool.OnFailure(conn)` and tries the next connection.
- Address paths are allowed and preserved (prepended to the request path).

This codebase is **not** the official client: it uses resty and already has its own retry mechanism
(`Config.RetryCount` + `RetryConditions`). We adapt the ideas (address list, per-request round-robin,
failover to the next node) without importing the official client and without changing `api/`, `trace/`,
or the shared `*http.Transport`.

---

## 2. Locked decisions

| # | Decision |
|---|----------|
| 1 | **Config shape:** `Config.URL` is **removed entirely**. `Config.Addresses []string` is the only address field; a single endpoint is `Addresses: []string{"https://node:9200"}`. |
| 2 | **Empty list allowed:** `nil` and `[]string{}` are equivalent → no base URL, each request must supply an absolute URL (matches today's empty-config behavior, now without a `URL` field). |
| 3 | **Single element:** `Addresses` with one element behaves identically to the old single-URL path (round-robin hook is a no-op because `len <= 1`; `SetBaseURL(first)` resolves the base URL). |
| 4 | **Failover is retry-driven:** round-robin advances on every request *attempt*; a failed request is retried against the **next** address only when `RetryCount > 0` (existing `DefaultRetryConditions`). With `RetryCount == 0` requests still round-robin for load balancing but do **not** fail over. |
| 5 | **Implementation hook:** a resty `OnBeforeRequest` middleware rewrites the relative request path into the next address's absolute URL. Verified against resty v2.16.5 source: user middleware runs *before* resty's `parseRequestURL` (base-URL resolution), and resty's retry loop resets `req.URL` to the original relative path on every attempt, so the hook re-runs and advances the address. |
| 6 | **No default port:** addresses must include a port (or rely on net/http's 80/443 defaults), matching the previous single-address behavior. (The current official client also does not inject a port.) |
| 7 | **No env-var support** (`ELASTICSEARCH_URL`) in the client library — out of scope; this codebase has no client-side env-var handling today. (The integration test helper reads its own `ELASTICSEARCH_URL` env var — see §12.6 — that is unrelated to `Config`.) |
| 8 | **Validation convention:** `New` returns `*types.ElasticsearchError` (status 400) for invalid addresses — it never panics. Matches the existing validation style. |

---

## 3. Design overview (resty hook mechanics)

Resty v2.16.5 request flow (verified in source):

1. `Request.Execute(method, path)` sets `r.URL = path` (the relative path, e.g. `"/_search"`).
2. `Client.execute` → `executeBefore` runs, in order:
   a. **user-defined `OnBeforeRequest` middleware** (`udBeforeRequest`) — *our hook lives here*,
   b. default middleware `parseRequestURL` (applies `BaseURL`, adds query params),
   c. `parseRequestHeader`, `parseRequestBody`, `createHTTPRequest`, `addCredentials`, `createCurlCmd`.
3. With `RetryCount > 0`, `Request.Execute` runs a `Backoff` loop; **each attempt** re-assigns
   `r.URL = <original relative path>` then calls `execute` again, re-running our hook.

Consequences:

- Our hook runs **before** `parseRequestURL`, so setting `req.URL` to an absolute URL makes resty skip
  the base-URL prepend (resty only prepends `BaseURL` when `req.URL` is relative).
- On every retry attempt the hook re-runs, so the round-robin counter advances and a failed request
  lands on the **next** address.
- For **0 addresses** the hook is a no-op (no base URL; per-request absolute URLs). For **1 address**
  the hook is a no-op and `SetBaseURL(first)` resolves the base URL exactly as the old single-URL path
  did.
- Because `req.URL` is absolute by the time the logging and OpenTelemetry `OnBeforeRequest` hooks run,
  they keep working and now log/record the actual node (URL redaction via `RedactURL` is unchanged).

The shared `*http.Transport` (cloned in `New`) is **kept as-is**, so
`RestyClient().Transport()` continues to return `*http.Transport` and every existing transport-level
test continues to pass. net/http already maintains per-host idle-connection pools, so no per-address
transport is required.

---

## 4. Data structures (exact Go)

### 4.1 `Config` (modified — the address field is replaced)

```go
// Config holds configuration for the Elasticsearch client.
type Config struct {
	// Addresses is a list of Elasticsearch node URLs to connect to. Requests
	// are distributed across the addresses in round-robin order. When
	// RetryCount > 0, a request that fails against one address is retried
	// against the next address (see DefaultRetryConditions).
	//
	// Example:
	//   Addresses: []string{"https://node1:9200", "https://node2:9200"}
	//
	// An empty (nil or zero-length) slice is allowed: in that case the client
	// has no base URL and every request must supply an absolute URL.
	Addresses []string

	// ... all other fields (Username, Password, APIKey, BearerToken,
	// TLSSkipVerify, AllowInsecureHTTP, CACert, Timeout, IdleConnTimeout,
	// DisableHTTP2, RetryCount, RetryWaitTime, RetryMaxWaitTime,
	// RetryConditions) are unchanged.
}
```

### 4.2 New `connectionPool` type (new file `connectionpool.go`)

```go
package elasticsearch

import (
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
)

// connectionPool holds the parsed node addresses and selects the next one in
// round-robin order for each request. It is safe for concurrent use: the
// counter is atomic and the address slice is immutable after construction.
type connectionPool struct {
	addresses []*url.URL
	next      atomic.Uint64
}

// newConnectionPool creates a connectionPool over the given addresses.
// Addresses must already be validated and normalized (see normalizeAddresses).
func newConnectionPool(addresses []*url.URL) *connectionPool {
	return &connectionPool{addresses: addresses}
}
```

`DefaultClient` is **unchanged** — the pool is captured by the `OnBeforeRequest` closure inside
`New` (no new struct fields on `DefaultClient` are required).

---

## 5. Function signatures (exact Go)

New/changed unexported helpers (all in `connectionpool.go` unless noted):

```go
// connectionpool.go
func newConnectionPool(addresses []*url.URL) *connectionPool

// next returns the next address in round-robin order. Never nil when the
// pool is non-empty.
func (p *connectionPool) next() *url.URL

// rewrite mutates req.URL to point at the next address, but only when the pool
// holds more than one address AND req.URL is a relative path. No-op otherwise.
func (p *connectionPool) rewrite(req *resty.Request)

// joinURL joins a base address with a relative request path, mirroring resty's
// BaseURL composition.
func joinURL(addr *url.URL, rawPath string) string

// normalizeAddresses parses and validates addresses; returns the parsed list.
func normalizeAddresses(addresses []string) ([]*url.URL, error)
```

Changed exported API:

```go
// client.go — signature UNCHANGED
func New(cfg *Config, logger *logrus.Entry) (Client, error)
```

No new exported functions/methods. `Config.Addresses` is the only address-related exported symbol.

Exact bodies (reference for the implementer):

```go
// next returns the next address in round-robin order: 0, 1, ..., n-1, 0, ...
func (p *connectionPool) next() *url.URL {
	n := p.next.Add(1) // 1, 2, 3, ...
	return p.addresses[(n-1)%uint64(len(p.addresses))]
}

// rewrite advances the round-robin selection and rewrites req.URL for the next
// address. It is a no-op for 0 or 1 address, or when the caller supplied an
// absolute URL (user override).
func (p *connectionPool) rewrite(req *resty.Request) {
	if len(p.addresses) <= 1 {
		return
	}
	u, err := url.Parse(req.URL)
	if err != nil || u.IsAbs() {
		return
	}
	req.URL = joinURL(p.next(), req.URL)
}

// joinURL mirrors resty's BaseURL composition: the address's trailing slash is
// trimmed, the path is ensured to start with "/", and any address path (e.g.
// "/es" for a reverse proxy) is preserved and prepended.
func joinURL(addr *url.URL, rawPath string) string {
	path := rawPath
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(addr.String(), "/") + path
}

// normalizeAddresses parses and validates each address. Each address must
// parse and have an http/https scheme. Trailing slashes are trimmed (matching
// the official client). A nil/empty slice yields nil (no base URL; resty uses
// per-request absolute URLs). Returns *types.ElasticsearchError{Status:400} on
// the first invalid address.
func normalizeAddresses(addresses []string) ([]*url.URL, error) {
	if len(addresses) == 0 {
		return nil, nil
	}
	urls := make([]*url.URL, 0, len(addresses))
	for _, addr := range addresses {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			return nil, &types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{
				Reason: "invalid address: empty string in Addresses",
			}}
		}
		u, err := url.Parse(strings.TrimRight(addr, "/"))
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return nil, &types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{
				Reason: fmt.Sprintf("invalid address: %s must be http(s)", addr),
			}}
		}
		urls = append(urls, u)
	}
	return urls, nil
}
```

---

## 6. Round-robin / failover algorithm (precise)

State: `connectionPool{addresses []*url.URL; next atomic.Uint64}`. `addresses` is read-only after
`New`; `next` starts at 0.

Selection (`next()`):
- `n := next.Add(1)` (atomic increment, 1,2,3,...).
- `idx := (n-1) % len(addresses)`; return `addresses[idx]`.
- First call → index 0, second → index 1, … wraps after `len(addresses)`.

Per request (via the `OnBeforeRequest` hook):
1. If `len(addresses) <= 1` → return unchanged (resty applies `SetBaseURL(first)` for the 1-address
   case; no base URL for the 0-address case).
2. If `req.URL` is already absolute (user passed a full URL) → return unchanged (explicit override).
3. Else `req.URL = joinURL(next(), req.URL)` — round-robin load balancing.

Failover (retry interplay):
- resty executes the request against the selected address.
- If it fails and a `RetryCondition` returns true (`DefaultRetryConditions`: network error, 429,
  5xx except 501), resty's `Backoff` loop performs the next attempt; `Request.Execute` resets
  `req.URL` to the original relative path, re-runs our hook, and `next()` yields the **next** address.
- Thus: node A network-error/5xx → retry hits node B → (if it fails) node C → … up to `RetryCount`
  extra attempts. When all addresses/attempts are exhausted, resty returns the last error.
- With `RetryCount == 0`: a single attempt against the round-robin-selected node; no failover
  (consistent with today's "no retries" default).

Thread safety: `atomic.Uint64` makes `next()` safe for concurrent goroutines; the slice is immutable.

---

## 7. Breaking change & migration

**Breaking change:** `Config.URL` is removed from the public API. Code that set
`URL: "https://x:9200"` will no longer compile. This is intentional.

**Migration (mechanical):**
```go
// before
elasticsearch.Config{ URL: "https://node1:9200" }

// after
elasticsearch.Config{ Addresses: []string{"https://node1:9200"} }

// multiple nodes
elasticsearch.Config{ Addresses: []string{"https://node1:9200", "https://node2:9200"} }
```

There is no persisted state and no runtime data migration; the change is purely a source-level API
change (one field renamed/retyped from `string` to `[]string`).

**What remains equivalent to the old single-URL behavior:**
- One element in `Addresses` ⇒ identical to the old `URL` (hook no-op; `SetBaseURL(first)`).
- Empty `Addresses` ⇒ identical to the old empty-config behavior (no base URL; per-request absolute
  URLs).
- `RestyClient()` returns the same `*resty.Client`; `.BaseURL` reflects the first address;
  `.Transport()` still returns `*http.Transport`.
- The invalid-address error is still `*types.ElasticsearchError` with status 400; only the wording
  changes from `"invalid URL: …"` to `"invalid address: …"` (the `URL` concept is gone).

---

## 8. Validation rules

Validated in `normalizeAddresses` (called from `New`, before the client is built):

| Rule | Accepted | Rejected → error |
|---|---|---|
| Parse | valid `url.Parse` | parse error (`"://bad"`) → status 400 |
| Scheme | `http` or `https` | anything else (`ftp://`, `google.com` no-scheme) → status 400 |
| Empty element | non-empty after `strings.TrimSpace` | `""` / whitespace → status 400 |
| Trailing slash | trimmed (e.g. `"https://h:9200/"` → `"https://h:9200"`) | — |
| Path | preserved (e.g. `"https://h:9200/es"`) | — |
| Port | not required (no default injected) | — |
| Userinfo | preserved; redacted in logs via `RedactURL` | — |
| Duplicates | allowed (not deduplicated) | — |
| Empty/nil list | allowed (no base URL) | — |

Additional checks in `New` (after `normalizeAddresses`):

- **Cleartext credentials**: if credentials are configured and any parsed address has scheme `http`
  while `AllowInsecureHTTP` is false → status 400 with message
  `"credentials configured over insecure http:// URL; use https:// or set Config.AllowInsecureHTTP=true"`.
- **`TLSSkipVerify` + credentials** → warning log (unchanged).

Validation runs **at construction (`New`)** and returns an error; it never panics (matches this
codebase's convention — `New` returns `(Client, error)`).

---

## 9. Edge cases

| Case | Behavior |
|---|---|
| Empty `Addresses` (`nil` or `[]string{}`) | No base URL; per-request absolute URLs (current empty-config behavior). Hook is a no-op. |
| Empty element in `Addresses` | `New` returns status-400 error. |
| Duplicate addresses | Allowed; the duplicate is selected twice per cycle (no dedupe, matching official client). |
| Address with path (`https://h:9200/es`) | Path preserved and prepended to every request path (e.g. `/es/_search`). |
| Trailing slash on address | Trimmed during normalization. |
| Mixed http/https addresses | Allowed. If credentials are set and any address is http, `New` errors unless `AllowInsecureHTTP`. |
| Unreachable node | With `RetryCount>0`: retry advances to the next address (failover). With `RetryCount==0`: request fails immediately. |
| All nodes down | All attempts exhausted; resty returns the last error (network error or parsed ES error). |
| Concurrent requests | Safe: `atomic.Uint64` counter; immutable slice. |
| Context cancellation | Propagates through the request context exactly as today (no change); resty aborts in-flight requests. |
| Absolute URL passed to a service/`RestyClient()` call | Hook leaves it untouched (explicit user override). |
| Address with userinfo | Preserved (net/http uses it for basic auth); `RedactURL` keeps it out of logs/telemetry. Recommend using `Username`/`Password` instead. |

---

## 10. Error handling (consistent with `errors.go`)

No new error types. Reuse `types.ElasticsearchError` / `types.ElasticsearchErrorDetails`:

- Invalid address → `&types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{Reason: "invalid address: … must be http(s)"}}`.
- Cleartext credentials over http → `&types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{Reason: "credentials configured over insecure http:// URL; …"}}`.

`errors.go` / `common.go` are **unchanged** (no new helpers needed; the existing `IsNotFound` /
`IsConflict` / `IsUnauthorized` remain untouched).

---

## 11. Files to add / modify (exact changes)

### 11.1 NEW: `connectionpool.go` (root package `elasticsearch`)
Add `connectionPool` type, `newConnectionPool`, `(*connectionPool).next`, `(*connectionPool).rewrite`,
`joinURL`, `normalizeAddresses` (exact code in §4–§5). Imports: `fmt`, `net/url`, `strings`,
`sync/atomic`, `github.com/disaster37/elasticsearch/v9/types`, `github.com/go-resty/resty/v2`.

### 11.2 MODIFY: `client.go`
1. `Config`: **remove** the `URL string` field and its doc comment (lines ~200-202); **add**
   `Addresses []string` with the doc comment from §4.1. All other fields unchanged.
2. `New` — replace the address/validation blocks:
   - Remove `if cfg.URL != "" { c.SetBaseURL(cfg.URL) }` (lines ~326-328).
   - Remove the `url.Parse(cfg.URL)` scheme check and the single-URL cleartext-credential check
     (lines ~358-372), **and** the `net/url` import (no longer used in `client.go`; `strings`, `fmt`,
     `net/http` remain used).
   - Insert, before the auth section (after the `Content-Type` header line):

```go
	// Parse and validate the configured addresses. An empty list is allowed
	// (resty then uses per-request absolute URLs). normalizeAddresses enforces
	// the http/https scheme.
	parsedAddrs, err := normalizeAddresses(cfg.Addresses)
	if err != nil {
		return nil, err
	}
	pool := newConnectionPool(parsedAddrs)

	// SetBaseURL makes RestyClient().BaseURL reflect the first address and
	// remains the fallback for single-address and absolute-URL requests.
	if len(parsedAddrs) > 0 {
		c.SetBaseURL(parsedAddrs[0].String())
	}

	// Round-robin across addresses. No-op for 0 or 1 address. Registered before
	// the logging hook below so the logged URL is the actual node.
	c.OnBeforeRequest(func(_ *resty.Client, req *resty.Request) error {
		pool.rewrite(req)
		return nil
	})
```

   - Keep the auth section (unchanged), but replace the old scheme/creds validation with:

```go
	// Guard against sending credentials in cleartext over any http:// address.
	// An empty address list is allowed (resty uses a per-request URL).
	credsConfigured := cfg.Username != "" || cfg.Password != "" || cfg.APIKey != "" || cfg.BearerToken != ""
	if credsConfigured {
		for _, u := range parsedAddrs {
			if u.Scheme == "http" && !cfg.AllowInsecureHTTP {
				return nil, &types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{
					Reason: "credentials configured over insecure http:// URL; use https:// or set Config.AllowInsecureHTTP=true",
				}}
			}
		}
	}
	if cfg.TLSSkipVerify && credsConfigured {
		logger.Warn("TLSSkipVerify is enabled while credentials are configured; credentials are protected only by the (unverified) TLS layer")
	}
```

   - Note: `parsedAddrs, err := …` is the first declaration of `err` in `New`; no other `err`
     redeclaration exists afterward, so no shadowing conflict.
   - All other `New` logic (TLS config, transport clone, retry config, service construction, logging
     hook) is **unchanged**. The existing logging `OnBeforeRequest`/`OnAfterResponse` hooks stay where
     they are and now observe the absolute URL (no code change).

3. `DefaultClient` struct and its methods: **unchanged**.

### 11.3 MODIFY: `client_test.go`
- Replace **every** `URL: "…"` Config field with `Addresses: []string{"…"}` (lines 28, 35, 41, 47,
  54, 59, 65, 71, 89, 139, 149, 157, 165, 172, 181, 190, 207, 215, 222, 229, 236, 245, 254, 264,
  276, 286, 298, 318, 329, 345). The two `New(&Config{}, …)` cases (lines 200, 308) are already field-
  free and remain as-is.
- Rename tests whose names reference the removed `URL` concept:
  - `TestNewInvalidURLNoScheme` → `TestNewInvalidAddressNoScheme`
  - `TestNewInvalidURL` → `TestNewInvalidAddress`
  - `TestNewFTPURLRejected` → `TestNewFTPAddressRejected`
  - `TestNewEmptyURL` → `TestNewEmptyAddresses`
  - `TestNewRestyClientEmptyURL` → `TestNewRestyClientEmptyAddresses`
- Update comment text that mentions "URL" where it now refers to an address (e.g. the comment above
  `TestNewInsecureHTTPWithCredentials` at ~line 136). `TestNewRestyClientBaseURL` keeps its name (it
  asserts resty's `BaseURL` field, not `Config.URL`) but its body changes to `Addresses`.

### 11.4 MODIFY: `client_retry_test.go`
Replace the five `URL: srv.URL` Config fields (lines 150, 180, 211, 242, 273) with
`Addresses: []string{srv.URL}`. Test names are unaffected.

### 11.5 MODIFY: `example_test.go`
Replace `URL: "https://localhost:9200"` with `Addresses: []string{"https://localhost:9200"}` in
`ExampleNew` (line 15) and `ExampleNew_apiKey` (line 35).

### 11.6 MODIFY: `api/integration_helper_test.go`
Replace `URL: url` (line 28) with `Addresses: []string{url}`. (The helper's own `ELASTICSEARCH_URL`
env-var read stays; it is test plumbing, not part of the client library.)

### 11.7 No changes required
`common.go`, `errors.go`, `types/*`, `api/*` (production code: `api/helpers.go`, `api/options.go`,
`api/errors.go`), `trace/opentelemetry/*`, `generate.go`, `go.mod`, `go.sum`. `api/test_helpers_test.go`
and `api/security_test.go` call `c.SetBaseURL(...)` on a `*resty.Client` directly (resty's own API, not
`Config`) — leave unchanged. (Note: the historical `.opencode/plans/elasticsearch-go-client-plan.md`
document still mentions `URL`; it is an archived planning doc and is out of scope for this change.)

---

## 12. Testing plan

Test style mirrors the repo: `package elasticsearch`, `testify` `assert`/`require`, `httptest`
servers, atomic hit counters (see `client_retry_test.go`).

### 12.1 NEW: `connectionpool_test.go` (unit tests — `package elasticsearch`)
- `TestNormalizeAddressesValid` — `["https://n1:9200","http://n2:9200"]` → 2 URLs, schemes kept, no err.
- `TestNormalizeAddressesTrimsTrailingSlash` — `["https://n1:9200/"]` → `URL.String()=="https://n1:9200"`.
- `TestNormalizeAddressesPreservesPath` — `["https://h:9200/es/"]` → path `/es`.
- `TestNormalizeAddressesInvalidScheme` — `["ftp://h:9200"]` → err, status 400.
- `TestNormalizeAddressesNoScheme` — `["google.com"]` → err, status 400.
- `TestNormalizeAddressesMalformed` — `["://bad"]` → err, status 400.
- `TestNormalizeAddressesEmptyElement` — `["https://a:9200","  "]` → err, status 400.
- `TestNormalizeAddressesEmptySlice` — `[]string{}` → `nil, nil`; also `nil` → `nil, nil`.
- `TestConnectionPoolNextRoundRobin` — 3 addresses; 6 `next()` calls → indices `[0,1,2,0,1,2]`.
- `TestJoinURL` — table: `("https://h:9200","/_search")→"https://h:9200/_search"`;
  `("https://h:9200/es/","/_search")→"https://h:9200/es/_search"`; `("https://h:9200/","/")→"https://h:9200/"`;
  `("https://h:9200","")→"https://h:9200/"`.
- `TestConnectionPoolRewrite` — pool of 2; `&resty.Request{URL:"/_search"}`; `rewrite` → `addr0/_search`;
  second request → `addr1/_search`; absolute URL request → unchanged; single-address pool → unchanged;
  empty pool → unchanged.

### 12.2 NEW integration tests (httptest) — add to `connectionpool_test.go`
- `TestNewMultipleAddressesRoundRobin` — 3 `httptest` servers each returning `200 {}`; client with
  `Addresses: [srv1,srv2,srv3]`; issue 3 requests via `client.RestyClient().R().Get("/_test")`;
  assert each server got exactly 1 hit.
- `TestNewMultipleAddressesFailover` — srvA always `503`, srvB `200`; client
  `Addresses:[A,B]`, `RetryCount:1`, `RetryWaitTime/MaxWaitTime:1ms`, `RetryConditions: DefaultRetryConditions()`;
  one request → `200`; assert A hit once and B hit once (failover to next address).
- `TestNewMultipleAddressesNoFailoverWithoutRetry` — srvA `503`, srvB `200`; `RetryCount` default (0);
  one request → returns error (status 503); assert B never hit.
- `TestNewAddressesBaseURL` — `New(&Config{Addresses:[]string{"https://a:9200","https://b:9200"}})` →
  `RestyClient().BaseURL == "https://a:9200"`.
- `TestNewInvalidAddress` — `New(&Config{Addresses:[]string{"ftp://a:9200"}})` → err, status 400.
- `TestNewInsecureHTTPAddressWithCredentials` — `Addresses:[]string{"http://a:9200"}` + creds → err 400;
  with `AllowInsecureHTTP:true` → ok.

### 12.3 Existing tests to update (enumerated in §11.3–§11.6)
- `client_test.go`: all `URL:` literals → `Addresses: []string{...}`; renames listed in §11.3.
- `client_retry_test.go`: 5 `URL: srv.URL` literals → `Addresses: []string{srv.URL}`.
- `example_test.go`: 2 examples.
- `api/integration_helper_test.go`: 1 literal.
Run `go test ./...` (and `go test -tags integration ./...` if an ES is available) to confirm.

---

## 13. Documentation updates

- `README.md`: **Quick start** must now use `Addresses: []string{"https://localhost:9200"}`; add a
  short "Connecting to multiple nodes" note showing a two-node `Addresses` list and that failover
  requires `RetryCount > 0`.
- `doc.go`: update the quick-start example to use `Addresses`; add a one-line mention of
  round-robin/failover.
- `example_test.go`: update the two examples and (optionally) add `ExampleNew_multipleAddresses`.
- `README_RETRY.md`: update the four code examples from `URL:` to `Addresses:`; add a "Failover across
  multiple addresses" subsection explaining that with `Addresses` + `RetryCount > 0` each retry attempt
  targets the next address in round-robin order.

---

## 14. Preserved behavior checklist (for reviewers)

- [ ] Single endpoint (`Addresses` with one element) behaves identically to the old `URL` path (base URL set, scheme validated, cleartext-credential guard, same error type/status).
- [ ] Empty `Addresses` behaves identically to the old empty-config path (no base URL; per-request absolute URLs).
- [ ] `RestyClient()` returns the same `*resty.Client`; `.BaseURL` equals the first address; `.Transport()` still returns `*http.Transport`.
- [ ] Retry behavior is unchanged (same `RetryCount`/`RetryConditions` semantics).
- [ ] No signature changes to `New`, `Client`, `DefaultClient`, or any service interface.
- [ ] No change to `types.ElasticsearchError` shape or the `IsNotFound`/`IsConflict`/`IsUnauthorized` helpers.
- [ ] OpenTelemetry middleware, logging redaction (`RedactURL`), and TLS/transport configuration are unchanged.
- [ ] `Config.URL` no longer exists anywhere in the public API, tests, examples, or docs.

---

## 15. Ordered implementation checklist

1. Create `connectionpool.go` with `connectionPool`, `newConnectionPool`, `next`, `rewrite`, `joinURL`,
   `normalizeAddresses` (exact code in §4–§5).
2. In `client.go` `Config`: remove the `URL string` field and its doc comment; add `Addresses []string`
   with the §4.1 doc comment.
3. In `client.go` `New`: remove the `SetBaseURL(cfg.URL)` block and the `url.Parse(cfg.URL)` validation
   block; remove the now-unused `net/url` import.
4. In `client.go` `New`: add the address-resolution block (normalize → pool → `SetBaseURL(first)` →
   round-robin `OnBeforeRequest` hook) before the auth section.
5. In `client.go` `New`: replace the scheme/creds validation with the per-address cleartext-credential
   loop + `TLSSkipVerify` warning.
6. Update `client_test.go` (§11.3): swap every `URL:` → `Addresses: []string{...}`; apply the renames.
7. Update `client_retry_test.go` (§11.4), `example_test.go` (§11.5), `api/integration_helper_test.go` (§11.6).
8. Add `connectionpool_test.go` unit tests (§12.1).
9. Add `connectionpool_test.go` httptest integration tests (§12.2).
10. Update `README.md`, `doc.go`, `README_RETRY.md` (§13).
11. Run `go test ./...`, `go vet ./...`, `gofmt -l .` (and the repo's lint/format targets) to confirm no
    regressions and no unused imports. Grep the repo to confirm no remaining `Config.URL` references.
12. (Optional) Run `go generate ./...` and confirm the generated `mock/client.go` still compiles (the
    `Client` interface is unchanged, so no mock diff is expected).

---

## 16. Out of scope

- `ELASTICSEARCH_URL` env-var support inside the client library (not requested; the integration test
  helper's env-var read is separate test plumbing and stays as-is).
- Cloud ID support, node discovery, and dead-node resurrection/mark-dead semantics (official-client
  extras that resty's retry model does not need).
- Per-address TLS/transport pools (net/http already pools per host on the single shared transport).
- Changing the default port or injecting `:9200`.
