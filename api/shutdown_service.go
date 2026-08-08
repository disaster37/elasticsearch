package api

import (
	"context"
	"fmt"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// ShutdownGetNodeResponse is the response from GET /_nodes/shutdown.
type ShutdownGetNodeResponse struct {
	Nodes []struct {
		NodeID            string         `json:"node_id"`
		Type              string         `json:"type"`
		Reason            string         `json:"reason"`
		ShutdownStarted   string         `json:"shutdown_started,omitempty"`
		ShutdownCompleted string         `json:"shutdown_completed,omitempty"`
		AllocationDelay   string         `json:"allocation_delay,omitempty"`
		TargetNodeName    string         `json:"target_node_name,omitempty"`
		Status            map[string]any `json:"status,omitempty"`
	} `json:"nodes"`
}

// ShutdownService provides access to the shutdown APIs.
type ShutdownService interface {
	GetNode(ctx context.Context, nodeId string) (*ShutdownGetNodeResponse, error)
	PutNode(ctx context.Context, nodeId string, body any) (*types.AcknowledgedResponse, error)
	DeleteNode(ctx context.Context, nodeId string) (*types.AcknowledgedResponse, error)
}

// DefaultShutdownService is the default ShutdownService implementation.
type DefaultShutdownService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewShutdownService creates a new ShutdownService.
func NewShutdownService(client *resty.Client, logger *logrus.Entry) ShutdownService {
	return &DefaultShutdownService{client: client, logger: logger.WithField("service", "shutdown")}
}

func (s *DefaultShutdownService) GetNode(ctx context.Context, nodeId string) (*ShutdownGetNodeResponse, error) {
	var path string
	if nodeId != "" {
		path = fmt.Sprintf("/_nodes/%s/shutdown", pathSegment(nodeId))
	} else {
		path = "/_nodes/shutdown"
	}
	var result ShutdownGetNodeResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultShutdownService) PutNode(ctx context.Context, nodeId string, body any) (*types.AcknowledgedResponse, error) {
	if nodeId == "" {
		return nil, fmt.Errorf("node_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_nodes/%s/shutdown", pathSegment(nodeId)), body, nil, &result)
}

func (s *DefaultShutdownService) DeleteNode(ctx context.Context, nodeId string) (*types.AcknowledgedResponse, error) {
	if nodeId == "" {
		return nil, fmt.Errorf("node_id is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_nodes/%s/shutdown", pathSegment(nodeId)), nil, nil, &result)
}
