package api

import "github.com/disaster37/elasticsearch/v9/querydsl"

// AsyncSearchSubmitRequest is the request for POST /{index}/_async_search.
type AsyncSearchSubmitRequest struct {
	Indices []string
	Body    any
	Params  *AsyncSearchSubmitParams
}

// AsyncSearchSubmitParams are the query parameters for async_search.submit.
type AsyncSearchSubmitParams struct {
	Common                    *CommonParams
	AllowNoIndices            *bool
	AllowPartialSearchResults *bool
	Analyzer                  string
	AnalyzeWildcard           *bool
	BatchedReduceSize         *int
	DefaultOperator           DefaultOperator
	Df                        string
	ExpandWildcards           ExpandWildcards
	Explain                   *bool
	From                      *int
	IgnoreThrottled           *bool
	IgnoreUnavailable         *bool
	KeepAlive                 string
	KeepOnCompletion          *bool
	Lenient                   *bool
	Preference                string
	Q                         string
	RequestCache              *bool
	Routing                   string
	SearchType                SearchType
	Size                      *int
	Sort                      []string
	Source                    string
	SourceExcludes            []string
	SourceIncludes            []string
	Stats                     []string
	StoredFields              []string
	SuggestField              string
	SuggestMode               SuggestMode
	SuggestSize               *int
	SuggestText               string
	TerminateAfter            *int
	Timeout                   string
	TrackScores               *bool
	TrackTotalHits            string
	TypedKeys                 *bool
	WaitForCompletionTimeout  string
}

// ToMap converts AsyncSearchSubmitParams to a query-parameter map.
func (p *AsyncSearchSubmitParams) ToMap() map[string]string {
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
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	setBoolParam(m, "ignore_throttled", p.IgnoreThrottled)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.KeepAlive != "" {
		m["keep_alive"] = p.KeepAlive
	}
	setBoolParam(m, "keep_on_completion", p.KeepOnCompletion)
	setBoolParam(m, "lenient", p.Lenient)
	if p.Preference != "" {
		m["preference"] = p.Preference
	}
	if p.Q != "" {
		m["q"] = p.Q
	}
	setBoolParam(m, "request_cache", p.RequestCache)
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.SearchType != "" {
		m["search_type"] = string(p.SearchType)
	}
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
	if p.WaitForCompletionTimeout != "" {
		m["wait_for_completion_timeout"] = p.WaitForCompletionTimeout
	}
	return mergeParams(m, p.Common)
}

// AsyncSearchGetParams are the query parameters for async_search.get.
type AsyncSearchGetParams struct {
	Common                   *CommonParams
	KeepAlive                string
	TypedKeys                *bool
	WaitForCompletionTimeout string
}

// ToMap converts AsyncSearchGetParams to a query-parameter map.
func (p *AsyncSearchGetParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.KeepAlive != "" {
		m["keep_alive"] = p.KeepAlive
	}
	setBoolParam(m, "typed_keys", p.TypedKeys)
	if p.WaitForCompletionTimeout != "" {
		m["wait_for_completion_timeout"] = p.WaitForCompletionTimeout
	}
	return mergeParams(m, p.Common)
}

// AsyncSearchResponse is the response from async_search submit/get.
type AsyncSearchResponse struct {
	Id                     string                 `json:"id,omitempty"`
	IsPartial              bool                   `json:"is_partial"`
	IsRunning              bool                   `json:"is_running"`
	StartTimeInMillis      int64                  `json:"start_time_in_millis,omitempty"`
	ExpirationTimeInMillis int64                  `json:"expiration_time_in_millis,omitempty"`
	Response               *querydsl.SearchResult `json:"response,omitempty"`
}

// AsyncSearchStatusResponse is the response from async_search.status.
type AsyncSearchStatusResponse struct {
	Id                     string `json:"id,omitempty"`
	IsRunning              bool   `json:"is_running"`
	IsPartial              bool   `json:"is_partial"`
	StartTimeInMillis      int64  `json:"start_time_in_millis,omitempty"`
	ExpirationTimeInMillis int64  `json:"expiration_time_in_millis,omitempty"`
	Shards                 *struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards,omitempty"`
}
