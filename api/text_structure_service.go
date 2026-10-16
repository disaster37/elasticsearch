package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// TextStructureFindParams are the shared query parameters for text structure endpoints.
type TextStructureFindParams struct {
	Common             *CommonParams
	Charset            string
	ColumnNames        []string
	Delimiter          string
	DocumentsToSample  *int
	EcsCompatibility   string
	Explain            *bool
	Format             string
	GrokPattern        string
	HasHeaderRow       *bool
	LineMergeSizeLimit *int
	LinesToSample      *int
	Quote              string
	ShouldTrimFields   *bool
	Timeout            string
	TimestampField     string
	TimestampFormat    string
}

// ToMap converts TextStructureFindParams to a query-parameter map.
func (p *TextStructureFindParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Charset != "" {
		m["charset"] = p.Charset
	}
	setListParam(m, "column_names", p.ColumnNames)
	if p.Delimiter != "" {
		m["delimiter"] = p.Delimiter
	}
	if p.DocumentsToSample != nil {
		m["documents_to_sample"] = intStr(*p.DocumentsToSample)
	}
	if p.EcsCompatibility != "" {
		m["ecs_compatibility"] = p.EcsCompatibility
	}
	setBoolParam(m, "explain", p.Explain)
	if p.Format != "" {
		m["format"] = p.Format
	}
	if p.GrokPattern != "" {
		m["grok_pattern"] = p.GrokPattern
	}
	setBoolParam(m, "has_header_row", p.HasHeaderRow)
	if p.LineMergeSizeLimit != nil {
		m["line_merge_size_limit"] = intStr(*p.LineMergeSizeLimit)
	}
	if p.LinesToSample != nil {
		m["lines_to_sample"] = intStr(*p.LinesToSample)
	}
	if p.Quote != "" {
		m["quote"] = p.Quote
	}
	setBoolParam(m, "should_trim_fields", p.ShouldTrimFields)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	if p.TimestampField != "" {
		m["timestamp_field"] = p.TimestampField
	}
	if p.TimestampFormat != "" {
		m["timestamp_format"] = p.TimestampFormat
	}
	return mergeParams(m, p.Common)
}

// TextStructureResponse is the response from find_structure.
type TextStructureResponse struct {
	NDJSON              bool           `json:"ndjson"`
	NeedClientTimezone  bool           `json:"need_client_timezone"`
	HasHeaderRow        bool           `json:"has_header_row"`
	HasByteOrderMarker  bool           `json:"has_byte_order_marker"`
	LinesMerged         int            `json:"lines_merged"`
	Format              string         `json:"format"`
	Charset             string         `json:"charset"`
	Delimiter           string         `json:"delimiter,omitempty"`
	Quote               string         `json:"quote,omitempty"`
	OriginalQuote       string         `json:"original_quote,omitempty"`
	ColumnNames         []string       `json:"column_names,omitempty"`
	GrokPattern         string         `json:"grok_pattern,omitempty"`
	TimestampField      string         `json:"timestamp_field,omitempty"`
	TimestampFormat     string         `json:"timestamp_format,omitempty"`
	ShouldTrimFields    bool           `json:"should_trim_fields"`
	Mappings            map[string]any `json:"mappings,omitempty"`
	IngestPipeline      map[string]any `json:"ingest_pipeline,omitempty"`
	Explanation         []string       `json:"explanation,omitempty"`
	SampleStart         string         `json:"sample_start,omitempty"`
	SampleSize          int            `json:"sample_size,omitempty"`
	NumLinesAnalyzed    int            `json:"num_lines_analyzed,omitempty"`
	NumMessagesAnalyzed int            `json:"num_messages_analyzed,omitempty"`
}

// TextStructureService provides access to the text structure APIs.
type TextStructureService interface {
	FindStructure(ctx context.Context, body any, params *TextStructureFindParams) (*TextStructureResponse, error)
	FindFieldStructure(ctx context.Context, index, field string, params *TextStructureFindParams) (json.RawMessage, error)
	FindMessageStructure(ctx context.Context, index string, params *TextStructureFindParams) (json.RawMessage, error)
	TestGrokPattern(ctx context.Context, body any, params *TextStructureFindParams) (json.RawMessage, error)
}

// DefaultTextStructureService is the default TextStructureService implementation.
type DefaultTextStructureService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewTextStructureService creates a new TextStructureService.
func NewTextStructureService(client *resty.Client, logger *logrus.Entry) TextStructureService {
	return &DefaultTextStructureService{client: client, logger: logger.WithField("service", "text_structure")}
}

func (s *DefaultTextStructureService) FindStructure(ctx context.Context, body any, params *TextStructureFindParams) (*TextStructureResponse, error) {
	var result TextStructureResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_text_structure/find_structure", body, params, &result)
}

func (s *DefaultTextStructureService) FindFieldStructure(ctx context.Context, index, field string, params *TextStructureFindParams) (json.RawMessage, error) {
	if index == "" || field == "" {
		return nil, fmt.Errorf("index and field are required")
	}
	r := s.client.R().SetContext(ctx).SetQueryParam("field", field)
	for k, v := range params.ToMap() {
		r.SetQueryParam(k, v)
	}
	resp, err := r.Get(fmt.Sprintf("/%s/_text_structure/field_structure", pathSegment(index)))
	if err != nil {
		return nil, wrapNetworkError(s.logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(s.logger, resp)
	}
	return json.RawMessage(resp.Body()), nil
}

func (s *DefaultTextStructureService) FindMessageStructure(ctx context.Context, index string, params *TextStructureFindParams) (json.RawMessage, error) {
	if index == "" {
		return nil, fmt.Errorf("index is required")
	}
	return doReqRaw(s.logger, ctx, s.client, "GET", fmt.Sprintf("/%s/_text_structure/message_structure", pathSegment(index)), nil, params)
}

func (s *DefaultTextStructureService) TestGrokPattern(ctx context.Context, body any, params *TextStructureFindParams) (json.RawMessage, error) {
	return doReqRaw(s.logger, ctx, s.client, "POST", "/_text_structure/test_grok_pattern", body, params)
}
