// client_retry_test.go contains unit tests for retry configuration and
// behavior: DefaultRetryConditions, ESSearchRetryConditions, retry execution
// against httptest servers, and the automatic retry defaults applied by New()
// when multiple addresses are configured. No Elasticsearch cluster is required.
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

// TestNewAutoRetryCountMultipleAddresses verifies that New auto-enables retry
// with a count equal to the number of configured addresses and installs the
// default retry conditions when no retry configuration is given.
func TestNewAutoRetryCountMultipleAddresses(t *testing.T) {
	c, err := New(&Config{
		Addresses: []string{"http://n1:9200", "http://n2:9200", "http://n3:9200"},
	}, testLogger())
	require.NoError(t, err)

	rc := c.RestyClient()
	assert.Equal(t, 3, rc.RetryCount)
	assert.NotEmpty(t, rc.RetryConditions)
}

// TestNewSingleAddressDefaultRetryDisabled verifies that a single address with
// the zero-value RetryCount keeps retries disabled and registers no conditions.
func TestNewSingleAddressDefaultRetryDisabled(t *testing.T) {
	c, err := New(&Config{Addresses: []string{"http://n1:9200"}}, testLogger())
	require.NoError(t, err)

	rc := c.RestyClient()
	assert.Equal(t, 0, rc.RetryCount)
	assert.Empty(t, rc.RetryConditions)
}

// TestNewExplicitRetryCountNotOverwritten verifies that an explicitly set
// RetryCount is preserved when multiple addresses are configured and that
// caller-provided conditions are not replaced by the defaults.
func TestNewExplicitRetryCountNotOverwritten(t *testing.T) {
	called := false
	cond := func(*resty.Response, error) bool {
		called = true
		return false
	}

	c, err := New(&Config{
		Addresses:       []string{"http://n1:9200", "http://n2:9200", "http://n3:9200"},
		RetryCount:      2,
		RetryConditions: []resty.RetryConditionFunc{cond},
	}, testLogger())
	require.NoError(t, err)

	rc := c.RestyClient()
	assert.Equal(t, 2, rc.RetryCount)
	require.Len(t, rc.RetryConditions, 1)
	rc.RetryConditions[0](nil, nil)
	assert.True(t, called)
}

// TestNewExplicitRetryCountInstallsDefaultConditions verifies that enabling
// retries without custom conditions installs the documented defaults.
func TestNewExplicitRetryCountInstallsDefaultConditions(t *testing.T) {
	c, err := New(&Config{
		Addresses:  []string{"http://n1:9200"},
		RetryCount: 2,
	}, testLogger())
	require.NoError(t, err)

	rc := c.RestyClient()
	assert.Equal(t, 2, rc.RetryCount)
	require.NotEmpty(t, rc.RetryConditions)

	// The installed condition is the default one: network errors, 429, 5xx.
	cond := rc.RetryConditions[0]
	assert.True(t, cond(nil, errors.New("connection reset by peer")))
	resp := &resty.Response{RawResponse: &http.Response{StatusCode: http.StatusServiceUnavailable}}
	assert.True(t, cond(resp, nil))
}

// ensure types import is used.
var _ = types.ElasticsearchError{}
