package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitNodesInfo(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"cluster_name":"cluster","nodes":{"node1":{"name":"node1","transport_address":"127.0.0.1:9300","host":"127.0.0.1","ip":"127.0.0.1","version":"8.0.0","build_hash":"abc","build_type":"docker","build_date":"2020","roles":["master"],"plugins":[{"name":"x-pack","version":"8.0.0","elasticsearch_version":"8.0.0","description":"xpack","has_native_controller":true}]}}}`)
	}))
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	resp, err := svc.Info(context.Background(), &NodesInfoRequest{
		NodeIds: []string{"node1"},
		Metrics: []string{"http"},
		Params:  &NodesInfoParams{FlatSettings: ptrBool(true), Timeout: "10s"},
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_nodes/node1/http", gotPath)
	assert.Equal(t, "true", gotQuery["flat_settings"])
	assert.Equal(t, "10s", gotQuery["timeout"])
	assert.Equal(t, "cluster", resp.ClusterName)
	require.Contains(t, resp.Nodes, "node1")
	n := resp.Nodes["node1"]
	assert.Equal(t, "node1", n.Name)
	assert.Equal(t, "8.0.0", n.Version)
	assert.Equal(t, []string{"master"}, n.Roles)
	require.Len(t, n.Plugins, 1)
	assert.Equal(t, "x-pack", n.Plugins[0].Name)
	assert.True(t, n.Plugins[0].HasNativeController)
}

func TestUnitNodesInfoAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"cluster_name":"c","nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	_, err := svc.Info(context.Background(), &NodesInfoRequest{})
	require.NoError(t, err)
	assert.Equal(t, "/_nodes", gotPath)
}

func TestUnitNodesInfoMetricsOnly(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"cluster_name":"c","nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	_, err := svc.Info(context.Background(), &NodesInfoRequest{Metrics: []string{"jvm", "os"}})
	require.NoError(t, err)
	assert.Equal(t, "/_nodes/jvm,os", gotPath)
}

func TestUnitNodesInfoError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	_, err := svc.Info(context.Background(), &NodesInfoRequest{})
	assert.Error(t, err)
}

func TestUnitNodesInfoNetworkError(t *testing.T) {
	svc := NewNodesService(deadClient(), testLogger())
	_, err := svc.Info(context.Background(), &NodesInfoRequest{})
	assert.Error(t, err)
}

func TestUnitNodesStats(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"cluster_name":"cluster","nodes":{"node1":{"timestamp":100,"name":"node1","transport_address":"127.0.0.1:9300","host":"127.0.0.1","ip":"127.0.0.1","roles":["data"]}}}`)
	}))
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	resp, err := svc.Stats(context.Background(), &NodesStatsRequest{
		NodeIds: []string{"node1"},
		Metrics: []string{"jvm"},
		Params:  &NodesStatsParams{Level: LevelIndices, Timeout: "5s", Fields: []string{"f"}, Groups: []string{"g"}, Types: []string{"t"}, CompletionFields: []string{"c"}, FielddataFields: []string{"fd"}, IncludeSegmentFileSizes: ptrBool(true)},
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_nodes/node1/stats/jvm", gotPath)
	assert.Equal(t, "indices", gotQuery["level"])
	assert.Equal(t, "5s", gotQuery["timeout"])
	assert.Equal(t, "f", gotQuery["fields"])
	assert.Equal(t, "g", gotQuery["groups"])
	assert.Equal(t, "t", gotQuery["types"])
	assert.Equal(t, "c", gotQuery["completion_fields"])
	assert.Equal(t, "fd", gotQuery["fielddata_fields"])
	assert.Equal(t, "true", gotQuery["include_segment_file_sizes"])
	assert.Equal(t, "cluster", resp.ClusterName)
	require.Contains(t, resp.Nodes, "node1")
	assert.Equal(t, int64(100), resp.Nodes["node1"].Timestamp)
	assert.Equal(t, []string{"data"}, resp.Nodes["node1"].Roles)
}

func TestUnitNodesStatsAllPaths(t *testing.T) {
	cases := []struct {
		name   string
		req    NodesStatsRequest
		expect string
	}{
		{"ids+metrics+indexmetrics", NodesStatsRequest{NodeIds: []string{"n"}, Metrics: []string{"m"}, IndexMetrics: []string{"im"}, Params: &NodesStatsParams{}}, "/_nodes/n/stats/m/im"},
		{"ids+metrics", NodesStatsRequest{NodeIds: []string{"n"}, Metrics: []string{"m"}, Params: &NodesStatsParams{}}, "/_nodes/n/stats/m"},
		{"ids", NodesStatsRequest{NodeIds: []string{"n"}, Params: &NodesStatsParams{}}, "/_nodes/n/stats"},
		{"metrics+indexmetrics", NodesStatsRequest{Metrics: []string{"m"}, IndexMetrics: []string{"im"}, Params: &NodesStatsParams{}}, "/_nodes/stats/m/im"},
		{"metrics", NodesStatsRequest{Metrics: []string{"m"}, Params: &NodesStatsParams{}}, "/_nodes/stats/m"},
		{"default", NodesStatsRequest{Params: &NodesStatsParams{}}, "/_nodes/stats"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				jsonResponse(w, 200, `{"cluster_name":"c","nodes":{}}`)
			}))
			defer srv.Close()
			svc := NewNodesService(restyClient(srv), testLogger())
			_, err := svc.Stats(context.Background(), &c.req)
			require.NoError(t, err)
			assert.Equal(t, c.expect, gotPath)
		})
	}
}

func TestUnitNodesStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	_, err := svc.Stats(context.Background(), &NodesStatsRequest{Params: &NodesStatsParams{}})
	assert.Error(t, err)
}

func TestUnitNodesStatsNetworkError(t *testing.T) {
	svc := NewNodesService(deadClient(), testLogger())
	_, err := svc.Stats(context.Background(), &NodesStatsRequest{Params: &NodesStatsParams{}})
	assert.Error(t, err)
}

func TestUnitNodesUsage(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"cluster_name":"cluster","nodes":{"node1":{"timestamp":100,"since":50,"rest_actions":{"indices.get":5}}}}`)
	}))
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	resp, err := svc.Usage(context.Background(), &NodesUsageRequest{
		NodeIds: []string{"node1"},
		Metrics: []string{"rest_actions"},
		Params:  &NodesUsageParams{Timeout: "5s"},
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_nodes/usage/node1/rest_actions", gotPath)
	assert.Equal(t, "5s", gotQuery["timeout"])
	assert.Equal(t, "cluster", resp.ClusterName)
	require.Contains(t, resp.Nodes, "node1")
	assert.Equal(t, int64(100), resp.Nodes["node1"].Timestamp)
	assert.Equal(t, int64(50), resp.Nodes["node1"].Since)
	assert.Equal(t, 5, resp.Nodes["node1"].RestActions["indices.get"])
}

func TestUnitNodesUsageDefault(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"cluster_name":"c","nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	_, err := svc.Usage(context.Background(), &NodesUsageRequest{})
	require.NoError(t, err)
	assert.Equal(t, "/_nodes/usage", gotPath)
}

func TestUnitNodesUsageError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	_, err := svc.Usage(context.Background(), &NodesUsageRequest{})
	assert.Error(t, err)
}

func TestUnitNodesUsageNetworkError(t *testing.T) {
	svc := NewNodesService(deadClient(), testLogger())
	_, err := svc.Usage(context.Background(), &NodesUsageRequest{})
	assert.Error(t, err)
}

func TestUnitNodesHotThreads(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`::: node1\n   thread1`))
	}))
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	out, err := svc.HotThreads(context.Background(), []string{"node1"}, &NodesHotThreadsParams{
		IgnoreIdleThreads: ptrBool(true),
		Interval:          "500ms",
		Snapshots:         ptrInt(10),
		Sort:              "cpu",
		Threads:           ptrInt(3),
		Timeout:           "5s",
		Type:              "cpu",
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_nodes/node1/hot_threads", gotPath)
	assert.Equal(t, "true", gotQuery["ignore_idle_threads"])
	assert.Equal(t, "500ms", gotQuery["interval"])
	assert.Equal(t, "10", gotQuery["snapshots"])
	assert.Equal(t, "cpu", gotQuery["sort"])
	assert.Equal(t, "3", gotQuery["threads"])
	assert.Equal(t, "5s", gotQuery["timeout"])
	assert.Equal(t, "cpu", gotQuery["type"])
	assert.Contains(t, out, "node1")
}

func TestUnitNodesHotThreadsAllNodes(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	_, err := svc.HotThreads(context.Background(), nil, &NodesHotThreadsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_nodes/hot_threads", gotPath)
}

func TestUnitNodesHotThreadsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	_, err := svc.HotThreads(context.Background(), nil, &NodesHotThreadsParams{})
	assert.Error(t, err)
}

func TestUnitNodesHotThreadsNetworkError(t *testing.T) {
	svc := NewNodesService(deadClient(), testLogger())
	_, err := svc.HotThreads(context.Background(), nil, &NodesHotThreadsParams{})
	assert.Error(t, err)
}

func TestUnitNodesReloadSecureSettings(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"cluster_name":"cluster","nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	resp, err := svc.ReloadSecureSettings(context.Background(), []string{"node1"}, map[string]any{"secure_settings": map[string]any{"key": "value"}}, &NodesReloadSecureSettingsParams{Timeout: "10s"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_nodes/node1/reload_secure_settings", gotPath)
	assert.Equal(t, "10s", gotQuery["timeout"])
	assert.Contains(t, gotBody, "secure_settings")
	assert.Equal(t, "cluster", resp.ClusterName)
}

func TestUnitNodesReloadSecureSettingsAllNodes(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"cluster_name":"c","nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	_, err := svc.ReloadSecureSettings(context.Background(), nil, nil, &NodesReloadSecureSettingsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_nodes/reload_secure_settings", gotPath)
}

func TestUnitNodesReloadSecureSettingsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewNodesService(restyClient(srv), testLogger())
	_, err := svc.ReloadSecureSettings(context.Background(), nil, nil, &NodesReloadSecureSettingsParams{})
	assert.Error(t, err)
}

func TestUnitNodesReloadSecureSettingsNetworkError(t *testing.T) {
	svc := NewNodesService(deadClient(), testLogger())
	_, err := svc.ReloadSecureSettings(context.Background(), nil, nil, &NodesReloadSecureSettingsParams{})
	assert.Error(t, err)
}
