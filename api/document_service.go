package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// DocumentService provides access to the document APIs.
type DocumentService interface {
	Index(ctx context.Context, req *IndexRequest) (*IndexResponse, error)
	Create(ctx context.Context, req *CreateRequest) (*IndexResponse, error)
	Get(ctx context.Context, req *GetRequest) (*GetResult, error)
	GetSource(ctx context.Context, req *GetSourceRequest) (json.RawMessage, error)
	Exists(ctx context.Context, index, id string, params *GetParams) (bool, error)
	ExistsSource(ctx context.Context, index, id string, params *GetSourceParams) (bool, error)
	Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error)
	MultiGet(ctx context.Context, req *MgetRequest) (*MgetResponse, error)
	Update(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error)
	Bulk(ctx context.Context, req *BulkRequest) (*BulkResponse, error)
	DeleteByQuery(ctx context.Context, req *DeleteByQueryRequest) (*BulkIndexByScrollResponse, error)
	DeleteByQueryRethrottle(ctx context.Context, req *RethrottleRequest) (*TasksListResponse, error)
	UpdateByQuery(ctx context.Context, req *UpdateByQueryRequest) (*BulkIndexByScrollResponse, error)
	UpdateByQueryRethrottle(ctx context.Context, req *RethrottleRequest) (*TasksListResponse, error)
	Reindex(ctx context.Context, req *ReindexRequest) (*BulkIndexByScrollResponse, error)
	ReindexRethrottle(ctx context.Context, req *RethrottleRequest) (*TasksListResponse, error)
	Explain(ctx context.Context, req *ExplainRequest) (*ExplainResponse, error)
	TermVectors(ctx context.Context, req *TermVectorsRequest) (*TermvectorsResponse, error)
	MultiTermVectors(ctx context.Context, req *MtermvectorsRequest) (*MultiTermvectorResponse, error)
	TermsEnum(ctx context.Context, req *TermsEnumRequest) (*TermsEnumResponse, error)
}

// DefaultDocumentService is the default DocumentService implementation.
type DefaultDocumentService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewDocumentService creates a new DocumentService.
func NewDocumentService(client *resty.Client, logger *logrus.Entry) DocumentService {
	return &DefaultDocumentService{client: client, logger: logger.WithField("service", "document")}
}

func (s *DefaultDocumentService) Index(ctx context.Context, req *IndexRequest) (*IndexResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path, method string
	if req.Id != "" {
		path = fmt.Sprintf("/%s/_doc/%s", pathSegment(req.Index), pathSegment(req.Id))
		method = "PUT"
	} else {
		path = fmt.Sprintf("/%s/_doc", pathSegment(req.Index))
		method = "POST"
	}
	r := s.client.R().SetContext(ctx).SetBody(req.Body)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Execute(method, path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IndexResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	result.Status = resp.StatusCode()
	return &result, nil
}

func (s *DefaultDocumentService) Create(ctx context.Context, req *CreateRequest) (*IndexResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result IndexResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/%s/_create/%s", pathSegment(req.Index), pathSegment(req.Id)), req.Body, req.Params, &result)
}

func (s *DefaultDocumentService) Get(ctx context.Context, req *GetRequest) (*GetResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/%s/_doc/%s", pathSegment(req.Index), pathSegment(req.Id))
	r := s.client.R().SetContext(ctx)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.StatusCode() == 404 {
		return &GetResult{Found: false, Index: req.Index, Id: req.Id}, nil
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result GetResult
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultDocumentService) GetSource(ctx context.Context, req *GetSourceRequest) (json.RawMessage, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", fmt.Sprintf("/%s/_source/%s", pathSegment(req.Index), pathSegment(req.Id)), nil, req.Params)
}

func (s *DefaultDocumentService) Exists(ctx context.Context, index, id string, params *GetParams) (bool, error) {
	if index == "" {
		return false, fmt.Errorf("index is required")
	}
	if id == "" {
		return false, fmt.Errorf("id is required")
	}
	return headReq(s.logger, ctx, s.client, fmt.Sprintf("/%s/_doc/%s", pathSegment(index), pathSegment(id)), params)
}

func (s *DefaultDocumentService) ExistsSource(ctx context.Context, index, id string, params *GetSourceParams) (bool, error) {
	if index == "" {
		return false, fmt.Errorf("index is required")
	}
	if id == "" {
		return false, fmt.Errorf("id is required")
	}
	return headReq(s.logger, ctx, s.client, fmt.Sprintf("/%s/_source/%s", pathSegment(index), pathSegment(id)), params)
}

func (s *DefaultDocumentService) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/%s/_doc/%s", pathSegment(req.Index), pathSegment(req.Id))
	r := s.client.R().SetContext(ctx)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Delete(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() && resp.StatusCode() != 404 {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result DeleteResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	if resp.StatusCode() == 404 {
		return &result, fmt.Errorf("document not found")
	}
	return &result, nil
}

func (s *DefaultDocumentService) MultiGet(ctx context.Context, req *MgetRequest) (*MgetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if req.Index != "" {
		path = fmt.Sprintf("/%s/_mget", pathSegment(req.Index))
	} else {
		path = "/_mget"
	}
	var result MgetResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", path, req.Body, req.Params, &result)
}

func (s *DefaultDocumentService) Update(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result UpdateResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_update/%s", pathSegment(req.Index), pathSegment(req.Id)), req.Body, req.Params, &result)
}

func (s *DefaultDocumentService) Bulk(ctx context.Context, req *BulkRequest) (*BulkResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if req.Index != "" {
		path = fmt.Sprintf("/%s/_bulk", pathSegment(req.Index))
	} else {
		path = "/_bulk"
	}
	r := s.client.R().
		SetContext(ctx).
		SetBody(ensureTrailingNewline(req.Body)).
		SetHeader("Content-Type", "application/x-ndjson")
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result BulkResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultDocumentService) DeleteByQuery(ctx context.Context, req *DeleteByQueryRequest) (*BulkIndexByScrollResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result BulkIndexByScrollResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_delete_by_query", escapedList(req.Indices)), req.Body, req.Params, &result)
}

func (s *DefaultDocumentService) UpdateByQuery(ctx context.Context, req *UpdateByQueryRequest) (*BulkIndexByScrollResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result BulkIndexByScrollResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_update_by_query", escapedList(req.Indices)), req.Body, req.Params, &result)
}

func (s *DefaultDocumentService) Reindex(ctx context.Context, req *ReindexRequest) (*BulkIndexByScrollResponse, error) {
	var result BulkIndexByScrollResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_reindex", req.Body, req.Params, &result)
}

func (s *DefaultDocumentService) rethrottle(ctx context.Context, req *RethrottleRequest, prefix string) (*TasksListResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/%s/%s/_rethrottle", pathSegment(prefix), pathSegment(req.TaskId))
	rps := req.RequestsPerSecond
	if req.Params != nil && req.Params.RequestsPerSecond != nil {
		rps = *req.Params.RequestsPerSecond
	}
	r := s.client.R().SetContext(ctx).SetQueryParam("requests_per_second", floatStr(rps))
	for k, v := range req.Params.ToMap() {
		if k == "requests_per_second" {
			continue
		}
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

func (s *DefaultDocumentService) DeleteByQueryRethrottle(ctx context.Context, req *RethrottleRequest) (*TasksListResponse, error) {
	return s.rethrottle(ctx, req, "_delete_by_query")
}

func (s *DefaultDocumentService) UpdateByQueryRethrottle(ctx context.Context, req *RethrottleRequest) (*TasksListResponse, error) {
	return s.rethrottle(ctx, req, "_update_by_query")
}

func (s *DefaultDocumentService) ReindexRethrottle(ctx context.Context, req *RethrottleRequest) (*TasksListResponse, error) {
	return s.rethrottle(ctx, req, "_reindex")
}

func (s *DefaultDocumentService) Explain(ctx context.Context, req *ExplainRequest) (*ExplainResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/%s/_explain/%s", pathSegment(req.Index), pathSegment(req.Id))
	r := s.client.R().SetContext(ctx)
	if req.Body != nil {
		r = r.SetBody(req.Body)
	}
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	var resp *resty.Response
	var err error
	if req.Body != nil {
		resp, err = r.Post(path)
	} else {
		resp, err = r.Get(path)
	}
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result ExplainResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultDocumentService) TermVectors(ctx context.Context, req *TermVectorsRequest) (*TermvectorsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if req.Id != "" {
		path = fmt.Sprintf("/%s/_termvectors/%s", pathSegment(req.Index), pathSegment(req.Id))
	} else {
		path = fmt.Sprintf("/%s/_termvectors", pathSegment(req.Index))
	}
	r := s.client.R().SetContext(ctx)
	if req.Body != nil {
		r = r.SetBody(req.Body)
	}
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	var resp *resty.Response
	var err error
	if req.Body != nil {
		resp, err = r.Post(path)
	} else {
		resp, err = r.Get(path)
	}
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result TermvectorsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultDocumentService) MultiTermVectors(ctx context.Context, req *MtermvectorsRequest) (*MultiTermvectorResponse, error) {
	var params *MtermvectorsSpecParams
	if req.Params != nil {
		params = &MtermvectorsSpecParams{TermVectorsParams: *req.Params}
	}
	return s.MultiTermVectorsWithRequest(ctx, &MtermvectorsSpecRequest{Index: req.Index, Body: req.Body, Params: params})
}

func (s *DefaultDocumentService) TermsEnum(ctx context.Context, req *TermsEnumRequest) (*TermsEnumResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result TermsEnumResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_terms_enum", pathSegment(req.Index)), req.Body, req.Params, &result)
}
