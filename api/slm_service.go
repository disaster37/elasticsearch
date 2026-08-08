package api

import (
	"context"
	"fmt"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// SlmService provides access to the snapshot lifecycle management APIs.
type SlmService interface {
	GetLifecycle(ctx context.Context, policyIds []string) (map[string]*SlmPolicy, error)
	PutLifecycle(ctx context.Context, policyId string, body any) (*types.AcknowledgedResponse, error)
	DeleteLifecycle(ctx context.Context, policyId string) (*types.AcknowledgedResponse, error)
	ExecuteLifecycle(ctx context.Context, policyId string) (*SlmExecuteResponse, error)
	ExecuteRetention(ctx context.Context) (*types.AcknowledgedResponse, error)
	GetStats(ctx context.Context) (*SlmStatsResponse, error)
	GetStatus(ctx context.Context) (*SlmStatusResponse, error)
	Start(ctx context.Context) (*types.AcknowledgedResponse, error)
	Stop(ctx context.Context) (*types.AcknowledgedResponse, error)
}

// DefaultSlmService is the default SlmService implementation.
type DefaultSlmService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewSlmService creates a new SlmService.
func NewSlmService(client *resty.Client, logger *logrus.Entry) SlmService {
	return &DefaultSlmService{client: client, logger: logger.WithField("service", "slm")}
}

func (s *DefaultSlmService) GetLifecycle(ctx context.Context, policyIds []string) (map[string]*SlmPolicy, error) {
	var path string
	if len(policyIds) > 0 {
		path = fmt.Sprintf("/_slm/policy/%s", escapedList(policyIds))
	} else {
		path = "/_slm/policy"
	}
	var result map[string]*SlmPolicy
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultSlmService) PutLifecycle(ctx context.Context, policyId string, body any) (*types.AcknowledgedResponse, error) {
	if policyId == "" {
		return nil, fmt.Errorf("policy_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_slm/policy/%s", pathSegment(policyId)), body, nil, &result)
}

func (s *DefaultSlmService) DeleteLifecycle(ctx context.Context, policyId string) (*types.AcknowledgedResponse, error) {
	if policyId == "" {
		return nil, fmt.Errorf("policy_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_slm/policy/%s", pathSegment(policyId)), nil, nil, &result)
}

func (s *DefaultSlmService) ExecuteLifecycle(ctx context.Context, policyId string) (*SlmExecuteResponse, error) {
	if policyId == "" {
		return nil, fmt.Errorf("policy_id is required")
	}
	var result SlmExecuteResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_slm/policy/%s/_execute", pathSegment(policyId)), nil, nil, &result)
}

func (s *DefaultSlmService) ExecuteRetention(ctx context.Context) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_slm/_execute_retention", nil, nil, &result)
}

func (s *DefaultSlmService) GetStats(ctx context.Context) (*SlmStatsResponse, error) {
	var result SlmStatsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_slm/stats", nil, nil, &result)
}

func (s *DefaultSlmService) GetStatus(ctx context.Context) (*SlmStatusResponse, error) {
	var result SlmStatusResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_slm/status", nil, nil, &result)
}

func (s *DefaultSlmService) Start(ctx context.Context) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_slm/start", nil, nil, &result)
}

func (s *DefaultSlmService) Stop(ctx context.Context) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_slm/stop", nil, nil, &result)
}
