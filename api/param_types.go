package api

// Refresh controls when changes become visible to search.
// Valid values: RefreshTrue, RefreshFalse, RefreshWaitFor.
type Refresh string

const (
	RefreshTrue    Refresh = "true"
	RefreshFalse   Refresh = "false"
	RefreshWaitFor Refresh = "wait_for"
)

// VersionType controls versioning strategy for document operations.
type VersionType string

const (
	VersionTypeInternal    VersionType = "internal"
	VersionTypeExternal    VersionType = "external"
	VersionTypeExternalGte VersionType = "external_gte"
)

// SearchType controls how distributed term frequencies affect scoring.
type SearchType string

const (
	SearchTypeQueryThenFetch    SearchType = "query_then_fetch"
	SearchTypeDfsQueryThenFetch SearchType = "dfs_query_then_fetch"
)

// OpType controls the operation type for index/create operations.
type OpType string

const (
	OpTypeIndex  OpType = "index"
	OpTypeCreate OpType = "create"
)

// HealthStatus values.
type HealthStatus string

const (
	HealthStatusGreen  HealthStatus = "green"
	HealthStatusYellow HealthStatus = "yellow"
	HealthStatusRed    HealthStatus = "red"
)

// ExpandWildcards values. Can be combined (e.g. "open,hidden").
type ExpandWildcards string

const (
	ExpandWildcardOpen   ExpandWildcards = "open"
	ExpandWildcardClosed ExpandWildcards = "closed"
	ExpandWildcardHidden ExpandWildcards = "hidden"
	ExpandWildcardNone   ExpandWildcards = "none"
	ExpandWildcardAll    ExpandWildcards = "all"
)

// BytesFormat values for cat/size endpoints.
type BytesFormat string

const (
	BytesB  BytesFormat = "b"
	BytesKb BytesFormat = "kb"
	BytesMb BytesFormat = "mb"
	BytesGb BytesFormat = "gb"
	BytesTb BytesFormat = "tb"
	BytesPb BytesFormat = "pb"
)

// Level values for stats endpoints.
type Level string

const (
	LevelCluster Level = "cluster"
	LevelIndices Level = "indices"
	LevelShards  Level = "shards"
)

// SuggestMode values.
type SuggestMode string

const (
	SuggestModeMissing SuggestMode = "missing"
	SuggestModePopular SuggestMode = "popular"
	SuggestModeAlways  SuggestMode = "always"
)

// GroupBy values for tasks listing.
type GroupBy string

const (
	GroupByNodes   GroupBy = "nodes"
	GroupByParents GroupBy = "parents"
	GroupByNone    GroupBy = "none"
)

// Conflicts values for by-query operations.
type Conflicts string

const (
	ConflictsAbort   Conflicts = "abort"
	ConflictsProceed Conflicts = "proceed"
)

// DefaultOperator values.
type DefaultOperator string

const (
	DefaultOperatorAnd DefaultOperator = "AND"
	DefaultOperatorOr  DefaultOperator = "OR"
)

// SortOrder values.
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// SnapshotSort values for snapshot.get sort.
type SnapshotSort string

const (
	SnapshotSortStartTime        SnapshotSort = "start_time"
	SnapshotSortDuration         SnapshotSort = "duration"
	SnapshotSortName             SnapshotSort = "name"
	SnapshotSortRepository       SnapshotSort = "repository"
	SnapshotSortIndexCount       SnapshotSort = "index_count"
	SnapshotSortShardCount       SnapshotSort = "shard_count"
	SnapshotSortFailedShardCount SnapshotSort = "failed_shard_count"
)

// CatFormat values. The client always requests JSON.
type CatFormat string

const (
	CatFormatJSON CatFormat = "json"
	CatFormatText CatFormat = "text"
)

// WaitForStatus values.
type WaitForStatus string

const (
	WaitForStatusGreen  WaitForStatus = "green"
	WaitForStatusYellow WaitForStatus = "yellow"
	WaitForStatusRed    WaitForStatus = "red"
)

// TimeType is a generic string alias for time/duration values ("30s", "1m").
type TimeType string
