package api

import (
	"context"
	"fmt"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// WatcherPutWatchParams are the query parameters for put watch.
type WatcherPutWatchParams struct {
	Common        *CommonParams
	Active        *bool
	IfPrimaryTerm *int64
	IfSeqNo       *int64
}

// ToMap converts WatcherPutWatchParams to a query-parameter map.
func (p *WatcherPutWatchParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "active", p.Active)
	if p.IfPrimaryTerm != nil {
		m["if_primary_term"] = int64Str(*p.IfPrimaryTerm)
	}
	if p.IfSeqNo != nil {
		m["if_seq_no"] = int64Str(*p.IfSeqNo)
	}
	return mergeParams(m, p.Common)
}

// WatcherExecuteWatchParams are the query parameters for execute watch.
type WatcherExecuteWatchParams struct {
	Common *CommonParams
	Debug  *bool
}

// ToMap converts WatcherExecuteWatchParams to a query-parameter map.
func (p *WatcherExecuteWatchParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "debug", p.Debug)
	return mergeParams(m, p.Common)
}

// WatcherStatsParams are the query parameters for watcher stats.
type WatcherStatsParams struct {
	Common          *CommonParams
	EmitStacktraces *bool
}

// ToMap converts WatcherStatsParams to a query-parameter map.
func (p *WatcherStatsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "emit_stacktraces", p.EmitStacktraces)
	return mergeParams(m, p.Common)
}

// WatcherPutWatchResponse is the response from put watch.
type WatcherPutWatchResponse struct {
	ID      string `json:"_id"`
	Version int64  `json:"_version"`
	Created bool   `json:"created"`
}

// WatcherGetWatchResponse is the response from get watch.
type WatcherGetWatchResponse struct {
	Found  bool           `json:"found"`
	ID     string         `json:"_id,omitempty"`
	Status map[string]any `json:"status,omitempty"`
	Watch  map[string]any `json:"watch,omitempty"`
}

// WatcherDeleteWatchResponse is the response from delete watch.
type WatcherDeleteWatchResponse struct {
	Found   bool   `json:"found"`
	ID      string `json:"_id,omitempty"`
	Version int64  `json:"_version,omitempty"`
}

// WatcherExecuteWatchResponse is the response from execute watch.
type WatcherExecuteWatchResponse struct {
	ID          string         `json:"id"`
	WatchRecord map[string]any `json:"watch_record"`
}

// WatcherAckWatchResponse is the response from ack watch.
type WatcherAckWatchResponse struct {
	Status map[string]any `json:"status"`
}

// WatcherActivateWatchResponse is the response from activate/deactivate watch.
type WatcherActivateWatchResponse struct {
	Status map[string]any `json:"status"`
}

// WatcherQueryWatchesResponse is the response from query watches.
type WatcherQueryWatchesResponse struct {
	Count   int `json:"count"`
	Watches []struct {
		ID     string         `json:"id"`
		Watch  map[string]any `json:"watch,omitempty"`
		Status map[string]any `json:"status,omitempty"`
	} `json:"watches"`
}

// WatcherStatsResponse is the response from watcher stats.
type WatcherStatsResponse struct {
	WatcherState   string           `json:"watcher_state"`
	Stats          map[string]any   `json:"stats,omitempty"`
	Queues         map[string]any   `json:"queues,omitempty"`
	CurrentWatches []map[string]any `json:"current_watches,omitempty"`
}

// WatcherService provides access to the watcher APIs.
type WatcherService interface {
	PutWatch(ctx context.Context, id string, body any, params *WatcherPutWatchParams) (*WatcherPutWatchResponse, error)
	GetWatch(ctx context.Context, id string) (*WatcherGetWatchResponse, error)
	DeleteWatch(ctx context.Context, id string) (*WatcherDeleteWatchResponse, error)
	ExecuteWatch(ctx context.Context, id string, body any, params *WatcherExecuteWatchParams) (*WatcherExecuteWatchResponse, error)
	AckWatch(ctx context.Context, id, actionId string) (*WatcherAckWatchResponse, error)
	ActivateWatch(ctx context.Context, id string) (*WatcherActivateWatchResponse, error)
	DeactivateWatch(ctx context.Context, id string) (*WatcherActivateWatchResponse, error)
	QueryWatches(ctx context.Context, body any) (*WatcherQueryWatchesResponse, error)
	Start(ctx context.Context) (*types.AcknowledgedResponse, error)
	Stop(ctx context.Context) (*types.AcknowledgedResponse, error)
	Stats(ctx context.Context, metrics []string, params *WatcherStatsParams) (*WatcherStatsResponse, error)
}

// DefaultWatcherService is the default WatcherService implementation.
type DefaultWatcherService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewWatcherService creates a new WatcherService.
func NewWatcherService(client *resty.Client, logger *logrus.Entry) WatcherService {
	return &DefaultWatcherService{client: client, logger: logger.WithField("service", "watcher")}
}

func (s *DefaultWatcherService) PutWatch(ctx context.Context, id string, body any, params *WatcherPutWatchParams) (*WatcherPutWatchResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result WatcherPutWatchResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_watcher/watch/%s", pathSegment(id)), body, params, &result)
}

func (s *DefaultWatcherService) GetWatch(ctx context.Context, id string) (*WatcherGetWatchResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result WatcherGetWatchResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_watcher/watch/%s", pathSegment(id)), nil, nil, &result)
}

func (s *DefaultWatcherService) DeleteWatch(ctx context.Context, id string) (*WatcherDeleteWatchResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result WatcherDeleteWatchResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_watcher/watch/%s", pathSegment(id)), nil, nil, &result)
}

func (s *DefaultWatcherService) ExecuteWatch(ctx context.Context, id string, body any, params *WatcherExecuteWatchParams) (*WatcherExecuteWatchResponse, error) {
	var path string
	if id != "" {
		path = fmt.Sprintf("/_watcher/watch/%s/_execute", pathSegment(id))
	} else {
		path = "/_watcher/watch/_execute"
	}
	var result WatcherExecuteWatchResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", path, body, params, &result)
}

func (s *DefaultWatcherService) AckWatch(ctx context.Context, id, actionId string) (*WatcherAckWatchResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var path string
	if actionId != "" {
		path = fmt.Sprintf("/_watcher/watch/%s/_ack/%s", pathSegment(id), pathSegment(actionId))
	} else {
		path = fmt.Sprintf("/_watcher/watch/%s/_ack", pathSegment(id))
	}
	var result WatcherAckWatchResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", path, nil, nil, &result)
}

func (s *DefaultWatcherService) ActivateWatch(ctx context.Context, id string) (*WatcherActivateWatchResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result WatcherActivateWatchResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_watcher/watch/%s/_activate", pathSegment(id)), nil, nil, &result)
}

func (s *DefaultWatcherService) DeactivateWatch(ctx context.Context, id string) (*WatcherActivateWatchResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result WatcherActivateWatchResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_watcher/watch/%s/_deactivate", pathSegment(id)), nil, nil, &result)
}

func (s *DefaultWatcherService) QueryWatches(ctx context.Context, body any) (*WatcherQueryWatchesResponse, error) {
	var result WatcherQueryWatchesResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_watcher/_query/watches", body, nil, &result)
}

func (s *DefaultWatcherService) Start(ctx context.Context) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_watcher/_start", nil, nil, &result)
}

func (s *DefaultWatcherService) Stop(ctx context.Context) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_watcher/_stop", nil, nil, &result)
}

func (s *DefaultWatcherService) Stats(ctx context.Context, metrics []string, params *WatcherStatsParams) (*WatcherStatsResponse, error) {
	var path string
	if len(metrics) > 0 {
		path = fmt.Sprintf("/_watcher/stats/%s", escapedList(metrics))
	} else {
		path = "/_watcher/stats"
	}
	var result WatcherStatsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}
