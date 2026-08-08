package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitFeaturesGetFeatures(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"features":[{"name":"tsdb","description":"time series"}]}`)
	}))
	defer srv.Close()
	svc := NewFeaturesService(restyClient(srv), testLogger())
	resp, err := svc.GetFeatures(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_features", gotPath)
	require.Len(t, resp.Features, 1)
	assert.Equal(t, "tsdb", resp.Features[0].Name)
}

func TestUnitFeaturesGetFeaturesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewFeaturesService(restyClient(srv), testLogger())
	_, err := svc.GetFeatures(context.Background())
	assert.Error(t, err)
}

func TestUnitFeaturesGetFeaturesNetworkError(t *testing.T) {
	svc := NewFeaturesService(deadClient(), testLogger())
	_, err := svc.GetFeatures(context.Background())
	assert.Error(t, err)
}

func TestUnitFeaturesResetFeatures(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewFeaturesService(restyClient(srv), testLogger())
	resp, err := svc.ResetFeatures(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_features/_reset", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitFeaturesResetFeaturesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewFeaturesService(restyClient(srv), testLogger())
	_, err := svc.ResetFeatures(context.Background())
	assert.Error(t, err)
}

func TestUnitFeaturesResetFeaturesNetworkError(t *testing.T) {
	svc := NewFeaturesService(deadClient(), testLogger())
	_, err := svc.ResetFeatures(context.Background())
	assert.Error(t, err)
}
