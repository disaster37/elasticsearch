package elasticsearch

import (
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
)

// maxAddresses is a config-time guard against accidentally huge address lists
// (e.g. misconfigured service discovery) that would cause memory pressure.
const maxAddresses = 1000

// connectionPool holds the parsed node addresses and selects the next one in
// round-robin order for each request. It is safe for concurrent use: the
// counter is atomic and the address slice is immutable after construction.
type connectionPool struct {
	addresses []*url.URL
	counter   atomic.Uint64
}

// newConnectionPool creates a connectionPool over the given addresses.
// Addresses must already be validated and normalized (see normalizeAddresses).
func newConnectionPool(addresses []*url.URL) *connectionPool {
	return &connectionPool{addresses: addresses}
}

// next returns the next address in round-robin order: 0, 1, ..., n-1, 0, ...
func (p *connectionPool) next() *url.URL {
	n := p.counter.Add(1) // 1, 2, 3, ...
	return p.addresses[(n-1)%uint64(len(p.addresses))]
}

// rewrite advances the round-robin selection and rewrites req.URL for the next
// address. It is a no-op for 0 or 1 address, or when the caller supplied an
// absolute URL (user override). url.Parse + IsAbs() correctly distinguishes
// absolute URLs from relative paths that happen to contain "://" in a query
// string (e.g. "/search?q=ab://cd").
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
//
// Security: addresses containing "?" or "#" are rejected because these
// characters would cause joinURL to silently corrupt the request path (the
// path would become part of the query string or fragment instead of the actual
// HTTP path — CWE-918 request-target confusion). Addresses with embedded
// userinfo ("user:pass@host" or "token@host") are also rejected: the "@" in
// the authority section creates a host-confusion risk where a developer might
// think they are connecting to host A while the URL parser routes the request
// to host B (CWE-918 SSRF). Use Config.Username/Password for authentication
// instead.
func normalizeAddresses(addresses []string) ([]*url.URL, error) {
	if len(addresses) == 0 {
		return nil, nil
	}
	if len(addresses) > maxAddresses {
		return nil, &types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{
			Reason: fmt.Sprintf("invalid address: too many addresses: %d (maximum %d)", len(addresses), maxAddresses),
		}}
	}
	urls := make([]*url.URL, 0, len(addresses))
	for _, addr := range addresses {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			return nil, &types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{
				Reason: "invalid address: empty string in Addresses",
			}}
		}
		// Reject addresses containing "?" or "#" before parsing: these
		// characters have no legitimate use in an Elasticsearch node address
		// and would cause joinURL to silently corrupt the request path
		// (CWE-918 request-target confusion).
		if strings.ContainsAny(addr, "?#") {
			return nil, &types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{
				Reason: "invalid address: address must not contain '?' or '#'",
			}}
		}
		u, err := url.Parse(strings.TrimRight(addr, "/"))
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return nil, &types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{
				Reason: "invalid address: must be http(s)",
			}}
		}
		// Reject addresses with embedded userinfo (CWE-918 host confusion).
		// The "@" in the authority creates ambiguity: url.Parse treats
		// "https://good@evil.com:9200" as connecting to evil.com:9200 with
		// username "good", which is almost certainly not what the developer
		// intended. Use Config.Username/Password for authentication.
		if u.User != nil {
			return nil, &types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{
				Reason: "invalid address: userinfo (@) not allowed in address; use Config.Username/Password for authentication",
			}}
		}
		urls = append(urls, u)
	}
	return urls, nil
}
