package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- SearchParams ----

func TestUnitSearchParamsToMap(t *testing.T) {
	p := &SearchParams{
		AllowNoIndices:             ptrBool(true),
		AllowPartialSearchResults:  ptrBool(true),
		Analyzer:                   "standard",
		AnalyzeWildcard:            ptrBool(true),
		BatchedReduceSize:          ptrInt(64),
		CcsMinimizeRoundtrips:      ptrBool(true),
		DefaultOperator:            DefaultOperatorAnd,
		Df:                         "title",
		DocvalueFields:             []string{"d1", "d2"},
		ExpandWildcards:            ExpandWildcardOpen,
		Explain:                    ptrBool(true),
		From:                       ptrInt(10),
		IgnoreThrottled:            ptrBool(true),
		IgnoreUnavailable:          ptrBool(true),
		Lenient:                    ptrBool(true),
		MaxConcurrentShardRequests: ptrInt(5),
		MinCompatibleShardNode:     "7.10.0",
		PreFilterShardSize:         ptrInt(128),
		Preference:                 "_local",
		Q:                          "foo",
		RequestCache:               ptrBool(true),
		RestTotalHitsAsInt:         ptrBool(true),
		Routing:                    "r",
		Scroll:                     "1m",
		SearchType:                 SearchTypeQueryThenFetch,
		SeqNoPrimaryTerm:           ptrBool(true),
		Size:                       ptrInt(50),
		Sort:                       []string{"a:asc"},
		Source:                     "true",
		SourceExcludes:             []string{"x"},
		SourceIncludes:             []string{"y"},
		Stats:                      []string{"s1", "s2"},
		StoredFields:               []string{"f"},
		SuggestField:               "sug",
		SuggestMode:                SuggestModeAlways,
		SuggestSize:                ptrInt(3),
		SuggestText:                "txt",
		TerminateAfter:             ptrInt(100),
		Timeout:                    "30s",
		TrackScores:                ptrBool(true),
		TrackTotalHits:             "true",
		TypedKeys:                  ptrBool(true),
		Version:                    ptrBool(true),
		Common:                     &CommonParams{Pretty: ptrBool(true)},
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "true", m["allow_partial_search_results"])
	assert.Equal(t, "standard", m["analyzer"])
	assert.Equal(t, "true", m["analyze_wildcard"])
	assert.Equal(t, "64", m["batched_reduce_size"])
	assert.Equal(t, "true", m["ccs_minimize_roundtrips"])
	assert.Equal(t, "AND", m["default_operator"])
	assert.Equal(t, "title", m["df"])
	assert.Equal(t, "d1,d2", m["docvalue_fields"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["explain"])
	assert.Equal(t, "10", m["from"])
	assert.Equal(t, "true", m["ignore_throttled"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["lenient"])
	assert.Equal(t, "5", m["max_concurrent_shard_requests"])
	assert.Equal(t, "7.10.0", m["min_compatible_shard_node"])
	assert.Equal(t, "128", m["pre_filter_shard_size"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "foo", m["q"])
	assert.Equal(t, "true", m["request_cache"])
	assert.Equal(t, "true", m["rest_total_hits_as_int"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "1m", m["scroll"])
	assert.Equal(t, "query_then_fetch", m["search_type"])
	assert.Equal(t, "true", m["seq_no_primary_term"])
	assert.Equal(t, "50", m["size"])
	assert.Equal(t, "a:asc", m["sort"])
	assert.Equal(t, "true", m["_source"])
	assert.Equal(t, "x", m["_source_excludes"])
	assert.Equal(t, "y", m["_source_includes"])
	assert.Equal(t, "s1,s2", m["stats"])
	assert.Equal(t, "f", m["stored_fields"])
	assert.Equal(t, "sug", m["suggest_field"])
	assert.Equal(t, "always", m["suggest_mode"])
	assert.Equal(t, "3", m["suggest_size"])
	assert.Equal(t, "txt", m["suggest_text"])
	assert.Equal(t, "100", m["terminate_after"])
	assert.Equal(t, "30s", m["timeout"])
	assert.Equal(t, "true", m["track_scores"])
	assert.Equal(t, "true", m["track_total_hits"])
	assert.Equal(t, "true", m["typed_keys"])
	assert.Equal(t, "true", m["version"])
	assert.Equal(t, "true", m["pretty"])
}

func TestUnitSearchParamsToMapNil(t *testing.T) {
	var p *SearchParams
	assert.Nil(t, p.ToMap())
}

func TestUnitSearchParamsToMapEmpty(t *testing.T) {
	assert.Empty(t, (&SearchParams{}).ToMap())
}

// ---- SearchMVTParams ----

func TestUnitSearchMVTParamsToMap(t *testing.T) {
	p := &SearchMVTParams{
		ExactBounds:    ptrBool(true),
		Extent:         ptrInt(4096),
		GridAgg:        "grid",
		GridPrecision:  ptrInt(6),
		GridType:       "hex",
		Size:           ptrInt(1000),
		TrackTotalHits: "true",
		WithLabels:     ptrBool(true),
		Common:         &CommonParams{Pretty: ptrBool(true)},
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["exact_bounds"])
	assert.Equal(t, "4096", m["extent"])
	assert.Equal(t, "grid", m["grid_agg"])
	assert.Equal(t, "6", m["grid_precision"])
	assert.Equal(t, "hex", m["grid_type"])
	assert.Equal(t, "1000", m["size"])
	assert.Equal(t, "true", m["track_total_hits"])
	assert.Equal(t, "true", m["with_labels"])
	assert.Equal(t, "true", m["pretty"])
}

func TestUnitSearchMVTParamsToMapNil(t *testing.T) {
	var p *SearchMVTParams
	assert.Nil(t, p.ToMap())
}

func TestUnitSearchMVTRequestValidate(t *testing.T) {
	require.NoError(t, (&SearchMVTRequest{
		Index: "idx", Field: "geo", Zoom: ptrInt(5), X: ptrInt(1), Y: ptrInt(2),
	}).Validate())
}

func TestUnitSearchMVTRequestValidateMissingIndex(t *testing.T) {
	err := (&SearchMVTRequest{Field: "geo", Zoom: ptrInt(5), X: ptrInt(1), Y: ptrInt(2)}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Index")
}

func TestUnitSearchMVTRequestValidateMissingZoom(t *testing.T) {
	err := (&SearchMVTRequest{Index: "idx", Field: "geo", X: ptrInt(1), Y: ptrInt(2)}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Zoom")
}

// ---- MultiSearchParams ----

func TestUnitMultiSearchParamsToMap(t *testing.T) {
	p := &MultiSearchParams{
		CcsMinimizeRoundtrips:      ptrBool(true),
		MaxConcurrentSearches:      ptrInt(5),
		MaxConcurrentShardRequests: ptrInt(10),
		PreFilterShardSize:         ptrInt(128),
		RestTotalHitsAsInt:         ptrBool(true),
		SearchType:                 SearchTypeDfsQueryThenFetch,
		TypedKeys:                  ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["ccs_minimize_roundtrips"])
	assert.Equal(t, "5", m["max_concurrent_searches"])
	assert.Equal(t, "10", m["max_concurrent_shard_requests"])
	assert.Equal(t, "128", m["pre_filter_shard_size"])
	assert.Equal(t, "true", m["rest_total_hits_as_int"])
	assert.Equal(t, "dfs_query_then_fetch", m["search_type"])
	assert.Equal(t, "true", m["typed_keys"])
}

func TestUnitMultiSearchParamsToMapNil(t *testing.T) {
	var p *MultiSearchParams
	assert.Nil(t, p.ToMap())
}

func TestUnitMultiSearchRequestValidate(t *testing.T) {
	require.NoError(t, (&MultiSearchRequest{Body: "{}\n"}).Validate())
}

func TestUnitMultiSearchRequestValidateMissingBody(t *testing.T) {
	require.Error(t, (&MultiSearchRequest{}).Validate())
}

// ---- SearchTemplateParams ----

func TestUnitSearchTemplateParamsToMap(t *testing.T) {
	p := &SearchTemplateParams{
		AllowNoIndices:        ptrBool(true),
		CcsMinimizeRoundtrips: ptrBool(true),
		Explain:               ptrBool(true),
		IgnoreThrottled:       ptrBool(true),
		IgnoreUnavailable:     ptrBool(true),
		Preference:            "_local",
		Profile:               ptrBool(true),
		RestTotalHitsAsInt:    ptrBool(true),
		Routing:               "r",
		Scroll:                "1m",
		SearchType:            SearchTypeQueryThenFetch,
		TypedKeys:             ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "true", m["ccs_minimize_roundtrips"])
	assert.Equal(t, "true", m["explain"])
	assert.Equal(t, "true", m["ignore_throttled"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "true", m["profile"])
	assert.Equal(t, "true", m["rest_total_hits_as_int"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "1m", m["scroll"])
	assert.Equal(t, "query_then_fetch", m["search_type"])
	assert.Equal(t, "true", m["typed_keys"])
}

func TestUnitSearchTemplateParamsToMapNil(t *testing.T) {
	var p *SearchTemplateParams
	assert.Nil(t, p.ToMap())
}

func TestUnitSearchTemplateRequestValidate(t *testing.T) {
	require.NoError(t, (&SearchTemplateRequest{Body: map[string]any{}}).Validate())
}

func TestUnitSearchTemplateRequestValidateMissing(t *testing.T) {
	require.Error(t, (&SearchTemplateRequest{}).Validate())
}

// ---- MultiSearchTemplateParams ----

func TestUnitMultiSearchTemplateParamsToMap(t *testing.T) {
	p := &MultiSearchTemplateParams{
		CcsMinimizeRoundtrips: ptrBool(true),
		MaxConcurrentSearches: ptrInt(5),
		RestTotalHitsAsInt:    ptrBool(true),
		SearchType:            SearchTypeQueryThenFetch,
		TypedKeys:             ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["ccs_minimize_roundtrips"])
	assert.Equal(t, "5", m["max_concurrent_searches"])
	assert.Equal(t, "true", m["rest_total_hits_as_int"])
	assert.Equal(t, "query_then_fetch", m["search_type"])
	assert.Equal(t, "true", m["typed_keys"])
}

func TestUnitMultiSearchTemplateParamsToMapNil(t *testing.T) {
	var p *MultiSearchTemplateParams
	assert.Nil(t, p.ToMap())
}

func TestUnitMultiSearchTemplateRequestValidate(t *testing.T) {
	require.NoError(t, (&MultiSearchTemplateRequest{Body: "{}\n"}).Validate())
}

func TestUnitMultiSearchTemplateRequestValidateMissing(t *testing.T) {
	require.Error(t, (&MultiSearchTemplateRequest{}).Validate())
}

// ---- RankEvalParams ----

func TestUnitRankEvalParamsToMap(t *testing.T) {
	p := &RankEvalParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
		SearchType:        SearchTypeQueryThenFetch,
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "query_then_fetch", m["search_type"])
}

func TestUnitRankEvalParamsToMapNil(t *testing.T) {
	var p *RankEvalParams
	assert.Nil(t, p.ToMap())
}

func TestUnitRankEvalRequestValidate(t *testing.T) {
	require.NoError(t, (&RankEvalRequest{Body: map[string]any{}}).Validate())
}

func TestUnitRankEvalRequestValidateMissing(t *testing.T) {
	require.Error(t, (&RankEvalRequest{}).Validate())
}

// ---- CountParams ----

func TestUnitCountParamsToMap(t *testing.T) {
	p := &CountParams{
		AllowNoIndices:    ptrBool(true),
		Analyzer:          "standard",
		AnalyzeWildcard:   ptrBool(true),
		DefaultOperator:   DefaultOperatorAnd,
		Df:                "title",
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreThrottled:   ptrBool(true),
		IgnoreUnavailable: ptrBool(true),
		Lenient:           ptrBool(true),
		MinScore:          ptrFloat64(1.5),
		Preference:        "_local",
		Q:                 "foo",
		Routing:           "r",
		TerminateAfter:    ptrInt(100),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "standard", m["analyzer"])
	assert.Equal(t, "true", m["analyze_wildcard"])
	assert.Equal(t, "AND", m["default_operator"])
	assert.Equal(t, "title", m["df"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_throttled"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["lenient"])
	assert.Equal(t, "1.5", m["min_score"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "foo", m["q"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "100", m["terminate_after"])
}

func TestUnitCountParamsToMapNil(t *testing.T) {
	var p *CountParams
	assert.Nil(t, p.ToMap())
}

// ---- ScrollParams ----

func TestUnitScrollParamsToMap(t *testing.T) {
	p := &ScrollParams{
		Scroll:             "1m",
		ScrollId:           "abc",
		RestTotalHitsAsInt: ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "1m", m["scroll"])
	assert.Equal(t, "abc", m["scroll_id"])
	assert.Equal(t, "true", m["rest_total_hits_as_int"])
}

func TestUnitScrollParamsToMapNil(t *testing.T) {
	var p *ScrollParams
	assert.Nil(t, p.ToMap())
}

// ---- ValidateQueryParams ----

func TestUnitValidateQueryParamsToMap(t *testing.T) {
	p := &ValidateQueryParams{
		AllShards:         ptrBool(true),
		AllowNoIndices:    ptrBool(true),
		Analyzer:          "standard",
		AnalyzeWildcard:   ptrBool(true),
		DefaultOperator:   DefaultOperatorAnd,
		Df:                "title",
		ExpandWildcards:   ExpandWildcardOpen,
		Explain:           ptrBool(true),
		IgnoreUnavailable: ptrBool(true),
		Lenient:           ptrBool(true),
		Q:                 "foo",
		Rewrite:           ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["all_shards"])
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "standard", m["analyzer"])
	assert.Equal(t, "true", m["analyze_wildcard"])
	assert.Equal(t, "AND", m["default_operator"])
	assert.Equal(t, "title", m["df"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["explain"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["lenient"])
	assert.Equal(t, "foo", m["q"])
	assert.Equal(t, "true", m["rewrite"])
}

func TestUnitValidateQueryParamsToMapNil(t *testing.T) {
	var p *ValidateQueryParams
	assert.Nil(t, p.ToMap())
}

// ---- SearchShardsParams ----

func TestUnitSearchShardsParamsToMap(t *testing.T) {
	p := &SearchShardsParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
		Local:             ptrBool(true),
		Preference:        "_local",
		Routing:           "r",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["local"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "r", m["routing"])
}

func TestUnitSearchShardsParamsToMapNil(t *testing.T) {
	var p *SearchShardsParams
	assert.Nil(t, p.ToMap())
}

// ---- FieldCapsParams ----

func TestUnitFieldCapsParamsToMap(t *testing.T) {
	p := &FieldCapsParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		Fields:            []string{"f1", "f2"},
		IgnoreUnavailable: ptrBool(true),
		IncludeUnmapped:   ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "f1,f2", m["fields"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["include_unmapped"])
}

func TestUnitFieldCapsParamsToMapNil(t *testing.T) {
	var p *FieldCapsParams
	assert.Nil(t, p.ToMap())
}

// ---- OpenPITParams ----

func TestUnitOpenPITParamsToMap(t *testing.T) {
	p := &OpenPITParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
		KeepAlive:         "1m",
		Preference:        "_local",
		Routing:           "r",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "1m", m["keep_alive"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "r", m["routing"])
}

func TestUnitOpenPITParamsToMapNil(t *testing.T) {
	var p *OpenPITParams
	assert.Nil(t, p.ToMap())
}

func TestUnitOpenPITRequestValidate(t *testing.T) {
	require.NoError(t, (&OpenPITRequest{Indices: []string{"idx"}}).Validate())
}

func TestUnitOpenPITRequestValidateMissingIndices(t *testing.T) {
	err := (&OpenPITRequest{}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Indices")
}

// ---- AsyncSearchSubmitParams ----

func TestUnitAsyncSearchSubmitParamsToMap(t *testing.T) {
	p := &AsyncSearchSubmitParams{
		AllowNoIndices:            ptrBool(true),
		AllowPartialSearchResults: ptrBool(true),
		Analyzer:                  "standard",
		AnalyzeWildcard:           ptrBool(true),
		BatchedReduceSize:         ptrInt(64),
		DefaultOperator:           DefaultOperatorAnd,
		Df:                        "title",
		ExpandWildcards:           ExpandWildcardOpen,
		Explain:                   ptrBool(true),
		From:                      ptrInt(10),
		IgnoreThrottled:           ptrBool(true),
		IgnoreUnavailable:         ptrBool(true),
		KeepAlive:                 "1m",
		KeepOnCompletion:          ptrBool(true),
		Lenient:                   ptrBool(true),
		Preference:                "_local",
		Q:                         "foo",
		RequestCache:              ptrBool(true),
		Routing:                   "r",
		SearchType:                SearchTypeQueryThenFetch,
		Size:                      ptrInt(50),
		Sort:                      []string{"a:asc"},
		Source:                    "true",
		SourceExcludes:            []string{"x"},
		SourceIncludes:            []string{"y"},
		Stats:                     []string{"s1"},
		StoredFields:              []string{"f"},
		SuggestField:              "sug",
		SuggestMode:               SuggestModeAlways,
		SuggestSize:               ptrInt(3),
		SuggestText:               "txt",
		TerminateAfter:            ptrInt(100),
		Timeout:                   "30s",
		TrackScores:               ptrBool(true),
		TrackTotalHits:            "true",
		TypedKeys:                 ptrBool(true),
		WaitForCompletionTimeout:  "10s",
		Common:                    &CommonParams{Pretty: ptrBool(true)},
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "true", m["allow_partial_search_results"])
	assert.Equal(t, "standard", m["analyzer"])
	assert.Equal(t, "true", m["analyze_wildcard"])
	assert.Equal(t, "64", m["batched_reduce_size"])
	assert.Equal(t, "AND", m["default_operator"])
	assert.Equal(t, "title", m["df"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["explain"])
	assert.Equal(t, "10", m["from"])
	assert.Equal(t, "true", m["ignore_throttled"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "1m", m["keep_alive"])
	assert.Equal(t, "true", m["keep_on_completion"])
	assert.Equal(t, "true", m["lenient"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "foo", m["q"])
	assert.Equal(t, "true", m["request_cache"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "query_then_fetch", m["search_type"])
	assert.Equal(t, "50", m["size"])
	assert.Equal(t, "a:asc", m["sort"])
	assert.Equal(t, "true", m["_source"])
	assert.Equal(t, "x", m["_source_excludes"])
	assert.Equal(t, "y", m["_source_includes"])
	assert.Equal(t, "s1", m["stats"])
	assert.Equal(t, "f", m["stored_fields"])
	assert.Equal(t, "sug", m["suggest_field"])
	assert.Equal(t, "always", m["suggest_mode"])
	assert.Equal(t, "3", m["suggest_size"])
	assert.Equal(t, "txt", m["suggest_text"])
	assert.Equal(t, "100", m["terminate_after"])
	assert.Equal(t, "30s", m["timeout"])
	assert.Equal(t, "true", m["track_scores"])
	assert.Equal(t, "true", m["track_total_hits"])
	assert.Equal(t, "true", m["typed_keys"])
	assert.Equal(t, "10s", m["wait_for_completion_timeout"])
	assert.Equal(t, "true", m["pretty"])
}

func TestUnitAsyncSearchSubmitParamsToMapNil(t *testing.T) {
	var p *AsyncSearchSubmitParams
	assert.Nil(t, p.ToMap())
}

// ---- AsyncSearchGetParams ----

func TestUnitAsyncSearchGetParamsToMap(t *testing.T) {
	p := &AsyncSearchGetParams{
		KeepAlive:                "1m",
		TypedKeys:                ptrBool(true),
		WaitForCompletionTimeout: "10s",
	}
	m := p.ToMap()
	assert.Equal(t, "1m", m["keep_alive"])
	assert.Equal(t, "true", m["typed_keys"])
	assert.Equal(t, "10s", m["wait_for_completion_timeout"])
}

func TestUnitAsyncSearchGetParamsToMapNil(t *testing.T) {
	var p *AsyncSearchGetParams
	assert.Nil(t, p.ToMap())
}
