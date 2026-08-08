package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitIngestGetPipeline(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"my-pipeline":{"description":"d","version":1,"processors":[{"set":{"field":"f","value":"v"}}],"deprecated":false}}`)
	}))
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	result, err := svc.GetPipeline(context.Background(), []string{"my-pipeline"}, &IngestGetPipelineParams{Summary: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ingest/pipeline/my-pipeline", gotPath)
	assert.Equal(t, "true", gotQuery["summary"])
	require.Contains(t, result, "my-pipeline")
	p := result["my-pipeline"]
	require.NotNil(t, p)
	assert.Equal(t, "d", p.Description)
	assert.Equal(t, 1, p.Version)
	require.Len(t, p.Processors, 1)
	assert.False(t, p.Deprecated)
}

func TestUnitIngestGetPipelineAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	_, err := svc.GetPipeline(context.Background(), nil, &IngestGetPipelineParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_ingest/pipeline", gotPath)
}

func TestUnitIngestGetPipelineError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	_, err := svc.GetPipeline(context.Background(), nil, &IngestGetPipelineParams{})
	assert.Error(t, err)
}

func TestUnitIngestGetPipelineNetworkError(t *testing.T) {
	svc := NewIngestService(deadClient(), testLogger())
	_, err := svc.GetPipeline(context.Background(), nil, &IngestGetPipelineParams{})
	assert.Error(t, err)
}

func TestUnitIngestPutPipeline(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	resp, err := svc.PutPipeline(context.Background(), &IngestPutPipelineRequest{
		Id:   "my-pipeline",
		Body: map[string]any{"description": "d", "processors": []map[string]any{{"set": map[string]any{"field": "f", "value": "v"}}}},
		Params: &IngestPutPipelineParams{IfVersion: ptrInt(2), Timeout: "10s"},
	})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_ingest/pipeline/my-pipeline", gotPath)
	assert.Equal(t, "2", gotQuery["if_version"])
	assert.Equal(t, "10s", gotQuery["timeout"])
	assert.Contains(t, gotBody, "processors")
	assert.True(t, resp.Acknowledged)
}

func TestUnitIngestPutPipelineValidation(t *testing.T) {
	svc := NewIngestService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	// missing id
	_, err := svc.PutPipeline(context.Background(), &IngestPutPipelineRequest{Body: map[string]any{"x": 1}})
	assert.Error(t, err)
	// missing body
	_, err = svc.PutPipeline(context.Background(), &IngestPutPipelineRequest{Id: "id"})
	assert.Error(t, err)
}

func TestUnitIngestPutPipelineError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	_, err := svc.PutPipeline(context.Background(), &IngestPutPipelineRequest{Id: "id", Body: map[string]any{"x": 1}})
	assert.Error(t, err)
}

func TestUnitIngestPutPipelineNetworkError(t *testing.T) {
	svc := NewIngestService(deadClient(), testLogger())
	_, err := svc.PutPipeline(context.Background(), &IngestPutPipelineRequest{Id: "id", Body: map[string]any{"x": 1}})
	assert.Error(t, err)
}

func TestUnitIngestDeletePipeline(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	resp, err := svc.DeletePipeline(context.Background(), "my-pipeline", &IngestDeletePipelineParams{Timeout: "5s"})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_ingest/pipeline/my-pipeline", gotPath)
	assert.Equal(t, "5s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitIngestDeletePipelineEmptyID(t *testing.T) {
	svc := NewIngestService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeletePipeline(context.Background(), "", &IngestDeletePipelineParams{})
	assert.Error(t, err)
}

func TestUnitIngestDeletePipelineError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	_, err := svc.DeletePipeline(context.Background(), "id", &IngestDeletePipelineParams{})
	assert.Error(t, err)
}

func TestUnitIngestDeletePipelineNetworkError(t *testing.T) {
	svc := NewIngestService(deadClient(), testLogger())
	_, err := svc.DeletePipeline(context.Background(), "id", &IngestDeletePipelineParams{})
	assert.Error(t, err)
}

func TestUnitIngestSimulate(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"docs":[{"doc":{"_index":"idx","_source":{"f":"v"}},"processor_results":[{"processor_type":{"name":"set"},"status":"success","doc":{"_index":"idx","_source":{"f":"v"}}}]}]}`)
	}))
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	resp, err := svc.Simulate(context.Background(), &IngestSimulatePipelineRequest{
		Id:   "my-pipeline",
		Body: map[string]any{"docs": []map[string]any{{"_source": map[string]any{"f": "v"}}}},
		Params: &IngestSimulatePipelineParams{Verbose: ptrBool(true)},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ingest/pipeline/my-pipeline/_simulate", gotPath)
	assert.Equal(t, "true", gotQuery["verbose"])
	assert.Contains(t, gotBody, "docs")
	require.Len(t, resp.Docs, 1)
	require.Len(t, resp.Docs[0].ProcessorResults, 1)
	assert.Equal(t, "set", resp.Docs[0].ProcessorResults[0].Processor["name"])
	assert.Equal(t, "success", resp.Docs[0].ProcessorResults[0].Status)
}

func TestUnitIngestSimulateNoID(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"docs":[]}`)
	}))
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	_, err := svc.Simulate(context.Background(), &IngestSimulatePipelineRequest{Body: map[string]any{"docs": []map[string]any{}}})
	require.NoError(t, err)
	assert.Equal(t, "/_ingest/pipeline/_simulate", gotPath)
}

func TestUnitIngestSimulateValidation(t *testing.T) {
	svc := NewIngestService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	// missing body
	_, err := svc.Simulate(context.Background(), &IngestSimulatePipelineRequest{Id: "id"})
	assert.Error(t, err)
}

func TestUnitIngestSimulateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	_, err := svc.Simulate(context.Background(), &IngestSimulatePipelineRequest{Body: map[string]any{"x": 1}})
	assert.Error(t, err)
}

func TestUnitIngestSimulateNetworkError(t *testing.T) {
	svc := NewIngestService(deadClient(), testLogger())
	_, err := svc.Simulate(context.Background(), &IngestSimulatePipelineRequest{Body: map[string]any{"x": 1}})
	assert.Error(t, err)
}

func TestUnitIngestProcessorGrok(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"patterns":{"LOGLEVEL":"[A-Z]+"}}`)
	}))
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	resp, err := svc.ProcessorGrok(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ingest/processor/grok", gotPath)
	require.Contains(t, resp.Patterns, "LOGLEVEL")
	assert.Equal(t, "[A-Z]+", resp.Patterns["LOGLEVEL"])
}

func TestUnitIngestProcessorGrokError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIngestService(restyClient(srv), testLogger())
	_, err := svc.ProcessorGrok(context.Background())
	assert.Error(t, err)
}

func TestUnitIngestProcessorGrokNetworkError(t *testing.T) {
	svc := NewIngestService(deadClient(), testLogger())
	_, err := svc.ProcessorGrok(context.Background())
	assert.Error(t, err)
}
