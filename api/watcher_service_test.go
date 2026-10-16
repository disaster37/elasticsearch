package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitWatcherPutWatch(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"_id":"w1","_version":2,"created":false}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.PutWatch(context.Background(), "w1", map[string]any{"trigger": map[string]any{}}, &WatcherPutWatchParams{
		Active:        ptrBool(true),
		IfPrimaryTerm: ptrInt64(1),
		IfSeqNo:       ptrInt64(2),
	})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_watcher/watch/w1", gotPath)
	assert.Equal(t, "true", gotQuery["active"])
	assert.Equal(t, "1", gotQuery["if_primary_term"])
	assert.Equal(t, "2", gotQuery["if_seq_no"])
	assert.Contains(t, gotBody, "trigger")
	assert.Equal(t, "w1", resp.ID)
	assert.Equal(t, int64(2), resp.Version)
	assert.False(t, resp.Created)
}

func TestUnitWatcherPutWatchEmpty(t *testing.T) {
	svc := NewWatcherService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutWatch(context.Background(), "", map[string]any{"x": 1}, &WatcherPutWatchParams{})
	assert.Error(t, err)
}

func TestUnitWatcherPutWatchError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.PutWatch(context.Background(), "w1", map[string]any{"x": 1}, &WatcherPutWatchParams{})
	assert.Error(t, err)
}

func TestUnitWatcherPutWatchNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.PutWatch(context.Background(), "w1", map[string]any{"x": 1}, &WatcherPutWatchParams{})
	assert.Error(t, err)
}

func TestUnitWatcherGetWatch(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"found":true,"_id":"w1","status":{"state":"active"},"watch":{"trigger":{}}}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.GetWatch(context.Background(), "w1")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_watcher/watch/w1", gotPath)
	assert.True(t, resp.Found)
	assert.Equal(t, "w1", resp.ID)
	require.Contains(t, resp.Status, "state")
	assert.Equal(t, "active", resp.Status["state"])
	require.Contains(t, resp.Watch, "trigger")
}

func TestUnitWatcherGetWatchEmpty(t *testing.T) {
	svc := NewWatcherService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.GetWatch(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitWatcherGetWatchError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.GetWatch(context.Background(), "w1")
	assert.Error(t, err)
}

func TestUnitWatcherGetWatchNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.GetWatch(context.Background(), "w1")
	assert.Error(t, err)
}

func TestUnitWatcherDeleteWatch(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"found":true,"_id":"w1","_version":3}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.DeleteWatch(context.Background(), "w1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_watcher/watch/w1", gotPath)
	assert.True(t, resp.Found)
	assert.Equal(t, "w1", resp.ID)
	assert.Equal(t, int64(3), resp.Version)
}

func TestUnitWatcherDeleteWatchEmpty(t *testing.T) {
	svc := NewWatcherService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteWatch(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitWatcherDeleteWatchError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.DeleteWatch(context.Background(), "w1")
	assert.Error(t, err)
}

func TestUnitWatcherDeleteWatchNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.DeleteWatch(context.Background(), "w1")
	assert.Error(t, err)
}

func TestUnitWatcherExecuteWatch(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"w1","watch_record":{"state":"executed"}}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.ExecuteWatch(context.Background(), "w1", map[string]any{"trigger": map[string]any{}}, &WatcherExecuteWatchParams{Debug: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_watcher/watch/w1/_execute", gotPath)
	assert.Equal(t, "true", gotQuery["debug"])
	assert.Contains(t, gotBody, "trigger")
	assert.Equal(t, "w1", resp.ID)
	require.Contains(t, resp.WatchRecord, "state")
	assert.Equal(t, "executed", resp.WatchRecord["state"])
}

func TestUnitWatcherExecuteWatchNoId(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"id":"w1","watch_record":{}}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.ExecuteWatch(context.Background(), "", nil, &WatcherExecuteWatchParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_watcher/watch/_execute", gotPath)
}

func TestUnitWatcherExecuteWatchError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.ExecuteWatch(context.Background(), "w1", map[string]any{"x": 1}, &WatcherExecuteWatchParams{})
	assert.Error(t, err)
}

func TestUnitWatcherExecuteWatchNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.ExecuteWatch(context.Background(), "w1", map[string]any{"x": 1}, &WatcherExecuteWatchParams{})
	assert.Error(t, err)
}

func TestUnitWatcherAckWatch(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"status":{"state":"acked"}}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.AckWatch(context.Background(), "w1", "action1")
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_watcher/watch/w1/_ack/action1", gotPath)
	require.Contains(t, resp.Status, "state")
	assert.Equal(t, "acked", resp.Status["state"])
}

func TestUnitWatcherAckWatchNoAction(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"status":{}}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.AckWatch(context.Background(), "w1", "")
	require.NoError(t, err)
	assert.Equal(t, "/_watcher/watch/w1/_ack", gotPath)
}

func TestUnitWatcherAckWatchEmpty(t *testing.T) {
	svc := NewWatcherService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.AckWatch(context.Background(), "", "action1")
	assert.Error(t, err)
}

func TestUnitWatcherAckWatchError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.AckWatch(context.Background(), "w1", "action1")
	assert.Error(t, err)
}

func TestUnitWatcherAckWatchNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.AckWatch(context.Background(), "w1", "action1")
	assert.Error(t, err)
}

func TestUnitWatcherActivateWatch(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"status":{"state":"active"}}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.ActivateWatch(context.Background(), "w1")
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_watcher/watch/w1/_activate", gotPath)
	require.Contains(t, resp.Status, "state")
	assert.Equal(t, "active", resp.Status["state"])
}

func TestUnitWatcherActivateWatchEmpty(t *testing.T) {
	svc := NewWatcherService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ActivateWatch(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitWatcherActivateWatchError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.ActivateWatch(context.Background(), "w1")
	assert.Error(t, err)
}

func TestUnitWatcherActivateWatchNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.ActivateWatch(context.Background(), "w1")
	assert.Error(t, err)
}

func TestUnitWatcherDeactivateWatch(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"status":{"state":"inactive"}}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.DeactivateWatch(context.Background(), "w1")
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_watcher/watch/w1/_deactivate", gotPath)
	require.Contains(t, resp.Status, "state")
	assert.Equal(t, "inactive", resp.Status["state"])
}

func TestUnitWatcherDeactivateWatchEmpty(t *testing.T) {
	svc := NewWatcherService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeactivateWatch(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitWatcherDeactivateWatchError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.DeactivateWatch(context.Background(), "w1")
	assert.Error(t, err)
}

func TestUnitWatcherDeactivateWatchNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.DeactivateWatch(context.Background(), "w1")
	assert.Error(t, err)
}

func TestUnitWatcherQueryWatches(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"count":1,"watches":[{"id":"w1","watch":{"trigger":{}}}]}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.QueryWatches(context.Background(), map[string]any{"query": map[string]any{}})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_watcher/_query/watches", gotPath)
	assert.Contains(t, gotBody, "query")
	assert.Equal(t, 1, resp.Count)
	require.Len(t, resp.Watches, 1)
	assert.Equal(t, "w1", resp.Watches[0].ID)
	require.Contains(t, resp.Watches[0].Watch, "trigger")
}

func TestUnitWatcherQueryWatchesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.QueryWatches(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitWatcherQueryWatchesNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.QueryWatches(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitWatcherStart(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.Start(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_watcher/_start", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitWatcherStartError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.Start(context.Background())
	assert.Error(t, err)
}

func TestUnitWatcherStartNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.Start(context.Background())
	assert.Error(t, err)
}

func TestUnitWatcherStop(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.Stop(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_watcher/_stop", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitWatcherStopError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.Stop(context.Background())
	assert.Error(t, err)
}

func TestUnitWatcherStopNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.Stop(context.Background())
	assert.Error(t, err)
}

func TestUnitWatcherStats(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"watcher_state":"started","stats":{"count":1},"queues":{"q1":{}},"current_watches":[{"id":"w1"}]}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	resp, err := svc.Stats(context.Background(), []string{"_all"}, &WatcherStatsParams{EmitStacktraces: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_watcher/stats/_all", gotPath)
	assert.Equal(t, "true", gotQuery["emit_stacktraces"])
	assert.Equal(t, "started", resp.WatcherState)
	require.Contains(t, resp.Stats, "count")
	require.Contains(t, resp.Queues, "q1")
	require.Len(t, resp.CurrentWatches, 1)
}

func TestUnitWatcherStatsNoMetrics(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"watcher_state":"started"}`)
	}))
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.Stats(context.Background(), nil, &WatcherStatsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_watcher/stats", gotPath)
}

func TestUnitWatcherStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	_, err := svc.Stats(context.Background(), []string{"_all"}, &WatcherStatsParams{})
	assert.Error(t, err)
}

func TestUnitWatcherStatsNetworkError(t *testing.T) {
	svc := NewWatcherService(deadClient(), testLogger())
	_, err := svc.Stats(context.Background(), []string{"_all"}, &WatcherStatsParams{})
	assert.Error(t, err)
}

func TestUnitWatcherB10Paths(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	svc := NewWatcherService(restyClient(srv), testLogger())
	ctx := context.Background()

	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
		query  map[string]string
	}{
		{"ack_named", func() error { _, err := svc.AckWatch(ctx, "w1", "action1"); return err }, "PUT", "/_watcher/watch/w1/_ack/action1", nil},
		{"ack_no_action", func() error { _, err := svc.AckWatch(ctx, "w1", ""); return err }, "PUT", "/_watcher/watch/w1/_ack", nil},
		{"activate", func() error { _, err := svc.ActivateWatch(ctx, "w1"); return err }, "PUT", "/_watcher/watch/w1/_activate", nil},
		{"deactivate", func() error { _, err := svc.DeactivateWatch(ctx, "w1"); return err }, "PUT", "/_watcher/watch/w1/_deactivate", nil},
		{"put_version", func() error {
			v := int64(3)
			_, err := svc.PutWatch(ctx, "w1", map[string]any{}, &WatcherPutWatchParams{Version: &v})
			return err
		}, "PUT", "/_watcher/watch/w1", map[string]string{"version": "3"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			*got = capturedSpecRequest{}
			assert.NoError(t, tc.call())
			assert.Equal(t, tc.method, got.method)
			assert.Equal(t, tc.path, got.path)
			for k, v := range tc.query {
				assert.Equal(t, v, got.query[k], "query %s", k)
			}
		})
	}
}
