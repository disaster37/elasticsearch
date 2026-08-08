package types

import (
	"errors"
	"testing"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElasticsearchErrorStructured(t *testing.T) {
	body := `{"error":{"type":"illegal_argument_exception","reason":"bad request","caused_by":{"type":"some","reason":"root"}},"status":400}`
	var e ElasticsearchError
	err := json.Unmarshal([]byte(body), &e)
	assert.NoError(t, err)
	assert.Equal(t, 400, e.Status)
	assert.Equal(t, "illegal_argument_exception", e.Details.Type)
	assert.Equal(t, "bad request", e.Details.Reason)
	assert.Contains(t, e.Error(), "elasticsearch error 400")
	assert.Contains(t, e.Error(), "caused by")
}

func TestElasticsearchErrorLegacyString(t *testing.T) {
	body := `{"error":"some string","status":400}`
	var e ElasticsearchError
	err := json.Unmarshal([]byte(body), &e)
	assert.NoError(t, err)
	assert.Equal(t, 400, e.Status)
	assert.Equal(t, "some string", e.Details.Reason)
}

func TestElasticsearchErrorNoDetails(t *testing.T) {
	e := &ElasticsearchError{Status: 500}
	assert.Contains(t, e.Error(), "elasticsearch error 500")
}

// TestElasticsearchErrorEmptyReason verifies Error() falls back to the bare
// status form when Details is present but Reason is empty.
func TestElasticsearchErrorEmptyReason(t *testing.T) {
	e := &ElasticsearchError{Status: 500, Details: &ElasticsearchErrorDetails{Type: "x"}}
	assert.Equal(t, "elasticsearch error 500", e.Error())
}

// TestElasticsearchErrorNilDetails verifies Error() handles nil Details.
func TestElasticsearchErrorNilDetails(t *testing.T) {
	e := &ElasticsearchError{Status: 404}
	assert.Equal(t, "elasticsearch error 404", e.Error())
}

// TestElasticsearchErrorWithDetailsNoCause verifies Error() formatting when
// Details has a reason but no cause.
func TestElasticsearchErrorWithDetailsNoCause(t *testing.T) {
	e := &ElasticsearchError{
		Status: 400,
		Details: &ElasticsearchErrorDetails{
			Type:   "illegal_argument_exception",
			Reason: "bad request",
		},
	}
	assert.Equal(t, "elasticsearch error 400: bad request [type=illegal_argument_exception]", e.Error())
}

// TestElasticsearchErrorStatusCode verifies StatusCode returns the status.
func TestElasticsearchErrorStatusCode(t *testing.T) {
	e := &ElasticsearchError{Status: 418}
	assert.Equal(t, 418, e.StatusCode())
}

// TestElasticsearchErrorUnmarshalEmptyBody verifies UnmarshalJSON handles an
// empty/null error field.
func TestElasticsearchErrorUnmarshalEmptyBody(t *testing.T) {
	body := `{"status":404}`
	var e ElasticsearchError
	err := json.Unmarshal([]byte(body), &e)
	assert.NoError(t, err)
	assert.Equal(t, 404, e.Status)
	assert.Nil(t, e.Details)
}

// TestElasticsearchErrorUnmarshalNullError verifies UnmarshalJSON handles an
// explicit null error field.
func TestElasticsearchErrorUnmarshalNullError(t *testing.T) {
	body := `{"error":null,"status":500}`
	var e ElasticsearchError
	err := json.Unmarshal([]byte(body), &e)
	assert.NoError(t, err)
	assert.Equal(t, 500, e.Status)
	assert.Nil(t, e.Details)
}

// TestElasticsearchErrorUnmarshalObjectNoType verifies UnmarshalJSON falls back
// to keeping the object form when Type is empty (last-resort branch).
func TestElasticsearchErrorUnmarshalObjectNoType(t *testing.T) {
	body := `{"error":{"reason":"no type here"},"status":400}`
	var e ElasticsearchError
	err := json.Unmarshal([]byte(body), &e)
	assert.NoError(t, err)
	assert.Equal(t, 400, e.Status)
	require.NotNil(t, e.Details)
	assert.Equal(t, "no type here", e.Details.Reason)
}

// TestElasticsearchErrorUnmarshalMalformedJSON verifies UnmarshalJSON returns
// an error for malformed JSON.
func TestElasticsearchErrorUnmarshalMalformedJSON(t *testing.T) {
	body := `{"error":{not valid`
	var e ElasticsearchError
	err := json.Unmarshal([]byte(body), &e)
	assert.Error(t, err)
}

// TestElasticsearchErrorUnmarshalNonStringNonObject verifies UnmarshalJSON
// tolerates an error field that is neither object nor string (e.g. number).
func TestElasticsearchErrorUnmarshalNonStringNonObject(t *testing.T) {
	body := `{"error":42,"status":500}`
	var e ElasticsearchError
	err := json.Unmarshal([]byte(body), &e)
	assert.NoError(t, err)
	assert.Equal(t, 500, e.Status)
}

// TestElasticsearchErrorComplexNested verifies UnmarshalJSON decodes a complex
// nested error with root_cause, caused_by, suppressed, failed_shards, and
// script_stack.
func TestElasticsearchErrorComplexNested(t *testing.T) {
	body := `{
		"error":{
			"type":"search_phase_execution_exception",
			"reason":"all shards failed",
			"phase":"query",
			"grouped":true,
			"caused_by":{"type":"query_shard_exception","reason":"failed to parse"},
			"root_cause":[
				{"type":"query_shard_exception","reason":"shard 0 failed","index":"idx"},
				{"type":"parse_exception","reason":"bad query","index":"idx"}
			],
			"suppressed":[
				{"type":"illegal_argument_exception","reason":"suppressed one"}
			],
			"failed_shards":[
				{"shard":0,"index":"idx","node":"n1","reason":{"type":"x","reason":"shard reason"}}
			],
			"script_stack":["ctx._source","return;"],
			"script":"return;",
			"lang":"painless",
			"position":{"offset":1,"start":0,"end":2}
		},
		"status":400
	}`
	var e ElasticsearchError
	err := json.Unmarshal([]byte(body), &e)
	require.NoError(t, err)
	assert.Equal(t, 400, e.Status)
	require.NotNil(t, e.Details)
	assert.Equal(t, "search_phase_execution_exception", e.Details.Type)
	assert.Equal(t, "all shards failed", e.Details.Reason)
	assert.Equal(t, "query", e.Details.Phase)
	assert.True(t, e.Details.Grouped)
	require.Len(t, e.Details.RootCause, 2)
	assert.Equal(t, "shard 0 failed", e.Details.RootCause[0].Reason)
	assert.Equal(t, "idx", e.Details.RootCause[0].Index)
	require.Len(t, e.Details.Suppressed, 1)
	assert.Equal(t, "suppressed one", e.Details.Suppressed[0].Reason)
	require.Len(t, e.Details.FailedShards, 1)
	assert.Equal(t, "shard reason", e.Details.FailedShards[0]["reason"].(map[string]any)["reason"])
	assert.Equal(t, []string{"ctx._source", "return;"}, e.Details.ScriptStack)
	assert.Equal(t, "return;", e.Details.Script)
	assert.Equal(t, "painless", e.Details.Lang)
	require.NotNil(t, e.Details.Position)
	assert.Equal(t, 1, e.Details.Position.Offset)
	assert.Equal(t, 0, e.Details.Position.Start)
	assert.Equal(t, 2, e.Details.Position.End)
}

// TestCauseStringCausedByWithType verifies causeString formats caused_by with
// a type.
func TestCauseStringCausedByWithType(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason:   "outer",
		CausedBy: map[string]any{"type": "inner_type", "reason": "inner reason"},
	}
	assert.Equal(t, "[type=inner_type] inner reason", causeString(d))
}

// TestCauseStringCausedByNoType verifies causeString formats caused_by without
// a type.
func TestCauseStringCausedByNoType(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason:   "outer",
		CausedBy: map[string]any{"reason": "inner reason"},
	}
	assert.Equal(t, "inner reason", causeString(d))
}

// TestCauseStringCausedBySameReason verifies causeString skips caused_by when
// the reason matches the parent.
func TestCauseStringCausedBySameReason(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason:   "same",
		CausedBy: map[string]any{"type": "t", "reason": "same"},
	}
	// Falls through to root_cause (none) → empty.
	assert.Empty(t, causeString(d))
}

// TestCauseStringRootCause verifies causeString formats a root_cause entry.
func TestCauseStringRootCause(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason: "outer",
		RootCause: []*ElasticsearchErrorDetails{
			{Type: "rc_type", Reason: "rc reason"},
		},
	}
	assert.Equal(t, "[type=rc_type] rc reason", causeString(d))
}

// TestCauseStringRootCauseNoType verifies causeString formats a root_cause
// entry without a type.
func TestCauseStringRootCauseNoType(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason: "outer",
		RootCause: []*ElasticsearchErrorDetails{
			{Reason: "rc reason"},
		},
	}
	assert.Equal(t, "rc reason", causeString(d))
}

// TestCauseStringRootCauseSameReason verifies causeString skips root_cause
// entries whose reason matches the parent.
func TestCauseStringRootCauseSameReason(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason: "same",
		RootCause: []*ElasticsearchErrorDetails{
			{Type: "t", Reason: "same"},
		},
	}
	assert.Empty(t, causeString(d))
}

// TestCauseStringRootCauseNil verifies causeString skips nil root_cause entries.
func TestCauseStringRootCauseNil(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason: "outer",
		RootCause: []*ElasticsearchErrorDetails{
			nil,
			{Reason: ""},
		},
	}
	assert.Empty(t, causeString(d))
}

// TestCauseStringFailedShardMap verifies causeString formats a failed_shards
// reason that is a map.
func TestCauseStringFailedShardMap(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason: "outer",
		FailedShards: []map[string]any{
			{"reason": map[string]any{"type": "fs_type", "reason": "fs reason"}},
		},
	}
	assert.Equal(t, "[type=fs_type] fs reason", causeString(d))
}

// TestCauseStringFailedShardMapNoType verifies causeString formats a
// failed_shards reason map without a type.
func TestCauseStringFailedShardMapNoType(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason: "outer",
		FailedShards: []map[string]any{
			{"reason": map[string]any{"reason": "fs reason"}},
		},
	}
	assert.Equal(t, "fs reason", causeString(d))
}

// TestCauseStringFailedShardString verifies causeString formats a
// failed_shards reason that is a plain string.
func TestCauseStringFailedShardString(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason: "outer",
		FailedShards: []map[string]any{
			{"reason": "plain string reason"},
		},
	}
	assert.Equal(t, "plain string reason", causeString(d))
}

// TestCauseStringFailedShardEmptyString verifies causeString skips empty
// string reasons.
func TestCauseStringFailedShardEmptyString(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason: "outer",
		FailedShards: []map[string]any{
			{"reason": ""},
		},
	}
	assert.Empty(t, causeString(d))
}

// TestCauseStringFailedShardNil verifies causeString skips nil failed_shards.
func TestCauseStringFailedShardNil(t *testing.T) {
	d := &ElasticsearchErrorDetails{
		Reason: "outer",
		FailedShards: []map[string]any{
			nil,
		},
	}
	assert.Empty(t, causeString(d))
}

// TestCauseStringEmpty verifies causeString returns "" for an empty details.
func TestCauseStringEmpty(t *testing.T) {
	assert.Empty(t, causeString(&ElasticsearchErrorDetails{}))
}

// TestErrorWithCauseChain verifies the full Error() output includes the cause.
func TestErrorWithCauseChain(t *testing.T) {
	e := &ElasticsearchError{
		Status: 400,
		Details: &ElasticsearchErrorDetails{
			Type:   "outer_type",
			Reason: "outer reason",
			CausedBy: map[string]any{
				"type":   "inner_type",
				"reason": "inner reason",
			},
		},
	}
	msg := e.Error()
	assert.Contains(t, msg, "elasticsearch error 400: outer reason [type=outer_type]")
	assert.Contains(t, msg, "caused by: [type=inner_type] inner reason")
}

func TestIsHelpers(t *testing.T) {
	assert.True(t, IsNotFound(&ElasticsearchError{Status: 404}))
	assert.True(t, IsConflict(&ElasticsearchError{Status: 409}))
	assert.True(t, IsUnauthorized(&ElasticsearchError{Status: 401}))
	assert.False(t, IsNotFound(nil))
}

// TestIsHelpersNonElasticsearchError verifies the helpers return false for
// non-ElasticsearchError errors.
func TestIsHelpersNonElasticsearchError(t *testing.T) {
	assert.False(t, IsNotFound(errors.New("not found")))
	assert.False(t, IsConflict(errors.New("conflict")))
	assert.False(t, IsUnauthorized(errors.New("unauthorized")))
}

// TestIsHelpersWrongStatus verifies the helpers return false for the wrong
// status.
func TestIsHelpersWrongStatus(t *testing.T) {
	assert.False(t, IsNotFound(&ElasticsearchError{Status: 200}))
	assert.False(t, IsConflict(&ElasticsearchError{Status: 404}))
	assert.False(t, IsUnauthorized(&ElasticsearchError{Status: 403}))
}

func TestCommonParamsToMap(t *testing.T) {
	p := &CommonParams{Pretty: ptrBool(true), FilterPath: []string{"hits", "_shards"}}
	m := p.ToMap()
	assert.Equal(t, "true", m["pretty"])
	assert.Equal(t, "hits,_shards", m["filter_path"])
}

func TestCommonParamsNil(t *testing.T) {
	var p *CommonParams
	assert.Nil(t, p.ToMap())
}

func ptrBool(b bool) *bool { return &b }

func TestRedactURL(t *testing.T) {
	assert.Equal(t, "https://host:9200/_search", RedactURL("https://user:pass@host:9200/_search"))
	// Username-only userinfo (often an API key) must also be stripped.
	assert.Equal(t, "https://host:9200/", RedactURL("https://api-key-12345@host:9200/"))
	assert.Equal(t, "https://host:9200/_search", RedactURL("https://host:9200/_search"))
	// Unparseable input is returned unchanged.
	assert.Equal(t, "://bad", RedactURL("://bad"))
}
