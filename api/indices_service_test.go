package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitIndicesCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/idx", r.URL.Path)
		jsonResponse(w, 200, `{"acknowledged":true,"shards_acknowledged":true,"index":"idx"}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Create(context.Background(), "idx", map[string]any{"settings": map[string]any{"number_of_shards": 1}}, &IndicesCreateParams{})
	assert.NoError(t, err)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/idx", r.URL.Path)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background(), []string{"idx"}, &IndicesDeleteParams{})
	assert.NoError(t, err)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesExists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "HEAD", r.Method)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	ok, err := svc.Exists(context.Background(), []string{"idx"}, &IndicesExistsParams{})
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUnitIndicesExistsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	ok, err := svc.Exists(context.Background(), []string{"idx"}, &IndicesExistsParams{})
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestUnitIndicesGetMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_mapping", r.URL.Path)
		jsonResponse(w, 200, `{"idx":{"mappings":{"properties":{"f":{"type":"keyword"}}}}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	result, err := svc.GetMapping(context.Background(), []string{"idx"}, &IndicesGetMappingParams{})
	assert.NoError(t, err)
	assert.Contains(t, result, "idx")
}

func TestUnitIndicesRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_refresh", r.URL.Path)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Refresh(context.Background(), []string{"idx"}, &IndicesRefreshParams{})
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Shards.Total)
}

func TestUnitIndicesRefreshAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_refresh", r.URL.Path)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Refresh(context.Background(), nil, &IndicesRefreshParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesRollover(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"old_index":"a-000001","new_index":"a-000002","rolled_over":true,"acknowledged":true,"shards_acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Rollover(context.Background(), &RolloverRequest{
		Alias:  "alias",
		Body:   map[string]any{"conditions": map[string]any{"max_age": "1d"}},
		Params: &RolloverParams{DryRun: ptrBool(true)},
	})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/alias/_rollover", gotPath)
	assert.NotEmpty(t, gotBody)
	assert.True(t, resp.RolledOver)
}

func TestUnitIndicesRolloverNewIndex(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"old_index":"a","new_index":"b","acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Rollover(context.Background(), &RolloverRequest{
		Alias:    "alias",
		NewIndex: "new-idx",
		Params:   &RolloverParams{WaitForActiveShards: "1"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "/alias/_rollover/new-idx", gotPath)
	assert.Equal(t, "1", gotQuery["wait_for_active_shards"])
}

func TestUnitIndicesRolloverValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Rollover(context.Background(), &RolloverRequest{})
	assert.Error(t, err)
}

func TestUnitIndicesShrink(t *testing.T) {
	var gotMethod string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true,"shards_acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Shrink(context.Background(), &ShrinkRequest{
		Source: "src",
		Target: "tgt",
		Body:   map[string]any{"settings": map[string]any{"number_of_shards": 1}},
		Params: &ShrinkParams{Timeout: "30s"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/src/_shrink/tgt", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesSplit(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Split(context.Background(), &SplitRequest{
		Source: "src",
		Target: "tgt",
		Params: &SplitParams{WaitForActiveShards: "1"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "/src/_split/tgt", gotPath)
}

func TestUnitIndicesClone(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Clone(context.Background(), &CloneRequest{
		Source: "src",
		Target: "tgt",
		Params: &CloneParams{},
	})
	assert.NoError(t, err)
	assert.Equal(t, "/src/_clone/tgt", gotPath)
}

func TestUnitIndicesCloneMissingFields(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Clone(context.Background(), &CloneRequest{Source: "src"})
	assert.Error(t, err)
}

func TestUnitIndicesFlush(t *testing.T) {
	var gotMethod string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Flush(context.Background(), []string{"idx"}, &IndicesFlushParams{Force: ptrBool(true)})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx/_flush", gotPath)
	assert.NotNil(t, resp.Shards)
}

func TestUnitIndicesFlushAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_flush", r.URL.Path)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Flush(context.Background(), nil, &IndicesFlushParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesForceMerge(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.ForceMerge(context.Background(), []string{"idx"}, &IndicesForceMergeParams{MaxNumSegments: ptrInt(1), OnlyExpungeDeletes: ptrBool(true)})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx/_forcemerge", gotPath)
	assert.Equal(t, "1", gotQuery["max_num_segments"])
	assert.Equal(t, "true", gotQuery["only_expunge_deletes"])
}

func TestUnitIndicesForceMergeAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_forcemerge", r.URL.Path)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.ForceMerge(context.Background(), nil, &IndicesForceMergeParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesClearCache(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.ClearCache(context.Background(), []string{"idx"}, &IndicesClearCacheParams{Fielddata: ptrBool(true), Fields: []string{"f1"}})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx/_cache/clear", gotPath)
	assert.Equal(t, "true", gotQuery["fielddata"])
	assert.Contains(t, gotQuery["fields"], "f1")
}

func TestUnitIndicesClearCacheAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_cache/clear", r.URL.Path)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.ClearCache(context.Background(), nil, &IndicesClearCacheParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesAnalyze(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"tokens":[{"token":"foo","start_offset":0,"end_offset":3,"position":0}]}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Analyze(context.Background(), "idx", map[string]any{"text": "foo bar"})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx/_analyze", gotPath)
	assert.NotEmpty(t, gotBody)
	assert.Len(t, resp.Tokens, 1)
}

func TestUnitIndicesAnalyzeNoIndex(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		jsonResponse(w, 200, `{"tokens":[]}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Analyze(context.Background(), "", nil)
	assert.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
}

func TestUnitIndicesStats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_stats", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Stats(context.Background(), &IndicesStatsRequest{Params: &IndicesStatsParams{Level: LevelIndices}})
	assert.NoError(t, err)
	assert.NotNil(t, resp.Shards)
}

func TestUnitIndicesStatsIndicesMetrics(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Stats(context.Background(), &IndicesStatsRequest{
		Indices: []string{"idx"},
		Metrics: []string{"docs"},
		Params:  &IndicesStatsParams{Level: LevelShards},
	})
	assert.NoError(t, err)
	assert.Equal(t, "/idx/_stats/docs", gotPath)
	assert.Equal(t, "shards", gotQuery["level"])
}

func TestUnitIndicesStatsMetricsOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_stats/docs", r.URL.Path)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Stats(context.Background(), &IndicesStatsRequest{Metrics: []string{"docs"}, Params: &IndicesStatsParams{}})
	assert.NoError(t, err)
}

func TestUnitIndicesSegments(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0},"indices":{}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Segments(context.Background(), []string{"idx"}, &IndicesSegmentsParams{Verbose: ptrBool(true)})
	assert.NoError(t, err)
	assert.Equal(t, "/idx/_segments", gotPath)
	assert.Equal(t, "true", gotQuery["verbose"])
}

func TestUnitIndicesSegmentsAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_segments", r.URL.Path)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0},"indices":{}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Segments(context.Background(), nil, &IndicesSegmentsParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesRecovery(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"idx":{"shards":[]}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Recovery(context.Background(), []string{"idx"}, &IndicesRecoveryParams{Detailed: ptrBool(true)})
	assert.NoError(t, err)
	assert.Equal(t, "/idx/_recovery", gotPath)
	assert.Equal(t, "true", gotQuery["detailed"])
	assert.Contains(t, *resp, "idx")
}

func TestUnitIndicesRecoveryAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_recovery", r.URL.Path)
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.Recovery(context.Background(), nil, &IndicesRecoveryParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesShardStores(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"indices":{}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.ShardStores(context.Background(), []string{"idx"}, &IndicesShardStoresParams{Status: []string{"green"}})
	assert.NoError(t, err)
	assert.Equal(t, "/idx/_shard_stores", gotPath)
	assert.Contains(t, gotQuery["status"], "green")
}

func TestUnitIndicesShardStoresAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_shard_stores", r.URL.Path)
		jsonResponse(w, 200, `{"indices":{}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.ShardStores(context.Background(), nil, &IndicesShardStoresParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesReloadSearchAnalyzers(t *testing.T) {
	var gotMethod string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.ReloadSearchAnalyzers(context.Background(), []string{"idx"}, &IndicesReloadAnalyzersParams{})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx/_reload_search_analyzers", gotPath)
}

func TestUnitIndicesReloadSearchAnalyzersEmpty(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ReloadSearchAnalyzers(context.Background(), nil, &IndicesReloadAnalyzersParams{})
	assert.Error(t, err)
}

func TestUnitIndicesDownsample(t *testing.T) {
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
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Downsample(context.Background(), "src", "tgt", map[string]any{"fixed_interval": "1h"})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/src/_downsample/tgt", gotPath)
	assert.NotEmpty(t, gotBody)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesDownsampleMissingFields(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Downsample(context.Background(), "", "tgt", nil)
	assert.Error(t, err)
}

func TestUnitIndicesAddBlock(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true,"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.AddBlock(context.Background(), &AddBlockRequest{
		Indices: []string{"idx"},
		Block:   "write",
		Params:  &AddBlockParams{Timeout: "30s"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/idx/_block/write", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesAddBlockValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.AddBlock(context.Background(), &AddBlockRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

func TestUnitIndicesPutAlias(t *testing.T) {
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
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.PutAlias(context.Background(), &PutAliasRequest{
		Index:  "idx",
		Alias:  "a1",
		Body:   map[string]any{"is_write_index": true},
		Params: &AliasParams{Timeout: "30s"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/idx/_alias/a1", gotPath)
	assert.NotEmpty(t, gotBody)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesPutAliasValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutAlias(context.Background(), &PutAliasRequest{Index: "idx"})
	assert.Error(t, err)
}

func TestUnitIndicesGetAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_alias/a1", r.URL.Path)
		jsonResponse(w, 200, `{"idx":{"aliases":{"a1":{}}}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.GetAlias(context.Background(), &GetAliasRequest{
		Indices: []string{"idx"},
		Names:   []string{"a1"},
		Params:  &GetAliasParams{},
	})
	assert.NoError(t, err)
	assert.Contains(t, resp, "idx")
}

func TestUnitIndicesGetAliasAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_alias", r.URL.Path)
		jsonResponse(w, 200, `{"idx":{"aliases":{}}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.GetAlias(context.Background(), &GetAliasRequest{Params: &GetAliasParams{}})
	assert.NoError(t, err)
}

func TestUnitIndicesGetAliasNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.GetAlias(context.Background(), &GetAliasRequest{Names: []string{"missing"}, Params: &GetAliasParams{}})
	assert.NoError(t, err)
	assert.Empty(t, resp)
}

func TestUnitIndicesExistsAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_alias/a1", r.URL.Path)
		assert.Equal(t, "HEAD", r.Method)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	ok, err := svc.ExistsAlias(context.Background(), &ExistsAliasRequest{
		Indices: []string{"idx"},
		Names:   []string{"a1"},
		Params:  &GetAliasParams{},
	})
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUnitIndicesExistsAliasNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	ok, err := svc.ExistsAlias(context.Background(), &ExistsAliasRequest{Names: []string{"a1"}, Params: &GetAliasParams{}})
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestUnitIndicesExistsAliasValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ExistsAlias(context.Background(), &ExistsAliasRequest{})
	assert.Error(t, err)
}

func TestUnitIndicesDeleteAlias(t *testing.T) {
	var gotMethod string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.DeleteAlias(context.Background(), &DeleteAliasRequest{
		Indices: []string{"idx"},
		Names:   []string{"a1"},
		Params:  &AliasParams{},
	})
	assert.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/idx/_alias/a1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesDeleteAliasValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteAlias(context.Background(), &DeleteAliasRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

func TestUnitIndicesUpdateAliases(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.UpdateAliases(context.Background(), map[string]any{"actions": []map[string]any{}}, &IndicesUpdateAliasesParams{Timeout: "30s"})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_aliases", gotPath)
	assert.NotEmpty(t, gotBody)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesGetSettings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_settings", r.URL.Path)
		jsonResponse(w, 200, `{"idx":{"settings":{}}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.GetSettings(context.Background(), &GetSettingsRequest{
		Indices: []string{"idx"},
		Params:  &GetSettingsParams{IncludeDefaults: ptrBool(true)},
	})
	assert.NoError(t, err)
	assert.Contains(t, resp, "idx")
}

func TestUnitIndicesGetSettingsAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_settings", r.URL.Path)
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.GetSettings(context.Background(), &GetSettingsRequest{Params: &GetSettingsParams{}})
	assert.NoError(t, err)
}

func TestUnitIndicesGetSettingsNames(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_settings/index.*", r.URL.Path)
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.GetSettings(context.Background(), &GetSettingsRequest{Names: []string{"index.*"}, Params: &GetSettingsParams{}})
	assert.NoError(t, err)
}

func TestUnitIndicesPutSettings(t *testing.T) {
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
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.PutSettings(context.Background(), []string{"idx"}, map[string]any{"index": map[string]any{"number_of_replicas": 0}}, &IndicesPutSettingsParams{PreserveExisting: ptrBool(true)})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/idx/_settings", gotPath)
	assert.NotEmpty(t, gotBody)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesPutSettingsAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_settings", r.URL.Path)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.PutSettings(context.Background(), nil, map[string]any{}, &IndicesPutSettingsParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesGetMappingAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_mapping", r.URL.Path)
		jsonResponse(w, 200, `{"idx":{"mappings":{"properties":{}}}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.GetMapping(context.Background(), nil, &IndicesGetMappingParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesPutMapping(t *testing.T) {
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
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.PutMapping(context.Background(), &PutMappingRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"properties": map[string]any{}},
		Params:  &PutMappingParams{Timeout: "30s"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/idx/_mapping", gotPath)
	assert.NotEmpty(t, gotBody)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesPutMappingValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutMapping(context.Background(), &PutMappingRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

func TestUnitIndicesGetFieldMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_mapping/field/f1", r.URL.Path)
		jsonResponse(w, 200, `{"idx":{"mappings":{}}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.GetFieldMapping(context.Background(), &GetFieldMappingRequest{
		Indices: []string{"idx"},
		Fields:  []string{"f1"},
		Params:  &GetFieldMappingParams{},
	})
	assert.NoError(t, err)
	assert.Contains(t, resp, "idx")
}

func TestUnitIndicesGetFieldMappingAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_mapping/field/f1", r.URL.Path)
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.GetFieldMapping(context.Background(), &GetFieldMappingRequest{Fields: []string{"f1"}, Params: &GetFieldMappingParams{}})
	assert.NoError(t, err)
}

func TestUnitIndicesGetFieldMappingValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.GetFieldMapping(context.Background(), &GetFieldMappingRequest{})
	assert.Error(t, err)
}

func TestUnitIndicesGetTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_template/t1", r.URL.Path)
		jsonResponse(w, 200, `{"t1":{"order":0,"index_patterns":["*"]}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.GetTemplate(context.Background(), []string{"t1"}, &IndicesGetTemplateParams{FlatSettings: ptrBool(true)})
	assert.NoError(t, err)
	assert.Contains(t, resp, "t1")
}

func TestUnitIndicesGetTemplateAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_template", r.URL.Path)
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.GetTemplate(context.Background(), nil, &IndicesGetTemplateParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesPutTemplate(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.PutTemplate(context.Background(), &PutTemplateRequest{
		Name:   "t1",
		Body:   map[string]any{"index_patterns": []string{"*"}},
		Params: &PutTemplateParams{Create: ptrBool(true), Order: ptrInt(10)},
	})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_template/t1", gotPath)
	assert.NotEmpty(t, gotBody)
	assert.Equal(t, "true", gotQuery["create"])
	assert.Equal(t, "10", gotQuery["order"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesPutTemplateValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutTemplate(context.Background(), &PutTemplateRequest{Name: "t1"})
	assert.Error(t, err)
}

func TestUnitIndicesExistsTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_template/t1", r.URL.Path)
		assert.Equal(t, "HEAD", r.Method)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	ok, err := svc.ExistsTemplate(context.Background(), "t1")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUnitIndicesExistsTemplateNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	ok, err := svc.ExistsTemplate(context.Background(), "t1")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestUnitIndicesExistsTemplateEmptyName(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ExistsTemplate(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitIndicesDeleteTemplate(t *testing.T) {
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
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.DeleteTemplate(context.Background(), "t1", &IndicesDeleteTemplateParams{Timeout: "30s"})
	assert.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_template/t1", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesDeleteTemplateEmptyName(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteTemplate(context.Background(), "", &IndicesDeleteTemplateParams{})
	assert.Error(t, err)
}

func TestUnitIndicesGetIndexTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_index_template", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"index_templates":[]}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.GetIndexTemplate(context.Background(), nil, &IndicesGetIndexTemplateParams{Local: ptrBool(true)})
	assert.NoError(t, err)
	assert.NotNil(t, resp.IndexTemplates)
}

func TestUnitIndicesGetIndexTemplateNamed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_index_template/t1", r.URL.Path)
		jsonResponse(w, 200, `{"index_templates":[]}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.GetIndexTemplate(context.Background(), []string{"t1"}, &IndicesGetIndexTemplateParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesPutIndexTemplate(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.PutIndexTemplate(context.Background(), &PutIndexTemplateRequest{
		Name:   "t1",
		Body:   map[string]any{"index_patterns": []string{"*"}},
		Params: &PutIndexTemplateParams{Create: ptrBool(true), Cause: "test"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_index_template/t1", gotPath)
	assert.NotEmpty(t, gotBody)
	assert.Equal(t, "true", gotQuery["create"])
	assert.Equal(t, "test", gotQuery["cause"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesPutIndexTemplateValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutIndexTemplate(context.Background(), &PutIndexTemplateRequest{Name: "t1"})
	assert.Error(t, err)
}

func TestUnitIndicesExistsIndexTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_index_template/t1", r.URL.Path)
		assert.Equal(t, "HEAD", r.Method)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	ok, err := svc.ExistsIndexTemplate(context.Background(), "t1")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUnitIndicesExistsIndexTemplateNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	ok, err := svc.ExistsIndexTemplate(context.Background(), "t1")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestUnitIndicesExistsIndexTemplateEmptyName(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ExistsIndexTemplate(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitIndicesDeleteIndexTemplate(t *testing.T) {
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
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.DeleteIndexTemplate(context.Background(), "t1", &IndicesDeleteIndexTemplateParams{Timeout: "30s"})
	assert.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_index_template/t1", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesDeleteIndexTemplateEmptyName(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteIndexTemplate(context.Background(), "", &IndicesDeleteIndexTemplateParams{})
	assert.Error(t, err)
}

func TestUnitIndicesSimulateIndexTemplate(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"template":{}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.SimulateIndexTemplate(context.Background(), &SimulateIndexTemplateRequest{
		Name:   "idx-*",
		Body:   map[string]any{"index_patterns": []string{"idx-*"}},
		Params: &SimulateIndexTemplateParams{Create: ptrBool(true)},
	})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_index_template/_simulate_index/idx-*", gotPath)
	assert.NotEmpty(t, gotBody)
}

func TestUnitIndicesSimulateIndexTemplateNoBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_index_template/_simulate_index/idx", r.URL.Path)
		jsonResponse(w, 200, `{"template":{}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.SimulateIndexTemplate(context.Background(), &SimulateIndexTemplateRequest{Name: "idx", Params: &SimulateIndexTemplateParams{}})
	assert.NoError(t, err)
}

func TestUnitIndicesSimulateIndexTemplateValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.SimulateIndexTemplate(context.Background(), &SimulateIndexTemplateRequest{})
	assert.Error(t, err)
}

func TestUnitIndicesSimulateTemplate(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"template":{}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.SimulateTemplate(context.Background(), &SimulateTemplateRequest{
		Name:   "t1",
		Body:   map[string]any{"index_patterns": []string{"*"}},
		Params: &SimulateTemplateParams{},
	})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_index_template/_simulate/t1", gotPath)
	assert.NotEmpty(t, gotBody)
}

func TestUnitIndicesSimulateTemplateNoName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_index_template/_simulate", r.URL.Path)
		jsonResponse(w, 200, `{"template":{}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.SimulateTemplate(context.Background(), &SimulateTemplateRequest{Params: &SimulateTemplateParams{}})
	assert.NoError(t, err)
}

func TestUnitIndicesCreateDataStream(t *testing.T) {
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
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.CreateDataStream(context.Background(), "logs", nil, &IndicesCreateDataStreamParams{Timeout: "30s"})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_data_stream/logs", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesCreateDataStreamEmptyName(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.CreateDataStream(context.Background(), "", nil, &IndicesCreateDataStreamParams{})
	assert.Error(t, err)
}

func TestUnitIndicesGetDataStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_data_stream", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"data_streams":[]}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.GetDataStream(context.Background(), nil, &IndicesGetDataStreamParams{IncludeDefaults: ptrBool(true)})
	assert.NoError(t, err)
	assert.NotNil(t, resp.DataStreams)
}

func TestUnitIndicesGetDataStreamNamed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_data_stream/logs", r.URL.Path)
		jsonResponse(w, 200, `{"data_streams":[]}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.GetDataStream(context.Background(), []string{"logs"}, &IndicesGetDataStreamParams{})
	assert.NoError(t, err)
}

func TestUnitIndicesDeleteDataStream(t *testing.T) {
	var gotMethod string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.DeleteDataStream(context.Background(), []string{"logs"}, &IndicesDeleteDataStreamParams{})
	assert.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_data_stream/logs", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesDeleteDataStreamEmpty(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteDataStream(context.Background(), nil, &IndicesDeleteDataStreamParams{})
	assert.Error(t, err)
}

func TestUnitIndicesDataStreamStats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_data_stream/_stats", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"data_stream_count":1,"backing_indices":1,"total_store_size_bytes":100,"data_streams":[]}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.DataStreamStats(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.DataStreamCount)
}

func TestUnitIndicesDataStreamStatsNamed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_data_stream/logs/_stats", r.URL.Path)
		jsonResponse(w, 200, `{"data_stream_count":1,"backing_indices":1,"total_store_size_bytes":100,"data_streams":[]}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	_, err := svc.DataStreamStats(context.Background(), []string{"logs"})
	assert.NoError(t, err)
}

func TestUnitIndicesMigrateToDataStream(t *testing.T) {
	var gotMethod string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.MigrateToDataStream(context.Background(), "alias")
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_data_stream/_migrate/alias", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesMigrateToDataStreamEmptyName(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.MigrateToDataStream(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitIndicesModifyDataStream(t *testing.T) {
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
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.ModifyDataStream(context.Background(), &ModifyDataStreamRequest{
		Actions: []*ModifyDataStreamAction{
			{Type: "add_backing_index", DataStream: "logs", Index: "logs-2024"},
		},
		Params: &ModifyDataStreamParams{Timeout: "30s"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_data_stream/_modify", gotPath)
	assert.Contains(t, gotBody, "actions")
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesModifyDataStreamValidate(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ModifyDataStream(context.Background(), &ModifyDataStreamRequest{})
	assert.Error(t, err)
}

func TestUnitIndicesPromoteDataStream(t *testing.T) {
	var gotMethod string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.PromoteDataStream(context.Background(), "logs")
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_data_stream/_promote/logs", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIndicesPromoteDataStreamEmptyName(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PromoteDataStream(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitIndicesResolveIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_resolve/index/idx", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"indices":[],"aliases":[],"data_streams":[]}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.ResolveIndex(context.Background(), "idx", &IndicesResolveIndexParams{ExpandWildcards: ExpandWildcardOpen})
	assert.NoError(t, err)
	assert.NotNil(t, resp.Indices)
}

func TestUnitIndicesResolveIndexEmptyName(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ResolveIndex(context.Background(), "", &IndicesResolveIndexParams{})
	assert.Error(t, err)
}

func TestUnitIndicesResolveCluster(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_resolve/cluster/remote1", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"remote1":{"connected":true,"mode":"sniff","seeds":[],"skip_unavailable":false}}`)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	resp, err := svc.ResolveCluster(context.Background(), []string{"remote1"}, &IndicesResolveClusterParams{})
	assert.NoError(t, err)
	assert.Contains(t, resp, "remote1")
}

func TestUnitIndicesResolveClusterEmpty(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ResolveCluster(context.Background(), nil, &IndicesResolveClusterParams{})
	assert.Error(t, err)
}

func TestUnitIndicesB3MasterTimeoutTransport(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	ctx := context.Background()

	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
		query  map[string]string
	}{
		{"shrink", func() error {
			_, err := svc.Shrink(ctx, &ShrinkRequest{Source: "src", Target: "tgt", Params: &ShrinkParams{MasterTimeout: "m1"}})
			return err
		}, "PUT", "/src/_shrink/tgt", map[string]string{"master_timeout": "m1"}},
		{"split", func() error {
			_, err := svc.Split(ctx, &SplitRequest{Source: "src", Target: "tgt", Params: &SplitParams{MasterTimeout: "m1"}})
			return err
		}, "PUT", "/src/_split/tgt", map[string]string{"master_timeout": "m1"}},
		{"clone", func() error {
			_, err := svc.Clone(ctx, &CloneRequest{Source: "src", Target: "tgt", Params: &CloneParams{MasterTimeout: "m1"}})
			return err
		}, "PUT", "/src/_clone/tgt", map[string]string{"master_timeout": "m1"}},
		{"open", func() error {
			_, err := svc.Open(ctx, []string{"idx"}, &IndicesOpenParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/idx/_open", map[string]string{"master_timeout": "m1"}},
		{"close", func() error {
			_, err := svc.Close(ctx, []string{"idx"}, &IndicesCloseParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/idx/_close", map[string]string{"master_timeout": "m1"}},
		{"put_alias", func() error {
			_, err := svc.PutAlias(ctx, &PutAliasRequest{Index: "idx", Alias: "al", Params: &AliasParams{MasterTimeout: "m1"}})
			return err
		}, "PUT", "/idx/_alias/al", map[string]string{"master_timeout": "m1"}},
		{"delete_alias", func() error {
			_, err := svc.DeleteAlias(ctx, &DeleteAliasRequest{Indices: []string{"idx"}, Names: []string{"al"}, Params: &AliasParams{MasterTimeout: "m1"}})
			return err
		}, "DELETE", "/idx/_alias/al", map[string]string{"master_timeout": "m1"}},
		{"get_alias", func() error {
			_, err := svc.GetAlias(ctx, &GetAliasRequest{Indices: []string{"idx"}, Names: []string{"al"}, Params: &GetAliasParams{MasterTimeout: "m1"}})
			return err
		}, "GET", "/idx/_alias/al", map[string]string{"master_timeout": "m1"}},
		{"exists_alias", func() error {
			_, err := svc.ExistsAlias(ctx, &ExistsAliasRequest{Indices: []string{"idx"}, Names: []string{"al"}, Params: &GetAliasParams{MasterTimeout: "m1"}})
			return err
		}, "HEAD", "/idx/_alias/al", map[string]string{"master_timeout": "m1"}},
		{"simulate_index_template", func() error {
			_, err := svc.SimulateIndexTemplate(ctx, &SimulateIndexTemplateRequest{Name: "tpl", Params: &SimulateIndexTemplateParams{MasterTimeout: "m1", IncludeDefaults: ptrBool(true)}})
			return err
		}, "POST", "/_index_template/_simulate_index/tpl", map[string]string{"master_timeout": "m1", "include_defaults": "true"}},
		{"simulate_template", func() error {
			_, err := svc.SimulateTemplate(ctx, &SimulateTemplateRequest{Name: "tpl", Params: &SimulateTemplateParams{MasterTimeout: "m1", IncludeDefaults: ptrBool(false)}})
			return err
		}, "POST", "/_index_template/_simulate/tpl", map[string]string{"master_timeout": "m1", "include_defaults": "false"}},
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
