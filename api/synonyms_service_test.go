package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSynonymsGetSynonymsSets(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"count":1,"results":[{"synonym_set":"s1","count":2}]}`)
	}))
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	resp, err := svc.GetSynonymsSets(context.Background(), &SynonymsListParams{
		From: ptrInt(0),
		Size: ptrInt(10),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_synonyms", gotPath)
	assert.Equal(t, "0", gotQuery["from"])
	assert.Equal(t, "10", gotQuery["size"])
	assert.Equal(t, 1, resp.Count)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "s1", resp.Results[0].SynonymSet)
	assert.Equal(t, 2, resp.Results[0].Count)
}

func TestUnitSynonymsGetSynonymsSetsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	_, err := svc.GetSynonymsSets(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSynonymsGetSynonymsSetsNetworkError(t *testing.T) {
	svc := NewSynonymsService(deadClient(), testLogger())
	_, err := svc.GetSynonymsSets(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSynonymsGetSynonym(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"count":1,"rules":[{"id":"r1","synonyms":"a,b,c"}]}`)
	}))
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	resp, err := svc.GetSynonym(context.Background(), "s1", nil)
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_synonyms/s1", gotPath)
	assert.Equal(t, 1, resp.Count)
	require.Len(t, resp.Rules, 1)
	assert.Equal(t, "r1", resp.Rules[0].ID)
	assert.Equal(t, "a,b,c", resp.Rules[0].Synonyms)
}

func TestUnitSynonymsGetSynonymEmpty(t *testing.T) {
	svc := NewSynonymsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.GetSynonym(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestUnitSynonymsGetSynonymError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	_, err := svc.GetSynonym(context.Background(), "s1", nil)
	assert.Error(t, err)
}

func TestUnitSynonymsGetSynonymNetworkError(t *testing.T) {
	svc := NewSynonymsService(deadClient(), testLogger())
	_, err := svc.GetSynonym(context.Background(), "s1", nil)
	assert.Error(t, err)
}

func TestUnitSynonymsPutSynonym(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	resp, err := svc.PutSynonym(context.Background(), "s1", map[string]any{"synonyms_set": []string{}})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_synonyms/s1", gotPath)
	assert.Contains(t, gotBody, "synonyms_set")
	assert.True(t, resp.Acknowledged)
}

func TestUnitSynonymsPutSynonymEmpty(t *testing.T) {
	svc := NewSynonymsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutSynonym(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSynonymsPutSynonymError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	_, err := svc.PutSynonym(context.Background(), "s1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSynonymsPutSynonymNetworkError(t *testing.T) {
	svc := NewSynonymsService(deadClient(), testLogger())
	_, err := svc.PutSynonym(context.Background(), "s1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSynonymsDeleteSynonym(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	resp, err := svc.DeleteSynonym(context.Background(), "s1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_synonyms/s1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSynonymsDeleteSynonymEmpty(t *testing.T) {
	svc := NewSynonymsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteSynonym(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSynonymsDeleteSynonymError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	_, err := svc.DeleteSynonym(context.Background(), "s1")
	assert.Error(t, err)
}

func TestUnitSynonymsDeleteSynonymNetworkError(t *testing.T) {
	svc := NewSynonymsService(deadClient(), testLogger())
	_, err := svc.DeleteSynonym(context.Background(), "s1")
	assert.Error(t, err)
}

func TestUnitSynonymsGetSynonymRule(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"id":"r1","synonyms":"a,b,c"}`)
	}))
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	resp, err := svc.GetSynonymRule(context.Background(), "s1", "r1")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_synonyms/s1/r1", gotPath)
	assert.Equal(t, "r1", resp.ID)
	assert.Equal(t, "a,b,c", resp.Synonyms)
}

func TestUnitSynonymsGetSynonymRuleEmptyId(t *testing.T) {
	svc := NewSynonymsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.GetSynonymRule(context.Background(), "", "r1")
	assert.Error(t, err)
}

func TestUnitSynonymsGetSynonymRuleEmptyRuleId(t *testing.T) {
	svc := NewSynonymsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.GetSynonymRule(context.Background(), "s1", "")
	assert.Error(t, err)
}

func TestUnitSynonymsGetSynonymRuleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	_, err := svc.GetSynonymRule(context.Background(), "s1", "r1")
	assert.Error(t, err)
}

func TestUnitSynonymsGetSynonymRuleNetworkError(t *testing.T) {
	svc := NewSynonymsService(deadClient(), testLogger())
	_, err := svc.GetSynonymRule(context.Background(), "s1", "r1")
	assert.Error(t, err)
}

func TestUnitSynonymsPutSynonymRule(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	resp, err := svc.PutSynonymRule(context.Background(), "s1", "r1", map[string]any{"synonyms": "a,b"})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_synonyms/s1/r1", gotPath)
	assert.Contains(t, gotBody, "synonyms")
	assert.True(t, resp.Acknowledged)
}

func TestUnitSynonymsPutSynonymRuleEmptyId(t *testing.T) {
	svc := NewSynonymsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutSynonymRule(context.Background(), "", "r1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSynonymsPutSynonymRuleEmptyRuleId(t *testing.T) {
	svc := NewSynonymsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutSynonymRule(context.Background(), "s1", "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSynonymsPutSynonymRuleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	_, err := svc.PutSynonymRule(context.Background(), "s1", "r1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSynonymsPutSynonymRuleNetworkError(t *testing.T) {
	svc := NewSynonymsService(deadClient(), testLogger())
	_, err := svc.PutSynonymRule(context.Background(), "s1", "r1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSynonymsDeleteSynonymRule(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	resp, err := svc.DeleteSynonymRule(context.Background(), "s1", "r1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_synonyms/s1/r1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSynonymsDeleteSynonymRuleEmptyId(t *testing.T) {
	svc := NewSynonymsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteSynonymRule(context.Background(), "", "r1")
	assert.Error(t, err)
}

func TestUnitSynonymsDeleteSynonymRuleEmptyRuleId(t *testing.T) {
	svc := NewSynonymsService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteSynonymRule(context.Background(), "s1", "")
	assert.Error(t, err)
}

func TestUnitSynonymsDeleteSynonymRuleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	_, err := svc.DeleteSynonymRule(context.Background(), "s1", "r1")
	assert.Error(t, err)
}

func TestUnitSynonymsDeleteSynonymRuleNetworkError(t *testing.T) {
	svc := NewSynonymsService(deadClient(), testLogger())
	_, err := svc.DeleteSynonymRule(context.Background(), "s1", "r1")
	assert.Error(t, err)
}

func TestUnitSynonymsB10Rules(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	svc := NewSynonymsService(restyClient(srv), testLogger())
	ctx := context.Background()

	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
	}{
		{"get", func() error { _, err := svc.GetSynonymRule(ctx, "set1", "rule1"); return err }, "GET", "/_synonyms/set1/rule1"},
		{"put", func() error {
			_, err := svc.PutSynonymRule(ctx, "set1", "rule1", map[string]any{"synonyms": "a,b"})
			return err
		}, "PUT", "/_synonyms/set1/rule1"},
		{"delete", func() error { _, err := svc.DeleteSynonymRule(ctx, "set2", "rule2"); return err }, "DELETE", "/_synonyms/set2/rule2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			*got = capturedSpecRequest{}
			assert.NoError(t, tc.call())
			assert.Equal(t, tc.method, got.method)
			assert.Equal(t, tc.path, got.path)
		})
	}
}
