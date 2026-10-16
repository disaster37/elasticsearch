package elasticsearch

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/disaster37/elasticsearch/v9/api"
	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// DefaultRetryConditions returns the default retry conditions that handle:
//   - Network errors (no response received): connection reset/refused, timeout,
//     i/o timeout, network unreachable, broken pipe, EOF
//   - 429 Too Many Requests
//   - 5xx server error status codes (excluding 501 Not Implemented)
//
// These conditions are automatically applied by [New] when RetryCount > 0
// and no custom conditions are supplied. Use this function to build custom
// retry logic that includes the defaults.
func DefaultRetryConditions() []resty.RetryConditionFunc {
	return []resty.RetryConditionFunc{
		func(res *resty.Response, err error) bool {
			if res == nil {
				if err != nil {
					errStr := err.Error()
					networkErrors := []string{
						"connection reset by peer",
						"connection refused",
						"timeout",
						"i/o timeout",
						"network is unreachable",
						"broken pipe",
						"EOF",
					}
					for _, netErr := range networkErrors {
						if strings.Contains(strings.ToLower(errStr), netErr) {
							return true
						}
					}
					return true
				}
				return false
			}
			status := res.StatusCode()
			return status == 429 || (status >= 500 && status != 501)
		},
	}
}

// ESSearchRetryConditions returns retry conditions optimized for long-running
// Elasticsearch search operations. It includes [DefaultRetryConditions] plus
// retrying on transient Elasticsearch error types found in the response body:
// search_phase_execution_exception, too_many_buckets_exception, and
// circuit_breaking_exception.
func ESSearchRetryConditions() []resty.RetryConditionFunc {
	defaultConds := DefaultRetryConditions()
	return append(
		defaultConds,
		func(res *resty.Response, err error) bool {
			if res == nil {
				return false
			}
			body := string(res.Body())
			return strings.Contains(body, "search_phase_execution_exception") ||
				strings.Contains(body, "too_many_buckets_exception") ||
				strings.Contains(body, "circuit_breaking_exception")
		},
	)
}

// RedactURL strips any embedded userinfo (username and/or password) from a
// raw URL. This is more aggressive than net/url.Redacted, which preserves
// the username when no password is present — we remove the entire userinfo
// section because tokens, API keys, and short-lived credentials are often
// placed in the username slot alone (e.g. "https://api-key-12345@host"),
// and we must not leak them into logs or exported telemetry attributes.
// If the URL cannot be parsed it is returned unchanged.
//
// The implementation lives in the types package ([types.RedactURL]) so that
// the api package — which cannot import the root package (import cycle) —
// can share the same redaction logic.
func RedactURL(rawURL string) string {
	return types.RedactURL(rawURL)
}

// sensitiveHeaderNames is the set of HTTP headers whose values must never be
// written to logs. The lookup is case-insensitive.
var sensitiveHeaderNames = map[string]struct{}{
	"authorization": {},
	"cookie":        {},
	"set-cookie":    {},
}

// redactSensitiveHeaders scans a (potentially multi-line) message and replaces
// the values of sensitive HTTP headers with "[REDACTED]". It is used to scrub
// resty's debug dump, which formats each header on its own line as
// "\t<right-aligned name>: <value>". Any line whose leading token (before the
// first colon) matches a sensitive header name (case-insensitive) has its
// value stripped. Lines without a colon, or whose key is not sensitive, are
// passed through unchanged.
func redactSensitiveHeaders(msg string) string {
	if !strings.Contains(msg, ":") {
		return msg
	}
	lines := strings.Split(msg, "\n")
	for i, line := range lines {
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:idx]))
		if _, ok := sensitiveHeaderNames[key]; ok {
			lines[i] = line[:idx] + ": [REDACTED]"
		}
	}
	return strings.Join(lines, "\n")
}

// redactingRestyLogger adapts a logrus.Entry to resty's Logger interface while
// redacting sensitive HTTP headers from every message before it is forwarded.
// resty emits its full request/response debug dump (including all headers)
// through a single Debugf call, so scrubbing here prevents Authorization,
// security tokens, and cookies from leaking into logs at trace level.
type redactingRestyLogger struct {
	entry *logrus.Entry
}

func (l *redactingRestyLogger) Errorf(format string, v ...interface{}) {
	l.entry.Errorf("%s", redactSensitiveHeaders(fmt.Sprintf(format, v...)))
}

func (l *redactingRestyLogger) Warnf(format string, v ...interface{}) {
	l.entry.Warnf("%s", redactSensitiveHeaders(fmt.Sprintf(format, v...)))
}

func (l *redactingRestyLogger) Debugf(format string, v ...interface{}) {
	l.entry.Debugf("%s", redactSensitiveHeaders(fmt.Sprintf(format, v...)))
}

// Client is the main entry point for all Elasticsearch operations.
//
// Obtain a Client via [New]. All API groups are accessible as methods on this
// interface.
type Client interface {
	// RestyClient returns the underlying resty client for middleware
	// configuration (e.g. adding OpenTelemetry tracing).
	RestyClient() *resty.Client

	Info() api.InfoService
	Document() api.DocumentService
	Search() api.SearchService
	AsyncSearch() api.AsyncSearchService
	Script() api.ScriptService
	Indices() api.IndicesService
	Cluster() api.ClusterService
	Nodes() api.NodesService
	Cat() api.CatService
	Ingest() api.IngestService
	Snapshot() api.SnapshotService
	SLM() api.SlmService
	Tasks() api.TasksService
	ILM() api.IlmService
	Transform() api.TransformService
	ML() api.MlService
	Security() api.SecurityService
	SQL() api.SqlService
	Watcher() api.WatcherService
	CCR() api.CcrService
	Autoscaling() api.AutoscalingService
	Enrich() api.EnrichService
	EQL() api.EqlService
	Graph() api.GraphService
	License() api.LicenseService
	Logstash() api.LogstashService
	Migration() api.MigrationService
	Shutdown() api.ShutdownService
	Features() api.FeaturesService
	Fleet() api.FleetService
	DanglingIndices() api.DanglingIndicesService
	HealthReport() api.HealthReportService
	QueryRules() api.QueryRulesService
	SearchApplication() api.SearchApplicationService
	Synonyms() api.SynonymsService
	TextStructure() api.TextStructureService
	Inference() api.InferenceService
	SearchableSnapshots() api.SearchableSnapshotsService
	SSL() api.SslService
	XPack() api.XpackService
	Streams() api.StreamsService
}

// Config holds configuration for the Elasticsearch client.
type Config struct {
	// Addresses is a list of Elasticsearch node URLs to connect to. Requests
	// are distributed across the addresses in round-robin order.
	//
	// Configuring more than one address automatically enables retry and
	// failover: when RetryCount is left at its zero value it is set to the
	// number of addresses (each attempt targets the next address, so every
	// node is tried at least once), and the default retry conditions are
	// installed when no custom conditions are supplied. There is currently no
	// opt-out field for multi-address configurations — configure a single
	// address for strict no-retry semantics.
	//
	// The maximum number of addresses is maxAddresses (1000). Exceeding this
	// limit returns an error from New.
	//
	// Example:
	//   Addresses: []string{"https://node1:9200", "https://node2:9200"}
	//
	// An empty (nil or zero-length) slice is allowed: in that case the client
	// has no base URL and every request must supply an absolute URL.
	Addresses []string

	// Username for HTTP basic authentication.
	Username string
	// Password for HTTP basic authentication.
	Password string

	// APIKey enables API-key authentication. Sent as
	// "Authorization: ApiKey <value>". Mutually exclusive with basic auth and
	// BearerToken. The value may be the base64 "id:api_key" form or a raw key.
	APIKey string

	// BearerToken enables bearer-token authentication. Sent as
	// "Authorization: Bearer <value>". Mutually exclusive with basic auth and
	// APIKey.
	BearerToken string

	// TLSSkipVerify disables TLS certificate verification.
	// WARNING: Only set to true for development. For production, use CACert.
	// When credentials are configured together with TLSSkipVerify, a warning
	// is logged because the credentials are then protected only by the
	// (unverified) TLS layer.
	TLSSkipVerify bool

	// AllowInsecureHTTP explicitly opts in to sending credentials over a
	// plaintext http:// address. By default New rejects any configuration that
	// combines credentials (Username/Password, APIKey, or BearerToken) with
	// an http:// address, because the credentials would travel in cleartext.
	// Set this to true only for local development against a non-TLS
	// Elasticsearch; never enable it in production.
	AllowInsecureHTTP bool

	// CACert is a PEM-encoded CA certificate used to verify the server.
	// If empty and TLSSkipVerify is false, the system cert pool is used.
	CACert []byte

	// Timeout is the HTTP request timeout. Zero means no timeout.
	Timeout time.Duration

	// IdleConnTimeout is the maximum amount of time an idle (keep-alive)
	// connection is kept in the pool before being closed. Default 60s.
	IdleConnTimeout time.Duration

	// DisableHTTP2 forces the client to use HTTP/1.1 instead of negotiating
	// HTTP/2 via TLS ALPN.
	DisableHTTP2 bool

	// RetryCount is the maximum number of retry attempts for failed requests.
	//
	// The default is 0 (no retries) when a single address is configured.
	// When more than one address is configured, New automatically sets
	// RetryCount to the number of addresses so a failing request fails over
	// to the next node: with N addresses, N retries allow up to N+1 attempts
	// and every address is tried before the request gives up. An explicitly
	// set RetryCount > 0 is never overwritten.
	//
	// Whenever RetryCount > 0 and no custom RetryConditions are supplied, the
	// default conditions from DefaultRetryConditions are installed.
	//
	// There is currently no opt-out field for multi-address configurations —
	// configure a single address for strict no-retry semantics.
	RetryCount int

	// RetryWaitTime is the minimum wait time between retry attempts.
	// Default is 100ms if not specified.
	RetryWaitTime time.Duration

	// RetryMaxWaitTime is the maximum wait time between retry attempts.
	// Default is 2s if not specified.
	RetryMaxWaitTime time.Duration

	// RetryConditions are custom functions that determine if a request should
	// be retried. Caller-provided conditions are used as-is and are never
	// replaced or extended by the defaults.
	//
	// When RetryCount > 0 (explicitly set, or auto-enabled because multiple
	// addresses are configured) and RetryConditions is empty,
	// DefaultRetryConditions is applied automatically: network errors, 429
	// Too Many Requests, and 5xx errors except 501.
	RetryConditions []resty.RetryConditionFunc
}

// DefaultClient is the default [Client] implementation returned by [New].
type DefaultClient struct {
	client *resty.Client
	logger *logrus.Entry

	info                api.InfoService
	document            api.DocumentService
	search              api.SearchService
	asyncSearch         api.AsyncSearchService
	script              api.ScriptService
	indices             api.IndicesService
	cluster             api.ClusterService
	nodes               api.NodesService
	cat                 api.CatService
	ingest              api.IngestService
	snapshot            api.SnapshotService
	slm                 api.SlmService
	tasks               api.TasksService
	ilm                 api.IlmService
	transform           api.TransformService
	ml                  api.MlService
	security            api.SecurityService
	sql                 api.SqlService
	watcher             api.WatcherService
	ccr                 api.CcrService
	autoscaling         api.AutoscalingService
	enrich              api.EnrichService
	eql                 api.EqlService
	graph               api.GraphService
	license             api.LicenseService
	logstash            api.LogstashService
	migration           api.MigrationService
	shutdown            api.ShutdownService
	features            api.FeaturesService
	fleet               api.FleetService
	danglingIndices     api.DanglingIndicesService
	healthReport        api.HealthReportService
	queryRules          api.QueryRulesService
	searchApplication   api.SearchApplicationService
	synonyms            api.SynonymsService
	textStructure       api.TextStructureService
	inference           api.InferenceService
	searchableSnapshots api.SearchableSnapshotsService
	ssl                 api.SslService
	xpack               api.XpackService
	streams             api.StreamsService
}

// New creates a new [Client] connecting to an Elasticsearch cluster.
//
// Exactly one authentication mechanism may be configured: basic auth
// (Username/Password), API key (APIKey), or bearer token (BearerToken).
// Setting more than one returns an error.
func New(cfg *Config, logger *logrus.Entry) (Client, error) {
	c := resty.New()

	// Always send application/json so that string/[]byte bodies are not sent
	// with the resty default of "text/plain; charset=utf-8".
	c.SetHeader("Content-Type", "application/json")

	// Parse and validate the configured addresses. An empty list is allowed
	// (resty then uses per-request absolute URLs). normalizeAddresses enforces
	// the http/https scheme.
	parsedAddrs, err := normalizeAddresses(cfg.Addresses)
	if err != nil {
		return nil, err
	}

	// Multi-address failover is enabled by default: when several nodes are
	// configured and the caller left RetryCount at its zero value, retry up to
	// len(parsedAddrs) times. The pool advances one address per attempt, so N
	// retries yield up to N+1 attempts and every address is tried at least
	// once before the request gives up. An explicit RetryCount > 0 always wins.
	if len(parsedAddrs) > 1 && cfg.RetryCount == 0 {
		cfg.RetryCount = len(parsedAddrs)
	}

	// Retries need conditions: when retries are enabled (explicitly or by the
	// multi-address default above) and the caller supplied none, install the
	// documented defaults (network errors, 429, 5xx except 501). Caller
	// conditions are never replaced.
	if cfg.RetryCount > 0 && len(cfg.RetryConditions) == 0 {
		cfg.RetryConditions = DefaultRetryConditions()
	}

	pool := newConnectionPool(parsedAddrs)

	// SetBaseURL makes RestyClient().BaseURL reflect the first address and
	// remains the fallback for single-address and absolute-URL requests.
	if len(parsedAddrs) > 0 {
		c.SetBaseURL(parsedAddrs[0].String())
	}

	// Round-robin across addresses. No-op for 0 or 1 address. Registered before
	// the logging hook below so the logged URL is the actual node.
	c.OnBeforeRequest(func(_ *resty.Client, req *resty.Request) error {
		pool.rewrite(req)
		return nil
	})

	// Auth: exactly one mechanism may be set.
	authSet := 0
	if cfg.Username != "" || cfg.Password != "" {
		authSet++
	}
	if cfg.APIKey != "" {
		authSet++
	}
	if cfg.BearerToken != "" {
		authSet++
	}
	if authSet > 1 {
		return nil, &types.ElasticsearchError{
			Details: &types.ElasticsearchErrorDetails{
				Reason: "at most one of basic auth (Username/Password), APIKey, or BearerToken may be set",
			},
		}
	}

	switch {
	case cfg.Username != "" || cfg.Password != "":
		c.SetBasicAuth(cfg.Username, cfg.Password)
	case cfg.APIKey != "":
		c.SetHeader("Authorization", "ApiKey "+cfg.APIKey)
	case cfg.BearerToken != "":
		c.SetHeader("Authorization", "Bearer "+cfg.BearerToken)
	}

	// Guard against sending credentials in cleartext over any http:// address.
	// An empty address list is allowed (resty uses a per-request URL).
	credsConfigured := cfg.Username != "" || cfg.Password != "" || cfg.APIKey != "" || cfg.BearerToken != ""
	if credsConfigured {
		for _, u := range parsedAddrs {
			if u.Scheme == "http" && !cfg.AllowInsecureHTTP {
				return nil, &types.ElasticsearchError{Status: 400, Details: &types.ElasticsearchErrorDetails{
					Reason: "credentials configured over insecure http:// URL; use https:// or set Config.AllowInsecureHTTP=true",
				}}
			}
		}
	}
	if cfg.TLSSkipVerify && credsConfigured {
		logger.Warn("TLSSkipVerify is enabled while credentials are configured; credentials are protected only by the (unverified) TLS layer")
	}

	if cfg.Timeout > 0 {
		c.SetTimeout(cfg.Timeout)
	}

	if cfg.IdleConnTimeout == 0 {
		cfg.IdleConnTimeout = 60 * time.Second
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.TLSSkipVerify,
	}

	if len(cfg.CACert) > 0 {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(cfg.CACert) {
			return nil, &types.ElasticsearchError{
				Details: &types.ElasticsearchErrorDetails{Reason: "failed to parse CA certificate"},
			}
		}
		tlsConfig.RootCAs = pool
	}

	c.SetTLSClientConfig(tlsConfig)

	// Customize the underlying transport so the connection pool can be
	// aligned with upstream proxies. Cloning the default transport keeps
	// resty/net-http defaults (proxy from env, dialer, HTTP/2, etc.).
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	if cfg.IdleConnTimeout > 0 {
		transport.IdleConnTimeout = cfg.IdleConnTimeout
	}
	if cfg.DisableHTTP2 {
		transport.ForceAttemptHTTP2 = false
		transport.TLSClientConfig.NextProtos = []string{"http/1.1"}
		transport.TLSNextProto = make(map[string]func(authority string, conn *tls.Conn) http.RoundTripper)
	}
	c.SetTransport(transport)

	c.OnBeforeRequest(func(_ *resty.Client, req *resty.Request) error {
		logger.WithFields(logrus.Fields{
			"http_method": req.Method,
			"http_path":   RedactURL(req.URL),
		}).Debug("Sending request to Elasticsearch")
		return nil
	})
	c.OnAfterResponse(func(_ *resty.Client, resp *resty.Response) error {
		logger.WithFields(logrus.Fields{
			"http_method":   resp.Request.Method,
			"http_path":     RedactURL(resp.Request.URL),
			"http_status":   resp.StatusCode(),
			"http_duration": resp.Time().String(),
		}).Debug("Received response from Elasticsearch")
		// Do NOT hard-fail on a missing X-Elastic-Product header: proxies and
		// load balancers commonly strip custom response headers. Only log at
		// debug level so misconfigured proxies remain diagnosable.
		if resp.Header().Get("X-Elastic-Product") != "Elasticsearch" {
			logger.Debug("Response is missing the X-Elastic-Product header; cannot verify the server is Elasticsearch")
		}
		return nil
	})

	// Route all resty logging (including the debug dump) through a logger that
	// redacts sensitive headers, and disable body dumping in debug mode as
	// defense in depth — request/response bodies may also carry credentials.
	c.SetLogger(&redactingRestyLogger{entry: logger})

	// Enable resty debug output when the logger is at trace level.
	if logger.Logger.IsLevelEnabled(logrus.TraceLevel) {
		c.SetDebug(true)
		c.SetDebugBodyLimit(0)
	}

	if cfg.RetryCount > 0 {
		c.SetRetryCount(cfg.RetryCount)
		if cfg.RetryWaitTime > 0 {
			c.SetRetryWaitTime(cfg.RetryWaitTime)
		}
		if cfg.RetryMaxWaitTime > 0 {
			c.SetRetryMaxWaitTime(cfg.RetryMaxWaitTime)
		}
		// cfg.RetryConditions is guaranteed non-empty here: the block above
		// installs DefaultRetryConditions when none were supplied.
		for _, condition := range cfg.RetryConditions {
			c.AddRetryCondition(condition)
		}
	}

	return &DefaultClient{
		client:              c,
		logger:              logger,
		info:                api.NewInfoService(c, logger),
		document:            api.NewDocumentService(c, logger),
		search:              api.NewSearchService(c, logger),
		asyncSearch:         api.NewAsyncSearchService(c, logger),
		script:              api.NewScriptService(c, logger),
		indices:             api.NewIndicesService(c, logger),
		cluster:             api.NewClusterService(c, logger),
		nodes:               api.NewNodesService(c, logger),
		cat:                 api.NewCatService(c, logger),
		ingest:              api.NewIngestService(c, logger),
		snapshot:            api.NewSnapshotService(c, logger),
		slm:                 api.NewSlmService(c, logger),
		tasks:               api.NewTasksService(c, logger),
		ilm:                 api.NewIlmService(c, logger),
		transform:           api.NewTransformService(c, logger),
		ml:                  api.NewMlService(c, logger),
		security:            api.NewSecurityService(c, logger),
		sql:                 api.NewSqlService(c, logger),
		watcher:             api.NewWatcherService(c, logger),
		ccr:                 api.NewCcrService(c, logger),
		autoscaling:         api.NewAutoscalingService(c, logger),
		enrich:              api.NewEnrichService(c, logger),
		eql:                 api.NewEqlService(c, logger),
		graph:               api.NewGraphService(c, logger),
		license:             api.NewLicenseService(c, logger),
		logstash:            api.NewLogstashService(c, logger),
		migration:           api.NewMigrationService(c, logger),
		shutdown:            api.NewShutdownService(c, logger),
		features:            api.NewFeaturesService(c, logger),
		fleet:               api.NewFleetService(c, logger),
		danglingIndices:     api.NewDanglingIndicesService(c, logger),
		healthReport:        api.NewHealthReportService(c, logger),
		queryRules:          api.NewQueryRulesService(c, logger),
		searchApplication:   api.NewSearchApplicationService(c, logger),
		synonyms:            api.NewSynonymsService(c, logger),
		textStructure:       api.NewTextStructureService(c, logger),
		inference:           api.NewInferenceService(c, logger),
		searchableSnapshots: api.NewSearchableSnapshotsService(c, logger),
		ssl:                 api.NewSslService(c, logger),
		xpack:               api.NewXpackService(c, logger),
		streams:             api.NewStreamsService(c, logger),
	}, nil
}

func (c *DefaultClient) RestyClient() *resty.Client { return c.client }

func (c *DefaultClient) Info() api.InfoService                           { return c.info }
func (c *DefaultClient) Document() api.DocumentService                   { return c.document }
func (c *DefaultClient) Search() api.SearchService                       { return c.search }
func (c *DefaultClient) AsyncSearch() api.AsyncSearchService             { return c.asyncSearch }
func (c *DefaultClient) Script() api.ScriptService                       { return c.script }
func (c *DefaultClient) Indices() api.IndicesService                     { return c.indices }
func (c *DefaultClient) Cluster() api.ClusterService                     { return c.cluster }
func (c *DefaultClient) Nodes() api.NodesService                         { return c.nodes }
func (c *DefaultClient) Cat() api.CatService                             { return c.cat }
func (c *DefaultClient) Ingest() api.IngestService                       { return c.ingest }
func (c *DefaultClient) Snapshot() api.SnapshotService                   { return c.snapshot }
func (c *DefaultClient) SLM() api.SlmService                             { return c.slm }
func (c *DefaultClient) Tasks() api.TasksService                         { return c.tasks }
func (c *DefaultClient) ILM() api.IlmService                             { return c.ilm }
func (c *DefaultClient) Transform() api.TransformService                 { return c.transform }
func (c *DefaultClient) ML() api.MlService                               { return c.ml }
func (c *DefaultClient) Security() api.SecurityService                   { return c.security }
func (c *DefaultClient) SQL() api.SqlService                             { return c.sql }
func (c *DefaultClient) Watcher() api.WatcherService                     { return c.watcher }
func (c *DefaultClient) CCR() api.CcrService                             { return c.ccr }
func (c *DefaultClient) Autoscaling() api.AutoscalingService             { return c.autoscaling }
func (c *DefaultClient) Enrich() api.EnrichService                       { return c.enrich }
func (c *DefaultClient) EQL() api.EqlService                             { return c.eql }
func (c *DefaultClient) Graph() api.GraphService                         { return c.graph }
func (c *DefaultClient) License() api.LicenseService                     { return c.license }
func (c *DefaultClient) Logstash() api.LogstashService                   { return c.logstash }
func (c *DefaultClient) Migration() api.MigrationService                 { return c.migration }
func (c *DefaultClient) Shutdown() api.ShutdownService                   { return c.shutdown }
func (c *DefaultClient) Features() api.FeaturesService                   { return c.features }
func (c *DefaultClient) Fleet() api.FleetService                         { return c.fleet }
func (c *DefaultClient) DanglingIndices() api.DanglingIndicesService     { return c.danglingIndices }
func (c *DefaultClient) HealthReport() api.HealthReportService           { return c.healthReport }
func (c *DefaultClient) QueryRules() api.QueryRulesService               { return c.queryRules }
func (c *DefaultClient) SearchApplication() api.SearchApplicationService { return c.searchApplication }
func (c *DefaultClient) Synonyms() api.SynonymsService                   { return c.synonyms }
func (c *DefaultClient) TextStructure() api.TextStructureService         { return c.textStructure }
func (c *DefaultClient) Inference() api.InferenceService                 { return c.inference }
func (c *DefaultClient) SearchableSnapshots() api.SearchableSnapshotsService {
	return c.searchableSnapshots
}
func (c *DefaultClient) SSL() api.SslService         { return c.ssl }
func (c *DefaultClient) XPack() api.XpackService     { return c.xpack }
func (c *DefaultClient) Streams() api.StreamsService { return c.streams }

// ensure fmt is used (error formatting helpers may be added later).
var _ = fmt.Sprintf
