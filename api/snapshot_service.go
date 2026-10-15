package api

import (
	"context"
	"fmt"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// SnapshotService provides access to the snapshot APIs.
type SnapshotService interface {
	CreateRepository(ctx context.Context, name string, body any, params *SnapshotCreateRepositoryParams) (*types.AcknowledgedResponse, error)
	GetRepository(ctx context.Context, names []string, params *SnapshotGetRepositoryParams) (map[string]*SnapshotRepository, error)
	DeleteRepository(ctx context.Context, names []string, params *SnapshotDeleteRepositoryParams) (*types.AcknowledgedResponse, error)
	VerifyRepository(ctx context.Context, name string, params *SnapshotVerifyRepositoryParams) (*SnapshotVerifyRepositoryResponse, error)
	RepositoryAnalyze(ctx context.Context, name string, params *SnapshotRepositoryAnalyzeParams) (*SnapshotRepositoryAnalyzeResponse, error)
	RepositoryVerifyIntegrity(ctx context.Context, name string, params *SnapshotRepositoryVerifyIntegrityParams) (*SnapshotRepositoryVerifyIntegrityResponse, error)
	Create(ctx context.Context, req *SnapshotCreateRequest) (*SnapshotCreateResponse, error)
	Get(ctx context.Context, req *SnapshotGetRequest) (*SnapshotGetResponse, error)
	Delete(ctx context.Context, req *SnapshotDeleteRequest) (*types.AcknowledgedResponse, error)
	Restore(ctx context.Context, req *SnapshotRestoreRequest) (*SnapshotRestoreResponse, error)
	Status(ctx context.Context, req *SnapshotStatusRequest) (*SnapshotStatusResponse, error)
	Clone(ctx context.Context, req *SnapshotCloneRequest) (*types.AcknowledgedResponse, error)
	CleanupRepository(ctx context.Context, name string, params *SnapshotCleanupRepositoryParams) (*SnapshotCleanupRepositoryResponse, error)
}

// DefaultSnapshotService is the default SnapshotService implementation.
type DefaultSnapshotService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewSnapshotService creates a new SnapshotService.
func NewSnapshotService(client *resty.Client, logger *logrus.Entry) SnapshotService {
	return &DefaultSnapshotService{client: client, logger: logger.WithField("service", "snapshot")}
}

func (s *DefaultSnapshotService) CreateRepository(ctx context.Context, name string, body any, params *SnapshotCreateRepositoryParams) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_snapshot/%s", pathSegment(name)), body, params, &result)
}

func (s *DefaultSnapshotService) GetRepository(ctx context.Context, names []string, params *SnapshotGetRepositoryParams) (map[string]*SnapshotRepository, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_snapshot/%s", escapedList(names))
	} else {
		path = "/_snapshot"
	}
	var result map[string]*SnapshotRepository
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, params, &result)
}

func (s *DefaultSnapshotService) DeleteRepository(ctx context.Context, names []string, params *SnapshotDeleteRepositoryParams) (*types.AcknowledgedResponse, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("names is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_snapshot/%s", escapedList(names)), nil, params, &result)
}

func (s *DefaultSnapshotService) VerifyRepository(ctx context.Context, name string, params *SnapshotVerifyRepositoryParams) (*SnapshotVerifyRepositoryResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result SnapshotVerifyRepositoryResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_snapshot/%s/_verify", pathSegment(name)), nil, params, &result)
}

func (s *DefaultSnapshotService) RepositoryAnalyze(ctx context.Context, name string, params *SnapshotRepositoryAnalyzeParams) (*SnapshotRepositoryAnalyzeResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result SnapshotRepositoryAnalyzeResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_snapshot/%s/_analyze", pathSegment(name)), nil, params, &result)
}

func (s *DefaultSnapshotService) RepositoryVerifyIntegrity(ctx context.Context, name string, params *SnapshotRepositoryVerifyIntegrityParams) (*SnapshotRepositoryVerifyIntegrityResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result SnapshotRepositoryVerifyIntegrityResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_snapshot/%s/_verify_integrity", pathSegment(name)), nil, params, &result)
}

func (s *DefaultSnapshotService) Create(ctx context.Context, req *SnapshotCreateRequest) (*SnapshotCreateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result SnapshotCreateResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_snapshot/%s/%s", pathSegment(req.Repository), pathSegment(req.Snapshot)), req.Body, req.Params, &result)
}

func (s *DefaultSnapshotService) Get(ctx context.Context, req *SnapshotGetRequest) (*SnapshotGetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var path string
	if len(req.Snapshots) > 0 {
		path = fmt.Sprintf("/_snapshot/%s/%s", pathSegment(req.Repository), escapedList(req.Snapshots))
	} else {
		path = fmt.Sprintf("/_snapshot/%s/_all", pathSegment(req.Repository))
	}
	var result SnapshotGetResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, req.Params, &result)
}

func (s *DefaultSnapshotService) Delete(ctx context.Context, req *SnapshotDeleteRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_snapshot/%s/%s", pathSegment(req.Repository), pathSegment(req.Snapshot)), nil, req.Params, &result)
}

func (s *DefaultSnapshotService) Restore(ctx context.Context, req *SnapshotRestoreRequest) (*SnapshotRestoreResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result SnapshotRestoreResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_snapshot/%s/%s/_restore", pathSegment(req.Repository), pathSegment(req.Snapshot)), req.Body, req.Params, &result)
}

func (s *DefaultSnapshotService) Status(ctx context.Context, req *SnapshotStatusRequest) (*SnapshotStatusResponse, error) {
	var path string
	switch {
	case req.Repository != "" && len(req.Snapshots) > 0:
		path = fmt.Sprintf("/_snapshot/%s/%s/_status", pathSegment(req.Repository), escapedList(req.Snapshots))
	case req.Repository != "":
		path = fmt.Sprintf("/_snapshot/%s/_status", pathSegment(req.Repository))
	default:
		path = "/_snapshot/_status"
	}
	var result SnapshotStatusResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, req.Params, &result)
}

func (s *DefaultSnapshotService) Clone(ctx context.Context, req *SnapshotCloneRequest) (*types.AcknowledgedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_snapshot/%s/%s/_clone/%s", pathSegment(req.Repository), pathSegment(req.Snapshot), pathSegment(req.TargetSnapshot)), req.Body, req.Params, &result)
}

func (s *DefaultSnapshotService) CleanupRepository(ctx context.Context, name string, params *SnapshotCleanupRepositoryParams) (*SnapshotCleanupRepositoryResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result SnapshotCleanupRepositoryResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_snapshot/%s/_cleanup", pathSegment(name)), nil, params, &result)
}
