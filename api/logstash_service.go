package api

import (
	"context"
	"fmt"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// LogstashService provides access to the logstash APIs.
type LogstashService interface {
	GetPipeline(ctx context.Context, ids []string) (map[string]map[string]any, error)
	PutPipeline(ctx context.Context, id string, body any) (*types.AcknowledgedResponse, error)
	DeletePipeline(ctx context.Context, id string) (*types.AcknowledgedResponse, error)
}

// DefaultLogstashService is the default LogstashService implementation.
type DefaultLogstashService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewLogstashService creates a new LogstashService.
func NewLogstashService(client *resty.Client, logger *logrus.Entry) LogstashService {
	return &DefaultLogstashService{client: client, logger: logger.WithField("service", "logstash")}
}

func (s *DefaultLogstashService) GetPipeline(ctx context.Context, ids []string) (map[string]map[string]any, error) {
	var path string
	if len(ids) > 0 {
		path = fmt.Sprintf("/_logstash/pipeline/%s", escapedList(ids))
	} else {
		path = "/_logstash/pipeline"
	}
	var result map[string]map[string]any
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultLogstashService) PutPipeline(ctx context.Context, id string, body any) (*types.AcknowledgedResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_logstash/pipeline/%s", pathSegment(id)), body, nil, &result)
}

func (s *DefaultLogstashService) DeletePipeline(ctx context.Context, id string) (*types.AcknowledgedResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_logstash/pipeline/%s", pathSegment(id)), nil, nil, &result)
}
