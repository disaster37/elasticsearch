package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// LicenseGetParams are the query parameters for GET /_license.
type LicenseGetParams struct {
	Common           *CommonParams
	AcceptEnterprise *bool
	Local            *bool
}

// ToMap converts LicenseGetParams to a query-parameter map.
func (p *LicenseGetParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "accept_enterprise", p.AcceptEnterprise)
	setBoolParam(m, "local", p.Local)
	return mergeParams(m, p.Common)
}

// LicensePostParams are the query parameters for PUT /_license.
type LicensePostParams struct {
	Common        *CommonParams
	Acknowledge   *bool
	MasterTimeout string
	Timeout       string
}

// ToMap converts LicensePostParams to a query-parameter map.
func (p *LicensePostParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "acknowledge", p.Acknowledge)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// LicenseStartTrialParams are the query parameters for start trial.
type LicenseStartTrialParams struct {
	Common        *CommonParams
	Acknowledge   *bool
	MasterTimeout string
	Type          string
}

// ToMap converts LicenseStartTrialParams to a query-parameter map.
func (p *LicenseStartTrialParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "acknowledge", p.Acknowledge)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Type != "" {
		m["type"] = p.Type
	}
	return mergeParams(m, p.Common)
}

// LicenseGetResponse is the response from GET /_license.
type LicenseGetResponse struct {
	License LicenseInfo `json:"license"`
}

// LicenseInfo describes a license.
type LicenseInfo struct {
	UID                string `json:"uid"`
	Type               string `json:"type"`
	Status             string `json:"status"`
	IssueDate          string `json:"issue_date,omitempty"`
	IssueDateInMillis  int64  `json:"issue_dateInMillis,omitempty"`
	ExpiryDate         string `json:"expiry_date,omitempty"`
	ExpiryDateInMillis int64  `json:"expiry_dateInMillis,omitempty"`
	MaxNodes           int    `json:"max_nodes,omitempty"`
	IssuedTo           string `json:"issued_to,omitempty"`
	Issuer             string `json:"issuer,omitempty"`
	StartDate          string `json:"start_date,omitempty"`
	StartDateInMillis  int64  `json:"start_dateInMillis,omitempty"`
	Signature          string `json:"signature,omitempty"`
}

// LicenseService provides access to the license APIs.
type LicenseService interface {
	Get(ctx context.Context, params *LicenseGetParams) (*LicenseGetResponse, error)
	Delete(ctx context.Context) (*types.AcknowledgedResponse, error)
	Post(ctx context.Context, body any, params *LicensePostParams) (*types.AcknowledgedResponse, error)
	GetBasicStatus(ctx context.Context) (json.RawMessage, error)
	PostStartBasic(ctx context.Context, params *LicensePostParams) (json.RawMessage, error)
	GetTrialStatus(ctx context.Context) (json.RawMessage, error)
	PostStartTrial(ctx context.Context, params *LicenseStartTrialParams) (json.RawMessage, error)
}

// DefaultLicenseService is the default LicenseService implementation.
type DefaultLicenseService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewLicenseService creates a new LicenseService.
func NewLicenseService(client *resty.Client, logger *logrus.Entry) LicenseService {
	return &DefaultLicenseService{client: client, logger: logger.WithField("service", "license")}
}

func (s *DefaultLicenseService) Get(ctx context.Context, params *LicenseGetParams) (*LicenseGetResponse, error) {
	var result LicenseGetResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_license", nil, params, &result)
}

func (s *DefaultLicenseService) Delete(ctx context.Context) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", "/_license", nil, nil, &result)
}

func (s *DefaultLicenseService) Post(ctx context.Context, body any, params *LicensePostParams) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", "/_license", body, params, &result)
}

func (s *DefaultLicenseService) GetBasicStatus(ctx context.Context) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "GET", "/_license/basic_status", nil, nil)
}

func (s *DefaultLicenseService) PostStartBasic(ctx context.Context, params *LicensePostParams) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "POST", "/_license/start_basic", nil, params)
}

func (s *DefaultLicenseService) GetTrialStatus(ctx context.Context) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "GET", "/_license/trial_status", nil, nil)
}

func (s *DefaultLicenseService) PostStartTrial(ctx context.Context, params *LicenseStartTrialParams) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "POST", "/_license/start_trial", nil, params)
}

// ensure fmt used.
var _ = fmt.Sprintf
