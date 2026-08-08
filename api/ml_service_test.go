package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// PutJob
// ---------------------------------------------------------------------------

func TestUnitMlPutJob(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"job_id":"j1","description":"d"}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.PutJob(context.Background(), "j1", map[string]any{"description": "d"})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_ml/anomaly_detectors/j1", gotPath)
	assert.Contains(t, gotBody, "d")
	assert.Equal(t, "j1", resp.JobId)
	assert.Equal(t, "d", resp.Description)
}

func TestUnitMlPutJobEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutJob(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitMlPutJobError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.PutJob(context.Background(), "j1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitMlPutJobNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.PutJob(context.Background(), "j1", map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetJobs
// ---------------------------------------------------------------------------

func TestUnitMlGetJobs(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"count":1,"jobs":[{"job_id":"j1","description":"d"}]}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.GetJobs(context.Background(), []string{"j1"}, &MlGetJobsParams{
		AllowNoMatch:     ptrBool(true),
		ExcludeGenerated: ptrBool(false),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ml/anomaly_detectors/j1", gotPath)
	assert.Equal(t, "true", gotQuery["allow_no_match"])
	assert.Equal(t, "false", gotQuery["exclude_generated"])
	assert.Equal(t, 1, resp.Count)
	require.Len(t, resp.Jobs, 1)
	assert.Equal(t, "j1", resp.Jobs[0].JobId)
	assert.Equal(t, "d", resp.Jobs[0].Description)
}

func TestUnitMlGetJobsAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"count":0,"jobs":[]}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.GetJobs(context.Background(), nil, &MlGetJobsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_ml/anomaly_detectors", gotPath)
}

func TestUnitMlGetJobsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.GetJobs(context.Background(), []string{"j1"}, &MlGetJobsParams{})
	assert.Error(t, err)
}

func TestUnitMlGetJobsNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.GetJobs(context.Background(), []string{"j1"}, &MlGetJobsParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// DeleteJob
// ---------------------------------------------------------------------------

func TestUnitMlDeleteJob(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.DeleteJob(context.Background(), "j1", &MlDeleteJobParams{
		Force:             ptrBool(true),
		WaitForCompletion: ptrBool(false),
	})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_ml/anomaly_detectors/j1", gotPath)
	assert.Equal(t, "true", gotQuery["force"])
	assert.Equal(t, "false", gotQuery["wait_for_completion"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitMlDeleteJobEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteJob(context.Background(), "", &MlDeleteJobParams{})
	assert.Error(t, err)
}

func TestUnitMlDeleteJobError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.DeleteJob(context.Background(), "j1", &MlDeleteJobParams{})
	assert.Error(t, err)
}

func TestUnitMlDeleteJobNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.DeleteJob(context.Background(), "j1", &MlDeleteJobParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// OpenJob
// ---------------------------------------------------------------------------

func TestUnitMlOpenJob(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"opened":true}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.OpenJob(context.Background(), "j1", &MlOpenJobParams{Timeout: "30s"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ml/anomaly_detectors/j1/_open", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Opened)
}

func TestUnitMlOpenJobEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.OpenJob(context.Background(), "", &MlOpenJobParams{})
	assert.Error(t, err)
}

func TestUnitMlOpenJobError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.OpenJob(context.Background(), "j1", &MlOpenJobParams{})
	assert.Error(t, err)
}

func TestUnitMlOpenJobNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.OpenJob(context.Background(), "j1", &MlOpenJobParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// CloseJob
// ---------------------------------------------------------------------------

func TestUnitMlCloseJob(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"closed":true}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.CloseJob(context.Background(), "j1", &MlCloseJobParams{
		AllowNoMatch: ptrBool(true),
		Force:        ptrBool(false),
		Timeout:      "30s",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ml/anomaly_detectors/j1/_close", gotPath)
	assert.Equal(t, "true", gotQuery["allow_no_match"])
	assert.Equal(t, "false", gotQuery["force"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Closed)
}

func TestUnitMlCloseJobEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.CloseJob(context.Background(), "", &MlCloseJobParams{})
	assert.Error(t, err)
}

func TestUnitMlCloseJobError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.CloseJob(context.Background(), "j1", &MlCloseJobParams{})
	assert.Error(t, err)
}

func TestUnitMlCloseJobNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.CloseJob(context.Background(), "j1", &MlCloseJobParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ResetJob
// ---------------------------------------------------------------------------

func TestUnitMlResetJob(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.ResetJob(context.Background(), "j1", &MlResetJobParams{WaitForCompletion: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ml/anomaly_detectors/j1/_reset", gotPath)
	assert.Equal(t, "true", gotQuery["wait_for_completion"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitMlResetJobEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ResetJob(context.Background(), "", &MlResetJobParams{})
	assert.Error(t, err)
}

func TestUnitMlResetJobError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.ResetJob(context.Background(), "j1", &MlResetJobParams{})
	assert.Error(t, err)
}

func TestUnitMlResetJobNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.ResetJob(context.Background(), "j1", &MlResetJobParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// FlushJob
// ---------------------------------------------------------------------------

func TestUnitMlFlushJob(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"flushed":true,"last":"2020-01-01"}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.FlushJob(context.Background(), "j1", map[string]any{"calc_interim": true}, &MlFlushJobParams{
		AdvanceTime: "1h",
		CalcInterim: ptrBool(true),
		End:         "2020-01-02",
		SkipTime:    "2020-01-01",
		Start:       "2020-01-01",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ml/anomaly_detectors/j1/_flush", gotPath)
	assert.Equal(t, "1h", gotQuery["advance_time"])
	assert.Equal(t, "true", gotQuery["calc_interim"])
	assert.Equal(t, "2020-01-02", gotQuery["end"])
	assert.Equal(t, "2020-01-01", gotQuery["skip_time"])
	assert.Equal(t, "2020-01-01", gotQuery["start"])
	assert.Contains(t, gotBody, "calc_interim")
	assert.True(t, resp.Flushed)
	assert.Equal(t, "2020-01-01", resp.Last)
}

func TestUnitMlFlushJobEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.FlushJob(context.Background(), "", map[string]any{"x": 1}, &MlFlushJobParams{})
	assert.Error(t, err)
}

func TestUnitMlFlushJobError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.FlushJob(context.Background(), "j1", map[string]any{"x": 1}, &MlFlushJobParams{})
	assert.Error(t, err)
}

func TestUnitMlFlushJobNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.FlushJob(context.Background(), "j1", map[string]any{"x": 1}, &MlFlushJobParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// PutDatafeed
// ---------------------------------------------------------------------------

func TestUnitMlPutDatafeed(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"datafeed_id":"df1","job_id":"j1"}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.PutDatafeed(context.Background(), "df1", map[string]any{"job_id": "j1"})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_ml/datafeeds/df1", gotPath)
	assert.Contains(t, gotBody, "j1")
	assert.Equal(t, "df1", resp.DatafeedId)
	assert.Equal(t, "j1", resp.JobId)
}

func TestUnitMlPutDatafeedEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutDatafeed(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitMlPutDatafeedError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.PutDatafeed(context.Background(), "df1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitMlPutDatafeedNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.PutDatafeed(context.Background(), "df1", map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetDatafeeds
// ---------------------------------------------------------------------------

func TestUnitMlGetDatafeeds(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"count":1,"datafeeds":[{"datafeed_id":"df1","job_id":"j1"}]}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.GetDatafeeds(context.Background(), []string{"df1"}, &MlGetDatafeedsParams{
		AllowNoMatch:     ptrBool(true),
		ExcludeGenerated: ptrBool(false),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ml/datafeeds/df1", gotPath)
	assert.Equal(t, "true", gotQuery["allow_no_match"])
	assert.Equal(t, "false", gotQuery["exclude_generated"])
	assert.Equal(t, 1, resp.Count)
	require.Len(t, resp.Datafeeds, 1)
	assert.Equal(t, "df1", resp.Datafeeds[0].DatafeedId)
	assert.Equal(t, "j1", resp.Datafeeds[0].JobId)
}

func TestUnitMlGetDatafeedsAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"count":0,"datafeeds":[]}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.GetDatafeeds(context.Background(), nil, &MlGetDatafeedsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_ml/datafeeds", gotPath)
}

func TestUnitMlGetDatafeedsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.GetDatafeeds(context.Background(), []string{"df1"}, &MlGetDatafeedsParams{})
	assert.Error(t, err)
}

func TestUnitMlGetDatafeedsNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.GetDatafeeds(context.Background(), []string{"df1"}, &MlGetDatafeedsParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// DeleteDatafeed
// ---------------------------------------------------------------------------

func TestUnitMlDeleteDatafeed(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.DeleteDatafeed(context.Background(), "df1", &MlDeleteDatafeedParams{Force: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_ml/datafeeds/df1", gotPath)
	assert.Equal(t, "true", gotQuery["force"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitMlDeleteDatafeedEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteDatafeed(context.Background(), "", &MlDeleteDatafeedParams{})
	assert.Error(t, err)
}

func TestUnitMlDeleteDatafeedError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.DeleteDatafeed(context.Background(), "df1", &MlDeleteDatafeedParams{})
	assert.Error(t, err)
}

func TestUnitMlDeleteDatafeedNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.DeleteDatafeed(context.Background(), "df1", &MlDeleteDatafeedParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// StartDatafeed
// ---------------------------------------------------------------------------

func TestUnitMlStartDatafeed(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"started":true}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.StartDatafeed(context.Background(), "df1", map[string]any{"start": "2020-01-01"}, &MlStartDatafeedParams{
		End:     "2020-01-02",
		Start:   "2020-01-01",
		Timeout: "30s",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ml/datafeeds/df1/_start", gotPath)
	assert.Equal(t, "2020-01-02", gotQuery["end"])
	assert.Equal(t, "2020-01-01", gotQuery["start"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Contains(t, gotBody, "start")
	assert.True(t, resp.Started)
}

func TestUnitMlStartDatafeedEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.StartDatafeed(context.Background(), "", map[string]any{"x": 1}, &MlStartDatafeedParams{})
	assert.Error(t, err)
}

func TestUnitMlStartDatafeedError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.StartDatafeed(context.Background(), "df1", map[string]any{"x": 1}, &MlStartDatafeedParams{})
	assert.Error(t, err)
}

func TestUnitMlStartDatafeedNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.StartDatafeed(context.Background(), "df1", map[string]any{"x": 1}, &MlStartDatafeedParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// StopDatafeed
// ---------------------------------------------------------------------------

func TestUnitMlStopDatafeed(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"stopped":true}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.StopDatafeed(context.Background(), "df1", &MlStopDatafeedParams{
		AllowNoMatch: ptrBool(true),
		Force:        ptrBool(false),
		Timeout:      "30s",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ml/datafeeds/df1/_stop", gotPath)
	assert.Equal(t, "true", gotQuery["allow_no_match"])
	assert.Equal(t, "false", gotQuery["force"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Stopped)
}

func TestUnitMlStopDatafeedEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.StopDatafeed(context.Background(), "", &MlStopDatafeedParams{})
	assert.Error(t, err)
}

func TestUnitMlStopDatafeedError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.StopDatafeed(context.Background(), "df1", &MlStopDatafeedParams{})
	assert.Error(t, err)
}

func TestUnitMlStopDatafeedNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.StopDatafeed(context.Background(), "df1", &MlStopDatafeedParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetFilters
// ---------------------------------------------------------------------------

func TestUnitMlGetFilters(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"count":1,"filters":[{"filter_id":"f1","items":["a"],"description":"d"}]}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.GetFilters(context.Background(), "f1", &MlGetFiltersParams{
		From: ptrInt(0),
		Size: ptrInt(10),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ml/filters/f1", gotPath)
	assert.Equal(t, "0", gotQuery["from"])
	assert.Equal(t, "10", gotQuery["size"])
	assert.Equal(t, 1, resp.Count)
	require.Len(t, resp.Filters, 1)
	assert.Equal(t, "f1", resp.Filters[0].FilterId)
	assert.Equal(t, []string{"a"}, resp.Filters[0].Items)
	assert.Equal(t, "d", resp.Filters[0].Description)
}

func TestUnitMlGetFiltersAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"count":0,"filters":[]}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.GetFilters(context.Background(), "", &MlGetFiltersParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_ml/filters", gotPath)
}

func TestUnitMlGetFiltersError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.GetFilters(context.Background(), "f1", &MlGetFiltersParams{})
	assert.Error(t, err)
}

func TestUnitMlGetFiltersNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.GetFilters(context.Background(), "f1", &MlGetFiltersParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetTrainedModels
// ---------------------------------------------------------------------------

func TestUnitMlGetTrainedModels(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"count":1,"trained_model_configs":[{"model_id":"m1","type":"regression"}]}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	resp, err := svc.GetTrainedModels(context.Background(), []string{"m1"}, &MlGetTrainedModelsParams{
		AllowNoMatch:         ptrBool(true),
		DecompressDefinition: ptrBool(false),
		ExcludeGenerated:     ptrBool(true),
		From:                 ptrInt(0),
		Include:              "definition",
		Size:                 ptrInt(10),
		Tags:                 []string{"a", "b"},
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ml/trained_models/m1", gotPath)
	assert.Equal(t, "true", gotQuery["allow_no_match"])
	assert.Equal(t, "false", gotQuery["decompress_definition"])
	assert.Equal(t, "true", gotQuery["exclude_generated"])
	assert.Equal(t, "0", gotQuery["from"])
	assert.Equal(t, "definition", gotQuery["include"])
	assert.Equal(t, "10", gotQuery["size"])
	assert.Equal(t, "a,b", gotQuery["tags"])
	assert.Equal(t, 1, resp.Count)
	require.Len(t, resp.TrainedModels, 1)
	assert.Equal(t, "m1", resp.TrainedModels[0].ModelId)
	assert.Equal(t, "regression", resp.TrainedModels[0].Type)
}

func TestUnitMlGetTrainedModelsAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"count":0,"trained_model_configs":[]}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.GetTrainedModels(context.Background(), nil, &MlGetTrainedModelsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_ml/trained_models", gotPath)
}

func TestUnitMlGetTrainedModelsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.GetTrainedModels(context.Background(), []string{"m1"}, &MlGetTrainedModelsParams{})
	assert.Error(t, err)
}

func TestUnitMlGetTrainedModelsNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.GetTrainedModels(context.Background(), []string{"m1"}, &MlGetTrainedModelsParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// InferTrainedModel
// ---------------------------------------------------------------------------

func TestUnitMlInferTrainedModel(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"results":[{"predicted_value":"x"}]}`)
	}))
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	raw, err := svc.InferTrainedModel(context.Background(), "m1", map[string]any{"docs": []any{}}, &MlInferTrainedModelParams{Timeout: "30s"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ml/trained_models/m1/_infer", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Contains(t, gotBody, "docs")
	assert.Contains(t, string(raw), "predicted_value")
}

func TestUnitMlInferTrainedModelEmpty(t *testing.T) {
	svc := NewMlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.InferTrainedModel(context.Background(), "", map[string]any{"x": 1}, &MlInferTrainedModelParams{})
	assert.Error(t, err)
}

func TestUnitMlInferTrainedModelError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewMlService(restyClient(srv), testLogger())
	_, err := svc.InferTrainedModel(context.Background(), "m1", map[string]any{"x": 1}, &MlInferTrainedModelParams{})
	assert.Error(t, err)
}

func TestUnitMlInferTrainedModelNetworkError(t *testing.T) {
	svc := NewMlService(deadClient(), testLogger())
	_, err := svc.InferTrainedModel(context.Background(), "m1", map[string]any{"x": 1}, &MlInferTrainedModelParams{})
	assert.Error(t, err)
}
