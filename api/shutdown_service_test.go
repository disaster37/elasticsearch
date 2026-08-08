package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitShutdownGetNode(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"nodes":[{"node_id":"node-1","type":"restart","reason":"test"}]}`)
	}))
	defer srv.Close()
	svc := NewShutdownService(restyClient(srv), testLogger())
	resp, err := svc.GetNode(context.Background(), "node-1")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_nodes/node-1/shutdown", gotPath)
	require.Len(t, resp.Nodes, 1)
	assert.Equal(t, "node-1", resp.Nodes[0].NodeID)
}

func TestUnitShutdownGetNodeAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"nodes":[]}`)
	}))
	defer srv.Close()
	svc := NewShutdownService(restyClient(srv), testLogger())
	_, err := svc.GetNode(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "/_nodes/shutdown", gotPath)
}

func TestUnitShutdownGetNodeError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewShutdownService(restyClient(srv), testLogger())
	_, err := svc.GetNode(context.Background(), "node-1")
	assert.Error(t, err)
}

func TestUnitShutdownGetNodeNetworkError(t *testing.T) {
	svc := NewShutdownService(deadClient(), testLogger())
	_, err := svc.GetNode(context.Background(), "node-1")
	assert.Error(t, err)
}

func TestUnitShutdownPutNode(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewShutdownService(restyClient(srv), testLogger())
	resp, err := svc.PutNode(context.Background(), "node-1", map[string]any{"type": "restart", "reason": "test"})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_nodes/node-1/shutdown", gotPath)
	assert.Contains(t, gotBody, "reason")
	assert.True(t, resp.Acknowledged)
}

func TestUnitShutdownPutNodeEmptyID(t *testing.T) {
	svc := NewShutdownService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutNode(context.Background(), "", map[string]any{})
	assert.Error(t, err)
}

func TestUnitShutdownPutNodeError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewShutdownService(restyClient(srv), testLogger())
	_, err := svc.PutNode(context.Background(), "node-1", map[string]any{})
	assert.Error(t, err)
}

func TestUnitShutdownPutNodeNetworkError(t *testing.T) {
	svc := NewShutdownService(deadClient(), testLogger())
	_, err := svc.PutNode(context.Background(), "node-1", map[string]any{})
	assert.Error(t, err)
}

func TestUnitShutdownDeleteNode(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewShutdownService(restyClient(srv), testLogger())
	resp, err := svc.DeleteNode(context.Background(), "node-1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_nodes/node-1/shutdown", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitShutdownDeleteNodeEmptyID(t *testing.T) {
	svc := NewShutdownService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteNode(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitShutdownDeleteNodeError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewShutdownService(restyClient(srv), testLogger())
	_, err := svc.DeleteNode(context.Background(), "node-1")
	assert.Error(t, err)
}

func TestUnitShutdownDeleteNodeNetworkError(t *testing.T) {
	svc := NewShutdownService(deadClient(), testLogger())
	_, err := svc.DeleteNode(context.Background(), "node-1")
	assert.Error(t, err)
}
