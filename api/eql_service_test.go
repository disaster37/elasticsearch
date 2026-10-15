package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitEqlSearch(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"eql-id","is_partial":false,"is_running":false,"timed_out":false,"took":5,"hits":{"total":{"value":1}}}`)
	}))
	defer srv.Close()
	svc := NewEqlService(restyClient(srv), testLogger())
	resp, err := svc.Search(context.Background(), "logs", map[string]any{"query": "process where true"}, &EqlSearchParams{
		KeepAlive:                "1m",
		KeepOnCompletion:         ptrBool(true),
		WaitForCompletionTimeout: "10s",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/logs/_eql/search", gotPath)
	assert.Equal(t, "1m", gotQuery["keep_alive"])
	assert.Equal(t, "true", gotQuery["keep_on_completion"])
	assert.Equal(t, "10s", gotQuery["wait_for_completion_timeout"])
	assert.Contains(t, gotBody, "process where true")
	assert.Equal(t, "eql-id", resp.ID)
	assert.False(t, resp.IsPartial)
	assert.False(t, resp.IsRunning)
	assert.Equal(t, int64(5), resp.Took)
	require.NotNil(t, resp.Hits)
}

func TestUnitEqlSearchEmptyIndex(t *testing.T) {
	svc := NewEqlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Search(context.Background(), "", map[string]any{}, nil)
	assert.Error(t, err)
}

func TestUnitEqlSearchError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewEqlService(restyClient(srv), testLogger())
	_, err := svc.Search(context.Background(), "logs", map[string]any{}, nil)
	assert.Error(t, err)
}

func TestUnitEqlSearchNetworkError(t *testing.T) {
	svc := NewEqlService(deadClient(), testLogger())
	_, err := svc.Search(context.Background(), "logs", map[string]any{}, nil)
	assert.Error(t, err)
}

func TestUnitEqlGet(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"id":"eql-id","is_partial":false,"is_running":false,"timed_out":false,"took":3}`)
	}))
	defer srv.Close()
	svc := NewEqlService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), "eql-id")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_eql/search/eql-id", gotPath)
	assert.Equal(t, "eql-id", resp.ID)
	assert.False(t, resp.IsRunning)
	assert.Equal(t, int64(3), resp.Took)
}

func TestUnitEqlGetEmptyID(t *testing.T) {
	svc := NewEqlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Get(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitEqlGetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewEqlService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "eql-id")
	assert.Error(t, err)
}

func TestUnitEqlGetNetworkError(t *testing.T) {
	svc := NewEqlService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), "eql-id")
	assert.Error(t, err)
}

func TestUnitEqlGetStatus(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"id":"eql-id","is_running":true}`)
	}))
	defer srv.Close()
	svc := NewEqlService(restyClient(srv), testLogger())
	raw, err := svc.GetStatus(context.Background(), "eql-id")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_eql/search/status/eql-id", gotPath)
	assert.Contains(t, string(raw), "is_running")
}

func TestUnitEqlGetStatusEmptyID(t *testing.T) {
	svc := NewEqlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.GetStatus(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitEqlGetStatusError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewEqlService(restyClient(srv), testLogger())
	_, err := svc.GetStatus(context.Background(), "eql-id")
	assert.Error(t, err)
}

func TestUnitEqlGetStatusNetworkError(t *testing.T) {
	svc := NewEqlService(deadClient(), testLogger())
	_, err := svc.GetStatus(context.Background(), "eql-id")
	assert.Error(t, err)
}

func TestUnitEqlDelete(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewEqlService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background(), "eql-id")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_eql/search/eql-id", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitEqlDeleteEmptyID(t *testing.T) {
	svc := NewEqlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Delete(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitEqlDeleteError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewEqlService(restyClient(srv), testLogger())
	_, err := svc.Delete(context.Background(), "eql-id")
	assert.Error(t, err)
}

func TestUnitEqlDeleteNetworkError(t *testing.T) {
	svc := NewEqlService(deadClient(), testLogger())
	_, err := svc.Delete(context.Background(), "eql-id")
	assert.Error(t, err)
}

type customEql struct{ EqlService }

func TestEqlSpecExtensionAdapters(t *testing.T) {
	svc := NewEqlService(nil, testLogger())
	if _, err := AsEqlServiceSpecExtensions(svc); err != nil {
		t.Fatalf("eql default: %v", err)
	}
	if _, err := AsEqlServiceSpecExtensions(nil); err == nil {
		t.Error("eql nil: want error")
	}
	var typedNil EqlService = (*DefaultEqlService)(nil)
	if _, err := AsEqlServiceSpecExtensions(typedNil); err == nil {
		t.Error("eql typed-nil: want error")
	}
	if _, err := AsEqlServiceSpecExtensions(customEql{svc}); err == nil {
		t.Error("eql custom: want error")
	}
}

func TestUnitEqlB7CarrierTransport(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	svc := NewEqlService(restyClient(srv), testLogger())
	ext, err := AsEqlServiceSpecExtensions(svc)
	assert.NoError(t, err)
	ctx := context.Background()
	body := map[string]any{"query": "x"}

	_, err = svc.Search(ctx, "idx", body, &EqlSearchParams{AllowPartialSearchResults: ptrBool(true), AllowPartialSequenceResults: ptrBool(false)})
	assert.NoError(t, err)
	assert.Equal(t, "/idx/_eql/search", got.path)
	assert.Equal(t, "true", got.query["allow_partial_search_results"])
	assert.Equal(t, "false", got.query["allow_partial_sequence_results"])

	*got = capturedSpecRequest{}
	_, err = ext.GetWithParams(ctx, "id1", &EqlGetSpecParams{KeepAlive: "1m", WaitForCompletionTimeout: "2s"})
	assert.NoError(t, err)
	assert.Equal(t, "/_eql/search/id1", got.path)
	assert.Equal(t, "1m", got.query["keep_alive"])
	assert.Equal(t, "2s", got.query["wait_for_completion_timeout"])
}
