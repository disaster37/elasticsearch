package api

import (
	"context"
	"fmt"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	json "github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
)

// InferenceGetResponse is the response from GET /_inference.
type InferenceGetResponse struct {
	InferenceID     string         `json:"inference_id"`
	TaskType        string         `json:"task_type"`
	Service         string         `json:"service"`
	ServiceSettings map[string]any `json:"service_settings"`
	TaskSettings    map[string]any `json:"task_settings"`
}

// InferenceDeleteParams are the query parameters for delete.
type InferenceDeleteParams struct {
	Common *CommonParams
	Force  *bool
	DryRun *bool
}

// ToMap converts InferenceDeleteParams to a query-parameter map.
func (p *InferenceDeleteParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "force", p.Force)
	setBoolParam(m, "dry_run", p.DryRun)
	return mergeParams(m, p.Common)
}

// InferenceService provides access to the inference APIs.
type InferenceService interface {
	Get(ctx context.Context, inferenceId, taskType string) (json.RawMessage, error)
	Put(ctx context.Context, taskType, inferenceId string, body any) (*types.AcknowledgedResponse, error)
	Delete(ctx context.Context, taskType, inferenceId string, params *InferenceDeleteParams) (*types.AcknowledgedResponse, error)
	Update(ctx context.Context, inferenceId string, body any) (*types.AcknowledgedResponse, error)
	Inference(ctx context.Context, inferenceId string, body any) (json.RawMessage, error)
	Completion(ctx context.Context, inferenceId string, body any) (json.RawMessage, error)
	Rerank(ctx context.Context, inferenceId string, body any) (json.RawMessage, error)
}

// DefaultInferenceService is the default InferenceService implementation.
type DefaultInferenceService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewInferenceService creates a new InferenceService.
func NewInferenceService(client *resty.Client, logger *logrus.Entry) InferenceService {
	return &DefaultInferenceService{client: client, logger: logger.WithField("service", "inference")}
}

func (s *DefaultInferenceService) Get(ctx context.Context, inferenceId, taskType string) (json.RawMessage, error) {
	var path string
	switch {
	case taskType != "" && inferenceId != "":
		path = fmt.Sprintf("/_inference/%s/%s", pathSegment(taskType), pathSegment(inferenceId))
	case inferenceId != "":
		path = fmt.Sprintf("/_inference/%s", pathSegment(inferenceId))
	default:
		path = "/_inference"
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", path, nil, nil)
}

func (s *DefaultInferenceService) Put(ctx context.Context, taskType, inferenceId string, body any) (*types.AcknowledgedResponse, error) {
	if taskType == "" || inferenceId == "" {
		return nil, fmt.Errorf("task_type and inference_id are required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_inference/%s/%s", pathSegment(taskType), pathSegment(inferenceId)), body, nil, &result)
}

func (s *DefaultInferenceService) Delete(ctx context.Context, taskType, inferenceId string, params *InferenceDeleteParams) (*types.AcknowledgedResponse, error) {
	var path string
	if taskType != "" && inferenceId != "" {
		path = fmt.Sprintf("/_inference/%s/%s", pathSegment(taskType), pathSegment(inferenceId))
	} else if inferenceId != "" {
		path = fmt.Sprintf("/_inference/%s", pathSegment(inferenceId))
	} else {
		return nil, fmt.Errorf("inference_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", path, nil, params, &result)
}

func (s *DefaultInferenceService) Update(ctx context.Context, inferenceId string, body any) (*types.AcknowledgedResponse, error) {
	if inferenceId == "" {
		return nil, fmt.Errorf("inference_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_inference/%s/_update", pathSegment(inferenceId)), body, nil, &result)
}

func (s *DefaultInferenceService) Inference(ctx context.Context, inferenceId string, body any) (json.RawMessage, error) {
	if inferenceId == "" {
		return nil, fmt.Errorf("inference_id is required")
	}
	return doReqRaw(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_inference/%s", pathSegment(inferenceId)), body, nil)
}

func (s *DefaultInferenceService) Completion(ctx context.Context, inferenceId string, body any) (json.RawMessage, error) {
	if inferenceId == "" {
		return nil, fmt.Errorf("inference_id is required")
	}
	return doReqRaw(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_inference/completion/%s", pathSegment(inferenceId)), body, nil)
}

func (s *DefaultInferenceService) Rerank(ctx context.Context, inferenceId string, body any) (json.RawMessage, error) {
	if inferenceId == "" {
		return nil, fmt.Errorf("inference_id is required")
	}
	return doReqRaw(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_inference/rerank/%s", pathSegment(inferenceId)), body, nil)
}
