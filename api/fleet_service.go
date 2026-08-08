package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// FleetGlobalCheckpointsParams are the query parameters for fleet global checkpoints.
type FleetGlobalCheckpointsParams struct {
	Common         *CommonParams
	Checkpoints    []string
	WaitForAdvance *bool
	WaitForIndex   *bool
	Timeout        string
}

// ToMap converts FleetGlobalCheckpointsParams to a query-parameter map.
func (p *FleetGlobalCheckpointsParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if len(p.Checkpoints) > 0 {
		// Query-parameter value: resty URL-encodes it, so use a plain join.
		m["checkpoints"] = strings.Join(p.Checkpoints, ",")
	}
	setBoolParam(m, "wait_for_advance", p.WaitForAdvance)
	setBoolParam(m, "wait_for_index", p.WaitForIndex)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// FleetGlobalCheckpointsResponse is the response from fleet global checkpoints.
type FleetGlobalCheckpointsResponse struct {
	GlobalCheckpoints []int64 `json:"global_checkpoints"`
	TimedOut          bool    `json:"timed_out"`
}

// FleetService provides access to the fleet APIs.
type FleetService interface {
	GlobalCheckpoints(ctx context.Context, index string, params *FleetGlobalCheckpointsParams) (*FleetGlobalCheckpointsResponse, error)
}

// DefaultFleetService is the default FleetService implementation.
type DefaultFleetService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewFleetService creates a new FleetService.
func NewFleetService(client *resty.Client, logger *logrus.Entry) FleetService {
	return &DefaultFleetService{client: client, logger: logger.WithField("service", "fleet")}
}

func (s *DefaultFleetService) GlobalCheckpoints(ctx context.Context, index string, params *FleetGlobalCheckpointsParams) (*FleetGlobalCheckpointsResponse, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	var result FleetGlobalCheckpointsResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/%s/_fleet/global_checkpoints", pathSegment(index)), nil, params, &result)
}
