package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitAsyncSearchSubmit(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"async-id","is_partial":true,"is_running":true,"start_time_in_millis":100,"expiration_time_in_millis":200,"response":{"took":1,"hits":{"total":{"value":1,"relation":"eq"},"hits":[{"_index":"idx","_id":"1"}]}}}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	resp, err := svc.Submit(context.Background(), &AsyncSearchSubmitRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"query": map[string]any{"match_all": map[string]any{}}},
		Params: &AsyncSearchSubmitParams{
			KeepAlive:                "1m",
			WaitForCompletionTimeout: "10s",
			Size:                     ptrInt(5),
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx/_async_search", gotPath)
	assert.Equal(t, "1m", gotQuery["keep_alive"])
	assert.Equal(t, "10s", gotQuery["wait_for_completion_timeout"])
	assert.Equal(t, "5", gotQuery["size"])
	assert.Contains(t, gotBody, "match_all")
	assert.Equal(t, "async-id", resp.Id)
	assert.True(t, resp.IsPartial)
	assert.True(t, resp.IsRunning)
	assert.Equal(t, int64(100), resp.StartTimeInMillis)
	assert.Equal(t, int64(200), resp.ExpirationTimeInMillis)
	require.NotNil(t, resp.Response)
	assert.Equal(t, int64(1), resp.Response.TookInMillis)
}

func TestUnitAsyncSearchSubmitClusterWide(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"id":"x","is_partial":false,"is_running":false}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	_, err := svc.Submit(context.Background(), &AsyncSearchSubmitRequest{Body: map[string]any{}})
	require.NoError(t, err)
	assert.Equal(t, "/_async_search", gotPath)
}

func TestUnitAsyncSearchSubmitNoBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"id":"x","is_running":true,"is_partial":true}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	resp, err := svc.Submit(context.Background(), &AsyncSearchSubmitRequest{})
	require.NoError(t, err)
	assert.Equal(t, "x", resp.Id)
}

func TestUnitAsyncSearchSubmitError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	_, err := svc.Submit(context.Background(), &AsyncSearchSubmitRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitAsyncSearchSubmitNetworkError(t *testing.T) {
	svc := NewAsyncSearchService(deadClient(), testLogger())
	_, err := svc.Submit(context.Background(), &AsyncSearchSubmitRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitAsyncSearchGet(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"async-id","is_partial":false,"is_running":false,"response":{"took":2}}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), "async-id", &AsyncSearchGetParams{
		KeepAlive:                "30s",
		WaitForCompletionTimeout: "5s",
		TypedKeys:                ptrBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_async_search/async-id", gotPath)
	assert.Equal(t, "30s", gotQuery["keep_alive"])
	assert.Equal(t, "5s", gotQuery["wait_for_completion_timeout"])
	assert.Equal(t, "true", gotQuery["typed_keys"])
	assert.Equal(t, "async-id", resp.Id)
	assert.False(t, resp.IsPartial)
	assert.False(t, resp.IsRunning)
	require.NotNil(t, resp.Response)
	assert.Equal(t, int64(2), resp.Response.TookInMillis)
}

func TestUnitAsyncSearchGetEmptyID(t *testing.T) {
	svc := NewAsyncSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Get(context.Background(), "", &AsyncSearchGetParams{})
	assert.Error(t, err)
}

func TestUnitAsyncSearchGetError(t *testing.T) {
	srv := errServer(404)
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "id", &AsyncSearchGetParams{})
	assert.Error(t, err)
}

func TestUnitAsyncSearchGetNetworkError(t *testing.T) {
	svc := NewAsyncSearchService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), "id", &AsyncSearchGetParams{})
	assert.Error(t, err)
}

func TestUnitAsyncSearchStatus(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"id":"async-id","is_running":true,"is_partial":true,"start_time_in_millis":100,"expiration_time_in_millis":200,"_shards":{"total":5,"successful":3,"skipped":1,"failed":1}}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	resp, err := svc.Status(context.Background(), "async-id")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_async_search/status/async-id", gotPath)
	assert.Equal(t, "async-id", resp.Id)
	assert.True(t, resp.IsRunning)
	assert.True(t, resp.IsPartial)
	assert.Equal(t, int64(100), resp.StartTimeInMillis)
	assert.Equal(t, int64(200), resp.ExpirationTimeInMillis)
	require.NotNil(t, resp.Shards)
	assert.Equal(t, 5, resp.Shards.Total)
	assert.Equal(t, 3, resp.Shards.Successful)
	assert.Equal(t, 1, resp.Shards.Skipped)
	assert.Equal(t, 1, resp.Shards.Failed)
}

func TestUnitAsyncSearchStatusEmptyID(t *testing.T) {
	svc := NewAsyncSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Status(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitAsyncSearchStatusError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	_, err := svc.Status(context.Background(), "id")
	assert.Error(t, err)
}

func TestUnitAsyncSearchStatusNetworkError(t *testing.T) {
	svc := NewAsyncSearchService(deadClient(), testLogger())
	_, err := svc.Status(context.Background(), "id")
	assert.Error(t, err)
}

func TestUnitAsyncSearchDelete(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background(), "async-id")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_async_search/async-id", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitAsyncSearchDeleteEmptyID(t *testing.T) {
	svc := NewAsyncSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Delete(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitAsyncSearchDeleteError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	_, err := svc.Delete(context.Background(), "id")
	assert.Error(t, err)
}

func TestUnitAsyncSearchDeleteNetworkError(t *testing.T) {
	svc := NewAsyncSearchService(deadClient(), testLogger())
	_, err := svc.Delete(context.Background(), "id")
	assert.Error(t, err)
}

func TestUnitAsyncSearchSubmitParamsSerialization(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"x","is_running":true,"is_partial":true}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	_, err := svc.Submit(context.Background(), &AsyncSearchSubmitRequest{
		Params: &AsyncSearchSubmitParams{
			AllowNoIndices:            ptrBool(true),
			AllowPartialSearchResults: ptrBool(false),
			Analyzer:                  "standard",
			AnalyzeWildcard:           ptrBool(true),
			BatchedReduceSize:         ptrInt(64),
			DefaultOperator:           DefaultOperatorAnd,
			Df:                        "field",
			ExpandWildcards:           ExpandWildcardOpen,
			Explain:                   ptrBool(true),
			From:                      ptrInt(10),
			IgnoreThrottled:           ptrBool(false),
			IgnoreUnavailable:         ptrBool(true),
			KeepOnCompletion:          ptrBool(false),
			Lenient:                   ptrBool(true),
			Preference:                "_local",
			Q:                         "q",
			RequestCache:              ptrBool(true),
			Routing:                   "r",
			SearchType:                SearchTypeQueryThenFetch,
			Size:                      ptrInt(20),
			Sort:                      []string{"a", "b"},
			Source:                    "true",
			SourceExcludes:            []string{"ex"},
			SourceIncludes:            []string{"in"},
			Stats:                     []string{"s"},
			StoredFields:              []string{"f"},
			SuggestField:              "sf",
			SuggestMode:               SuggestModeAlways,
			SuggestSize:               ptrInt(3),
			SuggestText:               "st",
			TerminateAfter:            ptrInt(100),
			Timeout:                   "30s",
			TrackScores:               ptrBool(true),
			TrackTotalHits:            "1000",
			TypedKeys:                 ptrBool(true),
			WaitForCompletionTimeout:  "1s",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "true", gotQuery["allow_no_indices"])
	assert.Equal(t, "false", gotQuery["allow_partial_search_results"])
	assert.Equal(t, "standard", gotQuery["analyzer"])
	assert.Equal(t, "true", gotQuery["analyze_wildcard"])
	assert.Equal(t, "64", gotQuery["batched_reduce_size"])
	assert.Equal(t, "AND", gotQuery["default_operator"])
	assert.Equal(t, "field", gotQuery["df"])
	assert.Equal(t, "open", gotQuery["expand_wildcards"])
	assert.Equal(t, "true", gotQuery["explain"])
	assert.Equal(t, "10", gotQuery["from"])
	assert.Equal(t, "false", gotQuery["ignore_throttled"])
	assert.Equal(t, "true", gotQuery["ignore_unavailable"])
	assert.Equal(t, "false", gotQuery["keep_on_completion"])
	assert.Equal(t, "true", gotQuery["lenient"])
	assert.Equal(t, "_local", gotQuery["preference"])
	assert.Equal(t, "q", gotQuery["q"])
	assert.Equal(t, "true", gotQuery["request_cache"])
	assert.Equal(t, "r", gotQuery["routing"])
	assert.Equal(t, "query_then_fetch", gotQuery["search_type"])
	assert.Equal(t, "20", gotQuery["size"])
	assert.Equal(t, "a,b", gotQuery["sort"])
	assert.Equal(t, "true", gotQuery["_source"])
	assert.Equal(t, "ex", gotQuery["_source_excludes"])
	assert.Equal(t, "in", gotQuery["_source_includes"])
	assert.Equal(t, "s", gotQuery["stats"])
	assert.Equal(t, "f", gotQuery["stored_fields"])
	assert.Equal(t, "sf", gotQuery["suggest_field"])
	assert.Equal(t, "always", gotQuery["suggest_mode"])
	assert.Equal(t, "3", gotQuery["suggest_size"])
	assert.Equal(t, "st", gotQuery["suggest_text"])
	assert.Equal(t, "100", gotQuery["terminate_after"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Equal(t, "true", gotQuery["track_scores"])
	assert.Equal(t, "1000", gotQuery["track_total_hits"])
	assert.Equal(t, "true", gotQuery["typed_keys"])
	assert.Equal(t, "1s", gotQuery["wait_for_completion_timeout"])
}

func TestUnitAsyncSearchGetParamsNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, `{"id":"x","is_running":false,"is_partial":false}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), "id", nil)
	require.NoError(t, err)
	assert.Equal(t, "x", resp.Id)
}

func TestUnitAsyncSearchSubmitB7Carrier(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"a1","is_running":true,"is_partial":false}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	_, err := svc.Submit(context.Background(), &AsyncSearchSubmitRequest{Indices: []string{"idx"}, Params: &AsyncSearchSubmitParams{
		RestTotalHitsAsInt: ptrBool(true), SeqNoPrimaryTerm: ptrBool(true), Version: ptrBool(true),
		CcsMinimizeRoundtrips: ptrBool(false), MaxConcurrentShardRequests: ptrInt(5), DocvalueFields: []string{"f1", "f2"},
	}})
	assert.NoError(t, err)
	assert.Equal(t, "true", gotQuery["rest_total_hits_as_int"])
	assert.Equal(t, "true", gotQuery["seq_no_primary_term"])
	assert.Equal(t, "true", gotQuery["version"])
	assert.Equal(t, "false", gotQuery["ccs_minimize_roundtrips"])
	assert.Equal(t, "5", gotQuery["max_concurrent_shard_requests"])
	assert.Equal(t, "f1,f2", gotQuery["docvalue_fields"])
}
