package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// SqlQueryParams are the query parameters for POST /_sql.
type SqlQueryParams struct {
	Common *CommonParams
	Format string
}

// ToMap converts SqlQueryParams to a query-parameter map.
func (p *SqlQueryParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Format != "" {
		m["format"] = p.Format
	}
	return mergeParams(m, p.Common)
}

// SqlGetAsyncParams are the query parameters for GET /_sql/async/{id}.
type SqlGetAsyncParams struct {
	Common                   *CommonParams
	Delimiter                string
	Format                   string
	KeepAlive                string
	WaitForCompletionTimeout string
}

// ToMap converts SqlGetAsyncParams to a query-parameter map.
func (p *SqlGetAsyncParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Delimiter != "" {
		m["delimiter"] = p.Delimiter
	}
	if p.Format != "" {
		m["format"] = p.Format
	}
	if p.KeepAlive != "" {
		m["keep_alive"] = p.KeepAlive
	}
	if p.WaitForCompletionTimeout != "" {
		m["wait_for_completion_timeout"] = p.WaitForCompletionTimeout
	}
	return mergeParams(m, p.Common)
}

// SqlQueryResponse is the response from POST /_sql.
type SqlQueryResponse struct {
	Columns []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"columns,omitempty"`
	Rows      [][]any `json:"rows,omitempty"`
	Cursor    string  `json:"cursor,omitempty"`
	IsRunning bool    `json:"is_running,omitempty"`
}

// SqlTranslateResponse is the response from POST /_sql/translate.
type SqlTranslateResponse struct {
	Size  int              `json:"size,omitempty"`
	Query map[string]any   `json:"query,omitempty"`
	Sort  []map[string]any `json:"sort,omitempty"`
}

// SqlClearCursorResponse is the response from POST /_sql/close.
type SqlClearCursorResponse struct {
	Succeeded bool `json:"succeeded"`
}

// SqlAsyncStatusResponse is the response from GET /_sql/async/status/{id}.
type SqlAsyncStatusResponse struct {
	ID                 string `json:"id"`
	IsRunning          bool   `json:"is_running"`
	IsPartial          bool   `json:"is_partial"`
	StartInMillis      int64  `json:"start_in_millis,omitempty"`
	ExpirationInMillis int64  `json:"expiration_in_millis,omitempty"`
	CompletionStatus   int    `json:"completion_status,omitempty"`
}

// SqlService provides access to the SQL APIs.
type SqlService interface {
	Query(ctx context.Context, body any, params *SqlQueryParams) (*SqlQueryResponse, error)
	Translate(ctx context.Context, body any) (*SqlTranslateResponse, error)
	ClearCursor(ctx context.Context, body any) (*SqlClearCursorResponse, error)
	GetAsync(ctx context.Context, id string, params *SqlGetAsyncParams) (*SqlQueryResponse, error)
	GetAsyncStatus(ctx context.Context, id string) (*SqlAsyncStatusResponse, error)
	DeleteAsync(ctx context.Context, id string) (*types.AcknowledgedResponse, error)
}

// DefaultSqlService is the default SqlService implementation.
type DefaultSqlService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewSqlService creates a new SqlService.
func NewSqlService(client *resty.Client, logger *logrus.Entry) SqlService {
	return &DefaultSqlService{client: client, logger: logger.WithField("service", "sql")}
}

func (s *DefaultSqlService) Query(ctx context.Context, body any, params *SqlQueryParams) (*SqlQueryResponse, error) {
	var result SqlQueryResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_sql", body, params, &result)
}

func (s *DefaultSqlService) Translate(ctx context.Context, body any) (*SqlTranslateResponse, error) {
	var result SqlTranslateResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_sql/translate", body, nil, &result)
}

func (s *DefaultSqlService) ClearCursor(ctx context.Context, body any) (*SqlClearCursorResponse, error) {
	var result SqlClearCursorResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_sql/close", body, nil, &result)
}

func (s *DefaultSqlService) GetAsync(ctx context.Context, id string, params *SqlGetAsyncParams) (*SqlQueryResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result SqlQueryResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_sql/async/%s", pathSegment(id)), nil, params, &result)
}

func (s *DefaultSqlService) GetAsyncStatus(ctx context.Context, id string) (*SqlAsyncStatusResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result SqlAsyncStatusResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_sql/async/status/%s", pathSegment(id)), nil, nil, &result)
}

func (s *DefaultSqlService) DeleteAsync(ctx context.Context, id string) (*types.AcknowledgedResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_sql/async/%s", pathSegment(id)), nil, nil, &result)
}

// ensure json import is used.
var _ = json.RawMessage{}
