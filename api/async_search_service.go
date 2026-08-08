package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// AsyncSearchService provides access to the async search APIs.
type AsyncSearchService interface {
	Submit(ctx context.Context, req *AsyncSearchSubmitRequest) (*AsyncSearchResponse, error)
	Get(ctx context.Context, id string, params *AsyncSearchGetParams) (*AsyncSearchResponse, error)
	Status(ctx context.Context, id string) (*AsyncSearchStatusResponse, error)
	Delete(ctx context.Context, id string) (*types.AcknowledgedResponse, error)
}

// DefaultAsyncSearchService is the default AsyncSearchService implementation.
type DefaultAsyncSearchService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewAsyncSearchService creates a new AsyncSearchService.
func NewAsyncSearchService(client *resty.Client, logger *logrus.Entry) AsyncSearchService {
	return &DefaultAsyncSearchService{client: client, logger: logger.WithField("service", "async_search")}
}

func (s *DefaultAsyncSearchService) Submit(ctx context.Context, req *AsyncSearchSubmitRequest) (*AsyncSearchResponse, error) {
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_async_search", escapedList(req.Indices))
	} else {
		path = "/_async_search"
	}
	r := s.client.R().SetContext(ctx)
	if req.Body != nil {
		r = r.SetBody(req.Body)
	}
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
	var result AsyncSearchResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultAsyncSearchService) Get(ctx context.Context, id string, params *AsyncSearchGetParams) (*AsyncSearchResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(fmt.Sprintf("/_async_search/%s", pathSegment(id)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result AsyncSearchResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultAsyncSearchService) Status(ctx context.Context, id string) (*AsyncSearchStatusResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	resp, err := s.client.R().SetContext(ctx).Get(fmt.Sprintf("/_async_search/status/%s", pathSegment(id)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result AsyncSearchStatusResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultAsyncSearchService) Delete(ctx context.Context, id string) (*types.AcknowledgedResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	resp, err := s.client.R().SetContext(ctx).Delete(fmt.Sprintf("/_async_search/%s", pathSegment(id)))
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
