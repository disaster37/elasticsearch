package api

import (
	"strconv"
	"strings"

	"github.com/disaster37/elasticsearch/v9/types"
)

// CommonParams is an alias for [types.CommonParams], embedded by every
// endpoint's Params struct so the common query parameters (pretty, human,
// error_trace, filter_path) are available everywhere.
type CommonParams = types.CommonParams

// mergeParams merges a service-specific query-parameter map with the common
// parameters (pretty, human, error_trace, filter_path). The common parameters
// are only added when not already present in m (service params take
// precedence). Returns m (possibly nil) for convenience.
func mergeParams(m map[string]string, common *types.CommonParams) map[string]string {
	if common == nil {
		return m
	}
	cm := common.ToMap()
	if len(cm) == 0 {
		return m
	}
	if m == nil {
		m = make(map[string]string, len(cm))
	}
	for k, v := range cm {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}
	return m
}

// setBoolParam adds a boolean query parameter when b is non-nil.
func setBoolParam(m map[string]string, key string, b *bool) {
	if b != nil {
		m[key] = strconv.FormatBool(*b)
	}
}

// setListParam adds a comma-joined list query parameter when the slice is
// non-empty.
func setListParam(m map[string]string, key string, vals []string) {
	if len(vals) > 0 {
		m[key] = strings.Join(vals, ",")
	}
}
