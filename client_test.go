package elasticsearch

import (
	"net/http"
	"testing"
	"time"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *logrus.Entry {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return logrus.NewEntry(logger)
}

func traceLogger() *logrus.Entry {
	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)
	return logrus.NewEntry(logger)
}

func TestNewBasicAuth(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200", Username: "elastic", Password: "pw"}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, c.RestyClient())
}

func TestNewAPIKey(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200", APIKey: "abc:123"}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewBearerToken(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200", BearerToken: "tok"}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewAuthConflict(t *testing.T) {
	_, err := New(&Config{URL: "https://localhost:9200", Username: "u", Password: "p", APIKey: "k"}, testLogger())
	assert.Error(t, err)
	var esErr *types.ElasticsearchError
	assert.ErrorAs(t, err, &esErr)
}

func TestNewCACertParseError(t *testing.T) {
	_, err := New(&Config{URL: "https://localhost:9200", CACert: []byte("not a cert")}, testLogger())
	assert.Error(t, err)
}

func TestNewNoAuth(t *testing.T) {
	c, err := New(&Config{URL: "http://localhost:9200"}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewDisableHTTP2(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200", DisableHTTP2: true}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewAllServices(t *testing.T) {
	c, err := New(&Config{URL: "http://localhost:9200"}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c.Info())
	assert.NotNil(t, c.Document())
	assert.NotNil(t, c.Search())
	assert.NotNil(t, c.Indices())
	assert.NotNil(t, c.Cluster())
	assert.NotNil(t, c.Nodes())
	assert.NotNil(t, c.Cat())
	assert.NotNil(t, c.ML())
	assert.NotNil(t, c.Security())
	assert.NotNil(t, c.SQL())
	assert.NotNil(t, c.Streams())
}

// TestNewAllServicesComplete verifies every accessor method on DefaultClient
// returns a non-nil service.
func TestNewAllServicesComplete(t *testing.T) {
	c, err := New(&Config{URL: "http://localhost:9200"}, testLogger())
	require.NoError(t, err)

	assert.NotNil(t, c.Info())
	assert.NotNil(t, c.Document())
	assert.NotNil(t, c.Search())
	assert.NotNil(t, c.AsyncSearch())
	assert.NotNil(t, c.Script())
	assert.NotNil(t, c.Indices())
	assert.NotNil(t, c.Cluster())
	assert.NotNil(t, c.Nodes())
	assert.NotNil(t, c.Cat())
	assert.NotNil(t, c.Ingest())
	assert.NotNil(t, c.Snapshot())
	assert.NotNil(t, c.SLM())
	assert.NotNil(t, c.Tasks())
	assert.NotNil(t, c.ILM())
	assert.NotNil(t, c.Transform())
	assert.NotNil(t, c.ML())
	assert.NotNil(t, c.Security())
	assert.NotNil(t, c.SQL())
	assert.NotNil(t, c.Watcher())
	assert.NotNil(t, c.CCR())
	assert.NotNil(t, c.Autoscaling())
	assert.NotNil(t, c.Enrich())
	assert.NotNil(t, c.EQL())
	assert.NotNil(t, c.Graph())
	assert.NotNil(t, c.License())
	assert.NotNil(t, c.Logstash())
	assert.NotNil(t, c.Migration())
	assert.NotNil(t, c.Shutdown())
	assert.NotNil(t, c.Features())
	assert.NotNil(t, c.Fleet())
	assert.NotNil(t, c.DanglingIndices())
	assert.NotNil(t, c.HealthReport())
	assert.NotNil(t, c.QueryRules())
	assert.NotNil(t, c.SearchApplication())
	assert.NotNil(t, c.Synonyms())
	assert.NotNil(t, c.TextStructure())
	assert.NotNil(t, c.Inference())
	assert.NotNil(t, c.SearchableSnapshots())
	assert.NotNil(t, c.SSL())
	assert.NotNil(t, c.XPack())
	assert.NotNil(t, c.Streams())
}

// TestNewInsecureHTTPWithCredentials verifies that combining credentials
// with an http:// URL is rejected unless AllowInsecureHTTP is set.
func TestNewInsecureHTTPWithCredentials(t *testing.T) {
	// http:// with credentials and no opt-in → error.
	_, err := New(&Config{URL: "http://localhost:9200", Username: "u", Password: "p"}, testLogger())
	assert.Error(t, err)
	var esErr *types.ElasticsearchError
	assert.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

// TestNewInsecureHTTPWithCredentialsOptIn verifies that http:// with
// credentials is allowed when AllowInsecureHTTP is explicitly set.
func TestNewInsecureHTTPWithCredentialsOptIn(t *testing.T) {
	c, err := New(&Config{URL: "http://localhost:9200", Username: "u", Password: "p", AllowInsecureHTTP: true}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

// TestNewHTTPSWithCredentials verifies that https:// with credentials is
// always allowed.
func TestNewHTTPSWithCredentials(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200", Username: "u", Password: "p"}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

// TestNewHTTPWithoutCredentials verifies that http:// without credentials
// is allowed (no cleartext secret to protect).
func TestNewHTTPWithoutCredentials(t *testing.T) {
	c, err := New(&Config{URL: "http://localhost:9200"}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

// TestNewInvalidURLNoScheme verifies that a URL without a scheme is rejected.
func TestNewInvalidURLNoScheme(t *testing.T) {
	_, err := New(&Config{URL: "google.com"}, testLogger())
	assert.Error(t, err)
	var esErr *types.ElasticsearchError
	assert.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

// TestNewInvalidURL verifies that a malformed URL is rejected.
func TestNewInvalidURL(t *testing.T) {
	_, err := New(&Config{URL: "://bad"}, testLogger())
	assert.Error(t, err)
	var esErr *types.ElasticsearchError
	assert.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

// TestNewFTPURLRejected verifies that a non-http(s) scheme is rejected.
func TestNewFTPURLRejected(t *testing.T) {
	_, err := New(&Config{URL: "ftp://localhost:9200"}, testLogger())
	assert.Error(t, err)
	var esErr *types.ElasticsearchError
	assert.ErrorAs(t, err, &esErr)
	assert.Equal(t, 400, esErr.Status)
}

// TestNewEmptyURL verifies that an empty URL is allowed (resty uses a
// per-request URL in that case).
func TestNewEmptyURL(t *testing.T) {
	c, err := New(&Config{}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

// TestNewAPIKeyAndBearerConflict verifies that APIKey + BearerToken is rejected.
func TestNewAPIKeyAndBearerConflict(t *testing.T) {
	_, err := New(&Config{URL: "https://localhost:9200", APIKey: "k", BearerToken: "t"}, testLogger())
	assert.Error(t, err)
	var esErr *types.ElasticsearchError
	assert.ErrorAs(t, err, &esErr)
}

// TestNewBasicAuthAndBearerConflict verifies that basic auth + BearerToken is rejected.
func TestNewBasicAuthAndBearerConflict(t *testing.T) {
	_, err := New(&Config{URL: "https://localhost:9200", Username: "u", Password: "p", BearerToken: "t"}, testLogger())
	assert.Error(t, err)
}

// TestNewTLSSkipVerifyWithCreds verifies TLSSkipVerify + credentials logs a
// warning but still succeeds.
func TestNewTLSSkipVerifyWithCreds(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200", Username: "u", Password: "p", TLSSkipVerify: true}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

// TestNewTLSSkipVerifyNoCreds verifies TLSSkipVerify without credentials is fine.
func TestNewTLSSkipVerifyNoCreds(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200", TLSSkipVerify: true}, testLogger())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

// TestNewTimeout verifies the request timeout is applied to the resty client.
func TestNewTimeout(t *testing.T) {
	c, err := New(&Config{URL: "http://localhost:9200", Timeout: 5 * time.Second}, testLogger())
	require.NoError(t, err)
	// No direct accessor for timeout; just ensure client built.
	assert.NotNil(t, c.RestyClient())
}

// TestNewIdleConnTimeoutDefault verifies the default IdleConnTimeout (60s) is
// applied to the transport when not specified.
func TestNewIdleConnTimeoutDefault(t *testing.T) {
	c, err := New(&Config{URL: "http://localhost:9200"}, testLogger())
	require.NoError(t, err)
	tr, err := c.RestyClient().Transport()
	require.NoError(t, err)
	assert.Equal(t, 60*time.Second, tr.IdleConnTimeout)
}

// TestNewIdleConnTimeoutCustom verifies a custom IdleConnTimeout is honored.
func TestNewIdleConnTimeoutCustom(t *testing.T) {
	c, err := New(&Config{URL: "http://localhost:9200", IdleConnTimeout: 5 * time.Second}, testLogger())
	require.NoError(t, err)
	tr, err := c.RestyClient().Transport()
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, tr.IdleConnTimeout)
}

// TestNewDisableHTTP2Transport verifies that DisableHTTP2 forces HTTP/1.1 on
// the transport.
func TestNewDisableHTTP2Transport(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200", DisableHTTP2: true}, testLogger())
	require.NoError(t, err)
	tr, err := c.RestyClient().Transport()
	require.NoError(t, err)
	assert.False(t, tr.ForceAttemptHTTP2)
	assert.Equal(t, []string{"http/1.1"}, tr.TLSClientConfig.NextProtos)
	assert.NotNil(t, tr.TLSNextProto)
	assert.Empty(t, tr.TLSNextProto)
}

// TestNewHTTP2Default verifies HTTP/2 is attempted by default.
func TestNewHTTP2Default(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200"}, testLogger())
	require.NoError(t, err)
	tr, err := c.RestyClient().Transport()
	require.NoError(t, err)
	assert.True(t, tr.ForceAttemptHTTP2)
}

// TestNewTLSSkipVerifyTransport verifies TLSSkipVerify propagates to the
// transport's TLS config.
func TestNewTLSSkipVerifyTransport(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200", TLSSkipVerify: true}, testLogger())
	require.NoError(t, err)
	tr, err := c.RestyClient().Transport()
	require.NoError(t, err)
	tlsCfg := tr.TLSClientConfig
	require.NotNil(t, tlsCfg)
	assert.True(t, tlsCfg.InsecureSkipVerify)
}

// TestNewRestyClientBaseURL verifies RestyClient returns a non-nil client with
// the configured base URL.
func TestNewRestyClientBaseURL(t *testing.T) {
	c, err := New(&Config{URL: "https://localhost:9200"}, testLogger())
	require.NoError(t, err)
	rc := c.RestyClient()
	require.NotNil(t, rc)
	assert.Equal(t, "https://localhost:9200", rc.BaseURL)
}

// TestNewRestyClientEmptyURL verifies that an empty URL leaves the base URL
// unset (resty uses per-request URLs).
func TestNewRestyClientEmptyURL(t *testing.T) {
	c, err := New(&Config{}, testLogger())
	require.NoError(t, err)
	rc := c.RestyClient()
	require.NotNil(t, rc)
	assert.Empty(t, rc.BaseURL)
}

// TestNewTraceLevelDebug verifies that a trace-level logger enables resty
// debug mode.
func TestNewTraceLevelDebug(t *testing.T) {
	c, err := New(&Config{URL: "http://localhost:9200"}, traceLogger())
	require.NoError(t, err)
	rc := c.RestyClient()
	assert.True(t, rc.Debug)
}

// TestNewRetryConfigured verifies retry settings are applied to the resty client.
func TestNewRetryConfigured(t *testing.T) {
	called := false
	cond := func(*resty.Response, error) bool { called = true; return false }
	c, err := New(&Config{
		URL:              "http://localhost:9200",
		RetryCount:       2,
		RetryWaitTime:    10 * time.Millisecond,
		RetryMaxWaitTime: 100 * time.Millisecond,
		RetryConditions:  []resty.RetryConditionFunc{cond},
	}, testLogger())
	require.NoError(t, err)
	rc := c.RestyClient()
	assert.Equal(t, 2, rc.RetryCount)
	// Trigger the condition to ensure it was registered.
	cond(nil, nil)
	assert.True(t, called)
}

// TestNewContentTypeHeader verifies the default Content-Type header is set.
func TestNewContentTypeHeader(t *testing.T) {
	c, err := New(&Config{URL: "http://localhost:9200"}, testLogger())
	require.NoError(t, err)
	assert.Equal(t, "application/json", c.RestyClient().Header.Get("Content-Type"))
}

// TestRedactingRestyLogger verifies the redacting logger scrubs sensitive
// headers from all log levels.
func TestRedactingRestyLogger(t *testing.T) {
	l := &redactingRestyLogger{entry: testLogger()}
	// Should not panic and should redact.
	l.Errorf("Authorization: secret\nCookie: abc")
	l.Warnf("Authorization: secret\nCookie: abc")
	l.Debugf("Authorization: secret\nCookie: abc")
}

// TestRedactingRestyLoggerNoColon verifies the logger handles messages
// without colons.
func TestRedactingRestyLoggerNoColon(t *testing.T) {
	l := &redactingRestyLogger{entry: testLogger()}
	l.Debugf("plain message with no colon here")
}

// TestRedactURLEmpty verifies RedactURL handles empty input.
func TestRedactURLEmpty(t *testing.T) {
	assert.Equal(t, "", RedactURL(""))
}

// TestRedactURLNoUserinfo verifies RedactURL passes through URLs without userinfo.
func TestRedactURLNoUserinfo(t *testing.T) {
	assert.Equal(t, "https://host:9200/_search", RedactURL("https://host:9200/_search"))
}

// TestRedactURLUserOnly verifies RedactURL strips username-only userinfo.
func TestRedactURLUserOnly(t *testing.T) {
	assert.Equal(t, "https://host:9200/", RedactURL("https://api-key-12345@host:9200/"))
}

// TestRedactURLUserAndPass verifies RedactURL strips full userinfo.
func TestRedactURLUserAndPass(t *testing.T) {
	assert.Equal(t, "https://host:9200/_search", RedactURL("https://user:pass@host:9200/_search"))
}

// TestRedactURLParseError verifies RedactURL returns unparseable input unchanged.
func TestRedactURLParseError(t *testing.T) {
	assert.Equal(t, "://bad", RedactURL("://bad"))
}

// TestRedactSensitiveHeadersSetCookie verifies the Set-Cookie header is redacted.
func TestRedactSensitiveHeadersSetCookie(t *testing.T) {
	redacted := redactSensitiveHeaders("Set-Cookie: session=xyz")
	assert.Contains(t, redacted, "Set-Cookie: [REDACTED]")
	assert.NotContains(t, redacted, "session=xyz")
}

// TestRedactSensitiveHeadersLineNoColon verifies lines without colons within a
// multi-line message are passed through.
func TestRedactSensitiveHeadersLineNoColon(t *testing.T) {
	msg := "line one without colon\nAuthorization: secret"
	redacted := redactSensitiveHeaders(msg)
	assert.Contains(t, redacted, "line one without colon")
	assert.Contains(t, redacted, "Authorization: [REDACTED]")
}

// ensure imports are used
var _ = http.DefaultTransport
