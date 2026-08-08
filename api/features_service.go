package api

import (
	"context"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// FeaturesGetResponse is the response from GET /_features.
type FeaturesGetResponse struct {
	Features []struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
	} `json:"features"`
}

// FeaturesService provides access to the features APIs.
type FeaturesService interface {
	GetFeatures(ctx context.Context) (*FeaturesGetResponse, error)
	ResetFeatures(ctx context.Context) (*types.AcknowledgedResponse, error)
}

// DefaultFeaturesService is the default FeaturesService implementation.
type DefaultFeaturesService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewFeaturesService creates a new FeaturesService.
func NewFeaturesService(client *resty.Client, logger *logrus.Entry) FeaturesService {
	return &DefaultFeaturesService{client: client, logger: logger.WithField("service", "features")}
}

func (s *DefaultFeaturesService) GetFeatures(ctx context.Context) (*FeaturesGetResponse, error) {
	var result FeaturesGetResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_features", nil, nil, &result)
}

func (s *DefaultFeaturesService) ResetFeatures(ctx context.Context) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_features/_reset", nil, nil, &result)
}
