package api

// SlmPolicy describes a snapshot lifecycle management policy.
type SlmPolicy struct {
	Version             int            `json:"version,omitempty"`
	ModifiedDate        string         `json:"modified_date,omitempty"`
	ModifiedDateMillis  int64          `json:"modified_date_millis,omitempty"`
	Policy              map[string]any `json:"policy,omitempty"`
	Config              map[string]any `json:"config,omitempty"`
	Name                string         `json:"name,omitempty"`
	Uuid                string         `json:"uuid,omitempty"`
	SnapshotName        string         `json:"snapshot_name,omitempty"`
	Retention           map[string]any `json:"retention,omitempty"`
	NextExecutionMillis int64          `json:"next_execution_millis,omitempty"`
}

// SlmExecuteResponse is the response from execute lifecycle.
type SlmExecuteResponse struct {
	SnapshotName string `json:"snapshot_name"`
}

// SlmStatsResponse is the response from SLM stats.
type SlmStatsResponse struct {
	RetentionRuns                 int64            `json:"retention_runs"`
	RetentionFailed               int64            `json:"retention_failed"`
	RetentionTimedOut             int64            `json:"retention_timed_out"`
	RetentionDeletionTime         string           `json:"retention_deletion_time"`
	RetentionDeletionTimeMillis   int64            `json:"retention_deletion_time_millis"`
	TotalSnapshotsTaken           int64            `json:"total_snapshots_taken"`
	TotalSnapshotsFailed          int64            `json:"total_snapshots_failed"`
	TotalSnapshotDeletions        int64            `json:"total_snapshot_deletions"`
	TotalSnapshotDeletionFailures int64            `json:"total_snapshot_deletion_failures"`
	PolicyStats                   []map[string]any `json:"policy_stats,omitempty"`
}

// SlmStatusResponse is the response from SLM status.
type SlmStatusResponse struct {
	OperationMode string `json:"operation_mode"`
}
