package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitDanglingIndicesList(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"dangling_indices":[{"index_uuid":"uuid-1","index_name":"old-index","creation_date":123}]}`)
	}))
	defer srv.Close()
	svc := NewDanglingIndicesService(restyClient(srv), testLogger())
	resp, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_dangling", gotPath)
	require.Len(t, resp.DanglingIndices, 1)
	assert.Equal(t, "uuid-1", resp.DanglingIndices[0].IndexUUID)
}

func TestUnitDanglingIndicesListError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDanglingIndicesService(restyClient(srv), testLogger())
	_, err := svc.List(context.Background())
	assert.Error(t, err)
}

func TestUnitDanglingIndicesListNetworkError(t *testing.T) {
	svc := NewDanglingIndicesService(deadClient(), testLogger())
	_, err := svc.List(context.Background())
	assert.Error(t, err)
}

func TestUnitDanglingIndicesImport(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewDanglingIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Import(context.Background(), "uuid-1", &DanglingIndicesImportParams{
		AcceptDataLoss: ptrBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_dangling/uuid-1", gotPath)
	assert.Equal(t, "true", gotQuery["accept_data_loss"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitDanglingIndicesImportEmptyUUID(t *testing.T) {
	svc := NewDanglingIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Import(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestUnitDanglingIndicesImportError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDanglingIndicesService(restyClient(srv), testLogger())
	_, err := svc.Import(context.Background(), "uuid-1", nil)
	assert.Error(t, err)
}

func TestUnitDanglingIndicesImportNetworkError(t *testing.T) {
	svc := NewDanglingIndicesService(deadClient(), testLogger())
	_, err := svc.Import(context.Background(), "uuid-1", nil)
	assert.Error(t, err)
}

func TestUnitDanglingIndicesDelete(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewDanglingIndicesService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background(), "uuid-1", &DanglingIndicesImportParams{
		AcceptDataLoss: ptrBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_dangling/uuid-1", gotPath)
	assert.Equal(t, "true", gotQuery["accept_data_loss"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitDanglingIndicesDeleteEmptyUUID(t *testing.T) {
	svc := NewDanglingIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Delete(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestUnitDanglingIndicesDeleteError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewDanglingIndicesService(restyClient(srv), testLogger())
	_, err := svc.Delete(context.Background(), "uuid-1", nil)
	assert.Error(t, err)
}

func TestUnitDanglingIndicesDeleteNetworkError(t *testing.T) {
	svc := NewDanglingIndicesService(deadClient(), testLogger())
	_, err := svc.Delete(context.Background(), "uuid-1", nil)
	assert.Error(t, err)
}

func TestUnitDanglingIndicesB3Timeouts(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewDanglingIndicesService(restyClient(srv), testLogger())
	_, err := svc.Import(context.Background(), "uuid-1", &DanglingIndicesImportParams{
		AcceptDataLoss: ptrBool(false),
		MasterTimeout:  "m1",
		Timeout:        "t1",
	})
	require.NoError(t, err)
	assert.Equal(t, "false", gotQuery["accept_data_loss"])
	assert.Equal(t, "m1", gotQuery["master_timeout"])
	assert.Equal(t, "t1", gotQuery["timeout"])

	_, err = svc.Delete(context.Background(), "uuid-2", &DanglingIndicesImportParams{MasterTimeout: "m2", Timeout: "t2"})
	require.NoError(t, err)
	assert.Equal(t, "m2", gotQuery["master_timeout"])
	assert.Equal(t, "t2", gotQuery["timeout"])
}
