package api

import "github.com/disaster37/elasticsearch/v9/types"

// IndicesCreateParams are the query parameters for PUT /{index}.
type IndicesCreateParams struct {
	Common              *CommonParams
	MasterTimeout       string
	Timeout             string
	WaitForActiveShards string
}

// ToMap converts IndicesCreateParams to a query-parameter map.
func (p *IndicesCreateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	return mergeParams(m, p.Common)
}

// IndicesDeleteParams are the query parameters for DELETE /{index}.
type IndicesDeleteParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	MasterTimeout     string
	Timeout           string
}

// ToMap converts IndicesDeleteParams to a query-parameter map.
func (p *IndicesDeleteParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// IndicesGetParams are the query parameters for GET /{index}.
type IndicesGetParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	Features          []string
	FlatSettings      *bool
	IgnoreUnavailable *bool
	IncludeDefaults   *bool
	Local             *bool
	MasterTimeout     string
}

// ToMap converts IndicesGetParams to a query-parameter map.
func (p *IndicesGetParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setListParam(m, "features", p.Features)
	setBoolParam(m, "flat_settings", p.FlatSettings)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "include_defaults", p.IncludeDefaults)
	setBoolParam(m, "local", p.Local)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// IndicesExistsParams are the query parameters for HEAD /{index}.
type IndicesExistsParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	FlatSettings      *bool
	IgnoreUnavailable *bool
	IncludeDefaults   *bool
	Local             *bool
}

// ToMap converts IndicesExistsParams to a query-parameter map.
func (p *IndicesExistsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "flat_settings", p.FlatSettings)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "include_defaults", p.IncludeDefaults)
	setBoolParam(m, "local", p.Local)
	return mergeParams(m, p.Common)
}

// IndicesOpenParams are the query parameters for POST /{index}/_open.
type IndicesOpenParams struct {
	Common              *CommonParams
	AllowNoIndices      *bool
	ExpandWildcards     ExpandWildcards
	IgnoreUnavailable   *bool
	MasterTimeout       string
	Timeout             string
	WaitForActiveShards string
}

// ToMap converts IndicesOpenParams to a query-parameter map.
func (p *IndicesOpenParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	return mergeParams(m, p.Common)
}

// IndicesCloseParams are the query parameters for POST /{index}/_close.
type IndicesCloseParams = IndicesOpenParams

// RolloverRequest is the request for POST /{alias}/_rollover.
type RolloverRequest struct {
	Alias    string `validate:"required"`
	NewIndex string
	Body     any
	Params   *RolloverParams
}

// Validate validates the RolloverRequest.
func (r *RolloverRequest) Validate() error { return validationError(validate.Struct(r)) }

// RolloverParams are the query parameters for the rollover endpoint.
type RolloverParams struct {
	Common              *CommonParams
	DryRun              *bool
	Lazy                *bool
	MasterTimeout       string
	Timeout             string
	WaitForActiveShards string
}

// ToMap converts RolloverParams to a query-parameter map.
func (p *RolloverParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "dry_run", p.DryRun)
	setBoolParam(m, "lazy", p.Lazy)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	return mergeParams(m, p.Common)
}

// ShrinkRequest is the request for PUT /{index}/_shrink/{target}.
type ShrinkRequest struct {
	Source string `validate:"required"`
	Target string `validate:"required"`
	Body   any
	Params *ShrinkParams
}

// Validate validates the ShrinkRequest.
func (r *ShrinkRequest) Validate() error { return validationError(validate.Struct(r)) }

// ShrinkParams are the query parameters for shrink/split/clone.
type ShrinkParams struct {
	Common              *CommonParams
	MasterTimeout       string
	Timeout             string
	WaitForActiveShards string
}

// ToMap converts ShrinkParams to a query-parameter map.
func (p *ShrinkParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	return mergeParams(m, p.Common)
}

// SplitRequest is the request for PUT /{index}/_split/{target}.
type SplitRequest struct {
	Source string `validate:"required"`
	Target string `validate:"required"`
	Body   any
	Params *SplitParams
}

// Validate validates the SplitRequest.
func (r *SplitRequest) Validate() error { return validationError(validate.Struct(r)) }

// SplitParams are the query parameters for the split endpoint.
type SplitParams = ShrinkParams

// CloneRequest is the request for PUT /{index}/_clone/{target}.
type CloneRequest struct {
	Source string `validate:"required"`
	Target string `validate:"required"`
	Body   any
	Params *CloneParams
}

// Validate validates the CloneRequest.
func (r *CloneRequest) Validate() error { return validationError(validate.Struct(r)) }

// CloneParams are the query parameters for the clone endpoint.
type CloneParams = ShrinkParams

// IndicesRefreshParams are the query parameters for POST /_refresh.
type IndicesRefreshParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
}

// ToMap converts IndicesRefreshParams to a query-parameter map.
func (p *IndicesRefreshParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	return mergeParams(m, p.Common)
}

// IndicesFlushParams are the query parameters for POST /_flush.
type IndicesFlushParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	Force             *bool
	IgnoreUnavailable *bool
	WaitIfOngoing     *bool
}

// ToMap converts IndicesFlushParams to a query-parameter map.
func (p *IndicesFlushParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "force", p.Force)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "wait_if_ongoing", p.WaitIfOngoing)
	return mergeParams(m, p.Common)
}

// IndicesForceMergeParams are the query parameters for POST /_forcemerge.
type IndicesForceMergeParams struct {
	Common             *CommonParams
	AllowNoIndices     *bool
	ExpandWildcards    ExpandWildcards
	Flush              *bool
	IgnoreUnavailable  *bool
	MaxNumSegments     *int
	OnlyExpungeDeletes *bool
	WaitForCompletion  *bool
}

// ToMap converts IndicesForceMergeParams to a query-parameter map.
func (p *IndicesForceMergeParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "flush", p.Flush)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.MaxNumSegments != nil {
		m["max_num_segments"] = intStr(*p.MaxNumSegments)
	}
	setBoolParam(m, "only_expunge_deletes", p.OnlyExpungeDeletes)
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// IndicesClearCacheParams are the query parameters for POST /_cache/clear.
type IndicesClearCacheParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	Fielddata         *bool
	Fields            []string
	IgnoreUnavailable *bool
	Query             *bool
	Request           *bool
}

// ToMap converts IndicesClearCacheParams to a query-parameter map.
func (p *IndicesClearCacheParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "fielddata", p.Fielddata)
	setListParam(m, "fields", p.Fields)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "query", p.Query)
	setBoolParam(m, "request", p.Request)
	return mergeParams(m, p.Common)
}

// IndicesStatsRequest is the request for GET /_stats.
type IndicesStatsRequest struct {
	Indices []string
	Metrics []string
	Params  *IndicesStatsParams
}

// IndicesStatsParams are the query parameters for the stats endpoint.
type IndicesStatsParams struct {
	Common                  *CommonParams
	CompletionFields        []string
	ExpandWildcards         ExpandWildcards
	FielddataFields         []string
	Fields                  []string
	ForbidClosedIndices     *bool
	Groups                  []string
	IncludeSegmentFileSizes *bool
	IncludeUnloadedSegments *bool
	Level                   Level
}

// ToMap converts IndicesStatsParams to a query-parameter map.
func (p *IndicesStatsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setListParam(m, "completion_fields", p.CompletionFields)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setListParam(m, "fielddata_fields", p.FielddataFields)
	setListParam(m, "fields", p.Fields)
	setBoolParam(m, "forbid_closed_indices", p.ForbidClosedIndices)
	setListParam(m, "groups", p.Groups)
	setBoolParam(m, "include_segment_file_sizes", p.IncludeSegmentFileSizes)
	setBoolParam(m, "include_unloaded_segments", p.IncludeUnloadedSegments)
	if p.Level != "" {
		m["level"] = string(p.Level)
	}
	return mergeParams(m, p.Common)
}

// IndicesSegmentsParams are the query parameters for GET /_segments.
type IndicesSegmentsParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	Verbose           *bool
}

// ToMap converts IndicesSegmentsParams to a query-parameter map.
func (p *IndicesSegmentsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "verbose", p.Verbose)
	return mergeParams(m, p.Common)
}

// IndicesRecoveryParams are the query parameters for GET /_recovery.
type IndicesRecoveryParams struct {
	Common     *CommonParams
	ActiveOnly *bool
	Detailed   *bool
}

// ToMap converts IndicesRecoveryParams to a query-parameter map.
func (p *IndicesRecoveryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "active_only", p.ActiveOnly)
	setBoolParam(m, "detailed", p.Detailed)
	return mergeParams(m, p.Common)
}

// IndicesShardStoresParams are the query parameters for GET /_shard_stores.
type IndicesShardStoresParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	Status            []string
}

// ToMap converts IndicesShardStoresParams to a query-parameter map.
func (p *IndicesShardStoresParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setListParam(m, "status", p.Status)
	return mergeParams(m, p.Common)
}

// IndicesReloadAnalyzersParams are the query parameters for reload_search_analyzers.
type IndicesReloadAnalyzersParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	Resource          string
}

// ToMap converts IndicesReloadAnalyzersParams to a query-parameter map.
func (p *IndicesReloadAnalyzersParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.Resource != "" {
		m["resource"] = p.Resource
	}
	return mergeParams(m, p.Common)
}

// AddBlockRequest is the request for PUT /{index}/_block/{block}.
type AddBlockRequest struct {
	Indices []string `validate:"required,min=1"`
	Block   string   `validate:"required"`
	Params  *AddBlockParams
}

// Validate validates the AddBlockRequest.
func (r *AddBlockRequest) Validate() error { return validationError(validate.Struct(r)) }

// AddBlockParams are the query parameters for the add_block endpoint.
type AddBlockParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	MasterTimeout     string
	Timeout           string
}

// ToMap converts AddBlockParams to a query-parameter map.
func (p *AddBlockParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// PutAliasRequest is the request for PUT|POST /{index}/_alias/{name}.
type PutAliasRequest struct {
	Index  string `validate:"required"`
	Alias  string `validate:"required"`
	Body   any
	Params *AliasParams
}

// Validate validates the PutAliasRequest.
func (r *PutAliasRequest) Validate() error { return validationError(validate.Struct(r)) }

// AliasParams are the query parameters for put/delete alias.
type AliasParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
}

// ToMap converts AliasParams to a query-parameter map.
func (p *AliasParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// GetAliasRequest is the request for GET /_alias.
type GetAliasRequest struct {
	Indices []string
	Names   []string
	Params  *GetAliasParams
}

// GetAliasParams are the query parameters for get/exists alias.
type GetAliasParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	Local             *bool
	MasterTimeout     string
}

// ToMap converts GetAliasParams to a query-parameter map.
func (p *GetAliasParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "local", p.Local)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// ExistsAliasRequest is the request for HEAD /_alias/{name}.
type ExistsAliasRequest struct {
	Indices []string
	Names   []string `validate:"required,min=1"`
	Params  *GetAliasParams
}

// Validate validates the ExistsAliasRequest.
func (r *ExistsAliasRequest) Validate() error { return validationError(validate.Struct(r)) }

// DeleteAliasRequest is the request for DELETE /{index}/_alias/{name}.
type DeleteAliasRequest struct {
	Indices []string `validate:"required,min=1"`
	Names   []string `validate:"required,min=1"`
	Params  *AliasParams
}

// Validate validates the DeleteAliasRequest.
func (r *DeleteAliasRequest) Validate() error { return validationError(validate.Struct(r)) }

// IndicesUpdateAliasesParams are the query parameters for POST /_aliases.
type IndicesUpdateAliasesParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
}

// ToMap converts IndicesUpdateAliasesParams to a query-parameter map.
func (p *IndicesUpdateAliasesParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// GetSettingsRequest is the request for GET /_settings.
type GetSettingsRequest struct {
	Indices []string
	Names   []string
	Params  *GetSettingsParams
}

// GetSettingsParams are the query parameters for get_settings.
type GetSettingsParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	FlatSettings      *bool
	IgnoreUnavailable *bool
	IncludeDefaults   *bool
	Local             *bool
	MasterTimeout     string
}

// ToMap converts GetSettingsParams to a query-parameter map.
func (p *GetSettingsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "flat_settings", p.FlatSettings)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "include_defaults", p.IncludeDefaults)
	setBoolParam(m, "local", p.Local)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// IndicesPutSettingsParams are the query parameters for PUT /_settings.
type IndicesPutSettingsParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	FlatSettings      *bool
	IgnoreUnavailable *bool
	MasterTimeout     string
	PreserveExisting  *bool
	Reopen            *bool
	Timeout           string
}

// ToMap converts IndicesPutSettingsParams to a query-parameter map.
func (p *IndicesPutSettingsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "flat_settings", p.FlatSettings)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	setBoolParam(m, "preserve_existing", p.PreserveExisting)
	setBoolParam(m, "reopen", p.Reopen)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// IndicesGetMappingParams are the query parameters for GET /_mapping.
type IndicesGetMappingParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	Local             *bool
	MasterTimeout     string
}

// ToMap converts IndicesGetMappingParams to a query-parameter map.
func (p *IndicesGetMappingParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "local", p.Local)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// PutMappingRequest is the request for PUT /{index}/_mapping.
type PutMappingRequest struct {
	Indices []string `validate:"required,min=1"`
	Body    any      `validate:"required"`
	Params  *PutMappingParams
}

// Validate validates the PutMappingRequest.
func (r *PutMappingRequest) Validate() error { return validationError(validate.Struct(r)) }

// PutMappingParams are the query parameters for put_mapping.
type PutMappingParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	MasterTimeout     string
	Timeout           string
	WriteIndexOnly    *bool
}

// ToMap converts PutMappingParams to a query-parameter map.
func (p *PutMappingParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	setBoolParam(m, "write_index_only", p.WriteIndexOnly)
	return mergeParams(m, p.Common)
}

// GetFieldMappingRequest is the request for GET /_mapping/field/{fields}.
type GetFieldMappingRequest struct {
	Indices []string
	Fields  []string `validate:"required,min=1"`
	Params  *GetFieldMappingParams
}

// Validate validates the GetFieldMappingRequest.
func (r *GetFieldMappingRequest) Validate() error { return validationError(validate.Struct(r)) }

// GetFieldMappingParams are the query parameters for get_field_mapping.
type GetFieldMappingParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	IncludeDefaults   *bool
	Local             *bool
}

// ToMap converts GetFieldMappingParams to a query-parameter map.
func (p *GetFieldMappingParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "include_defaults", p.IncludeDefaults)
	setBoolParam(m, "local", p.Local)
	return mergeParams(m, p.Common)
}

// IndicesGetTemplateParams are the query parameters for the legacy get_template.
type IndicesGetTemplateParams struct {
	Common        *CommonParams
	FlatSettings  *bool
	Local         *bool
	MasterTimeout string
}

// ToMap converts IndicesGetTemplateParams to a query-parameter map.
func (p *IndicesGetTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "flat_settings", p.FlatSettings)
	setBoolParam(m, "local", p.Local)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// PutTemplateRequest is the request for the legacy PUT /_template/{name}.
type PutTemplateRequest struct {
	Name   string `validate:"required"`
	Body   any    `validate:"required"`
	Params *PutTemplateParams
}

// Validate validates the PutTemplateRequest.
func (r *PutTemplateRequest) Validate() error { return validationError(validate.Struct(r)) }

// PutTemplateParams are the query parameters for the legacy put_template.
type PutTemplateParams struct {
	Common        *CommonParams
	Cause         string
	Create        *bool
	FlatSettings  *bool
	MasterTimeout string
	Order         *int
	Timeout       string
}

// ToMap converts PutTemplateParams to a query-parameter map.
func (p *PutTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Cause != "" {
		m["cause"] = p.Cause
	}
	setBoolParam(m, "create", p.Create)
	setBoolParam(m, "flat_settings", p.FlatSettings)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Order != nil {
		m["order"] = intStr(*p.Order)
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// IndicesDeleteTemplateParams are the query parameters for delete legacy template.
type IndicesDeleteTemplateParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
}

// ToMap converts IndicesDeleteTemplateParams to a query-parameter map.
func (p *IndicesDeleteTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// IndicesGetIndexTemplateParams are the query parameters for get_index_template.
type IndicesGetIndexTemplateParams struct {
	Common          *CommonParams
	FlatSettings    *bool
	IncludeDefaults *bool
	Local           *bool
	MasterTimeout   string
}

// ToMap converts IndicesGetIndexTemplateParams to a query-parameter map.
func (p *IndicesGetIndexTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "flat_settings", p.FlatSettings)
	setBoolParam(m, "include_defaults", p.IncludeDefaults)
	setBoolParam(m, "local", p.Local)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// PutIndexTemplateRequest is the request for PUT /_index_template/{name}.
type PutIndexTemplateRequest struct {
	Name   string `validate:"required"`
	Body   any    `validate:"required"`
	Params *PutIndexTemplateParams
}

// Validate validates the PutIndexTemplateRequest.
func (r *PutIndexTemplateRequest) Validate() error { return validationError(validate.Struct(r)) }

// PutIndexTemplateParams are the query parameters for put_index_template.
type PutIndexTemplateParams struct {
	Common        *CommonParams
	Cause         string
	Create        *bool
	MasterTimeout string
}

// ToMap converts PutIndexTemplateParams to a query-parameter map.
func (p *PutIndexTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Cause != "" {
		m["cause"] = p.Cause
	}
	setBoolParam(m, "create", p.Create)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// IndicesDeleteIndexTemplateParams are the query parameters for delete_index_template.
type IndicesDeleteIndexTemplateParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
}

// ToMap converts IndicesDeleteIndexTemplateParams to a query-parameter map.
func (p *IndicesDeleteIndexTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// SimulateIndexTemplateRequest is the request for POST /_index_template/_simulate_index.
type SimulateIndexTemplateRequest struct {
	Name   string `validate:"required"`
	Body   any
	Params *SimulateIndexTemplateParams
}

// Validate validates the SimulateIndexTemplateRequest.
func (r *SimulateIndexTemplateRequest) Validate() error { return validationError(validate.Struct(r)) }

// SimulateIndexTemplateParams are the query parameters for simulate_index_template.
type SimulateIndexTemplateParams struct {
	Common          *CommonParams
	Cause           string
	Create          *bool
	IncludeDefaults *bool
	MasterTimeout   string
}

// ToMap converts SimulateIndexTemplateParams to a query-parameter map.
func (p *SimulateIndexTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Cause != "" {
		m["cause"] = p.Cause
	}
	setBoolParam(m, "create", p.Create)
	setBoolParam(m, "include_defaults", p.IncludeDefaults)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// SimulateTemplateRequest is the request for POST /_index_template/_simulate.
type SimulateTemplateRequest struct {
	Name   string
	Body   any
	Params *SimulateTemplateParams
}

// SimulateTemplateParams are the query parameters for simulate_template.
type SimulateTemplateParams = SimulateIndexTemplateParams

// IndicesCreateDataStreamParams are the query parameters for create_data_stream.
type IndicesCreateDataStreamParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
}

// ToMap converts IndicesCreateDataStreamParams to a query-parameter map.
func (p *IndicesCreateDataStreamParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// IndicesGetDataStreamParams are the query parameters for get_data_stream.
type IndicesGetDataStreamParams struct {
	Common          *CommonParams
	ExpandWildcards ExpandWildcards
	IncludeDefaults *bool
	MasterTimeout   string
	Verbose         *bool
}

// ToMap converts IndicesGetDataStreamParams to a query-parameter map.
func (p *IndicesGetDataStreamParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "include_defaults", p.IncludeDefaults)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	setBoolParam(m, "verbose", p.Verbose)
	return mergeParams(m, p.Common)
}

// IndicesDeleteDataStreamParams are the query parameters for delete_data_stream.
type IndicesDeleteDataStreamParams struct {
	Common          *CommonParams
	ExpandWildcards ExpandWildcards
	MasterTimeout   string
}

// ToMap converts IndicesDeleteDataStreamParams to a query-parameter map.
func (p *IndicesDeleteDataStreamParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// ModifyDataStreamRequest is the request for POST /_data_stream/_modify.
type ModifyDataStreamRequest struct {
	Actions []*ModifyDataStreamAction `validate:"required,min=1,dive"`
	Params  *ModifyDataStreamParams
}

// Validate validates the ModifyDataStreamRequest.
func (r *ModifyDataStreamRequest) Validate() error { return validationError(validate.Struct(r)) }

// ModifyDataStreamAction describes a single add/remove backing-index action.
type ModifyDataStreamAction struct {
	Type       string `validate:"required,oneof=add_backing_index remove_backing_index"`
	DataStream string `validate:"required"`
	Index      string `validate:"required"`
}

// ModifyDataStreamParams are the query parameters for modify_data_stream.
type ModifyDataStreamParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts ModifyDataStreamParams to a query-parameter map.
func (p *ModifyDataStreamParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// IndicesResolveIndexParams are the query parameters for resolve_index.
type IndicesResolveIndexParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
}

// ToMap converts IndicesResolveIndexParams to a query-parameter map.
func (p *IndicesResolveIndexParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	return mergeParams(m, p.Common)
}

// IndicesResolveClusterParams are the query parameters for resolve_cluster.
type IndicesResolveClusterParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreThrottled   *bool
	IgnoreUnavailable *bool
	Timeout           string
}

// ToMap converts IndicesResolveClusterParams to a query-parameter map.
func (p *IndicesResolveClusterParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_throttled", p.IgnoreThrottled)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ---- Response structs ----

// IndicesGetResponse is the per-index info returned by GET /{index}.
type IndicesGetResponse struct {
	Aliases    map[string]AliasInfo `json:"aliases,omitempty"`
	Mappings   map[string]any       `json:"mappings,omitempty"`
	Settings   map[string]any       `json:"settings,omitempty"`
	Defaults   map[string]any       `json:"defaults,omitempty"`
	DataStream string               `json:"data_stream,omitempty"`
	System     bool                 `json:"system,omitempty"`
}

// AliasInfo describes a single alias.
type AliasInfo struct {
	Filter        map[string]any `json:"filter,omitempty"`
	IndexRouting  string         `json:"index_routing,omitempty"`
	SearchRouting string         `json:"search_routing,omitempty"`
	IsWriteIndex  *bool          `json:"is_write_index,omitempty"`
	IsHidden      *bool          `json:"is_hidden,omitempty"`
}

// IndicesRolloverResponse is the response from the rollover endpoint.
type IndicesRolloverResponse struct {
	OldIndex           string            `json:"old_index"`
	NewIndex           string            `json:"new_index"`
	RolledOver         bool              `json:"rolled_over"`
	DryRun             bool              `json:"dry_run"`
	Acknowledged       bool              `json:"acknowledged"`
	ShardsAcknowledged bool              `json:"shards_acknowledged"`
	Conditions         map[string]bool   `json:"conditions,omitempty"`
	Shards             *types.ShardsInfo `json:"shards,omitempty"`
}

// RefreshResult is the response from POST /_refresh.
type RefreshResult struct {
	Shards *types.ShardsInfo `json:"_shards,omitempty"`
}

// IndicesAnalyzeResponse is the response from GET|POST /_analyze.
type IndicesAnalyzeResponse struct {
	Tokens []AnalyzeToken `json:"tokens"`
}

// AnalyzeToken describes a single analysis token.
type AnalyzeToken struct {
	Token          string `json:"token"`
	StartOffset    int    `json:"start_offset"`
	EndOffset      int    `json:"end_offset"`
	Position       int    `json:"position"`
	Type           string `json:"type,omitempty"`
	PositionLength int    `json:"position_length,omitempty"`
	Bytes          string `json:"bytes,omitempty"`
	Keyword        bool   `json:"keyword,omitempty"`
}

// IndicesStatsResponse is the response from GET /_stats.
type IndicesStatsResponse struct {
	Shards      *types.ShardsInfo      `json:"_shards,omitempty"`
	All         *IndexStats            `json:"_all,omitempty"`
	Indices     map[string]*IndexStats `json:"indices,omitempty"`
	DataStreams map[string]any         `json:"data_streams,omitempty"`
}

// IndexStats holds per-index or aggregate stats.
type IndexStats struct {
	Primaries *IndexStatsDetail `json:"primaries,omitempty"`
	Total     *IndexStatsDetail `json:"total,omitempty"`
}

// IndexStatsDetail holds the detailed stats sections.
type IndexStatsDetail struct {
	Docs         map[string]any `json:"docs,omitempty"`
	Store        map[string]any `json:"store,omitempty"`
	Indexing     map[string]any `json:"indexing,omitempty"`
	Search       map[string]any `json:"search,omitempty"`
	Get          map[string]any `json:"get,omitempty"`
	Merges       map[string]any `json:"merges,omitempty"`
	Refresh      map[string]any `json:"refresh,omitempty"`
	Flush        map[string]any `json:"flush,omitempty"`
	Warmer       map[string]any `json:"warmer,omitempty"`
	QueryCache   map[string]any `json:"query_cache,omitempty"`
	Fielddata    map[string]any `json:"fielddata,omitempty"`
	Completion   map[string]any `json:"completion,omitempty"`
	Segments     map[string]any `json:"segments,omitempty"`
	Translog     map[string]any `json:"translog,omitempty"`
	RequestCache map[string]any `json:"request_cache,omitempty"`
	Recovery     map[string]any `json:"recovery,omitempty"`
}

// IndicesSegmentsResponse is the response from GET /_segments.
type IndicesSegmentsResponse struct {
	Shards  *types.ShardsInfo `json:"_shards,omitempty"`
	Indices map[string]struct {
		Shards map[string][]IndexSegmentShard `json:"shards"`
	} `json:"indices,omitempty"`
}

// IndexSegmentShard describes a shard's segments.
type IndexSegmentShard struct {
	Routing              IndexSegmentRouting `json:"routing"`
	NumCommittedSegments int                 `json:"num_committed_segments"`
	NumSearchSegments    int                 `json:"num_search_segments"`
	Segments             map[string]Segment  `json:"segments"`
}

// IndexSegmentRouting describes shard routing.
type IndexSegmentRouting struct {
	State   string `json:"state"`
	Primary bool   `json:"primary"`
	Node    string `json:"node"`
}

// Segment describes a single Lucene segment.
type Segment struct {
	Generation  int               `json:"generation"`
	NumDocs     int               `json:"num_docs"`
	DeletedDocs int               `json:"deleted_docs"`
	SizeInBytes int64             `json:"size_in_bytes"`
	Committed   bool              `json:"committed"`
	Search      bool              `json:"search"`
	Version     string            `json:"version"`
	Compound    bool              `json:"compound"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// IndicesRecoveryResponse is the response from GET /_recovery.
type IndicesRecoveryResponse map[string]IndexRecovery

// IndexRecovery holds recovery info for an index.
type IndexRecovery struct {
	Shards []ShardRecovery `json:"shards"`
}

// ShardRecovery describes a single shard recovery.
type ShardRecovery struct {
	Type              string         `json:"type"`
	Stage             string         `json:"stage"`
	Primary           bool           `json:"primary"`
	StartTimeInMillis int64          `json:"start_time_in_millis"`
	StopTimeInMillis  int64          `json:"stop_time_in_millis,omitempty"`
	TotalTimeInMillis int64          `json:"total_time_in_millis"`
	Source            RecoveryNode   `json:"source"`
	Target            RecoveryNode   `json:"target"`
	Index             map[string]any `json:"index"`
}

// RecoveryNode identifies a recovery source/target node.
type RecoveryNode struct {
	Host             string `json:"host"`
	TransportAddress string `json:"transport_address"`
	Name             string `json:"name"`
}

// IndicesShardStoresResponse is the response from GET /_shard_stores.
type IndicesShardStoresResponse struct {
	Indices map[string]struct {
		Shards map[string]IndexShardStores `json:"shards"`
	} `json:"indices"`
}

// IndexShardStores holds shard store info.
type IndexShardStores struct {
	Stores []ShardStore `json:"stores"`
}

// ShardStore describes a single shard store.
type ShardStore struct {
	ID             string `json:"id"`
	Path           string `json:"path"`
	Allocation     string `json:"allocation"`
	StoreException *struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	} `json:"store_exception,omitempty"`
}

// IndicesReloadAnalyzersResponse is the response from reload_search_analyzers.
type IndicesReloadAnalyzersResponse struct {
	Shards        *types.ShardsInfo `json:"_shards,omitempty"`
	ReloadDetails []ReloadDetails   `json:"reload_details,omitempty"`
}

// ReloadDetails describes per-node analyzer reload results.
type ReloadDetails struct {
	Index             string   `json:"index"`
	Node              string   `json:"node"`
	ReloadedAnalyzers []string `json:"reloaded_analyzers"`
}

// IndicesBlockResponse is the response from add_block.
type IndicesBlockResponse struct {
	Shards       *types.ShardsInfo `json:"_shards,omitempty"`
	Acknowledged bool              `json:"acknowledged"`
	Indices      []map[string]any  `json:"indices,omitempty"`
}

// IndicesGetAliasResponse is the per-index alias info.
type IndicesGetAliasResponse struct {
	Aliases map[string]AliasInfo `json:"aliases"`
}

// IndicesGetSettingsResponse is the per-index settings.
type IndicesGetSettingsResponse struct {
	Settings map[string]any `json:"settings,omitempty"`
	Defaults map[string]any `json:"defaults,omitempty"`
}

// IndicesGetMappingResponse is the per-index mapping.
type IndicesGetMappingResponse struct {
	Mappings map[string]any `json:"mappings"`
}

// IndicesGetTemplateResponse is the legacy index template.
type IndicesGetTemplateResponse struct {
	Order         int            `json:"order,omitempty"`
	Version       int            `json:"version,omitempty"`
	IndexPatterns []string       `json:"index_patterns,omitempty"`
	Settings      map[string]any `json:"settings,omitempty"`
	Mappings      map[string]any `json:"mappings,omitempty"`
	Aliases       map[string]any `json:"aliases,omitempty"`
	Deprecated    bool           `json:"deprecated,omitempty"`
}

// IndicesGetIndexTemplateResponse is the response from get_index_template.
type IndicesGetIndexTemplateResponse struct {
	IndexTemplates []IndexTemplateItem `json:"index_templates"`
}

// IndexTemplateItem wraps a composable index template.
type IndexTemplateItem struct {
	Name          string        `json:"name"`
	IndexTemplate IndexTemplate `json:"index_template"`
}

// IndexTemplate is a composable index template.
type IndexTemplate struct {
	IndexPatterns   []string       `json:"index_patterns"`
	Template        TemplateBody   `json:"template"`
	Priority        *int           `json:"priority,omitempty"`
	Version         *int           `json:"version,omitempty"`
	ComposedOf      []string       `json:"composed_of,omitempty"`
	DataStream      map[string]any `json:"data_stream,omitempty"`
	AllowAutoCreate *bool          `json:"allow_auto_create,omitempty"`
}

// TemplateBody is the settings/mappings/aliases template body.
type TemplateBody struct {
	Settings map[string]any `json:"settings,omitempty"`
	Mappings map[string]any `json:"mappings,omitempty"`
	Aliases  map[string]any `json:"aliases,omitempty"`
}

// IndicesSimulateTemplateResponse is the response from simulate template.
type IndicesSimulateTemplateResponse struct {
	Template    TemplateBody `json:"template"`
	Overlapping []struct {
		Name          string   `json:"name"`
		IndexPatterns []string `json:"index_patterns"`
	} `json:"overlapping,omitempty"`
}

// IndicesDataStreamGetResponse is the response from get_data_stream.
type IndicesDataStreamGetResponse struct {
	DataStreams []DataStream `json:"data_streams"`
}

// DataStream describes a data stream.
type DataStream struct {
	Name               string            `json:"name"`
	TimestampField     DataStreamField   `json:"timestamp_field"`
	Indices            []DataStreamIndex `json:"indices"`
	Generation         int               `json:"generation"`
	Status             string            `json:"status,omitempty"`
	Replicated         bool              `json:"replicated,omitempty"`
	Hidden             bool              `json:"hidden,omitempty"`
	System             bool              `json:"system,omitempty"`
	AllowCustomRouting bool              `json:"allow_custom_routing,omitempty"`
	Template           string            `json:"template,omitempty"`
}

// DataStreamField names the timestamp field.
type DataStreamField struct {
	Field string `json:"name"`
}

// DataStreamIndex identifies a backing index.
type DataStreamIndex struct {
	Index          string `json:"index_name"`
	IndexUUID      string `json:"index_uuid"`
	IsHidden       bool   `json:"is_hidden"`
	IsBackingIndex bool   `json:"is_backing_index"`
}

// IndicesDataStreamsStatsResponse is the response from data_stream_stats.
type IndicesDataStreamsStatsResponse struct {
	Shards              *types.ShardsInfo `json:"_shards,omitempty"`
	DataStreamCount     int               `json:"data_stream_count"`
	BackingIndices      int               `json:"backing_indices"`
	TotalStoreSizeBytes int64             `json:"total_store_size_bytes"`
	DataStreams         []struct {
		DataStream       string `json:"data_stream"`
		BackingIndices   int    `json:"backing_indices"`
		StoreSizeBytes   int64  `json:"store_size_bytes"`
		MaximumRetention string `json:"maximum_retention"`
	} `json:"data_streams"`
}

// IndicesResolveIndexResponse is the response from resolve_index.
type IndicesResolveIndexResponse struct {
	Indices     []ResolveIndexInfo      `json:"indices,omitempty"`
	Aliases     []ResolveAliasInfo      `json:"aliases,omitempty"`
	DataStreams []ResolveDataStreamInfo `json:"data_streams,omitempty"`
}

// ResolveIndexInfo describes a resolved index.
type ResolveIndexInfo struct {
	Name       string   `json:"name"`
	Attributes []string `json:"attributes,omitempty"`
	Aliases    []string `json:"aliases,omitempty"`
}

// ResolveAliasInfo describes a resolved alias.
type ResolveAliasInfo struct {
	Name    string   `json:"name"`
	Indices []string `json:"indices"`
}

// ResolveDataStreamInfo describes a resolved data stream.
type ResolveDataStreamInfo struct {
	Name           string   `json:"name"`
	BackingIndices []string `json:"backing_indices"`
	Aliases        []string `json:"aliases,omitempty"`
}

// ResolveClusterInfo describes a remote cluster.
type ResolveClusterInfo struct {
	Connected          bool     `json:"connected"`
	Mode               string   `json:"mode"`
	Seeds              []string `json:"seeds"`
	SkipUnavailable    bool     `json:"skip_unavailable"`
	ClusterCredentials string   `json:"cluster_credentials,omitempty"`
}
