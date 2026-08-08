package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---- doReqText ----

func TestUnitDoReqText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/text", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("hello world"))
	}))
	defer srv.Close()
	out, err := doReqText(testLogger(), context.Background(), restyClient(srv), "GET", "/text", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", out)
}

func TestUnitDoReqTextWithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, `{"q":"v"}`, readBody(r))
		w.WriteHeader(200)
		_, _ = w.Write([]byte("accepted"))
	}))
	defer srv.Close()
	out, err := doReqText(testLogger(), context.Background(), restyClient(srv), "POST", "/text", `{"q":"v"}`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "accepted", out)
}

func TestUnitDoReqTextError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	_, err := doReqText(testLogger(), context.Background(), restyClient(srv), "GET", "/text", nil, nil)
	assert.Error(t, err)
}

func TestUnitDoReqTextNetworkError(t *testing.T) {
	_, err := doReqText(testLogger(), context.Background(), deadClient(), "GET", "/text", nil, nil)
	assert.Error(t, err)
}

// ---- floatStr ----

func TestUnitFloatStr(t *testing.T) {
	assert.Equal(t, "0", floatStr(0))
	assert.Equal(t, "1", floatStr(1))
	assert.Equal(t, "-1", floatStr(-1))
	assert.Equal(t, "3.14", floatStr(3.14))
	assert.Equal(t, "-3.14", floatStr(-3.14))
	assert.Equal(t, "1000000", floatStr(1000000))
	assert.Equal(t, "0.5", floatStr(0.5))
	assert.Equal(t, "0.00000001", floatStr(1e-8))
}
