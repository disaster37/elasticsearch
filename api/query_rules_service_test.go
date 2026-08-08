package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitQueryRulesListRulesets(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"count":1,"rulesets":[{"ruleset_id":"r1"}]}`)
	}))
	defer srv.Close()
	svc := NewQueryRulesService(restyClient(srv), testLogger())
	resp, err := svc.ListRulesets(context.Background(), &QueryRulesListParams{
		From: ptrInt(0),
		Size: ptrInt(10),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_query_ruleset", gotPath)
	assert.Equal(t, "0", gotQuery["from"])
	assert.Equal(t, "10", gotQuery["size"])
	assert.Contains(t, string(resp), "rulesets")
}

func TestUnitQueryRulesListRulesetsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewQueryRulesService(restyClient(srv), testLogger())
	_, err := svc.ListRulesets(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitQueryRulesListRulesetsNetworkError(t *testing.T) {
	svc := NewQueryRulesService(deadClient(), testLogger())
	_, err := svc.ListRulesets(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitQueryRulesGetRuleset(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"ruleset_id":"r1","rules":[]}`)
	}))
	defer srv.Close()
	svc := NewQueryRulesService(restyClient(srv), testLogger())
	resp, err := svc.GetRuleset(context.Background(), "r1")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_query_ruleset/r1", gotPath)
	assert.Contains(t, string(resp), "ruleset_id")
}

func TestUnitQueryRulesGetRulesetEmpty(t *testing.T) {
	svc := NewQueryRulesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.GetRuleset(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitQueryRulesGetRulesetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewQueryRulesService(restyClient(srv), testLogger())
	_, err := svc.GetRuleset(context.Background(), "r1")
	assert.Error(t, err)
}

func TestUnitQueryRulesGetRulesetNetworkError(t *testing.T) {
	svc := NewQueryRulesService(deadClient(), testLogger())
	_, err := svc.GetRuleset(context.Background(), "r1")
	assert.Error(t, err)
}

func TestUnitQueryRulesPutRuleset(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewQueryRulesService(restyClient(srv), testLogger())
	resp, err := svc.PutRuleset(context.Background(), "r1", map[string]any{"rules": []string{}})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_query_ruleset/r1", gotPath)
	assert.Contains(t, gotBody, "rules")
	assert.True(t, resp.Acknowledged)
}

func TestUnitQueryRulesPutRulesetEmpty(t *testing.T) {
	svc := NewQueryRulesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutRuleset(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitQueryRulesPutRulesetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewQueryRulesService(restyClient(srv), testLogger())
	_, err := svc.PutRuleset(context.Background(), "r1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitQueryRulesPutRulesetNetworkError(t *testing.T) {
	svc := NewQueryRulesService(deadClient(), testLogger())
	_, err := svc.PutRuleset(context.Background(), "r1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitQueryRulesDeleteRuleset(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewQueryRulesService(restyClient(srv), testLogger())
	resp, err := svc.DeleteRuleset(context.Background(), "r1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_query_ruleset/r1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitQueryRulesDeleteRulesetEmpty(t *testing.T) {
	svc := NewQueryRulesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteRuleset(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitQueryRulesDeleteRulesetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewQueryRulesService(restyClient(srv), testLogger())
	_, err := svc.DeleteRuleset(context.Background(), "r1")
	assert.Error(t, err)
}

func TestUnitQueryRulesDeleteRulesetNetworkError(t *testing.T) {
	svc := NewQueryRulesService(deadClient(), testLogger())
	_, err := svc.DeleteRuleset(context.Background(), "r1")
	assert.Error(t, err)
}

func TestUnitQueryRulesTest(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"match":[{"id":"r1"}]}`)
	}))
	defer srv.Close()
	svc := NewQueryRulesService(restyClient(srv), testLogger())
	resp, err := svc.Test(context.Background(), "r1", map[string]any{"q": "test"}, &QueryRulesTestParams{
		QueryText:   "my query",
		QueryParams: []string{"key=val"},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_query_ruleset/r1/_test", gotPath)
	assert.Equal(t, "my query", gotQuery["query_text"])
	assert.Equal(t, "key=val", gotQuery["query_params"])
	assert.Contains(t, gotBody, "test")
	assert.Contains(t, string(resp), "match")
}

func TestUnitQueryRulesTestEmpty(t *testing.T) {
	svc := NewQueryRulesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Test(context.Background(), "", map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitQueryRulesTestError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewQueryRulesService(restyClient(srv), testLogger())
	_, err := svc.Test(context.Background(), "r1", map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}

func TestUnitQueryRulesTestNetworkError(t *testing.T) {
	svc := NewQueryRulesService(deadClient(), testLogger())
	_, err := svc.Test(context.Background(), "r1", map[string]any{"x": 1}, nil)
	assert.Error(t, err)
}
