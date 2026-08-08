package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// readBody reads and returns the request body as a string.
func readBody(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	return string(b)
}

// queryMap returns the request URL query parameters as a map.
func queryMap(r *http.Request) map[string]string {
	m := make(map[string]string)
	for k, vs := range r.URL.Query() {
		if len(vs) > 0 {
			m[k] = vs[0]
		}
	}
	return m
}

func testLogger() *logrus.Entry {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return logrus.NewEntry(logger)
}

func restyClient(ts *httptest.Server) *resty.Client {
	c := resty.New()
	c.SetBaseURL(ts.URL)
	c.SetHeader("Content-Type", "application/json")
	return c
}

func deadClient() *resty.Client {
	return resty.New().
		SetBaseURL("http://127.0.0.1:1").
		SetTimeout(100 * time.Millisecond)
}

func jsonResponse(w http.ResponseWriter, statusCode int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprint(w, body)
}

func errServer(status int) *httptest.Server {
	errJSON := `{"error":{"type":"test_error","reason":"test error"},"status":` + fmt.Sprint(status) + `}`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = fmt.Fprint(w, errJSON)
	}))
}

func badJSONServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `not json at all`)
	}))
}

func ptrInt(n int) *int       { return &n }
func ptrInt64(n int64) *int64 { return &n }
func ptrBool(b bool) *bool    { return &b }
