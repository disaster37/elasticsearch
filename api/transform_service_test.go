package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitTransformGet(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"count":1,"transforms":[{"id":"t1","description":"d","frequency":"1m","create_time":100}]}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), []string{"t1"}, &TransformGetParams{
		AllowNoMatch:     ptrBool(true),
		ExcludeGenerated: ptrBool(false),
		From:             ptrInt(0),
		Size:             ptrInt(10),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_transform/t1", gotPath)
	assert.Equal(t, "true", gotQuery["allow_no_match"])
	assert.Equal(t, "false", gotQuery["exclude_generated"])
	assert.Equal(t, "0", gotQuery["from"])
	assert.Equal(t, "10", gotQuery["size"])
	assert.Equal(t, 1, resp.Count)
	require.Len(t, resp.Transforms, 1)
	tr := resp.Transforms[0]
	assert.Equal(t, "t1", tr.ID)
	assert.Equal(t, "d", tr.Description)
	assert.Equal(t, "1m", tr.Frequency)
	assert.Equal(t, int64(100), tr.CreateTime)
}

func TestUnitTransformGetAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"count":0,"transforms":[]}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), nil, &TransformGetParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_transform", gotPath)
}

func TestUnitTransformGetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), []string{"t1"}, &TransformGetParams{})
	assert.Error(t, err)
}

func TestUnitTransformGetNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), []string{"t1"}, &TransformGetParams{})
	assert.Error(t, err)
}

func TestUnitTransformGetStats(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"count":1,"transforms":[{"id":"t1","state":"started"}]}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.GetStats(context.Background(), []string{"t1"}, &TransformGetStatsParams{
		AllowNoMatch: ptrBool(true),
		From:         ptrInt(0),
		Size:         ptrInt(10),
		Timeout:      "30s",
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_transform/t1/_stats", gotPath)
	assert.Equal(t, "true", gotQuery["allow_no_match"])
	assert.Equal(t, "0", gotQuery["from"])
	assert.Equal(t, "10", gotQuery["size"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Equal(t, 1, resp.Count)
	require.Len(t, resp.Transforms, 1)
	assert.Equal(t, "t1", resp.Transforms[0].ID)
	assert.Equal(t, "started", resp.Transforms[0].State)
}

func TestUnitTransformGetStatsAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"count":0,"transforms":[]}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.GetStats(context.Background(), nil, &TransformGetStatsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_transform/_stats", gotPath)
}

func TestUnitTransformGetStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.GetStats(context.Background(), []string{"t1"}, &TransformGetStatsParams{})
	assert.Error(t, err)
}

func TestUnitTransformGetStatsNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.GetStats(context.Background(), []string{"t1"}, &TransformGetStatsParams{})
	assert.Error(t, err)
}

func TestUnitTransformPut(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true,"created":true}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.Put(context.Background(), "t1", map[string]any{"source": map[string]any{"index": "src"}}, &TransformPutParams{
		DeferValidation: ptrBool(true),
		Timeout:         "30s",
	})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_transform/t1", gotPath)
	assert.Equal(t, "true", gotQuery["defer_validation"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Contains(t, gotBody, "src")
	assert.True(t, resp.Acknowledged)
	assert.True(t, resp.Created)
}

func TestUnitTransformPutEmpty(t *testing.T) {
	svc := NewTransformService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Put(context.Background(), "", map[string]any{"x": 1}, &TransformPutParams{})
	assert.Error(t, err)
}

func TestUnitTransformPutError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.Put(context.Background(), "t1", map[string]any{"x": 1}, &TransformPutParams{})
	assert.Error(t, err)
}

func TestUnitTransformPutNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.Put(context.Background(), "t1", map[string]any{"x": 1}, &TransformPutParams{})
	assert.Error(t, err)
}

func TestUnitTransformPreview(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"preview":[{"k":"v"}]}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.Preview(context.Background(), "t1", map[string]any{"source": map[string]any{}}, &TransformPreviewParams{Timeout: "30s"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_transform/t1/_preview", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Contains(t, gotBody, "source")
	require.Len(t, resp.Preview, 1)
	assert.Equal(t, "v", resp.Preview[0]["k"])
}

func TestUnitTransformPreviewNoId(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"preview":[]}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.Preview(context.Background(), "", nil, &TransformPreviewParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_transform/_preview", gotPath)
}

func TestUnitTransformPreviewError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.Preview(context.Background(), "t1", map[string]any{"x": 1}, &TransformPreviewParams{})
	assert.Error(t, err)
}

func TestUnitTransformPreviewNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.Preview(context.Background(), "t1", map[string]any{"x": 1}, &TransformPreviewParams{})
	assert.Error(t, err)
}

func TestUnitTransformDelete(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background(), "t1", &TransformDeleteParams{Force: ptrBool(true), Timeout: "30s"})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_transform/t1", gotPath)
	assert.Equal(t, "true", gotQuery["force"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitTransformDeleteEmpty(t *testing.T) {
	svc := NewTransformService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Delete(context.Background(), "", &TransformDeleteParams{})
	assert.Error(t, err)
}

func TestUnitTransformDeleteError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.Delete(context.Background(), "t1", &TransformDeleteParams{})
	assert.Error(t, err)
}

func TestUnitTransformDeleteNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.Delete(context.Background(), "t1", &TransformDeleteParams{})
	assert.Error(t, err)
}

func TestUnitTransformStart(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.Start(context.Background(), "t1", &TransformStartParams{Timeout: "30s"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_transform/t1/_start", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitTransformStartEmpty(t *testing.T) {
	svc := NewTransformService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Start(context.Background(), "", &TransformStartParams{})
	assert.Error(t, err)
}

func TestUnitTransformStartError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.Start(context.Background(), "t1", &TransformStartParams{})
	assert.Error(t, err)
}

func TestUnitTransformStartNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.Start(context.Background(), "t1", &TransformStartParams{})
	assert.Error(t, err)
}

func TestUnitTransformStop(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.Stop(context.Background(), "t1", &TransformStopParams{
		AllowNoMatch:      ptrBool(true),
		Force:             ptrBool(true),
		Timeout:           "30s",
		WaitForCheckpoint: ptrBool(true),
		WaitForCompletion: ptrBool(false),
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_transform/t1/_stop", gotPath)
	assert.Equal(t, "true", gotQuery["allow_no_match"])
	assert.Equal(t, "true", gotQuery["force"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Equal(t, "true", gotQuery["wait_for_checkpoint"])
	assert.Equal(t, "false", gotQuery["wait_for_completion"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitTransformStopEmpty(t *testing.T) {
	svc := NewTransformService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Stop(context.Background(), "", &TransformStopParams{})
	assert.Error(t, err)
}

func TestUnitTransformStopError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.Stop(context.Background(), "t1", &TransformStopParams{})
	assert.Error(t, err)
}

func TestUnitTransformStopNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.Stop(context.Background(), "t1", &TransformStopParams{})
	assert.Error(t, err)
}

func TestUnitTransformReset(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.Reset(context.Background(), "t1", &TransformResetParams{Force: ptrBool(true), Timeout: "30s"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_transform/t1/_reset", gotPath)
	assert.Equal(t, "true", gotQuery["force"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitTransformResetEmpty(t *testing.T) {
	svc := NewTransformService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Reset(context.Background(), "", &TransformResetParams{})
	assert.Error(t, err)
}

func TestUnitTransformResetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.Reset(context.Background(), "t1", &TransformResetParams{})
	assert.Error(t, err)
}

func TestUnitTransformResetNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.Reset(context.Background(), "t1", &TransformResetParams{})
	assert.Error(t, err)
}

func TestUnitTransformScheduleNow(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.ScheduleNow(context.Background(), "t1")
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_transform/t1/_schedule_now", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitTransformScheduleNowEmpty(t *testing.T) {
	svc := NewTransformService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ScheduleNow(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitTransformScheduleNowError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.ScheduleNow(context.Background(), "t1")
	assert.Error(t, err)
}

func TestUnitTransformScheduleNowNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.ScheduleNow(context.Background(), "t1")
	assert.Error(t, err)
}

func TestUnitTransformUpgrade(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"needs_update":true,"dry_run":true,"updated":["t1"],"no_action":["t2"]}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.Upgrade(context.Background(), &TransformUpgradeParams{DryRun: ptrBool(true), Timeout: "30s"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_transform/_upgrade", gotPath)
	assert.Equal(t, "true", gotQuery["dry_run"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.NeedsUpdate)
	assert.True(t, resp.DryRun)
	assert.Equal(t, []string{"t1"}, resp.Updated)
	assert.Equal(t, []string{"t2"}, resp.NoAction)
}

func TestUnitTransformUpgradeError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.Upgrade(context.Background(), &TransformUpgradeParams{})
	assert.Error(t, err)
}

func TestUnitTransformUpgradeNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.Upgrade(context.Background(), &TransformUpgradeParams{})
	assert.Error(t, err)
}

func TestUnitTransformGetNodeStats(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"nodes":{"node1":{"transforms":{"t1":{"count":1}}}}}`)
	}))
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	resp, err := svc.GetNodeStats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_transform/_node_stats", gotPath)
	require.Contains(t, resp.Nodes, "node1")
	require.Contains(t, resp.Nodes["node1"].Transforms, "t1")
}

func TestUnitTransformGetNodeStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTransformService(restyClient(srv), testLogger())
	_, err := svc.GetNodeStats(context.Background())
	assert.Error(t, err)
}

func TestUnitTransformGetNodeStatsNetworkError(t *testing.T) {
	svc := NewTransformService(deadClient(), testLogger())
	_, err := svc.GetNodeStats(context.Background())
	assert.Error(t, err)
}
