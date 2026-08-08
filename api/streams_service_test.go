package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitStreamsLogsEnable(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewStreamsService(restyClient(srv), testLogger())
	resp, err := svc.LogsEnable(context.Background(), map[string]any{"index": "logs"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_streams/logs/_enable", gotPath)
	assert.Contains(t, gotBody, "logs")
	assert.True(t, resp.Acknowledged)
}

func TestUnitStreamsLogsEnableError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewStreamsService(restyClient(srv), testLogger())
	_, err := svc.LogsEnable(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitStreamsLogsEnableNetworkError(t *testing.T) {
	svc := NewStreamsService(deadClient(), testLogger())
	_, err := svc.LogsEnable(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitStreamsLogsDisable(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewStreamsService(restyClient(srv), testLogger())
	resp, err := svc.LogsDisable(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_streams/logs/_disable", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitStreamsLogsDisableError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewStreamsService(restyClient(srv), testLogger())
	_, err := svc.LogsDisable(context.Background())
	assert.Error(t, err)
}

func TestUnitStreamsLogsDisableNetworkError(t *testing.T) {
	svc := NewStreamsService(deadClient(), testLogger())
	_, err := svc.LogsDisable(context.Background())
	assert.Error(t, err)
}

func TestUnitStreamsStatus(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"logs":{"enabled":true}}`)
	}))
	defer srv.Close()
	svc := NewStreamsService(restyClient(srv), testLogger())
	resp, err := svc.Status(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_streams/logs/_status", gotPath)
	assert.Contains(t, string(resp), "enabled")
}

func TestUnitStreamsStatusError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewStreamsService(restyClient(srv), testLogger())
	_, err := svc.Status(context.Background())
	assert.Error(t, err)
}

func TestUnitStreamsStatusNetworkError(t *testing.T) {
	svc := NewStreamsService(deadClient(), testLogger())
	_, err := svc.Status(context.Background())
	assert.Error(t, err)
}
