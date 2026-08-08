package api

// TransformGetParams are the query parameters for transform get.
type TransformGetParams struct {
	Common           *CommonParams
	AllowNoMatch     *bool
	ExcludeGenerated *bool
	From             *int
	Size             *int
}

// ToMap converts TransformGetParams to a query-parameter map.
func (p *TransformGetParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	setBoolParam(m, "exclude_generated", p.ExcludeGenerated)
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	return mergeParams(m, p.Common)
}

// TransformGetStatsParams are the query parameters for transform get stats.
type TransformGetStatsParams struct {
	Common       *CommonParams
	AllowNoMatch *bool
	From         *int
	Size         *int
	Timeout      string
}

// ToMap converts TransformGetStatsParams to a query-parameter map.
func (p *TransformGetStatsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// TransformPutParams are the query parameters for transform put.
type TransformPutParams struct {
	Common          *CommonParams
	DeferValidation *bool
	Timeout         string
}

// ToMap converts TransformPutParams to a query-parameter map.
func (p *TransformPutParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "defer_validation", p.DeferValidation)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// TransformPreviewParams are the query parameters for transform preview.
type TransformPreviewParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts TransformPreviewParams to a query-parameter map.
func (p *TransformPreviewParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// TransformDeleteParams are the query parameters for transform delete.
type TransformDeleteParams struct {
	Common  *CommonParams
	Force   *bool
	Timeout string
}

// ToMap converts TransformDeleteParams to a query-parameter map.
func (p *TransformDeleteParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "force", p.Force)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// TransformStartParams are the query parameters for transform start.
type TransformStartParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts TransformStartParams to a query-parameter map.
func (p *TransformStartParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// TransformStopParams are the query parameters for transform stop.
type TransformStopParams struct {
	Common            *CommonParams
	AllowNoMatch      *bool
	Force             *bool
	Timeout           string
	WaitForCheckpoint *bool
	WaitForCompletion *bool
}

// ToMap converts TransformStopParams to a query-parameter map.
func (p *TransformStopParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	setBoolParam(m, "force", p.Force)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	setBoolParam(m, "wait_for_checkpoint", p.WaitForCheckpoint)
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// TransformResetParams are the query parameters for transform reset.
type TransformResetParams struct {
	Common  *CommonParams
	Force   *bool
	Timeout string
}

// ToMap converts TransformResetParams to a query-parameter map.
func (p *TransformResetParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "force", p.Force)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// TransformUpgradeParams are the query parameters for transform upgrade.
type TransformUpgradeParams struct {
	Common  *CommonParams
	DryRun  *bool
	Timeout string
}

// ToMap converts TransformUpgradeParams to a query-parameter map.
func (p *TransformUpgradeParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "dry_run", p.DryRun)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// TransformGetResponse is the response from transform get.
type TransformGetResponse struct {
	Count      int               `json:"count"`
	Transforms []TransformConfig `json:"transforms"`
}

// TransformConfig describes a transform configuration.
type TransformConfig struct {
	ID          string         `json:"id,omitempty"`
	Description string         `json:"description,omitempty"`
	Source      map[string]any `json:"source,omitempty"`
	Dest        map[string]any `json:"dest,omitempty"`
	Frequency   string         `json:"frequency,omitempty"`
	Sync        map[string]any `json:"sync,omitempty"`
	Pivot       map[string]any `json:"pivot,omitempty"`
	Latest      map[string]any `json:"latest,omitempty"`
	Settings    map[string]any `json:"settings,omitempty"`
	Meta        map[string]any `json:"_meta,omitempty"`
	Version     *int           `json:"version,omitempty"`
	CreateTime  int64          `json:"create_time,omitempty"`
}

// TransformStatsResponse is the response from transform stats.
type TransformStatsResponse struct {
	Count      int              `json:"count"`
	Transforms []TransformStats `json:"transforms"`
}

// TransformStats describes transform statistics.
type TransformStats struct {
	ID             string         `json:"id"`
	State          string         `json:"state"`
	Reason         string         `json:"reason,omitempty"`
	Stats          map[string]any `json:"stats,omitempty"`
	CheckpointInfo map[string]any `json:"checkpointing_info,omitempty"`
	Health         map[string]any `json:"health,omitempty"`
}

// TransformPutResponse is the response from transform put.
type TransformPutResponse struct {
	Acknowledged bool `json:"acknowledged"`
	Created      bool `json:"created"`
}

// TransformPreviewResponse is the response from transform preview.
type TransformPreviewResponse struct {
	Preview []map[string]any `json:"preview"`
}

// TransformUpgradeResponse is the response from transform upgrade.
type TransformUpgradeResponse struct {
	NeedsUpdate bool     `json:"needs_update"`
	DryRun      bool     `json:"dry_run"`
	Updated     []string `json:"updated"`
	NoAction    []string `json:"no_action"`
}

// TransformNodeStatsResponse is the response from transform node stats.
type TransformNodeStatsResponse struct {
	Nodes map[string]struct {
		Transforms map[string]any `json:"transforms"`
	} `json:"nodes"`
}
