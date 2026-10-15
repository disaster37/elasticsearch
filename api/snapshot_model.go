package api

import (
	"strconv"

	"github.com/disaster37/elasticsearch/v9/types"
)

// SnapshotCreateRepositoryParams are the query parameters for create repository.
type SnapshotCreateRepositoryParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
	Verify        *bool
}

// ToMap converts SnapshotCreateRepositoryParams to a query-parameter map.
func (p *SnapshotCreateRepositoryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	setBoolParam(m, "verify", p.Verify)
	return mergeParams(m, p.Common)
}

// SnapshotGetRepositoryParams are the query parameters for get repository.
type SnapshotGetRepositoryParams struct {
	Common        *CommonParams
	Local         *bool
	MasterTimeout string
}

// ToMap converts SnapshotGetRepositoryParams to a query-parameter map.
func (p *SnapshotGetRepositoryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "local", p.Local)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// SnapshotDeleteRepositoryParams are the query parameters for delete repository.
type SnapshotDeleteRepositoryParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
}

// ToMap converts SnapshotDeleteRepositoryParams to a query-parameter map.
func (p *SnapshotDeleteRepositoryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// SnapshotVerifyRepositoryParams are the query parameters for verify repository.
type SnapshotVerifyRepositoryParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
}

// ToMap converts SnapshotVerifyRepositoryParams to a query-parameter map.
func (p *SnapshotVerifyRepositoryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// SnapshotRepositoryAnalyzeParams are the query parameters for repository analyze.
type SnapshotRepositoryAnalyzeParams struct {
	Common                *CommonParams
	BlobCount             *int
	Concurrency           *int
	Detailed              *bool
	EarlyReadNodeCount    *int
	MaxBlobSize           string
	MaxTotalDataSize      string
	RareActionProbability *float64
	RarelyAbortWrites     *bool
	ReadNodeCount         *int
	Seed                  *int
	Timeout               string
}

// ToMap converts SnapshotRepositoryAnalyzeParams to a query-parameter map.
func (p *SnapshotRepositoryAnalyzeParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.BlobCount != nil {
		m["blob_count"] = intStr(*p.BlobCount)
	}
	if p.Concurrency != nil {
		m["concurrency"] = intStr(*p.Concurrency)
	}
	setBoolParam(m, "detailed", p.Detailed)
	if p.EarlyReadNodeCount != nil {
		m["early_read_node_count"] = intStr(*p.EarlyReadNodeCount)
	}
	if p.MaxBlobSize != "" {
		m["max_blob_size"] = p.MaxBlobSize
	}
	if p.MaxTotalDataSize != "" {
		m["max_total_data_size"] = p.MaxTotalDataSize
	}
	if p.RareActionProbability != nil {
		m["rare_action_probability"] = strconv.FormatFloat(*p.RareActionProbability, 'g', -1, 64)
	}
	setBoolParam(m, "rarely_abort_writes", p.RarelyAbortWrites)
	if p.ReadNodeCount != nil {
		m["read_node_count"] = intStr(*p.ReadNodeCount)
	}
	if p.Seed != nil {
		m["seed"] = intStr(*p.Seed)
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// SnapshotRepositoryVerifyIntegrityParams are the query parameters for verify_integrity.
type SnapshotRepositoryVerifyIntegrityParams struct {
	Common                    *CommonParams
	IndexSnapshotVerification *bool
	SnapshotVerification      *bool
	Timeout                   string
}

// ToMap converts SnapshotRepositoryVerifyIntegrityParams to a query-parameter map.
func (p *SnapshotRepositoryVerifyIntegrityParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "index_snapshot_verification", p.IndexSnapshotVerification)
	setBoolParam(m, "snapshot_verification", p.SnapshotVerification)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// SnapshotCreateRequest is the request for PUT /_snapshot/{repository}/{snapshot}.
type SnapshotCreateRequest struct {
	Repository string `validate:"required"`
	Snapshot   string `validate:"required"`
	Body       any
	Params     *SnapshotCreateParams
}

// Validate validates the SnapshotCreateRequest.
func (r *SnapshotCreateRequest) Validate() error { return validationError(validate.Struct(r)) }

// SnapshotCreateParams are the query parameters for create snapshot.
type SnapshotCreateParams struct {
	Common            *CommonParams
	MasterTimeout     string
	WaitForCompletion *bool
}

// ToMap converts SnapshotCreateParams to a query-parameter map.
func (p *SnapshotCreateParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// SnapshotGetRequest is the request for GET /_snapshot/{repository}/{snapshot}.
type SnapshotGetRequest struct {
	Repository string `validate:"required"`
	Snapshots  []string
	Params     *SnapshotGetParams
}

// Validate validates the SnapshotGetRequest.
func (r *SnapshotGetRequest) Validate() error { return validationError(validate.Struct(r)) }

// SnapshotGetParams are the query parameters for get snapshot.
type SnapshotGetParams struct {
	Common            *CommonParams
	After             string
	FromSortValue     string
	IgnoreUnavailable *bool
	IncludeRepository *bool
	IndexDetails      *bool
	IndexNames        *bool
	MasterTimeout     string
	Offset            *int
	Order             SortOrder
	Size              *int
	SlmPolicyFilter   string
	Sort              SnapshotSort
	Verbose           *bool
}

// ToMap converts SnapshotGetParams to a query-parameter map.
func (p *SnapshotGetParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.After != "" {
		m["after"] = p.After
	}
	if p.FromSortValue != "" {
		m["from_sort_value"] = p.FromSortValue
	}
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	setBoolParam(m, "include_repository", p.IncludeRepository)
	setBoolParam(m, "index_details", p.IndexDetails)
	setBoolParam(m, "index_names", p.IndexNames)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Offset != nil {
		m["offset"] = intStr(*p.Offset)
	}
	if p.Order != "" {
		m["order"] = string(p.Order)
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	if p.SlmPolicyFilter != "" {
		m["slm_policy_filter"] = p.SlmPolicyFilter
	}
	if p.Sort != "" {
		m["sort"] = string(p.Sort)
	}
	setBoolParam(m, "verbose", p.Verbose)
	return mergeParams(m, p.Common)
}

// SnapshotDeleteSpecParams are the opt-in query parameters for snapshot.delete.
// The existing Snapshot field may be a comma-separated list.
type SnapshotDeleteSpecParams struct {
	Common            *CommonParams
	MasterTimeout     string
	WaitForCompletion *bool
}

// ToMap converts SnapshotDeleteSpecParams to a query-parameter map.
func (p *SnapshotDeleteSpecParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// SnapshotDeleteRequest is the request for DELETE /_snapshot/{repository}/{snapshot}.
type SnapshotDeleteRequest struct {
	Repository string `validate:"required"`
	Snapshot   string `validate:"required"`
	Params     *SnapshotDeleteSpecParams
}

// Validate validates the SnapshotDeleteRequest.
func (r *SnapshotDeleteRequest) Validate() error { return validationError(validate.Struct(r)) }

// SnapshotRestoreRequest is the request for POST /_snapshot/{repository}/{snapshot}/_restore.
type SnapshotRestoreRequest struct {
	Repository string `validate:"required"`
	Snapshot   string `validate:"required"`
	Body       any
	Params     *SnapshotRestoreParams
}

// Validate validates the SnapshotRestoreRequest.
func (r *SnapshotRestoreRequest) Validate() error { return validationError(validate.Struct(r)) }

// SnapshotRestoreParams are the query parameters for restore.
type SnapshotRestoreParams struct {
	Common            *CommonParams
	MasterTimeout     string
	WaitForCompletion *bool
}

// ToMap converts SnapshotRestoreParams to a query-parameter map.
func (p *SnapshotRestoreParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// SnapshotStatusRequest is the request for GET /_snapshot/status.
type SnapshotStatusRequest struct {
	Repository string
	Snapshots  []string
	Params     *SnapshotStatusParams
}

// SnapshotStatusParams are the query parameters for snapshot status.
type SnapshotStatusParams struct {
	Common            *CommonParams
	IgnoreUnavailable *bool
	MasterTimeout     string
}

// ToMap converts SnapshotStatusParams to a query-parameter map.
func (p *SnapshotStatusParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "ignore_unavailable", p.IgnoreUnavailable)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// SnapshotCloneRequest is the request for PUT /_snapshot/{repository}/{snapshot}/_clone.
type SnapshotCloneRequest struct {
	Repository     string `validate:"required"`
	Snapshot       string `validate:"required"`
	TargetSnapshot string `validate:"required"`
	Body           any
	Params         *MasterTimeoutSpecParams
}

// Validate validates the SnapshotCloneRequest.
func (r *SnapshotCloneRequest) Validate() error { return validationError(validate.Struct(r)) }

// SnapshotCleanupRepositoryParams are the query parameters for cleanup repository.
type SnapshotCleanupRepositoryParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
}

// ToMap converts SnapshotCleanupRepositoryParams to a query-parameter map.
func (p *SnapshotCleanupRepositoryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ---- Response structs ----

// SnapshotRepository describes a snapshot repository.
type SnapshotRepository struct {
	Type     string            `json:"type"`
	Settings map[string]string `json:"settings,omitempty"`
	UUID     string            `json:"uuid,omitempty"`
}

// SnapshotInfo describes a single snapshot.
type SnapshotInfo struct {
	Snapshot           string            `json:"snapshot"`
	UUID               string            `json:"uuid,omitempty"`
	Repository         string            `json:"repository,omitempty"`
	Indices            []string          `json:"indices,omitempty"`
	DataStreams        []string          `json:"data_streams,omitempty"`
	State              string            `json:"state,omitempty"`
	StartTime          string            `json:"start_time,omitempty"`
	StartTimeInMillis  int64             `json:"start_time_in_millis,omitempty"`
	EndTime            string            `json:"end_time,omitempty"`
	EndTimeInMillis    int64             `json:"end_time_in_millis,omitempty"`
	DurationInMillis   int64             `json:"duration_in_millis,omitempty"`
	Failures           []map[string]any  `json:"failures,omitempty"`
	Shards             *types.ShardsInfo `json:"shards,omitempty"`
	VersionId          string            `json:"version_id,omitempty"`
	Version            string            `json:"version,omitempty"`
	IncludeGlobalState *bool             `json:"include_global_state,omitempty"`
}

// SnapshotCreateResponse is the response from create snapshot.
type SnapshotCreateResponse struct {
	Snapshot SnapshotInfo `json:"snapshot"`
}

// SnapshotGetResponse is the response from get snapshot.
type SnapshotGetResponse struct {
	Snapshots []SnapshotInfo `json:"snapshots"`
	Total     int            `json:"total,omitempty"`
}

// SnapshotRestoreResponse is the response from restore.
type SnapshotRestoreResponse struct {
	Accepted bool `json:"accepted,omitempty"`
	Snapshot struct {
		Snapshot string            `json:"snapshot"`
		Indices  []string          `json:"indices"`
		Shards   *types.ShardsInfo `json:"shards,omitempty"`
	} `json:"snapshot"`
}

// SnapshotStatusResponse is the response from snapshot status.
type SnapshotStatusResponse struct {
	Snapshots []SnapshotStatus `json:"snapshots"`
}

// SnapshotStatus describes a snapshot's status.
type SnapshotStatus struct {
	Snapshot           string                         `json:"snapshot"`
	Repository         string                         `json:"repository"`
	UUID               string                         `json:"uuid,omitempty"`
	State              string                         `json:"state"`
	IncludeGlobalState *bool                          `json:"include_global_state,omitempty"`
	Shards             *types.ShardsInfo              `json:"shards,omitempty"`
	Indices            map[string]SnapshotIndexStatus `json:"indices,omitempty"`
}

// SnapshotIndexStatus is the per-index snapshot status.
type SnapshotIndexStatus struct {
	Shards *types.ShardsInfo `json:"shards,omitempty"`
	Stats  SnapshotStats     `json:"stats,omitempty"`
}

// SnapshotStats holds snapshot statistics.
type SnapshotStats struct {
	Incremental       SnapshotStatsCount `json:"incremental,omitempty"`
	Total             SnapshotStatsCount `json:"total,omitempty"`
	StartTimeInMillis int64              `json:"start_time_in_millis,omitempty"`
	TimeInMillis      int64              `json:"time_in_millis,omitempty"`
}

// SnapshotStatsCount holds file/size counts.
type SnapshotStatsCount struct {
	FileCount   int   `json:"file_count"`
	SizeInBytes int64 `json:"size_in_bytes"`
}

// SnapshotVerifyRepositoryResponse is the response from verify repository.
type SnapshotVerifyRepositoryResponse struct {
	Nodes map[string]struct {
		Name string `json:"name"`
	} `json:"nodes"`
}

// SnapshotRepositoryAnalyzeResponse is the response from repository analyze.
type SnapshotRepositoryAnalyzeResponse struct {
	BlobCount        int              `json:"blob_count"`
	Concurrency      int              `json:"concurrency"`
	ReadNodeCount    int              `json:"read_node_count"`
	CoordinatingNode string           `json:"coordinating_node"`
	Summary          map[string]any   `json:"summary,omitempty"`
	Details          []map[string]any `json:"details,omitempty"`
}

// SnapshotRepositoryVerifyIntegrityResponse is the response from verify_integrity.
type SnapshotRepositoryVerifyIntegrityResponse struct {
	Repositories []struct {
		Repository string `json:"repository"`
		Verified   bool   `json:"verified"`
	} `json:"repositories"`
}

// SnapshotCleanupRepositoryResponse is the response from cleanup repository.
type SnapshotCleanupRepositoryResponse struct {
	DeletedBytes int64 `json:"deleted_bytes"`
	DeletedBlobs int64 `json:"deleted_blobs"`
}
