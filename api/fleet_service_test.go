package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitFleetGlobalCheckpoints(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"global_checkpoints":[0,100],"timed_out":false}`)
	}))
	defer srv.Close()
	svc := NewFleetService(restyClient(srv), testLogger())
	resp, err := svc.GlobalCheckpoints(context.Background(), "my-index", &FleetGlobalCheckpointsParams{
		Checkpoints:    []string{"0", "100"},
		WaitForAdvance:  ptrBool(true),
		WaitForIndex:    ptrBool(false),
		Timeout:         "10s",
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/my-index/_fleet/global_checkpoints", gotPath)
	assert.Equal(t, "0,100", gotQuery["checkpoints"])
	assert.Equal(t, "true", gotQuery["wait_for_advance"])
	assert.Equal(t, "false", gotQuery["wait_for_index"])
	assert.Equal(t, "10s", gotQuery["timeout"])
	assert.Equal(t, []int64{0, 100}, resp.GlobalCheckpoints)
	assert.False(t, resp.TimedOut)
}

func TestUnitFleetGlobalCheckpointsNilParams(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"global_checkpoints":[],"timed_out":false}`)
	}))
	defer srv.Close()
	svc := NewFleetService(restyClient(srv), testLogger())
	_, err := svc.GlobalCheckpoints(context.Background(), "my-index", nil)
	require.NoError(t, err)
	assert.Equal(t, "/my-index/_fleet/global_checkpoints", gotPath)
}

func TestUnitFleetGlobalCheckpointsEmptyIndex(t *testing.T) {
	svc := NewFleetService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.GlobalCheckpoints(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestUnitFleetGlobalCheckpointsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewFleetService(restyClient(srv), testLogger())
	_, err := svc.GlobalCheckpoints(context.Background(), "my-index", nil)
	assert.Error(t, err)
}

func TestUnitFleetGlobalCheckpointsNetworkError(t *testing.T) {
	svc := NewFleetService(deadClient(), testLogger())
	_, err := svc.GlobalCheckpoints(context.Background(), "my-index", nil)
	assert.Error(t, err)
}
