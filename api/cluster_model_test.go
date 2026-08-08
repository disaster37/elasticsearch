package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- ClusterHealthParams ----

func TestUnitClusterHealthParamsToMap(t *testing.T) {
	p := &ClusterHealthParams{
		Level:                       LevelIndices,
		Local:                       ptrBool(true),
		Timeout:                     "30s",
		WaitForActiveShards:         "1",
		WaitForEvents:               "high",
		WaitForNoInitializingShards: ptrBool(true),
		WaitForNoRelocatingShards:   ptrBool(false),
		WaitForNodes:                ">=2",
		WaitForStatus:               WaitForStatusGreen,
		Common:                      &CommonParams{Pretty: ptrBool(true)},
	}
	m := p.ToMap()
	assert.Equal(t, "indices", m["level"])
	assert.Equal(t, "true", m["local"])
	assert.Equal(t, "30s", m["timeout"])
	assert.Equal(t, "1", m["wait_for_active_shards"])
	assert.Equal(t, "high", m["wait_for_events"])
	assert.Equal(t, "true", m["wait_for_no_initializing_shards"])
	assert.Equal(t, "false", m["wait_for_no_relocating_shards"])
	assert.Equal(t, ">=2", m["wait_for_nodes"])
	assert.Equal(t, "green", m["wait_for_status"])
	assert.Equal(t, "true", m["pretty"])
}

func TestUnitClusterHealthParamsToMapNil(t *testing.T) {
	var p *ClusterHealthParams
	assert.Nil(t, p.ToMap())
}

func TestUnitClusterHealthParamsToMapEmpty(t *testing.T) {
	assert.Empty(t, (&ClusterHealthParams{}).ToMap())
}

// ---- ClusterStateParams ----

func TestUnitClusterStateParamsToMap(t *testing.T) {
	p := &ClusterStateParams{
		AllowNoIndices:         ptrBool(true),
		ExpandWildcards:        ExpandWildcardOpen,
		FlatSettings:           ptrBool(true),
		IgnoreUnavailable:      ptrBool(false),
		Local:                  ptrBool(true),
		WaitForMetadataVersion: ptrInt64(5),
		WaitForTimeout:         "30s",
		Common:                 &CommonParams{Human: ptrBool(true)},
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["allow_no_indices"])
	assert.Equal(t, "open", m["expand_wildcards"])
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "false", m["ignore_unavailable"])
	assert.Equal(t, "true", m["local"])
	assert.Equal(t, "5", m["wait_for_metadata_version"])
	assert.Equal(t, "30s", m["wait_for_timeout"])
	assert.Equal(t, "true", m["human"])
}

func TestUnitClusterStateParamsToMapNil(t *testing.T) {
	var p *ClusterStateParams
	assert.Nil(t, p.ToMap())
}

// ---- ClusterStatsParams ----

func TestUnitClusterStatsParamsToMap(t *testing.T) {
	p := &ClusterStatsParams{
		FlatSettings: ptrBool(true),
		Timeout:      "30s",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitClusterStatsParamsToMapNil(t *testing.T) {
	var p *ClusterStatsParams
	assert.Nil(t, p.ToMap())
}

// ---- ClusterGetSettingsParams ----

func TestUnitClusterGetSettingsParamsToMap(t *testing.T) {
	p := &ClusterGetSettingsParams{
		FlatSettings:    ptrBool(true),
		IncludeDefaults: ptrBool(true),
		Timeout:         "10s",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "true", m["include_defaults"])
	assert.Equal(t, "10s", m["timeout"])
}

func TestUnitClusterGetSettingsParamsToMapNil(t *testing.T) {
	var p *ClusterGetSettingsParams
	assert.Nil(t, p.ToMap())
}

// ---- ClusterPutSettingsParams ----

func TestUnitClusterPutSettingsParamsToMap(t *testing.T) {
	p := &ClusterPutSettingsParams{
		FlatSettings: ptrBool(true),
		Timeout:      "10s",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "10s", m["timeout"])
}

func TestUnitClusterPutSettingsParamsToMapNil(t *testing.T) {
	var p *ClusterPutSettingsParams
	assert.Nil(t, p.ToMap())
}

// ---- ClusterAllocationExplainParams ----

func TestUnitClusterAllocationExplainParamsToMap(t *testing.T) {
	p := &ClusterAllocationExplainParams{
		IncludeDiskInfo:     ptrBool(true),
		IncludeYesDecisions: ptrBool(false),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["include_disk_info"])
	assert.Equal(t, "false", m["include_yes_decisions"])
}

func TestUnitClusterAllocationExplainParamsToMapNil(t *testing.T) {
	var p *ClusterAllocationExplainParams
	assert.Nil(t, p.ToMap())
}

// ---- ClusterRerouteParams ----

func TestUnitClusterRerouteParamsToMap(t *testing.T) {
	p := &ClusterRerouteParams{
		DryRun:      ptrBool(true),
		Explain:     ptrBool(true),
		Metric:      []string{"nodes", "routing_table"},
		RetryFailed: ptrBool(true),
		Timeout:     "30s",
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["dry_run"])
	assert.Equal(t, "true", m["explain"])
	assert.Equal(t, "nodes,routing_table", m["metric"])
	assert.Equal(t, "true", m["retry_failed"])
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitClusterRerouteParamsToMapNil(t *testing.T) {
	var p *ClusterRerouteParams
	assert.Nil(t, p.ToMap())
}

// ---- ClusterPendingTasksParams ----

func TestUnitClusterPendingTasksParamsToMap(t *testing.T) {
	p := &ClusterPendingTasksParams{Local: ptrBool(true)}
	m := p.ToMap()
	assert.Equal(t, "true", m["local"])
}

func TestUnitClusterPendingTasksParamsToMapNil(t *testing.T) {
	var p *ClusterPendingTasksParams
	assert.Nil(t, p.ToMap())
}

// ---- ClusterVotingConfigExclusionsParams ----

func TestUnitClusterVotingConfigExclusionsParamsToMap(t *testing.T) {
	p := &ClusterVotingConfigExclusionsParams{
		NodeIds:   []string{"n1", "n2"},
		NodeNames: []string{"node-a"},
		Timeout:   "30s",
	}
	m := p.ToMap()
	assert.Equal(t, "n1,n2", m["node_ids"])
	assert.Equal(t, "node-a", m["node_names"])
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitClusterVotingConfigExclusionsParamsToMapNil(t *testing.T) {
	var p *ClusterVotingConfigExclusionsParams
	assert.Nil(t, p.ToMap())
}

// ---- ClusterDeleteVotingConfigExclusionsParams ----

func TestUnitClusterDeleteVotingConfigExclusionsParamsToMap(t *testing.T) {
	p := &ClusterDeleteVotingConfigExclusionsParams{WaitForRemoval: ptrBool(true)}
	m := p.ToMap()
	assert.Equal(t, "true", m["wait_for_removal"])
}

func TestUnitClusterDeleteVotingConfigExclusionsParamsToMapNil(t *testing.T) {
	var p *ClusterDeleteVotingConfigExclusionsParams
	assert.Nil(t, p.ToMap())
}

// ---- ClusterGetComponentTemplateParams ----

func TestUnitClusterGetComponentTemplateParamsToMap(t *testing.T) {
	p := &ClusterGetComponentTemplateParams{
		FlatSettings: ptrBool(true),
		Local:        ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["flat_settings"])
	assert.Equal(t, "true", m["local"])
}

func TestUnitClusterGetComponentTemplateParamsToMapNil(t *testing.T) {
	var p *ClusterGetComponentTemplateParams
	assert.Nil(t, p.ToMap())
}

// ---- PutComponentTemplateParams ----

func TestUnitPutComponentTemplateParamsToMap(t *testing.T) {
	p := &PutComponentTemplateParams{
		Cause:   "test",
		Create:  ptrBool(true),
		Timeout: "30s",
	}
	m := p.ToMap()
	assert.Equal(t, "test", m["cause"])
	assert.Equal(t, "true", m["create"])
	assert.Equal(t, "30s", m["timeout"])
}

func TestUnitPutComponentTemplateParamsToMapNil(t *testing.T) {
	var p *PutComponentTemplateParams
	assert.Nil(t, p.ToMap())
}

func TestUnitPutComponentTemplateRequestValidate(t *testing.T) {
	require.NoError(t, (&PutComponentTemplateRequest{Name: "t", Body: map[string]any{}}).Validate())
}

func TestUnitPutComponentTemplateRequestValidateMissingName(t *testing.T) {
	err := (&PutComponentTemplateRequest{Body: map[string]any{}}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
}

func TestUnitPutComponentTemplateRequestValidateMissingBody(t *testing.T) {
	err := (&PutComponentTemplateRequest{Name: "t"}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Body")
}

// ---- ClusterDeleteComponentTemplateParams ----

func TestUnitClusterDeleteComponentTemplateParamsToMap(t *testing.T) {
	p := &ClusterDeleteComponentTemplateParams{Timeout: "10s"}
	m := p.ToMap()
	assert.Equal(t, "10s", m["timeout"])
}

func TestUnitClusterDeleteComponentTemplateParamsToMapNil(t *testing.T) {
	var p *ClusterDeleteComponentTemplateParams
	assert.Nil(t, p.ToMap())
}
