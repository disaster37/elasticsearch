package api

// IlmExplainParams are the query parameters for ILM explain.
type IlmExplainParams struct {
	Common        *CommonParams
	MasterTimeout string
	OnlyErrors    *bool
	OnlyManaged   *bool
}

// ToMap converts IlmExplainParams to a query-parameter map.
func (p *IlmExplainParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	setBoolParam(m, "only_errors", p.OnlyErrors)
	setBoolParam(m, "only_managed", p.OnlyManaged)
	return mergeParams(m, p.Common)
}

// IlmPolicy describes an ILM policy.
type IlmPolicy struct {
	Version            int            `json:"version,omitempty"`
	ModifiedDate       string         `json:"modified_date,omitempty"`
	ModifiedDateMillis int64          `json:"modified_date_millis,omitempty"`
	Policy             map[string]any `json:"policy,omitempty"`
	InUse              bool           `json:"in_use,omitempty"`
}

// IlmStatusResponse is the response from ILM status.
type IlmStatusResponse struct {
	OperationMode string `json:"operation_mode"`
}

// IlmExplainResponse is the response from ILM explain.
type IlmExplainResponse struct {
	Indices map[string]struct {
		Index                string         `json:"index"`
		Managed              bool           `json:"managed"`
		Phase                string         `json:"phase,omitempty"`
		Action               string         `json:"action,omitempty"`
		Step                 string         `json:"step,omitempty"`
		StepInfo             map[string]any `json:"step_info,omitempty"`
		FailedStep           string         `json:"failed_step,omitempty"`
		FailedStepRetryable  *bool          `json:"failed_step_retryable,omitempty"`
		StepTime             int64          `json:"step_time_millis,omitempty"`
		PhaseTime            int64          `json:"phase_time_millis,omitempty"`
		ActionTime           int64          `json:"action_time_millis,omitempty"`
		IsAutoRetryableError *bool          `json:"is_auto_retryable_error,omitempty"`
	} `json:"indices"`
}

// IlmRemovePolicyResponse is the response from ILM remove.
type IlmRemovePolicyResponse struct {
	HasFailures   bool     `json:"has_failures"`
	FailedIndexes []string `json:"failed_indexes"`
}

// IlmMigrateToDataTiersResponse is the response from migrate_to_data_tiers.
type IlmMigrateToDataTiersResponse struct {
	DryRun                      bool     `json:"dry_run"`
	MigratedComponentTemplates  []string `json:"migrated_component_templates"`
	MigratedComposableTemplates []string `json:"migrated_composable_templates"`
	MigratedIndices             []string `json:"migrated_indices"`
	MigratedIlmPolicies         []string `json:"migrated_ilm_policies"`
	MigratedLegacyTemplates     []string `json:"migrated_legacy_templates"`
	RemovedLegacyTemplate       []string `json:"removed_legacy_template,omitempty"`
	RemovedIndexSettings        []string `json:"removed_index_settings,omitempty"`
	RemovedClusterSettings      []string `json:"removed_cluster_settings,omitempty"`
}
