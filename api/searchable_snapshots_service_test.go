package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSearchableSnapshotsMount(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"snapshot":"snap1","repository":"repo1"}`)
	}))
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	resp, err := svc.Mount(context.Background(), "repo1", "snap1", map[string]any{"index": "idx"}, &SearchableSnapshotsMountParams{
		WaitForCompletion:   ptrBool(true),
		WaitForActiveShards: "1",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/snap1/_mount", gotPath)
	assert.Equal(t, "true", gotQuery["wait_for_completion"])
	assert.Equal(t, "1", gotQuery["wait_for_active_shards"])
	assert.Contains(t, gotBody, "idx")
	assert.Contains(t, string(resp), "snapshot")
}

func TestUnitSearchableSnapshotsMountEmptyRepo(t *testing.T) {
	svc := NewSearchableSnapshotsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Mount(context.Background(), "", "snap1", map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsMountEmptySnapshot(t *testing.T) {
	svc := NewSearchableSnapshotsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Mount(context.Background(), "repo1", "", map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsMountError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	_, err := svc.Mount(context.Background(), "repo1", "snap1", map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsMountNetworkError(t *testing.T) {
	svc := NewSearchableSnapshotsService(deadClient(), testLogger())
	_, err := svc.Mount(context.Background(), "repo1", "snap1", map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsStats(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"indices":{"idx":{"total":{}}}}`)
	}))
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	resp, err := svc.Stats(context.Background(), []string{"idx"}, &SearchableSnapshotsStatsParams{
		Level: LevelIndices,
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/idx/_searchable_snapshots/stats", gotPath)
	assert.Equal(t, "indices", gotQuery["level"])
	assert.Contains(t, string(resp), "indices")
}

func TestUnitSearchableSnapshotsStatsAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	_, err := svc.Stats(context.Background(), nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "/_searchable_snapshots/stats", gotPath)
}

func TestUnitSearchableSnapshotsStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	_, err := svc.Stats(context.Background(), nil, nil)
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsStatsNetworkError(t *testing.T) {
	svc := NewSearchableSnapshotsService(deadClient(), testLogger())
	_, err := svc.Stats(context.Background(), nil, nil)
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsCacheStats(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"nodes":{"node1":{}}}`)
	}))
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	resp, err := svc.CacheStats(context.Background(), []string{"node1"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_searchable_snapshots/cache/stats/node1", gotPath)
	assert.Contains(t, string(resp), "nodes")
}

func TestUnitSearchableSnapshotsCacheStatsAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	_, err := svc.CacheStats(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_searchable_snapshots/cache/stats", gotPath)
}

func TestUnitSearchableSnapshotsCacheStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	_, err := svc.CacheStats(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsCacheStatsNetworkError(t *testing.T) {
	svc := NewSearchableSnapshotsService(deadClient(), testLogger())
	_, err := svc.CacheStats(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsClearCache(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	resp, err := svc.ClearCache(context.Background(), []string{"idx"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx/_searchable_snapshots/cache/clear", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSearchableSnapshotsClearCacheAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	_, err := svc.ClearCache(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_searchable_snapshots/cache/clear", gotPath)
}

func TestUnitSearchableSnapshotsClearCacheError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	_, err := svc.ClearCache(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsClearCacheNetworkError(t *testing.T) {
	svc := NewSearchableSnapshotsService(deadClient(), testLogger())
	_, err := svc.ClearCache(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsRepositoryStats(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"repositories":{"repo1":{}}}`)
	}))
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	resp, err := svc.RepositoryStats(context.Background(), "repo1")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/_stats", gotPath)
	assert.Contains(t, string(resp), "repositories")
}

func TestUnitSearchableSnapshotsRepositoryStatsEmpty(t *testing.T) {
	svc := NewSearchableSnapshotsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.RepositoryStats(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsRepositoryStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchableSnapshotsService(restyClient(srv), testLogger())
	_, err := svc.RepositoryStats(context.Background(), "repo1")
	assert.Error(t, err)
}

func TestUnitSearchableSnapshotsRepositoryStatsNetworkError(t *testing.T) {
	svc := NewSearchableSnapshotsService(deadClient(), testLogger())
	_, err := svc.RepositoryStats(context.Background(), "repo1")
	assert.Error(t, err)
}
