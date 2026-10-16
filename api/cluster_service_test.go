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
		assert.Equal(t, "/_info/_all", r.URL.Path)
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
		assert.Equal(t, "/_info/version", r.URL.Path)
		jsonResponse(w, 200, `{"version":{"number":"8.0.0"}}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	_, err := svc.Info(context.Background(), []string{"version"})
	assert.NoError(t, err)
}

func TestUnitClusterInfoMultiTarget(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	_, err := svc.Info(context.Background(), []string{"version", "http"})
	assert.NoError(t, err)
	assert.Equal(t, "/_info/version,http", gotPath)
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

type customCluster struct{ ClusterService }

func TestClusterSpecExtensionAdapters(t *testing.T) {
	svc := NewClusterService(nil, testLogger())
	if _, err := AsClusterServiceSpecExtensions(svc); err != nil {
		t.Fatalf("cluster default: %v", err)
	}
	if _, err := AsClusterServiceSpecExtensions(nil); err == nil {
		t.Error("cluster nil: want error")
	}
	var typedNil ClusterService = (*DefaultClusterService)(nil)
	if _, err := AsClusterServiceSpecExtensions(typedNil); err == nil {
		t.Error("cluster typed-nil: want error")
	}
	if _, err := AsClusterServiceSpecExtensions(customCluster{svc}); err == nil {
		t.Error("cluster custom: want error")
	}
}

func TestUnitClusterB5CarrierTransport(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	svc := NewClusterService(restyClient(srv), testLogger())
	ext, err := AsClusterServiceSpecExtensions(svc)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	body := map[string]any{"current_node": "n1"}

	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
		query  map[string]string
	}{
		{"health", func() error {
			_, err := svc.Health(ctx, []string{"idx"}, &ClusterHealthParams{MasterTimeout: "m1", ExpandWildcards: ExpandWildcardOpen})
			return err
		}, "GET", "/_cluster/health/idx", map[string]string{"master_timeout": "m1", "expand_wildcards": "open"}},
		{"state", func() error {
			_, err := svc.State(ctx, &ClusterStateRequest{Metrics: []string{"metadata"}, Indices: []string{"idx"}, Params: &ClusterStateParams{MasterTimeout: "m1"}})
			return err
		}, "GET", "/_cluster/state/metadata/idx", map[string]string{"master_timeout": "m1"}},
		{"stats", func() error {
			_, err := svc.Stats(ctx, []string{"n1"}, &ClusterStatsParams{IncludeRemotes: ptrBool(true)})
			return err
		}, "GET", "/_cluster/stats/nodes/n1", map[string]string{"include_remotes": "true"}},
		{"get_settings", func() error {
			_, err := svc.GetSettings(ctx, &ClusterGetSettingsParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_cluster/settings", map[string]string{"master_timeout": "m1"}},
		{"put_settings", func() error {
			_, err := svc.PutSettings(ctx, body, &ClusterPutSettingsParams{MasterTimeout: "m1"})
			return err
		}, "PUT", "/_cluster/settings", map[string]string{"master_timeout": "m1"}},
		{"allocation_explain", func() error {
			_, err := svc.AllocationExplain(ctx, body, &ClusterAllocationExplainParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/_cluster/allocation/explain", map[string]string{"master_timeout": "m1"}},
		{"reroute", func() error {
			_, err := svc.Reroute(ctx, body, &ClusterRerouteParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/_cluster/reroute", map[string]string{"master_timeout": "m1"}},
		{"pending_tasks", func() error {
			_, err := svc.PendingTasks(ctx, &ClusterPendingTasksParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_cluster/pending_tasks", map[string]string{"master_timeout": "m1"}},
		{"post_voting_config_exclusions", func() error {
			return svc.PostVotingConfigExclusions(ctx, &ClusterVotingConfigExclusionsParams{MasterTimeout: "m1"})
		}, "POST", "/_cluster/voting_config_exclusions", map[string]string{"master_timeout": "m1"}},
		{"delete_voting_config_exclusions", func() error {
			return svc.DeleteVotingConfigExclusions(ctx, &ClusterDeleteVotingConfigExclusionsParams{MasterTimeout: "m1"})
		}, "DELETE", "/_cluster/voting_config_exclusions", map[string]string{"master_timeout": "m1"}},
		{"get_component_template", func() error {
			_, err := svc.GetComponentTemplate(ctx, []string{"c1"}, &ClusterGetComponentTemplateParams{MasterTimeout: "m1", IncludeDefaults: ptrBool(true)})
			return err
		}, "GET", "/_component_template/c1", map[string]string{"master_timeout": "m1", "include_defaults": "true"}},
		{"exists_component_template", func() error {
			_, err := ext.ExistsComponentTemplateWithParams(ctx, "c1", &ClusterGetComponentTemplateParams{MasterTimeout: "m1"})
			return err
		}, "HEAD", "/_component_template/c1", map[string]string{"master_timeout": "m1"}},
		{"put_component_template", func() error {
			_, err := svc.PutComponentTemplate(ctx, &PutComponentTemplateRequest{Name: "c1", Body: body, Params: &PutComponentTemplateParams{MasterTimeout: "m1"}})
			return err
		}, "PUT", "/_component_template/c1", map[string]string{"master_timeout": "m1"}},
		{"delete_component_template", func() error {
			_, err := svc.DeleteComponentTemplate(ctx, "c1", &ClusterDeleteComponentTemplateParams{MasterTimeout: "m1"})
			return err
		}, "DELETE", "/_component_template/c1", map[string]string{"master_timeout": "m1"}},
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
