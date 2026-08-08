package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/querydsl"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// SearchService provides access to the search APIs.
type SearchService interface {
	Search(ctx context.Context, req *SearchRequest) (*querydsl.SearchResult, error)
	SearchMVT(ctx context.Context, req *SearchMVTRequest) (json.RawMessage, error)
	MultiSearch(ctx context.Context, req *MultiSearchRequest) (*querydsl.MultiSearchResult, error)
	SearchTemplate(ctx context.Context, req *SearchTemplateRequest) (*querydsl.SearchResult, error)
	MultiSearchTemplate(ctx context.Context, req *MultiSearchTemplateRequest) (*querydsl.MultiSearchResult, error)
	RenderSearchTemplate(ctx context.Context, req *RenderSearchTemplateRequest) (*querydsl.RenderSearchTemplateResponse, error)
	RankEval(ctx context.Context, req *RankEvalRequest) (*querydsl.RankEvalResponse, error)
	Count(ctx context.Context, req *CountRequest) (int64, error)
	Scroll(ctx context.Context, req *ScrollRequest) (*querydsl.SearchResult, error)
	ClearScroll(ctx context.Context, scrollIds []string) (*querydsl.ClearScrollResponse, error)
	Validate(ctx context.Context, req *ValidateQueryRequest) (*querydsl.ValidateResponse, error)
	SearchShards(ctx context.Context, req *SearchShardsRequest) (*querydsl.SearchShardsResponse, error)
	FieldCaps(ctx context.Context, req *FieldCapsRequest) (*querydsl.FieldCapsResponse, error)
	OpenPointInTime(ctx context.Context, req *OpenPITRequest) (*OpenPITResponse, error)
	ClosePointInTime(ctx context.Context, pitIds []string) (*ClosePITResponse, error)
	ScriptsPainlessExec(ctx context.Context, body any) (json.RawMessage, error)
}

// DefaultSearchService is the default SearchService implementation.
type DefaultSearchService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewSearchService creates a new SearchService.
func NewSearchService(client *resty.Client, logger *logrus.Entry) SearchService {
	return &DefaultSearchService{client: client, logger: logger.WithField("service", "search")}
}

func (s *DefaultSearchService) Search(ctx context.Context, req *SearchRequest) (*querydsl.SearchResult, error) {
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_search", escapedList(req.Indices))
	} else {
		path = "/_search"
	}
	body, err := normalizeBody(req.Body)
	if err != nil {
		return nil, err
	}
	r := s.client.R().SetContext(ctx).SetBody(body)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	if req.RequestId != "" {
		r.SetHeader("X-Request-Id", req.RequestId)
	}
	resp, err := r.Post(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result querydsl.SearchResult
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) SearchMVT(ctx context.Context, req *SearchMVTRequest) (json.RawMessage, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/%s/_mvt/%s/%d/%d/%d", pathSegment(req.Index), pathSegment(req.Field), *req.Zoom, *req.X, *req.Y)
	r := s.client.R().SetContext(ctx)
	if req.Body != nil {
		r = r.SetBody(req.Body)
	}
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	return json.RawMessage(resp.Body()), nil
}

func (s *DefaultSearchService) MultiSearch(ctx context.Context, req *MultiSearchRequest) (*querydsl.MultiSearchResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_msearch", escapedList(req.Indices))
	} else {
		path = "/_msearch"
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
	var result querydsl.MultiSearchResult
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) SearchTemplate(ctx context.Context, req *SearchTemplateRequest) (*querydsl.SearchResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_search/template", escapedList(req.Indices))
	} else {
		path = "/_search/template"
	}
	r := s.client.R().SetContext(ctx).SetBody(req.Body)
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
	var result querydsl.SearchResult
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) MultiSearchTemplate(ctx context.Context, req *MultiSearchTemplateRequest) (*querydsl.MultiSearchResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_msearch/template", escapedList(req.Indices))
	} else {
		path = "/_msearch/template"
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
	var result querydsl.MultiSearchResult
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) RenderSearchTemplate(ctx context.Context, req *RenderSearchTemplateRequest) (*querydsl.RenderSearchTemplateResponse, error) {
	var path string
	if req.Id != "" {
		path = fmt.Sprintf("/_render/template/%s", pathSegment(req.Id))
	} else {
		path = "/_render/template"
	}
	r := s.client.R().SetContext(ctx)
	if req.Body != nil {
		r = r.SetBody(req.Body)
	}
	resp, err := r.Post(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result querydsl.RenderSearchTemplateResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) RankEval(ctx context.Context, req *RankEvalRequest) (*querydsl.RankEvalResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_rank_eval", escapedList(req.Indices))
	} else {
		path = "/_rank_eval"
	}
	r := s.client.R().SetContext(ctx).SetBody(req.Body)
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
	var result querydsl.RankEvalResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) Count(ctx context.Context, req *CountRequest) (int64, error) {
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_count", escapedList(req.Indices))
	} else {
		path = "/_count"
	}
	r := s.client.R().SetContext(ctx)
	if req.Body != nil {
		r = r.SetBody(req.Body)
	}
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post(path)
	if err != nil {
		return 0, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return 0, logAndReturnError(s.logger, resp)
	}
	var result querydsl.CountResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, wrapUnmarshalError(s.logger, resp, err)
	}
	return result.Count, nil
}

func (s *DefaultSearchService) Scroll(ctx context.Context, req *ScrollRequest) (*querydsl.SearchResult, error) {
	keepAlive := req.Scroll
	scrollId := req.ScrollId
	if req.Params != nil {
		if req.Params.Scroll != "" {
			keepAlive = req.Params.Scroll
		}
		if req.Params.ScrollId != "" {
			scrollId = req.Params.ScrollId
		}
	}
	if scrollId == "" {
		return nil, fmt.Errorf("scroll_id is required")
	}
	if keepAlive == "" {
		keepAlive = "5m"
	}
	body := map[string]string{
		"scroll":    keepAlive,
		"scroll_id": scrollId,
	}
	r := s.client.R().SetContext(ctx).SetBody(body)
	for k, v := range req.Params.ToMap() {
		if k == "scroll" || k == "scroll_id" {
			continue
		}
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post("/_search/scroll")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result querydsl.SearchResult
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) ClearScroll(ctx context.Context, scrollIds []string) (*querydsl.ClearScrollResponse, error) {
	if len(scrollIds) == 0 {
		return nil, fmt.Errorf("scroll_ids is required")
	}
	body := map[string][]string{"scroll_id": scrollIds}
	resp, err := s.client.R().SetContext(ctx).SetBody(body).Delete("/_search/scroll")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result querydsl.ClearScrollResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) Validate(ctx context.Context, req *ValidateQueryRequest) (*querydsl.ValidateResponse, error) {
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_validate/query", escapedList(req.Indices))
	} else {
		path = "/_validate/query"
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
	var result querydsl.ValidateResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) SearchShards(ctx context.Context, req *SearchShardsRequest) (*querydsl.SearchShardsResponse, error) {
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_search_shards", escapedList(req.Indices))
	} else {
		path = "/_search_shards"
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result querydsl.SearchShardsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) FieldCaps(ctx context.Context, req *FieldCapsRequest) (*querydsl.FieldCapsResponse, error) {
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_field_caps", escapedList(req.Indices))
	} else {
		path = "/_field_caps"
	}
	r := s.client.R().SetContext(ctx)
	if req.Body != nil {
		r = r.SetBody(req.Body)
	}
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		if resp.StatusCode() == 404 {
			return &querydsl.FieldCapsResponse{}, nil
		}
		return nil, logAndReturnError(s.logger, resp)
	}
	var result querydsl.FieldCapsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) OpenPointInTime(ctx context.Context, req *OpenPITRequest) (*OpenPITResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/%s/_pit", escapedList(req.Indices))
	r := s.client.R().SetContext(ctx)
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
	var result OpenPITResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) ClosePointInTime(ctx context.Context, pitIds []string) (*ClosePITResponse, error) {
	if len(pitIds) == 0 {
		return nil, fmt.Errorf("pit_ids is required")
	}
	body := map[string][]string{"id": pitIds}
	resp, err := s.client.R().SetContext(ctx).SetBody(body).Delete("/_pit")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result ClosePITResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultSearchService) ScriptsPainlessExec(ctx context.Context, body any) (json.RawMessage, error) {
	r := s.client.R().SetContext(ctx)
	if body != nil {
		r = r.SetBody(body)
	}
	resp, err := r.Post("/_scripts/painless/_execute")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	return json.RawMessage(resp.Body()), nil
}
