package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/", r.URL.Path)
		jsonResponse(w, 200, `{"name":"node1","cluster_name":"cluster","cluster_uuid":"uuid","version":{"number":"9.0.0","build_flavor":"default"},"tagline":"You Know, for Search"}`)
	}))
	defer srv.Close()
	svc := NewInfoService(restyClient(srv), testLogger())

	resp, err := svc.Info(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "node1", resp.Name)
	assert.Equal(t, "9.0.0", resp.Version.Number)
}

func TestUnitInfoError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewInfoService(restyClient(srv), testLogger())
	_, err := svc.Info(context.Background())
	assert.Error(t, err)
}

func TestUnitInfoNetworkError(t *testing.T) {
	svc := NewInfoService(deadClient(), testLogger())
	_, err := svc.Info(context.Background())
	assert.Error(t, err)
}

func TestUnitPing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "HEAD", r.Method)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	svc := NewInfoService(restyClient(srv), testLogger())
	ok, err := svc.Ping(context.Background())
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUnitPingNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	svc := NewInfoService(restyClient(srv), testLogger())
	ok, err := svc.Ping(context.Background())
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestUnitCapabilities(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_capabilities", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"capabilities":{"indices":{"data_stream":{"create":true}}}}`)
	}))
	defer srv.Close()
	svc := NewInfoService(restyClient(srv), testLogger())
	resp, err := svc.Capabilities(context.Background(), &CapabilitiesRequest{
		Method:     "PUT",
		Path:       "/idx",
		Parameters: []string{"p1", "p2"},
		Local:      ptrBool(true),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp.Capabilities)
	assert.Equal(t, "PUT", gotQuery["method"])
	assert.Equal(t, "/idx", gotQuery["path"])
	assert.Equal(t, "p1,p2", gotQuery["parameters"])
	assert.Equal(t, "true", gotQuery["local"])
}

func TestUnitCapabilitiesNilRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_capabilities", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)
		jsonResponse(w, 200, `{"capabilities":{}}`)
	}))
	defer srv.Close()
	svc := NewInfoService(restyClient(srv), testLogger())
	resp, err := svc.Capabilities(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestUnitCapabilitiesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewInfoService(restyClient(srv), testLogger())
	_, err := svc.Capabilities(context.Background(), &CapabilitiesRequest{Method: "GET"})
	assert.Error(t, err)
}

func TestUnitCapabilitiesNetworkError(t *testing.T) {
	svc := NewInfoService(deadClient(), testLogger())
	_, err := svc.Capabilities(context.Background(), &CapabilitiesRequest{Method: "GET"})
	assert.Error(t, err)
}
