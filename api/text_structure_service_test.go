package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitTextStructureFindStructure(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"format":"ndjson","charset":"UTF-8","has_header_row":true,"num_lines_analyzed":10}`)
	}))
	defer srv.Close()
	svc := NewTextStructureService(restyClient(srv), testLogger())
	resp, err := svc.FindStructure(context.Background(), map[string]any{"text": "sample"}, &TextStructureFindParams{
		Format:        "ndjson",
		Charset:       "UTF-8",
		HasHeaderRow:  ptrBool(true),
		LinesToSample: ptrInt(10),
		Explain:       ptrBool(false),
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_text_structure/find_structure", gotPath)
	assert.Equal(t, "ndjson", gotQuery["format"])
	assert.Equal(t, "UTF-8", gotQuery["charset"])
	assert.Equal(t, "true", gotQuery["has_header_row"])
	assert.Equal(t, "10", gotQuery["lines_to_sample"])
	assert.Equal(t, "false", gotQuery["explain"])
	assert.Contains(t, gotBody, "sample")
	assert.Equal(t, "ndjson", resp.Format)
	assert.Equal(t, "UTF-8", resp.Charset)
	assert.True(t, resp.HasHeaderRow)
	assert.Equal(t, 10, resp.NumLinesAnalyzed)
}

func TestUnitTextStructureFindStructureError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTextStructureService(restyClient(srv), testLogger())
	_, err := svc.FindStructure(context.Background(), map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitTextStructureFindStructureNetworkError(t *testing.T) {
	svc := NewTextStructureService(deadClient(), testLogger())
	_, err := svc.FindStructure(context.Background(), map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitTextStructureFindFieldStructure(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"field":"msg","format":"semi_structured_text"}`)
	}))
	defer srv.Close()
	svc := NewTextStructureService(restyClient(srv), testLogger())
	resp, err := svc.FindFieldStructure(context.Background(), "idx", "msg", &TextStructureFindParams{
		Format: "semi_structured_text",
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/idx/_text_structure/field_structure", gotPath)
	assert.Equal(t, "msg", gotQuery["field"])
	assert.Equal(t, "semi_structured_text", gotQuery["format"])
	assert.Contains(t, string(resp), "field")
}

func TestUnitTextStructureFindFieldStructureEmptyIndex(t *testing.T) {
	svc := NewTextStructureService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.FindFieldStructure(context.Background(), "", "msg", nil)
	assert.Error(t, err)
}

func TestUnitTextStructureFindFieldStructureEmptyField(t *testing.T) {
	svc := NewTextStructureService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.FindFieldStructure(context.Background(), "idx", "", nil)
	assert.Error(t, err)
}

func TestUnitTextStructureFindFieldStructureError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTextStructureService(restyClient(srv), testLogger())
	_, err := svc.FindFieldStructure(context.Background(), "idx", "msg", nil)
	assert.Error(t, err)
}

func TestUnitTextStructureFindFieldStructureNetworkError(t *testing.T) {
	svc := NewTextStructureService(deadClient(), testLogger())
	_, err := svc.FindFieldStructure(context.Background(), "idx", "msg", nil)
	assert.Error(t, err)
}

func TestUnitTextStructureFindMessageStructure(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"format":"logstash"}`)
	}))
	defer srv.Close()
	svc := NewTextStructureService(restyClient(srv), testLogger())
	resp, err := svc.FindMessageStructure(context.Background(), "idx", &TextStructureFindParams{
		Format: "logstash",
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/idx/_text_structure/message_structure", gotPath)
	assert.Equal(t, "logstash", gotQuery["format"])
	assert.Contains(t, string(resp), "logstash")
}

func TestUnitTextStructureFindMessageStructureEmptyIndex(t *testing.T) {
	svc := NewTextStructureService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.FindMessageStructure(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestUnitTextStructureFindMessageStructureError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTextStructureService(restyClient(srv), testLogger())
	_, err := svc.FindMessageStructure(context.Background(), "idx", nil)
	assert.Error(t, err)
}

func TestUnitTextStructureFindMessageStructureNetworkError(t *testing.T) {
	svc := NewTextStructureService(deadClient(), testLogger())
	_, err := svc.FindMessageStructure(context.Background(), "idx", nil)
	assert.Error(t, err)
}

func TestUnitTextStructureTestGrokPattern(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"matches":[]}`)
	}))
	defer srv.Close()
	svc := NewTextStructureService(restyClient(srv), testLogger())
	resp, err := svc.TestGrokPattern(context.Background(), map[string]any{"grok_pattern": "%{IP:ip}"}, &TextStructureFindParams{
		GrokPattern: "%{IP:ip}",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_text_structure/test_grok_pattern", gotPath)
	assert.Equal(t, "%{IP:ip}", gotQuery["grok_pattern"])
	assert.Contains(t, gotBody, "grok_pattern")
	assert.Contains(t, string(resp), "matches")
}

func TestUnitTextStructureTestGrokPatternError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTextStructureService(restyClient(srv), testLogger())
	_, err := svc.TestGrokPattern(context.Background(), map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitTextStructureTestGrokPatternNetworkError(t *testing.T) {
	svc := NewTextStructureService(deadClient(), testLogger())
	_, err := svc.TestGrokPattern(context.Background(), map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitTextStructureFindParamsToMap(t *testing.T) {
	p := &TextStructureFindParams{
		DocumentsToSample: ptrInt(500),
		EcsCompatibility:  "v1",
		Timeout:           "25s",
	}
	m := p.ToMap()
	assert.Equal(t, "500", m["documents_to_sample"])
	assert.Equal(t, "v1", m["ecs_compatibility"])
	assert.Equal(t, "25s", m["timeout"])
}
