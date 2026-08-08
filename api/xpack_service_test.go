package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitXpackInfo(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"build":{"hash":"abc"},"license":{"mode":"basic"},"features":{"security":{"available":true,"enabled":true}},"tagline":"You know, for X"}`)
	}))
	defer srv.Close()
	svc := NewXpackService(restyClient(srv), testLogger())
	resp, err := svc.Info(context.Background(), &XpackInfoParams{
		Categories: []string{"build", "license"},
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_xpack", gotPath)
	assert.Equal(t, "build,license", gotQuery["categories"])
	assert.Equal(t, "abc", resp.Build["hash"])
	assert.Equal(t, "basic", resp.License["mode"])
	require.Contains(t, resp.Features, "security")
	assert.True(t, resp.Features["security"].Available)
	assert.True(t, resp.Features["security"].Enabled)
	assert.Equal(t, "You know, for X", resp.Tagline)
}

func TestUnitXpackInfoError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewXpackService(restyClient(srv), testLogger())
	_, err := svc.Info(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitXpackInfoNetworkError(t *testing.T) {
	svc := NewXpackService(deadClient(), testLogger())
	_, err := svc.Info(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitXpackUsage(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"security":{"available":true,"enabled":true,"roles":10},"monitoring":{"available":true,"enabled":false}}`)
	}))
	defer srv.Close()
	svc := NewXpackService(restyClient(srv), testLogger())
	resp, err := svc.Usage(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_xpack/usage", gotPath)
	require.Contains(t, resp, "security")
	assert.True(t, resp["security"].Available)
	assert.True(t, resp["security"].Enabled)
	require.Contains(t, resp["security"].Data, "roles")
	assert.Equal(t, float64(10), resp["security"].Data["roles"])
	require.Contains(t, resp, "monitoring")
	assert.True(t, resp["monitoring"].Available)
	assert.False(t, resp["monitoring"].Enabled)
}

func TestUnitXpackUsageError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewXpackService(restyClient(srv), testLogger())
	_, err := svc.Usage(context.Background())
	assert.Error(t, err)
}

func TestUnitXpackUsageNetworkError(t *testing.T) {
	svc := NewXpackService(deadClient(), testLogger())
	_, err := svc.Usage(context.Background())
	assert.Error(t, err)
}
