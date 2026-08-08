package api

import (
	"context"
	"fmt"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	json "github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
)

// EnrichWaitForCompletionParams are the query parameters for execute policy.
type EnrichWaitForCompletionParams struct {
	Common            *CommonParams
	WaitForCompletion *bool
}

// ToMap converts EnrichWaitForCompletionParams to a query-parameter map.
func (p *EnrichWaitForCompletionParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// EnrichService provides access to the enrich APIs.
type EnrichService interface {
	GetPolicy(ctx context.Context, names []string) (json.RawMessage, error)
	PutPolicy(ctx context.Context, name string, body any) (*types.AcknowledgedResponse, error)
	DeletePolicy(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
	ExecutePolicy(ctx context.Context, name string, params *EnrichWaitForCompletionParams) (*types.AcknowledgedResponse, error)
	Stats(ctx context.Context) (json.RawMessage, error)
}

// DefaultEnrichService is the default EnrichService implementation.
type DefaultEnrichService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewEnrichService creates a new EnrichService.
func NewEnrichService(client *resty.Client, logger *logrus.Entry) EnrichService {
	return &DefaultEnrichService{client: client, logger: logger.WithField("service", "enrich")}
}

func (s *DefaultEnrichService) GetPolicy(ctx context.Context, names []string) (json.RawMessage, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_enrich/policy/%s", escapedList(names))
	} else {
		path = "/_enrich/policy"
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", path, nil, nil)
}

func (s *DefaultEnrichService) PutPolicy(ctx context.Context, name string, body any) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_enrich/policy/%s", pathSegment(name)), body, nil, &result)
}

func (s *DefaultEnrichService) DeletePolicy(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_enrich/policy/%s", pathSegment(name)), nil, nil, &result)
}

func (s *DefaultEnrichService) ExecutePolicy(ctx context.Context, name string, params *EnrichWaitForCompletionParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_enrich/policy/%s/_execute", pathSegment(name)), nil, params, &result)
}

func (s *DefaultEnrichService) Stats(ctx context.Context) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "GET", "/_enrich/_stats", nil, nil)
}
