package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// catServer returns a test server that records path/method/query and responds
// with the given JSON body.
func catServer(t *testing.T, gotPath, gotMethod *string, gotQuery *map[string]string, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gotPath != nil {
			*gotPath = r.URL.Path
		}
		if gotMethod != nil {
			*gotMethod = r.Method
		}
		if gotQuery != nil {
			*gotQuery = queryMap(r)
		}
		jsonResponse(w, 200, body)
	}))
}

func TestUnitCatAliases(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"alias":"a","index":"idx","filter":"-","routing.index":"-","routing.search":"-","is_write_index":"-"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Aliases(context.Background(), []string{"a"}, &CatAliasesParams{ExpandWildcards: ExpandWildcardOpen})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/aliases/a", gotPath)
	assert.Equal(t, "json", gotQuery["format"])
	assert.Equal(t, "open", gotQuery["expand_wildcards"])
	require.Len(t, resp, 1)
	assert.Equal(t, "a", resp[0].Alias)
	assert.Equal(t, "idx", resp[0].Index)
}

func TestUnitCatAliasesAll(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Aliases(context.Background(), nil, &CatAliasesParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/aliases", gotPath)
}

func TestUnitCatAliasesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Aliases(context.Background(), nil, &CatAliasesParams{})
	assert.Error(t, err)
}

func TestUnitCatAliasesNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Aliases(context.Background(), nil, &CatAliasesParams{})
	assert.Error(t, err)
}

func TestUnitCatAllocation(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"shards":1,"disk":"5gb","disk.indices":"1gb","disk.used":"2gb","disk.avail":"3gb","disk.total":"5gb","disk.percent":"40","host":"h","ip":"127.0.0.1","node":"n"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Allocation(context.Background(), []string{"n"}, &CatAllocationParams{Bytes: BytesGb})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/allocation/n", gotPath)
	assert.Equal(t, "gb", gotQuery["bytes"])
	require.Len(t, resp, 1)
	assert.Equal(t, 1, resp[0].Shards)
	assert.Equal(t, "n", resp[0].Node)
}

func TestUnitCatAllocationError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Allocation(context.Background(), nil, &CatAllocationParams{})
	assert.Error(t, err)
}

func TestUnitCatAllocationNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Allocation(context.Background(), nil, &CatAllocationParams{})
	assert.Error(t, err)
}

func TestUnitCatCircuitBreaker(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"node":"n","field":"fielddata","breaker":"parent","estimated_size":"0","limit_size":"0","overhead":"1","tripped":"0","limit":"0"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.CircuitBreaker(context.Background(), &CatCircuitBreakerParams{Bytes: BytesB})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/circuit_breakers", gotPath)
	assert.Equal(t, "b", gotQuery["bytes"])
	require.Len(t, resp, 1)
	assert.Equal(t, "n", resp[0].Node)
	assert.Equal(t, "parent", resp[0].Breaker)
}

func TestUnitCatCircuitBreakerError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.CircuitBreaker(context.Background(), &CatCircuitBreakerParams{})
	assert.Error(t, err)
}

func TestUnitCatCircuitBreakerNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.CircuitBreaker(context.Background(), &CatCircuitBreakerParams{})
	assert.Error(t, err)
}

func TestUnitCatComponentTemplates(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"name":"ct1","version":"1"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.ComponentTemplates(context.Background(), "ct1", &CatComponentTemplatesParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/component_templates/ct1", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "ct1", resp[0].Name)
}

func TestUnitCatComponentTemplatesAll(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.ComponentTemplates(context.Background(), "", &CatComponentTemplatesParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/component_templates", gotPath)
}

func TestUnitCatComponentTemplatesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.ComponentTemplates(context.Background(), "", &CatComponentTemplatesParams{})
	assert.Error(t, err)
}

func TestUnitCatComponentTemplatesNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.ComponentTemplates(context.Background(), "", &CatComponentTemplatesParams{})
	assert.Error(t, err)
}

func TestUnitCatCount(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"epoch":"0","timestamp":"00:00:00","count":"42"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Count(context.Background(), []string{"idx"}, &CatCountParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/count/idx", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "42", resp[0].Count)
}

func TestUnitCatCountError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Count(context.Background(), nil, &CatCountParams{})
	assert.Error(t, err)
}

func TestUnitCatCountNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Count(context.Background(), nil, &CatCountParams{})
	assert.Error(t, err)
}

func TestUnitCatFielddata(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"node":"n","field":"f","size":"1kb"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Fielddata(context.Background(), []string{"f"}, &CatFielddataParams{Bytes: BytesKb})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/fielddata/f", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "f", resp[0].Field)
}

func TestUnitCatFielddataError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Fielddata(context.Background(), nil, &CatFielddataParams{})
	assert.Error(t, err)
}

func TestUnitCatFielddataNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Fielddata(context.Background(), nil, &CatFielddataParams{})
	assert.Error(t, err)
}

func TestUnitCatHealth(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"epoch":"0","timestamp":"00:00:00","cluster":"c","status":"green","node.total":"1","node.data":"1","shards":"1","pri":"1","relo":"0","init":"0","unassign":"0","pending_tasks":"0","max_task_wait":"-","active_shards_percent":"100.0%"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Health(context.Background(), &CatHealthParams{Ts: ptrBool(false)})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/health", gotPath)
	assert.Equal(t, "false", gotQuery["ts"])
	require.Len(t, resp, 1)
	assert.Equal(t, "green", resp[0].Status)
	assert.Equal(t, "c", resp[0].Cluster)
}

func TestUnitCatHealthError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Health(context.Background(), &CatHealthParams{})
	assert.Error(t, err)
}

func TestUnitCatHealthNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Health(context.Background(), &CatHealthParams{})
	assert.Error(t, err)
}

func TestUnitCatHelp(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(200)
		_, _ = w.Write([]byte("help text"))
	}))
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	out, err := svc.Help(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_cat", gotPath)
	assert.Equal(t, "help text", out)
}

func TestUnitCatHelpError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Help(context.Background())
	assert.Error(t, err)
}

func TestUnitCatHelpNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Help(context.Background())
	assert.Error(t, err)
}

func TestUnitCatIndices(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"health":"green","status":"open","index":"idx","uuid":"u","pri":"1","rep":"0","docs.count":"10","docs.deleted":"0","store.size":"1kb","pri.store.size":"1kb"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Indices(context.Background(), []string{"idx"}, &CatIndicesParams{Bytes: BytesKb, Health: HealthStatusGreen, ExpandWildcards: ExpandWildcardOpen, IncludeUnloadedSegments: ptrBool(true), Pri: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/indices/idx", gotPath)
	assert.Equal(t, "kb", gotQuery["bytes"])
	assert.Equal(t, "green", gotQuery["health"])
	assert.Equal(t, "open", gotQuery["expand_wildcards"])
	assert.Equal(t, "true", gotQuery["include_unloaded_segments"])
	assert.Equal(t, "true", gotQuery["pri"])
	require.Len(t, resp, 1)
	assert.Equal(t, "idx", resp[0].Index)
	assert.Equal(t, "green", resp[0].Health)
}

func TestUnitCatIndicesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Indices(context.Background(), nil, &CatIndicesParams{})
	assert.Error(t, err)
}

func TestUnitCatIndicesNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Indices(context.Background(), nil, &CatIndicesParams{})
	assert.Error(t, err)
}

func TestUnitCatMaster(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"id":"id","host":"h","ip":"127.0.0.1","node":"n"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Master(context.Background(), &CatMasterParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/master", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "n", resp[0].Node)
}

func TestUnitCatMasterError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Master(context.Background(), &CatMasterParams{})
	assert.Error(t, err)
}

func TestUnitCatMasterNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Master(context.Background(), &CatMasterParams{})
	assert.Error(t, err)
}

func TestUnitCatMLDataFrameAnalytics(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"id":"a","state":"started","progress":"0.5"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.MLDataFrameAnalytics(context.Background(), "a", &CatMLDataFrameAnalyticsParams{AllowNoMatch: ptrBool(true), Bytes: BytesB, From: ptrInt(0), Size: ptrInt(10), Time: "1m"})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/ml/data_frame/analytics/a", gotPath)
	assert.Equal(t, "true", gotQuery["allow_no_match"])
	assert.Equal(t, "b", gotQuery["bytes"])
	assert.Equal(t, "0", gotQuery["from"])
	assert.Equal(t, "10", gotQuery["size"])
	assert.Equal(t, "1m", gotQuery["time"])
	require.Len(t, resp, 1)
	assert.Equal(t, "a", resp[0].ID)
}

func TestUnitCatMLDataFrameAnalyticsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.MLDataFrameAnalytics(context.Background(), "", &CatMLDataFrameAnalyticsParams{})
	assert.Error(t, err)
}

func TestUnitCatMLDataFrameAnalyticsNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.MLDataFrameAnalytics(context.Background(), "", &CatMLDataFrameAnalyticsParams{})
	assert.Error(t, err)
}

func TestUnitCatMLDatafeeds(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"id":"d","state":"started","node.assignment":"n"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.MLDatafeeds(context.Background(), "d", &CatMLDatafeedsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/ml/datafeeds/d", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "d", resp[0].ID)
}

func TestUnitCatMLDatafeedsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.MLDatafeeds(context.Background(), "", &CatMLDatafeedsParams{})
	assert.Error(t, err)
}

func TestUnitCatMLDatafeedsNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.MLDatafeeds(context.Background(), "", &CatMLDatafeedsParams{})
	assert.Error(t, err)
}

func TestUnitCatMLJobs(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"id":"j","state":"opened","opened_time":"2020"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.MLJobs(context.Background(), "j", &CatMLJobsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/ml/jobs/j", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "j", resp[0].ID)
}

func TestUnitCatMLJobsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.MLJobs(context.Background(), "", &CatMLJobsParams{})
	assert.Error(t, err)
}

func TestUnitCatMLJobsNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.MLJobs(context.Background(), "", &CatMLJobsParams{})
	assert.Error(t, err)
}

func TestUnitCatMLTrainedModels(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"id":"m","type":"regression","algorithm":"linear"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.MLTrainedModels(context.Background(), "m", &CatMLTrainedModelsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/ml/trained_models/m", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "m", resp[0].ID)
	assert.Equal(t, "regression", resp[0].Type)
}

func TestUnitCatMLTrainedModelsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.MLTrainedModels(context.Background(), "", &CatMLTrainedModelsParams{})
	assert.Error(t, err)
}

func TestUnitCatMLTrainedModelsNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.MLTrainedModels(context.Background(), "", &CatMLTrainedModelsParams{})
	assert.Error(t, err)
}

func TestUnitCatNodeAttrs(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"node":"n","attr":"a","value":"v"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.NodeAttrs(context.Background(), &CatNodeAttrsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/nodeattrs", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "v", resp[0].Value)
}

func TestUnitCatNodeAttrsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.NodeAttrs(context.Background(), &CatNodeAttrsParams{})
	assert.Error(t, err)
}

func TestUnitCatNodeAttrsNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.NodeAttrs(context.Background(), &CatNodeAttrsParams{})
	assert.Error(t, err)
}

func TestUnitCatNodes(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"ip":"127.0.0.1","heap.percent":"50","ram.percent":"50","cpu":"10","load_1m":"0.1","load_5m":"0.2","load_15m":"0.3","node.role":"d","master":"*","name":"n"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Nodes(context.Background(), &CatNodesParams{Bytes: BytesB, FullId: ptrBool(true), Time: "1m"})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/nodes", gotPath)
	assert.Equal(t, "b", gotQuery["bytes"])
	assert.Equal(t, "true", gotQuery["full_id"])
	assert.Equal(t, "1m", gotQuery["time"])
	require.Len(t, resp, 1)
	assert.Equal(t, "n", resp[0].Name)
}

func TestUnitCatNodesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Nodes(context.Background(), &CatNodesParams{})
	assert.Error(t, err)
}

func TestUnitCatNodesNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Nodes(context.Background(), &CatNodesParams{})
	assert.Error(t, err)
}

func TestUnitCatPendingTasks(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"insert_order":"1","time_in_queue":"1m","priority":"high","source":"s"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.PendingTasks(context.Background(), &CatPendingTasksParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/pending_tasks", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "high", resp[0].Priority)
}

func TestUnitCatPendingTasksError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.PendingTasks(context.Background(), &CatPendingTasksParams{})
	assert.Error(t, err)
}

func TestUnitCatPendingTasksNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.PendingTasks(context.Background(), &CatPendingTasksParams{})
	assert.Error(t, err)
}

func TestUnitCatPlugins(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"name":"p","version":"8.0.0","elasticsearch_version":"8.0.0","description":"d"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Plugins(context.Background(), &CatPluginsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/plugins", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "p", resp[0].Name)
}

func TestUnitCatPluginsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Plugins(context.Background(), &CatPluginsParams{})
	assert.Error(t, err)
}

func TestUnitCatPluginsNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Plugins(context.Background(), &CatPluginsParams{})
	assert.Error(t, err)
}

func TestUnitCatRecovery(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"index":"idx","shard":"0","start_time":"-","start_time_millis":"0","stop_time":"-","stop_time_millis":"0","total_time":"1m","total_time_millis":"60000","source":"n1","target":"n2","type":"peer","stage":"done","repository":"-","snapshot":"-","files":"1","files_recovered":"1","files_percent":"100%","bytes":"1b","bytes_recovered":"1b","bytes_percent":"100%"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Recovery(context.Background(), []string{"idx"}, &CatRecoveryParams{ActiveOnly: ptrBool(true), Bytes: BytesB, Detailed: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/recovery/idx", gotPath)
	assert.Equal(t, "true", gotQuery["active_only"])
	assert.Equal(t, "b", gotQuery["bytes"])
	assert.Equal(t, "true", gotQuery["detailed"])
	require.Len(t, resp, 1)
	assert.Equal(t, "idx", resp[0].Index)
	assert.Equal(t, "done", resp[0].Stage)
}

func TestUnitCatRecoveryError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Recovery(context.Background(), nil, &CatRecoveryParams{})
	assert.Error(t, err)
}

func TestUnitCatRecoveryNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Recovery(context.Background(), nil, &CatRecoveryParams{})
	assert.Error(t, err)
}

func TestUnitCatRepositories(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"id":"r","type":"fs"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Repositories(context.Background(), &CatRepositoriesParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/repositories", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "r", resp[0].ID)
	assert.Equal(t, "fs", resp[0].Type)
}

func TestUnitCatRepositoriesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Repositories(context.Background(), &CatRepositoriesParams{})
	assert.Error(t, err)
}

func TestUnitCatRepositoriesNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Repositories(context.Background(), &CatRepositoriesParams{})
	assert.Error(t, err)
}

func TestUnitCatSegments(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"index":"idx","shard":"0","prirep":"p","ip":"127.0.0.1","segment":"_0","generation":"0","docs.count":"1","docs.deleted":"0","size":"1kb","size.memory":"1","committed":"true","searchable":"true","version":"8.0.0","compound":"false"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Segments(context.Background(), []string{"idx"}, &CatSegmentsParams{Bytes: BytesKb, Verbose: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/segments/idx", gotPath)
	assert.Equal(t, "kb", gotQuery["bytes"])
	assert.Equal(t, "true", gotQuery["verbose"])
	require.Len(t, resp, 1)
	assert.Equal(t, "idx", resp[0].Index)
	assert.Equal(t, "_0", resp[0].Segment)
}

func TestUnitCatSegmentsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Segments(context.Background(), nil, &CatSegmentsParams{})
	assert.Error(t, err)
}

func TestUnitCatSegmentsNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Segments(context.Background(), nil, &CatSegmentsParams{})
	assert.Error(t, err)
}

func TestUnitCatShards(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"index":"idx","shard":"0","prirep":"p","state":"STARTED","docs":"1","store":"1kb","ip":"127.0.0.1","node":"n"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Shards(context.Background(), []string{"idx"}, &CatShardsParams{Bytes: BytesKb})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/shards/idx", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "STARTED", resp[0].State)
}

func TestUnitCatShardsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Shards(context.Background(), nil, &CatShardsParams{})
	assert.Error(t, err)
}

func TestUnitCatShardsNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Shards(context.Background(), nil, &CatShardsParams{})
	assert.Error(t, err)
}

func TestUnitCatSnapshots(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"id":"s","repository":"r","status":"SUCCESS","start_epoch":"0","start_time":"00:00:00","end_epoch":"1","end_time":"00:00:01","duration":"1s","indices":"1"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Snapshots(context.Background(), "r", &CatSnapshotsParams{IgnoreUnavailable: ptrBool(true), Time: "1m"})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/snapshots/r", gotPath)
	assert.Equal(t, "true", gotQuery["ignore_unavailable"])
	assert.Equal(t, "1m", gotQuery["time"])
	require.Len(t, resp, 1)
	assert.Equal(t, "s", resp[0].ID)
	assert.Equal(t, "SUCCESS", resp[0].Status)
}

func TestUnitCatSnapshotsAll(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Snapshots(context.Background(), "", &CatSnapshotsParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/snapshots", gotPath)
}

func TestUnitCatSnapshotsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Snapshots(context.Background(), "", &CatSnapshotsParams{})
	assert.Error(t, err)
}

func TestUnitCatSnapshotsNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Snapshots(context.Background(), "", &CatSnapshotsParams{})
	assert.Error(t, err)
}

func TestUnitCatTasks(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"id":"n:1","action":"cluster:monitor","type":"transport","parent":"-","start_epoch":"0","start_time":"00:00:00","node":"n"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Tasks(context.Background(), &CatTasksParams{Actions: []string{"cluster:*"}, Detailed: ptrBool(true), NodeId: []string{"n"}, ParentTaskId: "parent"})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/tasks", gotPath)
	assert.Equal(t, "cluster:*", gotQuery["actions"])
	assert.Equal(t, "true", gotQuery["detailed"])
	assert.Equal(t, "n", gotQuery["node_id"])
	assert.Equal(t, "parent", gotQuery["parent_task_id"])
	require.Len(t, resp, 1)
	assert.Equal(t, "cluster:monitor", resp[0].Action)
}

func TestUnitCatTasksError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Tasks(context.Background(), &CatTasksParams{})
	assert.Error(t, err)
}

func TestUnitCatTasksNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Tasks(context.Background(), &CatTasksParams{})
	assert.Error(t, err)
}

func TestUnitCatTemplates(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[{"name":"t","index_patterns":"*","order":"1","version":"1"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Templates(context.Background(), "t", &CatTemplatesParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/templates/t", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "t", resp[0].Name)
	assert.Equal(t, "*", resp[0].IndexPatterns)
}

func TestUnitCatTemplatesAll(t *testing.T) {
	var gotPath string
	srv := catServer(t, &gotPath, nil, nil, `[]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Templates(context.Background(), "", &CatTemplatesParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/templates", gotPath)
}

func TestUnitCatTemplatesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Templates(context.Background(), "", &CatTemplatesParams{})
	assert.Error(t, err)
}

func TestUnitCatTemplatesNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Templates(context.Background(), "", &CatTemplatesParams{})
	assert.Error(t, err)
}

func TestUnitCatThreadPool(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"node_name":"n","name":"generic","active":"0","queue":"0","rejected":"0","completed":"0"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.ThreadPool(context.Background(), []string{"generic"}, &CatThreadPoolParams{Size: ptrInt(10), Time: "1m"})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/thread_pool/generic", gotPath)
	assert.Equal(t, "10", gotQuery["size"])
	assert.Equal(t, "1m", gotQuery["time"])
	require.Len(t, resp, 1)
	assert.Equal(t, "generic", resp[0].Name)
}

func TestUnitCatThreadPoolError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.ThreadPool(context.Background(), nil, &CatThreadPoolParams{})
	assert.Error(t, err)
}

func TestUnitCatThreadPoolNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.ThreadPool(context.Background(), nil, &CatThreadPoolParams{})
	assert.Error(t, err)
}

func TestUnitCatTransforms(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := catServer(t, &gotPath, nil, &gotQuery, `[{"id":"t","state":"started","checkpoint":"1"}]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	resp, err := svc.Transforms(context.Background(), "t", &CatTransformsParams{AllowNoMatch: ptrBool(true), From: ptrInt(0), Size: ptrInt(10), Time: "1m"})
	require.NoError(t, err)
	assert.Equal(t, "/_cat/transforms/t", gotPath)
	assert.Equal(t, "true", gotQuery["allow_no_match"])
	assert.Equal(t, "0", gotQuery["from"])
	assert.Equal(t, "10", gotQuery["size"])
	assert.Equal(t, "1m", gotQuery["time"])
	require.Len(t, resp, 1)
	assert.Equal(t, "t", resp[0].ID)
	assert.Equal(t, "started", resp[0].State)
}

func TestUnitCatTransformsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Transforms(context.Background(), "", &CatTransformsParams{})
	assert.Error(t, err)
}

func TestUnitCatTransformsNetworkError(t *testing.T) {
	svc := NewCatService(deadClient(), testLogger())
	_, err := svc.Transforms(context.Background(), "", &CatTransformsParams{})
	assert.Error(t, err)
}

func TestUnitCatCommonParamsSerialization(t *testing.T) {
	var gotQuery map[string]string
	srv := catServer(t, nil, nil, &gotQuery, `[]`)
	defer srv.Close()
	svc := NewCatService(restyClient(srv), testLogger())
	_, err := svc.Aliases(context.Background(), nil, &CatAliasesParams{
		catCommonParams: catCommonParams{
			H:      []string{"alias", "index"},
			Help:   ptrBool(false),
			S:      []string{"index"},
			V:      ptrBool(true),
			Local:  ptrBool(true),
			Common: &CommonParams{Pretty: ptrBool(true)},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "json", gotQuery["format"])
	assert.Equal(t, "alias,index", gotQuery["h"])
	assert.Equal(t, "false", gotQuery["help"])
	assert.Equal(t, "index", gotQuery["s"])
	assert.Equal(t, "true", gotQuery["v"])
	assert.Equal(t, "true", gotQuery["local"])
	assert.Equal(t, "true", gotQuery["pretty"])
}
