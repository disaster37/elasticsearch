package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitScriptGet(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"_id":"my-script","found":true,"script":{"lang":"painless","source":"Math.log(_value)","options":{"foo":"bar"}}}`)
	}))
	defer srv.Close()
	svc := NewScriptService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), "my-script", &ScriptGetParams{Timeout: "30s"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_scripts/my-script", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Equal(t, "my-script", resp.Id)
	assert.True(t, resp.Found)
	require.NotNil(t, resp.Script)
	assert.Equal(t, "painless", resp.Script.Lang)
	assert.Equal(t, "Math.log(_value)", resp.Script.Source)
	assert.Equal(t, "bar", resp.Script.Options["foo"])
}

func TestUnitScriptGetNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error":"not found","status":404}`))
	}))
	defer srv.Close()
	svc := NewScriptService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), "missing", &ScriptGetParams{})
	require.NoError(t, err)
	assert.False(t, resp.Found)
}

func TestUnitScriptGetEmptyID(t *testing.T) {
	svc := NewScriptService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Get(context.Background(), "", &ScriptGetParams{})
	assert.Error(t, err)
}

func TestUnitScriptGetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewScriptService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "id", &ScriptGetParams{})
	assert.Error(t, err)
}

func TestUnitScriptGetNetworkError(t *testing.T) {
	svc := NewScriptService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), "id", &ScriptGetParams{})
	assert.Error(t, err)
}

func TestUnitScriptPut(t *testing.T) {
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
	svc := NewScriptService(restyClient(srv), testLogger())
	resp, err := svc.Put(context.Background(), &ScriptPutRequest{
		Id:   "my-script",
		Body: map[string]any{"script": map[string]any{"lang": "painless", "source": "_value * 2"}},
		Params: &ScriptPutParams{
			Context: "score",
			Timeout: "10s",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_scripts/my-script", gotPath)
	assert.Equal(t, "score", gotQuery["context"])
	assert.Equal(t, "10s", gotQuery["timeout"])
	assert.Contains(t, gotBody, "painless")
	assert.True(t, resp.Acknowledged)
}

func TestUnitScriptPutValidation(t *testing.T) {
	svc := NewScriptService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	// missing id
	_, err := svc.Put(context.Background(), &ScriptPutRequest{Body: map[string]any{"x": 1}})
	assert.Error(t, err)
	// missing body
	_, err = svc.Put(context.Background(), &ScriptPutRequest{Id: "id"})
	assert.Error(t, err)
}

func TestUnitScriptPutError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewScriptService(restyClient(srv), testLogger())
	_, err := svc.Put(context.Background(), &ScriptPutRequest{Id: "id", Body: map[string]any{"x": 1}})
	assert.Error(t, err)
}

func TestUnitScriptPutNetworkError(t *testing.T) {
	svc := NewScriptService(deadClient(), testLogger())
	_, err := svc.Put(context.Background(), &ScriptPutRequest{Id: "id", Body: map[string]any{"x": 1}})
	assert.Error(t, err)
}

func TestUnitScriptDelete(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewScriptService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background(), "my-script", &ScriptDeleteParams{Timeout: "5s"})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_scripts/my-script", gotPath)
	assert.Equal(t, "5s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitScriptDeleteEmptyID(t *testing.T) {
	svc := NewScriptService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Delete(context.Background(), "", &ScriptDeleteParams{})
	assert.Error(t, err)
}

func TestUnitScriptDeleteError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewScriptService(restyClient(srv), testLogger())
	_, err := svc.Delete(context.Background(), "id", &ScriptDeleteParams{})
	assert.Error(t, err)
}

func TestUnitScriptDeleteNetworkError(t *testing.T) {
	svc := NewScriptService(deadClient(), testLogger())
	_, err := svc.Delete(context.Background(), "id", &ScriptDeleteParams{})
	assert.Error(t, err)
}

func TestUnitScriptGetContext(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"contexts":[{"context":"score","methods":[{"name":"_score","params":["_value"]}]}]}`)
	}))
	defer srv.Close()
	svc := NewScriptService(restyClient(srv), testLogger())
	resp, err := svc.GetContext(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_script_context", gotPath)
	require.Len(t, resp.Contexts, 1)
	assert.Equal(t, "score", resp.Contexts[0].Context)
	require.Len(t, resp.Contexts[0].Methods, 1)
	assert.Equal(t, "_score", resp.Contexts[0].Methods[0].Name)
	assert.Equal(t, "_value", resp.Contexts[0].Methods[0].Params[0])
}

func TestUnitScriptGetContextError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewScriptService(restyClient(srv), testLogger())
	_, err := svc.GetContext(context.Background())
	assert.Error(t, err)
}

func TestUnitScriptGetContextNetworkError(t *testing.T) {
	svc := NewScriptService(deadClient(), testLogger())
	_, err := svc.GetContext(context.Background())
	assert.Error(t, err)
}

func TestUnitScriptGetLanguages(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"types_allowed":["painless","expression"],"language_contexts":[{"language":"painless","contexts":["score"]}]}`)
	}))
	defer srv.Close()
	svc := NewScriptService(restyClient(srv), testLogger())
	resp, err := svc.GetLanguages(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_script_language", gotPath)
	assert.Equal(t, []string{"painless", "expression"}, resp.TypesAllowed)
	require.Len(t, resp.LanguageContexts, 1)
	assert.Equal(t, "painless", resp.LanguageContexts[0].Language)
	assert.Equal(t, []string{"score"}, resp.LanguageContexts[0].Contexts)
}

func TestUnitScriptGetLanguagesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewScriptService(restyClient(srv), testLogger())
	_, err := svc.GetLanguages(context.Background())
	assert.Error(t, err)
}

func TestUnitScriptGetLanguagesNetworkError(t *testing.T) {
	svc := NewScriptService(deadClient(), testLogger())
	_, err := svc.GetLanguages(context.Background())
	assert.Error(t, err)
}
