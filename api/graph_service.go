package api

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// GraphExploreParams are the query parameters for graph explore.
type GraphExploreParams struct {
	Common  *CommonParams
	Routing string
	Timeout string
}

// ToMap converts GraphExploreParams to a query-parameter map.
func (p *GraphExploreParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Routing != "" {
		m["routing"] = p.Routing
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// GraphExploreResponse is the response from graph explore.
type GraphExploreResponse struct {
	Took        int64            `json:"took"`
	TimeOut     bool             `json:"timed_out"`
	Failures    []map[string]any `json:"failures,omitempty"`
	Vertices    []map[string]any `json:"vertices"`
	Connections []map[string]any `json:"connections"`
}

// GraphService provides access to the graph APIs.
type GraphService interface {
	Explore(ctx context.Context, index string, body any, params *GraphExploreParams) (*GraphExploreResponse, error)
}

// DefaultGraphService is the default GraphService implementation.
type DefaultGraphService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewGraphService creates a new GraphService.
func NewGraphService(client *resty.Client, logger *logrus.Entry) GraphService {
	return &DefaultGraphService{client: client, logger: logger.WithField("service", "graph")}
}

func (s *DefaultGraphService) Explore(ctx context.Context, index string, body any, params *GraphExploreParams) (*GraphExploreResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result GraphExploreResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/%s/_graph/explore", pathSegment(index)), body, params, &result)
}
