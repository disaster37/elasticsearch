package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// QueryRulesListParams are the query parameters for list rulesets.
type QueryRulesListParams struct {
	Common *CommonParams
	From   *int
	Size   *int
}

// ToMap converts QueryRulesListParams to a query-parameter map.
func (p *QueryRulesListParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	return mergeParams(m, p.Common)
}

// QueryRulesTestParams are the query parameters for test ruleset.
type QueryRulesTestParams struct {
	Common      *CommonParams
	QueryText   string
	QueryParams []string
}

// ToMap converts QueryRulesTestParams to a query-parameter map.
func (p *QueryRulesTestParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.QueryText != "" {
		m["query_text"] = p.QueryText
	}
	setListParam(m, "query_params", p.QueryParams)
	return mergeParams(m, p.Common)
}

// QueryRulesService provides access to the query rules APIs.
type QueryRulesService interface {
	ListRulesets(ctx context.Context, params *QueryRulesListParams) (json.RawMessage, error)
	GetRuleset(ctx context.Context, rulesetId string) (json.RawMessage, error)
	PutRuleset(ctx context.Context, rulesetId string, body any) (*types.AcknowledgedResponse, error)
	DeleteRuleset(ctx context.Context, rulesetId string) (*types.AcknowledgedResponse, error)
	Test(ctx context.Context, rulesetId string, body any, params *QueryRulesTestParams) (json.RawMessage, error)
}

// DefaultQueryRulesService is the default QueryRulesService implementation.
type DefaultQueryRulesService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewQueryRulesService creates a new QueryRulesService.
func NewQueryRulesService(client *resty.Client, logger *logrus.Entry) QueryRulesService {
	return &DefaultQueryRulesService{client: client, logger: logger.WithField("service", "query_rules")}
}

func (s *DefaultQueryRulesService) ListRulesets(ctx context.Context, params *QueryRulesListParams) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "GET", "/_query_ruleset", nil, params)
}

func (s *DefaultQueryRulesService) GetRuleset(ctx context.Context, rulesetId string) (json.RawMessage, error) {
	if rulesetId == "" {
		return nil, fmt.Errorf("ruleset_id is required")
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_query_ruleset/%s", pathSegment(rulesetId)), nil, nil)
}

func (s *DefaultQueryRulesService) PutRuleset(ctx context.Context, rulesetId string, body any) (*types.AcknowledgedResponse, error) {
	if rulesetId == "" {
		return nil, fmt.Errorf("ruleset_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_query_ruleset/%s", pathSegment(rulesetId)), body, nil, &result)
}

func (s *DefaultQueryRulesService) DeleteRuleset(ctx context.Context, rulesetId string) (*types.AcknowledgedResponse, error) {
	if rulesetId == "" {
		return nil, fmt.Errorf("ruleset_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_query_ruleset/%s", pathSegment(rulesetId)), nil, nil, &result)
}

func (s *DefaultQueryRulesService) Test(ctx context.Context, rulesetId string, body any, params *QueryRulesTestParams) (json.RawMessage, error) {
	if rulesetId == "" {
		return nil, fmt.Errorf("ruleset_id is required")
	}
	return doReqRaw(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_query_ruleset/%s/_test", pathSegment(rulesetId)), body, params)
}
