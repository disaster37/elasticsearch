package api

import (
	"github.com/disaster37/elasticsearch/v9/querydsl"
)

// SearchRequest is the request for GET|POST /_search | /{index}/_search.
type SearchRequest struct {
	Indices   []string
	Body      any
	Params    *SearchParams
	RequestId string
}

// NewSearchRequest creates a SearchRequest from a *querydsl.SearchRequest.
func NewSearchRequest(r *querydsl.SearchRequest) (*SearchRequest, error) {
	body, err := r.Body()
	if err != nil {
		return nil, err
	}
	return &SearchRequest{
		Indices: r.Indices(),
		Body:    rawBody(body),
	}, nil
}

// SearchParams are the query parameters for the search endpoint.
type SearchParams struct {
	Common                     *CommonParams
	AllowNoIndices             *bool
	AllowPartialSearchResults  *bool
	Analyzer                   string
	AnalyzeWildcard            *bool
	BatchedReduceSize          *int
	CcsMinimizeRoundtrips      *bool
	DefaultOperator            DefaultOperator
	Df                         string
	DocvalueFields             []string
	ExpandWildcards            ExpandWildcards
	Explain                    *bool
	From                       *int
	IgnoreThrottled            *bool
	IgnoreUnavailable          *bool
	Lenient                    *bool
	MaxConcurrentShardRequests *int
	MinCompatibleShardNode     string
	PreFilterShardSize         *int
	Preference                 string
	Q                          string
	RequestCache               *bool
	RestTotalHitsAsInt         *bool
	Routing                    string
	Scroll                     string
	SearchType                 SearchType
	SeqNoPrimaryTerm           *bool
	Size                       *int
	Sort                       []string
	Source                     string
	SourceExcludes             []string
	SourceIncludes             []string
	Stats                      []string
	StoredFields               []string
	SuggestField               string
	SuggestMode                SuggestMode
	SuggestSize                *int
	SuggestText                string
	TerminateAfter             *int
	Timeout                    string
	TrackScores                *bool
	TrackTotalHits             string
	TypedKeys                  *bool
	Version                    *bool
}

// ToMap converts SearchParams to a query-parameter map.
func (p *SearchParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	setBoolParam(m, "allow_partial_search_results", p.AllowPartialSearchResults)
	if p.Analyzer != "" {
		m["analyzer"] = p.Analyzer
	}
	setBoolParam(m, "analyze_wildcard", p.AnalyzeWildcard)
	if p.BatchedReduceSize != nil {
		m["batched_reduce_size"] = intStr(*p.BatchedReduceSize)
	}
	setBoolParam(m, "ccs_minimize_roundtrips", p.CcsMinimizeRoundtrips)
	if p.DefaultOperator != "" {
		m["default_operator"] = string(p.DefaultOperator)
	}
	if p.Df != "" {
		m["df"] = p.Df
	}
	setListParam(m, "docvalue_fields", p.DocvalueFields)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "explain", p.Explain)
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	setBoolParam(m, "ignore_throttled", p.IgnoreThrottled)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "lenient", p.Lenient)
	if p.MaxConcurrentShardRequests != nil {
		m["max_concurrent_shard_requests"] = intStr(*p.MaxConcurrentShardRequests)
	}
	if p.MinCompatibleShardNode != "" {
		m["min_compatible_shard_node"] = p.MinCompatibleShardNode
	}
	if p.PreFilterShardSize != nil {
		m["pre_filter_shard_size"] = intStr(*p.PreFilterShardSize)
	}
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	if p.Q != "" {
		m["q"] = p.Q
	}
	setBoolParam(m, "request_cache", p.RequestCache)
	setBoolParam(m, "rest_total_hits_as_int", p.RestTotalHitsAsInt)
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Scroll != "" {
		m["scroll"] = p.Scroll
	}
	if p.SearchType != "" {
		m["search_type"] = string(p.SearchType)
	}
	setBoolParam(m, "seq_no_primary_term", p.SeqNoPrimaryTerm)
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	setListParam(m, "sort", p.Sort)
	if p.Source != "" {
		m["_source"] = p.Source
	}
	setListParam(m, "_source_excludes", p.SourceExcludes)
	setListParam(m, "_source_includes", p.SourceIncludes)
	setListParam(m, "stats", p.Stats)
	setListParam(m, "stored_fields", p.StoredFields)
	if p.SuggestField != "" {
		m["suggest_field"] = p.SuggestField
	}
	if p.SuggestMode != "" {
		m["suggest_mode"] = string(p.SuggestMode)
	}
	if p.SuggestSize != nil {
		m["suggest_size"] = intStr(*p.SuggestSize)
	}
	if p.SuggestText != "" {
		m["suggest_text"] = p.SuggestText
	}
	if p.TerminateAfter != nil {
		m["terminate_after"] = intStr(*p.TerminateAfter)
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	setBoolParam(m, "track_scores", p.TrackScores)
	if p.TrackTotalHits != "" {
		m["track_total_hits"] = p.TrackTotalHits
	}
	setBoolParam(m, "typed_keys", p.TypedKeys)
	setBoolParam(m, "version", p.Version)
	return mergeParams(m, p.Common)
}

// SearchMVTRequest is the request for GET /{index}/_mvt/{field}/{zoom}/{x}/{y}.
// Zoom, X and Y are pointers because 0 is a valid tile coordinate/zoom level
// and required-ness must be distinguishable from the zero value.
type SearchMVTRequest struct {
	Index  string `validate:"required"`
	Field  string `validate:"required"`
	Zoom   *int   `validate:"required"`
	X      *int   `validate:"required"`
	Y      *int   `validate:"required"`
	Body   any
	Params *SearchMVTParams
}

// Validate validates the SearchMVTRequest.
func (r *SearchMVTRequest) Validate() error { return validationError(validate.Struct(r)) }

// SearchMVTParams are the query parameters for the search_mvt endpoint.
type SearchMVTParams struct {
	Common         *CommonParams
	ExactBounds    *bool
	Extent         *int
	GridAgg        string
	GridPrecision  *int
	GridType       string
	Size           *int
	TrackTotalHits string
	WithLabels     *bool
}

// ToMap converts SearchMVTParams to a query-parameter map.
func (p *SearchMVTParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "exact_bounds", p.ExactBounds)
	if p.Extent != nil {
		m["extent"] = intStr(*p.Extent)
	}
	if p.GridAgg != "" {
		m["grid_agg"] = p.GridAgg
	}
	if p.GridPrecision != nil {
		m["grid_precision"] = intStr(*p.GridPrecision)
	}
	if p.GridType != "" {
		m["grid_type"] = p.GridType
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	if p.TrackTotalHits != "" {
		m["track_total_hits"] = p.TrackTotalHits
	}
	setBoolParam(m, "with_labels", p.WithLabels)
	return mergeParams(m, p.Common)
}

// MultiSearchRequest is the request for GET|POST /_msearch (NDJSON body).
type MultiSearchRequest struct {
	Indices []string
	Body    string `validate:"required"`
	Params  *MultiSearchParams
}

// Validate validates the MultiSearchRequest.
func (r *MultiSearchRequest) Validate() error { return validationError(validate.Struct(r)) }

// MultiSearchParams are the query parameters for the msearch endpoint.
type MultiSearchParams struct {
	Common                     *CommonParams
	CcsMinimizeRoundtrips      *bool
	MaxConcurrentSearches      *int
	MaxConcurrentShardRequests *int
	PreFilterShardSize         *int
	RestTotalHitsAsInt         *bool
	SearchType                 SearchType
	TypedKeys                  *bool
}

// ToMap converts MultiSearchParams to a query-parameter map.
func (p *MultiSearchParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "ccs_minimize_roundtrips", p.CcsMinimizeRoundtrips)
	if p.MaxConcurrentSearches != nil {
		m["max_concurrent_searches"] = intStr(*p.MaxConcurrentSearches)
	}
	if p.MaxConcurrentShardRequests != nil {
		m["max_concurrent_shard_requests"] = intStr(*p.MaxConcurrentShardRequests)
	}
	if p.PreFilterShardSize != nil {
		m["pre_filter_shard_size"] = intStr(*p.PreFilterShardSize)
	}
	setBoolParam(m, "rest_total_hits_as_int", p.RestTotalHitsAsInt)
	if p.SearchType != "" {
		m["search_type"] = string(p.SearchType)
	}
	setBoolParam(m, "typed_keys", p.TypedKeys)
	return mergeParams(m, p.Common)
}

// SearchTemplateRequest is the request for the search_template endpoint.
type SearchTemplateRequest struct {
	Indices []string
	Body    any `validate:"required"`
	Params  *SearchTemplateParams
}

// Validate validates the SearchTemplateRequest.
func (r *SearchTemplateRequest) Validate() error { return validationError(validate.Struct(r)) }

// SearchTemplateParams are the query parameters for the search_template endpoint.
type SearchTemplateParams struct {
	Common                *CommonParams
	AllowNoIndices        *bool
	CcsMinimizeRoundtrips *bool
	Explain               *bool
	IgnoreThrottled       *bool
	IgnoreUnavailable     *bool
	Preference            string
	Profile               *bool
	RestTotalHitsAsInt    *bool
	Routing               string
	Scroll                string
	SearchType            SearchType
	TypedKeys             *bool
}

// ToMap converts SearchTemplateParams to a query-parameter map.
func (p *SearchTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	setBoolParam(m, "ccs_minimize_roundtrips", p.CcsMinimizeRoundtrips)
	setBoolParam(m, "explain", p.Explain)
	setBoolParam(m, "ignore_throttled", p.IgnoreThrottled)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	setBoolParam(m, "profile", p.Profile)
	setBoolParam(m, "rest_total_hits_as_int", p.RestTotalHitsAsInt)
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Scroll != "" {
		m["scroll"] = p.Scroll
	}
	if p.SearchType != "" {
		m["search_type"] = string(p.SearchType)
	}
	setBoolParam(m, "typed_keys", p.TypedKeys)
	return mergeParams(m, p.Common)
}

// MultiSearchTemplateRequest is the request for the msearch_template endpoint.
type MultiSearchTemplateRequest struct {
	Indices []string
	Body    string `validate:"required"`
	Params  *MultiSearchTemplateParams
}

// Validate validates the MultiSearchTemplateRequest.
func (r *MultiSearchTemplateRequest) Validate() error { return validationError(validate.Struct(r)) }

// MultiSearchTemplateParams are the query parameters for msearch_template.
type MultiSearchTemplateParams struct {
	Common                *CommonParams
	CcsMinimizeRoundtrips *bool
	MaxConcurrentSearches *int
	RestTotalHitsAsInt    *bool
	SearchType            SearchType
	TypedKeys             *bool
}

// ToMap converts MultiSearchTemplateParams to a query-parameter map.
func (p *MultiSearchTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "ccs_minimize_roundtrips", p.CcsMinimizeRoundtrips)
	if p.MaxConcurrentSearches != nil {
		m["max_concurrent_searches"] = intStr(*p.MaxConcurrentSearches)
	}
	setBoolParam(m, "rest_total_hits_as_int", p.RestTotalHitsAsInt)
	if p.SearchType != "" {
		m["search_type"] = string(p.SearchType)
	}
	setBoolParam(m, "typed_keys", p.TypedKeys)
	return mergeParams(m, p.Common)
}

// RenderSearchTemplateRequest is the request for POST /_render/template.
type RenderSearchTemplateRequest struct {
	Id   string
	Body any
}

// RankEvalRequest is the request for POST /_rank_eval.
type RankEvalRequest struct {
	Indices []string
	Body    any `validate:"required"`
	Params  *RankEvalParams
}

// Validate validates the RankEvalRequest.
func (r *RankEvalRequest) Validate() error { return validationError(validate.Struct(r)) }

// RankEvalParams are the query parameters for the rank_eval endpoint.
type RankEvalParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	SearchType        SearchType
}

// ToMap converts RankEvalParams to a query-parameter map.
func (p *RankEvalParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.SearchType != "" {
		m["search_type"] = string(p.SearchType)
	}
	return mergeParams(m, p.Common)
}

// CountRequest is the request for GET|POST /_count.
type CountRequest struct {
	Indices []string
	Body    any
	Params  *CountParams
}

// CountParams are the query parameters for the count endpoint.
type CountParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	Analyzer          string
	AnalyzeWildcard   *bool
	DefaultOperator   DefaultOperator
	Df                string
	ExpandWildcards   ExpandWildcards
	IgnoreThrottled   *bool
	IgnoreUnavailable *bool
	Lenient           *bool
	MinScore          *float64
	Preference        string
	Q                 string
	Routing           string
	TerminateAfter    *int
}

// ToMap converts CountParams to a query-parameter map.
func (p *CountParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
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
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_throttled", p.IgnoreThrottled)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "lenient", p.Lenient)
	if p.MinScore != nil {
		m["min_score"] = floatStr(*p.MinScore)
	}
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	if p.Q != "" {
		m["q"] = p.Q
	}
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.TerminateAfter != nil {
		m["terminate_after"] = intStr(*p.TerminateAfter)
	}
	return mergeParams(m, p.Common)
}

// ScrollRequest is the request for GET|POST /_search/scroll.
type ScrollRequest struct {
	ScrollId string
	Scroll   string
	Params   *ScrollParams
}

// ScrollParams are the query parameters for the scroll endpoint.
type ScrollParams struct {
	Common             *CommonParams
	Scroll             string
	ScrollId           string
	RestTotalHitsAsInt *bool
}

// ToMap converts ScrollParams to a query-parameter map.
func (p *ScrollParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Scroll != "" {
		m["scroll"] = p.Scroll
	}
	if p.ScrollId != "" {
		m["scroll_id"] = p.ScrollId
	}
	setBoolParam(m, "rest_total_hits_as_int", p.RestTotalHitsAsInt)
	return mergeParams(m, p.Common)
}

// ValidateQueryRequest is the request for GET|POST /_validate/query.
type ValidateQueryRequest struct {
	Indices []string
	Body    any
	Params  *ValidateQueryParams
}

// ValidateQueryParams are the query parameters for the validate_query endpoint.
type ValidateQueryParams struct {
	Common            *CommonParams
	AllShards         *bool
	AllowNoIndices    *bool
	Analyzer          string
	AnalyzeWildcard   *bool
	DefaultOperator   DefaultOperator
	Df                string
	ExpandWildcards   ExpandWildcards
	Explain           *bool
	IgnoreUnavailable *bool
	Lenient           *bool
	Q                 string
	Rewrite           *bool
}

// ToMap converts ValidateQueryParams to a query-parameter map.
func (p *ValidateQueryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "all_shards", p.AllShards)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
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
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "explain", p.Explain)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "lenient", p.Lenient)
	if p.Q != "" {
		m["q"] = p.Q
	}
	setBoolParam(m, "rewrite", p.Rewrite)
	return mergeParams(m, p.Common)
}

// SearchShardsRequest is the request for GET /{index}/_search_shards.
type SearchShardsRequest struct {
	Indices []string
	Params  *SearchShardsParams
}

// SearchShardsParams are the query parameters for the search_shards endpoint.
type SearchShardsParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	Local             *bool
	Preference        string
	Routing           string
}

// ToMap converts SearchShardsParams to a query-parameter map.
func (p *SearchShardsParams) ToMap() map[string]string {
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
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	return mergeParams(m, p.Common)
}

// FieldCapsRequest is the request for GET|POST /_field_caps.
type FieldCapsRequest struct {
	Indices []string
	Body    any
	Params  *FieldCapsParams
}

// FieldCapsParams are the query parameters for the field_caps endpoint.
type FieldCapsParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	Fields            []string
	IgnoreUnavailable *bool
	IncludeUnmapped   *bool
}

// ToMap converts FieldCapsParams to a query-parameter map.
func (p *FieldCapsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setListParam(m, "fields", p.Fields)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "include_unmapped", p.IncludeUnmapped)
	return mergeParams(m, p.Common)
}

// OpenPITRequest is the request for POST /{index}/_pit.
type OpenPITRequest struct {
	Indices []string `validate:"required,min=1"`
	Params  *OpenPITParams
}

// Validate validates the OpenPITRequest.
func (r *OpenPITRequest) Validate() error { return validationError(validate.Struct(r)) }

// OpenPITParams are the query parameters for the open_point_in_time endpoint.
type OpenPITParams struct {
	Common            *CommonParams
	AllowNoIndices    *bool
	ExpandWildcards   ExpandWildcards
	IgnoreUnavailable *bool
	KeepAlive         string `validate:"required"`
	Preference        string
	Routing           string
}

// ToMap converts OpenPITParams to a query-parameter map.
func (p *OpenPITParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_indices", p.AllowNoIndices)
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.KeepAlive != "" {
		m["keep_alive"] = p.KeepAlive
	}
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	return mergeParams(m, p.Common)
}

// OpenPITResponse is the response from POST /{index}/_pit.
type OpenPITResponse struct {
	Id string `json:"id"`
}

// ClosePITResponse is the response from DELETE /_pit.
type ClosePITResponse struct {
	Succeeded bool `json:"succeeded"`
	NumFreed  int  `json:"num_freed"`
}
