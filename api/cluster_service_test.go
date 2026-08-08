package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitClusterHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/health", r.URL.Path)
		jsonResponse(w, 200, `{"cluster_name":"cluster","status":"green","number_of_nodes":1,"number_of_data_nodes":1,"active_primary_shards":1,"active_shards":1}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.Health(context.Background(), nil, &ClusterHealthParams{})
	assert.NoError(t, err)
	assert.Equal(t, "green", resp.Status)
	assert.Equal(t, 1, resp.NumberOfNodes)
}

func TestUnitClusterHealthWaitForStatus(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		jsonResponse(w, 200, `{"cluster_name":"cluster","status":"yellow"}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	_, err := svc.Health(context.Background(), nil, &ClusterHealthParams{WaitForStatus: WaitForStatusGreen})
	assert.NoError(t, err)
	assert.Contains(t, gotQuery, "wait_for_status=green")
}

func TestUnitClusterGetSettings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/settings", r.URL.Path)
		jsonResponse(w, 200, `{"persistent":{},"transient":{}}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.GetSettings(context.Background(), &ClusterGetSettingsParams{})
	assert.NoError(t, err)
	assert.NotNil(t, resp.Persistent)
}

func TestUnitClusterStats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/stats", r.URL.Path)
		jsonResponse(w, 200, `{"cluster_name":"cluster","cluster_uuid":"uuid","status":"green","indices":{},"nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.Stats(context.Background(), nil, &ClusterStatsParams{})
	assert.NoError(t, err)
	assert.Equal(t, "green", resp.Status)
}

func TestUnitClusterStatsNodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/stats/nodes/node1,node2", r.URL.Path)
		jsonResponse(w, 200, `{"cluster_name":"cluster","status":"green","indices":{},"nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	_, err := svc.Stats(context.Background(), []string{"node1", "node2"}, &ClusterStatsParams{FlatSettings: ptrBool(true)})
	assert.NoError(t, err)
}

func TestUnitClusterState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/state", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"cluster_name":"test","cluster_uuid":"uuid","version":1}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.State(context.Background(), &ClusterStateRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "test", resp.ClusterName)
}

func TestUnitClusterStateMetrics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/state/metrics", r.URL.Path)
		jsonResponse(w, 200, `{"cluster_name":"test"}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	_, err := svc.State(context.Background(), &ClusterStateRequest{Metrics: []string{"metrics"}})
	assert.NoError(t, err)
}

func TestUnitClusterStateMetricsIndices(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/state/metrics/idx", r.URL.Path)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"cluster_name":"test"}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	_, err := svc.State(context.Background(), &ClusterStateRequest{
		Metrics: []string{"metrics"},
		Indices: []string{"idx"},
		Params:  &ClusterStateParams{Local: ptrBool(true), WaitForMetadataVersion: ptrInt64(5)},
	})
	assert.NoError(t, err)
	assert.Equal(t, "true", gotQuery["local"])
	assert.Equal(t, "5", gotQuery["wait_for_metadata_version"])
}

func TestUnitClusterPutSettings(t *testing.T) {
	var gotMethod string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/settings", r.URL.Path)
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true,"transient":{"action.auto_create_index":"true"}}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.PutSettings(context.Background(), map[string]any{"transient": map[string]any{"action.auto_create_index": "true"}}, &ClusterPutSettingsParams{FlatSettings: ptrBool(true)})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.NotEmpty(t, gotBody)
	assert.True(t, resp.Acknowledged)
}

func TestUnitClusterPutSettingsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	_, err := svc.PutSettings(context.Background(), map[string]any{}, &ClusterPutSettingsParams{})
	assert.Error(t, err)
}

func TestUnitClusterAllocationExplain(t *testing.T) {
	var gotMethod string
	var gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/allocation/explain", r.URL.Path)
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"index":"idx","shard":0,"primary":true,"current_state":"started"}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.AllocationExplain(context.Background(), map[string]any{"index": "idx"}, &ClusterAllocationExplainParams{IncludeDiskInfo: ptrBool(true)})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.NotEmpty(t, gotBody)
	assert.Equal(t, "true", gotQuery["include_disk_info"])
	assert.Equal(t, "idx", resp.Index)
}

func TestUnitClusterAllocationExplainNoBody(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		jsonResponse(w, 200, `{"current_state":"started"}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	_, err := svc.AllocationExplain(context.Background(), nil, &ClusterAllocationExplainParams{})
	assert.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
}

func TestUnitClusterReroute(t *testing.T) {
	var gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/reroute", r.URL.Path)
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.Reroute(context.Background(), map[string]any{"commands": []map[string]any{}}, &ClusterRerouteParams{DryRun: ptrBool(true), Metric: []string{"nodes"}})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "true", gotQuery["dry_run"])
	assert.Contains(t, gotQuery["metric"], "nodes")
	assert.True(t, resp.Acknowledged)
}

func TestUnitClusterPendingTasks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/pending_tasks", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"tasks":[]}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.PendingTasks(context.Background(), &ClusterPendingTasksParams{Local: ptrBool(true)})
	assert.NoError(t, err)
	assert.NotNil(t, resp.Tasks)
}

func TestUnitClusterRemoteInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_remote/info", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"remote_cluster":{"connected":true,"mode":"sniff","seeds":["host:9300"],"skip_unavailable":false}}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.RemoteInfo(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, *resp, "remote_cluster")
}

func TestUnitClusterInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/info", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"version":{"number":"8.0.0"}}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.Info(context.Background(), nil)
	assert.NoError(t, err)
	assert.Contains(t, *resp, "version")
}

func TestUnitClusterInfoFeatures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/info/version", r.URL.Path)
		jsonResponse(w, 200, `{"version":{"number":"8.0.0"}}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	_, err := svc.Info(context.Background(), []string{"version"})
	assert.NoError(t, err)
}

func TestUnitClusterPostVotingConfigExclusions(t *testing.T) {
	var gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/voting_config_exclusions", r.URL.Path)
		gotMethod = r.Method
		gotQuery = queryMap(r)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	err := svc.PostVotingConfigExclusions(context.Background(), &ClusterVotingConfigExclusionsParams{NodeNames: []string{"node1"}, Timeout: "30s"})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Contains(t, gotQuery["node_names"], "node1")
	assert.Equal(t, "30s", gotQuery["timeout"])
}

func TestUnitClusterDeleteVotingConfigExclusions(t *testing.T) {
	var gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cluster/voting_config_exclusions", r.URL.Path)
		gotMethod = r.Method
		gotQuery = queryMap(r)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	err := svc.DeleteVotingConfigExclusions(context.Background(), &ClusterDeleteVotingConfigExclusionsParams{WaitForRemoval: ptrBool(false)})
	assert.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "false", gotQuery["wait_for_removal"])
}

func TestUnitClusterGetComponentTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_component_template", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"component_templates":[]}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.GetComponentTemplate(context.Background(), nil, &ClusterGetComponentTemplateParams{Local: ptrBool(true)})
	assert.NoError(t, err)
	assert.NotNil(t, resp.ComponentTemplates)
}

func TestUnitClusterGetComponentTemplateNamed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_component_template/t1,t2", r.URL.Path)
		jsonResponse(w, 200, `{"component_templates":[]}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	_, err := svc.GetComponentTemplate(context.Background(), []string{"t1", "t2"}, &ClusterGetComponentTemplateParams{})
	assert.NoError(t, err)
}

func TestUnitClusterPutComponentTemplate(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.PutComponentTemplate(context.Background(), &PutComponentTemplateRequest{
		Name:   "t1",
		Body:   map[string]any{"template": map[string]any{}},
		Params: &PutComponentTemplateParams{Create: ptrBool(true)},
	})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_component_template/t1", gotPath)
	assert.NotEmpty(t, gotBody)
	assert.True(t, resp.Acknowledged)
}

func TestUnitClusterPutComponentTemplateValidate(t *testing.T) {
	svc := NewClusterService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutComponentTemplate(context.Background(), &PutComponentTemplateRequest{Name: "", Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitClusterExistsComponentTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_component_template/t1", r.URL.Path)
		assert.Equal(t, "HEAD", r.Method)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	ok, err := svc.ExistsComponentTemplate(context.Background(), "t1")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUnitClusterExistsComponentTemplateNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	ok, err := svc.ExistsComponentTemplate(context.Background(), "t1")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestUnitClusterExistsComponentTemplateEmptyName(t *testing.T) {
	svc := NewClusterService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ExistsComponentTemplate(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitClusterDeleteComponentTemplate(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	resp, err := svc.DeleteComponentTemplate(context.Background(), "t1", &ClusterDeleteComponentTemplateParams{Timeout: "30s"})
	assert.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_component_template/t1", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitClusterDeleteComponentTemplateEmptyName(t *testing.T) {
	svc := NewClusterService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteComponentTemplate(context.Background(), "", &ClusterDeleteComponentTemplateParams{})
	assert.Error(t, err)
}
