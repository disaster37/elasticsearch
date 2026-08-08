package api

import (
	"context"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// StreamsService provides access to the streams APIs (Elasticsearch 9.x).
// These endpoints are beta/experimental and may not exist in every 9.x patch;
// the speccheck allow-list accounts for their conditional presence.
type StreamsService interface {
	LogsEnable(ctx context.Context, body any) (*types.AcknowledgedResponse, error)
	LogsDisable(ctx context.Context) (*types.AcknowledgedResponse, error)
	Status(ctx context.Context) (json.RawMessage, error)
}

// DefaultStreamsService is the default StreamsService implementation.
type DefaultStreamsService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewStreamsService creates a new StreamsService.
func NewStreamsService(client *resty.Client, logger *logrus.Entry) StreamsService {
	return &DefaultStreamsService{client: client, logger: logger.WithField("service", "streams")}
}

func (s *DefaultStreamsService) LogsEnable(ctx context.Context, body any) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_streams/logs/_enable", body, nil, &result)
}

func (s *DefaultStreamsService) LogsDisable(ctx context.Context) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_streams/logs/_disable", nil, nil, &result)
}

func (s *DefaultStreamsService) Status(ctx context.Context) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "GET", "/_streams/logs/_status", nil, nil)
}
