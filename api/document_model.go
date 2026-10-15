package api

import (
	"fmt"
	"net/http"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
)

// hasControlChars reports whether s contains any ASCII control character
// (0x00-0x1F or 0x7F). Such characters are invalid in URL path segments and
// are rejected as a defense-in-depth measure alongside path escaping.
func hasControlChars(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c == 0x7F {
			return true
		}
	}
	return false
}

// validatePathValue returns an error if s contains ASCII control characters.
// It is a belt-and-suspenders check used by the document Validate() methods.
func validatePathValue(name, s string) error {
	if hasControlChars(s) {
		return fmt.Errorf("validation: %s contains forbidden ASCII control characters", name)
	}
	return nil
}

// IndexRequest is the request for PUT /{index}/_doc/{id} | POST /{index}/_doc.
type IndexRequest struct {
	Index  string `validate:"required"`
	Id     string
	Body   any
	Params *IndexParams
}

// Validate validates the IndexRequest.
func (r *IndexRequest) Validate() error {
	if err := validationError(validate.Struct(r)); err != nil {
		return err
	}
	if err := validatePathValue("index", r.Index); err != nil {
		return err
	}
	return validatePathValue("id", r.Id)
}

// IndexParams are the query parameters for the index endpoint.
type IndexParams struct {
	Common               *CommonParams
	IfPrimaryTerm        *int64
	IfSeqNo              *int64
	IncludeSourceOnError *bool
	OpType               OpType
	Pipeline             string
	Refresh              Refresh
	RequireAlias         *bool
	RequireDataStream    *bool
	Routing              string
	Timeout              string
	Version              *int64
	VersionType          VersionType
	WaitForActiveShards  string
}

// ToMap converts IndexParams to a query-parameter map.
func (p *IndexParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.IfPrimaryTerm != nil {
		m["if_primary_term"] = int64Str(*p.IfPrimaryTerm)
	}
	if p.IfSeqNo != nil {
		m["if_seq_no"] = int64Str(*p.IfSeqNo)
	}
	setBoolParam(m, "include_source_on_error", p.IncludeSourceOnError)
	if p.OpType != "" {
		m["op_type"] = string(p.OpType)
	}
	if p.Pipeline != "" {
		m["pipeline"] = p.Pipeline
	}
	if p.Refresh != "" {
		m["refresh"] = string(p.Refresh)
	}
	setBoolParam(m, "require_alias", p.RequireAlias)
	setBoolParam(m, "require_data_stream", p.RequireDataStream)
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.Version != nil {
		m["version"] = int64Str(*p.Version)
	}
	if p.VersionType != "" {
		m["version_type"] = string(p.VersionType)
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	return mergeParams(m, p.Common)
}

// CreateRequest is the request for PUT /{index}/_doc/{id}/_create.
type CreateRequest struct {
	Index  string `validate:"required"`
	Id     string `validate:"required"`
	Body   any    `validate:"required"`
	Params *CreateParams
}

// Validate validates the CreateRequest.
func (r *CreateRequest) Validate() error { return validationError(validate.Struct(r)) }

// CreateParams are the query parameters for the create endpoint (index params
// minus op_type).
type CreateParams = IndexParams

// GetRequest is the request for GET /{index}/_doc/{id}.
type GetRequest struct {
	Index  string `validate:"required"`
	Id     string `validate:"required"`
	Params *GetParams
}

// Validate validates the GetRequest.
func (r *GetRequest) Validate() error {
	if err := validationError(validate.Struct(r)); err != nil {
		return err
	}
	if err := validatePathValue("index", r.Index); err != nil {
		return err
	}
	return validatePathValue("id", r.Id)
}

// GetParams are the query parameters for the get endpoint.
type GetParams struct {
	Common               *CommonParams
	ForceSyntheticSource *bool
	Preference           string
	Realtime             *bool
	Refresh              *bool
	Routing              string
	Source               string
	SourceExcludes       []string
	SourceIncludes       []string
	StoredFields         []string
	Version              *int64
	VersionType          VersionType
}

// ToMap converts GetParams to a query-parameter map.
func (p *GetParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "force_synthetic_source", p.ForceSyntheticSource)
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	setBoolParam(m, "realtime", p.Realtime)
	setBoolParam(m, "refresh", p.Refresh)
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Source != "" {
		m["_source"] = p.Source
	}
	setListParam(m, "_source_excludes", p.SourceExcludes)
	setListParam(m, "_source_includes", p.SourceIncludes)
	setListParam(m, "stored_fields", p.StoredFields)
	if p.Version != nil {
		m["version"] = int64Str(*p.Version)
	}
	if p.VersionType != "" {
		m["version_type"] = string(p.VersionType)
	}
	return mergeParams(m, p.Common)
}

// GetSourceRequest is the request for GET /{index}/_source/{id}.
type GetSourceRequest struct {
	Index  string `validate:"required"`
	Id     string `validate:"required"`
	Params *GetSourceParams
}

// Validate validates the GetSourceRequest.
func (r *GetSourceRequest) Validate() error { return validationError(validate.Struct(r)) }

// GetSourceParams are the query parameters for the get_source endpoint.
type GetSourceParams = GetParams

// DeleteRequest is the request for DELETE /{index}/_doc/{id}.
type DeleteRequest struct {
	Index  string `validate:"required"`
	Id     string `validate:"required"`
	Params *DeleteParams
}

// Validate validates the DeleteRequest.
func (r *DeleteRequest) Validate() error {
	if err := validationError(validate.Struct(r)); err != nil {
		return err
	}
	if err := validatePathValue("index", r.Index); err != nil {
		return err
	}
	return validatePathValue("id", r.Id)
}

// DeleteParams are the query parameters for the delete endpoint.
type DeleteParams struct {
	Common              *CommonParams
	IfPrimaryTerm       *int64
	IfSeqNo             *int64
	Refresh             Refresh
	Routing             string
	Timeout             string
	Version             *int64
	VersionType         VersionType
	WaitForActiveShards string
}

// ToMap converts DeleteParams to a query-parameter map.
func (p *DeleteParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.IfPrimaryTerm != nil {
		m["if_primary_term"] = int64Str(*p.IfPrimaryTerm)
	}
	if p.IfSeqNo != nil {
		m["if_seq_no"] = int64Str(*p.IfSeqNo)
	}
	if p.Refresh != "" {
		m["refresh"] = string(p.Refresh)
	}
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.Version != nil {
		m["version"] = int64Str(*p.Version)
	}
	if p.VersionType != "" {
		m["version_type"] = string(p.VersionType)
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	return mergeParams(m, p.Common)
}

// UpdateRequest is the request for POST /{index}/_update/{id}.
type UpdateRequest struct {
	Index  string `validate:"required"`
	Id     string `validate:"required"`
	Body   any
	Params *UpdateParams
}

// Validate validates the UpdateRequest.
func (r *UpdateRequest) Validate() error {
	if err := validationError(validate.Struct(r)); err != nil {
		return err
	}
	if err := validatePathValue("index", r.Index); err != nil {
		return err
	}
	return validatePathValue("id", r.Id)
}

// UpdateParams are the query parameters for the update endpoint.
type UpdateParams struct {
	Common               *CommonParams
	IfPrimaryTerm        *int64
	IfSeqNo              *int64
	IncludeSourceOnError *bool
	Lang                 string
	Refresh              Refresh
	RequireAlias         *bool
	RetryOnConflict      *int
	Routing              string
	Source               string
	SourceExcludes       []string
	SourceIncludes       []string
	Timeout              string
	WaitForActiveShards  string
}

// ToMap converts UpdateParams to a query-parameter map.
func (p *UpdateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.IfPrimaryTerm != nil {
		m["if_primary_term"] = int64Str(*p.IfPrimaryTerm)
	}
	if p.IfSeqNo != nil {
		m["if_seq_no"] = int64Str(*p.IfSeqNo)
	}
	setBoolParam(m, "include_source_on_error", p.IncludeSourceOnError)
	if p.Lang != "" {
		m["lang"] = p.Lang
	}
	if p.Refresh != "" {
		m["refresh"] = string(p.Refresh)
	}
	setBoolParam(m, "require_alias", p.RequireAlias)
	if p.RetryOnConflict != nil {
		m["retry_on_conflict"] = intStr(*p.RetryOnConflict)
	}
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Source != "" {
		m["_source"] = p.Source
	}
	setListParam(m, "_source_excludes", p.SourceExcludes)
	setListParam(m, "_source_includes", p.SourceIncludes)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	return mergeParams(m, p.Common)
}

// MgetRequest is the request for POST /_mget | /{index}/_mget.
type MgetRequest struct {
	Index  string
	Body   any `validate:"required"`
	Params *MgetParams
}

// Validate validates the MgetRequest.
func (r *MgetRequest) Validate() error { return validationError(validate.Struct(r)) }

// MgetParams are the query parameters for the mget endpoint.
type MgetParams struct {
	Common               *CommonParams
	ForceSyntheticSource *bool
	Preference           string
	Realtime             *bool
	Refresh              *bool
	Routing              string
	Source               string
	SourceExcludes       []string
	SourceIncludes       []string
	StoredFields         []string
}

// ToMap converts MgetParams to a query-parameter map.
func (p *MgetParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "force_synthetic_source", p.ForceSyntheticSource)
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	setBoolParam(m, "realtime", p.Realtime)
	setBoolParam(m, "refresh", p.Refresh)
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Source != "" {
		m["_source"] = p.Source
	}
	setListParam(m, "_source_excludes", p.SourceExcludes)
	setListParam(m, "_source_includes", p.SourceIncludes)
	setListParam(m, "stored_fields", p.StoredFields)
	return mergeParams(m, p.Common)
}

// BulkRequest is the request for POST /_bulk | /{index}/_bulk (NDJSON body).
type BulkRequest struct {
	Index  string
	Body   string `validate:"required"`
	Params *BulkParams
}

// Validate validates the BulkRequest.
func (r *BulkRequest) Validate() error { return validationError(validate.Struct(r)) }

// BulkParams are the query parameters for the bulk endpoint.
type BulkParams struct {
	Common                *CommonParams
	IncludeSourceOnError  *bool
	ListExecutedPipelines []string
	Pipeline              string
	Refresh               Refresh
	RequireAlias          *bool
	RequireDataStream     *bool
	Routing               string
	Source                string
	SourceExcludes        []string
	SourceIncludes        []string
	Timeout               string
	WaitForActiveShards   string
}

// ToMap converts BulkParams to a query-parameter map.
func (p *BulkParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "include_source_on_error", p.IncludeSourceOnError)
	setListParam(m, "list_executed_pipelines", p.ListExecutedPipelines)
	if p.Pipeline != "" {
		m["pipeline"] = p.Pipeline
	}
	if p.Refresh != "" {
		m["refresh"] = string(p.Refresh)
	}
	setBoolParam(m, "require_alias", p.RequireAlias)
	setBoolParam(m, "require_data_stream", p.RequireDataStream)
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Source != "" {
		m["_source"] = p.Source
	}
	setListParam(m, "_source_excludes", p.SourceExcludes)
	setListParam(m, "_source_includes", p.SourceIncludes)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	return mergeParams(m, p.Common)
}

// DeleteByQueryRequest is the request for POST /{index}/_delete_by_query.
type DeleteByQueryRequest struct {
	Indices []string `validate:"required,min=1"`
	Body    any
	Params  *DeleteByQueryParams
}

// Validate validates the DeleteByQueryRequest.
func (r *DeleteByQueryRequest) Validate() error { return validationError(validate.Struct(r)) }

// DeleteByQueryParams are the query parameters for delete_by_query.
type DeleteByQueryParams struct {
	Common              *CommonParams
	AllowNoIndices      *bool
	Analyzer            string
	AnalyzeWildcard     *bool
	Conflicts           Conflicts
	DefaultOperator     DefaultOperator
	Df                  string
	ExpandWildcards     ExpandWildcards
	From                *int
	IgnoreUnavailable   *bool
	Lenient             *bool
	MaxDocs             *int
	Preference          string
	Q                   string
	Refresh             *bool
	RequestCache        *bool
	RequestsPerSecond   *float64
	Routing             string
	Scroll              string
	ScrollSize          *int
	SearchTimeout       string
	SearchType          SearchType
	SliceId             string
	SliceMax            *int
	Slices              string
	Sort                []string
	Stats               []string
	TerminateAfter      *int
	Timeout             string
	Version             *bool
	WaitForActiveShards string
	WaitForCompletion   *bool
}

// ToMap converts DeleteByQueryParams to a query-parameter map.
func (p *DeleteByQueryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.Analyzer != "" {
		m["analyzer"] = p.Analyzer
	}
	setBoolParam(m, "analyze_wildcard", p.AnalyzeWildcard)
	if p.Conflicts != "" {
		m["conflicts"] = string(p.Conflicts)
	}
	if p.DefaultOperator != "" {
		m["default_operator"] = string(p.DefaultOperator)
	}
	if p.Df != "" {
		m["df"] = p.Df
	}
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "lenient", p.Lenient)
	if p.MaxDocs != nil {
		m["max_docs"] = intStr(*p.MaxDocs)
	}
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	if p.Q != "" {
		m["q"] = p.Q
	}
	setBoolParam(m, "refresh", p.Refresh)
	setBoolParam(m, "request_cache", p.RequestCache)
	if p.RequestsPerSecond != nil {
		m["requests_per_second"] = floatStr(*p.RequestsPerSecond)
	}
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Scroll != "" {
		m["scroll"] = p.Scroll
	}
	if p.ScrollSize != nil {
		m["scroll_size"] = intStr(*p.ScrollSize)
	}
	if p.SearchTimeout != "" {
		m["search_timeout"] = p.SearchTimeout
	}
	if p.SearchType != "" {
		m["search_type"] = string(p.SearchType)
	}
	if p.SliceId != "" {
		m["slice_id"] = p.SliceId
	}
	if p.SliceMax != nil {
		m["slice_max"] = intStr(*p.SliceMax)
	}
	if p.Slices != "" {
		m["slices"] = p.Slices
	}
	setListParam(m, "sort", p.Sort)
	setListParam(m, "stats", p.Stats)
	if p.TerminateAfter != nil {
		m["terminate_after"] = intStr(*p.TerminateAfter)
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	setBoolParam(m, "version", p.Version)
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// UpdateByQueryRequest is the request for POST /{index}/_update_by_query.
type UpdateByQueryRequest struct {
	Indices []string `validate:"required,min=1"`
	Body    any
	Params  *UpdateByQueryParams
}

// Validate validates the UpdateByQueryRequest.
func (r *UpdateByQueryRequest) Validate() error { return validationError(validate.Struct(r)) }

// UpdateByQueryParams are the query parameters for update_by_query. They
// mirror DeleteByQueryParams plus the update-specific pipeline and
// version_type parameters.
type UpdateByQueryParams struct {
	Common              *CommonParams
	AllowNoIndices      *bool
	Analyzer            string
	AnalyzeWildcard     *bool
	Conflicts           Conflicts
	DefaultOperator     DefaultOperator
	Df                  string
	ExpandWildcards     ExpandWildcards
	From                *int
	IgnoreUnavailable   *bool
	Lenient             *bool
	MaxDocs             *int
	Pipeline            string
	Preference          string
	Q                   string
	Refresh             *bool
	RequestCache        *bool
	RequestsPerSecond   *float64
	Routing             string
	Scroll              string
	ScrollSize          *int
	SearchTimeout       string
	SearchType          SearchType
	SliceId             string
	SliceMax            *int
	Slices              string
	Sort                []string
	Stats               []string
	TerminateAfter      *int
	Timeout             string
	Version             *bool
	VersionType         VersionType
	WaitForActiveShards string
	WaitForCompletion   *bool
}

// ToMap converts UpdateByQueryParams to a query-parameter map.
func (p *UpdateByQueryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.Analyzer != "" {
		m["analyzer"] = p.Analyzer
	}
	setBoolParam(m, "analyze_wildcard", p.AnalyzeWildcard)
	if p.Conflicts != "" {
		m["conflicts"] = string(p.Conflicts)
	}
	if p.DefaultOperator != "" {
		m["default_operator"] = string(p.DefaultOperator)
	}
	if p.Df != "" {
		m["df"] = p.Df
	}
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "lenient", p.Lenient)
	if p.MaxDocs != nil {
		m["max_docs"] = intStr(*p.MaxDocs)
	}
	if p.Pipeline != "" {
		m["pipeline"] = p.Pipeline
	}
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	if p.Q != "" {
		m["q"] = p.Q
	}
	setBoolParam(m, "refresh", p.Refresh)
	setBoolParam(m, "request_cache", p.RequestCache)
	if p.RequestsPerSecond != nil {
		m["requests_per_second"] = floatStr(*p.RequestsPerSecond)
	}
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Scroll != "" {
		m["scroll"] = p.Scroll
	}
	if p.ScrollSize != nil {
		m["scroll_size"] = intStr(*p.ScrollSize)
	}
	if p.SearchTimeout != "" {
		m["search_timeout"] = p.SearchTimeout
	}
	if p.SearchType != "" {
		m["search_type"] = string(p.SearchType)
	}
	if p.SliceId != "" {
		m["slice_id"] = p.SliceId
	}
	if p.SliceMax != nil {
		m["slice_max"] = intStr(*p.SliceMax)
	}
	if p.Slices != "" {
		m["slices"] = p.Slices
	}
	setListParam(m, "sort", p.Sort)
	setListParam(m, "stats", p.Stats)
	if p.TerminateAfter != nil {
		m["terminate_after"] = intStr(*p.TerminateAfter)
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	setBoolParam(m, "version", p.Version)
	if p.VersionType != "" {
		m["version_type"] = string(p.VersionType)
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// RethrottleRequest is the request for the *_rethrottle endpoints.
type RethrottleRequest struct {
	TaskId            string `validate:"required"`
	RequestsPerSecond float64
	Params            *RethrottleParams
}

// Validate validates the RethrottleRequest.
func (r *RethrottleRequest) Validate() error { return validationError(validate.Struct(r)) }

// RethrottleParams are the query parameters for rethrottle endpoints.
type RethrottleParams struct {
	Common            *CommonParams
	RequestsPerSecond *float64
}

// ToMap converts RethrottleParams to a query-parameter map.
func (p *RethrottleParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.RequestsPerSecond != nil {
		m["requests_per_second"] = floatStr(*p.RequestsPerSecond)
	}
	return mergeParams(m, p.Common)
}

// ReindexRequest is the request for POST /_reindex.
type ReindexRequest struct {
	Body   any
	Params *ReindexParams
}

// ReindexParams are the query parameters for reindex.
type ReindexParams struct {
	Common              *CommonParams
	MaxDocs             *int
	Refresh             *bool
	RequestsPerSecond   *float64
	RequireAlias        *bool
	Scroll              string
	Slices              string
	Timeout             string
	WaitForActiveShards string
	WaitForCompletion   *bool
}

// ToMap converts ReindexParams to a query-parameter map.
func (p *ReindexParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MaxDocs != nil {
		m["max_docs"] = intStr(*p.MaxDocs)
	}
	setBoolParam(m, "refresh", p.Refresh)
	if p.RequestsPerSecond != nil {
		m["requests_per_second"] = floatStr(*p.RequestsPerSecond)
	}
	setBoolParam(m, "require_alias", p.RequireAlias)
	if p.Scroll != "" {
		m["scroll"] = p.Scroll
	}
	if p.Slices != "" {
		m["slices"] = p.Slices
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// ExplainRequest is the request for GET|POST /{index}/_explain/{id}.
type ExplainRequest struct {
	Index  string `validate:"required"`
	Id     string `validate:"required"`
	Body   any
	Params *ExplainParams
}

// Validate validates the ExplainRequest.
func (r *ExplainRequest) Validate() error { return validationError(validate.Struct(r)) }

// ExplainParams are the query parameters for the explain endpoint.
type ExplainParams struct {
	Common          *CommonParams
	Analyzer        string
	AnalyzeWildcard *bool
	DefaultOperator DefaultOperator
	Df              string
	Lenient         *bool
	Preference      string
	Q               string
	Routing         string
	Source          string
	SourceExcludes  []string
	SourceIncludes  []string
	StoredFields    []string
}

// ToMap converts ExplainParams to a query-parameter map.
func (p *ExplainParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Analyzer != "" {
		m["analyzer"] = p.Analyzer
	}
	setBoolParam(m, "analyze_wildcard", p.AnalyzeWildcard)
	if p.DefaultOperator != "" {
		m["default_operator"] = string(p.DefaultOperator)
	}
	if p.Df != "" {
		m["df"] = p.Df
	}
	setBoolParam(m, "lenient", p.Lenient)
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	if p.Q != "" {
		m["q"] = p.Q
	}
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Source != "" {
		m["_source"] = p.Source
	}
	setListParam(m, "_source_excludes", p.SourceExcludes)
	setListParam(m, "_source_includes", p.SourceIncludes)
	setListParam(m, "stored_fields", p.StoredFields)
	return mergeParams(m, p.Common)
}

// TermVectorsRequest is the request for GET|POST /{index}/_termvectors/{id}.
type TermVectorsRequest struct {
	Index  string `validate:"required"`
	Id     string
	Body   any
	Params *TermVectorsParams
}

// Validate validates the TermVectorsRequest.
func (r *TermVectorsRequest) Validate() error { return validationError(validate.Struct(r)) }

// TermVectorsParams are the query parameters for the termvectors endpoint.
type TermVectorsParams struct {
	Common          *CommonParams
	FieldStatistics *bool
	Fields          []string
	Offsets         *bool
	Payloads        *bool
	Positions       *bool
	Preference      string
	Realtime        *bool
	Routing         string
	TermStatistics  *bool
	Version         *int64
	VersionType     VersionType
}

// ToMap converts TermVectorsParams to a query-parameter map.
func (p *TermVectorsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "field_statistics", p.FieldStatistics)
	setListParam(m, "fields", p.Fields)
	setBoolParam(m, "offsets", p.Offsets)
	setBoolParam(m, "payloads", p.Payloads)
	setBoolParam(m, "positions", p.Positions)
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	setBoolParam(m, "realtime", p.Realtime)
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	setBoolParam(m, "term_statistics", p.TermStatistics)
	if p.Version != nil {
		m["version"] = int64Str(*p.Version)
	}
	if p.VersionType != "" {
		m["version_type"] = string(p.VersionType)
	}
	return mergeParams(m, p.Common)
}

// MtermvectorsRequest is the request for GET|POST /_mtermvectors.
type MtermvectorsRequest struct {
	Index  string
	Body   any
	Params *MtermvectorsParams
}

// MtermvectorsParams are the query parameters for the mtermvectors endpoint.
type MtermvectorsParams = TermVectorsParams

// TermsEnumRequest is the request for POST /{index}/_terms_enum.
type TermsEnumRequest struct {
	Index  string `validate:"required"`
	Body   any
	Params *TermsEnumParams
}

// Validate validates the TermsEnumRequest.
func (r *TermsEnumRequest) Validate() error { return validationError(validate.Struct(r)) }

// TermsEnumParams are the query parameters for the terms_enum endpoint.
type TermsEnumParams struct {
	Common          *CommonParams
	CaseInsensitive *bool
}

// ToMap converts TermsEnumParams to a query-parameter map.
func (p *TermsEnumParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "case_insensitive", p.CaseInsensitive)
	return mergeParams(m, p.Common)
}

// ---- Response structs ----

// IndexResponse represents the result of an index document operation.
type IndexResponse struct {
	Index         string            `json:"_index,omitempty"`
	Id            string            `json:"_id,omitempty"`
	Version       int64             `json:"_version,omitempty"`
	Result        string            `json:"result,omitempty"`
	Shards        *types.ShardsInfo `json:"_shards,omitempty"`
	SeqNo         int64             `json:"_seq_no,omitempty"`
	PrimaryTerm   int64             `json:"_primary_term,omitempty"`
	Status        int               `json:"status,omitempty"`
	ForcedRefresh bool              `json:"forced_refresh,omitempty"`
}

// GetResult represents the result of a get document operation.
type GetResult struct {
	Index       string                           `json:"_index"`
	Id          string                           `json:"_id"`
	Version     *int64                           `json:"_version"`
	SeqNo       *int64                           `json:"_seq_no"`
	PrimaryTerm *int64                           `json:"_primary_term"`
	Source      json.RawMessage                  `json:"_source,omitempty"`
	Found       bool                             `json:"found,omitempty"`
	Fields      map[string]any                   `json:"fields,omitempty"`
	Error       *types.ElasticsearchErrorDetails `json:"error,omitempty"`
}

// DeleteResponse represents the result of a delete document operation.
type DeleteResponse struct {
	Index         string            `json:"_index,omitempty"`
	Id            string            `json:"_id,omitempty"`
	Version       int64             `json:"_version,omitempty"`
	Result        string            `json:"result,omitempty"`
	Shards        *types.ShardsInfo `json:"_shards,omitempty"`
	SeqNo         int64             `json:"_seq_no,omitempty"`
	PrimaryTerm   int64             `json:"_primary_term,omitempty"`
	Status        int               `json:"status,omitempty"`
	ForcedRefresh bool              `json:"forced_refresh,omitempty"`
}

// UpdateResponse represents the result of an update document operation.
type UpdateResponse struct {
	Index         string            `json:"_index,omitempty"`
	Id            string            `json:"_id,omitempty"`
	Version       int64             `json:"_version,omitempty"`
	Result        string            `json:"result,omitempty"`
	Shards        *types.ShardsInfo `json:"_shards,omitempty"`
	SeqNo         int64             `json:"_seq_no,omitempty"`
	PrimaryTerm   int64             `json:"_primary_term,omitempty"`
	Status        int               `json:"status,omitempty"`
	ForcedRefresh bool              `json:"forced_refresh,omitempty"`
	GetResult     *GetResult        `json:"get,omitempty"`
}

// MgetResponse represents the result of a multi-get request.
type MgetResponse struct {
	Docs []*GetResult `json:"docs,omitempty"`
}

// BulkResponse represents the result of a bulk API request.
type BulkResponse struct {
	Took   int64                          `json:"took,omitempty"`
	Errors bool                           `json:"errors,omitempty"`
	Items  []map[string]*BulkResponseItem `json:"items,omitempty"`
}

// BulkResponseItem represents the result of a single operation within a bulk request.
type BulkResponseItem struct {
	Index         string                           `json:"_index,omitempty"`
	Id            string                           `json:"_id,omitempty"`
	Version       int64                            `json:"_version,omitempty"`
	Result        string                           `json:"result,omitempty"`
	Shards        *types.ShardsInfo                `json:"_shards,omitempty"`
	SeqNo         int64                            `json:"_seq_no,omitempty"`
	PrimaryTerm   int64                            `json:"_primary_term,omitempty"`
	Status        int                              `json:"status,omitempty"`
	ForcedRefresh bool                             `json:"forced_refresh,omitempty"`
	Error         *types.ElasticsearchErrorDetails `json:"error,omitempty"`
}

// BulkIndexByScrollResponse represents the result of delete-by-query,
// update-by-query, and reindex operations.
type BulkIndexByScrollResponse struct {
	Header           http.Header `json:"-"`
	Took             int64       `json:"took"`
	SliceId          *int64      `json:"slice_id,omitempty"`
	TimedOut         bool        `json:"timed_out"`
	Total            int64       `json:"total"`
	Updated          int64       `json:"updated,omitempty"`
	Created          int64       `json:"created,omitempty"`
	Deleted          int64       `json:"deleted"`
	Batches          int64       `json:"batches"`
	VersionConflicts int64       `json:"version_conflicts"`
	Noops            int64       `json:"noops"`
	Retries          struct {
		Bulk   int64 `json:"bulk"`
		Search int64 `json:"search"`
	} `json:"retries,omitempty"`
	Throttled            string           `json:"throttled"`
	ThrottledMillis      int64            `json:"throttled_millis"`
	RequestsPerSecond    float64          `json:"requests_per_second"`
	Canceled             string           `json:"canceled,omitempty"`
	ThrottledUntil       string           `json:"throttled_until"`
	ThrottledUntilMillis int64            `json:"throttled_until_millis"`
	Failures             []map[string]any `json:"failures"`
}

// ExplainResponse represents the result of an explain API request.
type ExplainResponse struct {
	Index       string         `json:"_index"`
	Id          string         `json:"_id"`
	Matched     bool           `json:"matched"`
	Explanation map[string]any `json:"explanation"`
}

// TermvectorsResponse represents the result of a term vectors request.
type TermvectorsResponse struct {
	Index       string                          `json:"_index"`
	Id          string                          `json:"_id,omitempty"`
	Version     int                             `json:"_version"`
	Found       bool                            `json:"found"`
	Took        int64                           `json:"took"`
	TermVectors map[string]TermVectorsFieldInfo `json:"term_vectors"`
}

// TermVectorsFieldInfo holds the term vector data for a single field.
type TermVectorsFieldInfo struct {
	FieldStatistics FieldStatistics      `json:"field_statistics"`
	Terms           map[string]TermsInfo `json:"terms"`
}

// FieldStatistics contains aggregate statistics for all terms in a field.
type FieldStatistics struct {
	DocCount   int64 `json:"doc_count"`
	SumDocFreq int64 `json:"sum_doc_freq"`
	SumTtf     int64 `json:"sum_ttf"`
}

// TermsInfo provides per-term statistics within a field's term vector.
type TermsInfo struct {
	DocFreq  int64       `json:"doc_freq"`
	Score    float64     `json:"score"`
	TermFreq int64       `json:"term_freq"`
	Ttf      int64       `json:"ttf"`
	Tokens   []TokenInfo `json:"tokens"`
}

// TokenInfo describes a single token occurrence within a term vector response.
type TokenInfo struct {
	StartOffset int64  `json:"start_offset"`
	EndOffset   int64  `json:"end_offset"`
	Position    int64  `json:"position"`
	Payload     string `json:"payload"`
}

// MultiTermvectorResponse represents the result of a multi-term-vectors request.
type MultiTermvectorResponse struct {
	Docs []*TermvectorsResponse `json:"docs"`
}

// TermsEnumResponse represents the result of a terms_enum request.
type TermsEnumResponse struct {
	Shards   *types.ShardsInfo `json:"_shards,omitempty"`
	Terms    []string          `json:"terms,omitempty"`
	Complete bool              `json:"complete,omitempty"`
}
