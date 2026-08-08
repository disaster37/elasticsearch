package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// NodesService provides access to the nodes APIs.
type NodesService interface {
	Info(ctx context.Context, req *NodesInfoRequest) (*NodesInfoResponse, error)
	Stats(ctx context.Context, req *NodesStatsRequest) (*NodesStatsResponse, error)
	Usage(ctx context.Context, req *NodesUsageRequest) (*NodesUsageResponse, error)
	HotThreads(ctx context.Context, nodeIds []string, params *NodesHotThreadsParams) (string, error)
	ReloadSecureSettings(ctx context.Context, nodeIds []string, body any, params *NodesReloadSecureSettingsParams) (*NodesReloadSecureSettingsResponse, error)
}

// DefaultNodesService is the default NodesService implementation.
type DefaultNodesService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewNodesService creates a new NodesService.
func NewNodesService(client *resty.Client, logger *logrus.Entry) NodesService {
	return &DefaultNodesService{client: client, logger: logger.WithField("service", "nodes")}
}

func (s *DefaultNodesService) Info(ctx context.Context, req *NodesInfoRequest) (*NodesInfoResponse, error) {
	path := nodesPath("/_nodes", req.NodeIds, req.Metrics)
	r := s.client.R().SetContext(ctx)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result NodesInfoResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultNodesService) Stats(ctx context.Context, req *NodesStatsRequest) (*NodesStatsResponse, error) {
	var path string
	switch {
	case len(req.NodeIds) > 0 && len(req.Metrics) > 0 && len(req.IndexMetrics) > 0:
		path = fmt.Sprintf("/_nodes/%s/stats/%s/%s", escapedList(req.NodeIds), escapedList(req.Metrics), escapedList(req.IndexMetrics))
	case len(req.NodeIds) > 0 && len(req.Metrics) > 0:
		path = fmt.Sprintf("/_nodes/%s/stats/%s", escapedList(req.NodeIds), escapedList(req.Metrics))
	case len(req.NodeIds) > 0:
		path = fmt.Sprintf("/_nodes/%s/stats", escapedList(req.NodeIds))
	case len(req.Metrics) > 0 && len(req.IndexMetrics) > 0:
		path = fmt.Sprintf("/_nodes/stats/%s/%s", escapedList(req.Metrics), escapedList(req.IndexMetrics))
	case len(req.Metrics) > 0:
		path = fmt.Sprintf("/_nodes/stats/%s", escapedList(req.Metrics))
	default:
		path = "/_nodes/stats"
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result NodesStatsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultNodesService) Usage(ctx context.Context, req *NodesUsageRequest) (*NodesUsageResponse, error) {
	path := nodesPath("/_nodes/usage", req.NodeIds, req.Metrics)
	r := s.client.R().SetContext(ctx)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result NodesUsageResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultNodesService) HotThreads(ctx context.Context, nodeIds []string, params *NodesHotThreadsParams) (string, error) {
	var path string
	if len(nodeIds) > 0 {
		path = fmt.Sprintf("/_nodes/%s/hot_threads", escapedList(nodeIds))
	} else {
		path = "/_nodes/hot_threads"
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return "", wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return "", logAndReturnError(s.logger, resp)
	}
	return string(resp.Body()), nil
}

func (s *DefaultNodesService) ReloadSecureSettings(ctx context.Context, nodeIds []string, body any, params *NodesReloadSecureSettingsParams) (*NodesReloadSecureSettingsResponse, error) {
	var path string
	if len(nodeIds) > 0 {
		path = fmt.Sprintf("/_nodes/%s/reload_secure_settings", escapedList(nodeIds))
	} else {
		path = "/_nodes/reload_secure_settings"
	}
	r := s.client.R().SetContext(ctx)
	if body != nil {
		r = r.SetBody(body)
	}
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result NodesReloadSecureSettingsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

// nodesPath builds a /_nodes[/{node_ids}][/{metrics}] path.
func nodesPath(base string, nodeIds, metrics []string) string {
	if len(nodeIds) > 0 && len(metrics) > 0 {
		return fmt.Sprintf("%s/%s/%s", base, escapedList(nodeIds), escapedList(metrics))
	}
	if len(nodeIds) > 0 {
		return fmt.Sprintf("%s/%s", base, escapedList(nodeIds))
	}
	if len(metrics) > 0 {
		return fmt.Sprintf("%s/%s", base, escapedList(metrics))
	}
	return base
}
