package api

import "github.com/disaster37/elasticsearch/v9/types"

// NodesInfoRequest is the request for GET /_nodes.
type NodesInfoRequest struct {
	NodeIds []string
	Metrics []string
	Params  *NodesInfoParams
}

// NodesInfoParams are the query parameters for nodes.info.
type NodesInfoParams struct {
	Common       *CommonParams
	FlatSettings *bool
	Timeout      string
}

// ToMap converts NodesInfoParams to a query-parameter map.
func (p *NodesInfoParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "flat_settings", p.FlatSettings)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// NodesStatsRequest is the request for GET /_nodes/stats.
type NodesStatsRequest struct {
	NodeIds      []string
	Metrics      []string
	IndexMetrics []string
	Params       *NodesStatsParams
}

// NodesStatsParams are the query parameters for nodes.stats.
type NodesStatsParams struct {
	Common                  *CommonParams
	CompletionFields        []string
	FielddataFields         []string
	Fields                  []string
	Groups                  []string
	IncludeSegmentFileSizes *bool
	IncludeUnloadedSegments *bool
	Level                   Level
	Timeout                 string
	Types                   []string
}

// ToMap converts NodesStatsParams to a query-parameter map.
func (p *NodesStatsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setListParam(m, "completion_fields", p.CompletionFields)
	setListParam(m, "fielddata_fields", p.FielddataFields)
	setListParam(m, "fields", p.Fields)
	setListParam(m, "groups", p.Groups)
	setBoolParam(m, "include_segment_file_sizes", p.IncludeSegmentFileSizes)
	setBoolParam(m, "include_unloaded_segments", p.IncludeUnloadedSegments)
	if p.Level != "" {
		m["level"] = string(p.Level)
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	setListParam(m, "types", p.Types)
	return mergeParams(m, p.Common)
}

// NodesUsageRequest is the request for GET /_nodes/usage.
type NodesUsageRequest struct {
	NodeIds []string
	Metrics []string
	Params  *NodesUsageParams
}

// NodesUsageParams are the query parameters for nodes.usage.
type NodesUsageParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts NodesUsageParams to a query-parameter map.
func (p *NodesUsageParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// NodesHotThreadsParams are the query parameters for hot_threads.
type NodesHotThreadsParams struct {
	Common            *CommonParams
	IgnoreIdleThreads *bool
	Interval          string
	Snapshots         *int
	Sort              string
	Threads           *int
	Timeout           string
	Type              string
}

// ToMap converts NodesHotThreadsParams to a query-parameter map.
func (p *NodesHotThreadsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "ignore_idle_threads", p.IgnoreIdleThreads)
	if p.Interval != "" {
		m["interval"] = p.Interval
	}
	if p.Snapshots != nil {
		m["snapshots"] = intStr(*p.Snapshots)
	}
	if p.Sort != "" {
		m["sort"] = p.Sort
	}
	if p.Threads != nil {
		m["threads"] = intStr(*p.Threads)
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.Type != "" {
		m["type"] = p.Type
	}
	return mergeParams(m, p.Common)
}

// NodesReloadSecureSettingsParams are the query parameters for reload_secure_settings.
type NodesReloadSecureSettingsParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts NodesReloadSecureSettingsParams to a query-parameter map.
func (p *NodesReloadSecureSettingsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ---- Response structs ----

// NodesInfoResponse is the response from GET /_nodes.
type NodesInfoResponse struct {
	ClusterName string              `json:"cluster_name"`
	Nodes       map[string]NodeInfo `json:"nodes"`
}

// NodeInfo describes a node.
type NodeInfo struct {
	Name             string            `json:"name"`
	TransportAddress string            `json:"transport_address"`
	Host             string            `json:"host"`
	IP               string            `json:"ip"`
	Version          string            `json:"version"`
	BuildHash        string            `json:"build_hash"`
	BuildType        string            `json:"build_type"`
	BuildDate        string            `json:"build_date"`
	Roles            []string          `json:"roles"`
	Attributes       map[string]string `json:"attributes,omitempty"`
	Settings         map[string]any    `json:"settings,omitempty"`
	OS               map[string]any    `json:"os,omitempty"`
	Process          map[string]any    `json:"process,omitempty"`
	JVM              map[string]any    `json:"jvm,omitempty"`
	ThreadPool       map[string]any    `json:"thread_pool,omitempty"`
	Transport        map[string]any    `json:"transport,omitempty"`
	HTTP             map[string]any    `json:"http,omitempty"`
	Plugins          []PluginInfo      `json:"plugins,omitempty"`
	Modules          []PluginInfo      `json:"modules,omitempty"`
	Ingest           map[string]any    `json:"ingest,omitempty"`
	Aggregations     map[string]any    `json:"aggregations,omitempty"`
}

// PluginInfo describes an installed plugin or module.
type PluginInfo struct {
	Name                 string   `json:"name"`
	Version              string   `json:"version"`
	ElasticsearchVersion string   `json:"elasticsearch_version"`
	JavaVersion          string   `json:"java_version,omitempty"`
	Description          string   `json:"description"`
	Classname            string   `json:"classname,omitempty"`
	ExtendedPlugins      []string `json:"extended_plugins,omitempty"`
	HasNativeController  bool     `json:"has_native_controller"`
}

// NodesStatsResponse is the response from GET /_nodes/stats.
type NodesStatsResponse struct {
	ClusterName string               `json:"cluster_name"`
	Nodes       map[string]NodeStats `json:"nodes"`
}

// NodeStats describes per-node stats.
type NodeStats struct {
	Timestamp         int64             `json:"timestamp"`
	Name              string            `json:"name"`
	TransportAddress  string            `json:"transport_address"`
	Host              string            `json:"host"`
	IP                string            `json:"ip"`
	Roles             []string          `json:"roles"`
	Attributes        map[string]string `json:"attributes,omitempty"`
	Indices           map[string]any    `json:"indices,omitempty"`
	OS                map[string]any    `json:"os,omitempty"`
	Process           map[string]any    `json:"process,omitempty"`
	JVM               map[string]any    `json:"jvm,omitempty"`
	ThreadPool        map[string]any    `json:"thread_pool,omitempty"`
	FS                map[string]any    `json:"fs,omitempty"`
	Transport         map[string]any    `json:"transport,omitempty"`
	HTTP              map[string]any    `json:"http,omitempty"`
	Breakers          map[string]any    `json:"breakers,omitempty"`
	Script            map[string]any    `json:"script,omitempty"`
	Discovery         map[string]any    `json:"discovery,omitempty"`
	Ingest            map[string]any    `json:"ingest,omitempty"`
	AdaptiveSelection map[string]any    `json:"adaptive_selection,omitempty"`
	IndexingPressure  map[string]any    `json:"indexing_pressure,omitempty"`
}

// NodesUsageResponse is the response from GET /_nodes/usage.
type NodesUsageResponse struct {
	ClusterName string `json:"cluster_name"`
	Nodes       map[string]struct {
		Timestamp   int64          `json:"timestamp"`
		Since       int64          `json:"since"`
		RestActions map[string]int `json:"rest_actions"`
	} `json:"nodes"`
}

// NodesReloadSecureSettingsResponse is the response from reload_secure_settings.
type NodesReloadSecureSettingsResponse struct {
	ClusterName string                       `json:"cluster_name"`
	Nodes       map[string]any               `json:"nodes"`
	Failures    []*types.FailedNodeException `json:"failures,omitempty"`
}
