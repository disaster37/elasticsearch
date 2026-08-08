package types

// CommonParams are the query parameters common to all Elasticsearch API
// requests (pretty, human, error_trace, filter_path). They are forwarded by
// all service methods and encoded into the request URL.
type CommonParams struct {
	// Pretty formats the JSON response with indentation (adds ?pretty=true).
	Pretty *bool
	// Human returns human-readable values (adds ?human=true).
	Human *bool
	// ErrorTrace includes the stack trace for errors (adds ?error_trace=true).
	ErrorTrace *bool
	// FilterPath is a comma-separated list of response fields to include.
	// Useful for reducing network payload.
	FilterPath []string
}

// ToMap converts CommonParams to a query-parameter map. Only non-zero values
// are emitted. Returns nil when no parameters are set.
func (p *CommonParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Pretty != nil {
		m["pretty"] = boolStr(*p.Pretty)
	}
	if p.Human != nil {
		m["human"] = boolStr(*p.Human)
	}
	if p.ErrorTrace != nil {
		m["error_trace"] = boolStr(*p.ErrorTrace)
	}
	if len(p.FilterPath) > 0 {
		m["filter_path"] = joinComma(p.FilterPath)
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// AcknowledgedResponse is the standard response for operations that request
// cluster acknowledgment. Most index creation, deletion, and mapping
// operations return this type.
type AcknowledgedResponse struct {
	Acknowledged       bool   `json:"acknowledged"`
	ShardsAcknowledged bool   `json:"shards_acknowledged,omitempty"`
	Index              string `json:"index,omitempty"`
}

// ShardsInfo represents the "_shards" section of most Elasticsearch responses.
type ShardsInfo struct {
	Total      int                              `json:"total"`
	Successful int                              `json:"successful"`
	Failed     int                              `json:"failed"`
	Failures   []*ShardOperationFailedException `json:"failures,omitempty"`
	Skipped    int                              `json:"skipped,omitempty"`
}

// ShardOperationFailedException describes a single shard failure within a
// [ShardsInfo] response.
type ShardOperationFailedException struct {
	Shard   int            `json:"shard,omitempty"`
	Index   string         `json:"index,omitempty"`
	Status  string         `json:"status,omitempty"`
	Reason  map[string]any `json:"reason,omitempty"`
	Node    string         `json:"_node,omitempty"`
	Primary bool           `json:"primary,omitempty"`
}

// BroadcastResponse is the common response for operations applied across all
// shards of a set of indices (flush, refresh, forcemerge, etc.).
type BroadcastResponse struct {
	Shards     *ShardsInfo                      `json:"_shards,omitempty"`
	Total      int                              `json:"total"`
	Successful int                              `json:"successful"`
	Failed     int                              `json:"failed"`
	Failures   []*ShardOperationFailedException `json:"failures,omitempty"`
}

// FailedNodeException represents a failure that occurred on a specific node
// during a cluster-wide operation. The embedded [ElasticsearchErrorDetails]
// carries the structured error; NodeId identifies the affected node.
type FailedNodeException struct {
	*ElasticsearchErrorDetails
	NodeId string `json:"node_id"`
}

// DocumentVersion carries the seq_no and primary_term used for optimistic
// concurrency control. Pass to document write operations (index, update,
// delete) to fail the operation if the document has been modified since it
// was last fetched.
type DocumentVersion struct {
	SeqNo       *int64
	PrimaryTerm *int64
}

// ListResponse is a generic list-response wrapper used by several services.
type ListResponse[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}
