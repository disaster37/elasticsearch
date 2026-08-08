package api

import (
	"context"
	"fmt"
	"strings"

	json "github.com/goccy/go-json"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// MigrationDeprecationsResponse is the response from deprecations.
type MigrationDeprecationsResponse struct {
	ClusterSettings []map[string]any            `json:"cluster_settings"`
	NodeSettings    []map[string]any            `json:"node_settings"`
	IndexSettings   map[string][]map[string]any `json:"index_settings"`
	MLSettings      []map[string]any            `json:"ml_settings"`
}

// MigrationFeatureUpgradeStatusResponse is the response from feature_upgrade status.
type MigrationFeatureUpgradeStatusResponse struct {
	Features []struct {
		FeatureName         string `json:"feature_name"`
		FeatureStatus       string `json:"feature_status"`
		MinimumIndexVersion int    `json:"minimum_index_version"`
		Indices             []struct {
			IndexName string `json:"index_name"`
		} `json:"indices"`
	} `json:"features"`
}

// MigrationService provides access to the migration APIs.
type MigrationService interface {
	Deprecations(ctx context.Context, index string) (*MigrationDeprecationsResponse, error)
	GetFeatureUpgradeStatus(ctx context.Context) (*MigrationFeatureUpgradeStatusResponse, error)
	PostFeatureUpgrade(ctx context.Context) (json.RawMessage, error)
}

// DefaultMigrationService is the default MigrationService implementation.
type DefaultMigrationService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewMigrationService creates a new MigrationService.
func NewMigrationService(client *resty.Client, logger *logrus.Entry) MigrationService {
	return &DefaultMigrationService{client: client, logger: logger.WithField("service", "migration")}
}

func (s *DefaultMigrationService) Deprecations(ctx context.Context, index string) (*MigrationDeprecationsResponse, error) {
	var path string
	if strings.TrimSpace(index) != "" {
		path = fmt.Sprintf("/%s/_migration/deprecations", pathSegment(index))
	} else {
		path = "/_migration/deprecations"
	}
	var result MigrationDeprecationsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultMigrationService) GetFeatureUpgradeStatus(ctx context.Context) (*MigrationFeatureUpgradeStatusResponse, error) {
	var result MigrationFeatureUpgradeStatusResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_migration/feature_upgrade", nil, nil, &result)
}

func (s *DefaultMigrationService) PostFeatureUpgrade(ctx context.Context) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "POST", "/_migration/feature_upgrade", nil, nil)
}
