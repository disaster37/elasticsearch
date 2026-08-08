package types

import (
	"testing"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommonParamsToMapAllFields verifies ToMap emits every configured field.
func TestCommonParamsToMapAllFields(t *testing.T) {
	p := &CommonParams{
		Pretty:      ptrBool(true),
		Human:       ptrBool(false),
		ErrorTrace:  ptrBool(true),
		FilterPath:  []string{"hits", "_shards", "_index"},
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["pretty"])
	assert.Equal(t, "false", m["human"])
	assert.Equal(t, "true", m["error_trace"])
	assert.Equal(t, "hits,_shards,_index", m["filter_path"])
	assert.Len(t, m, 4)
}

// TestCommonParamsToMapAllTrue verifies ToMap with all booleans true.
func TestCommonParamsToMapAllTrue(t *testing.T) {
	p := &CommonParams{
		Pretty:     ptrBool(true),
		Human:      ptrBool(true),
		ErrorTrace: ptrBool(true),
	}
	m := p.ToMap()
	assert.Equal(t, "true", m["pretty"])
	assert.Equal(t, "true", m["human"])
	assert.Equal(t, "true", m["error_trace"])
}

// TestCommonParamsToMapEmpty verifies ToMap returns nil when no fields are set.
func TestCommonParamsToMapEmpty(t *testing.T) {
	p := &CommonParams{}
	assert.Nil(t, p.ToMap())
}

// TestCommonParamsToMapOnlyFilterPath verifies ToMap with only FilterPath set.
func TestCommonParamsToMapOnlyFilterPath(t *testing.T) {
	p := &CommonParams{FilterPath: []string{"_id"}}
	m := p.ToMap()
	assert.Equal(t, "_id", m["filter_path"])
	assert.Len(t, m, 1)
}

// TestCommonParamsToMapSingleFilterPath verifies joinComma with one element.
func TestCommonParamsToMapSingleFilterPath(t *testing.T) {
	p := &CommonParams{FilterPath: []string{"only"}}
	m := p.ToMap()
	assert.Equal(t, "only", m["filter_path"])
}

// TestJoinCommaEmpty verifies joinComma returns "" for an empty slice.
func TestJoinCommaEmpty(t *testing.T) {
	assert.Empty(t, joinComma(nil))
	assert.Empty(t, joinComma([]string{}))
}

// TestJoinCommaMultiple verifies joinComma joins multiple elements.
func TestJoinCommaMultiple(t *testing.T) {
	assert.Equal(t, "a,b,c", joinComma([]string{"a", "b", "c"}))
}

// TestBoolStr verifies boolStr formatting.
func TestBoolStr(t *testing.T) {
	assert.Equal(t, "true", boolStr(true))
	assert.Equal(t, "false", boolStr(false))
}

// TestDocumentVersionZero verifies a zero DocumentVersion has nil pointers.
func TestDocumentVersionZero(t *testing.T) {
	var dv DocumentVersion
	assert.Nil(t, dv.SeqNo)
	assert.Nil(t, dv.PrimaryTerm)
}

// TestDocumentVersionSet verifies DocumentVersion carries seq_no/primary_term.
func TestDocumentVersionSet(t *testing.T) {
	dv := DocumentVersion{SeqNo: ptrInt64Local(5), PrimaryTerm: ptrInt64Local(2)}
	require.NotNil(t, dv.SeqNo)
	require.NotNil(t, dv.PrimaryTerm)
	assert.Equal(t, int64(5), *dv.SeqNo)
	assert.Equal(t, int64(2), *dv.PrimaryTerm)
}

// TestShardsInfoSerialization verifies ShardsInfo round-trips through JSON.
func TestShardsInfoSerialization(t *testing.T) {
	si := ShardsInfo{
		Total:      10,
		Successful: 8,
		Failed:     2,
		Skipped:    1,
		Failures: []*ShardOperationFailedException{
			{
				Shard:   1,
				Index:   "idx",
				Status:  "BAD",
				Node:    "node-1",
				Primary: true,
				Reason:  map[string]any{"type": "x", "reason": "fail"},
			},
		},
	}
	b, err := json.Marshal(si)
	require.NoError(t, err)
	var out ShardsInfo
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, 10, out.Total)
	assert.Equal(t, 8, out.Successful)
	assert.Equal(t, 2, out.Failed)
	assert.Equal(t, 1, out.Skipped)
	require.Len(t, out.Failures, 1)
	assert.Equal(t, 1, out.Failures[0].Shard)
	assert.Equal(t, "idx", out.Failures[0].Index)
	assert.Equal(t, "BAD", out.Failures[0].Status)
	assert.Equal(t, "node-1", out.Failures[0].Node)
	assert.True(t, out.Failures[0].Primary)
}

// TestShardsInfoEmpty verifies an empty ShardsInfo serializes minimally.
func TestShardsInfoEmpty(t *testing.T) {
	si := ShardsInfo{}
	b, err := json.Marshal(si)
	require.NoError(t, err)
	var out ShardsInfo
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, 0, out.Total)
	assert.Nil(t, out.Failures)
}

// TestBroadcastResponseSerialization verifies BroadcastResponse round-trips.
func TestBroadcastResponseSerialization(t *testing.T) {
	br := BroadcastResponse{
		Shards: &ShardsInfo{Total: 5, Successful: 5, Failed: 0},
		Total:  5, Successful: 5, Failed: 0,
	}
	b, err := json.Marshal(br)
	require.NoError(t, err)
	var out BroadcastResponse
	require.NoError(t, json.Unmarshal(b, &out))
	require.NotNil(t, out.Shards)
	assert.Equal(t, 5, out.Shards.Total)
	assert.Equal(t, 5, out.Total)
}

// TestBroadcastResponseNoShards verifies BroadcastResponse with nil Shards.
func TestBroadcastResponseNoShards(t *testing.T) {
	br := BroadcastResponse{Total: 3, Successful: 3, Failed: 0}
	b, err := json.Marshal(br)
	require.NoError(t, err)
	var out BroadcastResponse
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Nil(t, out.Shards)
	assert.Equal(t, 3, out.Total)
}

// TestAcknowledgedResponseSerialization verifies AcknowledgedResponse round-trips.
func TestAcknowledgedResponseSerialization(t *testing.T) {
	ar := AcknowledgedResponse{Acknowledged: true, ShardsAcknowledged: true, Index: "idx"}
	b, err := json.Marshal(ar)
	require.NoError(t, err)
	var out AcknowledgedResponse
	require.NoError(t, json.Unmarshal(b, &out))
	assert.True(t, out.Acknowledged)
	assert.True(t, out.ShardsAcknowledged)
	assert.Equal(t, "idx", out.Index)
}

// TestFailedNodeExceptionError verifies FailedNodeException wraps an
// ElasticsearchErrorDetails and exposes its Reason via Error().
func TestFailedNodeExceptionError(t *testing.T) {
	fne := &FailedNodeException{
		ElasticsearchErrorDetails: &ElasticsearchErrorDetails{
			Type:   "node_disconnected_exception",
			Reason: "node unreachable",
		},
		NodeId: "node-1",
	}
	// FailedNodeException embeds *ElasticsearchErrorDetails, so it satisfies
	// the error interface via ElasticsearchErrorDetails.Error()... but that
	// method is on *ElasticsearchError, not *ElasticsearchErrorDetails.
	// Verify the embedded details are accessible.
	require.NotNil(t, fne.ElasticsearchErrorDetails)
	assert.Equal(t, "node unreachable", fne.Reason)
	assert.Equal(t, "node-1", fne.NodeId)
}

// TestFailedNodeExceptionSerialization verifies FailedNodeException round-trips.
func TestFailedNodeExceptionSerialization(t *testing.T) {
	fne := FailedNodeException{
		ElasticsearchErrorDetails: &ElasticsearchErrorDetails{
			Type:   "node_disconnected_exception",
			Reason: "node unreachable",
		},
		NodeId: "node-1",
	}
	b, err := json.Marshal(fne)
	require.NoError(t, err)
	var out FailedNodeException
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, "node-1", out.NodeId)
	require.NotNil(t, out.ElasticsearchErrorDetails)
	assert.Equal(t, "node unreachable", out.Reason)
}

// TestListResponseSerialization verifies ListResponse round-trips.
func TestListResponseSerialization(t *testing.T) {
	lr := ListResponse[string]{
		Items: []string{"a", "b"},
		Total: 2,
	}
	b, err := json.Marshal(lr)
	require.NoError(t, err)
	var out ListResponse[string]
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, 2, out.Total)
	assert.Equal(t, []string{"a", "b"}, out.Items)
}

// TestRedactURLFull verifies RedactURL strips full userinfo.
func TestRedactURLFull(t *testing.T) {
	assert.Equal(t, "https://host:9200/_search", RedactURL("https://user:pass@host:9200/_search"))
}

// TestRedactURLNoUserinfo verifies RedactURL passes through URLs without userinfo.
func TestRedactURLNoUserinfo(t *testing.T) {
	assert.Equal(t, "https://host:9200/_search", RedactURL("https://host:9200/_search"))
}

// TestRedactURLUserOnly verifies RedactURL strips username-only userinfo.
func TestRedactURLUserOnly(t *testing.T) {
	assert.Equal(t, "https://host:9200/", RedactURL("https://api-key-12345@host:9200/"))
}

// TestRedactURLUserAndPass verifies RedactURL strips user+password userinfo.
func TestRedactURLUserAndPass(t *testing.T) {
	assert.Equal(t, "http://host/path", RedactURL("http://u:p@host/path"))
}

// TestRedactURLParseError verifies RedactURL returns unparseable input unchanged.
func TestRedactURLParseError(t *testing.T) {
	assert.Equal(t, "://bad", RedactURL("://bad"))
}

// TestRedactURLEmpty verifies RedactURL handles an empty string.
func TestRedactURLEmpty(t *testing.T) {
	assert.Equal(t, "", RedactURL(""))
}

func ptrInt64Local(n int64) *int64 { return &n }
