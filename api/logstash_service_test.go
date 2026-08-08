package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitLogstashGetPipeline(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"my-pipeline":{"description":"test"}}`)
	}))
	defer srv.Close()
	svc := NewLogstashService(restyClient(srv), testLogger())
	resp, err := svc.GetPipeline(context.Background(), []string{"my-pipeline"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_logstash/pipeline/my-pipeline", gotPath)
	assert.Contains(t, resp["my-pipeline"]["description"], "test")
}

func TestUnitLogstashGetPipelineMultiple(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewLogstashService(restyClient(srv), testLogger())
	_, err := svc.GetPipeline(context.Background(), []string{"a", "b"})
	require.NoError(t, err)
	assert.Equal(t, "/_logstash/pipeline/a,b", gotPath)
}

func TestUnitLogstashGetPipelineAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewLogstashService(restyClient(srv), testLogger())
	_, err := svc.GetPipeline(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_logstash/pipeline", gotPath)
}

func TestUnitLogstashGetPipelineError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewLogstashService(restyClient(srv), testLogger())
	_, err := svc.GetPipeline(context.Background(), []string{"my-pipeline"})
	assert.Error(t, err)
}

func TestUnitLogstashGetPipelineNetworkError(t *testing.T) {
	svc := NewLogstashService(deadClient(), testLogger())
	_, err := svc.GetPipeline(context.Background(), []string{"my-pipeline"})
	assert.Error(t, err)
}

func TestUnitLogstashPutPipeline(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewLogstashService(restyClient(srv), testLogger())
	resp, err := svc.PutPipeline(context.Background(), "my-pipeline", map[string]any{"description": "test"})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_logstash/pipeline/my-pipeline", gotPath)
	assert.Contains(t, gotBody, "description")
	assert.True(t, resp.Acknowledged)
}

func TestUnitLogstashPutPipelineEmptyID(t *testing.T) {
	svc := NewLogstashService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutPipeline(context.Background(), "", map[string]any{})
	assert.Error(t, err)
}

func TestUnitLogstashPutPipelineError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewLogstashService(restyClient(srv), testLogger())
	_, err := svc.PutPipeline(context.Background(), "my-pipeline", map[string]any{})
	assert.Error(t, err)
}

func TestUnitLogstashPutPipelineNetworkError(t *testing.T) {
	svc := NewLogstashService(deadClient(), testLogger())
	_, err := svc.PutPipeline(context.Background(), "my-pipeline", map[string]any{})
	assert.Error(t, err)
}

func TestUnitLogstashDeletePipeline(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewLogstashService(restyClient(srv), testLogger())
	resp, err := svc.DeletePipeline(context.Background(), "my-pipeline")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_logstash/pipeline/my-pipeline", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitLogstashDeletePipelineEmptyID(t *testing.T) {
	svc := NewLogstashService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeletePipeline(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitLogstashDeletePipelineError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewLogstashService(restyClient(srv), testLogger())
	_, err := svc.DeletePipeline(context.Background(), "my-pipeline")
	assert.Error(t, err)
}

func TestUnitLogstashDeletePipelineNetworkError(t *testing.T) {
	svc := NewLogstashService(deadClient(), testLogger())
	_, err := svc.DeletePipeline(context.Background(), "my-pipeline")
	assert.Error(t, err)
}
