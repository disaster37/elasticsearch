package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// AutoscalingService provides access to the autoscaling APIs.
type AutoscalingService interface {
	GetCapacity(ctx context.Context) (json.RawMessage, error)
	GetPolicy(ctx context.Context, name string) (json.RawMessage, error)
	PutPolicy(ctx context.Context, name string, body any) (*types.AcknowledgedResponse, error)
	DeletePolicy(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
}

// DefaultAutoscalingService is the default AutoscalingService implementation.
type DefaultAutoscalingService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewAutoscalingService creates a new AutoscalingService.
func NewAutoscalingService(client *resty.Client, logger *logrus.Entry) AutoscalingService {
	return &DefaultAutoscalingService{client: client, logger: logger.WithField("service", "autoscaling")}
}

func (s *DefaultAutoscalingService) GetCapacity(ctx context.Context) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "GET", "/_autoscaling/capacity", nil, nil)
}

func (s *DefaultAutoscalingService) GetPolicy(ctx context.Context, name string) (json.RawMessage, error) {
	if name == "" {
		return doReqRaw(s.logger, ctx, s.client, "GET", "/_autoscaling/policy", nil, nil)
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_autoscaling/policy/%s", pathSegment(name)), nil, nil)
}

func (s *DefaultAutoscalingService) PutPolicy(ctx context.Context, name string, body any) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_autoscaling/policy/%s", pathSegment(name)), body, nil, &result)
}

func (s *DefaultAutoscalingService) DeletePolicy(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_autoscaling/policy/%s", pathSegment(name)), nil, nil, &result)
}
