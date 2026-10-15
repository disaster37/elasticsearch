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

// MasterTimeoutSpecParams is the shared opt-in query carrier for endpoints
// whose only endpoint-local parameter is master_timeout.
type MasterTimeoutSpecParams struct {
	Common        *CommonParams
	MasterTimeout string
}

// ToMap converts MasterTimeoutSpecParams to a query-parameter map.
func (p *MasterTimeoutSpecParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	return mergeParams(m, p.Common)
}

// TimeoutSpecParams is the shared opt-in query carrier for endpoints whose
// only endpoint-local parameter is timeout.
type TimeoutSpecParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts TimeoutSpecParams to a query-parameter map.
func (p *TimeoutSpecParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// MasterTimeoutAndTimeoutSpecParams is the shared opt-in query carrier for
// endpoints that accept both master_timeout and timeout.
type MasterTimeoutAndTimeoutSpecParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
}

// ToMap converts MasterTimeoutAndTimeoutSpecParams to a query-parameter map.
func (p *MasterTimeoutAndTimeoutSpecParams) ToMap() map[string]string {
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

// CcrPutAutoFollowPatternRequest is the opt-in request for
// ccr.put_auto_follow_pattern. Params may be nil.
type CcrPutAutoFollowPatternRequest struct {
	Name   string
	Body   any
	Params *MasterTimeoutSpecParams
}

// CcrResumeFollowRequest is the opt-in request for ccr.resume_follow. Params
// may be nil.
type CcrResumeFollowRequest struct {
	Index  string
	Body   any
	Params *MasterTimeoutSpecParams
}

// CcrForgetFollowerRequest is the opt-in request for ccr.forget_follower.
// Params may be nil.
type CcrForgetFollowerRequest struct {
	Index  string
	Body   any
	Params *TimeoutSpecParams
}

// EnrichPutPolicyRequest is the opt-in request for enrich.put_policy. Params
// may be nil.
type EnrichPutPolicyRequest struct {
	Name   string
	Body   any
	Params *MasterTimeoutSpecParams
}

// IlmPutLifecycleRequest is the opt-in request for ilm.put_lifecycle. Params
// may be nil.
type IlmPutLifecycleRequest struct {
	Policy string
	Body   any
	Params *MasterTimeoutAndTimeoutSpecParams
}

// SlmPutLifecycleRequest is the opt-in request for slm.put_lifecycle. Params
// may be nil.
type SlmPutLifecycleRequest struct {
	PolicyID string
	Body     any
	Params   *MasterTimeoutAndTimeoutSpecParams
}
