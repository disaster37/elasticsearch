package api

// ClusterHealthParams are the query parameters for GET /_cluster/health.
type ClusterHealthParams struct {
	Common                      *CommonParams
	Level                       Level
	Local                       *bool
	Timeout                     string
	WaitForActiveShards         string
	WaitForEvents               string
	WaitForNoInitializingShards *bool
	WaitForNoRelocatingShards   *bool
	WaitForNodes                string
	WaitForStatus               WaitForStatus
}

// ToMap converts ClusterHealthParams to a query-parameter map.
func (p *ClusterHealthParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Level != "" {
		m["level"] = string(p.Level)
	}
	setBoolParam(m, "local", p.Local)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	if p.WaitForEvents != "" {
		m["wait_for_events"] = p.WaitForEvents
	}
	setBoolParam(m, "wait_for_no_initializing_shards", p.WaitForNoInitializingShards)
	setBoolParam(m, "wait_for_no_relocating_shards", p.WaitForNoRelocatingShards)
	if p.WaitForNodes != "" {
		m["wait_for_nodes"] = p.WaitForNodes
	}
	if p.WaitForStatus != "" {
		m["wait_for_status"] = string(p.WaitForStatus)
	}
	return mergeParams(m, p.Common)
}

// ClusterStateRequest is the request for GET /_cluster/state.
type ClusterStateRequest struct {
	Metrics []string
	Indices []string
	Params  *ClusterStateParams
}

// ClusterStateParams are the query parameters for cluster.state.
type ClusterStateParams struct {
	Common                 *CommonParams
	AllowNoIndices         *bool
	ExpandWildcards        ExpandWildcards
	FlatSettings           *bool
	IgnoreUnavailable      *bool
	Local                  *bool
	WaitForMetadataVersion *int64
	WaitForTimeout         string
}

// ToMap converts ClusterStateParams to a query-parameter map.
func (p *ClusterStateParams) ToMap() map[string]string {
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
	setBoolParam(m, "local", p.Local)
	if p.WaitForMetadataVersion != nil {
		m["wait_for_metadata_version"] = int64Str(*p.WaitForMetadataVersion)
	}
	if p.WaitForTimeout != "" {
		m["wait_for_timeout"] = p.WaitForTimeout
	}
	return mergeParams(m, p.Common)
}

// ClusterStatsParams are the query parameters for cluster.stats.
type ClusterStatsParams struct {
	Common       *CommonParams
	FlatSettings *bool
	Timeout      string
}

// ToMap converts ClusterStatsParams to a query-parameter map.
func (p *ClusterStatsParams) ToMap() map[string]string {
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

// ClusterGetSettingsParams are the query parameters for GET /_cluster/settings.
type ClusterGetSettingsParams struct {
	Common          *CommonParams
	FlatSettings    *bool
	IncludeDefaults *bool
	Timeout         string
}

// ToMap converts ClusterGetSettingsParams to a query-parameter map.
func (p *ClusterGetSettingsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "flat_settings", p.FlatSettings)
	setBoolParam(m, "include_defaults", p.IncludeDefaults)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ClusterPutSettingsParams are the query parameters for PUT /_cluster/settings.
type ClusterPutSettingsParams struct {
	Common       *CommonParams
	FlatSettings *bool
	Timeout      string
}

// ToMap converts ClusterPutSettingsParams to a query-parameter map.
func (p *ClusterPutSettingsParams) ToMap() map[string]string {
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

// ClusterAllocationExplainParams are the query parameters for allocation/explain.
type ClusterAllocationExplainParams struct {
	Common              *CommonParams
	IncludeDiskInfo     *bool
	IncludeYesDecisions *bool
}

// ToMap converts ClusterAllocationExplainParams to a query-parameter map.
func (p *ClusterAllocationExplainParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "include_disk_info", p.IncludeDiskInfo)
	setBoolParam(m, "include_yes_decisions", p.IncludeYesDecisions)
	return mergeParams(m, p.Common)
}

// ClusterRerouteParams are the query parameters for cluster.reroute.
type ClusterRerouteParams struct {
	Common      *CommonParams
	DryRun      *bool
	Explain     *bool
	Metric      []string
	RetryFailed *bool
	Timeout     string
}

// ToMap converts ClusterRerouteParams to a query-parameter map.
func (p *ClusterRerouteParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "dry_run", p.DryRun)
	setBoolParam(m, "explain", p.Explain)
	setListParam(m, "metric", p.Metric)
	setBoolParam(m, "retry_failed", p.RetryFailed)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ClusterPendingTasksParams are the query parameters for pending_tasks.
type ClusterPendingTasksParams struct {
	Common *CommonParams
	Local  *bool
}

// ToMap converts ClusterPendingTasksParams to a query-parameter map.
func (p *ClusterPendingTasksParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "local", p.Local)
	return mergeParams(m, p.Common)
}

// ClusterVotingConfigExclusionsParams are the query parameters for voting_config_exclusions.
type ClusterVotingConfigExclusionsParams struct {
	Common    *CommonParams
	NodeIds   []string
	NodeNames []string
	Timeout   string
}

// ToMap converts ClusterVotingConfigExclusionsParams to a query-parameter map.
func (p *ClusterVotingConfigExclusionsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setListParam(m, "node_ids", p.NodeIds)
	setListParam(m, "node_names", p.NodeNames)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ClusterDeleteVotingConfigExclusionsParams are the query parameters for delete voting_config_exclusions.
type ClusterDeleteVotingConfigExclusionsParams struct {
	Common         *CommonParams
	WaitForRemoval *bool
}

// ToMap converts ClusterDeleteVotingConfigExclusionsParams to a query-parameter map.
func (p *ClusterDeleteVotingConfigExclusionsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "wait_for_removal", p.WaitForRemoval)
	return mergeParams(m, p.Common)
}

// ClusterGetComponentTemplateParams are the query parameters for get component templates.
type ClusterGetComponentTemplateParams struct {
	Common       *CommonParams
	FlatSettings *bool
	Local        *bool
}

// ToMap converts ClusterGetComponentTemplateParams to a query-parameter map.
func (p *ClusterGetComponentTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "flat_settings", p.FlatSettings)
	setBoolParam(m, "local", p.Local)
	return mergeParams(m, p.Common)
}

// PutComponentTemplateRequest is the request for PUT /_component_template/{name}.
type PutComponentTemplateRequest struct {
	Name   string `validate:"required"`
	Body   any    `validate:"required"`
	Params *PutComponentTemplateParams
}

// Validate validates the PutComponentTemplateRequest.
func (r *PutComponentTemplateRequest) Validate() error { return validationError(validate.Struct(r)) }

// PutComponentTemplateParams are the query parameters for put component template.
type PutComponentTemplateParams struct {
	Common  *CommonParams
	Cause   string
	Create  *bool
	Timeout string
}

// ToMap converts PutComponentTemplateParams to a query-parameter map.
func (p *PutComponentTemplateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Cause != "" {
		m["cause"] = p.Cause
	}
	setBoolParam(m, "create", p.Create)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ClusterDeleteComponentTemplateParams are the query parameters for delete component template.
type ClusterDeleteComponentTemplateParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts ClusterDeleteComponentTemplateParams to a query-parameter map.
func (p *ClusterDeleteComponentTemplateParams) ToMap() map[string]string {
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

// ClusterHealthResponse is the response from GET /_cluster/health.
type ClusterHealthResponse struct {
	ClusterName                 string                        `json:"cluster_name"`
	Status                      string                        `json:"status"`
	TimedOut                    bool                          `json:"timed_out"`
	NumberOfNodes               int                           `json:"number_of_nodes"`
	NumberOfDataNodes           int                           `json:"number_of_data_nodes"`
	ActivePrimaryShards         int                           `json:"active_primary_shards"`
	ActiveShards                int                           `json:"active_shards"`
	RelocatingShards            int                           `json:"relocating_shards"`
	InitializingShards          int                           `json:"initializing_shards"`
	UnassignedShards            int                           `json:"unassigned_shards"`
	DelayedUnassignedShards     int                           `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int                           `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int                           `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int                           `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64                       `json:"active_shards_percent_as_number"`
	Indices                     map[string]ClusterIndexHealth `json:"indices,omitempty"`
}

// ClusterIndexHealth is the per-index health.
type ClusterIndexHealth struct {
	Status              string `json:"status"`
	NumberOfShards      int    `json:"number_of_shards"`
	NumberOfReplicas    int    `json:"number_of_replicas"`
	ActivePrimaryShards int    `json:"active_primary_shards"`
	ActiveShards        int    `json:"active_shards"`
	RelocatingShards    int    `json:"relocating_shards"`
	InitializingShards  int    `json:"initializing_shards"`
	UnassignedShards    int    `json:"unassigned_shards"`
}

// ClusterStateResponse is the response from GET /_cluster/state.
type ClusterStateResponse struct {
	ClusterName  string         `json:"cluster_name"`
	ClusterUUID  string         `json:"cluster_uuid"`
	Version      int64          `json:"version"`
	StateUUID    string         `json:"state_uuid"`
	MasterNode   string         `json:"master_node"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Nodes        map[string]any `json:"nodes,omitempty"`
	RoutingTable map[string]any `json:"routing_table,omitempty"`
	RoutingNodes map[string]any `json:"routing_nodes,omitempty"`
	Blocks       map[string]any `json:"blocks,omitempty"`
}

// ClusterStatsResponse is the response from GET /_cluster/stats.
type ClusterStatsResponse struct {
	ClusterName string         `json:"cluster_name"`
	ClusterUUID string         `json:"cluster_uuid"`
	Timestamp   int64          `json:"timestamp"`
	Status      string         `json:"status"`
	Indices     map[string]any `json:"indices"`
	Nodes       map[string]any `json:"nodes"`
}

// ClusterGetSettingsResponse is the response from GET /_cluster/settings.
type ClusterGetSettingsResponse struct {
	Persistent map[string]any `json:"persistent"`
	Transient  map[string]any `json:"transient"`
	Defaults   map[string]any `json:"defaults,omitempty"`
}

// ClusterPutSettingsResponse is the response from PUT /_cluster/settings.
type ClusterPutSettingsResponse struct {
	Acknowledged bool           `json:"acknowledged"`
	Persistent   map[string]any `json:"persistent,omitempty"`
	Transient    map[string]any `json:"transient,omitempty"`
}

// ClusterAllocationExplainResponse is the response from allocation/explain.
type ClusterAllocationExplainResponse struct {
	Index                   string         `json:"index,omitempty"`
	Shard                   int            `json:"shard,omitempty"`
	Primary                 bool           `json:"primary"`
	CurrentState            string         `json:"current_state"`
	UnassignedInfo          map[string]any `json:"unassigned_info,omitempty"`
	CanAllocate             string         `json:"can_allocate,omitempty"`
	AllocateExplanation     string         `json:"allocate_explanation,omitempty"`
	ConfiguredDelay         string         `json:"configured_delay,omitempty"`
	ConfiguredDelayInMillis int64          `json:"configured_delay_in_millis,omitempty"`
	CurrentNode             map[string]any `json:"current_node,omitempty"`
	RemainingDelay          string         `json:"remaining_delay,omitempty"`
	AllocationDelay         string         `json:"allocation_delay,omitempty"`
	AllocationDelayInMillis int64          `json:"allocation_delay_in_millis,omitempty"`
}

// ClusterRerouteResponse is the response from cluster.reroute.
type ClusterRerouteResponse struct {
	Acknowledged bool             `json:"acknowledged"`
	State        map[string]any   `json:"state,omitempty"`
	Explanations []map[string]any `json:"explanations,omitempty"`
}

// ClusterPendingTasksResponse is the response from pending_tasks.
type ClusterPendingTasksResponse struct {
	Tasks []struct {
		InsertOrder       int    `json:"insert_order"`
		Priority          string `json:"priority"`
		Source            string `json:"source"`
		TimeInQueueMillis int    `json:"time_in_queue_millis"`
		Executing         bool   `json:"executing"`
	} `json:"tasks"`
}

// ClusterRemoteInfoResponse is the response from GET /_remote/info.
type ClusterRemoteInfoResponse map[string]struct {
	Connected          bool     `json:"connected"`
	Mode               string   `json:"mode"`
	Seeds              []string `json:"seeds"`
	SkipUnavailable    bool     `json:"skip_unavailable"`
	ClusterCredentials string   `json:"cluster_credentials,omitempty"`
}

// ClusterInfoResponse is the response from GET /_cluster/info/{feature}.
type ClusterInfoResponse map[string]any

// ClusterGetComponentTemplateResponse is the response from get component templates.
type ClusterGetComponentTemplateResponse struct {
	ComponentTemplates []struct {
		Name              string       `json:"name"`
		ComponentTemplate TemplateBody `json:"component_template"`
	} `json:"component_templates"`
}
