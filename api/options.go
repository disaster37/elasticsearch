package api

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/querydsl"
	"github.com/go-playground/validator/v10"
)

// rawBody wraps a JSON string as json.RawMessage so that resty does not
// double-encode it when setting the request body.
func rawBody(s string) json.RawMessage {
	return json.RawMessage(s)
}

// bodyProvider is implemented by request types that can serialize themselves
// into a JSON search body. *querydsl.SearchRequest satisfies this interface
// via its Body() (string, error) method.
type bodyProvider interface {
	Body() (string, error)
}

// normalizeBody prepares a request body for transmission via resty.
//
//   - string, []byte, json.RawMessage and *json.RawMessage are treated as
//     already-serialized JSON. They are wrapped as json.RawMessage so that
//     resty transmits the exact bytes without re-marshaling (see rawBody).
//   - A value implementing bodyProvider (e.g. *querydsl.SearchRequest) has its
//     Body() method called and the result wrapped as json.RawMessage.
//   - nil and typed-nil values yield a nil body.
//   - Any other value (map, struct, ...) is returned unchanged so resty
//     marshals it.
func normalizeBody(body any) (any, error) {
	if body == nil {
		return nil, nil
	}

	switch b := body.(type) {
	case string:
		return rawBody(b), nil
	case []byte:
		return rawBody(string(b)), nil
	case json.RawMessage:
		return b, nil
	case *json.RawMessage:
		if b == nil {
			return nil, nil
		}
		return *b, nil
	}

	if bp, ok := body.(bodyProvider); ok {
		// Guard against a typed nil whose Body() method would panic.
		if isNilValue(body) {
			return nil, nil
		}
		s, err := bp.Body()
		if err != nil {
			return nil, fmt.Errorf("serialize request body: %w", err)
		}
		return rawBody(s), nil
	}

	return body, nil
}

// isNilValue reports whether v holds a nil pointer or nil interface value
// (a "typed nil").
func isNilValue(v any) bool {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	}
	return false
}

// validate is the shared validator instance used by all request Validate()
// methods. WithRequiredStructEnabled makes `validate:"required"` on a struct
// field enforce non-zero values.
var validate = validator.New(validator.WithRequiredStructEnabled())

// validationError wraps a validator error with a "validation: " prefix, or
// returns nil when err is nil.
func validationError(err error) error {
	if err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	return nil
}

// ensureTrailingNewline guarantees that an NDJSON body ends with a newline,
// as required by the bulk/msearch/msearch_template endpoints.
func ensureTrailingNewline(s string) string {
	if s == "" {
		return s
	}
	if !strings.HasSuffix(s, "\n") {
		return s + "\n"
	}
	return s
}

// pathSegment percent-encodes a single dynamic path segment so that
// user-supplied values (index names, doc IDs, usernames, etc.) cannot
// restructure the request URL. Characters with special meaning in a URL
// path — '?', '#', '/', space, and others — are escaped, which prevents
// path/query injection (CWE-20). Elasticsearch percent-decodes path
// segments before matching them, so legitimate values round-trip
// unchanged on the server side.
func pathSegment(s string) string {
	return url.PathEscape(s)
}

// escapedList escapes each element individually with pathSegment, then
// joins them with a literal comma. The comma is a legal path-segment
// character and is Elasticsearch's standard list separator, so it is left
// unescaped; each element, however, is fully percent-encoded so a
// malicious element cannot break out of its segment.
func escapedList(vals []string) string {
	e := make([]string, len(vals))
	for i, v := range vals {
		e[i] = url.PathEscape(v)
	}
	return strings.Join(e, ",")
}

// joinIndices joins index names with a comma, returning "" for an empty
// slice. Each name is percent-escaped first so a value containing '/', '?',
// '#', or space cannot restructure the request path.
func joinIndices(indices []string) string {
	return escapedList(indices)
}

// pathFromIndices returns "/{indices}" when indices is non-empty, or "" for
// the cluster-wide route. Each index name is percent-escaped before
// joining.
func pathFromIndices(indices []string) string {
	if len(indices) == 0 {
		return ""
	}
	return "/" + joinIndices(indices)
}

// ensure querydsl is referenced (used by normalizeBody via bodyProvider).
var _ = querydsl.NewSearchRequest

// SecurityHasPrivilegesSpecRequest is the opt-in request for
// security.has_privileges. Body is required; User is optional and selects the
// per-user route.
type SecurityHasPrivilegesSpecRequest struct {
	User string
	Body any
}

// Validate ensures a request body is present.
func (r *SecurityHasPrivilegesSpecRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("validation: request is required")
	}
	if r.Body == nil {
		return fmt.Errorf("validation: body is required")
	}
	return nil
}
