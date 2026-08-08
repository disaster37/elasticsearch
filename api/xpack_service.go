package api

import (
	"context"

	"github.com/go-resty/resty/v2"
	json "github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
)

// XpackInfoParams are the query parameters for GET /_xpack.
type XpackInfoParams struct {
	Common     *CommonParams
	Categories []string
}

// ToMap converts XpackInfoParams to a query-parameter map.
func (p *XpackInfoParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setListParam(m, "categories", p.Categories)
	return mergeParams(m, p.Common)
}

// XpackInfoResponse is the response from GET /_xpack.
type XpackInfoResponse struct {
	Build    map[string]any          `json:"build"`
	License  map[string]any          `json:"license"`
	Features map[string]XpackFeature `json:"features"`
	Tagline  string                  `json:"tagline"`
}

// XpackFeature describes an X-Pack feature.
type XpackFeature struct {
	Available    bool     `json:"available"`
	Enabled      bool     `json:"enabled"`
	NativeRealms []string `json:"native_realms,omitempty"`
	Description  string   `json:"description,omitempty"`
}

// XpackUsageResponse is the response from GET /_xpack/usage.
type XpackUsageResponse map[string]XpackUsage

// XpackUsage describes usage for an X-Pack feature. Feature-specific fields
// (which vary per feature) are collected into Data by the custom
// UnmarshalJSON implementation.
type XpackUsage struct {
	Available bool `json:"available"`
	Enabled   bool `json:"enabled"`
	// Data holds every feature-specific field that is neither available nor
	// enabled.
	Data map[string]any `json:"-"`
}

// UnmarshalJSON decodes an X-Pack usage object, capturing the known
// available/enabled flags and collecting all remaining feature-specific
// fields into Data.
func (u *XpackUsage) UnmarshalJSON(b []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["available"]; ok {
		_ = json.Unmarshal(v, &u.Available)
	}
	if v, ok := raw["enabled"]; ok {
		_ = json.Unmarshal(v, &u.Enabled)
	}
	data := make(map[string]any, len(raw))
	for k, v := range raw {
		if k == "available" || k == "enabled" {
			continue
		}
		var val any
		if err := json.Unmarshal(v, &val); err == nil {
			data[k] = val
		}
	}
	if len(data) > 0 {
		u.Data = data
	}
	return nil
}

// XpackService provides access to the X-Pack info and usage APIs.
type XpackService interface {
	Info(ctx context.Context, params *XpackInfoParams) (*XpackInfoResponse, error)
	Usage(ctx context.Context) (XpackUsageResponse, error)
}

// DefaultXpackService is the default XpackService implementation.
type DefaultXpackService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewXpackService creates a new XpackService.
func NewXpackService(client *resty.Client, logger *logrus.Entry) XpackService {
	return &DefaultXpackService{client: client, logger: logger.WithField("service", "xpack")}
}

func (s *DefaultXpackService) Info(ctx context.Context, params *XpackInfoParams) (*XpackInfoResponse, error) {
	var result XpackInfoResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_xpack", nil, params, &result)
}

func (s *DefaultXpackService) Usage(ctx context.Context) (XpackUsageResponse, error) {
	var result XpackUsageResponse
	return result, doReq(s.logger, ctx, s.client, "GET", "/_xpack/usage", nil, nil, &result)
}
