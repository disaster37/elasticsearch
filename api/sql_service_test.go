package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSqlQuery(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"columns":[{"name":"c1","type":"text"}],"rows":[["v1"]],"cursor":"cur1","is_running":false}`)
	}))
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	resp, err := svc.Query(context.Background(), map[string]any{"query": "SELECT * FROM idx"}, &SqlQueryParams{Format: "json"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_sql", gotPath)
	assert.Equal(t, "json", gotQuery["format"])
	assert.Contains(t, gotBody, "SELECT")
	require.Len(t, resp.Columns, 1)
	assert.Equal(t, "c1", resp.Columns[0].Name)
	assert.Equal(t, "text", resp.Columns[0].Type)
	require.Len(t, resp.Rows, 1)
	assert.Equal(t, "v1", resp.Rows[0][0])
	assert.Equal(t, "cur1", resp.Cursor)
	assert.False(t, resp.IsRunning)
}

func TestUnitSqlQueryError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	_, err := svc.Query(context.Background(), map[string]any{"query": "SELECT 1"}, &SqlQueryParams{})
	assert.Error(t, err)
}

func TestUnitSqlQueryNetworkError(t *testing.T) {
	svc := NewSqlService(deadClient(), testLogger())
	_, err := svc.Query(context.Background(), map[string]any{"query": "SELECT 1"}, &SqlQueryParams{})
	assert.Error(t, err)
}

func TestUnitSqlTranslate(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"size":100,"query":{"match":{"field":"value"}},"sort":[{"field":{"order":"asc"}}]}`)
	}))
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	resp, err := svc.Translate(context.Background(), map[string]any{"query": "SELECT * FROM idx"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_sql/translate", gotPath)
	assert.Contains(t, gotBody, "SELECT")
	assert.Equal(t, 100, resp.Size)
	require.Contains(t, resp.Query, "match")
	require.Len(t, resp.Sort, 1)
	assert.Contains(t, resp.Sort[0], "field")
}

func TestUnitSqlTranslateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	_, err := svc.Translate(context.Background(), map[string]any{"query": "SELECT 1"})
	assert.Error(t, err)
}

func TestUnitSqlTranslateNetworkError(t *testing.T) {
	svc := NewSqlService(deadClient(), testLogger())
	_, err := svc.Translate(context.Background(), map[string]any{"query": "SELECT 1"})
	assert.Error(t, err)
}

func TestUnitSqlClearCursor(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"succeeded":true}`)
	}))
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	resp, err := svc.ClearCursor(context.Background(), map[string]any{"cursor": "cur1"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_sql/close", gotPath)
	assert.Contains(t, gotBody, "cur1")
	assert.True(t, resp.Succeeded)
}

func TestUnitSqlClearCursorError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	_, err := svc.ClearCursor(context.Background(), map[string]any{"cursor": "cur1"})
	assert.Error(t, err)
}

func TestUnitSqlClearCursorNetworkError(t *testing.T) {
	svc := NewSqlService(deadClient(), testLogger())
	_, err := svc.ClearCursor(context.Background(), map[string]any{"cursor": "cur1"})
	assert.Error(t, err)
}

func TestUnitSqlGetAsync(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"columns":[{"name":"c1","type":"text"}],"rows":[["v1"]],"cursor":"cur1","is_running":false}`)
	}))
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	resp, err := svc.GetAsync(context.Background(), "async1", &SqlGetAsyncParams{
		Delimiter:                ",",
		Format:                   "csv",
		KeepAlive:                "1d",
		WaitForCompletionTimeout: "30s",
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_sql/async/async1", gotPath)
	assert.Equal(t, ",", gotQuery["delimiter"])
	assert.Equal(t, "csv", gotQuery["format"])
	assert.Equal(t, "1d", gotQuery["keep_alive"])
	assert.Equal(t, "30s", gotQuery["wait_for_completion_timeout"])
	require.Len(t, resp.Columns, 1)
	assert.Equal(t, "c1", resp.Columns[0].Name)
	assert.Equal(t, "cur1", resp.Cursor)
}

func TestUnitSqlGetAsyncEmpty(t *testing.T) {
	svc := NewSqlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.GetAsync(context.Background(), "", &SqlGetAsyncParams{})
	assert.Error(t, err)
}

func TestUnitSqlGetAsyncError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	_, err := svc.GetAsync(context.Background(), "async1", &SqlGetAsyncParams{})
	assert.Error(t, err)
}

func TestUnitSqlGetAsyncNetworkError(t *testing.T) {
	svc := NewSqlService(deadClient(), testLogger())
	_, err := svc.GetAsync(context.Background(), "async1", &SqlGetAsyncParams{})
	assert.Error(t, err)
}

func TestUnitSqlGetAsyncStatus(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"id":"async1","is_running":true,"is_partial":false,"start_in_millis":100,"expiration_in_millis":200,"completion_status":200}`)
	}))
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	resp, err := svc.GetAsyncStatus(context.Background(), "async1")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_sql/async/status/async1", gotPath)
	assert.Equal(t, "async1", resp.ID)
	assert.True(t, resp.IsRunning)
	assert.False(t, resp.IsPartial)
	assert.Equal(t, int64(100), resp.StartInMillis)
	assert.Equal(t, int64(200), resp.ExpirationInMillis)
	assert.Equal(t, 200, resp.CompletionStatus)
}

func TestUnitSqlGetAsyncStatusEmpty(t *testing.T) {
	svc := NewSqlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.GetAsyncStatus(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSqlGetAsyncStatusError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	_, err := svc.GetAsyncStatus(context.Background(), "async1")
	assert.Error(t, err)
}

func TestUnitSqlGetAsyncStatusNetworkError(t *testing.T) {
	svc := NewSqlService(deadClient(), testLogger())
	_, err := svc.GetAsyncStatus(context.Background(), "async1")
	assert.Error(t, err)
}

func TestUnitSqlDeleteAsync(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	resp, err := svc.DeleteAsync(context.Background(), "async1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_sql/async/async1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSqlDeleteAsyncEmpty(t *testing.T) {
	svc := NewSqlService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteAsync(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSqlDeleteAsyncError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSqlService(restyClient(srv), testLogger())
	_, err := svc.DeleteAsync(context.Background(), "async1")
	assert.Error(t, err)
}

func TestUnitSqlDeleteAsyncNetworkError(t *testing.T) {
	svc := NewSqlService(deadClient(), testLogger())
	_, err := svc.DeleteAsync(context.Background(), "async1")
	assert.Error(t, err)
}
