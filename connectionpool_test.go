// connectionpool_test.go contains unit tests for address normalization, the
// round-robin connection pool, and the multi-address failover behavior enabled
// by New(). Tests use httptest.NewServer for HTTP mocking and require no
// Elasticsearch cluster.
package elasticsearch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeAddressesValid(t *testing.T) {
	urls, err := normalizeAddresses([]string{"https://n1:9200", "http://n2:9200"})
	require.NoError(t, err)
	require.Len(t, urls, 2)
	assert.Equal(t, "https", urls[0].Scheme)
	assert.Equal(t, "http", urls[1].Scheme)
}

func TestNormalizeAddressesTrimsTrailingSlash(t *testing.T) {
	urls, err := normalizeAddresses([]string{"https://n1:9200/"})
	require.NoError(t, err)
	require.Len(t, urls, 1)
	assert.Equal(t, "https://n1:9200", urls[0].String())
}

func TestNormalizeAddressesPreservesPath(t *testing.T) {
	urls, err := normalizeAddresses([]string{"https://h:9200/es/"})
	require.NoError(t, err)
	require.Len(t, urls, 1)
	assert.Equal(t, "/es", urls[0].Path)
}

func TestNormalizeAddressesInvalidScheme(t *testing.T) {
	_, err := normalizeAddresses([]string{"ftp://h:9200"})
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

func TestNormalizeAddressesNoScheme(t *testing.T) {
	_, err := normalizeAddresses([]string{"google.com"})
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

func TestNormalizeAddressesMalformed(t *testing.T) {
	_, err := normalizeAddresses([]string{"://bad"})
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

func TestNormalizeAddressesEmptyElement(t *testing.T) {
	_, err := normalizeAddresses([]string{"https://a:9200", "  "})
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

func TestNormalizeAddressesEmptySlice(t *testing.T) {
	urls, err := normalizeAddresses([]string{})
	assert.Nil(t, urls)
	assert.NoError(t, err)

	urls, err = normalizeAddresses(nil)
	assert.Nil(t, urls)
	assert.NoError(t, err)
}

func TestConnectionPoolNextRoundRobin(t *testing.T) {
	urls, err := normalizeAddresses([]string{"https://n1:9200", "https://n2:9200", "https://n3:9200"})
	require.NoError(t, err)
	pool := newConnectionPool(urls)

	var got []string
	for i := 0; i < 6; i++ {
		got = append(got, pool.next().String())
	}
	assert.Equal(t, []string{
		"https://n1:9200",
		"https://n2:9200",
		"https://n3:9200",
		"https://n1:9200",
		"https://n2:9200",
		"https://n3:9200",
	}, got)
}

func TestJoinURL(t *testing.T) {
	cases := []struct {
		addr string
		path string
		want string
	}{
		{"https://h:9200", "/_search", "https://h:9200/_search"},
		{"https://h:9200/es/", "/_search", "https://h:9200/es/_search"},
		{"https://h:9200/", "/", "https://h:9200/"},
		{"https://h:9200", "", "https://h:9200/"},
		{"https://h:9200", "search", "https://h:9200/search"},
	}
	for _, tc := range cases {
		u, err := url.Parse(tc.addr)
		require.NoError(t, err)
		assert.Equal(t, tc.want, joinURL(u, tc.path), "addr=%q path=%q", tc.addr, tc.path)
	}
}

func TestConnectionPoolRewrite(t *testing.T) {
	urls, err := normalizeAddresses([]string{"https://n1:9200", "https://n2:9200"})
	require.NoError(t, err)
	pool := newConnectionPool(urls)

	req := &resty.Request{URL: "/_search"}
	pool.rewrite(req)
	assert.Equal(t, "https://n1:9200/_search", req.URL)

	req2 := &resty.Request{URL: "/_search"}
	pool.rewrite(req2)
	assert.Equal(t, "https://n2:9200/_search", req2.URL)

	// Absolute URL request → unchanged (explicit user override).
	abs := &resty.Request{URL: "https://other:9200/_search"}
	pool.rewrite(abs)
	assert.Equal(t, "https://other:9200/_search", abs.URL)

	// Single-address pool → unchanged.
	single, err := normalizeAddresses([]string{"https://n1:9200"})
	require.NoError(t, err)
	singlePool := newConnectionPool(single)
	sreq := &resty.Request{URL: "/_search"}
	singlePool.rewrite(sreq)
	assert.Equal(t, "/_search", sreq.URL)

	// Empty pool → unchanged.
	emptyPool := newConnectionPool(nil)
	ereq := &resty.Request{URL: "/_search"}
	emptyPool.rewrite(ereq)
	assert.Equal(t, "/_search", ereq.URL)

	// Relative URL with "://" in query string → still rewritten (not absolute).
	// Regression test: strings.Contains("://") would incorrectly skip this.
	pool2, err := normalizeAddresses([]string{"https://n1:9200", "https://n2:9200"})
	require.NoError(t, err)
	pool2Pool := newConnectionPool(pool2)
	qreq := &resty.Request{URL: "/search?q=ab://cd"}
	pool2Pool.rewrite(qreq)
	assert.Equal(t, "https://n1:9200/search?q=ab://cd", qreq.URL)
}

func TestNewMultipleAddressesRoundRobin(t *testing.T) {
	var h1, h2, h3 int32
	mkSrv := func(counter *int32) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(counter, 1)
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{}`)
		}))
	}
	srv1 := mkSrv(&h1)
	defer srv1.Close()
	srv2 := mkSrv(&h2)
	defer srv2.Close()
	srv3 := mkSrv(&h3)
	defer srv3.Close()

	client, err := New(&Config{Addresses: []string{srv1.URL, srv2.URL, srv3.URL}}, testLogger())
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		resp, err := client.RestyClient().R().Get("/_test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&h1))
	assert.Equal(t, int32(1), atomic.LoadInt32(&h2))
	assert.Equal(t, int32(1), atomic.LoadInt32(&h3))
}

func TestNewMultipleAddressesFailover(t *testing.T) {
	var aHits, bHits int32
	srvA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&aHits, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, `{"error":{"type":"es_error","reason":"unavailable"},"status":503}`)
	}))
	defer srvA.Close()
	srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&bHits, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{}`)
	}))
	defer srvB.Close()

	client, err := New(&Config{
		Addresses:        []string{srvA.URL, srvB.URL},
		RetryCount:       1,
		RetryWaitTime:    1 * time.Millisecond,
		RetryMaxWaitTime: 1 * time.Millisecond,
		RetryConditions:  DefaultRetryConditions(),
	}, testLogger())
	require.NoError(t, err)

	resp, err := client.RestyClient().R().Get("/_test")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, int32(1), atomic.LoadInt32(&aHits))
	assert.Equal(t, int32(1), atomic.LoadInt32(&bHits))
}

// TestNewMultipleAddressesAutoFailover verifies that configuring more than one
// address without any explicit retry configuration automatically enables retry
// and default conditions, so a request that fails against the first node fails
// over to the next one.
func TestNewMultipleAddressesAutoFailover(t *testing.T) {
	var aHits, bHits int32
	srvA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&aHits, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, `{"error":{"type":"es_error","reason":"unavailable"},"status":503}`)
	}))
	defer srvA.Close()
	srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&bHits, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{}`)
	}))
	defer srvB.Close()

	// No RetryCount/RetryConditions: the multi-address default must fail over.
	client, err := New(&Config{
		Addresses: []string{srvA.URL, srvB.URL},
	}, testLogger())
	require.NoError(t, err)

	resp, err := client.RestyClient().R().Get("/_test")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, int32(1), atomic.LoadInt32(&aHits))
	assert.Equal(t, int32(1), atomic.LoadInt32(&bHits))
}

func TestNewAddressesBaseURL(t *testing.T) {
	client, err := New(&Config{Addresses: []string{"https://a:9200", "https://b:9200"}}, testLogger())
	require.NoError(t, err)
	assert.Equal(t, "https://a:9200", client.RestyClient().BaseURL)
}

func TestNewInvalidAddressFTP(t *testing.T) {
	_, err := New(&Config{Addresses: []string{"ftp://a:9200"}}, testLogger())
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

func TestNewInsecureHTTPAddressWithCredentials(t *testing.T) {
	_, err := New(&Config{Addresses: []string{"http://a:9200"}, Username: "u", Password: "p"}, testLogger())
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)

	c, err := New(&Config{Addresses: []string{"http://a:9200"}, Username: "u", Password: "p", AllowInsecureHTTP: true}, testLogger())
	require.NoError(t, err)
	assert.NotNil(t, c)
}

// --- Security regression tests ---

// TestNormalizeAddressesRejectsQuery verifies that addresses containing "?"
// are rejected (CWE-918: request-target confusion — the query would swallow
// the request path in joinURL).
func TestNormalizeAddressesRejectsQuery(t *testing.T) {
	_, err := normalizeAddresses([]string{"https://host:9200/?x="})
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
	assert.Contains(t, esErr.Details.Reason, "?")
}

// TestNormalizeAddressesRejectsFragment verifies that addresses containing "#"
// are rejected (CWE-918: request-target confusion — the fragment would swallow
// the request path in joinURL).
func TestNormalizeAddressesRejectsFragment(t *testing.T) {
	_, err := normalizeAddresses([]string{"https://host:9200/#frag"})
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
	assert.Contains(t, esErr.Details.Reason, "#")
}

// TestNormalizeAddressesRejectsQueryInPath verifies that "?" in an address
// path is also rejected.
func TestNormalizeAddressesRejectsQueryInPath(t *testing.T) {
	_, err := normalizeAddresses([]string{"https://host:9200/es?x=1"})
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

// TestNormalizeAddressesRejectsFragmentInPath verifies that "#" in an address
// path is also rejected.
func TestNormalizeAddressesRejectsFragmentInPath(t *testing.T) {
	_, err := normalizeAddresses([]string{"https://host:9200/es#frag"})
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

// TestNormalizeAddressesRejectsUserinfo verifies that addresses with embedded
// userinfo (user:pass@host or token@host) are rejected (CWE-918: host confusion
// — url.Parse treats "good@evil.com:9200" as connecting to evil.com:9200).
func TestNormalizeAddressesRejectsUserinfo(t *testing.T) {
	tests := []string{
		"https://user:pass@host:9200",
		"https://token@host:9200",
		"https://api-key-12345@host:9200/",
	}
	for _, addr := range tests {
		t.Run(addr, func(t *testing.T) {
			_, err := normalizeAddresses([]string{addr})
			require.Error(t, err, "address %q should be rejected", addr)
			var esErr *types.ElasticsearchError
			require.ErrorAs(t, err, &esErr)
			assert.Equal(t, 400, esErr.Status)
			assert.Contains(t, esErr.Details.Reason, "userinfo")
		})
	}
}

// TestNormalizeAddressesRejectsUserinfoHostConfusion verifies the specific
// SSRF host-confusion case: "https://good@evil.com:9200" is parsed by
// url.Parse as connecting to evil.com:9200 (not good), so it must be rejected.
func TestNormalizeAddressesRejectsUserinfoHostConfusion(t *testing.T) {
	_, err := normalizeAddresses([]string{"https://good@evil.com:9200"})
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
	assert.Contains(t, esErr.Details.Reason, "userinfo")
}

// TestNormalizeAddressesErrorMessageDoesNotLeakInput verifies that error
// messages from normalizeAddresses do not echo the raw user-controlled address
// (CWE-200: information disclosure). Previously the error message included the
// raw address via fmt.Sprintf, which could leak credentials if an address like
// "ftp://user:pass@host" was provided.
func TestNormalizeAddressesErrorMessageDoesNotLeakInput(t *testing.T) {
	// Invalid scheme — error must not contain the raw address.
	_, err := normalizeAddresses([]string{"ftp://secret:token@evil:21"})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "secret")
	assert.NotContains(t, err.Error(), "token")
	assert.NotContains(t, err.Error(), "evil")

	// Malformed URL — error must not contain the raw address.
	_, err = normalizeAddresses([]string{"://bad"})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "://bad")
}

// TestNewRejectsAddressWithQuery verifies the full New() path rejects
// addresses with "?".
func TestNewRejectsAddressWithQuery(t *testing.T) {
	_, err := New(&Config{Addresses: []string{"https://host:9200/?x="}}, testLogger())
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

// TestNewRejectsAddressWithUserinfo verifies the full New() path rejects
// addresses with embedded userinfo.
func TestNewRejectsAddressWithUserinfo(t *testing.T) {
	_, err := New(&Config{Addresses: []string{"https://user:pass@host:9200"}}, testLogger())
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

// TestNormalizeAddressesTooMany verifies that more than maxAddresses addresses
// are rejected before any allocation (CWE-400: unbounded address list).
func TestNormalizeAddressesTooMany(t *testing.T) {
	addrs := make([]string, maxAddresses+1)
	for i := range addrs {
		addrs[i] = fmt.Sprintf("https://h%d.example:9200", i)
	}
	_, err := normalizeAddresses(addrs)
	require.Error(t, err)
	var esErr *types.ElasticsearchError
	require.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
	assert.Contains(t, esErr.Details.Reason, "too many")
}

// TestNormalizeAddressesAtMaxOK verifies that exactly maxAddresses addresses
// are accepted.
func TestNormalizeAddressesAtMaxOK(t *testing.T) {
	addrs := make([]string, maxAddresses)
	for i := range addrs {
		addrs[i] = fmt.Sprintf("https://h%d.example:9200", i)
	}
	urls, err := normalizeAddresses(addrs)
	require.NoError(t, err)
	assert.Len(t, urls, maxAddresses)
}

// TestNormalizeAddressesRejectsCRLF verifies that addresses containing CR/LF
// are rejected (Go's url.Parse rejects control characters).
func TestNormalizeAddressesRejectsCRLF(t *testing.T) {
	tests := []string{
		"https://host:9200\r\nX-Injected: true",
		"https://\rhost:9200",
		"https://host:9200/\npath",
	}
	for _, addr := range tests {
		t.Run(addr, func(t *testing.T) {
			_, err := normalizeAddresses([]string{addr})
			require.Error(t, err)
			var esErr *types.ElasticsearchError
			require.ErrorAs(t, err, &esErr)
			assert.Equal(t, 400, esErr.Status)
		})
	}
}
