package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// ScriptService provides access to the stored script APIs.
type ScriptService interface {
	Get(ctx context.Context, id string, params *ScriptGetParams) (*ScriptGetResponse, error)
	Put(ctx context.Context, req *ScriptPutRequest) (*types.AcknowledgedResponse, error)
	Delete(ctx context.Context, id string, params *ScriptDeleteParams) (*types.AcknowledgedResponse, error)
	GetContext(ctx context.Context) (*ScriptContextResponse, error)
	GetLanguages(ctx context.Context) (*ScriptLanguagesResponse, error)
}

// DefaultScriptService is the default ScriptService implementation.
type DefaultScriptService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewScriptService creates a new ScriptService.
func NewScriptService(client *resty.Client, logger *logrus.Entry) ScriptService {
	return &DefaultScriptService{client: client, logger: logger.WithField("service", "script")}
}

func (s *DefaultScriptService) Get(ctx context.Context, id string, params *ScriptGetParams) (*ScriptGetResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(fmt.Sprintf("/_scripts/%s", pathSegment(id)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.StatusCode() == 404 {
		return &ScriptGetResponse{Found: false}, nil
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result ScriptGetResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultScriptService) Put(ctx context.Context, req *ScriptPutRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_scripts/%s", pathSegment(req.Id)), req.Body, req.Params, &result)
}

func (s *DefaultScriptService) Delete(ctx context.Context, id string, params *ScriptDeleteParams) (*types.AcknowledgedResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_scripts/%s", pathSegment(id)), nil, params, &result)
}

func (s *DefaultScriptService) GetContext(ctx context.Context) (*ScriptContextResponse, error) {
	var result ScriptContextResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_script_context", nil, nil, &result)
}

func (s *DefaultScriptService) GetLanguages(ctx context.Context) (*ScriptLanguagesResponse, error) {
	var result ScriptLanguagesResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_script_language", nil, nil, &result)
}
