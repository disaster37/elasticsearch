package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitAutoscalingGetCapacity(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"policies":[]}`)
	}))
	defer srv.Close()
	svc := NewAutoscalingService(restyClient(srv), testLogger())
	raw, err := svc.GetCapacity(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_autoscaling/capacity", gotPath)
	assert.Contains(t, string(raw), "policies")
}

func TestUnitAutoscalingGetCapacityError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewAutoscalingService(restyClient(srv), testLogger())
	_, err := svc.GetCapacity(context.Background())
	assert.Error(t, err)
}

func TestUnitAutoscalingGetCapacityNetworkError(t *testing.T) {
	svc := NewAutoscalingService(deadClient(), testLogger())
	_, err := svc.GetCapacity(context.Background())
	assert.Error(t, err)
}

func TestUnitAutoscalingGetPolicy(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"policies":{"my-policy":{}}}`)
	}))
	defer srv.Close()
	svc := NewAutoscalingService(restyClient(srv), testLogger())
	raw, err := svc.GetPolicy(context.Background(), "my-policy")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_autoscaling/policy/my-policy", gotPath)
	assert.Contains(t, string(raw), "my-policy")
}

func TestUnitAutoscalingGetPolicyAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"policies":{}}`)
	}))
	defer srv.Close()
	svc := NewAutoscalingService(restyClient(srv), testLogger())
	_, err := svc.GetPolicy(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "/_autoscaling/policy", gotPath)
}

func TestUnitAutoscalingGetPolicyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewAutoscalingService(restyClient(srv), testLogger())
	_, err := svc.GetPolicy(context.Background(), "my-policy")
	assert.Error(t, err)
}

func TestUnitAutoscalingGetPolicyNetworkError(t *testing.T) {
	svc := NewAutoscalingService(deadClient(), testLogger())
	_, err := svc.GetPolicy(context.Background(), "my-policy")
	assert.Error(t, err)
}

func TestUnitAutoscalingPutPolicy(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewAutoscalingService(restyClient(srv), testLogger())
	resp, err := svc.PutPolicy(context.Background(), "my-policy", map[string]any{"roles": []string{"data"}})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_autoscaling/policy/my-policy", gotPath)
	assert.Contains(t, gotBody, "roles")
	assert.True(t, resp.Acknowledged)
}

func TestUnitAutoscalingPutPolicyEmptyName(t *testing.T) {
	svc := NewAutoscalingService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutPolicy(context.Background(), "", map[string]any{})
	assert.Error(t, err)
}

func TestUnitAutoscalingPutPolicyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewAutoscalingService(restyClient(srv), testLogger())
	_, err := svc.PutPolicy(context.Background(), "my-policy", map[string]any{})
	assert.Error(t, err)
}

func TestUnitAutoscalingPutPolicyNetworkError(t *testing.T) {
	svc := NewAutoscalingService(deadClient(), testLogger())
	_, err := svc.PutPolicy(context.Background(), "my-policy", map[string]any{})
	assert.Error(t, err)
}

func TestUnitAutoscalingDeletePolicy(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewAutoscalingService(restyClient(srv), testLogger())
	resp, err := svc.DeletePolicy(context.Background(), "my-policy")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_autoscaling/policy/my-policy", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitAutoscalingDeletePolicyEmptyName(t *testing.T) {
	svc := NewAutoscalingService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeletePolicy(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitAutoscalingDeletePolicyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewAutoscalingService(restyClient(srv), testLogger())
	_, err := svc.DeletePolicy(context.Background(), "my-policy")
	assert.Error(t, err)
}

func TestUnitAutoscalingDeletePolicyNetworkError(t *testing.T) {
	svc := NewAutoscalingService(deadClient(), testLogger())
	_, err := svc.DeletePolicy(context.Background(), "my-policy")
	assert.Error(t, err)
}
