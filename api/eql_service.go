package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// EqlSearchParams are the query parameters for EQL search.
type EqlSearchParams struct {
	Common                   *CommonParams
	KeepAlive                string
	KeepOnCompletion         *bool
	WaitForCompletionTimeout string
}

// ToMap converts EqlSearchParams to a query-parameter map.
func (p *EqlSearchParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.KeepAlive != "" {
		m["keep_alive"] = p.KeepAlive
	}
	setBoolParam(m, "keep_on_completion", p.KeepOnCompletion)
	if p.WaitForCompletionTimeout != "" {
		m["wait_for_completion_timeout"] = p.WaitForCompletionTimeout
	}
	return mergeParams(m, p.Common)
}

// EqlSearchResponse is the response from EQL search.
type EqlSearchResponse struct {
	ID        string         `json:"id,omitempty"`
	IsPartial bool           `json:"is_partial"`
	IsRunning bool           `json:"is_running"`
	TimedOut  bool           `json:"timed_out"`
	Took      int64          `json:"took"`
	Hits      map[string]any `json:"hits,omitempty"`
}

// EqlService provides access to the EQL APIs.
type EqlService interface {
	Search(ctx context.Context, index string, body any, params *EqlSearchParams) (*EqlSearchResponse, error)
	Get(ctx context.Context, id string) (*EqlSearchResponse, error)
	GetStatus(ctx context.Context, id string) (json.RawMessage, error)
	Delete(ctx context.Context, id string) (*types.AcknowledgedResponse, error)
}

// DefaultEqlService is the default EqlService implementation.
type DefaultEqlService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewEqlService creates a new EqlService.
func NewEqlService(client *resty.Client, logger *logrus.Entry) EqlService {
	return &DefaultEqlService{client: client, logger: logger.WithField("service", "eql")}
}

func (s *DefaultEqlService) Search(ctx context.Context, index string, body any, params *EqlSearchParams) (*EqlSearchResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result EqlSearchResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_eql/search", pathSegment(index)), body, params, &result)
}

func (s *DefaultEqlService) Get(ctx context.Context, id string) (*EqlSearchResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result EqlSearchResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_eql/search/%s", pathSegment(id)), nil, nil, &result)
}

func (s *DefaultEqlService) GetStatus(ctx context.Context, id string) (json.RawMessage, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_eql/search/status/%s", pathSegment(id)), nil, nil)
}

func (s *DefaultEqlService) Delete(ctx context.Context, id string) (*types.AcknowledgedResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_eql/search/%s", pathSegment(id)), nil, nil, &result)
}
