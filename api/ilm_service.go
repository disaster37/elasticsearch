package api

import (
	"context"
	"fmt"
	"strconv"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// IlmService provides access to the index lifecycle management APIs.
type IlmService interface {
	GetLifecycle(ctx context.Context, policies []string) (map[string]*IlmPolicy, error)
	PutLifecycle(ctx context.Context, policy string, body any) (*types.AcknowledgedResponse, error)
	DeleteLifecycle(ctx context.Context, policy string) (*types.AcknowledgedResponse, error)
	GetStatus(ctx context.Context) (*IlmStatusResponse, error)
	Start(ctx context.Context) (*types.AcknowledgedResponse, error)
	Stop(ctx context.Context) (*types.AcknowledgedResponse, error)
	ExplainLifecycle(ctx context.Context, indices []string, params *IlmExplainParams) (*IlmExplainResponse, error)
	MoveToStep(ctx context.Context, index string, body any) (*types.AcknowledgedResponse, error)
	RemovePolicy(ctx context.Context, indices []string) (*IlmRemovePolicyResponse, error)
	Retry(ctx context.Context, indices []string) (*types.AcknowledgedResponse, error)
	MigrateToDataTiers(ctx context.Context, dryRun *bool) (*IlmMigrateToDataTiersResponse, error)
}

// DefaultIlmService is the default IlmService implementation.
type DefaultIlmService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewIlmService creates a new IlmService.
func NewIlmService(client *resty.Client, logger *logrus.Entry) IlmService {
	return &DefaultIlmService{client: client, logger: logger.WithField("service", "ilm")}
}

func (s *DefaultIlmService) GetLifecycle(ctx context.Context, policies []string) (map[string]*IlmPolicy, error) {
	var path string
	if len(policies) > 0 {
		path = fmt.Sprintf("/_ilm/policy/%s", escapedList(policies))
	} else {
		path = "/_ilm/policy"
	}
	var result map[string]*IlmPolicy
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultIlmService) PutLifecycle(ctx context.Context, policy string, body any) (*types.AcknowledgedResponse, error) {
	if policy == "" {
		return nil, fmt.Errorf("policy is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_ilm/policy/%s", pathSegment(policy)), body, nil, &result)
}

func (s *DefaultIlmService) DeleteLifecycle(ctx context.Context, policy string) (*types.AcknowledgedResponse, error) {
	if policy == "" {
		return nil, fmt.Errorf("policy is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_ilm/policy/%s", pathSegment(policy)), nil, nil, &result)
}

func (s *DefaultIlmService) GetStatus(ctx context.Context) (*IlmStatusResponse, error) {
	var result IlmStatusResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_ilm/status", nil, nil, &result)
}

func (s *DefaultIlmService) Start(ctx context.Context) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_ilm/start", nil, nil, &result)
}

func (s *DefaultIlmService) Stop(ctx context.Context) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_ilm/stop", nil, nil, &result)
}

func (s *DefaultIlmService) ExplainLifecycle(ctx context.Context, indices []string, params *IlmExplainParams) (*IlmExplainResponse, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("indices is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(fmt.Sprintf("/%s/_ilm/explain", escapedList(indices)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IlmExplainResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIlmService) MoveToStep(ctx context.Context, index string, body any) (*types.AcknowledgedResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ilm/move/%s", pathSegment(index)), body, nil, &result)
}

func (s *DefaultIlmService) RemovePolicy(ctx context.Context, indices []string) (*IlmRemovePolicyResponse, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("indices is required")
	}
	var result IlmRemovePolicyResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ilm/remove", escapedList(indices)), nil, nil, &result)
}

func (s *DefaultIlmService) Retry(ctx context.Context, indices []string) (*types.AcknowledgedResponse, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("indices is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_ilm/retry", escapedList(indices)), nil, nil, &result)
}

// MigrateToDataTiers migrates ILM policies to data tiers.
//
// Deprecated: use the data tiers APIs directly.
func (s *DefaultIlmService) MigrateToDataTiers(ctx context.Context, dryRun *bool) (*IlmMigrateToDataTiersResponse, error) {
	r := s.client.R().SetContext(ctx)
	if dryRun != nil {
		r.SetQueryParam("dry_run", strconv.FormatBool(*dryRun))
	}
	resp, err := r.Post("/_ilm/migrate_to_data_tiers")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IlmMigrateToDataTiersResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}
