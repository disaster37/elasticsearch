package api

import "github.com/disaster37/elasticsearch/v9/types"

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

// SecurityRefreshSpecParams is the shared opt-in query carrier for security
// endpoints whose only endpoint-local parameter is refresh.
type SecurityRefreshSpecParams struct {
	Common  *CommonParams
	Refresh Refresh
}

// ToMap converts SecurityRefreshSpecParams to a query-parameter map.
func (p *SecurityRefreshSpecParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Refresh != "" {
		m["refresh"] = string(p.Refresh)
	}
	return mergeParams(m, p.Common)
}

// SecurityGetUserSpecParams are the opt-in query parameters for security.get_user.
type SecurityGetUserSpecParams struct {
	Common         *CommonParams
	WithProfileUid *bool
}

// ToMap converts SecurityGetUserSpecParams to a query-parameter map.
func (p *SecurityGetUserSpecParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "with_profile_uid", p.WithProfileUid)
	return mergeParams(m, p.Common)
}

// SecurityChangePasswordRequest is the opt-in request for
// security.change_password. Username may be empty to target the authenticated
// user.
type SecurityChangePasswordRequest struct {
	Username string
	Body     any
	Params   *SecurityRefreshSpecParams
}

// SecurityDeletePrivilegesRequest is the opt-in request for
// security.delete_privileges.
type SecurityDeletePrivilegesRequest struct {
	Application string
	Name        string
	Params      *SecurityRefreshSpecParams
}

// SecurityPutRoleRequest is the opt-in request for security.put_role.
type SecurityPutRoleRequest struct {
	Name   string
	Body   any
	Params *SecurityRefreshSpecParams
}

// SecurityPutRoleMappingRequest is the opt-in request for
// security.put_role_mapping.
type SecurityPutRoleMappingRequest struct {
	Name   string
	Body   any
	Params *SecurityRefreshSpecParams
}

// SecurityPutUserRequest is the opt-in request for security.put_user.
type SecurityPutUserRequest struct {
	Username string
	Body     any
	Params   *SecurityRefreshSpecParams
}

// IndicesForceMergeRequest is the opt-in request for indices.forcemerge. When
// Params.WaitForCompletion is false the response carries a task id.
type IndicesForceMergeRequest struct {
	Indices []string
	Params  *IndicesForceMergeParams
}

// IndicesForceMergeSpecResponse retains both the synchronous shard results and
// the asynchronous task id from the forcemerge response.
type IndicesForceMergeSpecResponse struct {
	types.BroadcastResponse
	Task string `json:"task,omitempty"`
}

// MtermvectorsSpecParams are the opt-in query parameters for mtermvectors. It
// embeds the termvectors carrier by value and adds the ids list so that ids
// never leak into the termvectors endpoint.
type MtermvectorsSpecParams struct {
	TermVectorsParams
	Ids []string
}

// ToMap converts MtermvectorsSpecParams to a query-parameter map.
func (p *MtermvectorsSpecParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := p.TermVectorsParams.ToMap()
	setListParam(m, "ids", p.Ids)
	return m
}

// MtermvectorsSpecRequest is the opt-in request for mtermvectors.
type MtermvectorsSpecRequest struct {
	Index  string
	Body   any
	Params *MtermvectorsSpecParams
}
