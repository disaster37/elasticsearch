package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrFloat64(f float64) *float64 { return &f }

// ---- IndexParams ----

func TestUnitIndexParamsToMap(t *testing.T) {
	p := &IndexParams{
		IfPrimaryTerm:       ptrInt64(2),
		IfSeqNo:             ptrInt64(10),
		OpType:              OpTypeCreate,
		Pipeline:            "my-pipe",
		Refresh:             RefreshTrue,
		RequireAlias:        ptrBool(true),
		Timeout:             "30s",
		Version:             ptrInt64(5),
		VersionType:         VersionTypeExternal,
		WaitForActiveShards: "1",
		Common:              &CommonParams{Pretty: ptrBool(true)},
	}
	m := p.ToMap()
	assert.Equal(t, "2", m["if_primary_term"])
	assert.Equal(t, "10", m["if_seq_no"])
	assert.Equal(t, "create", m["op_type"])
	assert.Equal(t, "my-pipe", m["pipeline"])
	assert.Equal(t, "true", m["refresh"])
	assert.Equal(t, "true", m["require_alias"])
	assert.Equal(t, "30s", m["timeout"])
	assert.Equal(t, "5", m["version"])
	assert.Equal(t, "external", m["version_type"])
	assert.Equal(t, "1", m["wait_for_active_shards"])
	assert.Equal(t, "true", m["pretty"])
}

func TestUnitIndexParamsToMapNil(t *testing.T) {
	var p *IndexParams
	assert.Nil(t, p.ToMap())
}

func TestUnitIndexParamsToMapEmpty(t *testing.T) {
	m := (&IndexParams{}).ToMap()
	assert.Empty(t, m)
}

func TestUnitIndexParamsToMapFalseBool(t *testing.T) {
	p := &IndexParams{RequireAlias: ptrBool(false)}
	m := p.ToMap()
	assert.Equal(t, "false", m["require_alias"])
}

func TestUnitIndexRequestValidate(t *testing.T) {
	require.NoError(t, (&IndexRequest{Index: "idx", Id: "1"}).Validate())
}

func TestUnitIndexRequestValidateMissingIndex(t *testing.T) {
	err := (&IndexRequest{Id: "1"}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Index")
}

func TestUnitIndexRequestValidateControlChars(t *testing.T) {
	err := (&IndexRequest{Index: "idx", Id: "bad\x00id"}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "control characters")
}

// ---- CreateParams (= IndexParams) ----

func TestUnitCreateParamsToMap(t *testing.T) {
	p := &CreateParams{
		Pipeline: "p",
		Refresh:  RefreshWaitFor,
	}
	m := p.ToMap()
	assert.Equal(t, "p", m["pipeline"])
	assert.Equal(t, "wait_for", m["refresh"])
}

func TestUnitCreateRequestValidate(t *testing.T) {
	require.NoError(t, (&CreateRequest{Index: "idx", Id: "1", Body: map[string]any{}}).Validate())
}

func TestUnitCreateRequestValidateMissing(t *testing.T) {
	err := (&CreateRequest{Index: "idx"}).Validate()
	require.Error(t, err)
}

// ---- GetParams ----

func TestUnitGetParamsToMap(t *testing.T) {
	p := &GetParams{
		ForceSyntheticSource: ptrBool(true),
		Preference:           "_local",
		Realtime:             ptrBool(false),
		Refresh:              ptrBool(true),
		Routing:              "r1",
		Source:               "false",
		SourceExcludes:       []string{"a", "b"},
		SourceIncludes:       []string{"c"},
		StoredFields:         []string{"f1", "f2"},
		Version:              ptrInt64(3),
		VersionType:          VersionTypeInternal,
		Common:               &CommonParams{Human: ptrBool(true)},
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["force_synthetic_source"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "false", m["realtime"])
	assert.Equal(t, "true", m["refresh"])
	assert.Equal(t, "r1", m["routing"])
	assert.Equal(t, "false", m["_source"])
	assert.Equal(t, "a,b", m["_source_excludes"])
	assert.Equal(t, "c", m["_source_includes"])
	assert.Equal(t, "f1,f2", m["stored_fields"])
	assert.Equal(t, "3", m["version"])
	assert.Equal(t, "internal", m["version_type"])
	assert.Equal(t, "true", m["human"])
}

func TestUnitGetParamsToMapNil(t *testing.T) {
	var p *GetParams
	assert.Nil(t, p.ToMap())
}

func TestUnitGetRequestValidate(t *testing.T) {
	require.NoError(t, (&GetRequest{Index: "idx", Id: "1"}).Validate())
}

func TestUnitGetRequestValidateMissing(t *testing.T) {
	err := (&GetRequest{Index: "idx"}).Validate()
	require.Error(t, err)
}

func TestUnitGetRequestValidateControlChars(t *testing.T) {
	err := (&GetRequest{Index: "idx\x01", Id: "1"}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "control characters")
}

// ---- GetSourceParams (= GetParams) ----

func TestUnitGetSourceParamsToMap(t *testing.T) {
	p := &GetSourceParams{Routing: "r", Refresh: ptrBool(true)}
	m := p.ToMap()
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "true", m["refresh"])
}

func TestUnitGetSourceRequestValidate(t *testing.T) {
	require.NoError(t, (&GetSourceRequest{Index: "idx", Id: "1"}).Validate())
}

func TestUnitGetSourceRequestValidateMissing(t *testing.T) {
	require.Error(t, (&GetSourceRequest{Index: "idx"}).Validate())
}

// ---- DeleteParams ----

func TestUnitDeleteParamsToMap(t *testing.T) {
	p := &DeleteParams{
		IfPrimaryTerm: ptrInt64(1),
		IfSeqNo:       ptrInt64(2),
		Refresh:       RefreshFalse,
		Routing:       "r",
		Timeout:       "5s",
		Version:       ptrInt64(4),
		VersionType:   VersionTypeExternalGte,
		Common:        &CommonParams{ErrorTrace: ptrBool(true)},
	}
	m := p.ToMap()
	assert.Equal(t, "1", m["if_primary_term"])
	assert.Equal(t, "2", m["if_seq_no"])
	assert.Equal(t, "false", m["refresh"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "5s", m["timeout"])
	assert.Equal(t, "4", m["version"])
	assert.Equal(t, "external_gte", m["version_type"])
	assert.Equal(t, "true", m["error_trace"])
}

func TestUnitDeleteParamsToMapNil(t *testing.T) {
	var p *DeleteParams
	assert.Nil(t, p.ToMap())
}

func TestUnitDeleteRequestValidate(t *testing.T) {
	require.NoError(t, (&DeleteRequest{Index: "idx", Id: "1"}).Validate())
}

func TestUnitDeleteRequestValidateMissing(t *testing.T) {
	require.Error(t, (&DeleteRequest{Index: "idx"}).Validate())
}

func TestUnitDeleteRequestValidateControlChars(t *testing.T) {
	err := (&DeleteRequest{Index: "idx", Id: "bad\x7F"}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "control characters")
}

// ---- UpdateParams ----

func TestUnitUpdateParamsToMap(t *testing.T) {
	p := &UpdateParams{
		IfPrimaryTerm:       ptrInt64(1),
		IfSeqNo:             ptrInt64(2),
		Lang:                "painless",
		Refresh:             RefreshTrue,
		RequireAlias:        ptrBool(false),
		RetryOnConflict:     ptrInt(3),
		Routing:             "r",
		Source:              "true",
		SourceExcludes:      []string{"x"},
		SourceIncludes:      []string{"y"},
		Timeout:             "10s",
		WaitForActiveShards: "all",
		Common:              &CommonParams{FilterPath: []string{"a", "b"}},
	}
	m := p.ToMap()
	assert.Equal(t, "1", m["if_primary_term"])
	assert.Equal(t, "2", m["if_seq_no"])
	assert.Equal(t, "painless", m["lang"])
	assert.Equal(t, "true", m["refresh"])
	assert.Equal(t, "false", m["require_alias"])
	assert.Equal(t, "3", m["retry_on_conflict"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "true", m["_source"])
	assert.Equal(t, "x", m["_source_excludes"])
	assert.Equal(t, "y", m["_source_includes"])
	assert.Equal(t, "10s", m["timeout"])
	assert.Equal(t, "all", m["wait_for_active_shards"])
	assert.Equal(t, "a,b", m["filter_path"])
}

func TestUnitUpdateParamsToMapNil(t *testing.T) {
	var p *UpdateParams
	assert.Nil(t, p.ToMap())
}

func TestUnitUpdateRequestValidate(t *testing.T) {
	require.NoError(t, (&UpdateRequest{Index: "idx", Id: "1"}).Validate())
}

func TestUnitUpdateRequestValidateMissing(t *testing.T) {
	require.Error(t, (&UpdateRequest{Index: "idx"}).Validate())
}

// ---- MgetParams ----

func TestUnitMgetParamsToMap(t *testing.T) {
	p := &MgetParams{
		ForceSyntheticSource: ptrBool(true),
		Preference:           "_only_node:abc",
		Realtime:             ptrBool(false),
		Refresh:              ptrBool(true),
		Routing:              "r",
		Source:               "true",
		SourceExcludes:       []string{"a"},
		SourceIncludes:       []string{"b"},
		StoredFields:         []string{"f"},
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["force_synthetic_source"])
	assert.Equal(t, "_only_node:abc", m["preference"])
	assert.Equal(t, "false", m["realtime"])
	assert.Equal(t, "true", m["refresh"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "true", m["_source"])
	assert.Equal(t, "a", m["_source_excludes"])
	assert.Equal(t, "b", m["_source_includes"])
	assert.Equal(t, "f", m["stored_fields"])
}

func TestUnitMgetParamsToMapNil(t *testing.T) {
	var p *MgetParams
	assert.Nil(t, p.ToMap())
}

func TestUnitMgetRequestValidate(t *testing.T) {
	require.NoError(t, (&MgetRequest{Body: map[string]any{}}).Validate())
}

func TestUnitMgetRequestValidateMissingBody(t *testing.T) {
	require.Error(t, (&MgetRequest{}).Validate())
}

// ---- BulkParams ----

func TestUnitBulkParamsToMap(t *testing.T) {
	p := &BulkParams{
		ListExecutedPipelines: []string{"p1", "p2"},
		Pipeline:              "pipe",
		Refresh:               RefreshTrue,
		RequireAlias:          ptrBool(true),
		Routing:               "r",
		Source:                "false",
		SourceExcludes:        []string{"a"},
		SourceIncludes:        []string{"b"},
		Timeout:               "30s",
		WaitForActiveShards:   "1",
	}
	m := p.ToMap()
	assert.Equal(t, "p1,p2", m["list_executed_pipelines"])
	assert.Equal(t, "pipe", m["pipeline"])
	assert.Equal(t, "true", m["refresh"])
	assert.Equal(t, "true", m["require_alias"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "false", m["_source"])
	assert.Equal(t, "a", m["_source_excludes"])
	assert.Equal(t, "b", m["_source_includes"])
	assert.Equal(t, "30s", m["timeout"])
	assert.Equal(t, "1", m["wait_for_active_shards"])
}

func TestUnitBulkParamsToMapNil(t *testing.T) {
	var p *BulkParams
	assert.Nil(t, p.ToMap())
}

func TestUnitBulkRequestValidate(t *testing.T) {
	require.NoError(t, (&BulkRequest{Body: "{}\n"}).Validate())
}

func TestUnitBulkRequestValidateMissingBody(t *testing.T) {
	require.Error(t, (&BulkRequest{}).Validate())
}

// ---- DeleteByQueryParams ----

func TestUnitDeleteByQueryParamsToMap(t *testing.T) {
	p := &DeleteByQueryParams{
		Analyzer:            "standard",
		AnalyzeWildcard:     ptrBool(true),
		Conflicts:           ConflictsProceed,
		DefaultOperator:     DefaultOperatorAnd,
		Df:                  "title",
		ExpandWildcards:     ExpandWildcardOpen,
		From:                ptrInt(5),
		IgnoreUnavailable:   ptrBool(true),
		Lenient:             ptrBool(false),
		MaxDocs:             ptrInt(100),
		Preference:          "_local",
		Q:                   "foo",
		Refresh:              ptrBool(true),
		RequestCache:        ptrBool(false),
		RequestsPerSecond:   ptrFloat64(10.5),
		Routing:             "r",
		Scroll:              "1m",
		ScrollSize:          ptrInt(50),
		SearchTimeout:       "30s",
		SearchType:          SearchTypeQueryThenFetch,
		SliceId:             "0",
		SliceMax:            ptrInt(2),
		Slices:              "auto",
		Sort:                []string{"a:asc", "b:desc"},
		Stats:               []string{"s1"},
		TerminateAfter:      ptrInt(1000),
		Timeout:             "1m",
		Version:             ptrBool(true),
		WaitForActiveShards: "1",
		WaitForCompletion:   ptrBool(true),
		Common:              &CommonParams{Pretty: ptrBool(true)},
	}
	m := p.ToMap()
	assert.Equal(t, "standard", m["analyzer"])
	assert.Equal(t, "true", m["analyze_wildcard"])
	assert.Equal(t, "proceed", m["conflicts"])
	assert.Equal(t, "AND", m["default_operator"])
	assert.Equal(t, "title", m["df"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "5", m["from"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "false", m["lenient"])
	assert.Equal(t, "100", m["max_docs"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "foo", m["q"])
	assert.Equal(t, "true", m["refresh"])
	assert.Equal(t, "false", m["request_cache"])
	assert.Equal(t, "10.5", m["requests_per_second"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "1m", m["scroll"])
	assert.Equal(t, "50", m["scroll_size"])
	assert.Equal(t, "30s", m["search_timeout"])
	assert.Equal(t, "query_then_fetch", m["search_type"])
	assert.Equal(t, "0", m["slice_id"])
	assert.Equal(t, "2", m["slice_max"])
	assert.Equal(t, "auto", m["slices"])
	assert.Equal(t, "a:asc,b:desc", m["sort"])
	assert.Equal(t, "s1", m["stats"])
	assert.Equal(t, "1000", m["terminate_after"])
	assert.Equal(t, "1m", m["timeout"])
	assert.Equal(t, "true", m["version"])
	assert.Equal(t, "1", m["wait_for_active_shards"])
	assert.Equal(t, "true", m["wait_for_completion"])
	assert.Equal(t, "true", m["pretty"])
}

func TestUnitDeleteByQueryParamsToMapNil(t *testing.T) {
	var p *DeleteByQueryParams
	assert.Nil(t, p.ToMap())
}

func TestUnitDeleteByQueryRequestValidate(t *testing.T) {
	require.NoError(t, (&DeleteByQueryRequest{Indices: []string{"idx"}}).Validate())
}

func TestUnitDeleteByQueryRequestValidateMissingIndices(t *testing.T) {
	err := (&DeleteByQueryRequest{}).Validate()
	require.Error(t, err)
}

// ---- UpdateByQueryParams ----

func TestUnitUpdateByQueryParamsToMap(t *testing.T) {
	p := &UpdateByQueryParams{
		Analyzer:          "standard",
		AnalyzeWildcard:   ptrBool(true),
		Conflicts:         ConflictsAbort,
		DefaultOperator:   DefaultOperatorOr,
		Df:                 "title",
		ExpandWildcards:   ExpandWildcardAll,
		From:              ptrInt(1),
		IgnoreUnavailable: ptrBool(true),
		Lenient:           ptrBool(true),
		MaxDocs:           ptrInt(2),
		Pipeline:          "pipe",
		Preference:        "_local",
		Q:                 "bar",
		Refresh:           ptrBool(true),
		RequestCache:      ptrBool(true),
		RequestsPerSecond: ptrFloat64(1),
		Routing:           "r",
		Scroll:            "1m",
		ScrollSize:        ptrInt(10),
		SearchTimeout:    "5s",
		SearchType:       SearchTypeDfsQueryThenFetch,
		SliceId:          "1",
		SliceMax:         ptrInt(3),
		Slices:           "2",
		Sort:             []string{"x"},
		Stats:            []string{"y"},
		TerminateAfter:   ptrInt(5),
		Timeout:          "1m",
		Version:          ptrBool(false),
		VersionType:      VersionTypeInternal,
		WaitForActiveShards: "all",
		WaitForCompletion:   ptrBool(false),
	}
	m := p.ToMap()
	assert.Equal(t, "standard", m["analyzer"])
	assert.Equal(t, "true", m["analyze_wildcard"])
	assert.Equal(t, "abort", m["conflicts"])
	assert.Equal(t, "OR", m["default_operator"])
	assert.Equal(t, "title", m["df"])
	assert.Equal(t, "all", m["expand_wildcards"])
	assert.Equal(t, "1", m["from"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["lenient"])
	assert.Equal(t, "2", m["max_docs"])
	assert.Equal(t, "pipe", m["pipeline"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "bar", m["q"])
	assert.Equal(t, "true", m["refresh"])
	assert.Equal(t, "true", m["request_cache"])
	assert.Equal(t, "1", m["requests_per_second"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "1m", m["scroll"])
	assert.Equal(t, "10", m["scroll_size"])
	assert.Equal(t, "5s", m["search_timeout"])
	assert.Equal(t, "dfs_query_then_fetch", m["search_type"])
	assert.Equal(t, "1", m["slice_id"])
	assert.Equal(t, "3", m["slice_max"])
	assert.Equal(t, "2", m["slices"])
	assert.Equal(t, "x", m["sort"])
	assert.Equal(t, "y", m["stats"])
	assert.Equal(t, "5", m["terminate_after"])
	assert.Equal(t, "1m", m["timeout"])
	assert.Equal(t, "false", m["version"])
	assert.Equal(t, "internal", m["version_type"])
	assert.Equal(t, "all", m["wait_for_active_shards"])
	assert.Equal(t, "false", m["wait_for_completion"])
}

func TestUnitUpdateByQueryParamsToMapNil(t *testing.T) {
	var p *UpdateByQueryParams
	assert.Nil(t, p.ToMap())
}

func TestUnitUpdateByQueryRequestValidate(t *testing.T) {
	require.NoError(t, (&UpdateByQueryRequest{Indices: []string{"idx"}}).Validate())
}

func TestUnitUpdateByQueryRequestValidateMissing(t *testing.T) {
	require.Error(t, (&UpdateByQueryRequest{}).Validate())
}

// ---- ReindexParams ----

func TestUnitReindexParamsToMap(t *testing.T) {
	p := &ReindexParams{
		MaxDocs:             ptrInt(100),
		Refresh:             ptrBool(true),
		RequestsPerSecond:   ptrFloat64(2.5),
		RequireAlias:        ptrBool(true),
		Scroll:              "5m",
		Slices:              "auto",
		Timeout:             "1m",
		WaitForActiveShards: "1",
		WaitForCompletion:   ptrBool(false),
	}
	m := p.ToMap()
	assert.Equal(t, "100", m["max_docs"])
	assert.Equal(t, "true", m["refresh"])
	assert.Equal(t, "2.5", m["requests_per_second"])
	assert.Equal(t, "true", m["require_alias"])
	assert.Equal(t, "5m", m["scroll"])
	assert.Equal(t, "auto", m["slices"])
	assert.Equal(t, "1m", m["timeout"])
	assert.Equal(t, "1", m["wait_for_active_shards"])
	assert.Equal(t, "false", m["wait_for_completion"])
}

func TestUnitReindexParamsToMapNil(t *testing.T) {
	var p *ReindexParams
	assert.Nil(t, p.ToMap())
}

// ---- ExplainParams ----

func TestUnitExplainParamsToMap(t *testing.T) {
	p := &ExplainParams{
		Analyzer:        "standard",
		AnalyzeWildcard: ptrBool(true),
		DefaultOperator: DefaultOperatorAnd,
		Df:              "title",
		Lenient:         ptrBool(true),
		Preference:      "_local",
		Q:               "foo",
		Routing:         "r",
		Source:          "true",
		SourceExcludes:  []string{"a"},
		SourceIncludes:  []string{"b"},
		StoredFields:    []string{"f"},
	}
	m := p.ToMap()
	assert.Equal(t, "standard", m["analyzer"])
	assert.Equal(t, "true", m["analyze_wildcard"])
	assert.Equal(t, "AND", m["default_operator"])
	assert.Equal(t, "title", m["df"])
	assert.Equal(t, "true", m["lenient"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "foo", m["q"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "true", m["_source"])
	assert.Equal(t, "a", m["_source_excludes"])
	assert.Equal(t, "b", m["_source_includes"])
	assert.Equal(t, "f", m["stored_fields"])
}

func TestUnitExplainParamsToMapNil(t *testing.T) {
	var p *ExplainParams
	assert.Nil(t, p.ToMap())
}

func TestUnitExplainRequestValidate(t *testing.T) {
	require.NoError(t, (&ExplainRequest{Index: "idx", Id: "1"}).Validate())
}

func TestUnitExplainRequestValidateMissing(t *testing.T) {
	require.Error(t, (&ExplainRequest{Index: "idx"}).Validate())
}

// ---- TermVectorsParams ----

func TestUnitTermVectorsParamsToMap(t *testing.T) {
	p := &TermVectorsParams{
		FieldStatistics: ptrBool(true),
		Fields:          []string{"title", "body"},
		Offsets:         ptrBool(true),
		Payloads:        ptrBool(false),
		Positions:       ptrBool(true),
		Preference:      "_local",
		Realtime:        ptrBool(true),
		Routing:         "r",
		TermStatistics:  ptrBool(true),
		Version:         ptrInt64(2),
		VersionType:     VersionTypeInternal,
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["field_statistics"])
	assert.Equal(t, "title,body", m["fields"])
	assert.Equal(t, "true", m["offsets"])
	assert.Equal(t, "false", m["payloads"])
	assert.Equal(t, "true", m["positions"])
	assert.Equal(t, "_local", m["preference"])
	assert.Equal(t, "true", m["realtime"])
	assert.Equal(t, "r", m["routing"])
	assert.Equal(t, "true", m["term_statistics"])
	assert.Equal(t, "2", m["version"])
	assert.Equal(t, "internal", m["version_type"])
}

func TestUnitTermVectorsParamsToMapNil(t *testing.T) {
	var p *TermVectorsParams
	assert.Nil(t, p.ToMap())
}

func TestUnitTermVectorsRequestValidate(t *testing.T) {
	require.NoError(t, (&TermVectorsRequest{Index: "idx"}).Validate())
}

func TestUnitTermVectorsRequestValidateMissing(t *testing.T) {
	require.Error(t, (&TermVectorsRequest{}).Validate())
}

// ---- MtermvectorsParams (= TermVectorsParams) ----

func TestUnitMtermvectorsParamsToMap(t *testing.T) {
	p := &MtermvectorsParams{Fields: []string{"a"}, Routing: "r"}
	m := p.ToMap()
	assert.Equal(t, "a", m["fields"])
	assert.Equal(t, "r", m["routing"])
}

// ---- TermsEnumParams ----

func TestUnitTermsEnumParamsToMap(t *testing.T) {
	p := &TermsEnumParams{CaseInsensitive: ptrBool(true)}
	m := p.ToMap()
	assert.Equal(t, "true", m["case_insensitive"])
}

func TestUnitTermsEnumParamsToMapNil(t *testing.T) {
	var p *TermsEnumParams
	assert.Nil(t, p.ToMap())
}

func TestUnitTermsEnumRequestValidate(t *testing.T) {
	require.NoError(t, (&TermsEnumRequest{Index: "idx"}).Validate())
}

func TestUnitTermsEnumRequestValidateMissing(t *testing.T) {
	require.Error(t, (&TermsEnumRequest{}).Validate())
}

// ---- RethrottleParams ----

func TestUnitRethrottleParamsToMap(t *testing.T) {
	p := &RethrottleParams{RequestsPerSecond: ptrFloat64(100)}
	m := p.ToMap()
	assert.Equal(t, "100", m["requests_per_second"])
}

func TestUnitRethrottleParamsToMapNil(t *testing.T) {
	var p *RethrottleParams
	assert.Nil(t, p.ToMap())
}

func TestUnitRethrottleRequestValidate(t *testing.T) {
	require.NoError(t, (&RethrottleRequest{TaskId: "task:1"}).Validate())
}

func TestUnitRethrottleRequestValidateMissing(t *testing.T) {
	require.Error(t, (&RethrottleRequest{}).Validate())
}
