package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- IndicesCreateParams ----

func TestUnitIndicesCreateParamsToMap(t *testing.T) {
	p := &IndicesCreateParams{WaitForActiveShards: "1", Common: &CommonParams{Pretty: ptrBool(true)}}
	m := p.ToMap()
	assert.Equal(t, "1", m["wait_for_active_shards"])
	assert.Equal(t, "true", m["pretty"])
}

func TestUnitIndicesCreateParamsToMapNil(t *testing.T) {
	var p *IndicesCreateParams
	assert.Nil(t, p.ToMap())
}

func TestUnitIndicesCreateParamsToMapEmpty(t *testing.T) {
	assert.Empty(t, (&IndicesCreateParams{}).ToMap())
}

// ---- IndicesDeleteParams ----

func TestUnitIndicesDeleteParamsToMap(t *testing.T) {
	p := &IndicesDeleteParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(false),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "false", m["ignore_unavailable"])
}

func TestUnitIndicesDeleteParamsToMapNil(t *testing.T) {
	var p *IndicesDeleteParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesGetParams ----

func TestUnitIndicesGetParamsToMap(t *testing.T) {
	p := &IndicesGetParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardAll,
		Features:          []string{"aliases", "mappings"},
		FlatSettings:      ptrBool(true),
		IgnoreUnavailable: ptrBool(true),
		IncludeDefaults:   ptrBool(true),
		Local:             ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "all", m["expand_wildcards"])
	assert.Equal(t, "aliases,mappings", m["features"])
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["include_defaults"])
	assert.Equal(t, "true", m["local"])
}

func TestUnitIndicesGetParamsToMapNil(t *testing.T) {
	var p *IndicesGetParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesExistsParams ----

func TestUnitIndicesExistsParamsToMap(t *testing.T) {
	p := &IndicesExistsParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardClosed,
		FlatSettings:      ptrBool(true),
		IgnoreUnavailable: ptrBool(true),
		IncludeDefaults:   ptrBool(true),
		Local:             ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "closed", m["expand_wildcards"])
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["include_defaults"])
	assert.Equal(t, "true", m["local"])
}

func TestUnitIndicesExistsParamsToMapNil(t *testing.T) {
	var p *IndicesExistsParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesOpenParams ----

func TestUnitIndicesOpenParamsToMap(t *testing.T) {
	p := &IndicesOpenParams{
		AllowNoIndices:      ptrBool(true),
		ExpandWildcards:     ExpandWildcardOpen,
		IgnoreUnavailable:   ptrBool(true),
		Timeout:             "30s",
		WaitForActiveShards: "1",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "30s", m["timeout"])
	assert.Equal(t, "1", m["wait_for_active_shards"])
}

func TestUnitIndicesOpenParamsToMapNil(t *testing.T) {
	var p *IndicesOpenParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesCloseParams (= IndicesOpenParams) ----

func TestUnitIndicesCloseParamsToMap(t *testing.T) {
	p := &IndicesCloseParams{Timeout: "10s", WaitForActiveShards: "all"}
	m := p.ToMap()
	assert.Equal(t, "10s", m["timeout"])
	assert.Equal(t, "all", m["wait_for_active_shards"])
}

// ---- RolloverParams ----

func TestUnitRolloverParamsToMap(t *testing.T) {
	p := &RolloverParams{
		DryRun:              ptrBool(true),
		Timeout:             "30s",
		WaitForActiveShards: "1",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["dry_run"])
	assert.Equal(t, "30s", m["timeout"])
	assert.Equal(t, "1", m["wait_for_active_shards"])
}

func TestUnitRolloverParamsToMapNil(t *testing.T) {
	var p *RolloverParams
	assert.Nil(t, p.ToMap())
}

func TestUnitRolloverRequestValidate(t *testing.T) {
	require.NoError(t, (&RolloverRequest{Alias: "a"}).Validate())
}

func TestUnitRolloverRequestValidateMissing(t *testing.T) {
	err := (&RolloverRequest{}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Alias")
}

// ---- ShrinkParams ----

func TestUnitShrinkParamsToMap(t *testing.T) {
	p := &ShrinkParams{
		Timeout:             "30s",
		WaitForActiveShards: "1",
	}
	m := p.ToMap()
	assert.Equal(t, "30s", m["timeout"])
	assert.Equal(t, "1", m["wait_for_active_shards"])
}

func TestUnitShrinkParamsToMapNil(t *testing.T) {
	var p *ShrinkParams
	assert.Nil(t, p.ToMap())
}

func TestUnitShrinkRequestValidate(t *testing.T) {
	require.NoError(t, (&ShrinkRequest{Source: "s", Target: "t"}).Validate())
}

func TestUnitShrinkRequestValidateMissing(t *testing.T) {
	err := (&ShrinkRequest{Source: "s"}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Target")
}

// ---- SplitParams (= ShrinkParams) ----

func TestUnitSplitParamsToMap(t *testing.T) {
	p := &SplitParams{Timeout: "5s"}
	m := p.ToMap()
	assert.Equal(t, "5s", m["timeout"])
}

func TestUnitSplitRequestValidate(t *testing.T) {
	require.NoError(t, (&SplitRequest{Source: "s", Target: "t"}).Validate())
}

func TestUnitSplitRequestValidateMissing(t *testing.T) {
	require.Error(t, (&SplitRequest{Source: "s"}).Validate())
}

// ---- CloneParams (= ShrinkParams) ----

func TestUnitCloneParamsToMap(t *testing.T) {
	p := &CloneParams{WaitForActiveShards: "all"}
	m := p.ToMap()
	assert.Equal(t, "all", m["wait_for_active_shards"])
}

func TestUnitCloneRequestValidate(t *testing.T) {
	require.NoError(t, (&CloneRequest{Source: "s", Target: "t"}).Validate())
}

func TestUnitCloneRequestValidateMissing(t *testing.T) {
	require.Error(t, (&CloneRequest{Source: "s"}).Validate())
}

// ---- IndicesRefreshParams ----

func TestUnitIndicesRefreshParamsToMap(t *testing.T) {
	p := &IndicesRefreshParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
}

func TestUnitIndicesRefreshParamsToMapNil(t *testing.T) {
	var p *IndicesRefreshParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesFlushParams ----

func TestUnitIndicesFlushParamsToMap(t *testing.T) {
	p := &IndicesFlushParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		Force:             ptrBool(true),
		IgnoreUnavailable: ptrBool(true),
		WaitIfOngoing:     ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["force"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["wait_if_ongoing"])
}

func TestUnitIndicesFlushParamsToMapNil(t *testing.T) {
	var p *IndicesFlushParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesForceMergeParams ----

func TestUnitIndicesForceMergeParamsToMap(t *testing.T) {
	p := &IndicesForceMergeParams{
		AllowNoIndices:     ptrBool(true),
		ExpandWildcards:    ExpandWildcardOpen,
		Flush:              ptrBool(true),
		IgnoreUnavailable:  ptrBool(true),
		MaxNumSegments:     ptrInt(1),
		OnlyExpungeDeletes: ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["flush"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "1", m["max_num_segments"])
	assert.Equal(t, "true", m["only_expunge_deletes"])
}

func TestUnitIndicesForceMergeParamsToMapNil(t *testing.T) {
	var p *IndicesForceMergeParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesClearCacheParams ----

func TestUnitIndicesClearCacheParamsToMap(t *testing.T) {
	p := &IndicesClearCacheParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		Fielddata:         ptrBool(true),
		Fields:            []string{"f1", "f2"},
		IgnoreUnavailable: ptrBool(true),
		Query:             ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["fielddata"])
	assert.Equal(t, "f1,f2", m["fields"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["query"])
}

func TestUnitIndicesClearCacheParamsToMapNil(t *testing.T) {
	var p *IndicesClearCacheParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesStatsParams ----

func TestUnitIndicesStatsParamsToMap(t *testing.T) {
	p := &IndicesStatsParams{
		CompletionFields:        []string{"c"},
		ExpandWildcards:         ExpandWildcardOpen,
		FielddataFields:         []string{"fd"},
		Fields:                  []string{"f"},
		ForbidClosedIndices:     ptrBool(true),
		Groups:                  []string{"g1", "g2"},
		IncludeSegmentFileSizes: ptrBool(true),
		IncludeUnloadedSegments: ptrBool(true),
		Level:                   LevelShards,
	}
	m := p.ToMap()
	assert.Equal(t, "c", m["completion_fields"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "fd", m["fielddata_fields"])
	assert.Equal(t, "f", m["fields"])
	assert.Equal(t, "true", m["forbid_closed_indices"])
	assert.Equal(t, "g1,g2", m["groups"])
	assert.Equal(t, "true", m["include_segment_file_sizes"])
	assert.Equal(t, "true", m["include_unloaded_segments"])
	assert.Equal(t, "shards", m["level"])
}

func TestUnitIndicesStatsParamsToMapNil(t *testing.T) {
	var p *IndicesStatsParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesSegmentsParams ----

func TestUnitIndicesSegmentsParamsToMap(t *testing.T) {
	p := &IndicesSegmentsParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
		Verbose:           ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["verbose"])
}

func TestUnitIndicesSegmentsParamsToMapNil(t *testing.T) {
	var p *IndicesSegmentsParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesRecoveryParams ----

func TestUnitIndicesRecoveryParamsToMap(t *testing.T) {
	p := &IndicesRecoveryParams{
		ActiveOnly: ptrBool(true),
		Detailed:   ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["active_only"])
	assert.Equal(t, "true", m["detailed"])
}

func TestUnitIndicesRecoveryParamsToMapNil(t *testing.T) {
	var p *IndicesRecoveryParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesShardStoresParams ----

func TestUnitIndicesShardStoresParamsToMap(t *testing.T) {
	p := &IndicesShardStoresParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
		Status:            []string{"green", "yellow"},
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "green,yellow", m["status"])
}

func TestUnitIndicesShardStoresParamsToMapNil(t *testing.T) {
	var p *IndicesShardStoresParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesReloadAnalyzersParams ----

func TestUnitIndicesReloadAnalyzersParamsToMap(t *testing.T) {
	p := &IndicesReloadAnalyzersParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
}

func TestUnitIndicesReloadAnalyzersParamsToMapNil(t *testing.T) {
	var p *IndicesReloadAnalyzersParams
	assert.Nil(t, p.ToMap())
}

// ---- AddBlockParams ----

func TestUnitAddBlockParamsToMap(t *testing.T) {
	p := &AddBlockParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
		Timeout:           "30s",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitAddBlockParamsToMapNil(t *testing.T) {
	var p *AddBlockParams
	assert.Nil(t, p.ToMap())
}

func TestUnitAddBlockRequestValidate(t *testing.T) {
	require.NoError(t, (&AddBlockRequest{Indices: []string{"idx"}, Block: "write"}).Validate())
}

func TestUnitAddBlockRequestValidateMissingIndices(t *testing.T) {
	err := (&AddBlockRequest{Block: "write"}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Indices")
}

func TestUnitAddBlockRequestValidateMissingBlock(t *testing.T) {
	err := (&AddBlockRequest{Indices: []string{"idx"}}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Block")
}

// ---- AliasParams ----

func TestUnitAliasParamsToMap(t *testing.T) {
	p := &AliasParams{Timeout: "30s"}
	m := p.ToMap()
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitAliasParamsToMapNil(t *testing.T) {
	var p *AliasParams
	assert.Nil(t, p.ToMap())
}

func TestUnitPutAliasRequestValidate(t *testing.T) {
	require.NoError(t, (&PutAliasRequest{Index: "idx", Alias: "a"}).Validate())
}

func TestUnitPutAliasRequestValidateMissing(t *testing.T) {
	err := (&PutAliasRequest{Index: "idx"}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Alias")
}

// ---- GetAliasParams ----

func TestUnitGetAliasParamsToMap(t *testing.T) {
	p := &GetAliasParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
		Local:             ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["local"])
}

func TestUnitGetAliasParamsToMapNil(t *testing.T) {
	var p *GetAliasParams
	assert.Nil(t, p.ToMap())
}

func TestUnitExistsAliasRequestValidate(t *testing.T) {
	require.NoError(t, (&ExistsAliasRequest{Names: []string{"a"}}).Validate())
}

func TestUnitExistsAliasRequestValidateMissing(t *testing.T) {
	err := (&ExistsAliasRequest{}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Names")
}

func TestUnitDeleteAliasRequestValidate(t *testing.T) {
	require.NoError(t, (&DeleteAliasRequest{Indices: []string{"i"}, Names: []string{"a"}}).Validate())
}

func TestUnitDeleteAliasRequestValidateMissing(t *testing.T) {
	err := (&DeleteAliasRequest{Indices: []string{"i"}}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Names")
}

// ---- IndicesUpdateAliasesParams ----

func TestUnitIndicesUpdateAliasesParamsToMap(t *testing.T) {
	p := &IndicesUpdateAliasesParams{Timeout: "30s"}
	m := p.ToMap()
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitIndicesUpdateAliasesParamsToMapNil(t *testing.T) {
	var p *IndicesUpdateAliasesParams
	assert.Nil(t, p.ToMap())
}

// ---- GetSettingsParams ----

func TestUnitGetSettingsParamsToMap(t *testing.T) {
	p := &GetSettingsParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		FlatSettings:      ptrBool(true),
		IgnoreUnavailable: ptrBool(true),
		IncludeDefaults:   ptrBool(true),
		Local:             ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["include_defaults"])
	assert.Equal(t, "true", m["local"])
}

func TestUnitGetSettingsParamsToMapNil(t *testing.T) {
	var p *GetSettingsParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesPutSettingsParams ----

func TestUnitIndicesPutSettingsParamsToMap(t *testing.T) {
	p := &IndicesPutSettingsParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		FlatSettings:      ptrBool(true),
		IgnoreUnavailable: ptrBool(true),
		PreserveExisting:  ptrBool(true),
		Timeout:           "30s",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["preserve_existing"])
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitIndicesPutSettingsParamsToMapNil(t *testing.T) {
	var p *IndicesPutSettingsParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesGetMappingParams ----

func TestUnitIndicesGetMappingParamsToMap(t *testing.T) {
	p := &IndicesGetMappingParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
		Local:             ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["local"])
}

func TestUnitIndicesGetMappingParamsToMapNil(t *testing.T) {
	var p *IndicesGetMappingParams
	assert.Nil(t, p.ToMap())
}

// ---- PutMappingParams ----

func TestUnitPutMappingParamsToMap(t *testing.T) {
	p := &PutMappingParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
		Timeout:           "30s",
		WriteIndexOnly:    ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "30s", m["timeout"])
	assert.Equal(t, "true", m["write_index_only"])
}

func TestUnitPutMappingParamsToMapNil(t *testing.T) {
	var p *PutMappingParams
	assert.Nil(t, p.ToMap())
}

func TestUnitPutMappingRequestValidate(t *testing.T) {
	require.NoError(t, (&PutMappingRequest{Indices: []string{"idx"}, Body: map[string]any{}}).Validate())
}

func TestUnitPutMappingRequestValidateMissingIndices(t *testing.T) {
	err := (&PutMappingRequest{Body: map[string]any{}}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Indices")
}

func TestUnitPutMappingRequestValidateMissingBody(t *testing.T) {
	err := (&PutMappingRequest{Indices: []string{"idx"}}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Body")
}

// ---- GetFieldMappingParams ----

func TestUnitGetFieldMappingParamsToMap(t *testing.T) {
	p := &GetFieldMappingParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
		IncludeDefaults:   ptrBool(true),
		Local:             ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
	assert.Equal(t, "true", m["include_defaults"])
	assert.Equal(t, "true", m["local"])
}

func TestUnitGetFieldMappingParamsToMapNil(t *testing.T) {
	var p *GetFieldMappingParams
	assert.Nil(t, p.ToMap())
}

func TestUnitGetFieldMappingRequestValidate(t *testing.T) {
	require.NoError(t, (&GetFieldMappingRequest{Fields: []string{"f"}}).Validate())
}

func TestUnitGetFieldMappingRequestValidateMissing(t *testing.T) {
	err := (&GetFieldMappingRequest{}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Fields")
}

// ---- IndicesGetTemplateParams ----

func TestUnitIndicesGetTemplateParamsToMap(t *testing.T) {
	p := &IndicesGetTemplateParams{
		FlatSettings: ptrBool(true),
		Local:        ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "true", m["local"])
}

func TestUnitIndicesGetTemplateParamsToMapNil(t *testing.T) {
	var p *IndicesGetTemplateParams
	assert.Nil(t, p.ToMap())
}

// ---- PutTemplateParams ----

func TestUnitPutTemplateParamsToMap(t *testing.T) {
	p := &PutTemplateParams{
		Create:       ptrBool(true),
		FlatSettings:  ptrBool(true),
		Order:        ptrInt(5),
		Timeout:      "30s",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["create"])
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "5", m["order"])
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitPutTemplateParamsToMapNil(t *testing.T) {
	var p *PutTemplateParams
	assert.Nil(t, p.ToMap())
}

func TestUnitPutTemplateRequestValidate(t *testing.T) {
	require.NoError(t, (&PutTemplateRequest{Name: "t", Body: map[string]any{}}).Validate())
}

func TestUnitPutTemplateRequestValidateMissingName(t *testing.T) {
	err := (&PutTemplateRequest{Body: map[string]any{}}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
}

func TestUnitPutTemplateRequestValidateMissingBody(t *testing.T) {
	err := (&PutTemplateRequest{Name: "t"}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Body")
}

// ---- IndicesDeleteTemplateParams ----

func TestUnitIndicesDeleteTemplateParamsToMap(t *testing.T) {
	p := &IndicesDeleteTemplateParams{Timeout: "30s"}
	m := p.ToMap()
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitIndicesDeleteTemplateParamsToMapNil(t *testing.T) {
	var p *IndicesDeleteTemplateParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesGetIndexTemplateParams ----

func TestUnitIndicesGetIndexTemplateParamsToMap(t *testing.T) {
	p := &IndicesGetIndexTemplateParams{
		FlatSettings: ptrBool(true),
		Local:        ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "true", m["local"])
}

func TestUnitIndicesGetIndexTemplateParamsToMapNil(t *testing.T) {
	var p *IndicesGetIndexTemplateParams
	assert.Nil(t, p.ToMap())
}

// ---- PutIndexTemplateParams ----

func TestUnitPutIndexTemplateParamsToMap(t *testing.T) {
	p := &PutIndexTemplateParams{
		Cause:  "test",
		Create: ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "test", m["cause"])
	assert.Equal(t, "true", m["create"])
}

func TestUnitPutIndexTemplateParamsToMapNil(t *testing.T) {
	var p *PutIndexTemplateParams
	assert.Nil(t, p.ToMap())
}

func TestUnitPutIndexTemplateRequestValidate(t *testing.T) {
	require.NoError(t, (&PutIndexTemplateRequest{Name: "t", Body: map[string]any{}}).Validate())
}

func TestUnitPutIndexTemplateRequestValidateMissingName(t *testing.T) {
	err := (&PutIndexTemplateRequest{Body: map[string]any{}}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
}

func TestUnitPutIndexTemplateRequestValidateMissingBody(t *testing.T) {
	err := (&PutIndexTemplateRequest{Name: "t"}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Body")
}

// ---- IndicesDeleteIndexTemplateParams ----

func TestUnitIndicesDeleteIndexTemplateParamsToMap(t *testing.T) {
	p := &IndicesDeleteIndexTemplateParams{Timeout: "30s"}
	m := p.ToMap()
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitIndicesDeleteIndexTemplateParamsToMapNil(t *testing.T) {
	var p *IndicesDeleteIndexTemplateParams
	assert.Nil(t, p.ToMap())
}

// ---- SimulateIndexTemplateParams ----

func TestUnitSimulateIndexTemplateParamsToMap(t *testing.T) {
	p := &SimulateIndexTemplateParams{
		Cause:  "test",
		Create: ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "test", m["cause"])
	assert.Equal(t, "true", m["create"])
}

func TestUnitSimulateIndexTemplateParamsToMapNil(t *testing.T) {
	var p *SimulateIndexTemplateParams
	assert.Nil(t, p.ToMap())
}

func TestUnitSimulateIndexTemplateRequestValidate(t *testing.T) {
	require.NoError(t, (&SimulateIndexTemplateRequest{Name: "n"}).Validate())
}

func TestUnitSimulateIndexTemplateRequestValidateMissing(t *testing.T) {
	err := (&SimulateIndexTemplateRequest{}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
}

// ---- SimulateTemplateParams (= SimulateIndexTemplateParams) ----

func TestUnitSimulateTemplateParamsToMap(t *testing.T) {
	p := &SimulateTemplateParams{Cause: "c", Create: ptrBool(true)}
	m := p.ToMap()
	assert.Equal(t, "c", m["cause"])
	assert.Equal(t, "true", m["create"])
}

// ---- IndicesCreateDataStreamParams ----

func TestUnitIndicesCreateDataStreamParamsToMap(t *testing.T) {
	p := &IndicesCreateDataStreamParams{Timeout: "30s"}
	m := p.ToMap()
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitIndicesCreateDataStreamParamsToMapNil(t *testing.T) {
	var p *IndicesCreateDataStreamParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesGetDataStreamParams ----

func TestUnitIndicesGetDataStreamParamsToMap(t *testing.T) {
	p := &IndicesGetDataStreamParams{
		ExpandWildcards: ExpandWildcardOpen,
		IncludeDefaults: ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["include_defaults"])
}

func TestUnitIndicesGetDataStreamParamsToMapNil(t *testing.T) {
	var p *IndicesGetDataStreamParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesDeleteDataStreamParams ----

func TestUnitIndicesDeleteDataStreamParamsToMap(t *testing.T) {
	p := &IndicesDeleteDataStreamParams{ExpandWildcards: ExpandWildcardAll}
	m := p.ToMap()
	assert.Equal(t, "all", m["expand_wildcards"])
}

func TestUnitIndicesDeleteDataStreamParamsToMapNil(t *testing.T) {
	var p *IndicesDeleteDataStreamParams
	assert.Nil(t, p.ToMap())
}

// ---- ModifyDataStreamParams ----

func TestUnitModifyDataStreamParamsToMap(t *testing.T) {
	p := &ModifyDataStreamParams{Timeout: "30s"}
	m := p.ToMap()
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitModifyDataStreamParamsToMapNil(t *testing.T) {
	var p *ModifyDataStreamParams
	assert.Nil(t, p.ToMap())
}

func TestUnitModifyDataStreamRequestValidate(t *testing.T) {
	require.NoError(t, (&ModifyDataStreamRequest{
		Actions: []*ModifyDataStreamAction{
			{Type: "add_backing_index", DataStream: "ds", Index: "idx"},
		},
	}).Validate())
}

func TestUnitModifyDataStreamRequestValidateMissingActions(t *testing.T) {
	err := (&ModifyDataStreamRequest{}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Actions")
}

func TestUnitModifyDataStreamRequestValidateInvalidActionType(t *testing.T) {
	err := (&ModifyDataStreamRequest{
		Actions: []*ModifyDataStreamAction{
			{Type: "bad", DataStream: "ds", Index: "idx"},
		},
	}).Validate()
	require.Error(t, err)
}

// ---- IndicesResolveIndexParams ----

func TestUnitIndicesResolveIndexParamsToMap(t *testing.T) {
	p := &IndicesResolveIndexParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
}

func TestUnitIndicesResolveIndexParamsToMapNil(t *testing.T) {
	var p *IndicesResolveIndexParams
	assert.Nil(t, p.ToMap())
}

// ---- IndicesResolveClusterParams ----

func TestUnitIndicesResolveClusterParamsToMap(t *testing.T) {
	p := &IndicesResolveClusterParams{
		AllowNoIndices:    ptrBool(true),
		ExpandWildcards:   ExpandWildcardOpen,
		IgnoreUnavailable: ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["ignore_unavailable"])
}

func TestUnitIndicesResolveClusterParamsToMapNil(t *testing.T) {
	var p *IndicesResolveClusterParams
	assert.Nil(t, p.ToMap())
}
