package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// ClusterService provides access to the cluster APIs.
type ClusterService interface {
	Health(ctx context.Context, indices []string, params *ClusterHealthParams) (*ClusterHealthResponse, error)
	State(ctx context.Context, req *ClusterStateRequest) (*ClusterStateResponse, error)
	Stats(ctx context.Context, nodeIds []string, params *ClusterStatsParams) (*ClusterStatsResponse, error)
	GetSettings(ctx context.Context, params *ClusterGetSettingsParams) (*ClusterGetSettingsResponse, error)
	PutSettings(ctx context.Context, body any, params *ClusterPutSettingsParams) (*ClusterPutSettingsResponse, error)
	AllocationExplain(ctx context.Context, body any, params *ClusterAllocationExplainParams) (*ClusterAllocationExplainResponse, error)
	Reroute(ctx context.Context, body any, params *ClusterRerouteParams) (*ClusterRerouteResponse, error)
	PendingTasks(ctx context.Context, params *ClusterPendingTasksParams) (*ClusterPendingTasksResponse, error)
	RemoteInfo(ctx context.Context) (*ClusterRemoteInfoResponse, error)
	Info(ctx context.Context, features []string) (*ClusterInfoResponse, error)
	PostVotingConfigExclusions(ctx context.Context, params *ClusterVotingConfigExclusionsParams) error
	DeleteVotingConfigExclusions(ctx context.Context, params *ClusterDeleteVotingConfigExclusionsParams) error
	GetComponentTemplate(ctx context.Context, names []string, params *ClusterGetComponentTemplateParams) (*ClusterGetComponentTemplateResponse, error)
	PutComponentTemplate(ctx context.Context, req *PutComponentTemplateRequest) (*types.AcknowledgedResponse, error)
	ExistsComponentTemplate(ctx context.Context, name string) (bool, error)
	DeleteComponentTemplate(ctx context.Context, name string, params *ClusterDeleteComponentTemplateParams) (*types.AcknowledgedResponse, error)
}

// DefaultClusterService is the default ClusterService implementation.
type DefaultClusterService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewClusterService creates a new ClusterService.
func NewClusterService(client *resty.Client, logger *logrus.Entry) ClusterService {
	return &DefaultClusterService{client: client, logger: logger.WithField("service", "cluster")}
}

func (s *DefaultClusterService) Health(ctx context.Context, indices []string, params *ClusterHealthParams) (*ClusterHealthResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/_cluster/health/%s", escapedList(indices))
	} else {
		path = "/_cluster/health"
	}
	var result ClusterHealthResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultClusterService) State(ctx context.Context, req *ClusterStateRequest) (*ClusterStateResponse, error) {
	var path string
	switch {
	case len(req.Metrics) > 0 && len(req.Indices) > 0:
		path = fmt.Sprintf("/_cluster/state/%s/%s", escapedList(req.Metrics), escapedList(req.Indices))
	case len(req.Metrics) > 0:
		path = fmt.Sprintf("/_cluster/state/%s", escapedList(req.Metrics))
	default:
		path = "/_cluster/state"
	}
	var result ClusterStateResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, req.Params, &result)
}

func (s *DefaultClusterService) Stats(ctx context.Context, nodeIds []string, params *ClusterStatsParams) (*ClusterStatsResponse, error) {
	var path string
	if len(nodeIds) > 0 {
		path = fmt.Sprintf("/_cluster/stats/nodes/%s", escapedList(nodeIds))
	} else {
		path = "/_cluster/stats"
	}
	var result ClusterStatsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultClusterService) GetSettings(ctx context.Context, params *ClusterGetSettingsParams) (*ClusterGetSettingsResponse, error) {
	var result ClusterGetSettingsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_cluster/settings", nil, params, &result)
}

func (s *DefaultClusterService) PutSettings(ctx context.Context, body any, params *ClusterPutSettingsParams) (*ClusterPutSettingsResponse, error) {
	var result ClusterPutSettingsResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", "/_cluster/settings", body, params, &result)
}

func (s *DefaultClusterService) AllocationExplain(ctx context.Context, body any, params *ClusterAllocationExplainParams) (*ClusterAllocationExplainResponse, error) {
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	var resp *resty.Response
	var err error
	if body != nil {
		r = r.SetBody(body)
		resp, err = r.Post("/_cluster/allocation/explain")
	} else {
		resp, err = r.Get("/_cluster/allocation/explain")
	}
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result ClusterAllocationExplainResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultClusterService) Reroute(ctx context.Context, body any, params *ClusterRerouteParams) (*ClusterRerouteResponse, error) {
	var result ClusterRerouteResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_cluster/reroute", body, params, &result)
}

func (s *DefaultClusterService) PendingTasks(ctx context.Context, params *ClusterPendingTasksParams) (*ClusterPendingTasksResponse, error) {
	var result ClusterPendingTasksResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_cluster/pending_tasks", nil, params, &result)
}

func (s *DefaultClusterService) RemoteInfo(ctx context.Context) (*ClusterRemoteInfoResponse, error) {
	var result ClusterRemoteInfoResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_remote/info", nil, nil, &result)
}

func (s *DefaultClusterService) Info(ctx context.Context, features []string) (*ClusterInfoResponse, error) {
	var path string
	if len(features) > 0 {
		path = fmt.Sprintf("/_info/%s", escapedList(features))
	} else {
		path = "/_info/_all"
	}
	var result ClusterInfoResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultClusterService) PostVotingConfigExclusions(ctx context.Context, params *ClusterVotingConfigExclusionsParams) error {
	return doReq(s.logger, ctx, s.client, "POST", "/_cluster/voting_config_exclusions", nil, params, nil)
}

func (s *DefaultClusterService) DeleteVotingConfigExclusions(ctx context.Context, params *ClusterDeleteVotingConfigExclusionsParams) error {
	return doReq(s.logger, ctx, s.client, "DELETE", "/_cluster/voting_config_exclusions", nil, params, nil)
}

func (s *DefaultClusterService) GetComponentTemplate(ctx context.Context, names []string, params *ClusterGetComponentTemplateParams) (*ClusterGetComponentTemplateResponse, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_component_template/%s", escapedList(names))
	} else {
		path = "/_component_template"
	}
	var result ClusterGetComponentTemplateResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultClusterService) PutComponentTemplate(ctx context.Context, req *PutComponentTemplateRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_component_template/%s", pathSegment(req.Name)), req.Body, req.Params, &result)
}

func (s *DefaultClusterService) ExistsComponentTemplate(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name is required")
	}
	return headReq(s.logger, ctx, s.client, fmt.Sprintf("/_component_template/%s", pathSegment(name)), nil)
}

func (s *DefaultClusterService) DeleteComponentTemplate(ctx context.Context, name string, params *ClusterDeleteComponentTemplateParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_component_template/%s", pathSegment(name)), nil, params, &result)
}
