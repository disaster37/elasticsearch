package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_search", r.URL.Path)
		jsonResponse(w, 200, `{"took":1,"hits":{"total":{"value":1,"relation":"eq"},"hits":[{"_index":"idx","_id":"1","_source":{"f":"v"}}]}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.Search(context.Background(), &SearchRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"query": map[string]any{"match_all": map[string]any{}}},
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.Hits.TotalHits.Value)
}

func TestUnitSearchClusterWide(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_search", r.URL.Path)
		jsonResponse(w, 200, `{"took":0,"hits":{"total":{"value":0,"relation":"eq"},"hits":[]}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Search(context.Background(), &SearchRequest{Body: map[string]any{}})
	assert.NoError(t, err)
}

func TestUnitCount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_count", r.URL.Path)
		jsonResponse(w, 200, `{"count":42}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	n, err := svc.Count(context.Background(), &CountRequest{Indices: []string{"idx"}})
	assert.NoError(t, err)
	assert.Equal(t, int64(42), n)
}

func TestUnitSearchError(t *testing.T) {
	srv := errServer(400)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Search(context.Background(), &SearchRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitOpenClosePIT(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/idx/_pit":
			assert.Equal(t, "keep_alive=1m", r.URL.RawQuery)
			jsonResponse(w, 200, `{"id":"pit-id"}`)
		case "/_pit":
			assert.Equal(t, "DELETE", r.Method)
			jsonResponse(w, 200, `{"succeeded":true,"num_freed":1}`)
		}
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.OpenPointInTime(context.Background(), &OpenPITRequest{Indices: []string{"idx"}, Params: &OpenPITParams{KeepAlive: "1m"}})
	assert.NoError(t, err)
	assert.Equal(t, "pit-id", resp.Id)
	closeResp, err := svc.ClosePointInTime(context.Background(), []string{"pit-id"})
	assert.NoError(t, err)
	assert.True(t, closeResp.Succeeded)
	assert.Equal(t, 1, closeResp.NumFreed)
}

// -----------------------------------------------------------------------------
// MultiSearch
// -----------------------------------------------------------------------------

func TestUnitSearchMultiSearch(t *testing.T) {
	var gotPath, gotMethod, gotBody, gotCT string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotCT = r.Header.Get("Content-Type")
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"took":1,"responses":[]}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.MultiSearch(context.Background(), &MultiSearchRequest{
		Body:   `{"index":"idx"}` + "\n" + `{"query":{"match_all":{}}}` + "\n",
		Params: &MultiSearchParams{MaxConcurrentSearches: ptrInt(5)},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_msearch", gotPath)
	assert.Contains(t, gotCT, "x-ndjson")
	assert.Contains(t, gotBody, "match_all")
	assert.Empty(t, resp.Responses)
	assert.Equal(t, "5", gotQuery["max_concurrent_searches"])
}

func TestUnitSearchMultiSearchWithIndices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx1,idx2/_msearch", r.URL.Path)
		jsonResponse(w, 200, `{"responses":[]}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.MultiSearch(context.Background(), &MultiSearchRequest{
		Indices: []string{"idx1", "idx2"},
		Body:    "{}\n{}\n",
	})
	assert.NoError(t, err)
}

func TestUnitSearchMultiSearchEnsuresTrailingNewline(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"responses":[]}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.MultiSearch(context.Background(), &MultiSearchRequest{Body: "{}\n{}"})
	require.NoError(t, err)
	assert.True(t, len(gotBody) > 0 && gotBody[len(gotBody)-1] == '\n')
}

func TestUnitSearchMultiSearchError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.MultiSearch(context.Background(), &MultiSearchRequest{Body: "{}\n"})
	assert.Error(t, err)
}

func TestUnitSearchMultiSearchNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.MultiSearch(context.Background(), &MultiSearchRequest{Body: "{}\n"})
	assert.Error(t, err)
}

func TestUnitSearchMultiSearchBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.MultiSearch(context.Background(), &MultiSearchRequest{Body: "{}\n"})
	assert.Error(t, err)
}

func TestUnitSearchMultiSearchValidation(t *testing.T) {
	svc := NewSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.MultiSearch(context.Background(), &MultiSearchRequest{Body: ""})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// SearchMVT
// -----------------------------------------------------------------------------

func TestUnitSearchMVT(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/vnd.mapbox-vector-tile")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("binary-mvt-data"))
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.SearchMVT(context.Background(), &SearchMVTRequest{
		Index:  "idx",
		Field:  "geo",
		Zoom:   ptrInt(10),
		X:      ptrInt(5),
		Y:      ptrInt(3),
		Body:   map[string]any{"size": 1000},
		Params: &SearchMVTParams{GridPrecision: ptrInt(6)},
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/idx/_mvt/geo/10/5/3", gotPath)
	assert.Equal(t, "binary-mvt-data", string(resp))
}

func TestUnitSearchMVTZeroCoords(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.SearchMVT(context.Background(), &SearchMVTRequest{
		Index: "idx", Field: "geo", Zoom: ptrInt(0), X: ptrInt(0), Y: ptrInt(0),
	})
	require.NoError(t, err)
	assert.Equal(t, "/idx/_mvt/geo/0/0/0", gotPath)
}

func TestUnitSearchMVTError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.SearchMVT(context.Background(), &SearchMVTRequest{
		Index: "idx", Field: "geo", Zoom: ptrInt(1), X: ptrInt(1), Y: ptrInt(1),
	})
	assert.Error(t, err)
}

func TestUnitSearchMVTNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.SearchMVT(context.Background(), &SearchMVTRequest{
		Index: "idx", Field: "geo", Zoom: ptrInt(1), X: ptrInt(1), Y: ptrInt(1),
	})
	assert.Error(t, err)
}

func TestUnitSearchMVTValidation(t *testing.T) {
	svc := NewSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	// Missing index and field
	_, err := svc.SearchMVT(context.Background(), &SearchMVTRequest{
		Zoom: ptrInt(1), X: ptrInt(1), Y: ptrInt(1),
	})
	assert.Error(t, err)
	// Missing coordinates
	_, err = svc.SearchMVT(context.Background(), &SearchMVTRequest{
		Index: "idx", Field: "geo",
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// SearchTemplate
// -----------------------------------------------------------------------------

func TestUnitSearchTemplate(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"took":1,"hits":{"total":{"value":0,"relation":"eq"},"hits":[]}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.SearchTemplate(context.Background(), &SearchTemplateRequest{
		Body:   map[string]any{"id": "my-template", "params": map[string]any{"q": "foo"}},
		Params: &SearchTemplateParams{Explain: ptrBool(true)},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_search/template", gotPath)
	assert.Contains(t, gotBody, "my-template")
	assert.NotNil(t, resp)
}

func TestUnitSearchTemplateWithIndices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_search/template", r.URL.Path)
		jsonResponse(w, 200, `{"hits":{"total":{"value":0,"relation":"eq"},"hits":[]}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.SearchTemplate(context.Background(), &SearchTemplateRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"id": "t"},
	})
	assert.NoError(t, err)
}

func TestUnitSearchTemplateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.SearchTemplate(context.Background(), &SearchTemplateRequest{Body: map[string]any{"id": "t"}})
	assert.Error(t, err)
}

func TestUnitSearchTemplateNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.SearchTemplate(context.Background(), &SearchTemplateRequest{Body: map[string]any{"id": "t"}})
	assert.Error(t, err)
}

func TestUnitSearchTemplateBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.SearchTemplate(context.Background(), &SearchTemplateRequest{Body: map[string]any{"id": "t"}})
	assert.Error(t, err)
}

func TestUnitSearchTemplateValidation(t *testing.T) {
	svc := NewSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.SearchTemplate(context.Background(), &SearchTemplateRequest{Body: nil})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// MultiSearchTemplate
// -----------------------------------------------------------------------------

func TestUnitSearchMultiSearchTemplate(t *testing.T) {
	var gotPath, gotMethod, gotBody, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotCT = r.Header.Get("Content-Type")
		jsonResponse(w, 200, `{"took":1,"responses":[]}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.MultiSearchTemplate(context.Background(), &MultiSearchTemplateRequest{
		Body: `{"id":"t"}` + "\n" + `{"params":{}}` + "\n",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_msearch/template", gotPath)
	assert.Contains(t, gotCT, "x-ndjson")
	assert.Contains(t, gotBody, "params")
	assert.Empty(t, resp.Responses)
}

func TestUnitSearchMultiSearchTemplateWithIndices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_msearch/template", r.URL.Path)
		jsonResponse(w, 200, `{"responses":[]}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.MultiSearchTemplate(context.Background(), &MultiSearchTemplateRequest{
		Indices: []string{"idx"},
		Body:    "{}\n{}\n",
	})
	assert.NoError(t, err)
}

func TestUnitSearchMultiSearchTemplateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.MultiSearchTemplate(context.Background(), &MultiSearchTemplateRequest{Body: "{}\n"})
	assert.Error(t, err)
}

func TestUnitSearchMultiSearchTemplateNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.MultiSearchTemplate(context.Background(), &MultiSearchTemplateRequest{Body: "{}\n"})
	assert.Error(t, err)
}

func TestUnitSearchMultiSearchTemplateBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.MultiSearchTemplate(context.Background(), &MultiSearchTemplateRequest{Body: "{}\n"})
	assert.Error(t, err)
}

func TestUnitSearchMultiSearchTemplateValidation(t *testing.T) {
	svc := NewSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.MultiSearchTemplate(context.Background(), &MultiSearchTemplateRequest{Body: ""})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// RenderSearchTemplate
// -----------------------------------------------------------------------------

func TestUnitSearchRenderSearchTemplateWithID(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"template_output":{"query":{"match":{"foo":"bar"}}}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.RenderSearchTemplate(context.Background(), &RenderSearchTemplateRequest{
		Id:   "my-template",
		Body: map[string]any{"params": map[string]any{"foo": "bar"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_render/template/my-template", gotPath)
	assert.Contains(t, gotBody, "params")
	assert.Contains(t, string(resp.TemplateOutput), "match")
}

func TestUnitSearchRenderSearchTemplateWithoutID(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"template_output":{}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.RenderSearchTemplate(context.Background(), &RenderSearchTemplateRequest{
		Body: map[string]any{"source": `{"query":{"match_all":{}}}`},
	})
	require.NoError(t, err)
	assert.Equal(t, "/_render/template", gotPath)
}

func TestUnitSearchRenderSearchTemplateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.RenderSearchTemplate(context.Background(), &RenderSearchTemplateRequest{Id: "t", Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitSearchRenderSearchTemplateNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.RenderSearchTemplate(context.Background(), &RenderSearchTemplateRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitSearchRenderSearchTemplateBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.RenderSearchTemplate(context.Background(), &RenderSearchTemplateRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// RankEval
// -----------------------------------------------------------------------------

func TestUnitSearchRankEval(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"rank_eval_score":0.87}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.RankEval(context.Background(), &RankEvalRequest{
		Body: map[string]any{"requests": []any{}},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_rank_eval", gotPath)
	assert.Contains(t, gotBody, "requests")
	assert.InDelta(t, 0.87, resp.RankEvalScore, 0.0001)
}

func TestUnitSearchRankEvalWithIndices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_rank_eval", r.URL.Path)
		jsonResponse(w, 200, `{"rank_eval_score":0.5}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.RankEval(context.Background(), &RankEvalRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"requests": []any{}},
	})
	assert.NoError(t, err)
}

func TestUnitSearchRankEvalError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.RankEval(context.Background(), &RankEvalRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitSearchRankEvalNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.RankEval(context.Background(), &RankEvalRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitSearchRankEvalBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.RankEval(context.Background(), &RankEvalRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitSearchRankEvalValidation(t *testing.T) {
	svc := NewSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.RankEval(context.Background(), &RankEvalRequest{Body: nil})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Count (additional variants)
// -----------------------------------------------------------------------------

func TestUnitCountClusterWide(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_count", r.URL.Path)
		jsonResponse(w, 200, `{"count":7}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	n, err := svc.Count(context.Background(), &CountRequest{Body: map[string]any{"query": map[string]any{"match_all": map[string]any{}}}})
	require.NoError(t, err)
	assert.Equal(t, int64(7), n)
}

func TestUnitCountError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Count(context.Background(), &CountRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

func TestUnitCountNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.Count(context.Background(), &CountRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

func TestUnitCountBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Count(context.Background(), &CountRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Scroll
// -----------------------------------------------------------------------------

func TestUnitSearchScroll(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"took":1,"hits":{"total":{"value":0,"relation":"eq"},"hits":[]}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.Scroll(context.Background(), &ScrollRequest{
		ScrollId: "scroll-1",
		Scroll:   "1m",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_search/scroll", gotPath)
	assert.Contains(t, gotBody, "scroll-1")
	assert.Contains(t, gotBody, "1m")
	assert.NotNil(t, resp)
}

func TestUnitSearchScrollDefaultKeepAlive(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"hits":{"total":{"value":0,"relation":"eq"},"hits":[]}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Scroll(context.Background(), &ScrollRequest{ScrollId: "scroll-1"})
	require.NoError(t, err)
	assert.Contains(t, gotBody, "5m")
}

func TestUnitSearchScrollFromParams(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"hits":{"total":{"value":0,"relation":"eq"},"hits":[]}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Scroll(context.Background(), &ScrollRequest{
		Params: &ScrollParams{ScrollId: "from-params", Scroll: "2m"},
	})
	require.NoError(t, err)
	assert.Contains(t, gotBody, "from-params")
	assert.Contains(t, gotBody, "2m")
}

func TestUnitSearchScrollError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Scroll(context.Background(), &ScrollRequest{ScrollId: "scroll-1"})
	assert.Error(t, err)
}

func TestUnitSearchScrollNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.Scroll(context.Background(), &ScrollRequest{ScrollId: "scroll-1"})
	assert.Error(t, err)
}

func TestUnitSearchScrollBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Scroll(context.Background(), &ScrollRequest{ScrollId: "scroll-1"})
	assert.Error(t, err)
}

func TestUnitSearchScrollValidation(t *testing.T) {
	svc := NewSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Scroll(context.Background(), &ScrollRequest{ScrollId: ""})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// ClearScroll
// -----------------------------------------------------------------------------

func TestUnitSearchClearScroll(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"succeeded":true,"num_freed":2}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.ClearScroll(context.Background(), []string{"scroll-1", "scroll-2"})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_search/scroll", gotPath)
	assert.Contains(t, gotBody, "scroll-1")
	assert.Contains(t, gotBody, "scroll-2")
	assert.True(t, resp.Succeeded)
	assert.Equal(t, 2, resp.NumFreed)
}

func TestUnitSearchClearScrollError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.ClearScroll(context.Background(), []string{"scroll-1"})
	assert.Error(t, err)
}

func TestUnitSearchClearScrollNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.ClearScroll(context.Background(), []string{"scroll-1"})
	assert.Error(t, err)
}

func TestUnitSearchClearScrollBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.ClearScroll(context.Background(), []string{"scroll-1"})
	assert.Error(t, err)
}

func TestUnitSearchClearScrollValidation(t *testing.T) {
	svc := NewSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ClearScroll(context.Background(), []string{})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Validate
// -----------------------------------------------------------------------------

func TestUnitSearchValidateWithBody(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"valid":true}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.Validate(context.Background(), &ValidateQueryRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"query": map[string]any{"match_all": map[string]any{}}},
		Params:  &ValidateQueryParams{AllShards: ptrBool(true)},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx/_validate/query", gotPath)
	assert.Contains(t, gotBody, "match_all")
	assert.True(t, resp.Valid)
}

func TestUnitSearchValidateWithoutBody(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		jsonResponse(w, 200, `{"valid":true}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Validate(context.Background(), &ValidateQueryRequest{Params: &ValidateQueryParams{Q: "foo"}})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
}

func TestUnitSearchValidateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Validate(context.Background(), &ValidateQueryRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitSearchValidateNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.Validate(context.Background(), &ValidateQueryRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

func TestUnitSearchValidateBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.Validate(context.Background(), &ValidateQueryRequest{Body: map[string]any{}})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// SearchShards
// -----------------------------------------------------------------------------

func TestUnitSearchSearchShards(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"nodes":{},"indices":{},"shards":[]}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.SearchShards(context.Background(), &SearchShardsRequest{
		Indices: []string{"idx"},
		Params:  &SearchShardsParams{Local: ptrBool(true)},
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/idx/_search_shards", gotPath)
	assert.NotNil(t, resp)
}

func TestUnitSearchSearchShardsClusterWide(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_search_shards", r.URL.Path)
		jsonResponse(w, 200, `{"nodes":{},"indices":{},"shards":[]}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.SearchShards(context.Background(), &SearchShardsRequest{})
	assert.NoError(t, err)
}

func TestUnitSearchSearchShardsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.SearchShards(context.Background(), &SearchShardsRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

func TestUnitSearchSearchShardsNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.SearchShards(context.Background(), &SearchShardsRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

func TestUnitSearchSearchShardsBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.SearchShards(context.Background(), &SearchShardsRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// FieldCaps
// -----------------------------------------------------------------------------

func TestUnitSearchFieldCaps(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"indices":["idx"],"fields":{}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.FieldCaps(context.Background(), &FieldCapsRequest{
		Indices: []string{"idx"},
		Params:  &FieldCapsParams{Fields: []string{"foo"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/idx/_field_caps", gotPath)
	assert.Equal(t, []string{"idx"}, resp.Indices)
}

func TestUnitSearchFieldCapsClusterWide(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_field_caps", r.URL.Path)
		jsonResponse(w, 200, `{"indices":[],"fields":{}}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.FieldCaps(context.Background(), &FieldCapsRequest{Body: map[string]any{"fields": []string{"foo"}}})
	assert.NoError(t, err)
}

func TestUnitSearchFieldCapsNotFoundReturnsEmpty(t *testing.T) {
	srv := errServer(404)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.FieldCaps(context.Background(), &FieldCapsRequest{Indices: []string{"missing"}})
	require.NoError(t, err)
	assert.Empty(t, resp.Indices)
	assert.Empty(t, resp.Fields)
}

func TestUnitSearchFieldCapsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.FieldCaps(context.Background(), &FieldCapsRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

func TestUnitSearchFieldCapsNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.FieldCaps(context.Background(), &FieldCapsRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

func TestUnitSearchFieldCapsBadJSON(t *testing.T) {
	srv := badJSONServer()
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.FieldCaps(context.Background(), &FieldCapsRequest{Indices: []string{"idx"}})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// ScriptsPainlessExec
// -----------------------------------------------------------------------------

func TestUnitSearchScriptsPainlessExec(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"result":42}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.ScriptsPainlessExec(context.Background(), map[string]any{"script": map[string]any{"source": "return 42"}})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_scripts/painless/_execute", gotPath)
	assert.Contains(t, gotBody, "return 42")
	assert.Contains(t, string(resp), "result")
}

func TestUnitSearchScriptsPainlessExecNilBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, `{"result":0}`)
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.ScriptsPainlessExec(context.Background(), nil)
	assert.NoError(t, err)
}

func TestUnitSearchScriptsPainlessExecError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	_, err := svc.ScriptsPainlessExec(context.Background(), map[string]any{})
	assert.Error(t, err)
}

func TestUnitSearchScriptsPainlessExecNetworkError(t *testing.T) {
	svc := NewSearchService(deadClient(), testLogger())
	_, err := svc.ScriptsPainlessExec(context.Background(), map[string]any{})
	assert.Error(t, err)
}
