package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitGraphExplore(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"took":10,"timed_out":false,"vertices":[],"connections":[]}`)
	}))
	defer srv.Close()
	svc := NewGraphService(restyClient(srv), testLogger())
	resp, err := svc.Explore(context.Background(), "my-index", map[string]any{"query": map[string]any{"match_all": map[string]any{}}}, &GraphExploreParams{
		Routing: "r",
		Timeout: "30s",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/my-index/_graph/explore", gotPath)
	assert.Equal(t, "r", gotQuery["routing"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Contains(t, gotBody, "match_all")
	assert.Equal(t, int64(10), resp.Took)
	assert.False(t, resp.TimeOut)
	assert.Empty(t, resp.Vertices)
	assert.Empty(t, resp.Connections)
}

func TestUnitGraphExploreEmptyIndex(t *testing.T) {
	svc := NewGraphService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Explore(context.Background(), "", map[string]any{}, nil)
	assert.Error(t, err)
}

func TestUnitGraphExploreError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewGraphService(restyClient(srv), testLogger())
	_, err := svc.Explore(context.Background(), "my-index", map[string]any{}, nil)
	assert.Error(t, err)
}

func TestUnitGraphExploreNetworkError(t *testing.T) {
	svc := NewGraphService(deadClient(), testLogger())
	_, err := svc.Explore(context.Background(), "my-index", map[string]any{}, nil)
	assert.Error(t, err)
}
