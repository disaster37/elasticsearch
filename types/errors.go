package types

import (
	"fmt"

	json "github.com/goccy/go-json"
)

// ElasticsearchError is the structured error returned by Elasticsearch when a
// request fails. Status is the HTTP status code (e.g. 400, 404, 409, 500) and
// Details holds the typed payload.
//
// Use [IsNotFound], [IsConflict] and [IsUnauthorized] helpers for common cases,
// or type-assert:
//
//	var esErr *types.ElasticsearchError
//	if errors.As(err, &esErr) {
//	    log.Printf("Elasticsearch %d: %s", esErr.Status, esErr.Details.Reason)
//	}
type ElasticsearchError struct {
	Status  int                        `json:"status"`
	Details *ElasticsearchErrorDetails `json:"error,omitempty"`
}

// Error returns a human-readable string combining the HTTP status code,
// error reason, and error type from the Elasticsearch response.
func (e *ElasticsearchError) Error() string {
	if e.Details == nil || e.Details.Reason == "" {
		return fmt.Sprintf("elasticsearch error %d", e.Status)
	}
	msg := fmt.Sprintf("elasticsearch error %d: %s [type=%s]", e.Status, e.Details.Reason, e.Details.Type)
	if cause := causeString(e.Details); cause != "" {
		msg += "; caused by: " + cause
	}
	return msg
}

// StatusCode returns the HTTP status code reported by Elasticsearch.
func (e *ElasticsearchError) StatusCode() int {
	return e.Status
}

// UnmarshalJSON decodes an Elasticsearch error payload. It tolerates both the
// structured form ({"error": {...}, "status": N}) and the legacy/plain-string
// form ({"error": "some string", "status": N}), mapping the latter to
// Details{Reason: "some string"}.
func (e *ElasticsearchError) UnmarshalJSON(b []byte) error {
	// Decode the status and the "error" field which may be an object or a
	// plain string.
	raw := &struct {
		Status int             `json:"status"`
		Error  json.RawMessage `json:"error"`
	}{}
	if err := json.Unmarshal(b, raw); err != nil {
		return err
	}
	e.Status = raw.Status
	if len(raw.Error) == 0 || string(raw.Error) == "null" {
		return nil
	}
	// Try the structured object form first.
	d := &ElasticsearchErrorDetails{}
	if err := json.Unmarshal(raw.Error, d); err == nil && d.Type != "" {
		e.Details = d
		return nil
	}
	// Fall back to the legacy plain-string form.
	var s string
	if err := json.Unmarshal(raw.Error, &s); err == nil {
		e.Details = &ElasticsearchErrorDetails{Reason: s}
		return nil
	}
	// Last resort: keep the object form even if Type was empty.
	if err := json.Unmarshal(raw.Error, d); err == nil {
		e.Details = d
	}
	return nil
}

// causeString extracts the most informative cause from an
// ElasticsearchErrorDetails, checking caused_by, root_cause, and
// failed_shards in that priority order.
func causeString(d *ElasticsearchErrorDetails) string {
	if len(d.CausedBy) > 0 {
		reason, _ := d.CausedBy["reason"].(string)
		typ, _ := d.CausedBy["type"].(string)
		if reason != "" && reason != d.Reason {
			if typ != "" {
				return fmt.Sprintf("[type=%s] %s", typ, reason)
			}
			return reason
		}
	}

	for _, rc := range d.RootCause {
		if rc == nil || rc.Reason == "" || rc.Reason == d.Reason {
			continue
		}
		if rc.Type != "" {
			return fmt.Sprintf("[type=%s] %s", rc.Type, rc.Reason)
		}
		return rc.Reason
	}

	for _, shard := range d.FailedShards {
		if shard == nil {
			continue
		}
		switch rv := shard["reason"].(type) {
		case map[string]any:
			reason, _ := rv["reason"].(string)
			typ, _ := rv["type"].(string)
			if reason != "" {
				if typ != "" {
					return fmt.Sprintf("[type=%s] %s", typ, reason)
				}
				return reason
			}
		case string:
			if rv != "" {
				return rv
			}
		}
	}

	return ""
}

// ElasticsearchErrorDetails is the structured payload inside an
// [ElasticsearchError].
type ElasticsearchErrorDetails struct {
	Type         string                       `json:"type"`
	Reason       string                       `json:"reason"`
	ResourceType string                       `json:"resource.type,omitempty"`
	ResourceId   string                       `json:"resource.id,omitempty"`
	Index        string                       `json:"index,omitempty"`
	Phase        string                       `json:"phase,omitempty"`
	Grouped      bool                         `json:"grouped,omitempty"`
	CausedBy     map[string]any               `json:"caused_by,omitempty"`
	RootCause    []*ElasticsearchErrorDetails `json:"root_cause,omitempty"`
	Suppressed   []*ElasticsearchErrorDetails `json:"suppressed,omitempty"`
	FailedShards []map[string]any             `json:"failed_shards,omitempty"`
	Header       map[string]any               `json:"header,omitempty"`
	ScriptStack  []string                     `json:"script_stack,omitempty"`
	Script       string                       `json:"script,omitempty"`
	Lang         string                       `json:"lang,omitempty"`
	Position     *ScriptErrorPosition         `json:"position,omitempty"`
}

// ScriptErrorPosition describes where in a script an error occurred.
type ScriptErrorPosition struct {
	Offset int `json:"offset"`
	Start  int `json:"start"`
	End    int `json:"end"`
}

// IsNotFound reports whether err is an [*ElasticsearchError] with HTTP status 404.
func IsNotFound(err error) bool {
	if e, ok := err.(*ElasticsearchError); ok {
		return e.Status == 404
	}
	return false
}

// IsConflict reports whether err is an [*ElasticsearchError] with HTTP status 409.
func IsConflict(err error) bool {
	if e, ok := err.(*ElasticsearchError); ok {
		return e.Status == 409
	}
	return false
}

// IsUnauthorized reports whether err is an [*ElasticsearchError] with HTTP status 401.
func IsUnauthorized(err error) bool {
	if e, ok := err.(*ElasticsearchError); ok {
		return e.Status == 401
	}
	return false
}
