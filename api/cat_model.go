package api

// catCommonParams embeds the cat-common query parameters shared by every cat
// endpoint (h, help, s, v, format, local) plus CommonParams.
type catCommonParams struct {
	Common *CommonParams
	H      []string
	Help   *bool
	S      []string
	V      *bool
	Local  *bool
}

func (p *catCommonParams) toMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	// Always request JSON unless the caller overrides via H/S/etc.
	m["format"] = "json"
	setListParam(m, "h", p.H)
	setBoolParam(m, "help", p.Help)
	setListParam(m, "s", p.S)
	setBoolParam(m, "v", p.V)
	setBoolParam(m, "local", p.Local)
	return mergeParams(m, p.Common)
}

// CatAliasesParams are the query parameters for cat.aliases.
type CatAliasesParams struct {
	catCommonParams
	ExpandWildcards ExpandWildcards
}

// ToMap converts CatAliasesParams to a query-parameter map.
func (p *CatAliasesParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	return m
}

// CatAllocationParams are the query parameters for cat.allocation.
type CatAllocationParams struct {
	catCommonParams
	Bytes BytesFormat
}

// ToMap converts CatAllocationParams to a query-parameter map.
func (p *CatAllocationParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	return m
}

// CatCircuitBreakerParams are the query parameters for cat.circuit_breaker.
type CatCircuitBreakerParams struct {
	catCommonParams
	Bytes BytesFormat
}

// ToMap converts CatCircuitBreakerParams to a query-parameter map.
func (p *CatCircuitBreakerParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	return m
}

// CatComponentTemplatesParams are the query parameters for cat.component_templates.
type CatComponentTemplatesParams struct {
	catCommonParams
}

// ToMap converts CatComponentTemplatesParams to a query-parameter map.
func (p *CatComponentTemplatesParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	return p.toMap()
}

// CatCountParams are the query parameters for cat.count.
type CatCountParams struct {
	catCommonParams
}

// ToMap converts CatCountParams to a query-parameter map.
func (p *CatCountParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	return p.toMap()
}

// CatFielddataParams are the query parameters for cat.fielddata.
type CatFielddataParams struct {
	catCommonParams
	Bytes BytesFormat
}

// ToMap converts CatFielddataParams to a query-parameter map.
func (p *CatFielddataParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	return m
}

// CatHealthParams are the query parameters for cat.health.
type CatHealthParams struct {
	catCommonParams
	Ts *bool
}

// ToMap converts CatHealthParams to a query-parameter map.
func (p *CatHealthParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	setBoolParam(m, "ts", p.Ts)
	return m
}

// CatIndicesParams are the query parameters for cat.indices.
type CatIndicesParams struct {
	catCommonParams
	Bytes                   BytesFormat
	ExpandWildcards         ExpandWildcards
	Health                  HealthStatus
	IncludeUnloadedSegments *bool
	Pri                     *bool
}

// ToMap converts CatIndicesParams to a query-parameter map.
func (p *CatIndicesParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	if p.ExpandWildcards != "" {
		m["expand_wildcards"] = string(p.ExpandWildcards)
	}
	if p.Health != "" {
		m["health"] = string(p.Health)
	}
	setBoolParam(m, "include_unloaded_segments", p.IncludeUnloadedSegments)
	setBoolParam(m, "pri", p.Pri)
	return m
}

// CatMasterParams are the query parameters for cat.master.
type CatMasterParams struct {
	catCommonParams
}

// ToMap converts CatMasterParams to a query-parameter map.
func (p *CatMasterParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	return p.toMap()
}

// CatMLDataFrameAnalyticsParams are the query parameters for cat.ml_data_frame_analytics.
type CatMLDataFrameAnalyticsParams struct {
	catCommonParams
	AllowNoMatch *bool
	Bytes        BytesFormat
	From         *int
	Size         *int
	Time         string
}

// ToMap converts CatMLDataFrameAnalyticsParams to a query-parameter map.
func (p *CatMLDataFrameAnalyticsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	if p.Time != "" {
		m["time"] = p.Time
	}
	return m
}

// CatMLDatafeedsParams are the query parameters for cat.ml_datafeeds.
type CatMLDatafeedsParams struct {
	catCommonParams
	AllowNoMatch *bool
	From         *int
	Size         *int
	Time         string
}

// ToMap converts CatMLDatafeedsParams to a query-parameter map.
func (p *CatMLDatafeedsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	if p.Time != "" {
		m["time"] = p.Time
	}
	return m
}

// CatMLJobsParams are the query parameters for cat.ml_jobs.
type CatMLJobsParams struct {
	catCommonParams
	AllowNoMatch *bool
	Bytes        BytesFormat
	From         *int
	Size         *int
}

// ToMap converts CatMLJobsParams to a query-parameter map.
func (p *CatMLJobsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	return m
}

// CatMLTrainedModelsParams are the query parameters for cat.ml_trained_models.
type CatMLTrainedModelsParams struct {
	catCommonParams
	AllowNoMatch *bool
	Bytes        BytesFormat
	From         *int
	Size         *int
}

// ToMap converts CatMLTrainedModelsParams to a query-parameter map.
func (p *CatMLTrainedModelsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	return m
}

// CatNodeAttrsParams are the query parameters for cat.nodeattrs.
type CatNodeAttrsParams struct {
	catCommonParams
}

// ToMap converts CatNodeAttrsParams to a query-parameter map.
func (p *CatNodeAttrsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	return p.toMap()
}

// CatNodesParams are the query parameters for cat.nodes.
type CatNodesParams struct {
	catCommonParams
	Bytes  BytesFormat
	FullId *bool
	Time   string
}

// ToMap converts CatNodesParams to a query-parameter map.
func (p *CatNodesParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	setBoolParam(m, "full_id", p.FullId)
	if p.Time != "" {
		m["time"] = p.Time
	}
	return m
}

// CatPendingTasksParams are the query parameters for cat.pending_tasks.
type CatPendingTasksParams struct {
	catCommonParams
}

// ToMap converts CatPendingTasksParams to a query-parameter map.
func (p *CatPendingTasksParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	return p.toMap()
}

// CatPluginsParams are the query parameters for cat.plugins.
type CatPluginsParams struct {
	catCommonParams
}

// ToMap converts CatPluginsParams to a query-parameter map.
func (p *CatPluginsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	return p.toMap()
}

// CatRecoveryParams are the query parameters for cat.recovery.
type CatRecoveryParams struct {
	catCommonParams
	ActiveOnly *bool
	Bytes      BytesFormat
	Detailed   *bool
}

// ToMap converts CatRecoveryParams to a query-parameter map.
func (p *CatRecoveryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	setBoolParam(m, "active_only", p.ActiveOnly)
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	setBoolParam(m, "detailed", p.Detailed)
	return m
}

// CatRepositoriesParams are the query parameters for cat.repositories.
type CatRepositoriesParams struct {
	catCommonParams
}

// ToMap converts CatRepositoriesParams to a query-parameter map.
func (p *CatRepositoriesParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	return p.toMap()
}

// CatSegmentsParams are the query parameters for cat.segments.
type CatSegmentsParams struct {
	catCommonParams
	Bytes   BytesFormat
	Verbose *bool
}

// ToMap converts CatSegmentsParams to a query-parameter map.
func (p *CatSegmentsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	setBoolParam(m, "verbose", p.Verbose)
	return m
}

// CatShardsParams are the query parameters for cat.shards.
type CatShardsParams struct {
	catCommonParams
	Bytes BytesFormat
}

// ToMap converts CatShardsParams to a query-parameter map.
func (p *CatShardsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	if p.Bytes != "" {
		m["bytes"] = string(p.Bytes)
	}
	return m
}

// CatSnapshotsParams are the query parameters for cat.snapshots.
type CatSnapshotsParams struct {
	catCommonParams
	IgnoreUnavailable *bool
	Time              string
}

// ToMap converts CatSnapshotsParams to a query-parameter map.
func (p *CatSnapshotsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.Time != "" {
		m["time"] = p.Time
	}
	return m
}

// CatTasksParams are the query parameters for cat.tasks.
type CatTasksParams struct {
	catCommonParams
	Actions      []string
	Detailed     *bool
	NodeId       []string
	ParentTaskId string
}

// ToMap converts CatTasksParams to a query-parameter map.
func (p *CatTasksParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	setListParam(m, "actions", p.Actions)
	setBoolParam(m, "detailed", p.Detailed)
	setListParam(m, "node_id", p.NodeId)
	if p.ParentTaskId != "" {
		m["parent_task_id"] = p.ParentTaskId
	}
	return m
}

// CatTemplatesParams are the query parameters for cat.templates.
type CatTemplatesParams struct {
	catCommonParams
}

// ToMap converts CatTemplatesParams to a query-parameter map.
func (p *CatTemplatesParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	return p.toMap()
}

// CatThreadPoolParams are the query parameters for cat.thread_pool.
type CatThreadPoolParams struct {
	catCommonParams
	Size *int
	Time string
}

// ToMap converts CatThreadPoolParams to a query-parameter map.
func (p *CatThreadPoolParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	if p.Time != "" {
		m["time"] = p.Time
	}
	return m
}

// CatTransformsParams are the query parameters for cat.transforms.
type CatTransformsParams struct {
	catCommonParams
	AllowNoMatch *bool
	From         *int
	Size         *int
	Time         string
}

// ToMap converts CatTransformsParams to a query-parameter map.
func (p *CatTransformsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.toMap()
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	if p.Time != "" {
		m["time"] = p.Time
	}
	return m
}

// ---- Row structs ----

// CatAliasesResponse is a slice of cat.aliases rows.
type CatAliasesResponse []CatAliasesRow

// CatAliasesRow is a single cat.aliases row.
type CatAliasesRow struct {
	Alias         string `json:"alias"`
	Index         string `json:"index"`
	Filter        string `json:"filter"`
	RoutingIndex  string `json:"routing.index"`
	RoutingSearch string `json:"routing.search"`
	IsWriteIndex  string `json:"is_write_index"`
}

// CatAllocationResponse is a slice of cat.allocation rows.
type CatAllocationResponse []CatAllocationRow

// CatAllocationRow is a single cat.allocation row.
type CatAllocationRow struct {
	Shards      int    `json:"shards"`
	Disk        string `json:"disk"`
	DiskIndex   string `json:"disk.indices"`
	DiskUsed    string `json:"disk.used"`
	DiskAvail   string `json:"disk.avail"`
	DiskTotal   string `json:"disk.total"`
	DiskPercent string `json:"disk.percent"`
	Host        string `json:"host"`
	IP          string `json:"ip"`
	Node        string `json:"node"`
}

// CatCircuitBreakerResponse is a slice of cat.circuit_breaker rows.
type CatCircuitBreakerResponse []CatCircuitBreakerRow

// CatCircuitBreakerRow is a single cat.circuit_breaker row.
type CatCircuitBreakerRow struct {
	Node          string `json:"node"`
	Field         string `json:"field"`
	Breaker       string `json:"breaker"`
	EstimatedSize string `json:"estimated_size"`
	LimitSize     string `json:"limit_size"`
	Overhead      string `json:"overhead"`
	Tripped       string `json:"tripped"`
	Limit         string `json:"limit"`
}

// CatComponentTemplatesResponse is a slice of cat.component_templates rows.
type CatComponentTemplatesResponse []CatComponentTemplatesRow

// CatComponentTemplatesRow is a single cat.component_templates row.
type CatComponentTemplatesRow struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// CatCountResponse is a slice of cat.count rows.
type CatCountResponse []CatCountRow

// CatCountRow is a single cat.count row.
type CatCountRow struct {
	Epoch     string `json:"epoch"`
	Timestamp string `json:"timestamp"`
	Count     string `json:"count"`
}

// CatFielddataResponse is a slice of cat.fielddata rows.
type CatFielddataResponse []CatFielddataRow

// CatFielddataRow is a single cat.fielddata row.
type CatFielddataRow struct {
	Node  string `json:"node"`
	Field string `json:"field"`
	Size  string `json:"size"`
}

// CatHealthResponse is a slice of cat.health rows.
type CatHealthResponse []CatHealthRow

// CatHealthRow is a single cat.health row.
type CatHealthRow struct {
	Epoch               string `json:"epoch"`
	Timestamp           string `json:"timestamp"`
	Cluster             string `json:"cluster"`
	Status              string `json:"status"`
	NodeTotal           string `json:"node.total"`
	NodeData            string `json:"node.data"`
	Shards              string `json:"shards"`
	Pri                 string `json:"pri"`
	Relo                string `json:"relo"`
	Init                string `json:"init"`
	Unassign            string `json:"unassign"`
	PendingTasks        string `json:"pending_tasks"`
	MaxTaskWait         string `json:"max_task_wait"`
	ActiveShardsPercent string `json:"active_shards_percent"`
}

// CatIndicesResponse is a slice of cat.indices rows.
type CatIndicesResponse []CatIndicesRow

// CatIndicesRow is a single cat.indices row.
type CatIndicesRow struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	UUID         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
}

// CatMasterResponse is a slice of cat.master rows.
type CatMasterResponse []CatMasterRow

// CatMasterRow is a single cat.master row.
type CatMasterRow struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	IP   string `json:"ip"`
	Node string `json:"node"`
}

// CatMLDataFrameAnalyticsResponse is a slice of cat.ml_data_frame_analytics rows.
type CatMLDataFrameAnalyticsResponse []CatMLDataFrameAnalyticsRow

// CatMLDataFrameAnalyticsRow is a single cat.ml_data_frame_analytics row.
type CatMLDataFrameAnalyticsRow struct {
	ID       string `json:"id"`
	State    string `json:"state"`
	Progress string `json:"progress"`
}

// CatMLDatafeedsResponse is a slice of cat.ml_datafeeds rows.
type CatMLDatafeedsResponse []CatMLDatafeedsRow

// CatMLDatafeedsRow is a single cat.ml_datafeeds row.
type CatMLDatafeedsRow struct {
	ID         string `json:"id"`
	State      string `json:"state"`
	Assignment string `json:"node.assignment"`
}

// CatMLJobsResponse is a slice of cat.ml_jobs rows.
type CatMLJobsResponse []CatMLJobsRow

// CatMLJobsRow is a single cat.ml_jobs row.
type CatMLJobsRow struct {
	ID         string `json:"id"`
	State      string `json:"state"`
	OpenedTime string `json:"opened_time"`
}

// CatMLTrainedModelsResponse is a slice of cat.ml_trained_models rows.
type CatMLTrainedModelsResponse []CatMLTrainedModelsRow

// CatMLTrainedModelsRow is a single cat.ml_trained_models row.
type CatMLTrainedModelsRow struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Algorithm string `json:"algorithm"`
}

// CatNodeAttrsResponse is a slice of cat.nodeattrs rows.
type CatNodeAttrsResponse []CatNodeAttrsRow

// CatNodeAttrsRow is a single cat.nodeattrs row.
type CatNodeAttrsRow struct {
	Node  string `json:"node"`
	Attr  string `json:"attr"`
	Value string `json:"value"`
}

// CatNodesResponse is a slice of cat.nodes rows.
type CatNodesResponse []CatNodesRow

// CatNodesRow is a single cat.nodes row.
type CatNodesRow struct {
	IP          string `json:"ip"`
	HeapPercent string `json:"heap.percent"`
	RamPercent  string `json:"ram.percent"`
	CPU         string `json:"cpu"`
	Load1m      string `json:"load_1m"`
	Load5m      string `json:"load_5m"`
	Load15m     string `json:"load_15m"`
	NodeRole    string `json:"node.role"`
	Master      string `json:"master"`
	Name        string `json:"name"`
}

// CatPendingTasksResponse is a slice of cat.pending_tasks rows.
type CatPendingTasksResponse []CatPendingTasksRow

// CatPendingTasksRow is a single cat.pending_tasks row.
type CatPendingTasksRow struct {
	InsertOrder string `json:"insert_order"`
	TimeInQueue string `json:"time_in_queue"`
	Priority    string `json:"priority"`
	Source      string `json:"source"`
}

// CatPluginsResponse is a slice of cat.plugins rows.
type CatPluginsResponse []CatPluginsRow

// CatPluginsRow is a single cat.plugins row.
type CatPluginsRow struct {
	Name                 string `json:"name"`
	Version              string `json:"version"`
	ElasticsearchVersion string `json:"elasticsearch_version"`
	Description          string `json:"description"`
}

// CatRecoveryResponse is a slice of cat.recovery rows.
type CatRecoveryResponse []CatRecoveryRow

// CatRecoveryRow is a single cat.recovery row.
type CatRecoveryRow struct {
	Index           string `json:"index"`
	Shard           string `json:"shard"`
	StartTime       string `json:"start_time"`
	StartTimeMillis string `json:"start_time_millis"`
	StopTime        string `json:"stop_time"`
	StopTimeMillis  string `json:"stop_time_millis"`
	TotalTime       string `json:"total_time"`
	TotalTimeMillis string `json:"total_time_millis"`
	Source          string `json:"source"`
	Target          string `json:"target"`
	Type            string `json:"type"`
	Stage           string `json:"stage"`
	Repository      string `json:"repository"`
	Snapshot        string `json:"snapshot"`
	Files           string `json:"files"`
	FilesRecovered  string `json:"files_recovered"`
	FilesPercent    string `json:"files_percent"`
	Bytes           string `json:"bytes"`
	BytesRecovered  string `json:"bytes_recovered"`
	BytesPercent    string `json:"bytes_percent"`
}

// CatRepositoriesResponse is a slice of cat.repositories rows.
type CatRepositoriesResponse []CatRepositoriesRow

// CatRepositoriesRow is a single cat.repositories row.
type CatRepositoriesRow struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// CatSegmentsResponse is a slice of cat.segments rows.
type CatSegmentsResponse []CatSegmentsRow

// CatSegmentsRow is a single cat.segments row.
type CatSegmentsRow struct {
	Index       string `json:"index"`
	Shard       string `json:"shard"`
	Prirep      string `json:"prirep"`
	IP          string `json:"ip"`
	Segment     string `json:"segment"`
	Generation  string `json:"generation"`
	DocsCount   string `json:"docs.count"`
	DocsDeleted string `json:"docs.deleted"`
	Size        string `json:"size"`
	SizeMemory  string `json:"size.memory"`
	Committed   string `json:"committed"`
	Searchable  string `json:"searchable"`
	Version     string `json:"version"`
	Compound    string `json:"compound"`
}

// CatShardsResponse is a slice of cat.shards rows.
type CatShardsResponse []CatShardsRow

// CatShardsRow is a single cat.shards row.
type CatShardsRow struct {
	Index  string `json:"index"`
	Shard  string `json:"shard"`
	Prirep string `json:"prirep"`
	State  string `json:"state"`
	Docs   string `json:"docs"`
	Store  string `json:"store"`
	IP     string `json:"ip"`
	Node   string `json:"node"`
}

// CatSnapshotsResponse is a slice of cat.snapshots rows.
type CatSnapshotsResponse []CatSnapshotsRow

// CatSnapshotsRow is a single cat.snapshots row.
type CatSnapshotsRow struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
	Status     string `json:"status"`
	StartEpoch string `json:"start_epoch"`
	StartTime  string `json:"start_time"`
	EndEpoch   string `json:"end_epoch"`
	EndTime    string `json:"end_time"`
	Duration   string `json:"duration"`
	Indices    string `json:"indices"`
}

// CatTasksResponse is a slice of cat.tasks rows.
type CatTasksResponse []CatTasksRow

// CatTasksRow is a single cat.tasks row.
type CatTasksRow struct {
	ID         string `json:"id"`
	Action     string `json:"action"`
	TaskType   string `json:"type"`
	Parent     string `json:"parent"`
	StartEpoch string `json:"start_epoch"`
	StartTime  string `json:"start_time"`
	Node       string `json:"node"`
}

// CatTemplatesResponse is a slice of cat.templates rows.
type CatTemplatesResponse []CatTemplatesRow

// CatTemplatesRow is a single cat.templates row.
type CatTemplatesRow struct {
	Name          string `json:"name"`
	IndexPatterns string `json:"index_patterns"`
	Order         string `json:"order"`
	Version       string `json:"version,omitempty"`
}

// CatThreadPoolResponse is a slice of cat.thread_pool rows.
type CatThreadPoolResponse []CatThreadPoolRow

// CatThreadPoolRow is a single cat.thread_pool row.
type CatThreadPoolRow struct {
	NodeName  string `json:"node_name"`
	Name      string `json:"name"`
	Active    string `json:"active"`
	Queue     string `json:"queue"`
	Rejected  string `json:"rejected"`
	Completed string `json:"completed"`
}

// CatTransformsResponse is a slice of cat.transforms rows.
type CatTransformsResponse []CatTransformsRow

// CatTransformsRow is a single cat.transforms row.
type CatTransformsRow struct {
	ID         string `json:"id"`
	State      string `json:"state"`
	Checkpoint string `json:"checkpoint"`
}
