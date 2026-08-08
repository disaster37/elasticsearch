package elasticsearch

import "github.com/disaster37/elasticsearch/v9/types"

// CommonParams are the query parameters common to all Elasticsearch API
// requests. They are forwarded to all service methods.
type CommonParams = types.CommonParams

// AcknowledgedResponse is the common response for operations that are simply
// acknowledged or rejected. Returned by most indices and cluster operations.
type AcknowledgedResponse = types.AcknowledgedResponse

// ShardsInfo represents the "_shards" section of most Elasticsearch responses.
type ShardsInfo = types.ShardsInfo

// ShardOperationFailedException describes a single shard failure in a
// [ShardsInfo] response.
type ShardOperationFailedException = types.ShardOperationFailedException

// BroadcastResponse is the common response for broadcast-style operations.
type BroadcastResponse = types.BroadcastResponse

// FailedNodeException represents a failure that occurred on a specific node
// during a multi-node operation.
type FailedNodeException = types.FailedNodeException

// ElasticsearchError is the structured error returned by Elasticsearch when a
// request fails. Use [IsNotFound], [IsConflict] and [IsUnauthorized] for
// common checks.
type ElasticsearchError = types.ElasticsearchError

// ElasticsearchErrorDetails contains the detailed error payload of an
// [ElasticsearchError].
type ElasticsearchErrorDetails = types.ElasticsearchErrorDetails

// ScriptErrorPosition describes where in a script an error occurred.
type ScriptErrorPosition = types.ScriptErrorPosition

// DocumentVersion carries the seq_no and primary_term used for optimistic
// concurrency control in document write operations.
type DocumentVersion = types.DocumentVersion

// UnixMilliTime is a [time.Time] that serializes to/from Unix milliseconds.
type UnixMilliTime = types.UnixMilliTime

// ListResponse is a generic list-response wrapper used by several services.
type ListResponse[T any] = types.ListResponse[T]
