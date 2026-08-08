package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitMigrationDeprecations(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"cluster_settings":[],"node_settings":[],"index_settings":{},"ml_settings":[]}`)
	}))
	defer srv.Close()
	svc := NewMigrationService(restyClient(srv), testLogger())
	resp, err := svc.Deprecations(context.Background(), "my-index")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/my-index/_migration/deprecations", gotPath)
	assert.NotNil(t, resp)
}

func TestUnitMigrationDeprecationsAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"cluster_settings":[],"node_settings":[],"index_settings":{},"ml_settings":[]}`)
	}))
	defer srv.Close()
	svc := NewMigrationService(restyClient(srv), testLogger())
	_, err := svc.Deprecations(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "/_migration/deprecations", gotPath)
}

func TestUnitMigrationDeprecationsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMigrationService(restyClient(srv), testLogger())
	_, err := svc.Deprecations(context.Background(), "my-index")
	assert.Error(t, err)
}

func TestUnitMigrationDeprecationsNetworkError(t *testing.T) {
	svc := NewMigrationService(deadClient(), testLogger())
	_, err := svc.Deprecations(context.Background(), "my-index")
	assert.Error(t, err)
}

func TestUnitMigrationGetFeatureUpgradeStatus(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"features":[{"feature_name":"tsdb","feature_status":"upgraded"}]}`)
	}))
	defer srv.Close()
	svc := NewMigrationService(restyClient(srv), testLogger())
	resp, err := svc.GetFeatureUpgradeStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_migration/feature_upgrade", gotPath)
	require.Len(t, resp.Features, 1)
	assert.Equal(t, "tsdb", resp.Features[0].FeatureName)
}

func TestUnitMigrationGetFeatureUpgradeStatusError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMigrationService(restyClient(srv), testLogger())
	_, err := svc.GetFeatureUpgradeStatus(context.Background())
	assert.Error(t, err)
}

func TestUnitMigrationGetFeatureUpgradeStatusNetworkError(t *testing.T) {
	svc := NewMigrationService(deadClient(), testLogger())
	_, err := svc.GetFeatureUpgradeStatus(context.Background())
	assert.Error(t, err)
}

func TestUnitMigrationPostFeatureUpgrade(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewMigrationService(restyClient(srv), testLogger())
	raw, err := svc.PostFeatureUpgrade(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_migration/feature_upgrade", gotPath)
	assert.Contains(t, string(raw), "acknowledged")
}

func TestUnitMigrationPostFeatureUpgradeError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMigrationService(restyClient(srv), testLogger())
	_, err := svc.PostFeatureUpgrade(context.Background())
	assert.Error(t, err)
}

func TestUnitMigrationPostFeatureUpgradeNetworkError(t *testing.T) {
	svc := NewMigrationService(deadClient(), testLogger())
	_, err := svc.PostFeatureUpgrade(context.Background())
	assert.Error(t, err)
}
