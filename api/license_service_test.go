package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitLicenseGet(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"license":{"uid":"uid","type":"basic","status":"active","issue_date":"2020-01-01","max_nodes":1000,"issued_to":"me","issuer":"es"}}`)
	}))
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), &LicenseGetParams{
		Local: ptrBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_license", gotPath)
	assert.Equal(t, "true", gotQuery["local"])
	assert.Equal(t, "uid", resp.License.UID)
	assert.Equal(t, "basic", resp.License.Type)
	assert.Equal(t, "active", resp.License.Status)
	assert.Equal(t, 1000, resp.License.MaxNodes)
	assert.Equal(t, "me", resp.License.IssuedTo)
}

func TestUnitLicenseGetNilParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, `{"license":{"uid":"uid","type":"basic","status":"active"}}`)
	}))
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "uid", resp.License.UID)
}

func TestUnitLicenseGetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitLicenseGetNetworkError(t *testing.T) {
	svc := NewLicenseService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitLicenseDelete(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_license", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitLicenseDeleteError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	_, err := svc.Delete(context.Background())
	assert.Error(t, err)
}

func TestUnitLicenseDeleteNetworkError(t *testing.T) {
	svc := NewLicenseService(deadClient(), testLogger())
	_, err := svc.Delete(context.Background())
	assert.Error(t, err)
}

func TestUnitLicensePost(t *testing.T) {
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
	svc := NewLicenseService(restyClient(srv), testLogger())
	resp, err := svc.Post(context.Background(), map[string]any{"license": map[string]any{"type": "gold"}}, &LicensePostParams{
		Acknowledge: ptrBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_license", gotPath)
	assert.Equal(t, "true", gotQuery["acknowledge"])
	assert.Contains(t, gotBody, "gold")
	assert.True(t, resp.Acknowledged)
}

func TestUnitLicensePostError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	_, err := svc.Post(context.Background(), map[string]any{}, nil)
	assert.Error(t, err)
}

func TestUnitLicensePostNetworkError(t *testing.T) {
	svc := NewLicenseService(deadClient(), testLogger())
	_, err := svc.Post(context.Background(), map[string]any{}, nil)
	assert.Error(t, err)
}

func TestUnitLicenseGetBasicStatus(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"eligible_to_start_basic":true}`)
	}))
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	raw, err := svc.GetBasicStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_license/basic_status", gotPath)
	assert.Contains(t, string(raw), "eligible_to_start_basic")
}

func TestUnitLicenseGetBasicStatusError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	_, err := svc.GetBasicStatus(context.Background())
	assert.Error(t, err)
}

func TestUnitLicenseGetBasicStatusNetworkError(t *testing.T) {
	svc := NewLicenseService(deadClient(), testLogger())
	_, err := svc.GetBasicStatus(context.Background())
	assert.Error(t, err)
}

func TestUnitLicensePostStartBasic(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"basic_started":true}`)
	}))
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	raw, err := svc.PostStartBasic(context.Background(), &LicensePostParams{
		Acknowledge: ptrBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_license/start_basic", gotPath)
	assert.Equal(t, "true", gotQuery["acknowledge"])
	assert.Contains(t, string(raw), "basic_started")
}

func TestUnitLicensePostStartBasicError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	_, err := svc.PostStartBasic(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitLicensePostStartBasicNetworkError(t *testing.T) {
	svc := NewLicenseService(deadClient(), testLogger())
	_, err := svc.PostStartBasic(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitLicenseGetTrialStatus(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"eligible_to_start_trial":true}`)
	}))
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	raw, err := svc.GetTrialStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_license/trial_status", gotPath)
	assert.Contains(t, string(raw), "eligible_to_start_trial")
}

func TestUnitLicenseGetTrialStatusError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	_, err := svc.GetTrialStatus(context.Background())
	assert.Error(t, err)
}

func TestUnitLicenseGetTrialStatusNetworkError(t *testing.T) {
	svc := NewLicenseService(deadClient(), testLogger())
	_, err := svc.GetTrialStatus(context.Background())
	assert.Error(t, err)
}

func TestUnitLicensePostStartTrial(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"trial_started":true}`)
	}))
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	raw, err := svc.PostStartTrial(context.Background(), &LicenseStartTrialParams{
		Acknowledge: ptrBool(true),
		Type:        "trial",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_license/start_trial", gotPath)
	assert.Equal(t, "true", gotQuery["acknowledge"])
	assert.Equal(t, "trial", gotQuery["type"])
	assert.Contains(t, string(raw), "trial_started")
}

func TestUnitLicensePostStartTrialError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	_, err := svc.PostStartTrial(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitLicensePostStartTrialNetworkError(t *testing.T) {
	svc := NewLicenseService(deadClient(), testLogger())
	_, err := svc.PostStartTrial(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitLicenseB3Timeouts(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	_, err := svc.Post(context.Background(), map[string]any{"license": "x"}, &LicensePostParams{MasterTimeout: "m1", Timeout: "t1"})
	require.NoError(t, err)
	assert.Equal(t, "m1", gotQuery["master_timeout"])
	assert.Equal(t, "t1", gotQuery["timeout"])

	_, err = svc.PostStartBasic(context.Background(), &LicensePostParams{MasterTimeout: "m2"})
	require.NoError(t, err)
	assert.Equal(t, "m2", gotQuery["master_timeout"])
}

type customLicense struct{ LicenseService }

func TestLicenseSpecExtensionAdapters(t *testing.T) {
	svc := NewLicenseService(nil, testLogger())
	if _, err := AsLicenseServiceSpecExtensions(svc); err != nil {
		t.Fatalf("license default: %v", err)
	}
	if _, err := AsLicenseServiceSpecExtensions(nil); err == nil {
		t.Error("license nil: want error")
	}
	var typedNil LicenseService = (*DefaultLicenseService)(nil)
	if _, err := AsLicenseServiceSpecExtensions(typedNil); err == nil {
		t.Error("license typed-nil: want error")
	}
	if _, err := AsLicenseServiceSpecExtensions(customLicense{svc}); err == nil {
		t.Error("license custom: want error")
	}
}

func TestUnitLicenseB8Transport(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	svc := NewLicenseService(restyClient(srv), testLogger())
	ext, err := AsLicenseServiceSpecExtensions(svc)
	assert.NoError(t, err)
	ctx := context.Background()

	_, err = ext.DeleteWithParams(ctx, &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1", Timeout: "t1"})
	assert.NoError(t, err)
	assert.Equal(t, "DELETE", got.method)
	assert.Equal(t, "/_license", got.path)
	assert.Equal(t, "m1", got.query["master_timeout"])
	assert.Equal(t, "t1", got.query["timeout"])

	*got = capturedSpecRequest{}
	_, err = svc.Get(ctx, &LicenseGetParams{AcceptEnterprise: ptrBool(true)})
	assert.NoError(t, err)
	assert.Equal(t, "true", got.query["accept_enterprise"])

	*got = capturedSpecRequest{}
	_, err = svc.PostStartTrial(ctx, &LicenseStartTrialParams{MasterTimeout: "m1"})
	assert.NoError(t, err)
	assert.Equal(t, "/_license/start_trial", got.path)
	assert.Equal(t, "m1", got.query["master_timeout"])
}
