package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// TransformService provides access to the transform APIs.
type TransformService interface {
	Get(ctx context.Context, transformIds []string, params *TransformGetParams) (*TransformGetResponse, error)
	GetStats(ctx context.Context, transformIds []string, params *TransformGetStatsParams) (*TransformStatsResponse, error)
	Put(ctx context.Context, transformId string, body any, params *TransformPutParams) (*TransformPutResponse, error)
	Preview(ctx context.Context, transformId string, body any, params *TransformPreviewParams) (*TransformPreviewResponse, error)
	Delete(ctx context.Context, transformId string, params *TransformDeleteParams) (*types.AcknowledgedResponse, error)
	Start(ctx context.Context, transformId string, params *TransformStartParams) (*types.AcknowledgedResponse, error)
	Stop(ctx context.Context, transformId string, params *TransformStopParams) (*types.AcknowledgedResponse, error)
	Reset(ctx context.Context, transformId string, params *TransformResetParams) (*types.AcknowledgedResponse, error)
	ScheduleNow(ctx context.Context, transformId string) (*types.AcknowledgedResponse, error)
	Upgrade(ctx context.Context, params *TransformUpgradeParams) (*TransformUpgradeResponse, error)
	GetNodeStats(ctx context.Context) (*TransformNodeStatsResponse, error)
}

// DefaultTransformService is the default TransformService implementation.
type DefaultTransformService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewTransformService creates a new TransformService.
func NewTransformService(client *resty.Client, logger *logrus.Entry) TransformService {
	return &DefaultTransformService{client: client, logger: logger.WithField("service", "transform")}
}

func (s *DefaultTransformService) Get(ctx context.Context, transformIds []string, params *TransformGetParams) (*TransformGetResponse, error) {
	var path string
	if len(transformIds) > 0 {
		path = fmt.Sprintf("/_transform/%s", escapedList(transformIds))
	} else {
		path = "/_transform"
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result TransformGetResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTransformService) GetStats(ctx context.Context, transformIds []string, params *TransformGetStatsParams) (*TransformStatsResponse, error) {
	var path string
	if len(transformIds) > 0 {
		path = fmt.Sprintf("/_transform/%s/_stats", escapedList(transformIds))
	} else {
		path = "/_transform/_stats"
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result TransformStatsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTransformService) Put(ctx context.Context, transformId string, body any, params *TransformPutParams) (*TransformPutResponse, error) {
	if transformId == "" {
		return nil, fmt.Errorf("transform_id is required")
	}
	r := s.client.R().SetContext(ctx).SetBody(body)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Put(fmt.Sprintf("/_transform/%s", pathSegment(transformId)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result TransformPutResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTransformService) Preview(ctx context.Context, transformId string, body any, params *TransformPreviewParams) (*TransformPreviewResponse, error) {
	var path string
	r := s.client.R().SetContext(ctx)
	if body != nil {
		r = r.SetBody(body)
	}
	if transformId != "" {
		path = fmt.Sprintf("/_transform/%s/_preview", pathSegment(transformId))
	} else {
		path = "/_transform/_preview"
	}
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result TransformPreviewResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTransformService) Delete(ctx context.Context, transformId string, params *TransformDeleteParams) (*types.AcknowledgedResponse, error) {
	if transformId == "" {
		return nil, fmt.Errorf("transform_id is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Delete(fmt.Sprintf("/_transform/%s", pathSegment(transformId)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result types.AcknowledgedResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTransformService) Start(ctx context.Context, transformId string, params *TransformStartParams) (*types.AcknowledgedResponse, error) {
	if transformId == "" {
		return nil, fmt.Errorf("transform_id is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post(fmt.Sprintf("/_transform/%s/_start", pathSegment(transformId)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result types.AcknowledgedResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTransformService) Stop(ctx context.Context, transformId string, params *TransformStopParams) (*types.AcknowledgedResponse, error) {
	if transformId == "" {
		return nil, fmt.Errorf("transform_id is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post(fmt.Sprintf("/_transform/%s/_stop", pathSegment(transformId)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result types.AcknowledgedResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTransformService) Reset(ctx context.Context, transformId string, params *TransformResetParams) (*types.AcknowledgedResponse, error) {
	if transformId == "" {
		return nil, fmt.Errorf("transform_id is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post(fmt.Sprintf("/_transform/%s/_reset", pathSegment(transformId)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result types.AcknowledgedResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTransformService) ScheduleNow(ctx context.Context, transformId string) (*types.AcknowledgedResponse, error) {
	if transformId == "" {
		return nil, fmt.Errorf("transform_id is required")
	}
	resp, err := s.client.R().SetContext(ctx).Post(fmt.Sprintf("/_transform/%s/_schedule_now", pathSegment(transformId)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result types.AcknowledgedResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTransformService) Upgrade(ctx context.Context, params *TransformUpgradeParams) (*TransformUpgradeResponse, error) {
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post("/_transform/_upgrade")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result TransformUpgradeResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTransformService) GetNodeStats(ctx context.Context) (*TransformNodeStatsResponse, error) {
	resp, err := s.client.R().SetContext(ctx).Get("/_transform/_node_stats")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result TransformNodeStatsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}
