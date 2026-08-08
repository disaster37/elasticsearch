package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitEnrichGetPolicy(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"policies":[]}`)
	}))
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	raw, err := svc.GetPolicy(context.Background(), []string{"my-policy"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_enrich/policy/my-policy", gotPath)
	assert.Contains(t, string(raw), "policies")
}

func TestUnitEnrichGetPolicyAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"policies":[]}`)
	}))
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	_, err := svc.GetPolicy(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_enrich/policy", gotPath)
}

func TestUnitEnrichGetPolicyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	_, err := svc.GetPolicy(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitEnrichGetPolicyNetworkError(t *testing.T) {
	svc := NewEnrichService(deadClient(), testLogger())
	_, err := svc.GetPolicy(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitEnrichPutPolicy(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	resp, err := svc.PutPolicy(context.Background(), "my-policy", map[string]any{"match": map[string]any{"indices": "source"}})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_enrich/policy/my-policy", gotPath)
	assert.Contains(t, gotBody, "match")
	assert.True(t, resp.Acknowledged)
}

func TestUnitEnrichPutPolicyEmptyName(t *testing.T) {
	svc := NewEnrichService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutPolicy(context.Background(), "", map[string]any{})
	assert.Error(t, err)
}

func TestUnitEnrichPutPolicyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	_, err := svc.PutPolicy(context.Background(), "my-policy", map[string]any{})
	assert.Error(t, err)
}

func TestUnitEnrichPutPolicyNetworkError(t *testing.T) {
	svc := NewEnrichService(deadClient(), testLogger())
	_, err := svc.PutPolicy(context.Background(), "my-policy", map[string]any{})
	assert.Error(t, err)
}

func TestUnitEnrichDeletePolicy(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	resp, err := svc.DeletePolicy(context.Background(), "my-policy")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_enrich/policy/my-policy", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitEnrichDeletePolicyEmptyName(t *testing.T) {
	svc := NewEnrichService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeletePolicy(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitEnrichDeletePolicyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	_, err := svc.DeletePolicy(context.Background(), "my-policy")
	assert.Error(t, err)
}

func TestUnitEnrichDeletePolicyNetworkError(t *testing.T) {
	svc := NewEnrichService(deadClient(), testLogger())
	_, err := svc.DeletePolicy(context.Background(), "my-policy")
	assert.Error(t, err)
}

func TestUnitEnrichExecutePolicy(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	resp, err := svc.ExecutePolicy(context.Background(), "my-policy", &EnrichWaitForCompletionParams{
		WaitForCompletion: ptrBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_enrich/policy/my-policy/_execute", gotPath)
	assert.Equal(t, "true", gotQuery["wait_for_completion"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitEnrichExecutePolicyEmptyName(t *testing.T) {
	svc := NewEnrichService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ExecutePolicy(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestUnitEnrichExecutePolicyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	_, err := svc.ExecutePolicy(context.Background(), "my-policy", nil)
	assert.Error(t, err)
}

func TestUnitEnrichExecutePolicyNetworkError(t *testing.T) {
	svc := NewEnrichService(deadClient(), testLogger())
	_, err := svc.ExecutePolicy(context.Background(), "my-policy", nil)
	assert.Error(t, err)
}

func TestUnitEnrichStats(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"stats":{}}`)
	}))
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	raw, err := svc.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_enrich/_stats", gotPath)
	assert.Contains(t, string(raw), "stats")
}

func TestUnitEnrichStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewEnrichService(restyClient(srv), testLogger())
	_, err := svc.Stats(context.Background())
	assert.Error(t, err)
}

func TestUnitEnrichStatsNetworkError(t *testing.T) {
	svc := NewEnrichService(deadClient(), testLogger())
	_, err := svc.Stats(context.Background())
	assert.Error(t, err)
}
