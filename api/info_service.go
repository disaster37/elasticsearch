package api

import (
	"context"
	"strconv"
	"strings"

	json "github.com/goccy/go-json"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// InfoService provides access to cluster-level info, ping, and capabilities
// endpoints.
type InfoService interface {
	Info(ctx context.Context) (*InfoResponse, error)
	Ping(ctx context.Context) (bool, error)
	Capabilities(ctx context.Context, req *CapabilitiesRequest) (*CapabilitiesResponse, error)
}

// DefaultInfoService is the default InfoService implementation.
type DefaultInfoService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewInfoService creates a new InfoService.
func NewInfoService(client *resty.Client, logger *logrus.Entry) InfoService {
	return &DefaultInfoService{client: client, logger: logger.WithField("service", "info")}
}

func (s *DefaultInfoService) Info(ctx context.Context) (*InfoResponse, error) {
	var result InfoResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/", nil, nil, &result)
}

func (s *DefaultInfoService) Ping(ctx context.Context) (bool, error) {
	return headReq(s.logger, ctx, s.client, "/", nil)
}

func (s *DefaultInfoService) Capabilities(ctx context.Context, req *CapabilitiesRequest) (*CapabilitiesResponse, error) {
	r := s.client.R().SetContext(ctx)
	if req != nil {
		if req.Method != "" {
			r.SetQueryParam("method", req.Method)
		}
		if req.Path != "" {
			r.SetQueryParam("path", req.Path)
		}
		if len(req.Parameters) > 0 {
			// Query-parameter value: resty URL-encodes it, so use a plain
			// join (do not pre-escape with pathSegment/escapedList).
			r.SetQueryParam("parameters", strings.Join(req.Parameters, ","))
		}
		if req.Local != nil {
			r.SetQueryParam("local", strconv.FormatBool(*req.Local))
		}
		m := mergeParams(nil, req.Common)
		for k, v := range m {
			r.SetQueryParam(k, v)
		}
	}
	resp, err := r.Get("/_capabilities")
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	var result CapabilitiesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, wrapUnmarshalError(s.logger, resp, err)
	}
	return &result, nil
}
