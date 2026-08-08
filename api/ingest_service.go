package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// IngestService provides access to the ingest APIs.
type IngestService interface {
	GetPipeline(ctx context.Context, ids []string, params *IngestGetPipelineParams) (map[string]*IngestPipeline, error)
	PutPipeline(ctx context.Context, req *IngestPutPipelineRequest) (*types.AcknowledgedResponse, error)
	DeletePipeline(ctx context.Context, id string, params *IngestDeletePipelineParams) (*types.AcknowledgedResponse, error)
	Simulate(ctx context.Context, req *IngestSimulatePipelineRequest) (*IngestSimulateResponse, error)
	ProcessorGrok(ctx context.Context) (*IngestGrokResponse, error)
}

// DefaultIngestService is the default IngestService implementation.
type DefaultIngestService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewIngestService creates a new IngestService.
func NewIngestService(client *resty.Client, logger *logrus.Entry) IngestService {
	return &DefaultIngestService{client: client, logger: logger.WithField("service", "ingest")}
}

func (s *DefaultIngestService) GetPipeline(ctx context.Context, ids []string, params *IngestGetPipelineParams) (map[string]*IngestPipeline, error) {
	var path string
	if len(ids) > 0 {
		path = fmt.Sprintf("/_ingest/pipeline/%s", joinIndices(ids))
	} else {
		path = "/_ingest/pipeline"
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result map[string]*IngestPipeline
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return result, nil
}

func (s *DefaultIngestService) PutPipeline(ctx context.Context, req *IngestPutPipelineRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	r := s.client.R().SetContext(ctx).SetBody(req.Body)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Put(fmt.Sprintf("/_ingest/pipeline/%s", pathSegment(req.Id)))
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

func (s *DefaultIngestService) DeletePipeline(ctx context.Context, id string, params *IngestDeletePipelineParams) (*types.AcknowledgedResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Delete(fmt.Sprintf("/_ingest/pipeline/%s", pathSegment(id)))
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

func (s *DefaultIngestService) Simulate(ctx context.Context, req *IngestSimulatePipelineRequest) (*IngestSimulateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if req.Id != "" {
		path = fmt.Sprintf("/_ingest/pipeline/%s/_simulate", pathSegment(req.Id))
	} else {
		path = "/_ingest/pipeline/_simulate"
	}
	r := s.client.R().SetContext(ctx).SetBody(req.Body)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IngestSimulateResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIngestService) ProcessorGrok(ctx context.Context) (*IngestGrokResponse, error) {
	resp, err := s.client.R().SetContext(ctx).Get("/_ingest/processor/grok")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IngestGrokResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}
