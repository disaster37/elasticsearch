package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// IndicesService provides access to the indices APIs.
type IndicesService interface {
	Create(ctx context.Context, index string, body any, params *IndicesCreateParams) (*types.AcknowledgedResponse, error)
	Delete(ctx context.Context, indices []string, params *IndicesDeleteParams) (*types.AcknowledgedResponse, error)
	Get(ctx context.Context, indices []string, params *IndicesGetParams) (map[string]*IndicesGetResponse, error)
	Exists(ctx context.Context, indices []string, params *IndicesExistsParams) (bool, error)
	Open(ctx context.Context, indices []string, params *IndicesOpenParams) (*types.AcknowledgedResponse, error)
	Close(ctx context.Context, indices []string, params *IndicesCloseParams) (*types.AcknowledgedResponse, error)
	Rollover(ctx context.Context, req *RolloverRequest) (*IndicesRolloverResponse, error)
	Shrink(ctx context.Context, req *ShrinkRequest) (*types.AcknowledgedResponse, error)
	Split(ctx context.Context, req *SplitRequest) (*types.AcknowledgedResponse, error)
	Clone(ctx context.Context, req *CloneRequest) (*types.AcknowledgedResponse, error)
	Refresh(ctx context.Context, indices []string, params *IndicesRefreshParams) (*RefreshResult, error)
	Flush(ctx context.Context, indices []string, params *IndicesFlushParams) (*types.BroadcastResponse, error)
	ForceMerge(ctx context.Context, indices []string, params *IndicesForceMergeParams) (*types.BroadcastResponse, error)
	ClearCache(ctx context.Context, indices []string, params *IndicesClearCacheParams) (*types.BroadcastResponse, error)
	Analyze(ctx context.Context, index string, body any) (*IndicesAnalyzeResponse, error)
	Stats(ctx context.Context, req *IndicesStatsRequest) (*IndicesStatsResponse, error)
	Segments(ctx context.Context, indices []string, params *IndicesSegmentsParams) (*IndicesSegmentsResponse, error)
	Recovery(ctx context.Context, indices []string, params *IndicesRecoveryParams) (*IndicesRecoveryResponse, error)
	ShardStores(ctx context.Context, indices []string, params *IndicesShardStoresParams) (*IndicesShardStoresResponse, error)
	ReloadSearchAnalyzers(ctx context.Context, indices []string, params *IndicesReloadAnalyzersParams) (*IndicesReloadAnalyzersResponse, error)
	Downsample(ctx context.Context, index, targetIndex string, body any) (*types.AcknowledgedResponse, error)
	AddBlock(ctx context.Context, req *AddBlockRequest) (*IndicesBlockResponse, error)
	PutAlias(ctx context.Context, req *PutAliasRequest) (*types.AcknowledgedResponse, error)
	GetAlias(ctx context.Context, req *GetAliasRequest) (map[string]*IndicesGetAliasResponse, error)
	ExistsAlias(ctx context.Context, req *ExistsAliasRequest) (bool, error)
	DeleteAlias(ctx context.Context, req *DeleteAliasRequest) (*types.AcknowledgedResponse, error)
	UpdateAliases(ctx context.Context, body any, params *IndicesUpdateAliasesParams) (*types.AcknowledgedResponse, error)
	GetSettings(ctx context.Context, req *GetSettingsRequest) (map[string]*IndicesGetSettingsResponse, error)
	PutSettings(ctx context.Context, indices []string, body any, params *IndicesPutSettingsParams) (*types.AcknowledgedResponse, error)
	GetMapping(ctx context.Context, indices []string, params *IndicesGetMappingParams) (map[string]*IndicesGetMappingResponse, error)
	PutMapping(ctx context.Context, req *PutMappingRequest) (*types.AcknowledgedResponse, error)
	GetFieldMapping(ctx context.Context, req *GetFieldMappingRequest) (map[string]any, error)
	GetTemplate(ctx context.Context, names []string, params *IndicesGetTemplateParams) (map[string]*IndicesGetTemplateResponse, error)
	PutTemplate(ctx context.Context, req *PutTemplateRequest) (*types.AcknowledgedResponse, error)
	ExistsTemplate(ctx context.Context, name string) (bool, error)
	DeleteTemplate(ctx context.Context, name string, params *IndicesDeleteTemplateParams) (*types.AcknowledgedResponse, error)
	GetIndexTemplate(ctx context.Context, names []string, params *IndicesGetIndexTemplateParams) (*IndicesGetIndexTemplateResponse, error)
	PutIndexTemplate(ctx context.Context, req *PutIndexTemplateRequest) (*types.AcknowledgedResponse, error)
	ExistsIndexTemplate(ctx context.Context, name string) (bool, error)
	DeleteIndexTemplate(ctx context.Context, name string, params *IndicesDeleteIndexTemplateParams) (*types.AcknowledgedResponse, error)
	SimulateIndexTemplate(ctx context.Context, req *SimulateIndexTemplateRequest) (*IndicesSimulateTemplateResponse, error)
	SimulateTemplate(ctx context.Context, req *SimulateTemplateRequest) (*IndicesSimulateTemplateResponse, error)
	CreateDataStream(ctx context.Context, name string, body any, params *IndicesCreateDataStreamParams) (*types.AcknowledgedResponse, error)
	GetDataStream(ctx context.Context, names []string, params *IndicesGetDataStreamParams) (*IndicesDataStreamGetResponse, error)
	DeleteDataStream(ctx context.Context, names []string, params *IndicesDeleteDataStreamParams) (*types.AcknowledgedResponse, error)
	DataStreamStats(ctx context.Context, names []string) (*IndicesDataStreamsStatsResponse, error)
	MigrateToDataStream(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
	ModifyDataStream(ctx context.Context, req *ModifyDataStreamRequest) (*types.AcknowledgedResponse, error)
	PromoteDataStream(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
	ResolveIndex(ctx context.Context, name string, params *IndicesResolveIndexParams) (*IndicesResolveIndexResponse, error)
	ResolveCluster(ctx context.Context, names []string, params *IndicesResolveClusterParams) (map[string]*ResolveClusterInfo, error)
}

// DefaultIndicesService is the default IndicesService implementation.
type DefaultIndicesService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewIndicesService creates a new IndicesService.
func NewIndicesService(client *resty.Client, logger *logrus.Entry) IndicesService {
	return &DefaultIndicesService{client: client, logger: logger.WithField("service", "indices")}
}

func (s *DefaultIndicesService) Create(ctx context.Context, index string, body any, params *IndicesCreateParams) (*types.AcknowledgedResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/%s", pathSegment(index)), body, params, &result)
}

func (s *DefaultIndicesService) Delete(ctx context.Context, indices []string, params *IndicesDeleteParams) (*types.AcknowledgedResponse, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("indices is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/%s", escapedList(indices)), nil, params, &result)
}

func (s *DefaultIndicesService) Get(ctx context.Context, indices []string, params *IndicesGetParams) (map[string]*IndicesGetResponse, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("indices is required")
	}
	var result map[string]*IndicesGetResponse
	return result, doReq(s.logger, ctx, s.client, "GET", "/"+escapedList(indices), nil, params, &result)
}

func (s *DefaultIndicesService) Exists(ctx context.Context, indices []string, params *IndicesExistsParams) (bool, error) {
	if len(indices) == 0 {
		return false, fmt.Errorf("indices is required")
	}
	return headReq(s.logger, ctx, s.client, "/"+escapedList(indices), params)
}

func (s *DefaultIndicesService) Open(ctx context.Context, indices []string, params *IndicesOpenParams) (*types.AcknowledgedResponse, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("indices is required")
	}
	return s.ackPost(ctx, fmt.Sprintf("/%s/_open", escapedList(indices)), params)
}

func (s *DefaultIndicesService) Close(ctx context.Context, indices []string, params *IndicesCloseParams) (*types.AcknowledgedResponse, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("indices is required")
	}
	return s.ackPost(ctx, fmt.Sprintf("/%s/_close", escapedList(indices)), params)
}

func (s *DefaultIndicesService) ackPost(ctx context.Context, path string, params interface{ ToMap() map[string]string }) (*types.AcknowledgedResponse, error) {
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", path, nil, params, &result)
}

func (s *DefaultIndicesService) Rollover(ctx context.Context, req *RolloverRequest) (*IndicesRolloverResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if req.NewIndex != "" {
		path = fmt.Sprintf("/%s/_rollover/%s", pathSegment(req.Alias), pathSegment(req.NewIndex))
	} else {
		path = fmt.Sprintf("/%s/_rollover", pathSegment(req.Alias))
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
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IndicesRolloverResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) Shrink(ctx context.Context, req *ShrinkRequest) (*types.AcknowledgedResponse, error) {
	return s.cloneOp(ctx, "shrink", req.Source, req.Target, req.Body, req.Params)
}
func (s *DefaultIndicesService) Split(ctx context.Context, req *SplitRequest) (*types.AcknowledgedResponse, error) {
	return s.cloneOp(ctx, "split", req.Source, req.Target, req.Body, req.Params)
}
func (s *DefaultIndicesService) Clone(ctx context.Context, req *CloneRequest) (*types.AcknowledgedResponse, error) {
	return s.cloneOp(ctx, "clone", req.Source, req.Target, req.Body, req.Params)
}

func (s *DefaultIndicesService) cloneOp(ctx context.Context, op, source, target string, body any, params interface{ ToMap() map[string]string }) (*types.AcknowledgedResponse, error) {
	if source == "" || target == "" {
		return nil, fmt.Errorf("source and target are required")
	}
	r := s.client.R().SetContext(ctx)
	if body != nil {
		r = r.SetBody(body)
	}
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Put(fmt.Sprintf("/%s/_%s/%s", pathSegment(source), pathSegment(op), pathSegment(target)))
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

func (s *DefaultIndicesService) Refresh(ctx context.Context, indices []string, params *IndicesRefreshParams) (*RefreshResult, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_refresh", escapedList(indices))
	} else {
		path = "/_refresh"
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
	var result RefreshResult
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) Flush(ctx context.Context, indices []string, params *IndicesFlushParams) (*types.BroadcastResponse, error) {
	return s.broadcast(ctx, "flush", indices, params)
}

func (s *DefaultIndicesService) ForceMerge(ctx context.Context, indices []string, params *IndicesForceMergeParams) (*types.BroadcastResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_forcemerge", escapedList(indices))
	} else {
		path = "/_forcemerge"
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
	var result types.BroadcastResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) ClearCache(ctx context.Context, indices []string, params *IndicesClearCacheParams) (*types.BroadcastResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_cache/clear", escapedList(indices))
	} else {
		path = "/_cache/clear"
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
	var result types.BroadcastResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) broadcast(ctx context.Context, op string, indices []string, params interface{ ToMap() map[string]string }) (*types.BroadcastResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_%s", escapedList(indices), pathSegment(op))
	} else {
		path = "/_" + pathSegment(op)
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
	var result types.BroadcastResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) Analyze(ctx context.Context, index string, body any) (*IndicesAnalyzeResponse, error) {
	var path string
	r := s.client.R().SetContext(ctx)
	if body != nil {
		r = r.SetBody(body)
	}
	var resp *resty.Response
	var err error
	if index != "" {
		path = fmt.Sprintf("/%s/_analyze", pathSegment(index))
	} else {
		path = "/_analyze"
	}
	if body != nil {
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
	var result IndicesAnalyzeResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) Stats(ctx context.Context, req *IndicesStatsRequest) (*IndicesStatsResponse, error) {
	var path string
	switch {
	case len(req.Indices) > 0 && len(req.Metrics) > 0:
		path = fmt.Sprintf("/%s/_stats/%s", escapedList(req.Indices), escapedList(req.Metrics))
	case len(req.Indices) > 0:
		path = fmt.Sprintf("/%s/_stats", escapedList(req.Indices))
	case len(req.Metrics) > 0:
		path = fmt.Sprintf("/_stats/%s", escapedList(req.Metrics))
	default:
		path = "/_stats"
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
	var result IndicesStatsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) Segments(ctx context.Context, indices []string, params *IndicesSegmentsParams) (*IndicesSegmentsResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_segments", escapedList(indices))
	} else {
		path = "/_segments"
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
	var result IndicesSegmentsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) Recovery(ctx context.Context, indices []string, params *IndicesRecoveryParams) (*IndicesRecoveryResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_recovery", escapedList(indices))
	} else {
		path = "/_recovery"
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
	var result IndicesRecoveryResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) ShardStores(ctx context.Context, indices []string, params *IndicesShardStoresParams) (*IndicesShardStoresResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_shard_stores", escapedList(indices))
	} else {
		path = "/_shard_stores"
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
	var result IndicesShardStoresResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) ReloadSearchAnalyzers(ctx context.Context, indices []string, params *IndicesReloadAnalyzersParams) (*IndicesReloadAnalyzersResponse, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("indices is required")
	}
	path := fmt.Sprintf("/%s/_reload_search_analyzers", escapedList(indices))
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	var resp *resty.Response
	var err error
	resp, err = r.Post(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IndicesReloadAnalyzersResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) Downsample(ctx context.Context, index, targetIndex string, body any) (*types.AcknowledgedResponse, error) {
	if index == "" || targetIndex == "" {
		return nil, fmt.Errorf("index and target_index are required")
	}
	r := s.client.R().SetContext(ctx)
	if body != nil {
		r = r.SetBody(body)
	}
	resp, err := r.Post(fmt.Sprintf("/%s/_downsample/%s", pathSegment(index), pathSegment(targetIndex)))
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

func (s *DefaultIndicesService) AddBlock(ctx context.Context, req *AddBlockRequest) (*IndicesBlockResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/%s/_block/%s", escapedList(req.Indices), pathSegment(req.Block))
	r := s.client.R().SetContext(ctx)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Put(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IndicesBlockResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) PutAlias(ctx context.Context, req *PutAliasRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	r := s.client.R().SetContext(ctx)
	if req.Body != nil {
		r = r.SetBody(req.Body)
	}
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Put(fmt.Sprintf("/%s/_alias/%s", pathSegment(req.Index), pathSegment(req.Alias)))
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

func (s *DefaultIndicesService) GetAlias(ctx context.Context, req *GetAliasRequest) (map[string]*IndicesGetAliasResponse, error) {
	var path string
	switch {
	case len(req.Indices) > 0 && len(req.Names) > 0:
		path = fmt.Sprintf("/%s/_alias/%s", escapedList(req.Indices), escapedList(req.Names))
	case len(req.Indices) > 0:
		path = fmt.Sprintf("/%s/_alias", escapedList(req.Indices))
	case len(req.Names) > 0:
		path = fmt.Sprintf("/_alias/%s", escapedList(req.Names))
	default:
		path = "/_alias"
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.StatusCode() == 404 {
		return map[string]*IndicesGetAliasResponse{}, nil
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result map[string]*IndicesGetAliasResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return result, nil
}

func (s *DefaultIndicesService) ExistsAlias(ctx context.Context, req *ExistsAliasRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, err
	}
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_alias/%s", escapedList(req.Indices), escapedList(req.Names))
	} else {
		path = fmt.Sprintf("/_alias/%s", escapedList(req.Names))
	}
	return headReq(s.logger, ctx, s.client, path, req.Params)
}

func (s *DefaultIndicesService) DeleteAlias(ctx context.Context, req *DeleteAliasRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/%s/_alias/%s", escapedList(req.Indices), escapedList(req.Names))
	r := s.client.R().SetContext(ctx)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Delete(path)
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

func (s *DefaultIndicesService) UpdateAliases(ctx context.Context, body any, params *IndicesUpdateAliasesParams) (*types.AcknowledgedResponse, error) {
	r := s.client.R().SetContext(ctx).SetBody(body)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post("/_aliases")
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

func (s *DefaultIndicesService) GetSettings(ctx context.Context, req *GetSettingsRequest) (map[string]*IndicesGetSettingsResponse, error) {
	var path string
	switch {
	case len(req.Indices) > 0 && len(req.Names) > 0:
		path = fmt.Sprintf("/%s/_settings/%s", escapedList(req.Indices), escapedList(req.Names))
	case len(req.Indices) > 0:
		path = fmt.Sprintf("/%s/_settings", escapedList(req.Indices))
	case len(req.Names) > 0:
		path = fmt.Sprintf("/_settings/%s", escapedList(req.Names))
	default:
		path = "/_settings"
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
	var result map[string]*IndicesGetSettingsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return result, nil
}

func (s *DefaultIndicesService) PutSettings(ctx context.Context, indices []string, body any, params *IndicesPutSettingsParams) (*types.AcknowledgedResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_settings", escapedList(indices))
	} else {
		path = "/_settings"
	}
	r := s.client.R().SetContext(ctx).SetBody(body)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Put(path)
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

func (s *DefaultIndicesService) GetMapping(ctx context.Context, indices []string, params *IndicesGetMappingParams) (map[string]*IndicesGetMappingResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_mapping", escapedList(indices))
	} else {
		path = "/_mapping"
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
	var result map[string]*IndicesGetMappingResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return result, nil
}

func (s *DefaultIndicesService) PutMapping(ctx context.Context, req *PutMappingRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/%s/_mapping", escapedList(req.Indices))
	r := s.client.R().SetContext(ctx).SetBody(req.Body)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Put(path)
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

func (s *DefaultIndicesService) GetFieldMapping(ctx context.Context, req *GetFieldMappingRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if len(req.Indices) > 0 {
		path = fmt.Sprintf("/%s/_mapping/field/%s", escapedList(req.Indices), escapedList(req.Fields))
	} else {
		path = fmt.Sprintf("/_mapping/field/%s", escapedList(req.Fields))
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
	var result map[string]any
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return result, nil
}

// ---- Legacy templates ----

func (s *DefaultIndicesService) GetTemplate(ctx context.Context, names []string, params *IndicesGetTemplateParams) (map[string]*IndicesGetTemplateResponse, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_template/%s", escapedList(names))
	} else {
		path = "/_template"
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
	var result map[string]*IndicesGetTemplateResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return result, nil
}

func (s *DefaultIndicesService) PutTemplate(ctx context.Context, req *PutTemplateRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	r := s.client.R().SetContext(ctx).SetBody(req.Body)
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Put(fmt.Sprintf("/_template/%s", pathSegment(req.Name)))
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

func (s *DefaultIndicesService) ExistsTemplate(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name is required")
	}
	return headReq(s.logger, ctx, s.client, fmt.Sprintf("/_template/%s", pathSegment(name)), nil)
}

func (s *DefaultIndicesService) DeleteTemplate(ctx context.Context, name string, params *IndicesDeleteTemplateParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_template/%s", pathSegment(name)), nil, params, &result)
}

// ---- Composable templates ----

func (s *DefaultIndicesService) GetIndexTemplate(ctx context.Context, names []string, params *IndicesGetIndexTemplateParams) (*IndicesGetIndexTemplateResponse, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_index_template/%s", escapedList(names))
	} else {
		path = "/_index_template"
	}
	var result IndicesGetIndexTemplateResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultIndicesService) PutIndexTemplate(ctx context.Context, req *PutIndexTemplateRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_index_template/%s", pathSegment(req.Name)), req.Body, req.Params, &result)
}

func (s *DefaultIndicesService) ExistsIndexTemplate(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name is required")
	}
	return headReq(s.logger, ctx, s.client, fmt.Sprintf("/_index_template/%s", pathSegment(name)), nil)
}

func (s *DefaultIndicesService) DeleteIndexTemplate(ctx context.Context, name string, params *IndicesDeleteIndexTemplateParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Delete(fmt.Sprintf("/_index_template/%s", pathSegment(name)))
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

func (s *DefaultIndicesService) SimulateIndexTemplate(ctx context.Context, req *SimulateIndexTemplateRequest) (*IndicesSimulateTemplateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	r := s.client.R().SetContext(ctx)
	if req.Body != nil {
		r = r.SetBody(req.Body)
	}
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post(fmt.Sprintf("/_index_template/_simulate_index/%s", pathSegment(req.Name)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IndicesSimulateTemplateResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) SimulateTemplate(ctx context.Context, req *SimulateTemplateRequest) (*IndicesSimulateTemplateResponse, error) {
	var path string
	r := s.client.R().SetContext(ctx)
	if req.Body != nil {
		r = r.SetBody(req.Body)
	}
	if req.Name != "" {
		path = fmt.Sprintf("/_index_template/_simulate/%s", pathSegment(req.Name))
	} else {
		path = "/_index_template/_simulate"
	}
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
	var result IndicesSimulateTemplateResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

// ---- Data streams ----

func (s *DefaultIndicesService) CreateDataStream(ctx context.Context, name string, body any, params *IndicesCreateDataStreamParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	r := s.client.R().SetContext(ctx)
	if body != nil {
		r = r.SetBody(body)
	}
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Put(fmt.Sprintf("/_data_stream/%s", pathSegment(name)))
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

func (s *DefaultIndicesService) GetDataStream(ctx context.Context, names []string, params *IndicesGetDataStreamParams) (*IndicesDataStreamGetResponse, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_data_stream/%s", escapedList(names))
	} else {
		path = "/_data_stream"
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
	var result IndicesDataStreamGetResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) DeleteDataStream(ctx context.Context, names []string, params *IndicesDeleteDataStreamParams) (*types.AcknowledgedResponse, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("names is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Delete(fmt.Sprintf("/_data_stream/%s", escapedList(names)))
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

func (s *DefaultIndicesService) DataStreamStats(ctx context.Context, names []string) (*IndicesDataStreamsStatsResponse, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_data_stream/%s/_stats", escapedList(names))
	} else {
		path = "/_data_stream/_stats"
	}
	resp, err := s.client.R().SetContext(ctx).Get(path)
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IndicesDataStreamsStatsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) MigrateToDataStream(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	resp, err := s.client.R().SetContext(ctx).Post(fmt.Sprintf("/_data_stream/_migrate/%s", pathSegment(name)))
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

func (s *DefaultIndicesService) ModifyDataStream(ctx context.Context, req *ModifyDataStreamRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	r := s.client.R().SetContext(ctx).SetBody(map[string]any{"actions": req.Actions})
	for k, v := range req.Params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Post("/_data_stream/_modify")
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

func (s *DefaultIndicesService) PromoteDataStream(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	resp, err := s.client.R().SetContext(ctx).Post(fmt.Sprintf("/_data_stream/_promote/%s", pathSegment(name)))
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

func (s *DefaultIndicesService) ResolveIndex(ctx context.Context, name string, params *IndicesResolveIndexParams) (*IndicesResolveIndexResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(fmt.Sprintf("/_resolve/index/%s", pathSegment(name)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result IndicesResolveIndexResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}

func (s *DefaultIndicesService) ResolveCluster(ctx context.Context, names []string, params *IndicesResolveClusterParams) (map[string]*ResolveClusterInfo, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("names is required")
	}
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(fmt.Sprintf("/_resolve/cluster/%s", escapedList(names)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result map[string]*ResolveClusterInfo
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return result, nil
}
