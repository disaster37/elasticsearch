package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitHealthReportGet(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"cluster_name":"my-cluster","status":"green","indicators":{}}`)
	}))
	defer srv.Close()
	svc := NewHealthReportService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), "shards", &HealthReportParams{
		Size:    ptrInt(10),
		Time:    "10s",
		Verbose: ptrBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_health_report/shards", gotPath)
	assert.Equal(t, "10", gotQuery["size"])
	assert.Equal(t, "10s", gotQuery["time"])
	assert.Equal(t, "true", gotQuery["verbose"])
	assert.Equal(t, "my-cluster", resp.ClusterName)
	assert.Equal(t, "green", resp.Status)
}

func TestUnitHealthReportGetAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"cluster_name":"my-cluster","status":"green","indicators":{}}`)
	}))
	defer srv.Close()
	svc := NewHealthReportService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "", nil)
	require.NoError(t, err)
	assert.Equal(t, "/_health_report", gotPath)
}

func TestUnitHealthReportGetNilParams(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"cluster_name":"my-cluster","status":"green","indicators":{}}`)
	}))
	defer srv.Close()
	svc := NewHealthReportService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "shards", nil)
	require.NoError(t, err)
	assert.Equal(t, "/_health_report/shards", gotPath)
}

func TestUnitHealthReportGetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewHealthReportService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "shards", nil)
	assert.Error(t, err)
}

func TestUnitHealthReportGetNetworkError(t *testing.T) {
	svc := NewHealthReportService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), "shards", nil)
	assert.Error(t, err)
}

func TestUnitHealthReportB8Timeout(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	svc := NewHealthReportService(restyClient(srv), testLogger())
	ctx := context.Background()

	_, err := svc.Get(ctx, "", &HealthReportParams{Timeout: "t1"})
	assert.NoError(t, err)
	assert.Equal(t, "/_health_report", got.path)
	assert.Equal(t, "t1", got.query["timeout"])

	*got = capturedSpecRequest{}
	_, err = svc.Get(ctx, "shards_capacity", &HealthReportParams{Timeout: "t2"})
	assert.NoError(t, err)
	assert.Equal(t, "/_health_report/shards_capacity", got.path)
	assert.Equal(t, "t2", got.query["timeout"])
}
