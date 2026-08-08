package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSlmGetLifecycle(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"policy1":{"version":1,"modified_date":"2020","modified_date_millis":100,"name":"policy1","uuid":"u1","snapshot_name":"snap1","next_execution_millis":200}}`)
	}))
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	result, err := svc.GetLifecycle(context.Background(), []string{"policy1"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_slm/policy/policy1", gotPath)
	require.Contains(t, result, "policy1")
	p := result["policy1"]
	require.NotNil(t, p)
	assert.Equal(t, 1, p.Version)
	assert.Equal(t, "2020", p.ModifiedDate)
	assert.Equal(t, int64(100), p.ModifiedDateMillis)
	assert.Equal(t, "policy1", p.Name)
	assert.Equal(t, "u1", p.Uuid)
	assert.Equal(t, "snap1", p.SnapshotName)
	assert.Equal(t, int64(200), p.NextExecutionMillis)
}

func TestUnitSlmGetLifecycleAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	_, err := svc.GetLifecycle(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_slm/policy", gotPath)
}

func TestUnitSlmGetLifecycleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	_, err := svc.GetLifecycle(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSlmGetLifecycleNetworkError(t *testing.T) {
	svc := NewSlmService(deadClient(), testLogger())
	_, err := svc.GetLifecycle(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSlmPutLifecycle(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	resp, err := svc.PutLifecycle(context.Background(), "policy1", map[string]any{"schedule": "daily"})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_slm/policy/policy1", gotPath)
	assert.Contains(t, gotBody, "daily")
	assert.True(t, resp.Acknowledged)
}

func TestUnitSlmPutLifecycleEmpty(t *testing.T) {
	svc := NewSlmService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutLifecycle(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSlmPutLifecycleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	_, err := svc.PutLifecycle(context.Background(), "policy1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSlmPutLifecycleNetworkError(t *testing.T) {
	svc := NewSlmService(deadClient(), testLogger())
	_, err := svc.PutLifecycle(context.Background(), "policy1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSlmDeleteLifecycle(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	resp, err := svc.DeleteLifecycle(context.Background(), "policy1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_slm/policy/policy1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSlmDeleteLifecycleEmpty(t *testing.T) {
	svc := NewSlmService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteLifecycle(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSlmDeleteLifecycleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	_, err := svc.DeleteLifecycle(context.Background(), "policy1")
	assert.Error(t, err)
}

func TestUnitSlmDeleteLifecycleNetworkError(t *testing.T) {
	svc := NewSlmService(deadClient(), testLogger())
	_, err := svc.DeleteLifecycle(context.Background(), "policy1")
	assert.Error(t, err)
}

func TestUnitSlmExecuteLifecycle(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"snapshot_name":"snap1"}`)
	}))
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	resp, err := svc.ExecuteLifecycle(context.Background(), "policy1")
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_slm/policy/policy1/_execute", gotPath)
	assert.Equal(t, "snap1", resp.SnapshotName)
}

func TestUnitSlmExecuteLifecycleEmpty(t *testing.T) {
	svc := NewSlmService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ExecuteLifecycle(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSlmExecuteLifecycleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	_, err := svc.ExecuteLifecycle(context.Background(), "policy1")
	assert.Error(t, err)
}

func TestUnitSlmExecuteLifecycleNetworkError(t *testing.T) {
	svc := NewSlmService(deadClient(), testLogger())
	_, err := svc.ExecuteLifecycle(context.Background(), "policy1")
	assert.Error(t, err)
}

func TestUnitSlmExecuteRetention(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	resp, err := svc.ExecuteRetention(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_slm/_execute_retention", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSlmExecuteRetentionError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	_, err := svc.ExecuteRetention(context.Background())
	assert.Error(t, err)
}

func TestUnitSlmExecuteRetentionNetworkError(t *testing.T) {
	svc := NewSlmService(deadClient(), testLogger())
	_, err := svc.ExecuteRetention(context.Background())
	assert.Error(t, err)
}

func TestUnitSlmGetStats(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"retention_runs":1,"retention_failed":2,"retention_timed_out":3,"retention_deletion_time":"1s","retention_deletion_time_millis":100,"total_snapshots_taken":10,"total_snapshots_failed":1,"total_snapshot_deletions":5,"total_snapshot_deletion_failures":0,"policy_stats":[{"policy":"policy1"}]}`)
	}))
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	resp, err := svc.GetStats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_slm/stats", gotPath)
	assert.Equal(t, int64(1), resp.RetentionRuns)
	assert.Equal(t, int64(2), resp.RetentionFailed)
	assert.Equal(t, int64(3), resp.RetentionTimedOut)
	assert.Equal(t, "1s", resp.RetentionDeletionTime)
	assert.Equal(t, int64(100), resp.RetentionDeletionTimeMillis)
	assert.Equal(t, int64(10), resp.TotalSnapshotsTaken)
	assert.Equal(t, int64(1), resp.TotalSnapshotsFailed)
	assert.Equal(t, int64(5), resp.TotalSnapshotDeletions)
	assert.Equal(t, int64(0), resp.TotalSnapshotDeletionFailures)
	require.Len(t, resp.PolicyStats, 1)
	assert.Equal(t, "policy1", resp.PolicyStats[0]["policy"])
}

func TestUnitSlmGetStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	_, err := svc.GetStats(context.Background())
	assert.Error(t, err)
}

func TestUnitSlmGetStatsNetworkError(t *testing.T) {
	svc := NewSlmService(deadClient(), testLogger())
	_, err := svc.GetStats(context.Background())
	assert.Error(t, err)
}

func TestUnitSlmGetStatus(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"operation_mode":"RUNNING"}`)
	}))
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	resp, err := svc.GetStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_slm/status", gotPath)
	assert.Equal(t, "RUNNING", resp.OperationMode)
}

func TestUnitSlmGetStatusError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	_, err := svc.GetStatus(context.Background())
	assert.Error(t, err)
}

func TestUnitSlmGetStatusNetworkError(t *testing.T) {
	svc := NewSlmService(deadClient(), testLogger())
	_, err := svc.GetStatus(context.Background())
	assert.Error(t, err)
}

func TestUnitSlmStart(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	resp, err := svc.Start(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_slm/start", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSlmStartError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	_, err := svc.Start(context.Background())
	assert.Error(t, err)
}

func TestUnitSlmStartNetworkError(t *testing.T) {
	svc := NewSlmService(deadClient(), testLogger())
	_, err := svc.Start(context.Background())
	assert.Error(t, err)
}

func TestUnitSlmStop(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	resp, err := svc.Stop(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_slm/stop", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSlmStopError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSlmService(restyClient(srv), testLogger())
	_, err := svc.Stop(context.Background())
	assert.Error(t, err)
}

func TestUnitSlmStopNetworkError(t *testing.T) {
	svc := NewSlmService(deadClient(), testLogger())
	_, err := svc.Stop(context.Background())
	assert.Error(t, err)
}
