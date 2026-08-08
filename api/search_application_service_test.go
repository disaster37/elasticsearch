package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSearchApplicationList(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"count":1,"results":[{"name":"app1"}]}`)
	}))
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	resp, err := svc.List(context.Background(), &SearchApplicationListParams{
		From: ptrInt(0),
		Size: ptrInt(10),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_application/search_application", gotPath)
	assert.Equal(t, "0", gotQuery["from"])
	assert.Equal(t, "10", gotQuery["size"])
	assert.Contains(t, string(resp), "app1")
}

func TestUnitSearchApplicationListError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	_, err := svc.List(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSearchApplicationListNetworkError(t *testing.T) {
	svc := NewSearchApplicationService(deadClient(), testLogger())
	_, err := svc.List(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSearchApplicationGet(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"name":"app1"}`)
	}))
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), "app1")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_application/search_application/app1", gotPath)
	assert.Contains(t, string(resp), "app1")
}

func TestUnitSearchApplicationGetEmpty(t *testing.T) {
	svc := NewSearchApplicationService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Get(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSearchApplicationGetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "app1")
	assert.Error(t, err)
}

func TestUnitSearchApplicationGetNetworkError(t *testing.T) {
	svc := NewSearchApplicationService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), "app1")
	assert.Error(t, err)
}

func TestUnitSearchApplicationPut(t *testing.T) {
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
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	resp, err := svc.Put(context.Background(), "app1", map[string]any{"indices": []string{"idx"}}, &SearchApplicationPutParams{
		Create: ptrBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_application/search_application/app1", gotPath)
	assert.Equal(t, "true", gotQuery["create"])
	assert.Contains(t, gotBody, "idx")
	assert.True(t, resp.Acknowledged)
}

func TestUnitSearchApplicationPutEmpty(t *testing.T) {
	svc := NewSearchApplicationService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Put(context.Background(), "", map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitSearchApplicationPutError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	_, err := svc.Put(context.Background(), "app1", map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitSearchApplicationPutNetworkError(t *testing.T) {
	svc := NewSearchApplicationService(deadClient(), testLogger())
	_, err := svc.Put(context.Background(), "app1", map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitSearchApplicationDelete(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background(), "app1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_application/search_application/app1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSearchApplicationDeleteEmpty(t *testing.T) {
	svc := NewSearchApplicationService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Delete(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSearchApplicationDeleteError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	_, err := svc.Delete(context.Background(), "app1")
	assert.Error(t, err)
}

func TestUnitSearchApplicationDeleteNetworkError(t *testing.T) {
	svc := NewSearchApplicationService(deadClient(), testLogger())
	_, err := svc.Delete(context.Background(), "app1")
	assert.Error(t, err)
}

func TestUnitSearchApplicationSearch(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"took":5,"hits":{"total":{"value":0},"hits":[]}}`)
	}))
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	resp, err := svc.Search(context.Background(), "app1", map[string]any{"query": map[string]any{"match_all": map[string]any{}}})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_application/search_application/app1/_search", gotPath)
	assert.Contains(t, gotBody, "match_all")
	assert.Equal(t, int64(5), resp.TookInMillis)
}

func TestUnitSearchApplicationSearchEmpty(t *testing.T) {
	svc := NewSearchApplicationService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Search(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSearchApplicationSearchError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	_, err := svc.Search(context.Background(), "app1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSearchApplicationSearchNetworkError(t *testing.T) {
	svc := NewSearchApplicationService(deadClient(), testLogger())
	_, err := svc.Search(context.Background(), "app1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSearchApplicationGetBehavioralAnalytics(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"count":1,"results":[{"name":"a1"}]}`)
	}))
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	resp, err := svc.GetBehavioralAnalytics(context.Background(), "a1")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_application/analytics/a1", gotPath)
	assert.Contains(t, string(resp), "a1")
}

func TestUnitSearchApplicationGetBehavioralAnalyticsAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"count":0,"results":[]}`)
	}))
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	_, err := svc.GetBehavioralAnalytics(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "/_application/analytics", gotPath)
}

func TestUnitSearchApplicationGetBehavioralAnalyticsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	_, err := svc.GetBehavioralAnalytics(context.Background(), "a1")
	assert.Error(t, err)
}

func TestUnitSearchApplicationGetBehavioralAnalyticsNetworkError(t *testing.T) {
	svc := NewSearchApplicationService(deadClient(), testLogger())
	_, err := svc.GetBehavioralAnalytics(context.Background(), "a1")
	assert.Error(t, err)
}

func TestUnitSearchApplicationPutBehavioralAnalytics(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	resp, err := svc.PutBehavioralAnalytics(context.Background(), "a1", map[string]any{"indices": []string{"idx"}})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_application/analytics/a1", gotPath)
	assert.Contains(t, gotBody, "idx")
	assert.True(t, resp.Acknowledged)
}

func TestUnitSearchApplicationPutBehavioralAnalyticsEmpty(t *testing.T) {
	svc := NewSearchApplicationService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutBehavioralAnalytics(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSearchApplicationPutBehavioralAnalyticsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	_, err := svc.PutBehavioralAnalytics(context.Background(), "a1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSearchApplicationPutBehavioralAnalyticsNetworkError(t *testing.T) {
	svc := NewSearchApplicationService(deadClient(), testLogger())
	_, err := svc.PutBehavioralAnalytics(context.Background(), "a1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSearchApplicationDeleteBehavioralAnalytics(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	resp, err := svc.DeleteBehavioralAnalytics(context.Background(), "a1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_application/analytics/a1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSearchApplicationDeleteBehavioralAnalyticsEmpty(t *testing.T) {
	svc := NewSearchApplicationService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteBehavioralAnalytics(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSearchApplicationDeleteBehavioralAnalyticsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchApplicationService(restyClient(srv), testLogger())
	_, err := svc.DeleteBehavioralAnalytics(context.Background(), "a1")
	assert.Error(t, err)
}

func TestUnitSearchApplicationDeleteBehavioralAnalyticsNetworkError(t *testing.T) {
	svc := NewSearchApplicationService(deadClient(), testLogger())
	_, err := svc.DeleteBehavioralAnalytics(context.Background(), "a1")
	assert.Error(t, err)
}
