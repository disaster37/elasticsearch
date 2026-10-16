package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// SearchableSnapshotsMountParams are the query parameters for mount.
type SearchableSnapshotsMountParams struct {
	Common              *CommonParams
	MasterTimeout       string
	Storage             string
	WaitForCompletion   *bool
	WaitForActiveShards string
}

// ToMap converts SearchableSnapshotsMountParams to a query-parameter map.
func (p *SearchableSnapshotsMountParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Storage != "" {
		m["storage"] = p.Storage
	}
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	if p.WaitForActiveShards != "" {
		m["wait_for_active_shards"] = p.WaitForActiveShards
	}
	return mergeParams(m, p.Common)
}

// SearchableSnapshotsStatsParams are the query parameters for stats.
type SearchableSnapshotsStatsParams struct {
	Common *CommonParams
	Level  Level
}

// ToMap converts SearchableSnapshotsStatsParams to a query-parameter map.
func (p *SearchableSnapshotsStatsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Level != "" {
		m["level"] = string(p.Level)
	}
	return mergeParams(m, p.Common)
}

// SearchableSnapshotsService provides access to the searchable snapshots APIs.
type SearchableSnapshotsService interface {
	Mount(ctx context.Context, repository, snapshot string, body any, params *SearchableSnapshotsMountParams) (json.RawMessage, error)
	Stats(ctx context.Context, indices []string, params *SearchableSnapshotsStatsParams) (json.RawMessage, error)
	CacheStats(ctx context.Context, nodeIds []string) (json.RawMessage, error)
	ClearCache(ctx context.Context, indices []string) (*types.AcknowledgedResponse, error)
	RepositoryStats(ctx context.Context, repository string) (json.RawMessage, error)
}

// DefaultSearchableSnapshotsService is the default SearchableSnapshotsService implementation.
type DefaultSearchableSnapshotsService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewSearchableSnapshotsService creates a new SearchableSnapshotsService.
func NewSearchableSnapshotsService(client *resty.Client, logger *logrus.Entry) SearchableSnapshotsService {
	return &DefaultSearchableSnapshotsService{client: client, logger: logger.WithField("service", "searchable_snapshots")}
}

func (s *DefaultSearchableSnapshotsService) Mount(ctx context.Context, repository, snapshot string, body any, params *SearchableSnapshotsMountParams) (json.RawMessage, error) {
	if repository == "" || snapshot == "" {
		return nil, fmt.Errorf("repository and snapshot are required")
	}
	return doReqRaw(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_snapshot/%s/%s/_mount", pathSegment(repository), pathSegment(snapshot)), body, params)
}

func (s *DefaultSearchableSnapshotsService) Stats(ctx context.Context, indices []string, params *SearchableSnapshotsStatsParams) (json.RawMessage, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_searchable_snapshots/stats", escapedList(indices))
	} else {
		path = "/_searchable_snapshots/stats"
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", path, nil, params)
}

func (s *DefaultSearchableSnapshotsService) CacheStats(ctx context.Context, nodeIds []string) (json.RawMessage, error) {
	var path string
	if len(nodeIds) > 0 {
		path = fmt.Sprintf("/_searchable_snapshots/cache/stats/%s", escapedList(nodeIds))
	} else {
		path = "/_searchable_snapshots/cache/stats"
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", path, nil, nil)
}

func (s *DefaultSearchableSnapshotsService) ClearCache(ctx context.Context, indices []string) (*types.AcknowledgedResponse, error) {
	var path string
	if len(indices) > 0 {
		path = fmt.Sprintf("/%s/_searchable_snapshots/cache/clear", escapedList(indices))
	} else {
		path = "/_searchable_snapshots/cache/clear"
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", path, nil, nil, &result)
}

func (s *DefaultSearchableSnapshotsService) RepositoryStats(ctx context.Context, repository string) (json.RawMessage, error) {
	if repository == "" {
		return nil, fmt.Errorf("repository is required")
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_snapshot/%s/_stats", pathSegment(repository)), nil, nil)
}
