package api

import (
	"context"
	"fmt"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// CcrFollowParams are the query parameters for CCR follow.
type CcrFollowParams struct {
	Common              *CommonParams
	MasterTimeout       string
	WaitForActiveShards string
}

// ToMap converts CcrFollowParams to a query-parameter map.
func (p *CcrFollowParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	return mergeParams(m, p.Common)
}

// CcrGetAutoFollowPatternResponse is the response from get auto-follow pattern.
type CcrGetAutoFollowPatternResponse struct {
	Patterns []struct {
		Name    string         `json:"name"`
		Pattern map[string]any `json:"pattern"`
	} `json:"patterns"`
}

// CcrFollowResponse is the response from follow index.
type CcrFollowResponse struct {
	FollowIndex                   string `json:"follow_index"`
	ShardsAcknowledged            bool   `json:"shards_acknowledged"`
	FollowIndexCreated            bool   `json:"follow_index_created"`
	FollowIndexShardsAcknowledged bool   `json:"follow_index_shards_acknowledged"`
}

// CcrFollowInfoResponse is the response from follow info.
type CcrFollowInfoResponse struct {
	Indices []struct {
		Index         string `json:"index"`
		FollowIndex   string `json:"follow_index"`
		RemoteCluster string `json:"remote_cluster"`
		LeaderIndex   string `json:"leader_index"`
	} `json:"follower_indices"`
}

// CcrFollowStatsResponse is the response from follow stats.
type CcrFollowStatsResponse struct {
	Indices []struct {
		Index  string         `json:"index"`
		Status map[string]any `json:"shards"`
	} `json:"indices"`
}

// CcrStatsResponse is the response from CCR stats.
type CcrStatsResponse struct {
	AutoFollowStats map[string]any `json:"auto_follow_stats"`
	FollowStats     map[string]any `json:"follow_stats"`
}

// CcrForgetFollowerResponse is the response from forget follower.
type CcrForgetFollowerResponse struct {
	Shards *types.ShardsInfo `json:"_shards,omitempty"`
}

// CcrService provides access to the cross-cluster replication APIs.
type CcrService interface {
	PutAutoFollowPattern(ctx context.Context, name string, body any) (*types.AcknowledgedResponse, error)
	GetAutoFollowPattern(ctx context.Context, names []string) (*CcrGetAutoFollowPatternResponse, error)
	DeleteAutoFollowPattern(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
	PauseAutoFollowPattern(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
	ResumeAutoFollowPattern(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
	Follow(ctx context.Context, index string, body any, params *CcrFollowParams) (*CcrFollowResponse, error)
	PauseFollow(ctx context.Context, index string) (*types.AcknowledgedResponse, error)
	ResumeFollow(ctx context.Context, index string, body any) (*types.AcknowledgedResponse, error)
	Unfollow(ctx context.Context, index string) (*types.AcknowledgedResponse, error)
	ForgetFollower(ctx context.Context, index string, body any) (*CcrForgetFollowerResponse, error)
	FollowInfo(ctx context.Context, indices []string) (*CcrFollowInfoResponse, error)
	FollowStats(ctx context.Context, indices []string) (*CcrFollowStatsResponse, error)
	Stats(ctx context.Context) (*CcrStatsResponse, error)
}

// DefaultCcrService is the default CcrService implementation.
type DefaultCcrService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewCcrService creates a new CcrService.
func NewCcrService(client *resty.Client, logger *logrus.Entry) CcrService {
	return &DefaultCcrService{client: client, logger: logger.WithField("service", "ccr")}
}

func (s *DefaultCcrService) PutAutoFollowPattern(ctx context.Context, name string, body any) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_ccr/auto_follow/%s", pathSegment(name)), body, nil, &result)
}

func (s *DefaultCcrService) GetAutoFollowPattern(ctx context.Context, names []string) (*CcrGetAutoFollowPatternResponse, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_ccr/auto_follow/%s", escapedList(names))
	} else {
		path = "/_ccr/auto_follow"
	}
	var result CcrGetAutoFollowPatternResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultCcrService) DeleteAutoFollowPattern(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_ccr/auto_follow/%s", pathSegment(name)), nil, nil, &result)
}

func (s *DefaultCcrService) PauseAutoFollowPattern(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ccr/auto_follow/%s/pause", pathSegment(name)), nil, nil, &result)
}

func (s *DefaultCcrService) ResumeAutoFollowPattern(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ccr/auto_follow/%s/resume", pathSegment(name)), nil, nil, &result)
}

func (s *DefaultCcrService) Follow(ctx context.Context, index string, body any, params *CcrFollowParams) (*CcrFollowResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result CcrFollowResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ccr/follow", pathSegment(index)), body, params, &result)
}

func (s *DefaultCcrService) PauseFollow(ctx context.Context, index string) (*types.AcknowledgedResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ccr/pause_follow", pathSegment(index)), nil, nil, &result)
}

func (s *DefaultCcrService) ResumeFollow(ctx context.Context, index string, body any) (*types.AcknowledgedResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ccr/resume_follow", pathSegment(index)), body, nil, &result)
}

func (s *DefaultCcrService) Unfollow(ctx context.Context, index string) (*types.AcknowledgedResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ccr/unfollow", pathSegment(index)), nil, nil, &result)
}

func (s *DefaultCcrService) ForgetFollower(ctx context.Context, index string, body any) (*CcrForgetFollowerResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result CcrForgetFollowerResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ccr/forget_follower", pathSegment(index)), body, nil, &result)
}

func (s *DefaultCcrService) FollowInfo(ctx context.Context, indices []string) (*CcrFollowInfoResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_ccr/info", escapedList(indices))
	} else {
		path = "/_ccr/info"
	}
	var result CcrFollowInfoResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultCcrService) FollowStats(ctx context.Context, indices []string) (*CcrFollowStatsResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_ccr/stats", escapedList(indices))
	} else {
		path = "/_ccr/stats"
	}
	var result CcrFollowStatsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultCcrService) Stats(ctx context.Context) (*CcrStatsResponse, error) {
	var result CcrStatsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_ccr/stats", nil, nil, &result)
}
