package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// HealthReportParams are the query parameters for GET /_health_report.
type HealthReportParams struct {
	Common  *CommonParams
	Size    *int
	Time    string
	Timeout string
	Verbose *bool
}

// ToMap converts HealthReportParams to a query-parameter map.
func (p *HealthReportParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	if p.Time != "" {
		m["time"] = p.Time
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	setBoolParam(m, "verbose", p.Verbose)
	return mergeParams(m, p.Common)
}

// HealthReportResponse is the response from GET /_health_report.
type HealthReportResponse struct {
	ClusterName string                     `json:"cluster_name"`
	Status      string                     `json:"status"`
	Indicators  map[string]HealthIndicator `json:"indicators"`
}

// HealthIndicator describes a health indicator.
type HealthIndicator struct {
	Status      string `json:"status"`
	Summary     string `json:"summary"`
	Detail      string `json:"detail,omitempty"`
	UserActions []struct {
		Action  string `json:"action"`
		HelpURL string `json:"help_url,omitempty"`
	} `json:"user_actions,omitempty"`
	Impacts []struct {
		ID          string   `json:"id"`
		Severity    string   `json:"severity"`
		Description string   `json:"description"`
		ImpactAreas []string `json:"impact_areas"`
	} `json:"impacts,omitempty"`
	Diagnoses []struct {
		ID                string              `json:"id"`
		Cause             string              `json:"cause"`
		Action            string              `json:"action"`
		HelpURL           string              `json:"help_url,omitempty"`
		AffectedResources map[string][]string `json:"affected_resources,omitempty"`
	} `json:"diagnoses,omitempty"`
}

// HealthReportService provides access to the health report APIs.
type HealthReportService interface {
	Get(ctx context.Context, feature string, params *HealthReportParams) (*HealthReportResponse, error)
}

// DefaultHealthReportService is the default HealthReportService implementation.
type DefaultHealthReportService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewHealthReportService creates a new HealthReportService.
func NewHealthReportService(client *resty.Client, logger *logrus.Entry) HealthReportService {
	return &DefaultHealthReportService{client: client, logger: logger.WithField("service", "health_report")}
}

func (s *DefaultHealthReportService) Get(ctx context.Context, feature string, params *HealthReportParams) (*HealthReportResponse, error) {
	var path string
	if strings.TrimSpace(feature) != "" {
		path = fmt.Sprintf("/_health_report/%s", pathSegment(feature))
	} else {
		path = "/_health_report"
	}
	var result HealthReportResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}
