package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsAsyncSearchServiceSpecExtensionsDefault(t *testing.T) {
	svc := NewAsyncSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	ext, err := AsAsyncSearchServiceSpecExtensions(svc)
	require.NoError(t, err)
	assert.NotNil(t, ext)
}

func TestAsAsyncSearchServiceSpecExtensionsNil(t *testing.T) {
	_, err := AsAsyncSearchServiceSpecExtensions(nil)
	assert.Error(t, err)
}

func TestAsAsyncSearchServiceSpecExtensionsTypedNil(t *testing.T) {
	var svc AsyncSearchService = (*DefaultAsyncSearchService)(nil)
	_, err := AsAsyncSearchServiceSpecExtensions(svc)
	assert.Error(t, err)
}

type customAsyncSearchService struct{}

func (customAsyncSearchService) Submit(context.Context, *AsyncSearchSubmitRequest) (*AsyncSearchResponse, error) {
	return nil, nil
}
func (customAsyncSearchService) Get(context.Context, string, *AsyncSearchGetParams) (*AsyncSearchResponse, error) {
	return nil, nil
}
func (customAsyncSearchService) Status(context.Context, string) (*AsyncSearchStatusResponse, error) {
	return nil, nil
}
func (customAsyncSearchService) Delete(context.Context, string) (*types.AcknowledgedResponse, error) {
	return nil, nil
}

func TestAsAsyncSearchServiceSpecExtensionsCustom(t *testing.T) {
	_, err := AsAsyncSearchServiceSpecExtensions(customAsyncSearchService{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not implement")
}

func TestStatusWithParamsTransport(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"async-id","is_running":true,"is_partial":false}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	ext, err := AsAsyncSearchServiceSpecExtensions(svc)
	require.NoError(t, err)
	resp, err := ext.StatusWithParams(context.Background(), "async-id", &AsyncSearchStatusParams{KeepAlive: "1m"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_async_search/status/async-id", gotPath)
	assert.Equal(t, "1m", gotQuery["keep_alive"])
	assert.Equal(t, "async-id", resp.Id)
}

func TestStatusWithParamsNilParams(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"x"}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	ext, err := AsAsyncSearchServiceSpecExtensions(svc)
	require.NoError(t, err)
	_, err = ext.StatusWithParams(context.Background(), "x", nil)
	require.NoError(t, err)
	assert.NotContains(t, gotQuery, "keep_alive")
}

func TestStatusWrapperUnchanged(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"x"}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	_, err := svc.Status(context.Background(), "x")
	require.NoError(t, err)
	assert.NotContains(t, gotQuery, "keep_alive")
}

func TestStatusWithParamsEmptyID(t *testing.T) {
	svc := NewAsyncSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	ext, err := AsAsyncSearchServiceSpecExtensions(svc)
	require.NoError(t, err)
	_, err = ext.StatusWithParams(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestAsyncSearchStatusParamsToMap(t *testing.T) {
	assert.Nil(t, (*AsyncSearchStatusParams)(nil).ToMap())
	assert.Empty(t, (&AsyncSearchStatusParams{}).ToMap())
	m := (&AsyncSearchStatusParams{KeepAlive: "2m"}).ToMap()
	assert.Equal(t, "2m", m["keep_alive"])
}
