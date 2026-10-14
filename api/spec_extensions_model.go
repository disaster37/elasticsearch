package api

// AsyncSearchStatusParams are the opt-in query parameters for
// async_search.status. The original Status method has no carrier, so callers
// that need the keep_alive parameter use
// [AsyncSearchServiceSpecExtensions.StatusWithParams].
type AsyncSearchStatusParams struct {
	Common    *CommonParams
	KeepAlive string
}

// ToMap converts AsyncSearchStatusParams to a query-parameter map.
func (p *AsyncSearchStatusParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.KeepAlive != "" {
		m["keep_alive"] = p.KeepAlive
	}
	return mergeParams(m, p.Common)
}
