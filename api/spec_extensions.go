package api

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/querydsl"
	"github.com/disaster37/elasticsearch/v9/types"
)

// This file adds opt-in extension APIs for spec endpoints whose original
// service method has no query carrier or no path selector. Existing service
// interfaces and their method signatures are deliberately untouched:
// extensions live on separate <Group>ServiceSpecExtensions interfaces that the
// default implementations satisfy, and callers obtain them through the
// As<Group>ServiceSpecExtensions adapters.

// AsyncSearchServiceSpecExtensions exposes the opt-in async_search.status
// carrier. Obtain it with [AsAsyncSearchServiceSpecExtensions].
type AsyncSearchServiceSpecExtensions interface {
	StatusWithParams(ctx context.Context, id string, params *AsyncSearchStatusParams) (*AsyncSearchStatusResponse, error)
}

// Compile-time assertion that the default implementation supports the
// extension interface.
var _ AsyncSearchServiceSpecExtensions = (*DefaultAsyncSearchService)(nil)

// AsAsyncSearchServiceSpecExtensions adapts an AsyncSearchService to its
// extension interface. It rejects nil and typed-nil services and returns a
// descriptive error for custom implementations without support.
func AsAsyncSearchServiceSpecExtensions(service AsyncSearchService) (AsyncSearchServiceSpecExtensions, error) {
	if isNilService(service) {
		return nil, fmt.Errorf("as async search spec extensions: nil service")
	}
	ext, ok := service.(AsyncSearchServiceSpecExtensions)
	if !ok {
		return nil, fmt.Errorf("as async search spec extensions: %T does not implement AsyncSearchServiceSpecExtensions", service)
	}
	if isNilService(ext) {
		return nil, fmt.Errorf("as async search spec extensions: typed-nil service")
	}
	return ext, nil
}

// StatusWithParams performs GET /_async_search/status/{id} with the optional
// keep_alive query parameter.
func (s *DefaultAsyncSearchService) StatusWithParams(ctx context.Context, id string, params *AsyncSearchStatusParams) (*AsyncSearchStatusResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result AsyncSearchStatusResponse
	err := doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_async_search/status/%s", pathSegment(id)), nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// isNilService reports whether a service interface holds no usable value,
// including a typed-nil pointer.
func isNilService(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return rv.IsNil()
	default:
		return false
	}
}

// asSpecExtensions adapts a service to an extension interface, rejecting nil
// and typed-nil services and returning a descriptive error for custom
// implementations without support.
func asSpecExtensions[T any](group string, service any) (T, error) {
	var zero T
	if isNilService(service) {
		return zero, fmt.Errorf("as %s spec extensions: nil service", group)
	}
	ext, ok := service.(T)
	if !ok {
		return zero, fmt.Errorf("as %s spec extensions: %T does not implement the spec extension interface", group, service)
	}
	if isNilService(ext) {
		return zero, fmt.Errorf("as %s spec extensions: typed-nil service", group)
	}
	return ext, nil
}

// CcrServiceSpecExtensions exposes the opt-in CCR query carriers. Obtain it
// with [AsCcrServiceSpecExtensions].
type CcrServiceSpecExtensions interface {
	PutAutoFollowPatternWithRequest(ctx context.Context, req *CcrPutAutoFollowPatternRequest) (*types.AcknowledgedResponse, error)
	DeleteAutoFollowPatternWithParams(ctx context.Context, name string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	GetAutoFollowPatternWithParams(ctx context.Context, names []string, params *MasterTimeoutSpecParams) (*CcrGetAutoFollowPatternResponse, error)
	PauseAutoFollowPatternWithParams(ctx context.Context, name string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	ResumeAutoFollowPatternWithParams(ctx context.Context, name string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	PauseFollowWithParams(ctx context.Context, index string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	ResumeFollowWithRequest(ctx context.Context, req *CcrResumeFollowRequest) (*types.AcknowledgedResponse, error)
	UnfollowWithParams(ctx context.Context, index string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	ForgetFollowerWithRequest(ctx context.Context, req *CcrForgetFollowerRequest) (*CcrForgetFollowerResponse, error)
	FollowInfoWithParams(ctx context.Context, indices []string, params *MasterTimeoutSpecParams) (*CcrFollowInfoResponse, error)
	FollowStatsWithParams(ctx context.Context, indices []string, params *TimeoutSpecParams) (*CcrFollowStatsResponse, error)
	StatsWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*CcrStatsResponse, error)
}

// Compile-time assertion that the default implementation supports the
// extension interface.
var _ CcrServiceSpecExtensions = (*DefaultCcrService)(nil)

// AsCcrServiceSpecExtensions adapts a CcrService to its extension interface.
func AsCcrServiceSpecExtensions(service CcrService) (CcrServiceSpecExtensions, error) {
	return asSpecExtensions[CcrServiceSpecExtensions]("ccr", service)
}

func (s *DefaultCcrService) PutAutoFollowPatternWithRequest(ctx context.Context, req *CcrPutAutoFollowPatternRequest) (*types.AcknowledgedResponse, error) {
	if req == nil || req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_ccr/auto_follow/%s", pathSegment(req.Name)), req.Body, req.Params, &result)
}

func (s *DefaultCcrService) DeleteAutoFollowPatternWithParams(ctx context.Context, name string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_ccr/auto_follow/%s", pathSegment(name)), nil, params, &result)
}

func (s *DefaultCcrService) GetAutoFollowPatternWithParams(ctx context.Context, names []string, params *MasterTimeoutSpecParams) (*CcrGetAutoFollowPatternResponse, error) {
	path := "/_ccr/auto_follow"
	if len(names) > 0 {
		path = fmt.Sprintf("/_ccr/auto_follow/%s", escapedList(names))
	}
	var result CcrGetAutoFollowPatternResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultCcrService) PauseAutoFollowPatternWithParams(ctx context.Context, name string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ccr/auto_follow/%s/pause", pathSegment(name)), nil, params, &result)
}

func (s *DefaultCcrService) ResumeAutoFollowPatternWithParams(ctx context.Context, name string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ccr/auto_follow/%s/resume", pathSegment(name)), nil, params, &result)
}

func (s *DefaultCcrService) PauseFollowWithParams(ctx context.Context, index string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ccr/pause_follow", pathSegment(index)), nil, params, &result)
}

func (s *DefaultCcrService) ResumeFollowWithRequest(ctx context.Context, req *CcrResumeFollowRequest) (*types.AcknowledgedResponse, error) {
	if req == nil || req.Index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ccr/resume_follow", pathSegment(req.Index)), req.Body, req.Params, &result)
}

func (s *DefaultCcrService) UnfollowWithParams(ctx context.Context, index string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ccr/unfollow", pathSegment(index)), nil, params, &result)
}

func (s *DefaultCcrService) ForgetFollowerWithRequest(ctx context.Context, req *CcrForgetFollowerRequest) (*CcrForgetFollowerResponse, error) {
	if req == nil || req.Index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result CcrForgetFollowerResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ccr/forget_follower", pathSegment(req.Index)), req.Body, req.Params, &result)
}

func (s *DefaultCcrService) FollowInfoWithParams(ctx context.Context, indices []string, params *MasterTimeoutSpecParams) (*CcrFollowInfoResponse, error) {
	path := "/_ccr/info"
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_ccr/info", escapedList(indices))
	}
	var result CcrFollowInfoResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultCcrService) FollowStatsWithParams(ctx context.Context, indices []string, params *TimeoutSpecParams) (*CcrFollowStatsResponse, error) {
	path := "/_ccr/stats"
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_ccr/stats", escapedList(indices))
	}
	var result CcrFollowStatsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultCcrService) StatsWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*CcrStatsResponse, error) {
	var result CcrStatsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_ccr/stats", nil, params, &result)
}

// SlmServiceSpecExtensions exposes the opt-in SLM query carriers. Obtain it
// with [AsSlmServiceSpecExtensions].
type SlmServiceSpecExtensions interface {
	GetLifecycleWithParams(ctx context.Context, policyIds []string, params *MasterTimeoutAndTimeoutSpecParams) (map[string]*SlmPolicy, error)
	PutLifecycleWithRequest(ctx context.Context, req *SlmPutLifecycleRequest) (*types.AcknowledgedResponse, error)
	DeleteLifecycleWithParams(ctx context.Context, policyId string, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	ExecuteLifecycleWithParams(ctx context.Context, policyId string, params *MasterTimeoutAndTimeoutSpecParams) (*SlmExecuteResponse, error)
	ExecuteRetentionWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	GetStatsWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*SlmStatsResponse, error)
	GetStatusWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*SlmStatusResponse, error)
	StartWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	StopWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error)
}

var _ SlmServiceSpecExtensions = (*DefaultSlmService)(nil)

// AsSlmServiceSpecExtensions adapts a SlmService to its extension interface.
func AsSlmServiceSpecExtensions(service SlmService) (SlmServiceSpecExtensions, error) {
	return asSpecExtensions[SlmServiceSpecExtensions]("slm", service)
}

func (s *DefaultSlmService) GetLifecycleWithParams(ctx context.Context, policyIds []string, params *MasterTimeoutAndTimeoutSpecParams) (map[string]*SlmPolicy, error) {
	path := "/_slm/policy"
	if len(policyIds) > 0 {
		path = fmt.Sprintf("/_slm/policy/%s", escapedList(policyIds))
	}
	var result map[string]*SlmPolicy
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultSlmService) PutLifecycleWithRequest(ctx context.Context, req *SlmPutLifecycleRequest) (*types.AcknowledgedResponse, error) {
	if req == nil || req.PolicyID == "" {
		return nil, fmt.Errorf("policy_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_slm/policy/%s", pathSegment(req.PolicyID)), req.Body, req.Params, &result)
}

func (s *DefaultSlmService) DeleteLifecycleWithParams(ctx context.Context, policyId string, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if policyId == "" {
		return nil, fmt.Errorf("policy_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_slm/policy/%s", pathSegment(policyId)), nil, params, &result)
}

func (s *DefaultSlmService) ExecuteLifecycleWithParams(ctx context.Context, policyId string, params *MasterTimeoutAndTimeoutSpecParams) (*SlmExecuteResponse, error) {
	if policyId == "" {
		return nil, fmt.Errorf("policy_id is required")
	}
	var result SlmExecuteResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_slm/policy/%s/_execute", pathSegment(policyId)), nil, params, &result)
}

func (s *DefaultSlmService) ExecuteRetentionWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_slm/_execute_retention", nil, params, &result)
}

func (s *DefaultSlmService) GetStatsWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*SlmStatsResponse, error) {
	var result SlmStatsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_slm/stats", nil, params, &result)
}

func (s *DefaultSlmService) GetStatusWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*SlmStatusResponse, error) {
	var result SlmStatusResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_slm/status", nil, params, &result)
}

func (s *DefaultSlmService) StartWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_slm/start", nil, params, &result)
}

func (s *DefaultSlmService) StopWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_slm/stop", nil, params, &result)
}

// IlmServiceSpecExtensions exposes the opt-in ILM query carriers. Obtain it
// with [AsIlmServiceSpecExtensions].
type IlmServiceSpecExtensions interface {
	GetLifecycleWithParams(ctx context.Context, policies []string, params *MasterTimeoutAndTimeoutSpecParams) (map[string]*IlmPolicy, error)
	PutLifecycleWithRequest(ctx context.Context, req *IlmPutLifecycleRequest) (*types.AcknowledgedResponse, error)
	DeleteLifecycleWithParams(ctx context.Context, policy string, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	StartWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	StopWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	MigrateToDataTiersWithParams(ctx context.Context, dryRun *bool, params *MasterTimeoutSpecParams) (*IlmMigrateToDataTiersResponse, error)
}

var _ IlmServiceSpecExtensions = (*DefaultIlmService)(nil)

// AsIlmServiceSpecExtensions adapts an IlmService to its extension interface.
func AsIlmServiceSpecExtensions(service IlmService) (IlmServiceSpecExtensions, error) {
	return asSpecExtensions[IlmServiceSpecExtensions]("ilm", service)
}

func (s *DefaultIlmService) GetLifecycleWithParams(ctx context.Context, policies []string, params *MasterTimeoutAndTimeoutSpecParams) (map[string]*IlmPolicy, error) {
	path := "/_ilm/policy"
	if len(policies) > 0 {
		path = fmt.Sprintf("/_ilm/policy/%s", escapedList(policies))
	}
	var result map[string]*IlmPolicy
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultIlmService) PutLifecycleWithRequest(ctx context.Context, req *IlmPutLifecycleRequest) (*types.AcknowledgedResponse, error) {
	if req == nil || req.Policy == "" {
		return nil, fmt.Errorf("policy is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_ilm/policy/%s", pathSegment(req.Policy)), req.Body, req.Params, &result)
}

func (s *DefaultIlmService) DeleteLifecycleWithParams(ctx context.Context, policy string, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if policy == "" {
		return nil, fmt.Errorf("policy is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_ilm/policy/%s", pathSegment(policy)), nil, params, &result)
}

func (s *DefaultIlmService) StartWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_ilm/start", nil, params, &result)
}

func (s *DefaultIlmService) StopWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_ilm/stop", nil, params, &result)
}

func (s *DefaultIlmService) MigrateToDataTiersWithParams(ctx context.Context, dryRun *bool, params *MasterTimeoutSpecParams) (*IlmMigrateToDataTiersResponse, error) {
	r, err := buildRequest(s.client, ctx, nil, params)
	if err != nil {
		return nil, err
	}
	if dryRun != nil {
		r.SetQueryParam("dry_run", strconv.FormatBool(*dryRun))
	}
	resp, err := r.Post("/_ilm/migrate_to_data_tiers")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IlmMigrateToDataTiersResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

// EnrichServiceSpecExtensions exposes the opt-in enrich query carriers.
// Obtain it with [AsEnrichServiceSpecExtensions].
type EnrichServiceSpecExtensions interface {
	GetPolicyWithParams(ctx context.Context, names []string, params *MasterTimeoutSpecParams) (json.RawMessage, error)
	PutPolicyWithRequest(ctx context.Context, req *EnrichPutPolicyRequest) (*types.AcknowledgedResponse, error)
	DeletePolicyWithParams(ctx context.Context, name string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	StatsWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (json.RawMessage, error)
}

var _ EnrichServiceSpecExtensions = (*DefaultEnrichService)(nil)

// AsEnrichServiceSpecExtensions adapts an EnrichService to its extension
// interface.
func AsEnrichServiceSpecExtensions(service EnrichService) (EnrichServiceSpecExtensions, error) {
	return asSpecExtensions[EnrichServiceSpecExtensions]("enrich", service)
}

func (s *DefaultEnrichService) GetPolicyWithParams(ctx context.Context, names []string, params *MasterTimeoutSpecParams) (json.RawMessage, error) {
	path := "/_enrich/policy"
	if len(names) > 0 {
		path = fmt.Sprintf("/_enrich/policy/%s", escapedList(names))
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", path, nil, params)
}

func (s *DefaultEnrichService) PutPolicyWithRequest(ctx context.Context, req *EnrichPutPolicyRequest) (*types.AcknowledgedResponse, error) {
	if req == nil || req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_enrich/policy/%s", pathSegment(req.Name)), req.Body, req.Params, &result)
}

func (s *DefaultEnrichService) DeletePolicyWithParams(ctx context.Context, name string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_enrich/policy/%s", pathSegment(name)), nil, params, &result)
}

func (s *DefaultEnrichService) StatsWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "GET", "/_enrich/_stats", nil, params)
}

// FeaturesServiceSpecExtensions exposes the opt-in features query carriers.
// Obtain it with [AsFeaturesServiceSpecExtensions].
type FeaturesServiceSpecExtensions interface {
	GetFeaturesWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (*FeaturesGetResponse, error)
	ResetFeaturesWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error)
}

var _ FeaturesServiceSpecExtensions = (*DefaultFeaturesService)(nil)

// AsFeaturesServiceSpecExtensions adapts a FeaturesService to its extension
// interface.
func AsFeaturesServiceSpecExtensions(service FeaturesService) (FeaturesServiceSpecExtensions, error) {
	return asSpecExtensions[FeaturesServiceSpecExtensions]("features", service)
}

func (s *DefaultFeaturesService) GetFeaturesWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (*FeaturesGetResponse, error) {
	var result FeaturesGetResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_features", nil, params, &result)
}

func (s *DefaultFeaturesService) ResetFeaturesWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_features/_reset", nil, params, &result)
}

// WatcherServiceSpecExtensions exposes the opt-in watcher query carriers.
// Obtain it with [AsWatcherServiceSpecExtensions].
type WatcherServiceSpecExtensions interface {
	StartWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	StopWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error)
}

var _ WatcherServiceSpecExtensions = (*DefaultWatcherService)(nil)

// AsWatcherServiceSpecExtensions adapts a WatcherService to its extension
// interface.
func AsWatcherServiceSpecExtensions(service WatcherService) (WatcherServiceSpecExtensions, error) {
	return asSpecExtensions[WatcherServiceSpecExtensions]("watcher", service)
}

func (s *DefaultWatcherService) StartWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_watcher/_start", nil, params, &result)
}

func (s *DefaultWatcherService) StopWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_watcher/_stop", nil, params, &result)
}

// XpackServiceSpecExtensions exposes the opt-in xpack query carriers. Obtain
// it with [AsXpackServiceSpecExtensions].
type XpackServiceSpecExtensions interface {
	UsageWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (XpackUsageResponse, error)
}

var _ XpackServiceSpecExtensions = (*DefaultXpackService)(nil)

// AsXpackServiceSpecExtensions adapts an XpackService to its extension
// interface.
func AsXpackServiceSpecExtensions(service XpackService) (XpackServiceSpecExtensions, error) {
	return asSpecExtensions[XpackServiceSpecExtensions]("xpack", service)
}

func (s *DefaultXpackService) UsageWithParams(ctx context.Context, params *MasterTimeoutSpecParams) (XpackUsageResponse, error) {
	var result XpackUsageResponse
	return result, doReq(s.logger, ctx, s.client, "GET", "/_xpack/usage", nil, params, &result)
}

// SecurityServiceSpecExtensions exposes the opt-in security query carriers.
// Obtain it with [AsSecurityServiceSpecExtensions].
type SecurityServiceSpecExtensions interface {
	ChangePasswordWithRequest(ctx context.Context, req *SecurityChangePasswordRequest) (*types.AcknowledgedResponse, error)
	CreateAPIKeyWithParams(ctx context.Context, body any, params *SecurityRefreshSpecParams) (*SecurityCreateAPIKeyResponse, error)
	DeletePrivilegesWithRequest(ctx context.Context, req *SecurityDeletePrivilegesRequest) (*types.AcknowledgedResponse, error)
	DeleteRoleWithParams(ctx context.Context, name string, params *SecurityRefreshSpecParams) (*types.AcknowledgedResponse, error)
	DeleteRoleMappingWithParams(ctx context.Context, name string, params *SecurityRefreshSpecParams) (*types.AcknowledgedResponse, error)
	DeleteUserWithParams(ctx context.Context, username string, params *SecurityRefreshSpecParams) (*types.AcknowledgedResponse, error)
	DisableUserWithParams(ctx context.Context, username string, params *SecurityRefreshSpecParams) (*types.AcknowledgedResponse, error)
	EnableUserWithParams(ctx context.Context, username string, params *SecurityRefreshSpecParams) (*types.AcknowledgedResponse, error)
	PutPrivilegesWithParams(ctx context.Context, body any, params *SecurityRefreshSpecParams) (*SecurityPutPrivilegesResponse, error)
	PutRoleWithRequest(ctx context.Context, req *SecurityPutRoleRequest) (*SecurityCreateResponse, error)
	PutRoleMappingWithRequest(ctx context.Context, req *SecurityPutRoleMappingRequest) (*SecurityRoleMappingCreateResponse, error)
	PutUserWithRequest(ctx context.Context, req *SecurityPutUserRequest) (*SecurityCreateResponse, error)
	GetUserWithParams(ctx context.Context, usernames []string, params *SecurityGetUserSpecParams) (map[string]*SecurityUser, error)
	HasPrivilegesWithRequest(ctx context.Context, req *SecurityHasPrivilegesSpecRequest) (*SecurityHasPrivilegesResponse, error)
}

var _ SecurityServiceSpecExtensions = (*DefaultSecurityService)(nil)

// AsSecurityServiceSpecExtensions adapts a SecurityService to its extension
// interface.
func AsSecurityServiceSpecExtensions(service SecurityService) (SecurityServiceSpecExtensions, error) {
	return asSpecExtensions[SecurityServiceSpecExtensions]("security", service)
}

func (s *DefaultSecurityService) ChangePasswordWithRequest(ctx context.Context, req *SecurityChangePasswordRequest) (*types.AcknowledgedResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	path := "/_security/user/_password"
	if req.Username != "" {
		path = fmt.Sprintf("/_security/user/%s/_password", pathSegment(req.Username))
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", path, req.Body, req.Params, &result)
}

func (s *DefaultSecurityService) CreateAPIKeyWithParams(ctx context.Context, body any, params *SecurityRefreshSpecParams) (*SecurityCreateAPIKeyResponse, error) {
	var result SecurityCreateAPIKeyResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_security/api_key", body, params, &result)
}

func (s *DefaultSecurityService) DeletePrivilegesWithRequest(ctx context.Context, req *SecurityDeletePrivilegesRequest) (*types.AcknowledgedResponse, error) {
	if req == nil || req.Application == "" || req.Name == "" {
		return nil, fmt.Errorf("application and name are required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_security/privilege/%s/%s", pathSegment(req.Application), pathSegment(req.Name)), nil, req.Params, &result)
}

func (s *DefaultSecurityService) DeleteRoleWithParams(ctx context.Context, name string, params *SecurityRefreshSpecParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_security/role/%s", pathSegment(name)), nil, params, &result)
}

func (s *DefaultSecurityService) DeleteRoleMappingWithParams(ctx context.Context, name string, params *SecurityRefreshSpecParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_security/role_mapping/%s", pathSegment(name)), nil, params, &result)
}

func (s *DefaultSecurityService) DeleteUserWithParams(ctx context.Context, username string, params *SecurityRefreshSpecParams) (*types.AcknowledgedResponse, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_security/user/%s", pathSegment(username)), nil, params, &result)
}

func (s *DefaultSecurityService) DisableUserWithParams(ctx context.Context, username string, params *SecurityRefreshSpecParams) (*types.AcknowledgedResponse, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/user/%s/_disable", pathSegment(username)), nil, params, &result)
}

func (s *DefaultSecurityService) EnableUserWithParams(ctx context.Context, username string, params *SecurityRefreshSpecParams) (*types.AcknowledgedResponse, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/user/%s/_enable", pathSegment(username)), nil, params, &result)
}

func (s *DefaultSecurityService) PutPrivilegesWithParams(ctx context.Context, body any, params *SecurityRefreshSpecParams) (*SecurityPutPrivilegesResponse, error) {
	var result SecurityPutPrivilegesResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", "/_security/privilege", body, params, &result)
}

func (s *DefaultSecurityService) PutRoleWithRequest(ctx context.Context, req *SecurityPutRoleRequest) (*SecurityCreateResponse, error) {
	if req == nil || req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result SecurityCreateResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/role/%s", pathSegment(req.Name)), req.Body, req.Params, &result)
}

func (s *DefaultSecurityService) PutRoleMappingWithRequest(ctx context.Context, req *SecurityPutRoleMappingRequest) (*SecurityRoleMappingCreateResponse, error) {
	if req == nil || req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result SecurityRoleMappingCreateResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/role_mapping/%s", pathSegment(req.Name)), req.Body, req.Params, &result)
}

func (s *DefaultSecurityService) PutUserWithRequest(ctx context.Context, req *SecurityPutUserRequest) (*SecurityCreateResponse, error) {
	if req == nil || req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	var result SecurityCreateResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/user/%s", pathSegment(req.Username)), req.Body, req.Params, &result)
}

func (s *DefaultSecurityService) GetUserWithParams(ctx context.Context, usernames []string, params *SecurityGetUserSpecParams) (map[string]*SecurityUser, error) {
	path := "/_security/user"
	if len(usernames) > 0 {
		path = fmt.Sprintf("/_security/user/%s", escapedList(usernames))
	}
	var result map[string]*SecurityUser
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultSecurityService) HasPrivilegesWithRequest(ctx context.Context, req *SecurityHasPrivilegesSpecRequest) (*SecurityHasPrivilegesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	path := "/_security/user/_has_privileges"
	if req.User != "" {
		path = fmt.Sprintf("/_security/user/%s/_has_privileges", pathSegment(req.User))
	}
	var result SecurityHasPrivilegesResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", path, req.Body, nil, &result)
}

// IndicesServiceSpecExtensions exposes the opt-in indices query carriers.
// Obtain it with [AsIndicesServiceSpecExtensions].
type IndicesServiceSpecExtensions interface {
	ExistsTemplateWithParams(ctx context.Context, name string, params *IndicesGetTemplateParams) (bool, error)
	ExistsIndexTemplateWithParams(ctx context.Context, name string, params *IndicesGetIndexTemplateParams) (bool, error)
	MigrateToDataStreamWithParams(ctx context.Context, name string, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	PromoteDataStreamWithParams(ctx context.Context, name string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error)
	ForceMergeWithRequest(ctx context.Context, req *IndicesForceMergeRequest) (*IndicesForceMergeSpecResponse, error)
}

var _ IndicesServiceSpecExtensions = (*DefaultIndicesService)(nil)

// AsIndicesServiceSpecExtensions adapts an IndicesService to its extension
// interface.
func AsIndicesServiceSpecExtensions(service IndicesService) (IndicesServiceSpecExtensions, error) {
	return asSpecExtensions[IndicesServiceSpecExtensions]("indices", service)
}

func (s *DefaultIndicesService) ExistsTemplateWithParams(ctx context.Context, name string, params *IndicesGetTemplateParams) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name is required")
	}
	return headReq(s.logger, ctx, s.client, fmt.Sprintf("/_template/%s", pathSegment(name)), params)
}

func (s *DefaultIndicesService) ExistsIndexTemplateWithParams(ctx context.Context, name string, params *IndicesGetIndexTemplateParams) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name is required")
	}
	return headReq(s.logger, ctx, s.client, fmt.Sprintf("/_index_template/%s", pathSegment(name)), params)
}

func (s *DefaultIndicesService) MigrateToDataStreamWithParams(ctx context.Context, name string, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_data_stream/_migrate/%s", pathSegment(name)), nil, params, &result)
}

func (s *DefaultIndicesService) PromoteDataStreamWithParams(ctx context.Context, name string, params *MasterTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_data_stream/_promote/%s", pathSegment(name)), nil, params, &result)
}

func (s *DefaultIndicesService) ForceMergeWithRequest(ctx context.Context, req *IndicesForceMergeRequest) (*IndicesForceMergeSpecResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	path := "/_forcemerge"
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_forcemerge", escapedList(req.Indices))
	}
	var result IndicesForceMergeSpecResponse
	err := doReq(s.logger, ctx, s.client, "POST", path, nil, req.Params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ClusterServiceSpecExtensions exposes the opt-in cluster query carriers.
// Obtain it with [AsClusterServiceSpecExtensions].
type ClusterServiceSpecExtensions interface {
	ExistsComponentTemplateWithParams(ctx context.Context, name string, params *ClusterGetComponentTemplateParams) (bool, error)
}

var _ ClusterServiceSpecExtensions = (*DefaultClusterService)(nil)

// AsClusterServiceSpecExtensions adapts a ClusterService to its extension
// interface.
func AsClusterServiceSpecExtensions(service ClusterService) (ClusterServiceSpecExtensions, error) {
	return asSpecExtensions[ClusterServiceSpecExtensions]("cluster", service)
}

func (s *DefaultClusterService) ExistsComponentTemplateWithParams(ctx context.Context, name string, params *ClusterGetComponentTemplateParams) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name is required")
	}
	return headReq(s.logger, ctx, s.client, fmt.Sprintf("/_component_template/%s", pathSegment(name)), params)
}

// DocumentServiceSpecExtensions exposes the opt-in mtermvectors carrier.
// Obtain it with [AsDocumentServiceSpecExtensions].
type DocumentServiceSpecExtensions interface {
	MultiTermVectorsWithRequest(ctx context.Context, req *MtermvectorsSpecRequest) (*MultiTermvectorResponse, error)
}

var _ DocumentServiceSpecExtensions = (*DefaultDocumentService)(nil)

// AsDocumentServiceSpecExtensions adapts a DocumentService to its extension
// interface.
func AsDocumentServiceSpecExtensions(service DocumentService) (DocumentServiceSpecExtensions, error) {
	return asSpecExtensions[DocumentServiceSpecExtensions]("document", service)
}

func (s *DefaultDocumentService) MultiTermVectorsWithRequest(ctx context.Context, req *MtermvectorsSpecRequest) (*MultiTermvectorResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	path := "/_mtermvectors"
	if req.Index != "" {
		path = fmt.Sprintf("/%s/_mtermvectors", pathSegment(req.Index))
	}
	var result MultiTermvectorResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", path, req.Body, req.Params, &result)
}

// EqlServiceSpecExtensions exposes the opt-in EQL query carriers. Obtain it
// with [AsEqlServiceSpecExtensions].
type EqlServiceSpecExtensions interface {
	GetWithParams(ctx context.Context, id string, params *EqlGetSpecParams) (*EqlSearchResponse, error)
}

var _ EqlServiceSpecExtensions = (*DefaultEqlService)(nil)

// AsEqlServiceSpecExtensions adapts an EqlService to its extension interface.
func AsEqlServiceSpecExtensions(service EqlService) (EqlServiceSpecExtensions, error) {
	return asSpecExtensions[EqlServiceSpecExtensions]("eql", service)
}

func (s *DefaultEqlService) GetWithParams(ctx context.Context, id string, params *EqlGetSpecParams) (*EqlSearchResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result EqlSearchResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_eql/search/%s", pathSegment(id)), nil, params, &result)
}

// LicenseServiceSpecExtensions exposes the opt-in license query carriers.
// Obtain it with [AsLicenseServiceSpecExtensions].
type LicenseServiceSpecExtensions interface {
	DeleteWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error)
}

var _ LicenseServiceSpecExtensions = (*DefaultLicenseService)(nil)

// AsLicenseServiceSpecExtensions adapts a LicenseService to its extension
// interface.
func AsLicenseServiceSpecExtensions(service LicenseService) (LicenseServiceSpecExtensions, error) {
	return asSpecExtensions[LicenseServiceSpecExtensions]("license", service)
}

func (s *DefaultLicenseService) DeleteWithParams(ctx context.Context, params *MasterTimeoutAndTimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", "/_license", nil, params, &result)
}

// MlServiceSpecExtensions exposes the opt-in ML put carriers. Obtain it with
// [AsMlServiceSpecExtensions].
type MlServiceSpecExtensions interface {
	PutJobWithRequest(ctx context.Context, req *MlPutJobSpecRequest) (*MlJobConfig, error)
	PutDatafeedWithRequest(ctx context.Context, req *MlPutDatafeedSpecRequest) (*MlDatafeedConfig, error)
}

var _ MlServiceSpecExtensions = (*DefaultMlService)(nil)

// AsMlServiceSpecExtensions adapts an MlService to its extension interface.
func AsMlServiceSpecExtensions(service MlService) (MlServiceSpecExtensions, error) {
	return asSpecExtensions[MlServiceSpecExtensions]("ml", service)
}

func (s *DefaultMlService) PutJobWithRequest(ctx context.Context, req *MlPutJobSpecRequest) (*MlJobConfig, error) {
	if req == nil || req.JobID == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	var result MlJobConfig
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_ml/anomaly_detectors/%s", pathSegment(req.JobID)), req.Body, req.Params, &result)
}

func (s *DefaultMlService) PutDatafeedWithRequest(ctx context.Context, req *MlPutDatafeedSpecRequest) (*MlDatafeedConfig, error) {
	if req == nil || req.DatafeedID == "" {
		return nil, fmt.Errorf("datafeed_id is required")
	}
	var result MlDatafeedConfig
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_ml/datafeeds/%s", pathSegment(req.DatafeedID)), req.Body, req.Params, &result)
}

// SearchableSnapshotsServiceSpecExtensions exposes the opt-in clear-cache
// carrier. Obtain it with [AsSearchableSnapshotsServiceSpecExtensions].
type SearchableSnapshotsServiceSpecExtensions interface {
	ClearCacheWithParams(ctx context.Context, indices []string, params *SearchableSnapshotsClearCacheSpecParams) (*types.AcknowledgedResponse, error)
}

var _ SearchableSnapshotsServiceSpecExtensions = (*DefaultSearchableSnapshotsService)(nil)

// AsSearchableSnapshotsServiceSpecExtensions adapts a SearchableSnapshotsService
// to its extension interface.
func AsSearchableSnapshotsServiceSpecExtensions(service SearchableSnapshotsService) (SearchableSnapshotsServiceSpecExtensions, error) {
	return asSpecExtensions[SearchableSnapshotsServiceSpecExtensions]("searchable_snapshots", service)
}

func (s *DefaultSearchableSnapshotsService) ClearCacheWithParams(ctx context.Context, indices []string, params *SearchableSnapshotsClearCacheSpecParams) (*types.AcknowledgedResponse, error) {
	r, err := buildRequest(s.client, ctx, nil, params)
	if err != nil {
		return nil, err
	}
	path := "/_searchable_snapshots/cache/clear"
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_searchable_snapshots/cache/clear", escapedList(indices))
		r.SetQueryParam("index", strings.Join(indices, ","))
	}
	resp, err := r.Post(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result types.AcknowledgedResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

// TransformServiceSpecExtensions exposes the opt-in schedule-now carrier.
// Obtain it with [AsTransformServiceSpecExtensions].
type TransformServiceSpecExtensions interface {
	ScheduleNowWithParams(ctx context.Context, transformId string, params *TimeoutSpecParams) (*types.AcknowledgedResponse, error)
}

var _ TransformServiceSpecExtensions = (*DefaultTransformService)(nil)

// AsTransformServiceSpecExtensions adapts a TransformService to its extension
// interface.
func AsTransformServiceSpecExtensions(service TransformService) (TransformServiceSpecExtensions, error) {
	return asSpecExtensions[TransformServiceSpecExtensions]("transform", service)
}

func (s *DefaultTransformService) ScheduleNowWithParams(ctx context.Context, transformId string, params *TimeoutSpecParams) (*types.AcknowledgedResponse, error) {
	if transformId == "" {
		return nil, fmt.Errorf("transform_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_transform/%s/_schedule_now", pathSegment(transformId)), nil, params, &result)
}

// SearchApplicationServiceSpecExtensions exposes the opt-in search carrier.
// Obtain it with [AsSearchApplicationServiceSpecExtensions].
type SearchApplicationServiceSpecExtensions interface {
	SearchWithRequest(ctx context.Context, req *SearchApplicationSearchSpecRequest) (*querydsl.SearchResult, error)
}

var _ SearchApplicationServiceSpecExtensions = (*DefaultSearchApplicationService)(nil)

// AsSearchApplicationServiceSpecExtensions adapts a SearchApplicationService to
// its extension interface.
func AsSearchApplicationServiceSpecExtensions(service SearchApplicationService) (SearchApplicationServiceSpecExtensions, error) {
	return asSpecExtensions[SearchApplicationServiceSpecExtensions]("search_application", service)
}

func (s *DefaultSearchApplicationService) SearchWithRequest(ctx context.Context, req *SearchApplicationSearchSpecRequest) (*querydsl.SearchResult, error) {
	if req == nil || req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result querydsl.SearchResult
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_application/search_application/%s/_search", pathSegment(req.Name)), req.Body, req.Params, &result)
}

// InferenceServiceSpecExtensions exposes the opt-in inference.update carrier.
// Obtain it with [AsInferenceServiceSpecExtensions].
type InferenceServiceSpecExtensions interface {
	UpdateWithRequest(ctx context.Context, req *InferenceUpdateSpecRequest) (*types.AcknowledgedResponse, error)
}

var _ InferenceServiceSpecExtensions = (*DefaultInferenceService)(nil)

// AsInferenceServiceSpecExtensions adapts an InferenceService to its extension
// interface.
func AsInferenceServiceSpecExtensions(service InferenceService) (InferenceServiceSpecExtensions, error) {
	return asSpecExtensions[InferenceServiceSpecExtensions]("inference", service)
}

func (s *DefaultInferenceService) UpdateWithRequest(ctx context.Context, req *InferenceUpdateSpecRequest) (*types.AcknowledgedResponse, error) {
	if req == nil || req.InferenceId == "" {
		return nil, fmt.Errorf("inference_id is required")
	}
	path := fmt.Sprintf("/_inference/%s/_update", pathSegment(req.InferenceId))
	if req.TaskType != "" {
		path = fmt.Sprintf("/_inference/%s/%s/_update", pathSegment(req.TaskType), pathSegment(req.InferenceId))
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", path, req.Body, nil, &result)
}

// TextStructureServiceSpecExtensions exposes correctly routed text-structure
// endpoints. The original TextStructureService methods are unchanged and keep
// using the legacy routes; these extensions target the pinned v9 routes.
// Obtain the extensions with [AsTextStructureServiceSpecExtensions].
type TextStructureServiceSpecExtensions interface {
	FindFieldStructureWithRequest(ctx context.Context, req *TextStructureFindFieldSpecRequest) (json.RawMessage, error)
	FindMessageStructureWithRequest(ctx context.Context, req *TextStructureFindMessageSpecRequest) (json.RawMessage, error)
}

var _ TextStructureServiceSpecExtensions = (*DefaultTextStructureService)(nil)

// AsTextStructureServiceSpecExtensions adapts a TextStructureService to its
// extension interface.
func AsTextStructureServiceSpecExtensions(service TextStructureService) (TextStructureServiceSpecExtensions, error) {
	return asSpecExtensions[TextStructureServiceSpecExtensions]("text_structure", service)
}

func (s *DefaultTextStructureService) FindFieldStructureWithRequest(ctx context.Context, req *TextStructureFindFieldSpecRequest) (json.RawMessage, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	r, err := buildRequest(s.client, ctx, nil, req.Params)
	if err != nil {
		return nil, err
	}
	// Required selectors are applied last so a Params carrier cannot drop them.
	r.SetQueryParam("index", req.Index)
	r.SetQueryParam("field", req.Field)
	resp, err := r.Get("/_text_structure/find_field_structure")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	return json.RawMessage(resp.Body()), nil
}

func (s *DefaultTextStructureService) FindMessageStructureWithRequest(ctx context.Context, req *TextStructureFindMessageSpecRequest) (json.RawMessage, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	r, err := buildRequest(s.client, ctx, map[string][]string{"messages": req.Messages}, req.Params)
	if err != nil {
		return nil, err
	}
	resp, err := r.Post("/_text_structure/find_message_structure")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	return json.RawMessage(resp.Body()), nil
}
