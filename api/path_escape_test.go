package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathEscape(t *testing.T) {
	tests := []struct{ input, want string }{
		{"simple", "simple"},
		{"foo/bar", "foo%2Fbar"},
		{"1?inject=1", "1%3Finject=1"}, // '?' is escaped (prevents query injection); '=' is a legal path sub-delim
		{"a#fragment", "a%23fragment"},
		{"has space", "has%20space"},
		{"normal-index", "normal-index"},
		{"test*", "test%2A"}, // wildcards work because ES decodes
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := pathSegment(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEscapedList(t *testing.T) {
	assert.Equal(t, "", escapedList(nil))
	assert.Equal(t, "a", escapedList([]string{"a"}))
	// Each element is escaped individually; the comma separator is literal.
	assert.Equal(t, "a,b%2Fc,d", escapedList([]string{"a", "b/c", "d"}))
	// A malicious element cannot break out of its segment.
	assert.Equal(t, "a,b%3Fc,d", escapedList([]string{"a", "b?c", "d"}))
}

// TestPathFromIndicesEscaped ensures pathFromIndices escapes each index.
func TestPathFromIndicesEscaped(t *testing.T) {
	assert.Equal(t, "", pathFromIndices(nil))
	assert.Equal(t, "/a", pathFromIndices([]string{"a"}))
	assert.Equal(t, "/a,b%2Fc", pathFromIndices([]string{"a", "b/c"}))
}

// TestJoinIndicesEscaped ensures joinIndices escapes each element.
func TestJoinIndicesEscaped(t *testing.T) {
	assert.Equal(t, "", joinIndices(nil))
	assert.Equal(t, "a,b%2Fc", joinIndices([]string{"a", "b/c"}))
}

// TestPathInjectionBlocked is a regression test for CWE-20: a malicious
// index name containing '?' must not be able to inject a query parameter
// into the request. Before escaping, "/%s/_search" with index "idx?pretty=true"
// would produce "/idx?pretty=true/_search" — restructuring the request. After
// escaping, the '?' is percent-encoded so the index stays a single path
// segment and no query parameter is injected.
func TestPathInjectionBlocked(t *testing.T) {
	var rawURI, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawURI = r.RequestURI
		gotQuery = r.URL.RawQuery
		jsonResponse(w, 200, `{"took":0,"hits":{"hits":[]}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())

	// "idx?pretty=true" must NOT inject a pretty query param.
	_, err := svc.Search(context.Background(), &SearchRequest{
		Indices: []string{"idx?pretty=true"},
	})
	assert.NoError(t, err)
	// The '?' is percent-encoded on the wire (r.RequestURI is the raw line).
	assert.Contains(t, rawURI, "/idx%3Fpretty=true/_search")
	// No injected query parameter.
	assert.Equal(t, "", gotQuery)
}

// TestPathInjectionBlockedDocID is the same regression for document IDs: a
// doc id containing '/' must not be able to escape into a different
// endpoint path.
func TestPathInjectionBlockedDocID(t *testing.T) {
	var rawURI string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawURI = r.RequestURI
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"_index":"idx","_id":"x","found":false}`))
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())

	// "1/_delete" must not become "/idx/_doc/1/_delete" (a different endpoint).
	_, err := svc.Get(context.Background(), &GetRequest{
		Index: "idx",
		Id:    "1/_delete",
	})
	assert.NoError(t, err)
	// The '/' is percent-encoded on the wire so the id stays a single segment.
	assert.Contains(t, rawURI, "/idx/_doc/1%2F_delete")
}

// TestValidateRejectsControlChars is a regression test for the
// defense-in-depth CWE-20 check: IndexRequest, GetRequest, DeleteRequest,
// and UpdateRequest must reject index/id values containing ASCII control
// characters (0x00-0x1F, 0x7F).
func TestValidateRejectsControlChars(t *testing.T) {
	bad := "idx\u0001bad"
	good := "normal-index"

	cases := []struct {
		name string
		fn   func() error
	}{
		{"IndexRequest", func() error { return (&IndexRequest{Index: bad, Id: "1"}).Validate() }},
		{"IndexRequest.good", func() error { return (&IndexRequest{Index: good, Id: "1"}).Validate() }},
		{"GetRequest", func() error { return (&GetRequest{Index: bad, Id: "1"}).Validate() }},
		{"GetRequest.id", func() error { return (&GetRequest{Index: good, Id: "id\u007F"}).Validate() }},
		{"DeleteRequest", func() error { return (&DeleteRequest{Index: bad, Id: "1"}).Validate() }},
		{"UpdateRequest", func() error { return (&UpdateRequest{Index: bad, Id: "1"}).Validate() }},
		{"UpdateRequest.good", func() error { return (&UpdateRequest{Index: good, Id: "1"}).Validate() }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.fn()
			if c.name == "IndexRequest.good" || c.name == "UpdateRequest.good" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
