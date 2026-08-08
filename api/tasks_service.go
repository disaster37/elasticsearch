package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// TasksService provides access to the tasks APIs.
type TasksService interface {
	List(ctx context.Context, params *TasksListParams) (*TasksListResponse, error)
	Get(ctx context.Context, taskId string, params *TasksGetParams) (*TaskGetResponse, error)
	Cancel(ctx context.Context, taskId string, params *TasksCancelParams) (*TasksListResponse, error)
}

// DefaultTasksService is the default TasksService implementation.
type DefaultTasksService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewTasksService creates a new TasksService.
func NewTasksService(client *resty.Client, logger *logrus.Entry) TasksService {
	return &DefaultTasksService{client: client, logger: logger.WithField("service", "tasks")}
}

func (s *DefaultTasksService) List(ctx context.Context, params *TasksListParams) (*TasksListResponse, error) {
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get("/_tasks")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result TasksListResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTasksService) Get(ctx context.Context, taskId string, params *TasksGetParams) (*TaskGetResponse, error) {
	if taskId == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(fmt.Sprintf("/_tasks/%s", pathSegment(taskId)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result TaskGetResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultTasksService) Cancel(ctx context.Context, taskId string, params *TasksCancelParams) (*TasksListResponse, error) {
	var path string
	if taskId != "" {
		path = fmt.Sprintf("/_tasks/%s/_cancel", pathSegment(taskId))
	} else {
		path = "/_tasks/_cancel"
	}
	r := s.client.R().SetContext(ctx)
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
	var result TasksListResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}
