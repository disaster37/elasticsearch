package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/querydsl"
	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// SearchApplicationListParams are the query parameters for list.
type SearchApplicationListParams struct {
	Common *CommonParams
	From   *int
	Q      string
	Size   *int
}

// ToMap converts SearchApplicationListParams to a query-parameter map.
func (p *SearchApplicationListParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Q != "" {
		m["q"] = p.Q
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	return mergeParams(m, p.Common)
}

// SearchApplicationPutParams are the query parameters for put.
type SearchApplicationPutParams struct {
	Common *CommonParams
	Create *bool
}

// ToMap converts SearchApplicationPutParams to a query-parameter map.
func (p *SearchApplicationPutParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "create", p.Create)
	return mergeParams(m, p.Common)
}

// SearchApplicationService provides access to the search application APIs.
type SearchApplicationService interface {
	List(ctx context.Context, params *SearchApplicationListParams) (json.RawMessage, error)
	Get(ctx context.Context, name string) (json.RawMessage, error)
	Put(ctx context.Context, name string, body any, params *SearchApplicationPutParams) (*types.AcknowledgedResponse, error)
	Delete(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
	Search(ctx context.Context, name string, body any) (*querydsl.SearchResult, error)
	GetBehavioralAnalytics(ctx context.Context, name string) (json.RawMessage, error)
	PutBehavioralAnalytics(ctx context.Context, name string, body any) (*types.AcknowledgedResponse, error)
	DeleteBehavioralAnalytics(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
}

// DefaultSearchApplicationService is the default SearchApplicationService implementation.
type DefaultSearchApplicationService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewSearchApplicationService creates a new SearchApplicationService.
func NewSearchApplicationService(client *resty.Client, logger *logrus.Entry) SearchApplicationService {
	return &DefaultSearchApplicationService{client: client, logger: logger.WithField("service", "search_application")}
}

func (s *DefaultSearchApplicationService) List(ctx context.Context, params *SearchApplicationListParams) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "GET", "/_application/search_application", nil, params)
}

func (s *DefaultSearchApplicationService) Get(ctx context.Context, name string) (json.RawMessage, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_application/search_application/%s", pathSegment(name)), nil, nil)
}

func (s *DefaultSearchApplicationService) Put(ctx context.Context, name string, body any, params *SearchApplicationPutParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_application/search_application/%s", pathSegment(name)), body, params, &result)
}

func (s *DefaultSearchApplicationService) Delete(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_application/search_application/%s", pathSegment(name)), nil, nil, &result)
}

func (s *DefaultSearchApplicationService) Search(ctx context.Context, name string, body any) (*querydsl.SearchResult, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result querydsl.SearchResult
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_application/search_application/%s/_search", pathSegment(name)), body, nil, &result)
}

func (s *DefaultSearchApplicationService) GetBehavioralAnalytics(ctx context.Context, name string) (json.RawMessage, error) {
	var path string
	if name != "" {
		path = fmt.Sprintf("/_application/analytics/%s", pathSegment(name))
	} else {
		path = "/_application/analytics"
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", path, nil, nil)
}

func (s *DefaultSearchApplicationService) PutBehavioralAnalytics(ctx context.Context, name string, body any) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_application/analytics/%s", pathSegment(name)), body, nil, &result)
}

func (s *DefaultSearchApplicationService) DeleteBehavioralAnalytics(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_application/analytics/%s", pathSegment(name)), nil, nil, &result)
}
