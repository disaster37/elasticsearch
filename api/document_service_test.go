package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitDocumentIndex(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		assert.Equal(t, "refresh=true", r.URL.RawQuery)
		jsonResponse(w, 201, `{"_index":"idx","_id":"1","_version":1,"result":"created","_seq_no":0,"_primary_term":1,"status":201}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())

	resp, err := svc.Index(context.Background(), &IndexRequest{
		Index:  "idx",
		Id:     "1",
		Body:   map[string]any{"f": "v"},
		Params: &IndexParams{Refresh: RefreshTrue},
	})
	assert.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/idx/_doc/1", gotPath)
	assert.Equal(t, "idx", resp.Index)
	assert.Equal(t, "created", resp.Result)
	assert.Equal(t, 201, resp.Status)
}

func TestUnitDocumentIndexAutoID(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		jsonResponse(w, 201, `{"_index":"idx","_id":"auto","_version":1,"result":"created"}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.Index(context.Background(), &IndexRequest{Index: "idx", Body: map[string]any{"f": "v"}})
	assert.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx/_doc", gotPath)
}

func TestUnitDocumentIndexValidation(t *testing.T) {
	svc := NewDocumentService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Index(context.Background(), &IndexRequest{})
	assert.Error(t, err)
}

func TestUnitDocumentGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_doc/1", r.URL.Path)
		jsonResponse(w, 200, `{"_index":"idx","_id":"1","_version":1,"found":true,"_source":{"f":"v"}}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), &GetRequest{Index: "idx", Id: "1"})
	assert.NoError(t, err)
	assert.True(t, resp.Found)
	assert.Equal(t, "idx", resp.Index)
}

func TestUnitDocumentGetNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"_index":"idx","_id":"missing","found":false}`))
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), &GetRequest{Index: "idx", Id: "missing"})
	assert.NoError(t, err)
	assert.False(t, resp.Found)
}

func TestUnitDocumentDeleteNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"_index":"idx","_id":"missing","_version":1,"result":"not_found"}`))
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background(), &DeleteRequest{Index: "idx", Id: "missing"})
	assert.Error(t, err)
	assert.Equal(t, "not_found", resp.Result)
}

func TestUnitDocumentExists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "HEAD", r.Method)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	ok, err := svc.Exists(context.Background(), "idx", "1", &GetParams{})
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUnitDocumentExistsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	ok, err := svc.Exists(context.Background(), "idx", "1", &GetParams{})
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestUnitDocumentBulk(t *testing.T) {
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		assert.Equal(t, "/_bulk", r.URL.Path)
		jsonResponse(w, 200, `{"took":1,"errors":false,"items":[]}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.Bulk(context.Background(), &BulkRequest{Body: `{"index":{}}\n{"f":"v"}\n`})
	assert.NoError(t, err)
	assert.Equal(t, "application/x-ndjson", gotCT)
}

func TestUnitDocumentBulkEnsuresTrailingNewline(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		body = string(buf)
		jsonResponse(w, 200, `{"took":1,"errors":false,"items":[]}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.Bulk(context.Background(), &BulkRequest{Body: `{"index":{}}\n{"f":"v"}`})
	assert.NoError(t, err)
	assert.True(t, len(body) > 0)
	assert.Equal(t, "\n", string(body[len(body)-1]))
}

func TestUnitDocumentNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), &GetRequest{Index: "idx", Id: "1"})
	assert.Error(t, err)
}

// ---- Create ----

func TestUnitDocumentCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/test-index/_doc/1/_create", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)
		jsonResponse(w, 201, `{"_index":"test-index","_id":"1","_version":1,"result":"created","_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.Create(context.Background(), &CreateRequest{
		Index: "test-index",
		Id:    "1",
		Body:  map[string]any{"field": "value"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "created", resp.Result)
	assert.Equal(t, "test-index", resp.Index)
}

func TestUnitDocumentCreateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.Create(context.Background(), &CreateRequest{
		Index: "test-index",
		Id:    "1",
		Body:  map[string]any{"field": "value"},
	})
	assert.Error(t, err)
}

func TestUnitDocumentCreateNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.Create(context.Background(), &CreateRequest{
		Index: "test-index",
		Id:    "1",
		Body:  map[string]any{"field": "value"},
	})
	assert.Error(t, err)
}

// ---- GetSource ----

func TestUnitDocumentGetSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_source/1", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"f":"v"}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	raw, err := svc.GetSource(context.Background(), &GetSourceRequest{Index: "idx", Id: "1"})
	assert.NoError(t, err)
	assert.Equal(t, `{"f":"v"}`, string(raw))
}

func TestUnitDocumentGetSourceError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.GetSource(context.Background(), &GetSourceRequest{Index: "idx", Id: "1"})
	assert.Error(t, err)
}

func TestUnitDocumentGetSourceNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.GetSource(context.Background(), &GetSourceRequest{Index: "idx", Id: "1"})
	assert.Error(t, err)
}

// ---- ExistsSource ----

func TestUnitDocumentExistsSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "HEAD", r.Method)
		assert.Equal(t, "/idx/_source/1", r.URL.Path)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	ok, err := svc.ExistsSource(context.Background(), "idx", "1", &GetSourceParams{})
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUnitDocumentExistsSourceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	ok, err := svc.ExistsSource(context.Background(), "idx", "1", &GetSourceParams{})
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestUnitDocumentExistsSourceError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.ExistsSource(context.Background(), "idx", "1", &GetSourceParams{})
	assert.Error(t, err)
}

func TestUnitDocumentExistsSourceNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.ExistsSource(context.Background(), "idx", "1", &GetSourceParams{})
	assert.Error(t, err)
}

// ---- MultiGet ----

func TestUnitDocumentMultiGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_mget", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"docs":[{"_index":"idx","_id":"1","found":true,"_source":{"f":"v"}}]}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.MultiGet(context.Background(), &MgetRequest{
		Index: "idx",
		Body:  map[string]any{"ids": []string{"1"}},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Docs, 1)
	assert.True(t, resp.Docs[0].Found)
}

func TestUnitDocumentMultiGetNoIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_mget", r.URL.Path)
		jsonResponse(w, 200, `{"docs":[]}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.MultiGet(context.Background(), &MgetRequest{Body: map[string]any{"ids": []string{"1"}}})
	assert.NoError(t, err)
}

func TestUnitDocumentMultiGetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.MultiGet(context.Background(), &MgetRequest{
		Index: "idx",
		Body:  map[string]any{"ids": []string{"1"}},
	})
	assert.Error(t, err)
}

func TestUnitDocumentMultiGetNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.MultiGet(context.Background(), &MgetRequest{
		Index: "idx",
		Body:  map[string]any{"ids": []string{"1"}},
	})
	assert.Error(t, err)
}

// ---- Update ----

func TestUnitDocumentUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_update/1", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"_index":"idx","_id":"1","_version":2,"result":"updated"}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.Update(context.Background(), &UpdateRequest{
		Index: "idx",
		Id:    "1",
		Body:  map[string]any{"doc": map[string]any{"f": "v2"}},
	})
	assert.NoError(t, err)
	assert.Equal(t, "updated", resp.Result)
	assert.Equal(t, int64(2), resp.Version)
}

func TestUnitDocumentUpdateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.Update(context.Background(), &UpdateRequest{
		Index: "idx",
		Id:    "1",
		Body:  map[string]any{"doc": map[string]any{"f": "v2"}},
	})
	assert.Error(t, err)
}

func TestUnitDocumentUpdateNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.Update(context.Background(), &UpdateRequest{
		Index: "idx",
		Id:    "1",
		Body:  map[string]any{"doc": map[string]any{"f": "v2"}},
	})
	assert.Error(t, err)
}

// ---- Bulk (error + network; success covered above) ----

func TestUnitDocumentBulkError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.Bulk(context.Background(), &BulkRequest{Body: `{"index":{}}\n{"f":"v"}\n`})
	assert.Error(t, err)
}

func TestUnitDocumentBulkNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.Bulk(context.Background(), &BulkRequest{Body: `{"index":{}}\n{"f":"v"}\n`})
	assert.Error(t, err)
}

// ---- DeleteByQuery ----

func TestUnitDocumentDeleteByQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_delete_by_query", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"took":12,"timed_out":false,"total":1,"deleted":1,"batches":1,"version_conflicts":0,"noops":0}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.DeleteByQuery(context.Background(), &DeleteByQueryRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"query": map[string]any{"match_all": map[string]any{}}},
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.Deleted)
	assert.Equal(t, int64(1), resp.Total)
}

func TestUnitDocumentDeleteByQueryError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.DeleteByQuery(context.Background(), &DeleteByQueryRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"query": map[string]any{"match_all": map[string]any{}}},
	})
	assert.Error(t, err)
}

func TestUnitDocumentDeleteByQueryNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.DeleteByQuery(context.Background(), &DeleteByQueryRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"query": map[string]any{"match_all": map[string]any{}}},
	})
	assert.Error(t, err)
}

// ---- UpdateByQuery ----

func TestUnitDocumentUpdateByQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_update_by_query", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"took":10,"timed_out":false,"total":2,"updated":2,"batches":1,"version_conflicts":0,"noops":0}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.UpdateByQuery(context.Background(), &UpdateByQueryRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"query": map[string]any{"match_all": map[string]any{}}},
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Updated)
}

func TestUnitDocumentUpdateByQueryError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.UpdateByQuery(context.Background(), &UpdateByQueryRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"query": map[string]any{"match_all": map[string]any{}}},
	})
	assert.Error(t, err)
}

func TestUnitDocumentUpdateByQueryNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.UpdateByQuery(context.Background(), &UpdateByQueryRequest{
		Indices: []string{"idx"},
		Body:    map[string]any{"query": map[string]any{"match_all": map[string]any{}}},
	})
	assert.Error(t, err)
}

// ---- Reindex ----

func TestUnitDocumentReindex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_reindex", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"took":20,"timed_out":false,"total":3,"created":3,"batches":1,"version_conflicts":0,"noops":0}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.Reindex(context.Background(), &ReindexRequest{
		Body: map[string]any{"source": map[string]any{"index": "src"}, "dest": map[string]any{"index": "dst"}},
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), resp.Created)
}

func TestUnitDocumentReindexError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.Reindex(context.Background(), &ReindexRequest{
		Body: map[string]any{"source": map[string]any{"index": "src"}, "dest": map[string]any{"index": "dst"}},
	})
	assert.Error(t, err)
}

func TestUnitDocumentReindexNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.Reindex(context.Background(), &ReindexRequest{
		Body: map[string]any{"source": map[string]any{"index": "src"}, "dest": map[string]any{"index": "dst"}},
	})
	assert.Error(t, err)
}

// ---- Rethrottle (shared by DeleteByQuery/UpdateByQuery/Reindex) ----

func TestUnitDocumentDeleteByQueryRethrottle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_delete_by_query/task1/_rethrottle", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "10", r.URL.Query().Get("requests_per_second"))
		jsonResponse(w, 200, `{"nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.DeleteByQueryRethrottle(context.Background(), &RethrottleRequest{
		TaskId:            "task1",
		RequestsPerSecond: 10,
	})
	assert.NoError(t, err)
}

func TestUnitDocumentDeleteByQueryRethrottleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.DeleteByQueryRethrottle(context.Background(), &RethrottleRequest{
		TaskId:            "task1",
		RequestsPerSecond: 10,
	})
	assert.Error(t, err)
}

func TestUnitDocumentDeleteByQueryRethrottleNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.DeleteByQueryRethrottle(context.Background(), &RethrottleRequest{
		TaskId:            "task1",
		RequestsPerSecond: 10,
	})
	assert.Error(t, err)
}

func TestUnitDocumentUpdateByQueryRethrottle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_update_by_query/task1/_rethrottle", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.UpdateByQueryRethrottle(context.Background(), &RethrottleRequest{
		TaskId:            "task1",
		RequestsPerSecond: 5,
	})
	assert.NoError(t, err)
}

func TestUnitDocumentUpdateByQueryRethrottleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.UpdateByQueryRethrottle(context.Background(), &RethrottleRequest{
		TaskId:            "task1",
		RequestsPerSecond: 5,
	})
	assert.Error(t, err)
}

func TestUnitDocumentUpdateByQueryRethrottleNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.UpdateByQueryRethrottle(context.Background(), &RethrottleRequest{
		TaskId:            "task1",
		RequestsPerSecond: 5,
	})
	assert.Error(t, err)
}

func TestUnitDocumentReindexRethrottle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_reindex/task1/_rethrottle", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.ReindexRethrottle(context.Background(), &RethrottleRequest{
		TaskId:            "task1",
		RequestsPerSecond: 0,
	})
	assert.NoError(t, err)
}

func TestUnitDocumentReindexRethrottleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.ReindexRethrottle(context.Background(), &RethrottleRequest{
		TaskId:            "task1",
		RequestsPerSecond: 0,
	})
	assert.Error(t, err)
}

func TestUnitDocumentReindexRethrottleNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.ReindexRethrottle(context.Background(), &RethrottleRequest{
		TaskId:            "task1",
		RequestsPerSecond: 0,
	})
	assert.Error(t, err)
}

// ---- Explain ----

func TestUnitDocumentExplain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_explain/1", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"_index":"idx","_id":"1","matched":true,"explanation":{"value":1.0,"description":"match","details":[]}}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.Explain(context.Background(), &ExplainRequest{
		Index: "idx",
		Id:    "1",
		Body:  map[string]any{"query": map[string]any{"match": map[string]any{"f": "v"}}},
	})
	assert.NoError(t, err)
	assert.True(t, resp.Matched)
	assert.Equal(t, "idx", resp.Index)
}

func TestUnitDocumentExplainNoBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"_index":"idx","_id":"1","matched":false,"explanation":{}}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.Explain(context.Background(), &ExplainRequest{Index: "idx", Id: "1"})
	assert.NoError(t, err)
	assert.False(t, resp.Matched)
}

func TestUnitDocumentExplainError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.Explain(context.Background(), &ExplainRequest{
		Index: "idx",
		Id:    "1",
		Body:  map[string]any{"query": map[string]any{"match": map[string]any{"f": "v"}}},
	})
	assert.Error(t, err)
}

func TestUnitDocumentExplainNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.Explain(context.Background(), &ExplainRequest{
		Index: "idx",
		Id:    "1",
		Body:  map[string]any{"query": map[string]any{"match": map[string]any{"f": "v"}}},
	})
	assert.Error(t, err)
}

// ---- TermVectors ----

func TestUnitDocumentTermVectors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_termvectors/1", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"_index":"idx","_id":"1","_version":1,"found":true,"took":0,"term_vectors":{}}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.TermVectors(context.Background(), &TermVectorsRequest{
		Index: "idx",
		Id:    "1",
		Body:  map[string]any{"doc": map[string]any{"f": "v"}},
	})
	assert.NoError(t, err)
	assert.True(t, resp.Found)
}

func TestUnitDocumentTermVectorsNoId(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_termvectors", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		jsonResponse(w, 200, `{"_index":"idx","_version":1,"found":true,"took":0,"term_vectors":{}}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.TermVectors(context.Background(), &TermVectorsRequest{Index: "idx"})
	assert.NoError(t, err)
}

func TestUnitDocumentTermVectorsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.TermVectors(context.Background(), &TermVectorsRequest{
		Index: "idx",
		Id:    "1",
		Body:  map[string]any{"doc": map[string]any{"f": "v"}},
	})
	assert.Error(t, err)
}

func TestUnitDocumentTermVectorsNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.TermVectors(context.Background(), &TermVectorsRequest{
		Index: "idx",
		Id:    "1",
		Body:  map[string]any{"doc": map[string]any{"f": "v"}},
	})
	assert.Error(t, err)
}

// ---- MultiTermVectors ----

func TestUnitDocumentMultiTermVectors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_mtermvectors", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"docs":[{"_index":"idx","_id":"1","found":true,"term_vectors":{}}]}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.MultiTermVectors(context.Background(), &MtermvectorsRequest{
		Index: "idx",
		Body:  map[string]any{"ids": []string{"1"}},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Docs, 1)
}

func TestUnitDocumentMultiTermVectorsNoIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/_mtermvectors", r.URL.Path)
		jsonResponse(w, 200, `{"docs":[]}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.MultiTermVectors(context.Background(), &MtermvectorsRequest{
		Body: map[string]any{"ids": []string{"1"}},
	})
	assert.NoError(t, err)
}

func TestUnitDocumentMultiTermVectorsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.MultiTermVectors(context.Background(), &MtermvectorsRequest{
		Index: "idx",
		Body:  map[string]any{"ids": []string{"1"}},
	})
	assert.Error(t, err)
}

func TestUnitDocumentMultiTermVectorsNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.MultiTermVectors(context.Background(), &MtermvectorsRequest{
		Index: "idx",
		Body:  map[string]any{"ids": []string{"1"}},
	})
	assert.Error(t, err)
}

// ---- TermsEnum ----

func TestUnitDocumentTermsEnum(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_terms_enum", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0},"terms":["a","b"],"complete":true}`)
	}))
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	resp, err := svc.TermsEnum(context.Background(), &TermsEnumRequest{
		Index: "idx",
		Body:  map[string]any{"field": "f", "string": "a"},
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, resp.Terms)
	assert.True(t, resp.Complete)
}

func TestUnitDocumentTermsEnumError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDocumentService(restyClient(srv), testLogger())
	_, err := svc.TermsEnum(context.Background(), &TermsEnumRequest{
		Index: "idx",
		Body:  map[string]any{"field": "f"},
	})
	assert.Error(t, err)
}

func TestUnitDocumentTermsEnumNetworkError(t *testing.T) {
	svc := NewDocumentService(deadClient(), testLogger())
	_, err := svc.TermsEnum(context.Background(), &TermsEnumRequest{
		Index: "idx",
		Body:  map[string]any{"field": "f"},
	})
	assert.Error(t, err)
}
