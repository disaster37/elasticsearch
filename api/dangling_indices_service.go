package api

import (
	"context"
	"fmt"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// DanglingIndicesListResponse is the response from GET /_dangling.
type DanglingIndicesListResponse struct {
	DanglingIndices []struct {
		IndexUUID    string `json:"index_uuid"`
		IndexName    string `json:"index_name"`
		CreationDate int64  `json:"creation_date,omitempty"`
	} `json:"dangling_indices"`
}

// DanglingIndicesImportParams are the query parameters for import.
type DanglingIndicesImportParams struct {
	Common         *CommonParams
	AcceptDataLoss *bool `validate:"required"`
}

// ToMap converts DanglingIndicesImportParams to a query-parameter map.
func (p *DanglingIndicesImportParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "accept_data_loss", p.AcceptDataLoss)
	return mergeParams(m, p.Common)
}

// DanglingIndicesService provides access to the dangling indices APIs.
type DanglingIndicesService interface {
	List(ctx context.Context) (*DanglingIndicesListResponse, error)
	Import(ctx context.Context, indexUUID string, params *DanglingIndicesImportParams) (*types.AcknowledgedResponse, error)
	Delete(ctx context.Context, indexUUID string, params *DanglingIndicesImportParams) (*types.AcknowledgedResponse, error)
}

// DefaultDanglingIndicesService is the default DanglingIndicesService implementation.
type DefaultDanglingIndicesService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewDanglingIndicesService creates a new DanglingIndicesService.
func NewDanglingIndicesService(client *resty.Client, logger *logrus.Entry) DanglingIndicesService {
	return &DefaultDanglingIndicesService{client: client, logger: logger.WithField("service", "dangling_indices")}
}

func (s *DefaultDanglingIndicesService) List(ctx context.Context) (*DanglingIndicesListResponse, error) {
	var result DanglingIndicesListResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_dangling", nil, nil, &result)
}

func (s *DefaultDanglingIndicesService) Import(ctx context.Context, indexUUID string, params *DanglingIndicesImportParams) (*types.AcknowledgedResponse, error) {
	if indexUUID == "" {
		return nil, fmt.Errorf("index_uuid is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_dangling/%s", pathSegment(indexUUID)), nil, params, &result)
}

func (s *DefaultDanglingIndicesService) Delete(ctx context.Context, indexUUID string, params *DanglingIndicesImportParams) (*types.AcknowledgedResponse, error) {
	if indexUUID == "" {
		return nil, fmt.Errorf("index_uuid is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_dangling/%s", pathSegment(indexUUID)), nil, params, &result)
}
