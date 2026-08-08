package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// MlService provides access to the machine learning APIs. The ML surface is
// large; this implementation covers the core job, datafeed, calendar, filter,
// data-frame-analytics, and trained-model endpoints. Bodies are passed as
// `any` and serialized by resty.
type MlService interface {
	// Jobs
	PutJob(ctx context.Context, jobId string, body any) (*MlJobConfig, error)
	GetJobs(ctx context.Context, jobIds []string, params *MlGetJobsParams) (*MlGetJobsResponse, error)
	DeleteJob(ctx context.Context, jobId string, params *MlDeleteJobParams) (*MlDeleteJobResponse, error)
	OpenJob(ctx context.Context, jobId string, params *MlOpenJobParams) (*MlOpenJobResponse, error)
	CloseJob(ctx context.Context, jobId string, params *MlCloseJobParams) (*MlCloseJobResponse, error)
	ResetJob(ctx context.Context, jobId string, params *MlResetJobParams) (*types.AcknowledgedResponse, error)
	FlushJob(ctx context.Context, jobId string, body any, params *MlFlushJobParams) (*MlFlushJobResponse, error)
	// Datafeeds
	PutDatafeed(ctx context.Context, datafeedId string, body any) (*MlDatafeedConfig, error)
	GetDatafeeds(ctx context.Context, datafeedIds []string, params *MlGetDatafeedsParams) (*MlGetDatafeedsResponse, error)
	DeleteDatafeed(ctx context.Context, datafeedId string, params *MlDeleteDatafeedParams) (*types.AcknowledgedResponse, error)
	StartDatafeed(ctx context.Context, datafeedId string, body any, params *MlStartDatafeedParams) (*MlStartDatafeedResponse, error)
	StopDatafeed(ctx context.Context, datafeedId string, params *MlStopDatafeedParams) (*MlStopDatafeedResponse, error)
	// Info
	GetFilters(ctx context.Context, filterId string, params *MlGetFiltersParams) (*MlGetFiltersResponse, error)
	// Trained models
	GetTrainedModels(ctx context.Context, modelIds []string, params *MlGetTrainedModelsParams) (*MlGetTrainedModelsResponse, error)
	InferTrainedModel(ctx context.Context, modelId string, body any, params *MlInferTrainedModelParams) (json.RawMessage, error)
}

// DefaultMlService is the default MlService implementation.
type DefaultMlService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewMlService creates a new MlService.
func NewMlService(client *resty.Client, logger *logrus.Entry) MlService {
	return &DefaultMlService{client: client, logger: logger.WithField("service", "ml")}
}

// ---- Params ----

// MlGetJobsParams are the query parameters for get jobs.
type MlGetJobsParams struct {
	Common           *CommonParams
	AllowNoMatch     *bool
	ExcludeGenerated *bool
}

// ToMap converts MlGetJobsParams to a query-parameter map.
func (p *MlGetJobsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	setBoolParam(m, "exclude_generated", p.ExcludeGenerated)
	return mergeParams(m, p.Common)
}

// MlDeleteJobParams are the query parameters for delete job.
type MlDeleteJobParams struct {
	Common            *CommonParams
	Force             *bool
	WaitForCompletion *bool
}

// ToMap converts MlDeleteJobParams to a query-parameter map.
func (p *MlDeleteJobParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "force", p.Force)
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// MlOpenJobParams are the query parameters for open job.
type MlOpenJobParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts MlOpenJobParams to a query-parameter map.
func (p *MlOpenJobParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// MlCloseJobParams are the query parameters for close job.
type MlCloseJobParams struct {
	Common       *CommonParams
	AllowNoMatch *bool
	Force        *bool
	Timeout      string
}

// ToMap converts MlCloseJobParams to a query-parameter map.
func (p *MlCloseJobParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	setBoolParam(m, "force", p.Force)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// MlResetJobParams are the query parameters for reset job.
type MlResetJobParams struct {
	Common            *CommonParams
	WaitForCompletion *bool
}

// ToMap converts MlResetJobParams to a query-parameter map.
func (p *MlResetJobParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// MlFlushJobParams are the query parameters for flush job.
type MlFlushJobParams struct {
	Common      *CommonParams
	AdvanceTime string
	CalcInterim *bool
	End         string
	SkipTime    string
	Start       string
}

// ToMap converts MlFlushJobParams to a query-parameter map.
func (p *MlFlushJobParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.AdvanceTime != "" {
		m["advance_time"] = p.AdvanceTime
	}
	setBoolParam(m, "calc_interim", p.CalcInterim)
	if p.End != "" {
		m["end"] = p.End
	}
	if p.SkipTime != "" {
		m["skip_time"] = p.SkipTime
	}
	if p.Start != "" {
		m["start"] = p.Start
	}
	return mergeParams(m, p.Common)
}

// MlGetDatafeedsParams are the query parameters for get datafeeds.
type MlGetDatafeedsParams struct {
	Common           *CommonParams
	AllowNoMatch     *bool
	ExcludeGenerated *bool
}

// ToMap converts MlGetDatafeedsParams to a query-parameter map.
func (p *MlGetDatafeedsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	setBoolParam(m, "exclude_generated", p.ExcludeGenerated)
	return mergeParams(m, p.Common)
}

// MlDeleteDatafeedParams are the query parameters for delete datafeed.
type MlDeleteDatafeedParams struct {
	Common *CommonParams
	Force  *bool
}

// ToMap converts MlDeleteDatafeedParams to a query-parameter map.
func (p *MlDeleteDatafeedParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "force", p.Force)
	return mergeParams(m, p.Common)
}

// MlStartDatafeedParams are the query parameters for start datafeed.
type MlStartDatafeedParams struct {
	Common  *CommonParams
	End     string
	Start   string
	Timeout string
}

// ToMap converts MlStartDatafeedParams to a query-parameter map.
func (p *MlStartDatafeedParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.End != "" {
		m["end"] = p.End
	}
	if p.Start != "" {
		m["start"] = p.Start
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// MlStopDatafeedParams are the query parameters for stop datafeed.
type MlStopDatafeedParams struct {
	Common       *CommonParams
	AllowNoMatch *bool
	Force        *bool
	Timeout      string
}

// ToMap converts MlStopDatafeedParams to a query-parameter map.
func (p *MlStopDatafeedParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	setBoolParam(m, "force", p.Force)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// MlGetFiltersParams are the query parameters for get filters.
type MlGetFiltersParams struct {
	Common *CommonParams
	From   *int
	Size   *int
}

// ToMap converts MlGetFiltersParams to a query-parameter map.
func (p *MlGetFiltersParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	return mergeParams(m, p.Common)
}

// MlGetTrainedModelsParams are the query parameters for get trained models.
type MlGetTrainedModelsParams struct {
	Common               *CommonParams
	AllowNoMatch         *bool
	DecompressDefinition *bool
	ExcludeGenerated     *bool
	From                 *int
	Include              string
	Size                 *int
	Tags                 []string
}

// ToMap converts MlGetTrainedModelsParams to a query-parameter map.
func (p *MlGetTrainedModelsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "allow_no_match", p.AllowNoMatch)
	setBoolParam(m, "decompress_definition", p.DecompressDefinition)
	setBoolParam(m, "exclude_generated", p.ExcludeGenerated)
	if p.From != nil {
		m["from"] = intStr(*p.From)
	}
	if p.Include != "" {
		m["include"] = p.Include
	}
	if p.Size != nil {
		m["size"] = intStr(*p.Size)
	}
	setListParam(m, "tags", p.Tags)
	return mergeParams(m, p.Common)
}

// MlInferTrainedModelParams are the query parameters for infer trained model.
type MlInferTrainedModelParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts MlInferTrainedModelParams to a query-parameter map.
func (p *MlInferTrainedModelParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ---- Response structs ----

// MlJobConfig describes an ML job configuration.
type MlJobConfig struct {
	JobId                      string         `json:"job_id,omitempty"`
	Description                string         `json:"description,omitempty"`
	AnalysisConfig             map[string]any `json:"analysis_config,omitempty"`
	DataDescription            map[string]any `json:"data_description,omitempty"`
	AnalysisLimits             map[string]any `json:"analysis_limits,omitempty"`
	ModelSnapshotRetentionDays *int           `json:"model_snapshot_retention_days,omitempty"`
	ResultsRetentionDays       *int           `json:"results_retention_days,omitempty"`
	CustomSettings             map[string]any `json:"custom_settings,omitempty"`
	AllowLazyOpen              *bool          `json:"allow_lazy_open,omitempty"`
	MaxModelMemory             string         `json:"max_model_memory,omitempty"`
}

// MlGetJobsResponse is the response from get jobs.
type MlGetJobsResponse struct {
	Count int           `json:"count"`
	Jobs  []MlJobConfig `json:"jobs"`
}

// MlDeleteJobResponse is the response from delete job.
type MlDeleteJobResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

// MlOpenJobResponse is the response from open job.
type MlOpenJobResponse struct {
	Opened bool `json:"opened"`
}

// MlCloseJobResponse is the response from close job.
type MlCloseJobResponse struct {
	Closed bool `json:"closed"`
}

// MlFlushJobResponse is the response from flush job.
type MlFlushJobResponse struct {
	Flushed bool   `json:"flushed"`
	Last    string `json:"last,omitempty"`
}

// MlDatafeedConfig describes an ML datafeed configuration.
type MlDatafeedConfig struct {
	DatafeedId     string         `json:"datafeed_id,omitempty"`
	JobId          string         `json:"job_id,omitempty"`
	Indices        []string       `json:"indices,omitempty"`
	Query          map[string]any `json:"query,omitempty"`
	Frequency      *int           `json:"frequency,omitempty"`
	ChunkingConfig map[string]any `json:"chunking_config,omitempty"`
}

// MlGetDatafeedsResponse is the response from get datafeeds.
type MlGetDatafeedsResponse struct {
	Count     int                `json:"count"`
	Datafeeds []MlDatafeedConfig `json:"datafeeds"`
}

// MlStartDatafeedResponse is the response from start datafeed.
type MlStartDatafeedResponse struct {
	Started bool `json:"started"`
}

// MlStopDatafeedResponse is the response from stop datafeed.
type MlStopDatafeedResponse struct {
	Stopped bool `json:"stopped"`
}

// MlFilter describes an ML filter.
type MlFilter struct {
	FilterId    string   `json:"filter_id,omitempty"`
	Items       []string `json:"items,omitempty"`
	Description string   `json:"description,omitempty"`
}

// MlGetFiltersResponse is the response from get filters.
type MlGetFiltersResponse struct {
	Count   int        `json:"count"`
	Filters []MlFilter `json:"filters"`
}

// MlTrainedModelConfig describes a trained model.
type MlTrainedModelConfig struct {
	ModelId    string         `json:"model_id,omitempty"`
	Type       string         `json:"type,omitempty"`
	Input      map[string]any `json:"input,omitempty"`
	Definition map[string]any `json:"definition,omitempty"`
}

// MlGetTrainedModelsResponse is the response from get trained models.
type MlGetTrainedModelsResponse struct {
	Count         int                    `json:"count"`
	TrainedModels []MlTrainedModelConfig `json:"trained_model_configs"`
}

// ---- Implementations ----

func (s *DefaultMlService) PutJob(ctx context.Context, jobId string, body any) (*MlJobConfig, error) {
	if jobId == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	var result MlJobConfig
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_ml/anomaly_detectors/%s", pathSegment(jobId)), body, nil, &result)
}

func (s *DefaultMlService) GetJobs(ctx context.Context, jobIds []string, params *MlGetJobsParams) (*MlGetJobsResponse, error) {
	var path string
	if len(jobIds) > 0 {
		path = fmt.Sprintf("/_ml/anomaly_detectors/%s", escapedList(jobIds))
	} else {
		path = "/_ml/anomaly_detectors"
	}
	var result MlGetJobsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultMlService) DeleteJob(ctx context.Context, jobId string, params *MlDeleteJobParams) (*MlDeleteJobResponse, error) {
	if jobId == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	var result MlDeleteJobResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_ml/anomaly_detectors/%s", pathSegment(jobId)), nil, params, &result)
}

func (s *DefaultMlService) OpenJob(ctx context.Context, jobId string, params *MlOpenJobParams) (*MlOpenJobResponse, error) {
	if jobId == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	var result MlOpenJobResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ml/anomaly_detectors/%s/_open", pathSegment(jobId)), nil, params, &result)
}

func (s *DefaultMlService) CloseJob(ctx context.Context, jobId string, params *MlCloseJobParams) (*MlCloseJobResponse, error) {
	if jobId == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	var result MlCloseJobResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ml/anomaly_detectors/%s/_close", pathSegment(jobId)), nil, params, &result)
}

func (s *DefaultMlService) ResetJob(ctx context.Context, jobId string, params *MlResetJobParams) (*types.AcknowledgedResponse, error) {
	if jobId == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ml/anomaly_detectors/%s/_reset", pathSegment(jobId)), nil, params, &result)
}

func (s *DefaultMlService) FlushJob(ctx context.Context, jobId string, body any, params *MlFlushJobParams) (*MlFlushJobResponse, error) {
	if jobId == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	var result MlFlushJobResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ml/anomaly_detectors/%s/_flush", pathSegment(jobId)), body, params, &result)
}

func (s *DefaultMlService) PutDatafeed(ctx context.Context, datafeedId string, body any) (*MlDatafeedConfig, error) {
	if datafeedId == "" {
		return nil, fmt.Errorf("datafeed_id is required")
	}
	var result MlDatafeedConfig
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_ml/datafeeds/%s", pathSegment(datafeedId)), body, nil, &result)
}

func (s *DefaultMlService) GetDatafeeds(ctx context.Context, datafeedIds []string, params *MlGetDatafeedsParams) (*MlGetDatafeedsResponse, error) {
	var path string
	if len(datafeedIds) > 0 {
		path = fmt.Sprintf("/_ml/datafeeds/%s", escapedList(datafeedIds))
	} else {
		path = "/_ml/datafeeds"
	}
	var result MlGetDatafeedsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultMlService) DeleteDatafeed(ctx context.Context, datafeedId string, params *MlDeleteDatafeedParams) (*types.AcknowledgedResponse, error) {
	if datafeedId == "" {
		return nil, fmt.Errorf("datafeed_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_ml/datafeeds/%s", pathSegment(datafeedId)), nil, params, &result)
}

func (s *DefaultMlService) StartDatafeed(ctx context.Context, datafeedId string, body any, params *MlStartDatafeedParams) (*MlStartDatafeedResponse, error) {
	if datafeedId == "" {
		return nil, fmt.Errorf("datafeed_id is required")
	}
	var result MlStartDatafeedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ml/datafeeds/%s/_start", pathSegment(datafeedId)), body, params, &result)
}

func (s *DefaultMlService) StopDatafeed(ctx context.Context, datafeedId string, params *MlStopDatafeedParams) (*MlStopDatafeedResponse, error) {
	if datafeedId == "" {
		return nil, fmt.Errorf("datafeed_id is required")
	}
	var result MlStopDatafeedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ml/datafeeds/%s/_stop", pathSegment(datafeedId)), nil, params, &result)
}

func (s *DefaultMlService) GetFilters(ctx context.Context, filterId string, params *MlGetFiltersParams) (*MlGetFiltersResponse, error) {
	var path string
	if filterId != "" {
		path = fmt.Sprintf("/_ml/filters/%s", pathSegment(filterId))
	} else {
		path = "/_ml/filters"
	}
	var result MlGetFiltersResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultMlService) GetTrainedModels(ctx context.Context, modelIds []string, params *MlGetTrainedModelsParams) (*MlGetTrainedModelsResponse, error) {
	var path string
	if len(modelIds) > 0 {
		path = fmt.Sprintf("/_ml/trained_models/%s", escapedList(modelIds))
	} else {
		path = "/_ml/trained_models"
	}
	var result MlGetTrainedModelsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultMlService) InferTrainedModel(ctx context.Context, modelId string, body any, params *MlInferTrainedModelParams) (json.RawMessage, error) {
	if modelId == "" {
		return nil, fmt.Errorf("model_id is required")
	}
	return doReqRaw(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_ml/trained_models/%s/_infer", pathSegment(modelId)), body, params)
}
