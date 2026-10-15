package api

import (
	"context"
	"fmt"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// SynonymsListParams are the query parameters for list synonyms sets.
type SynonymsListParams struct {
	Common *CommonParams
	From   *int
	Size   *int
}

// ToMap converts SynonymsListParams to a query-parameter map.
func (p *SynonymsListParams) ToMap() map[string]string {
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

// SynonymsGetParams are the query parameters for get synonym.
type SynonymsGetParams struct {
	Common *CommonParams
	From   *int
	Size   *int
}

// ToMap converts SynonymsGetParams to a query-parameter map.
func (p *SynonymsGetParams) ToMap() map[string]string {
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

// SynonymsSetsResponse is the response from GET /_synonyms.
type SynonymsSetsResponse struct {
	Count   int              `json:"count"`
	Results []SynonymSetItem `json:"results"`
}

// SynonymSetItem describes a synonym set.
type SynonymSetItem struct {
	SynonymSet string `json:"synonym_set"`
	Count      int    `json:"count"`
}

// SynonymRulesResponse is the response from GET /_synonyms/{id}.
type SynonymRulesResponse struct {
	Count int           `json:"count"`
	Rules []SynonymRule `json:"rules"`
}

// SynonymRule describes a single synonym rule.
type SynonymRule struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Synonyms string `json:"synonyms"`
}

// SynonymsService provides access to the synonyms APIs.
type SynonymsService interface {
	GetSynonymsSets(ctx context.Context, params *SynonymsListParams) (*SynonymsSetsResponse, error)
	GetSynonym(ctx context.Context, id string, params *SynonymsGetParams) (*SynonymRulesResponse, error)
	PutSynonym(ctx context.Context, id string, body any) (*types.AcknowledgedResponse, error)
	DeleteSynonym(ctx context.Context, id string) (*types.AcknowledgedResponse, error)
	GetSynonymRule(ctx context.Context, id, ruleId string) (*SynonymRule, error)
	PutSynonymRule(ctx context.Context, id, ruleId string, body any) (*types.AcknowledgedResponse, error)
	DeleteSynonymRule(ctx context.Context, id, ruleId string) (*types.AcknowledgedResponse, error)
}

// DefaultSynonymsService is the default SynonymsService implementation.
type DefaultSynonymsService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewSynonymsService creates a new SynonymsService.
func NewSynonymsService(client *resty.Client, logger *logrus.Entry) SynonymsService {
	return &DefaultSynonymsService{client: client, logger: logger.WithField("service", "synonyms")}
}

func (s *DefaultSynonymsService) GetSynonymsSets(ctx context.Context, params *SynonymsListParams) (*SynonymsSetsResponse, error) {
	var result SynonymsSetsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_synonyms", nil, params, &result)
}

func (s *DefaultSynonymsService) GetSynonym(ctx context.Context, id string, params *SynonymsGetParams) (*SynonymRulesResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result SynonymRulesResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_synonyms/%s", pathSegment(id)), nil, params, &result)
}

func (s *DefaultSynonymsService) PutSynonym(ctx context.Context, id string, body any) (*types.AcknowledgedResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_synonyms/%s", pathSegment(id)), body, nil, &result)
}

func (s *DefaultSynonymsService) DeleteSynonym(ctx context.Context, id string) (*types.AcknowledgedResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_synonyms/%s", pathSegment(id)), nil, nil, &result)
}

func (s *DefaultSynonymsService) GetSynonymRule(ctx context.Context, id, ruleId string) (*SynonymRule, error) {
	if id == "" || ruleId == "" {
		return nil, fmt.Errorf("id and rule_id are required")
	}
	var result SynonymRule
	return &result, doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_synonyms/%s/%s", pathSegment(id), pathSegment(ruleId)), nil, nil, &result)
}

func (s *DefaultSynonymsService) PutSynonymRule(ctx context.Context, id, ruleId string, body any) (*types.AcknowledgedResponse, error) {
	if id == "" || ruleId == "" {
		return nil, fmt.Errorf("id and rule_id are required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_synonyms/%s/%s", pathSegment(id), pathSegment(ruleId)), body, nil, &result)
}

func (s *DefaultSynonymsService) DeleteSynonymRule(ctx context.Context, id, ruleId string) (*types.AcknowledgedResponse, error) {
	if id == "" || ruleId == "" {
		return nil, fmt.Errorf("id and rule_id are required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_synonyms/%s/%s", pathSegment(id), pathSegment(ruleId)), nil, nil, &result)
}
