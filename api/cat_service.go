package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// CatService provides access to the cat APIs. All methods request format=json
// and decode into typed row slices.
type CatService interface {
	Aliases(ctx context.Context, names []string, p *CatAliasesParams) (CatAliasesResponse, error)
	Allocation(ctx context.Context, nodeIds []string, p *CatAllocationParams) (CatAllocationResponse, error)
	CircuitBreaker(ctx context.Context, p *CatCircuitBreakerParams) (CatCircuitBreakerResponse, error)
	ComponentTemplates(ctx context.Context, name string, p *CatComponentTemplatesParams) (CatComponentTemplatesResponse, error)
	Count(ctx context.Context, indices []string, p *CatCountParams) (CatCountResponse, error)
	Fielddata(ctx context.Context, fields []string, p *CatFielddataParams) (CatFielddataResponse, error)
	Health(ctx context.Context, p *CatHealthParams) (CatHealthResponse, error)
	Help(ctx context.Context) (string, error)
	Indices(ctx context.Context, indices []string, p *CatIndicesParams) (CatIndicesResponse, error)
	Master(ctx context.Context, p *CatMasterParams) (CatMasterResponse, error)
	MLDataFrameAnalytics(ctx context.Context, id string, p *CatMLDataFrameAnalyticsParams) (CatMLDataFrameAnalyticsResponse, error)
	MLDatafeeds(ctx context.Context, datafeedId string, p *CatMLDatafeedsParams) (CatMLDatafeedsResponse, error)
	MLJobs(ctx context.Context, jobId string, p *CatMLJobsParams) (CatMLJobsResponse, error)
	MLTrainedModels(ctx context.Context, modelId string, p *CatMLTrainedModelsParams) (CatMLTrainedModelsResponse, error)
	NodeAttrs(ctx context.Context, p *CatNodeAttrsParams) (CatNodeAttrsResponse, error)
	Nodes(ctx context.Context, p *CatNodesParams) (CatNodesResponse, error)
	PendingTasks(ctx context.Context, p *CatPendingTasksParams) (CatPendingTasksResponse, error)
	Plugins(ctx context.Context, p *CatPluginsParams) (CatPluginsResponse, error)
	Recovery(ctx context.Context, indices []string, p *CatRecoveryParams) (CatRecoveryResponse, error)
	Repositories(ctx context.Context, p *CatRepositoriesParams) (CatRepositoriesResponse, error)
	Segments(ctx context.Context, indices []string, p *CatSegmentsParams) (CatSegmentsResponse, error)
	Shards(ctx context.Context, indices []string, p *CatShardsParams) (CatShardsResponse, error)
	Snapshots(ctx context.Context, repository string, p *CatSnapshotsParams) (CatSnapshotsResponse, error)
	Tasks(ctx context.Context, p *CatTasksParams) (CatTasksResponse, error)
	Templates(ctx context.Context, name string, p *CatTemplatesParams) (CatTemplatesResponse, error)
	ThreadPool(ctx context.Context, patterns []string, p *CatThreadPoolParams) (CatThreadPoolResponse, error)
	Transforms(ctx context.Context, transformId string, p *CatTransformsParams) (CatTransformsResponse, error)
}

// DefaultCatService is the default CatService implementation.
type DefaultCatService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewCatService creates a new CatService.
func NewCatService(client *resty.Client, logger *logrus.Entry) CatService {
	return &DefaultCatService{client: client, logger: logger.WithField("service", "cat")}
}

// catGet performs a GET cat request decoding JSON into out.
func (s *DefaultCatService) catGet(ctx context.Context, path string, params interface{ ToMap() map[string]string }, out any) error {
	r := s.client.R().SetContext(ctx)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(path)
	if err != nil {
		return wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return logAndReturnError(s.logger, resp)
	}
	if err := json.Unmarshal(resp.Body(), out); err != nil {
		return wrapUnmarshalError(s.logger, resp, err)
	}
	return nil
}

func (s *DefaultCatService) Aliases(ctx context.Context, names []string, p *CatAliasesParams) (CatAliasesResponse, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_cat/aliases/%s", escapedList(names))
	} else {
		path = "/_cat/aliases"
	}
	var result CatAliasesResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) Allocation(ctx context.Context, nodeIds []string, p *CatAllocationParams) (CatAllocationResponse, error) {
	var path string
	if len(nodeIds) > 0 {
		path = fmt.Sprintf("/_cat/allocation/%s", escapedList(nodeIds))
	} else {
		path = "/_cat/allocation"
	}
	var result CatAllocationResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) CircuitBreaker(ctx context.Context, p *CatCircuitBreakerParams) (CatCircuitBreakerResponse, error) {
	var result CatCircuitBreakerResponse
	return result, s.catGet(ctx, "/_cat/circuit_breakers", p, &result)
}

func (s *DefaultCatService) ComponentTemplates(ctx context.Context, name string, p *CatComponentTemplatesParams) (CatComponentTemplatesResponse, error) {
	var path string
	if name != "" {
		path = fmt.Sprintf("/_cat/component_templates/%s", pathSegment(name))
	} else {
		path = "/_cat/component_templates"
	}
	var result CatComponentTemplatesResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) Count(ctx context.Context, indices []string, p *CatCountParams) (CatCountResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/_cat/count/%s", escapedList(indices))
	} else {
		path = "/_cat/count"
	}
	var result CatCountResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) Fielddata(ctx context.Context, fields []string, p *CatFielddataParams) (CatFielddataResponse, error) {
	var path string
	if len(fields) > 0 {
		path = fmt.Sprintf("/_cat/fielddata/%s", escapedList(fields))
	} else {
		path = "/_cat/fielddata"
	}
	var result CatFielddataResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) Health(ctx context.Context, p *CatHealthParams) (CatHealthResponse, error) {
	var result CatHealthResponse
	return result, s.catGet(ctx, "/_cat/health", p, &result)
}

func (s *DefaultCatService) Help(ctx context.Context) (string, error) {
	resp, err := s.client.R().SetContext(ctx).Get("/_cat")
	if err != nil {
		return "", wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return "", logAndReturnError(s.logger, resp)
	}
	return string(resp.Body()), nil
}

func (s *DefaultCatService) Indices(ctx context.Context, indices []string, p *CatIndicesParams) (CatIndicesResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/_cat/indices/%s", escapedList(indices))
	} else {
		path = "/_cat/indices"
	}
	var result CatIndicesResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) Master(ctx context.Context, p *CatMasterParams) (CatMasterResponse, error) {
	var result CatMasterResponse
	return result, s.catGet(ctx, "/_cat/master", p, &result)
}

func (s *DefaultCatService) MLDataFrameAnalytics(ctx context.Context, id string, p *CatMLDataFrameAnalyticsParams) (CatMLDataFrameAnalyticsResponse, error) {
	var path string
	if id != "" {
		path = fmt.Sprintf("/_cat/ml/data_frame/analytics/%s", pathSegment(id))
	} else {
		path = "/_cat/ml/data_frame/analytics"
	}
	var result CatMLDataFrameAnalyticsResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) MLDatafeeds(ctx context.Context, datafeedId string, p *CatMLDatafeedsParams) (CatMLDatafeedsResponse, error) {
	var path string
	if datafeedId != "" {
		path = fmt.Sprintf("/_cat/ml/datafeeds/%s", pathSegment(datafeedId))
	} else {
		path = "/_cat/ml/datafeeds"
	}
	var result CatMLDatafeedsResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) MLJobs(ctx context.Context, jobId string, p *CatMLJobsParams) (CatMLJobsResponse, error) {
	var path string
	if jobId != "" {
		path = fmt.Sprintf("/_cat/ml/jobs/%s", pathSegment(jobId))
	} else {
		path = "/_cat/ml/jobs"
	}
	var result CatMLJobsResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) MLTrainedModels(ctx context.Context, modelId string, p *CatMLTrainedModelsParams) (CatMLTrainedModelsResponse, error) {
	var path string
	if modelId != "" {
		path = fmt.Sprintf("/_cat/ml/trained_models/%s", pathSegment(modelId))
	} else {
		path = "/_cat/ml/trained_models"
	}
	var result CatMLTrainedModelsResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) NodeAttrs(ctx context.Context, p *CatNodeAttrsParams) (CatNodeAttrsResponse, error) {
	var result CatNodeAttrsResponse
	return result, s.catGet(ctx, "/_cat/nodeattrs", p, &result)
}

func (s *DefaultCatService) Nodes(ctx context.Context, p *CatNodesParams) (CatNodesResponse, error) {
	var result CatNodesResponse
	return result, s.catGet(ctx, "/_cat/nodes", p, &result)
}

func (s *DefaultCatService) PendingTasks(ctx context.Context, p *CatPendingTasksParams) (CatPendingTasksResponse, error) {
	var result CatPendingTasksResponse
	return result, s.catGet(ctx, "/_cat/pending_tasks", p, &result)
}

func (s *DefaultCatService) Plugins(ctx context.Context, p *CatPluginsParams) (CatPluginsResponse, error) {
	var result CatPluginsResponse
	return result, s.catGet(ctx, "/_cat/plugins", p, &result)
}

func (s *DefaultCatService) Recovery(ctx context.Context, indices []string, p *CatRecoveryParams) (CatRecoveryResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/_cat/recovery/%s", escapedList(indices))
	} else {
		path = "/_cat/recovery"
	}
	var result CatRecoveryResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) Repositories(ctx context.Context, p *CatRepositoriesParams) (CatRepositoriesResponse, error) {
	var result CatRepositoriesResponse
	return result, s.catGet(ctx, "/_cat/repositories", p, &result)
}

func (s *DefaultCatService) Segments(ctx context.Context, indices []string, p *CatSegmentsParams) (CatSegmentsResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/_cat/segments/%s", escapedList(indices))
	} else {
		path = "/_cat/segments"
	}
	var result CatSegmentsResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) Shards(ctx context.Context, indices []string, p *CatShardsParams) (CatShardsResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/_cat/shards/%s", escapedList(indices))
	} else {
		path = "/_cat/shards"
	}
	var result CatShardsResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) Snapshots(ctx context.Context, repository string, p *CatSnapshotsParams) (CatSnapshotsResponse, error) {
	var path string
	if repository != "" {
		path = fmt.Sprintf("/_cat/snapshots/%s", pathSegment(repository))
	} else {
		path = "/_cat/snapshots"
	}
	var result CatSnapshotsResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) Tasks(ctx context.Context, p *CatTasksParams) (CatTasksResponse, error) {
	var result CatTasksResponse
	return result, s.catGet(ctx, "/_cat/tasks", p, &result)
}

func (s *DefaultCatService) Templates(ctx context.Context, name string, p *CatTemplatesParams) (CatTemplatesResponse, error) {
	var path string
	if name != "" {
		path = fmt.Sprintf("/_cat/templates/%s", pathSegment(name))
	} else {
		path = "/_cat/templates"
	}
	var result CatTemplatesResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) ThreadPool(ctx context.Context, patterns []string, p *CatThreadPoolParams) (CatThreadPoolResponse, error) {
	var path string
	if len(patterns) > 0 {
		path = fmt.Sprintf("/_cat/thread_pool/%s", escapedList(patterns))
	} else {
		path = "/_cat/thread_pool"
	}
	var result CatThreadPoolResponse
	return result, s.catGet(ctx, path, p, &result)
}

func (s *DefaultCatService) Transforms(ctx context.Context, transformId string, p *CatTransformsParams) (CatTransformsResponse, error) {
	var path string
	if transformId != "" {
		path = fmt.Sprintf("/_cat/transforms/%s", pathSegment(transformId))
	} else {
		path = "/_cat/transforms"
	}
	var result CatTransformsResponse
	return result, s.catGet(ctx, path, p, &result)
}
