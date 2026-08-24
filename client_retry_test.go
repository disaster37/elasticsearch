package elasticsearch

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRetryConditionsNetworkError(t *testing.T) {
	conds := DefaultRetryConditions()
	for _, c := range conds {
		assert.True(t, c(nil, errors.New("connection reset by peer")))
		assert.True(t, c(nil, errors.New("connection refused")))
		assert.True(t, c(nil, errors.New("i/o timeout")))
		assert.False(t, c(nil, nil))
	}
}

// TestDefaultRetryConditionsNetworkErrors enumerates each network error
// substring that should trigger a retry.
func TestDefaultRetryConditionsNetworkErrors(t *testing.T) {
	conds := DefaultRetryConditions()
	cond := conds[0]
	cases := []string{
		"connection reset by peer",
		"connection refused",
		"timeout waiting for response",
		"i/o timeout",
		"network is unreachable",
		"broken pipe",
		"EOF",
	}
	for _, msg := range cases {
		assert.True(t, cond(nil, errors.New(msg)), "expected retry for %q", msg)
	}
}

// TestDefaultRetryConditionsUnknownNetworkError verifies that any non-nil
// error with no response still triggers a retry (defensive default).
func TestDefaultRetryConditionsUnknownNetworkError(t *testing.T) {
	cond := DefaultRetryConditions()[0]
	assert.True(t, cond(nil, errors.New("some unknown socket error")))
}

// TestDefaultRetryConditions429 verifies 429 responses are retried.
func TestDefaultRetryConditions429(t *testing.T) {
	cond := DefaultRetryConditions()[0]
	resp := &resty.Response{RawResponse: &http.Response{StatusCode: 429}}
	assert.True(t, cond(resp, nil))
}

// TestDefaultRetryConditions5xx verifies 5xx responses (except 501) are retried.
func TestDefaultRetryConditions5xx(t *testing.T) {
	cond := DefaultRetryConditions()[0]
	for code := 500; code < 600; code++ {
		resp := &resty.Response{RawResponse: &http.Response{StatusCode: code}}
		if code == 501 {
			assert.False(t, cond(resp, nil), "501 must not be retried")
		} else {
			assert.True(t, cond(resp, nil), "%d should be retried", code)
		}
	}
}

// TestDefaultRetryConditions501NotRetried verifies 501 is explicitly excluded.
func TestDefaultRetryConditions501NotRetried(t *testing.T) {
	cond := DefaultRetryConditions()[0]
	resp := &resty.Response{RawResponse: &http.Response{StatusCode: 501}}
	assert.False(t, cond(resp, nil))
}

// TestDefaultRetryConditions4xxNotRetried verifies normal 4xx responses are
// not retried.
func TestDefaultRetryConditions4xxNotRetried(t *testing.T) {
	cond := DefaultRetryConditions()[0]
	for _, code := range []int{400, 401, 403, 404, 409} {
		resp := &resty.Response{RawResponse: &http.Response{StatusCode: code}}
		assert.False(t, cond(resp, nil), "%d should not be retried", code)
	}
}

// TestESSearchRetryConditionsLength verifies ESSearchRetryConditions includes
// the default conditions plus the ES search condition.
func TestESSearchRetryConditionsLength(t *testing.T) {
	conds := ESSearchRetryConditions()
	assert.GreaterOrEqual(t, len(conds), 2, "should have default + ES search conditions")
}

// TestESSearchRetryConditionsTransientErrors verifies the ES search condition
// retries on transient error types in the response body.
func TestESSearchRetryConditionsTransientErrors(t *testing.T) {
	conds := ESSearchRetryConditions()
	searchCond := conds[len(conds)-1]
	transient := []string{
		"search_phase_execution_exception",
		"too_many_buckets_exception",
		"circuit_breaking_exception",
	}
	for _, body := range transient {
		resp := &resty.Response{RawResponse: &http.Response{StatusCode: 200}}
		resp.SetBody([]byte(fmt.Sprintf(`{"error":{"type":"%s","reason":"x"}}`, body)))
		assert.True(t, searchCond(resp, nil), "expected retry for %s", body)
	}
}

// TestESSearchRetryConditionsNoMatch verifies the ES search condition does not
// retry on bodies without transient errors.
func TestESSearchRetryConditionsNoMatch(t *testing.T) {
	conds := ESSearchRetryConditions()
	searchCond := conds[len(conds)-1]
	resp := &resty.Response{RawResponse: &http.Response{StatusCode: 200}}
	resp.SetBody([]byte(`{"took":1,"hits":{"total":0}}`))
	assert.False(t, searchCond(resp, nil))
}

// TestESSearchRetryConditionsNilResponse verifies the ES search condition
// returns false when the response is nil.
func TestESSearchRetryConditionsNilResponse(t *testing.T) {
	conds := ESSearchRetryConditions()
	searchCond := conds[len(conds)-1]
	assert.False(t, searchCond(nil, nil))
}

// TestRetryOn503 verifies that 5xx (except 501) responses are retried under
// DefaultRetryConditions.
func TestRetryOn503(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprint(w, `{"error":{"type":"es_error","reason":"unavailable"},"status":503}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"cluster_name":"test","version":{"number":"9.0.0"}}`)
	}))
	defer srv.Close()

	client, err := New(&Config{
		Addresses:        []string{srv.URL},
		RetryCount:       3,
		RetryWaitTime:    1 * time.Millisecond,
		RetryMaxWaitTime: 1 * time.Millisecond,
		RetryConditions:  DefaultRetryConditions(),
	}, testLogger())
	require.NoError(t, err)

	resp, err := client.RestyClient().R().Get("/_test")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

// TestRetryOn429 verifies that 429 responses are retried.
func TestRetryOn429(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprint(w, `{"error":{"type":"too_many_requests","reason":"rate limited"},"status":429}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"ok":true}`)
	}))
	defer srv.Close()

	client, err := New(&Config{
		Addresses:        []string{srv.URL},
		RetryCount:       2,
		RetryWaitTime:    1 * time.Millisecond,
		RetryMaxWaitTime: 1 * time.Millisecond,
		RetryConditions:  DefaultRetryConditions(),
	}, testLogger())
	require.NoError(t, err)

	resp, err := client.RestyClient().R().Get("/_test")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

// TestRetryOnSearchPhaseException verifies that ESSearchRetryConditions retries
// on a transient search_phase_execution_exception in the response body.
func TestRetryOnSearchPhaseException(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprint(w, `{"error":{"type":"search_phase_execution_exception","reason":"failed"},"status":503}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"took":1,"hits":{"total":{"value":0,"relation":"eq"},"hits":[]}}`)
	}))
	defer srv.Close()

	client, err := New(&Config{
		Addresses:        []string{srv.URL},
		RetryCount:       2,
		RetryWaitTime:    1 * time.Millisecond,
		RetryMaxWaitTime: 1 * time.Millisecond,
		RetryConditions:  ESSearchRetryConditions(),
	}, testLogger())
	require.NoError(t, err)

	resp, err := client.RestyClient().R().Get("/_search")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

// TestRetryOnTooManyBuckets verifies ESSearchRetryConditions retries on
// too_many_buckets_exception.
func TestRetryOnTooManyBuckets(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"error":{"type":"too_many_buckets_exception","reason":"too many"}}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"took":1}`)
	}))
	defer srv.Close()

	client, err := New(&Config{
		Addresses:        []string{srv.URL},
		RetryCount:       2,
		RetryWaitTime:    1 * time.Millisecond,
		RetryMaxWaitTime: 1 * time.Millisecond,
		RetryConditions:  ESSearchRetryConditions(),
	}, testLogger())
	require.NoError(t, err)

	resp, err := client.RestyClient().R().Get("/_search")
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
	_ = resp
}

// TestRetryOnCircuitBreaking verifies ESSearchRetryConditions retries on
// circuit_breaking_exception.
func TestRetryOnCircuitBreaking(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"error":{"type":"circuit_breaking_exception","reason":"breaker tripped"}}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"took":1}`)
	}))
	defer srv.Close()

	client, err := New(&Config{
		Addresses:        []string{srv.URL},
		RetryCount:       2,
		RetryWaitTime:    1 * time.Millisecond,
		RetryMaxWaitTime: 1 * time.Millisecond,
		RetryConditions:  ESSearchRetryConditions(),
	}, testLogger())
	require.NoError(t, err)

	_, err = client.RestyClient().R().Get("/_search")
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

// ensure types import is used.
var _ = types.ElasticsearchError{}
