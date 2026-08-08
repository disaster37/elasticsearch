package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitInferenceGet(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"inference_id":"i1","task_type":"completion","service":"openai"}`)
	}))
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), "i1", "completion")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_inference/completion/i1", gotPath)
	assert.Contains(t, string(resp), "inference_id")
}

func TestUnitInferenceGetInferenceIdOnly(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"inference_id":"i1"}`)
	}))
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "i1", "")
	require.NoError(t, err)
	assert.Equal(t, "/_inference/i1", gotPath)
}

func TestUnitInferenceGetAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `[]`)
	}))
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "", "")
	require.NoError(t, err)
	assert.Equal(t, "/_inference", gotPath)
}

func TestUnitInferenceGetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "i1", "completion")
	assert.Error(t, err)
}

func TestUnitInferenceGetNetworkError(t *testing.T) {
	svc := NewInferenceService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), "i1", "completion")
	assert.Error(t, err)
}

func TestUnitInferencePut(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	resp, err := svc.Put(context.Background(), "completion", "i1", map[string]any{"service": "openai"})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_inference/completion/i1", gotPath)
	assert.Contains(t, gotBody, "openai")
	assert.True(t, resp.Acknowledged)
}

func TestUnitInferencePutEmptyTaskType(t *testing.T) {
	svc := NewInferenceService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Put(context.Background(), "", "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferencePutEmptyInferenceId(t *testing.T) {
	svc := NewInferenceService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Put(context.Background(), "completion", "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferencePutError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	_, err := svc.Put(context.Background(), "completion", "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferencePutNetworkError(t *testing.T) {
	svc := NewInferenceService(deadClient(), testLogger())
	_, err := svc.Put(context.Background(), "completion", "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceDelete(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background(), "completion", "i1", &InferenceDeleteParams{
		Force:  ptrBool(true),
		DryRun: ptrBool(false),
	})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_inference/completion/i1", gotPath)
	assert.Equal(t, "true", gotQuery["force"])
	assert.Equal(t, "false", gotQuery["dry_run"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitInferenceDeleteInferenceIdOnly(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	_, err := svc.Delete(context.Background(), "", "i1", nil)
	require.NoError(t, err)
	assert.Equal(t, "/_inference/i1", gotPath)
}

func TestUnitInferenceDeleteEmpty(t *testing.T) {
	svc := NewInferenceService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Delete(context.Background(), "", "", nil)
	assert.Error(t, err)
}

func TestUnitInferenceDeleteError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	_, err := svc.Delete(context.Background(), "completion", "i1", nil)
	assert.Error(t, err)
}

func TestUnitInferenceDeleteNetworkError(t *testing.T) {
	svc := NewInferenceService(deadClient(), testLogger())
	_, err := svc.Delete(context.Background(), "completion", "i1", nil)
	assert.Error(t, err)
}

func TestUnitInferenceUpdate(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	resp, err := svc.Update(context.Background(), "i1", map[string]any{"task_settings": map[string]any{}})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_inference/i1/_update", gotPath)
	assert.Contains(t, gotBody, "task_settings")
	assert.True(t, resp.Acknowledged)
}

func TestUnitInferenceUpdateEmpty(t *testing.T) {
	svc := NewInferenceService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Update(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceUpdateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	_, err := svc.Update(context.Background(), "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceUpdateNetworkError(t *testing.T) {
	svc := NewInferenceService(deadClient(), testLogger())
	_, err := svc.Update(context.Background(), "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceInference(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"completion":[{"result":"ok"}]}`)
	}))
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	resp, err := svc.Inference(context.Background(), "i1", map[string]any{"input": "hello"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_inference/i1", gotPath)
	assert.Contains(t, gotBody, "hello")
	assert.Contains(t, string(resp), "completion")
}

func TestUnitInferenceInferenceEmpty(t *testing.T) {
	svc := NewInferenceService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Inference(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceInferenceError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	_, err := svc.Inference(context.Background(), "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceInferenceNetworkError(t *testing.T) {
	svc := NewInferenceService(deadClient(), testLogger())
	_, err := svc.Inference(context.Background(), "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceCompletion(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"completion":[{"result":"ok"}]}`)
	}))
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	resp, err := svc.Completion(context.Background(), "i1", map[string]any{"input": "hello"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_inference/completion/i1", gotPath)
	assert.Contains(t, gotBody, "hello")
	assert.Contains(t, string(resp), "completion")
}

func TestUnitInferenceCompletionEmpty(t *testing.T) {
	svc := NewInferenceService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Completion(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceCompletionError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	_, err := svc.Completion(context.Background(), "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceCompletionNetworkError(t *testing.T) {
	svc := NewInferenceService(deadClient(), testLogger())
	_, err := svc.Completion(context.Background(), "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceRerank(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"rerank":[{"index":0,"relevance_score":0.9}]}`)
	}))
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	resp, err := svc.Rerank(context.Background(), "i1", map[string]any{"query": "hello"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_inference/rerank/i1", gotPath)
	assert.Contains(t, gotBody, "hello")
	assert.Contains(t, string(resp), "rerank")
}

func TestUnitInferenceRerankEmpty(t *testing.T) {
	svc := NewInferenceService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Rerank(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceRerankError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewInferenceService(restyClient(srv), testLogger())
	_, err := svc.Rerank(context.Background(), "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitInferenceRerankNetworkError(t *testing.T) {
	svc := NewInferenceService(deadClient(), testLogger())
	_, err := svc.Rerank(context.Background(), "i1", map[string]any{"x": 1})
	assert.Error(t, err)
}
