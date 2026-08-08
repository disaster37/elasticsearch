# Implementation Plan: Elasticsearch Go Client (`github.com/disaster37/elasticsearch/v9`)

Target repo: `/projects/elasticsearch` (remote `github.com/disaster37/elasticsearch.git`, currently empty).
Reference architecture: `/projects/opensearch` (`github.com/disaster37/opensearch/v4`) — follow it exactly.
API source of truth: Elasticsearch **9.x** `rest-api-spec` JSON files.

---

## 1. Locked decisions

| # | Decision |
|---|----------|
| 1 | Target Elasticsearch **9.x**. Spec cloned from `elastic/elasticsearch`, checked out at the latest `v9.*` release tag. Acceptance image: `docker.elastic.co/elasticsearch/elasticsearch:<latest 9.x patch>` |
| 2 | Go module path: `github.com/disaster37/elasticsearch/v9`, Go version aligned with reference (1.26 in go.mod; CI uses the version declared in go.mod) |
| 3 | Port the **full `querydsl` package** from the opensearch client (rename imports, fix ES incompatibilities) |
| 4 | Auth: Basic (Username/Password) **+ API key + Bearer token**; mutually exclusive, validated in `New` |
| 5 | CI: full parity with opensearch CI — Dagger module consuming disaster37's Daggerverse modules `golang`, `codecov`, `git-module` (`github.com/disaster37/dagger-library-go/*@2.0.0`); build + vet + lint + format + unit tests + acceptance tests + CodeCov upload + git push-back of `coverage.out` |
| 6 | CI ES container: **security enabled** (HTTPS, fixed `ELASTIC_PASSWORD`, trial license, `path.repo` for snapshot tests) |
| 7 | **Fully typed** request/response structs for every endpoint |
| 8 | Coverage: every endpoint in the 9.x spec **except `_internal.*`**; deprecated endpoints implemented with `// Deprecated:` comments. Rollup and `_knn_search` were removed in ES 9 — documented as N/A after spec verification |
| 9 | Acceptance tests: every service, happy-path CRUD cycles + key error cases (404/409), `-tags integration` |
| 10 | Libraries: `go-resty/resty/v2`, `sirupsen/logrus`, `goccy/go-json`, `go-playground/validator/v10`, `stretchr/testify`, `go.opentelemetry.io/otel*` (trace package), `mockgen` via `go:generate` |

---

## 2. Source-of-truth setup (Task 0 — implementer runs this first)

```bash
# REST spec (authoritative for endpoints + parameters)
git clone https://github.com/elastic/elasticsearch /tmp/elasticsearch
cd /tmp/elasticsearch && git checkout <latest v9.* tag>   # e.g. v9.1.x
ls rest-api-spec/src/main/resources/rest-api-spec/api/*.json   # ~600 files

# Official Go client (reference for naming/patterns only — we do NOT copy its architecture)
git clone https://github.com/elastic/go-elasticsearch /tmp/go-elasticsearch
```

Spec JSON anatomy (per endpoint file):
- `url.paths[]` — all route variants with `{placeholders}` (path parts)
- `url.parts` — path part names/types (required if present in every path variant)
- `params` — query parameters: name, type (`boolean|number|list|string|time|enum`), `required`, `default`, `options` (enum values), `deprecated`
- `body` — `serialize: single`/`bulk`, required flag
- `stability` (`stable|beta|experimental`), `deprecated` flag, `visibility` (`public|feature_flag|private`)

**Spec audit tool** (see §8.4): `internal/speccheck` parses every spec JSON (excluding `_internal.*` and `visibility: private`) and verifies: (a) a service method exists for the endpoint, (b) every `params` entry has a field in the endpoint's `*Params` struct, (c) every path part is represented. CI runs `go run ./internal/speccheck -spec /path/to/rest-api-spec/api`. The spec always wins over this plan's tables; discrepancies found during implementation must be fixed in code, not in the spec.

Common parameters (from `_common.json`) apply to **every** endpoint and are modeled once via embedded `CommonParams` (§4.3): `error_trace`, `filter_path`, `human`, `pretty`.

---

## 3. Directory structure

```
elasticsearch/
├── .github/workflows/workflow.yaml          # GitHub Actions -> dagger call (mirrors opensearch)
├── .gitignore                               # copy from opensearch
├── LICENSE                                  # Apache-2.0 (copy from opensearch)
├── README.md                                # usage, services table, CI badges
├── README_RETRY.md                          # retry behavior doc (port from opensearch)
├── CODE_OF_CONDUCT.md, CONTRIBUTING.md, ISSUE_TEMPLATE.md   # port from opensearch
├── go.mod / go.sum
├── dagger.json                              # name "elasticsearch", deps: codecov, git-module, golang
├── doc.go                                   # package elasticsearch overview docs
├── client.go                                # Config, Client iface, DefaultClient, New, retry conditions
├── client_test.go                           # unit: New, TLS, auth headers, product check
├── client_retry_test.go                     # unit: retry conditions (port pattern)
├── example_test.go, example_retry_test.go
├── common.go                                # root-level type aliases (types.* re-exports)
├── errors.go                                # IsNotFound / IsConflict / IsUnauthorized helpers
├── errors_test.go
├── generate.go                              # mockgen directives for client.go + every service iface
├── types/
│   ├── doc.go
│   ├── common.go                            # CommonParams, AcknowledgedResponse, ShardsInfo, BroadcastResponse, FailedNodeException, DocumentVersion, UnixMilliTime aliases...
│   ├── errors.go                            # ElasticsearchError, ElasticsearchErrorDetails, IsNotFound/IsConflict/IsUnauthorized
│   ├── errors_test.go, time.go, time_test.go, example_test.go
├── querydsl/                                # FULL PORT of opensearch querydsl (152 files)
│   └── ... (search_request.go, query.go, search_aggs_*.go, suggester_*.go, sort.go, highlight.go, pit.go, ...)
├── trace/opentelemetry/
│   ├── transport.go                         # OTel RoundTripper for resty (port)
│   └── transport_test.go
├── api/
│   ├── doc.go
│   ├── errors.go, errors_test.go            # parseErrorResponse, logAndReturnError, wrapNetworkError, wrapUnmarshalError
│   ├── options.go, options_test.go          # rawBody, normalizeBody, bodyProvider, validate, request structs
│   ├── param_types.go                       # Refresh, VersionType, SearchType, ExpandWildcard*, HealthStatus, Bytes, Level, SuggestMode, GroupBy, WaitForStatus...
│   ├── common_params.go                     # CommonParams embedding helpers + mergeParams
│   ├── test_helpers_test.go                 # restyClient(srv), deadClient(), testLogger()
│   ├── integration_helper_test.go           # //go:build integration; env-driven client
│   │
│   │   # one triplet per service (model + service + tests); integration tests only where a live ES can exercise it:
│   ├── info_model.go            info_service.go            info_service_test.go            info_service_integration_test.go
│   ├── document_model.go        document_service.go        document_service_test.go        document_service_integration_test.go
│   ├── search_model.go          search_service.go          search_service_test.go          search_service_integration_test.go
│   ├── async_search_model.go    async_search_service.go    async_search_service_test.go    async_search_service_integration_test.go
│   ├── script_model.go          script_service.go          script_service_test.go          script_service_integration_test.go
│   ├── indices_model.go         indices_service.go         indices_service_test.go         indices_service_integration_test.go
│   ├── cluster_model.go         cluster_service.go         cluster_service_test.go         cluster_service_integration_test.go
│   ├── nodes_model.go           nodes_service.go           nodes_service_test.go           nodes_service_integration_test.go
│   ├── cat_model.go             cat_service.go             cat_service_test.go             cat_service_integration_test.go
│   ├── ingest_model.go          ingest_service.go          ingest_service_test.go          ingest_service_integration_test.go
│   ├── snapshot_model.go        snapshot_service.go        snapshot_service_test.go        snapshot_service_integration_test.go
│   ├── slm_model.go             slm_service.go             slm_service_test.go             slm_service_integration_test.go
│   ├── tasks_model.go           tasks_service.go           tasks_service_test.go           tasks_service_integration_test.go
│   ├── ilm_model.go             ilm_service.go             ilm_service_test.go             ilm_service_integration_test.go
│   ├── transform_model.go       transform_service.go       transform_service_test.go       transform_service_integration_test.go
│   ├── ml_model.go              ml_service.go              ml_service_test.go              ml_service_integration_test.go
│   ├── security_model.go        security_service.go        security_service_test.go        security_service_integration_test.go
│   ├── sql_model.go             sql_service.go             sql_service_test.go             sql_service_integration_test.go
│   ├── watcher_model.go         watcher_service.go         watcher_service_test.go         watcher_service_integration_test.go
│   ├── ccr_model.go             ccr_service.go             ccr_service_test.go             ccr_service_integration_test.go
│   ├── autoscaling_model.go     autoscaling_service.go     autoscaling_service_test.go     autoscaling_service_integration_test.go
│   ├── enrich_model.go          enrich_service.go          enrich_service_test.go          enrich_service_integration_test.go
│   ├── eql_model.go             eql_service.go             eql_service_test.go             eql_service_integration_test.go
│   ├── graph_model.go           graph_service.go           graph_service_test.go           graph_service_integration_test.go
│   ├── license_model.go         license_service.go         license_service_test.go         license_service_integration_test.go
│   ├── logstash_model.go        logstash_service.go        logstash_service_test.go        logstash_service_integration_test.go
│   ├── migration_model.go       migration_service.go       migration_service_test.go       migration_service_integration_test.go
│   ├── shutdown_model.go        shutdown_service.go        shutdown_service_test.go        shutdown_service_integration_test.go
│   ├── features_model.go        features_service.go        features_service_test.go        features_service_integration_test.go
│   ├── fleet_model.go           fleet_service.go           fleet_service_test.go           fleet_service_integration_test.go
│   ├── dangling_indices_model.go dangling_indices_service.go dangling_indices_service_test.go
│   ├── health_report_model.go   health_report_service.go   health_report_service_test.go   health_report_service_integration_test.go
│   ├── query_rules_model.go     query_rules_service.go     query_rules_service_test.go     query_rules_service_integration_test.go
│   ├── search_application_model.go search_application_service.go search_application_service_test.go search_application_service_integration_test.go
│   ├── synonyms_model.go        synonyms_service.go        synonyms_service_test.go        synonyms_service_integration_test.go
│   ├── text_structure_model.go  text_structure_service.go  text_structure_service_test.go
│   ├── inference_model.go       inference_service.go       inference_service_test.go
│   ├── searchable_snapshots_model.go searchable_snapshots_service.go searchable_snapshots_service_test.go searchable_snapshots_service_integration_test.go
│   ├── ssl_model.go             ssl_service.go             ssl_service_test.go
│   ├── xpack_model.go           xpack_service.go           xpack_service_test.go           xpack_service_integration_test.go
│   └── streams_model.go         streams_service.go         streams_service_test.go        # only if streams.* present in pinned 9.x spec
├── internal/speccheck/
│   ├── main.go                              # spec audit CLI (parses spec JSON, checks coverage via registry)
│   └── registry.go                          # endpoint->method/params registry emitted alongside services
└── ci/dagger/
    ├── go.mod, go.sum, main.go, .gitignore, .gitattributes
```

File conventions (mirror opensearch exactly):
- `*_service.go`: `XxxService` interface + `DefaultXxxService` struct + `NewXxxService(client *resty.Client, logger *logrus.Entry) XxxService`
- `*_model.go`: request structs, `*Params` structs with `ToMap()`, response structs
- `*_service_test.go`: package `api`, `TestUnitXxx...` with `httptest` servers
- `*_service_integration_test.go`: package `api_test`, `//go:build integration`, `TestIntegrationXxx...`
- mocks: `generate.go` with one `//go:generate mockgen -source=api/<svc>_service.go -destination=mock/<svc>_service.go -package=mock` line per service + `client.go`

---

## 4. Core plumbing (exact data structures)

### 4.1 Root package `elasticsearch` — `client.go`

```go
// DefaultRetryConditions: retries network errors (connection reset/refused,
// timeout, i/o timeout, network unreachable, broken pipe, EOF), HTTP 429,
// and 5xx except 501. Port verbatim from opensearch client.go.
func DefaultRetryConditions() []resty.RetryConditionFunc

// ESSearchRetryConditions: default conditions + retry on transient ES error
// types in body: search_phase_execution_exception, too_many_buckets_exception,
// circuit_breaking_exception (rename of opensearch PITSearchRetryConditions).
func ESSearchRetryConditions() []resty.RetryConditionFunc

type Client interface {
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
    Streams() api.StreamsService   // only if streams.* in pinned spec
}

type Config struct {
    URL             string        // required, e.g. https://localhost:9200
    Username        string        // basic auth
    Password        string
    APIKey          string        // Authorization: ApiKey <value> (base64 id:api_key or raw)
    BearerToken     string        // Authorization: Bearer <value>
    TLSSkipVerify   bool
    CACert          []byte        // PEM CA pool
    Timeout         time.Duration
    IdleConnTimeout time.Duration // default 60s (as opensearch)
    DisableHTTP2    bool
    RetryCount      int
    RetryWaitTime   time.Duration // default 100ms
    RetryMaxWaitTime time.Duration // default 2s
    RetryConditions []resty.RetryConditionFunc
}

type DefaultClient struct { /* resty client + logger + one field per service */ }

func New(cfg *Config, logger *logrus.Entry) (Client, error)
```

`New` behavior (opensearch pattern + ES additions):
1. `resty.New()`, default header `Content-Type: application/json`, `SetBaseURL(cfg.URL)`.
2. Auth precedence, exactly one may be set (else return `*types.ElasticsearchError` with reason):
   - `Username`/`Password` -> `SetBasicAuth`
   - `APIKey` -> `SetHeader("Authorization", "ApiKey "+cfg.APIKey)`
   - `BearerToken` -> `SetHeader("Authorization", "Bearer "+cfg.BearerToken)`
3. TLS: same as opensearch (InsecureSkipVerify, CACert pool parse error -> typed error).
4. Transport clone, IdleConnTimeout default 60s, DisableHTTP2 handling — port verbatim.
5. `OnBeforeRequest`/`OnAfterResponse` debug logging with method/path/status/duration — port verbatim; `SetDebug(true)` when trace level.
6. Retry config — port verbatim.
7. Instantiate all services with `api.NewXxxService(c, logger)`.

Root `errors.go`: `IsNotFound(err) bool` (404), `IsConflict(err) bool` (409), `IsUnauthorized(err) bool` (401) delegating to `types`.
Root `common.go`: aliases `CommonParams`, `AcknowledgedResponse`, `ShardsInfo`, `ShardOperationFailedException`, `BroadcastResponse`, `FailedNodeException`, `ElasticsearchError`, `ElasticsearchErrorDetails`, `DocumentVersion`, `UnixMilliTime` (same pattern as opensearch `common.go`).

### 4.2 `types` package

```go
type CommonParams struct {
    Pretty     *bool
    Human      *bool
    ErrorTrace *bool
    FilterPath []string
}
func (p *CommonParams) ToMap() map[string]string   // pretty, human, error_trace, filter_path (comma-joined)

type AcknowledgedResponse struct {
    Acknowledged       bool   `json:"acknowledged"`
    ShardsAcknowledged bool   `json:"shards_acknowledged,omitempty"`
    Index              string `json:"index,omitempty"`
}

type ShardsInfo struct { Total, Successful, Failed int; Skipped int; Failures []*ShardOperationFailedException }
type ShardOperationFailedException struct { Shard int; Index, Status string; Reason map[string]any; Node string; Primary bool }
type BroadcastResponse struct { Shards *ShardsInfo; Total, Successful, Failed int; Failures []*ShardOperationFailedException }
type FailedNodeException struct { *ElasticsearchErrorDetails; NodeId string }

type ElasticsearchError struct {
    Status  int                       `json:"status"`
    Details *ElasticsearchErrorDetails `json:"error,omitempty"`
}
func (e *ElasticsearchError) Error() string  // "elasticsearch error <status>: <type>: <reason>"
// UnmarshalJSON must also accept the legacy/plain-string error body form:
// {"error": "some string", "status": 400} -> Details{Reason: "some string"}

type ElasticsearchErrorDetails struct {
    Type, Reason, ResourceType, ResourceId, Index, Phase string
    Grouped  bool
    CausedBy map[string]any
    RootCause []*ElasticsearchErrorDetails
    Suppressed []*ElasticsearchErrorDetails
    FailedShards []map[string]any
    Header map[string]any
    ScriptStack []string; Script, Lang string
    Position *ScriptErrorPosition
}
type ScriptErrorPosition struct { Offset, Start, End int }
func IsNotFound(err error) bool; func IsConflict(err error) bool; func IsUnauthorized(err error) bool

type DocumentVersion struct { SeqNo, PrimaryTerm *int64 }
type UnixMilliTime time.Time  // MarshalJSON/UnmarshalJSON as unix millis (port time.go)
```

### 4.3 `api` package plumbing

`api/errors.go` — port verbatim from opensearch (`parseErrorResponse`, `logAndReturnError`, `wrapNetworkError`, `wrapUnmarshalError`), using `types.ElasticsearchError`.

`api/options.go` — port `rawBody`, `bodyProvider`, `normalizeBody`, `isNilValue`, `validate` (validator v10 with `WithRequiredStructEnabled`), `validationError`.

`api/common_params.go` (new — improvement over opensearch, where CommonParams is defined but unused):

```go
// every endpoint Params struct embeds CommonParams; ToMap implementations call
// mergeParams so pretty/human/error_trace/filter_path are always available.
func mergeParams(m map[string]string, common *types.CommonParams) map[string]string
```

`api/param_types.go` — typed enums + constants:

```go
type Refresh string            // RefreshTrue / RefreshFalse / RefreshWaitFor
type VersionType string         // internal / external / external_gte
type SearchType string          // query_then_fetch / dfs_query_then_fetch
type OpType string              // OpTypeIndex / OpTypeCreate
type HealthStatus string        // green / yellow / red
type ExpandWildcards string     // open / closed / hidden / none / all (combinable "open,hidden")
type BytesFormat string         // b / kb / mb / gb / tb / pb
type Level string               // cluster / indices / shards
type SuggestMode string        // missing / popular / always
type GroupBy string             // nodes / parents / none
type Conflicts string           // abort / proceed
type DefaultOperator string     // AND / OR
type SortOrder string           // asc / desc
type CatFormat string           // json / text (client always uses json)
type WaitForStatus string       // green / yellow / red
type TimeType string            // generic string alias for time values ("30s", "1m")
```

### 4.4 Request/Params conventions (apply to every endpoint)

- Path parts -> fields on a `*Request` struct with `validate:"required"` where the spec marks them required (or where they appear in all path variants); `Validate() error` via validator.
- Query params -> `*XxxParams` struct; **every** spec param gets a field; `ToMap() map[string]string` emits only non-zero values (bools as `"true"`/`"false"` via pointers where `false` is meaningful, e.g. `*bool`); embeds `Common *types.CommonParams`.
- Go naming: spec `snake_case` -> `CamelCase` field, map key stays `snake_case`.
- Body: `any` (normalized by `normalizeBody`); NDJSON endpoints (`bulk`, `msearch`, `msearch_template`, `mtermvectors` body form) take `string` and set `Content-Type: application/x-ndjson`.
- Method shape: `Do(ctx context.Context, ...) (*XxxResponse, error)`; endpoints with body+parts use Request structs; simple endpoints use positional args + trailing `params *XxxParams` (nil allowed).
- HEAD endpoints return `(bool, error)` — 200 true / 404 false / else error (opensearch pattern).
- `Get`-style 404 on documents returns `&GetResult{Found:false}` (opensearch pattern); index-level 404 returns `ElasticsearchError` checked via `IsNotFound`.

---

## 5. API catalog — services, signatures, and complete parameter tables

Notation: `P:` path parts, `Q:` query params = fields of the Params struct (all optional unless marked **req**), `B:` body. `*` = embedded `CommonParams`. Every method below exists on `DefaultXxxService` with the standard error/unmarshal flow (§4.3). "Verify" = confirm exact param list against the pinned spec file during Task 0 audit.

### 5.1 InfoService — `info_service.go` (spec: `info` implicit, `ping`, `capabilities.json`)

```go
type InfoService interface {
    Info(ctx context.Context) (*InfoResponse, error)                       // GET /
    Ping(ctx context.Context) (bool, error)                                // HEAD /
    Capabilities(ctx context.Context, req *CapabilitiesRequest) (*CapabilitiesResponse, error) // GET /_capabilities
}
```
- `CapabilitiesRequest`: `Q: method, path, parameters (list), local` — verify.
- `InfoResponse`: `Name, ClusterName, ClusterUUID, Version{Number, BuildFlavor, BuildType, BuildHash, BuildDate, BuildSnapshot, LuceneVersion, MinimumWireCompatibilityVersion, MinimumIndexCompatibilityVersion}, Tagline`.

### 5.2 DocumentService — spec: `index, create, get, get_source, exists, exists_source, delete, mget, update, bulk, delete_by_query, delete_by_query_rethrottle, update_by_query, update_by_query_rethrottle, reindex, reindex_rethrottle, explain, termvectors, mtermvectors, terms_enum`

```go
type DocumentService interface {
    Index(ctx context.Context, req *IndexRequest) (*IndexResponse, error)                 // PUT /{index}/_doc/{id} | POST /{index}/_doc
    Create(ctx context.Context, req *CreateRequest) (*IndexResponse, error)               // PUT /{index}/_doc/{id}/_create | POST /{index}/{id}/_create
    Get(ctx context.Context, req *GetRequest) (*GetResult, error)                         // GET /{index}/_doc/{id}
    GetSource(ctx context.Context, req *GetSourceRequest) (json.RawMessage, error)        // GET /{index}/_source/{id}
    Exists(ctx context.Context, index, id string, params *GetParams) (bool, error)        // HEAD /{index}/_doc/{id}
    ExistsSource(ctx context.Context, index, id string, params *GetSourceParams) (bool, error) // HEAD /{index}/_source/{id}
    Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error)              // DELETE /{index}/_doc/{id}
    MultiGet(ctx context.Context, req *MgetRequest) (*MgetResponse, error)                // POST /_mget | /{index}/_mget
    Update(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error)              // POST /{index}/_update/{id}
    Bulk(ctx context.Context, req *BulkRequest) (*BulkResponse, error)                    // POST /_bulk | /{index}/_bulk (ndjson)
    DeleteByQuery(ctx context.Context, req *DeleteByQueryRequest) (*BulkIndexByScrollResponse, error)   // POST /{index}/_delete_by_query
    DeleteByQueryRethrottle(ctx context.Context, req *RethrottleRequest) (*TasksListResponse, error)    // POST /_delete_by_query/{task_id}/_rethrottle
    UpdateByQuery(ctx context.Context, req *UpdateByQueryRequest) (*BulkIndexByScrollResponse, error)   // POST /{index}/_update_by_query
    UpdateByQueryRethrottle(ctx context.Context, req *RethrottleRequest) (*TasksListResponse, error)    // POST /_update_by_query/{task_id}/_rethrottle
    Reindex(ctx context.Context, req *ReindexRequest) (*BulkIndexByScrollResponse, error) // POST /_reindex
    ReindexRethrottle(ctx context.Context, req *RethrottleRequest) (*TasksListResponse, error) // POST /_reindex/{task_id}/_rethrottle
    Explain(ctx context.Context, req *ExplainRequest) (*ExplainResponse, error)           // GET|POST /{index}/_explain/{id}
    TermVectors(ctx context.Context, req *TermVectorsRequest) (*TermvectorsResponse, error) // GET|POST /{index}/_termvectors/{id}
    MultiTermVectors(ctx context.Context, req *MtermvectorsRequest) (*MultiTermvectorResponse, error) // GET|POST /_mtermvectors | /{index}/_mtermvectors
    TermsEnum(ctx context.Context, req *TermsEnumRequest) (*TermsEnumResponse, error)     // POST /{index}/_terms_enum
}
```

Parameter tables (Q fields; `*` = CommonParams embedded):

| Endpoint | P | Params struct | Q fields (spec names) |
|---|---|---|---|
| index | index **req**, id | `IndexParams` | if_primary_term, if_seq_no, op_type(enum index/create), pipeline, refresh(enum), require_alias, timeout, version, version_type, wait_for_active_shards |
| create | index **req**, id **req** | `CreateParams` (= IndexParams minus op_type) | if_primary_term, if_seq_no, pipeline, refresh, require_alias, timeout, version, version_type, wait_for_active_shards |
| get | index **req**, id **req** | `GetParams` | force_synthetic_source, preference, realtime, refresh, routing, _source, _source_excludes, _source_includes, stored_fields, version, version_type |
| get_source | index **req**, id **req** | `GetSourceParams` | preference, realtime, refresh, routing, _source, _source_excludes, _source_includes, version, version_type |
| exists / exists_source | same as get / get_source | reuse `GetParams` / `GetSourceParams` | same |
| delete | index **req**, id **req** | `DeleteParams` | if_primary_term, if_seq_no, refresh, routing, timeout, version, version_type |
| mget | index | `MgetParams` | force_synthetic_source, preference, realtime, refresh, routing, _source, _source_excludes, _source_includes, stored_fields |
| update | index **req**, id **req** | `UpdateParams` | if_primary_term, if_seq_no, lang, refresh, require_alias, retry_on_conflict, routing, _source, _source_excludes, _source_includes, timeout, wait_for_active_shards |
| bulk | index | `BulkParams` | list_executed_pipelines, pipeline, refresh, require_alias, routing, _source, _source_excludes, _source_includes, timeout, wait_for_active_shards |
| delete_by_query | indices **req** (list) | `DeleteByQueryParams` | analyzer, analyze_wildcard, conflicts(abort/proceed), default_operator, df, expand_wildcards, from, ignore_unavailable, lenient, max_docs, preference, q, refresh, request_cache, requests_per_second, routing, scroll, scroll_size, search_timeout, search_type, slice_id, slice_max, slices, sort, stats, terminate_after, timeout, version, wait_for_active_shards, wait_for_completion |
| update_by_query | indices **req** | `UpdateByQueryParams` | analyzer, analyze_wildcard, conflicts, default_operator, df, expand_wildcards, from, ignore_unavailable, lenient, max_docs, pipeline, preference, q, refresh, request_cache, requests_per_second, routing, scroll, scroll_size, search_timeout, slices, slice_id, slice_max, sort, stats, terminate_after, timeout, version, version_type, wait_for_active_shards, wait_for_completion |
| delete_by_query_rethrottle / update_by_query_rethrottle / reindex_rethrottle | task_id **req** | `RethrottleParams` | requests_per_second |
| reindex | — | `ReindexParams` | max_docs, refresh, requests_per_second, require_alias, scroll, slices, timeout, wait_for_active_shards, wait_for_completion |
| explain | index **req**, id **req** | `ExplainParams` | analyzer, analyze_wildcard, default_operator, df, lenient, preference, q, routing, _source, _source_excludes, _source_includes, stored_fields |
| termvectors | index **req**, id | `TermVectorsParams` | field_statistics, fields, offsets, payloads, positions, preference, realtime, routing, term_statistics, version, version_type |
| mtermvectors | index | `MtermvectorsParams` | ids, field_statistics, fields, offsets, payloads, positions, preference, realtime, routing, term_statistics, version, version_type |
| terms_enum | index **req** | `TermsEnumParams` | case_insensitive |

Bodies: `IndexRequest.Body any`; `UpdateRequest.Body any` ({doc}/{script}/{upsert}); `BulkRequest.Body string` (NDJSON); `MgetRequest.Body any` ({docs|ids}); `ReindexRequest.Body any` ({source,dest,...}); `ExplainRequest.Body any` ({query}); `TermsEnumRequest.Body any` ({field,string,size,timeout}).

Response structs (all typed):
- `IndexResponse`: _index, _id, _version, result, _shards *ShardsInfo, _seq_no, _primary_term, status, forced_refresh (port from opensearch).
- `GetResult`: _index, _id, _version *, _seq_no *, _primary_term *, _source json.RawMessage, found, fields map[string]any, error *ElasticsearchErrorDetails (port).
- `DeleteResponse`: as opensearch.
- `UpdateResponse`: as opensearch (+ GetResult *GetResult `json:"get"`).
- `MgetResponse`: docs []GetResult.
- `BulkResponse`: took int64, errors bool, items []BulkResponseItem{_index,_id,status,seq_no,primary_term,version,result,error *ElasticsearchErrorDetails,forced_refresh}.
- `BulkIndexByScrollResponse`: took, timed_out, total, updated, created, deleted, batches, version_conflicts, noops, retries{bulk,search}, throttled_millis, requests_per_second, throttled_until_millis, failures []map[string]any, slice_id.
- `TasksListResponse`: nodes map[string]TaskNode{node_id, transport_address, host, ip, name, tasks map[string]TaskInfo{node, id, type, action, status map[string]any, description, start_time_in_millis, running_time_in_nanos, cancellable, cancelled, parent_task_id, headers}}.
- `ExplainResponse`: _index, _id, matched bool, explanation map[string]any.
- `TermvectorsResponse`: _index, _id, _version, found, took, term_vectors map[string]TermVector{field_statistics{sum_doc_freq,doc_count,sum_ttf}, terms map[string]TermInfo{term_freq, tokens[]Token{start_offset,end_offset,position,payload}, score}}.
- `MultiTermvectorResponse`: docs []TermvectorsResponse.
- `TermsEnumResponse`: _shards, terms []string, complete bool.

### 5.3 SearchService — spec: `search, search_mvt, scroll, clear_scroll, count, field_caps, search_shards, search_template, msearch, msearch_template, render_search_template, rank_eval, validate_query, open_point_in_time, close_point_in_time, scripts_painless_exec`

```go
type SearchService interface {
    Search(ctx context.Context, req *SearchRequest) (*querydsl.SearchResult, error)        // GET|POST /_search | /{index}/_search
    SearchMVT(ctx context.Context, req *SearchMVTRequest) (json.RawMessage, error)         // GET /{index}/_mvt/{field}/{zoom}/{x}/{y} (map binary/vector tiles)
    MultiSearch(ctx context.Context, req *MultiSearchRequest) (*querydsl.MultiSearchResult, error) // GET|POST /_msearch | /{index}/_msearch (ndjson)
    SearchTemplate(ctx context.Context, req *SearchTemplateRequest) (*querydsl.SearchResult, error)
    MultiSearchTemplate(ctx context.Context, req *MultiSearchTemplateRequest) (*querydsl.MultiSearchResult, error)
    RenderSearchTemplate(ctx context.Context, req *RenderSearchTemplateRequest) (*querydsl.RenderSearchTemplateResponse, error)
    RankEval(ctx context.Context, req *RankEvalRequest) (*querydsl.RankEvalResponse, error)
    Count(ctx context.Context, req *CountRequest) (int64, error)                            // GET|POST /_count | /{index}/_count
    Scroll(ctx context.Context, req *ScrollRequest) (*querydsl.SearchResult, error)         // GET|POST /_search/scroll
    ClearScroll(ctx context.Context, scrollIds []string) (*querydsl.ClearScrollResponse, error) // DELETE /_search/scroll
    Validate(ctx context.Context, req *ValidateQueryRequest) (*querydsl.ValidateResponse, error) // GET|POST /_validate/query | /{index}/_validate/query
    SearchShards(ctx context.Context, req *SearchShardsRequest) (*querydsl.SearchShardsResponse, error)
    FieldCaps(ctx context.Context, req *FieldCapsRequest) (*querydsl.FieldCapsResponse, error) // GET|POST /_field_caps | /{index}/_field_caps
    OpenPointInTime(ctx context.Context, req *OpenPITRequest) (*OpenPITResponse, error)     // POST /{index}/_pit
    ClosePointInTime(ctx context.Context, pitIds []string) (*ClosePITResponse, error)       // DELETE /_pit
    ScriptsPainlessExec(ctx context.Context, body any) (json.RawMessage, error)             // POST /_scripts/painless/_execute
}
```

| Endpoint | P | Params struct | Q fields |
|---|---|---|---|
| search | indices | `SearchParams` | allow_no_indices, allow_partial_search_results, analyzer, analyze_wildcard, batched_reduce_size, ccs_minimize_roundtrips, default_operator, df, docvalue_fields, expand_wildcards, explain, from, ignore_throttled, ignore_unavailable, include_named_queries_score (verify), lenient, max_concurrent_shard_requests, min_compatible_shard_node, pre_filter_shard_size, preference, q, request_cache, rest_total_hits_as_int, routing, scroll, search_type, seq_no_primary_term, size, sort, _source, _source_excludes, _source_includes, stats, stored_fields, suggest_field, suggest_mode, suggest_size, suggest_text, terminate_after, timeout, track_scores, track_total_hits, typed_keys, version |
| search_mvt | index **req**, field **req**, zoom **req**, x **req**, y **req** | `SearchMVTParams` | exact_bounds, extent, grid_agg, grid_precision, grid_type, size, track_total_hits, with_labels |
| msearch | indices | `MultiSearchParams` | ccs_minimize_roundtrips, max_concurrent_searches, max_concurrent_shard_requests, pre_filter_shard_size, rest_total_hits_as_int, search_type, typed_keys |
| search_template | indices | `SearchTemplateParams` | allow_no_indices, ccs_minimize_roundtrips, explain, ignore_throttled, ignore_unavailable, preference, profile, rest_total_hits_as_int, routing, scroll, search_type, typed_keys |
| msearch_template | indices | `MultiSearchTemplateParams` | ccs_minimize_roundtrips, max_concurrent_searches, rest_total_hits_as_int, search_type, typed_keys |
| render_search_template | id | — (body only) | — |
| rank_eval | indices | `RankEvalParams` | allow_no_indices, expand_wildcards, ignore_unavailable, search_type |
| count | indices | `CountParams` | allow_no_indices, analyzer, analyze_wildcard, default_operator, df, expand_wildcards, ignore_throttled, ignore_unavailable, lenient, min_score, preference, q, routing, terminate_after |
| scroll | — | `ScrollParams` | scroll, scroll_id, rest_total_hits_as_int |
| clear_scroll | — | body {scroll_id:[]} | — |
| validate_query | indices | `ValidateQueryParams` | all_shards, allow_no_indices, analyzer, analyze_wildcard, default_operator, df, expand_wildcards, explain, ignore_unavailable, lenient, q, rewrite |
| search_shards | indices | `SearchShardsParams` | allow_no_indices, expand_wildcards, ignore_unavailable, local, preference, routing |
| field_caps | indices | `FieldCapsParams` | allow_no_indices, expand_wildcards, fields, ignore_unavailable, include_unmapped; B: {index_filter, runtime_mappings} |
| open_point_in_time | indices **req** | `OpenPITParams` | allow_no_indices, expand_wildcards, ignore_unavailable, keep_alive **req**, preference, routing |
| close_point_in_time | — | body {id} | — |

`SearchRequest` mirrors opensearch: `{Indices []string; Body any; Params *SearchParams}` + `NewSearchRequest(*querydsl.SearchRequest)` adapter. Responses reuse `querydsl.SearchResult` etc. (ported). `OpenPITResponse{Id string json:"id"}`, `ClosePITResponse{Succeeded bool; NumFreed int json:"num_freed"}`.

### 5.4 AsyncSearchService — spec: `async_search.submit/get/status/delete`

```go
Submit(ctx, req *AsyncSearchSubmitRequest) (*AsyncSearchResponse, error)   // POST /{index}/_async_search
Get(ctx, id string, params *AsyncSearchGetParams) (*AsyncSearchResponse, error) // GET /_async_search/{id}
Status(ctx, id string) (*AsyncSearchStatusResponse, error)                 // GET /_async_search/status/{id}
Delete(ctx, id string) (*AcknowledgedResponse, error)                      // DELETE /_async_search/{id}
```
- `AsyncSearchSubmitParams`: allow_no_indices, allow_partial_search_results, analyzer, analyze_wildcard, batched_reduce_size, default_operator, df, expand_wildcards, explain, from, ignore_throttled, ignore_unavailable, keep_alive, keep_on_completion, lenient, preference, q, request_cache, routing, search_type, size, sort, _source, _source_excludes, _source_includes, stats, stored_fields, suggest_field, suggest_mode, suggest_size, suggest_text, terminate_after, timeout, track_scores, track_total_hits, typed_keys, wait_for_completion_timeout.
- `AsyncSearchGetParams`: keep_alive, typed_keys, wait_for_completion_timeout.
- `AsyncSearchResponse`: id, is_partial, is_running, start_time_in_millis, expiration_time_in_millis, response *querydsl.SearchResult.
- `AsyncSearchStatusResponse`: id, is_running, is_partial, start_time_in_millis, expiration_time_in_millis, _shards{total,successful,skipped,failed}.

### 5.5 ScriptService — spec: `get_script, put_script, delete_script, get_script_context, get_script_languages`

```go
Get(ctx, id string, params *ScriptGetParams) (*ScriptGetResponse, error)          // GET /_scripts/{id}
Put(ctx, req *ScriptPutRequest) (*AcknowledgedResponse, error)                    // PUT|POST /_scripts/{id}
Delete(ctx, id string, params *ScriptDeleteParams) (*AcknowledgedResponse, error) // DELETE /_scripts/{id}
GetContext(ctx) (*ScriptContextResponse, error)                                   // GET /_script_context
GetLanguages(ctx) (*ScriptLanguagesResponse, error)                               // GET /_script_language
```
- `ScriptGetParams`: timeout (master_timeout removed in ES9 — verify).
- `ScriptPutParams` (in `ScriptPutRequest`): context, timeout.
- `ScriptGetResponse`: _id, found, script{lang, source, options}.
- `ScriptContextResponse`: contexts []ScriptContextInfo{context, methods[]{name, params[]}}.
- `ScriptLanguagesResponse`: types_allowed[]string, language_contexts[]{language, contexts[]}.

### 5.6 IndicesService — spec: `indices.*` (~45 endpoints)

```go
type IndicesService interface {
    Create(ctx, index string, body any, params *IndicesCreateParams) (*types.AcknowledgedResponse, error)      // PUT /{index}
    Delete(ctx, indices []string, params *IndicesDeleteParams) (*types.AcknowledgedResponse, error)            // DELETE /{index}
    Get(ctx, indices []string, params *IndicesGetParams) (map[string]*IndicesGetResponse, error)               // GET /{index}
    Exists(ctx, indices []string, params *IndicesExistsParams) (bool, error)                                   // HEAD /{index}
    Open(ctx, indices []string, params *IndicesOpenParams) (*types.AcknowledgedResponse, error)                // POST /{index}/_open
    Close(ctx, indices []string, params *IndicesCloseParams) (*types.AcknowledgedResponse, error)              // POST /{index}/_close
    Rollover(ctx, req *RolloverRequest) (*IndicesRolloverResponse, error)                                      // POST /{alias}/_rollover | /{alias}/_rollover/{new_index}
    Shrink(ctx, req *ShrinkRequest) (*types.AcknowledgedResponse, error)                                       // PUT /{index}/_shrink/{target}
    Split(ctx, req *SplitRequest) (*types.AcknowledgedResponse, error)                                         // PUT /{index}/_split/{target}
    Clone(ctx, req *CloneRequest) (*types.AcknowledgedResponse, error)                                         // PUT /{index}/_clone/{target}
    Refresh(ctx, indices []string, params *IndicesRefreshParams) (*RefreshResult, error)                       // POST /_refresh | /{index}/_refresh
    Flush(ctx, indices []string, params *IndicesFlushParams) (*BroadcastResponse, error)                       // POST /_flush | /{index}/_flush
    ForceMerge(ctx, indices []string, params *IndicesForceMergeParams) (*BroadcastResponse, error)             // POST /_forcemerge | /{index}/_forcemerge
    ClearCache(ctx, indices []string, params *IndicesClearCacheParams) (*BroadcastResponse, error)             // POST /_cache/clear | /{index}/_cache/clear
    Analyze(ctx, index string, body any) (*IndicesAnalyzeResponse, error)                                      // GET|POST /_analyze | /{index}/_analyze
    Stats(ctx, req *IndicesStatsRequest) (*IndicesStatsResponse, error)                                        // GET /_stats | /_stats/{metric} | /{index}/_stats | /{index}/_stats/{metric}
    Segments(ctx, indices []string, params *IndicesSegmentsParams) (*IndicesSegmentsResponse, error)           // GET /_segments | /{index}/_segments
    Recovery(ctx, indices []string, params *IndicesRecoveryParams) (*IndicesRecoveryResponse, error)           // GET /_recovery | /{index}/_recovery
    ShardStores(ctx, indices []string, params *IndicesShardStoresParams) (*IndicesShardStoresResponse, error)  // GET /_shard_stores | /{index}/_shard_stores
    ReloadSearchAnalyzers(ctx, indices []string, params *IndicesReloadAnalyzersParams) (*IndicesReloadAnalyzersResponse, error) // GET|POST /{index}/_reload_search_analyzers
    Downsample(ctx, index, targetIndex string, body any) (*types.AcknowledgedResponse, error)                  // POST /{index}/_downsample/{target_index}
    AddBlock(ctx, req *AddBlockRequest) (*IndicesBlockResponse, error)                                         // PUT /{index}/_block/{block}
    // aliases
    PutAlias(ctx, req *PutAliasRequest) (*types.AcknowledgedResponse, error)                                   // PUT|POST /{index}/_alias/{name}
    GetAlias(ctx, req *GetAliasRequest) (map[string]*IndicesGetAliasResponse, error)                           // GET /_alias | /{index}/_alias | /_alias/{name} | /{index}/_alias/{name}
    ExistsAlias(ctx, req *ExistsAliasRequest) (bool, error)                                                    // HEAD /_alias/{name} | /{index}/_alias/{name}
    DeleteAlias(ctx, req *DeleteAliasRequest) (*types.AcknowledgedResponse, error)                             // DELETE /{index}/_alias/{name}
    UpdateAliases(ctx, body any, params *IndicesUpdateAliasesParams) (*types.AcknowledgedResponse, error)      // POST /_aliases
    // settings / mappings
    GetSettings(ctx, req *GetSettingsRequest) (map[string]*IndicesGetSettingsResponse, error)                  // GET /_settings | /{index}/_settings | /{index}/_settings/{name}
    PutSettings(ctx, indices []string, body any, params *IndicesPutSettingsParams) (*types.AcknowledgedResponse, error) // PUT /_settings | /{index}/_settings
    GetMapping(ctx, indices []string, params *IndicesGetMappingParams) (map[string]*IndicesGetMappingResponse, error)   // GET /_mapping | /{index}/_mapping
    PutMapping(ctx, req *PutMappingRequest) (*types.AcknowledgedResponse, error)                               // PUT /{index}/_mapping
    GetFieldMapping(ctx, req *GetFieldMappingRequest) (map[string]any, error)                                  // GET /_mapping/field/{fields} | /{index}/_mapping/field/{fields}
    // legacy templates (Deprecated in ES9)
    GetTemplate(ctx, names []string, params *IndicesGetTemplateParams) (map[string]*IndicesGetTemplateResponse, error)  // Deprecated
    PutTemplate(ctx, req *PutTemplateRequest) (*types.AcknowledgedResponse, error)                             // Deprecated
    ExistsTemplate(ctx, name string) (bool, error)                                                             // Deprecated
    DeleteTemplate(ctx, name string, params *IndicesDeleteTemplateParams) (*types.AcknowledgedResponse, error) // Deprecated
    // composable templates
    GetIndexTemplate(ctx, names []string, params *IndicesGetIndexTemplateParams) (*IndicesGetIndexTemplateResponse, error)
    PutIndexTemplate(ctx, req *PutIndexTemplateRequest) (*types.AcknowledgedResponse, error)
    ExistsIndexTemplate(ctx, name string) (bool, error)
    DeleteIndexTemplate(ctx, name string, params *IndicesDeleteIndexTemplateParams) (*types.AcknowledgedResponse, error)
    SimulateIndexTemplate(ctx, req *SimulateIndexTemplateRequest) (*IndicesSimulateTemplateResponse, error)
    SimulateTemplate(ctx, req *SimulateTemplateRequest) (*IndicesSimulateTemplateResponse, error)
    // data streams
    CreateDataStream(ctx, name string, body any, params *IndicesCreateDataStreamParams) (*types.AcknowledgedResponse, error)
    GetDataStream(ctx, names []string, params *IndicesGetDataStreamParams) (*IndicesDataStreamGetResponse, error)
    DeleteDataStream(ctx, names []string, params *IndicesDeleteDataStreamParams) (*types.AcknowledgedResponse, error)
    DataStreamStats(ctx, names []string) (*IndicesDataStreamsStatsResponse, error)
    MigrateToDataStream(ctx, name string) (*types.AcknowledgedResponse, error)                                 // POST /_data_stream/_migrate/{name}
    ModifyDataStream(ctx, req *ModifyDataStreamRequest) (*types.AcknowledgedResponse, error)                   // POST /_data_stream/_modify
    PromoteDataStream(ctx, name string) (*types.AcknowledgedResponse, error)                                   // POST /_data_stream/_promote/{name} (verify presence in pinned spec)
    // resolve
    ResolveIndex(ctx, name string, params *IndicesResolveIndexParams) (*IndicesResolveIndexResponse, error)    // GET /_resolve/index/{name}
    ResolveCluster(ctx, names []string, params *IndicesResolveClusterParams) (map[string]*ResolveClusterInfo, error) // GET /_resolve/cluster/{name}
}
```

Params tables:

| Endpoint | Params struct | Q fields |
|---|---|---|
| create | `IndicesCreateParams` | wait_for_active_shards (master_timeout removed in ES9 — verify) |
| delete | `IndicesDeleteParams` | allow_no_indices, expand_wildcards, ignore_unavailable |
| get | `IndicesGetParams` | allow_no_indices, expand_wildcards, features, flat_settings, ignore_unavailable, include_defaults, local |
| exists | `IndicesExistsParams` | allow_no_indices, expand_wildcards, flat_settings, ignore_unavailable, include_defaults, local |
| open | `IndicesOpenParams` | allow_no_indices, expand_wildcards, ignore_unavailable, timeout, wait_for_active_shards |
| close | `IndicesCloseParams` | allow_no_indices, expand_wildcards, ignore_unavailable, timeout, wait_for_active_shards |
| rollover | `RolloverParams` | dry_run, timeout, wait_for_active_shards |
| shrink/split/clone | `ShrinkParams`/`SplitParams`/`CloneParams` | timeout, wait_for_active_shards |
| refresh | `IndicesRefreshParams` | allow_no_indices, expand_wildcards, ignore_unavailable |
| flush | `IndicesFlushParams` | allow_no_indices, expand_wildcards, force, ignore_unavailable, wait_if_ongoing |
| forcemerge | `IndicesForceMergeParams` | allow_no_indices, expand_wildcards, flush, ignore_unavailable, max_num_segments, only_expunge_deletes |
| clear_cache | `IndicesClearCacheParams` | allow_no_indices, expand_wildcards, fielddata, fields, ignore_unavailable, query |
| stats | `IndicesStatsParams` | completion_fields, expand_wildcards, fielddata_fields, fields, forbid_closed_indices, groups, include_segment_file_sizes, include_unloaded_segments, level |
| segments | `IndicesSegmentsParams` | allow_no_indices, expand_wildcards, ignore_unavailable, verbose |
| recovery | `IndicesRecoveryParams` | active_only, detailed |
| shard_stores | `IndicesShardStoresParams` | allow_no_indices, expand_wildcards, ignore_unavailable, status |
| reload_search_analyzers | `IndicesReloadAnalyzersParams` | allow_no_indices, expand_wildcards, ignore_unavailable |
| add_block | `AddBlockParams` | allow_no_indices, expand_wildcards, ignore_unavailable, timeout |
| put_alias / delete_alias | `AliasParams` | timeout |
| get_alias / exists_alias | `GetAliasParams` | allow_no_indices, expand_wildcards, ignore_unavailable, local |
| update_aliases | `IndicesUpdateAliasesParams` | timeout |
| get_settings | `GetSettingsParams` | allow_no_indices, expand_wildcards, flat_settings, ignore_unavailable, include_defaults, local |
| put_settings | `IndicesPutSettingsParams` | allow_no_indices, expand_wildcards, flat_settings, ignore_unavailable, preserve_existing, timeout |
| get_mapping | `IndicesGetMappingParams` | allow_no_indices, expand_wildcards, ignore_unavailable, local |
| put_mapping | `PutMappingParams` | allow_no_indices, expand_wildcards, ignore_unavailable, timeout, write_index_only |
| get_field_mapping | `GetFieldMappingParams` | allow_no_indices, expand_wildcards, ignore_unavailable, include_defaults, local |
| get_template (dep.) | `IndicesGetTemplateParams` | flat_settings, local |
| put_template (dep.) | `PutTemplateParams` | create, flat_settings, order, timeout |
| delete_template (dep.) | `IndicesDeleteTemplateParams` | timeout |
| get_index_template | `IndicesGetIndexTemplateParams` | flat_settings, local |
| put_index_template | `PutIndexTemplateParams` | cause, create |
| delete_index_template | `IndicesDeleteIndexTemplateParams` | timeout |
| simulate_index_template | `SimulateIndexTemplateParams` | cause, create |
| simulate_template | `SimulateTemplateParams` | cause, create |
| create_data_stream | `IndicesCreateDataStreamParams` | timeout |
| get_data_stream | `IndicesGetDataStreamParams` | expand_wildcards, include_defaults |
| delete_data_stream | `IndicesDeleteDataStreamParams` | expand_wildcards |
| migrate/promote data stream | — | timeout |
| modify_data_stream | — | timeout |
| resolve_index | `IndicesResolveIndexParams` | allow_no_indices, expand_wildcards, ignore_unavailable |
| resolve_cluster | `IndicesResolveClusterParams` | allow_no_indices, expand_wildcards, ignore_unavailable |

Response structs (all typed; key fields):
- `IndicesGetResponse`: Aliases map[string]AliasInfo{Filter map[string]any, IndexRouting, SearchRouting, IsWriteIndex *bool, IsHidden *bool}; Mappings map[string]any; Settings map[string]any; Defaults map[string]any (when include_defaults); DataStream string; System bool.
- `IndicesRolloverResponse`: OldIndex, NewIndex, RolledOver bool, DryRun bool, Acknowledged, ShardsAcknowledged, Conditions map[string]bool, Shards *ShardsInfo.
- `RefreshResult`: Shards *ShardsInfo.
- `IndicesAnalyzeResponse`: Tokens []AnalyzeToken{Token, StartOffset, EndOffset, Position, Type, PositionLength, Bytes, Keyword, GraphDetails}.
- `IndicesStatsResponse`: Shards *ShardsInfo, All *IndexStats, Indices map[string]*IndexStats, DataStreams map[string]any; `IndexStats{Primaries *IndexStatsDetail, Total *IndexStatsDetail}`; `IndexStatsDetail` with Docs{Count,Deleted}, Store{SizeInBytes}, Indexing{IndexTotal,IndexTimeInMillis,...}, Search{QueryTotal,QueryTimeInMillis,FetchTotal,...}, Get, Merges, Refresh, Flush, Warmer, QueryCache, Fielddata, Completion, Segments, Translog, RequestCache, Recovery (all sub-structs typed).
- `IndicesSegmentsResponse`: Shards *ShardsInfo, Indices map[string]{Shards map[string][]IndexSegmentShard{Routing{State,Primary,Node}, NumCommittedSegments, NumSearchSegments, Segments map[string]Segment{Generation, NumDocs, DeletedDocs, SizeInBytes, Committed, Search, Version, Compound, Attributes map[string]string}}}`.
- `IndicesRecoveryResponse`: map[string]IndexRecovery{Shards []ShardRecovery{Type, Stage, Primary, StartTimeInMillis, StopTimeInMillis, TotalTimeInMillis, Source{Host,TransportAddress,Name}, Target{...}, Index{Files{Total,Recovered,...}, Size{...}, Translog{...}}}}.
- `IndicesShardStoresResponse`: Indices map[string]{Shards map[string]IndexShardStores{Stores []ShardStore{ID, Path, Allocation, StoreException{Type,Reason}}}}.
- `IndicesReloadAnalyzersResponse`: Shards *ShardsInfo, ReloadDetails []ReloadDetails{Index, Node, ReloadedAnalyzers []string}.
- `IndicesBlockResponse`: Shards *ShardsInfo, Acknowledged, Indices []map[string]any.
- `IndicesGetAliasResponse`: Aliases map[string]AliasInfo.
- `IndicesGetSettingsResponse`: Settings map[string]any, Defaults map[string]any.
- `IndicesGetMappingResponse`: Mappings map[string]any.
- `IndicesGetTemplateResponse` (legacy): Order int, Version int, IndexPatterns []string, Settings/Mappings/Aliases map[string]any, Deprecated bool.
- `IndicesGetIndexTemplateResponse`: IndexTemplates []IndexTemplateItem{Name, IndexTemplate{IndexPatterns []string, Template{Settings, Mappings, Aliases}, Priority, Version, ComposedOf []string, DataStream map[string]any, AllowAutoCreate *bool}}.
- `IndicesSimulateTemplateResponse`: Template{Settings map[string]any, Mappings map[string]any, Aliases map[string]any}, Overlapping []{Name, IndexPatterns []string}.
- `IndicesDataStreamGetResponse`: DataStreams []DataStream{Name, TimestampField{Field}, Indices []DataStreamIndex{Index, IndexUUID, IsHidden, IsBackingIndex}, Generation, Status, Replicated bool, Hidden, System, AllowCustomRouting, Template}.
- `IndicesDataStreamsStatsResponse`: Shards *ShardsInfo, DataStreamCount int, BackingIndices int, TotalStoreSizeBytes int64, DataStreams []{DataStream, BackingIndices, StoreSizeBytes, MaximumRetention}.
- `IndicesResolveIndexResponse`: Indices []{Name, Attributes []string, Aliases []string}, Aliases []{Name, Indices []string}, DataStreams []{Name, BackingIndices []string, Aliases []string}.
- `ResolveClusterInfo`: Connected bool, Mode string, Seeds []string, SkipUnavailable bool, ClusterCredentials string.

### 5.7 ClusterService — spec: `cluster.*` (+ component templates)

```go
Health(ctx, indices []string, params *ClusterHealthParams) (*ClusterHealthResponse, error)          // GET /_cluster/health | /_cluster/health/{index}
State(ctx, req *ClusterStateRequest) (*ClusterStateResponse, error)                                  // GET /_cluster/state | /{metric} | /{metric}/{index}
Stats(ctx, nodeIds []string, params *ClusterStatsParams) (*ClusterStatsResponse, error)              // GET /_cluster/stats | /_cluster/stats/nodes/{node_id}
GetSettings(ctx, params *ClusterGetSettingsParams) (*ClusterGetSettingsResponse, error)              // GET /_cluster/settings
PutSettings(ctx, body any, params *ClusterPutSettingsParams) (*ClusterPutSettingsResponse, error)    // PUT /_cluster/settings
AllocationExplain(ctx, body any, params *ClusterAllocationExplainParams) (*ClusterAllocationExplainResponse, error) // GET|POST /_cluster/allocation/explain
Reroute(ctx, body any, params *ClusterRerouteParams) (*ClusterRerouteResponse, error)                // POST /_cluster/reroute
PendingTasks(ctx, params *ClusterPendingTasksParams) (*ClusterPendingTasksResponse, error)           // GET /_cluster/pending_tasks
RemoteInfo(ctx) (*ClusterRemoteInfoResponse, error)                                                  // GET /_remote/info
Info(ctx, features []string) (*ClusterInfoResponse, error)                                           // GET /_cluster/info/{feature}
PostVotingConfigExclusions(ctx, params *ClusterVotingConfigExclusionsParams) error                   // POST /_cluster/voting_config_exclusions
DeleteVotingConfigExclusions(ctx, params *ClusterDeleteVotingConfigExclusionsParams) error           // DELETE /_cluster/voting_config_exclusions
GetComponentTemplate(ctx, names []string, params *ClusterGetComponentTemplateParams) (*ClusterGetComponentTemplateResponse, error)
PutComponentTemplate(ctx, req *PutComponentTemplateRequest) (*types.AcknowledgedResponse, error)
ExistsComponentTemplate(ctx, name string) (bool, error)
DeleteComponentTemplate(ctx, name string, params *ClusterDeleteComponentTemplateParams) (*types.AcknowledgedResponse, error)
```

| Endpoint | Q fields |
|---|---|
| health | allow_no_indices? (no), expand_wildcards? (no) — actual: level, local, timeout, wait_for_active_shards, wait_for_events, wait_for_no_initializing_shards, wait_for_no_relocating_shards, wait_for_nodes, wait_for_status |
| state | allow_no_indices, expand_wildcards, flat_settings, ignore_unavailable, local, wait_for_metadata_version, wait_for_timeout |
| stats | flat_settings, timeout |
| get_settings | flat_settings, include_defaults, timeout |
| put_settings | flat_settings, timeout |
| allocation_explain | include_disk_info, include_yes_decisions |
| reroute | dry_run, explain, metric, retry_failed, timeout |
| pending_tasks | local |
| voting config exclusions | node_ids, node_names, timeout |
| delete voting config exclusions | wait_for_removal |
| component templates get | flat_settings, local; put: cause, create, timeout; delete: timeout |

Responses (typed): `ClusterHealthResponse{ClusterName, Status, TimedOut bool, NumberOfNodes, NumberOfDataNodes, ActivePrimaryShards, ActiveShards, RelocatingShards, InitializingShards, UnassignedShards, DelayedUnassignedShards, NumberOfPendingTasks, NumberOfInFlightFetch, TaskMaxWaitingInQueueMillis, ActiveShardsPercentAsNumber float64, Indices map[string]ClusterIndexHealth{Status, NumberOfShards, NumberOfReplicas, ActivePrimaryShards, ActiveShards, RelocatingShards, InitializingShards, UnassignedShards}}`; `ClusterStateResponse{ClusterName, ClusterUUID, Version int64, StateUUID, MasterNode, Metadata map[string]any, Nodes map[string]any, RoutingTable map[string]any, RoutingNodes map[string]any, Blocks map[string]any}`; `ClusterStatsResponse{ClusterName, ClusterUUID, Timestamp, Status, Indices ClusterStatsIndices{Count, Shards{Total, Replication, Primaries, Index{...}}, Docs, Store, Fielddata, QueryCache, Completion, Fields, Segments, Mappings, Analysis, Versions, Mappings}, Nodes ClusterStatsNodes{Count{Total, CoordinatingOnly, Data, Ingest, Master, RemoteClusterClient, VotingOnly}, Versions []string, OS, Process, JVM, FS, Plugins, NetworkTypes, DiscoveryTypes, Ingest, IndexingPressure, PackagingTypes, Memory}}`; `ClusterGetSettingsResponse{Persistent, Transient, Defaults map[string]any}`; `ClusterAllocationExplainResponse{Index, Shard, Primary bool, CurrentState, UnassignedInfo{...}, CanAllocate, AllocateExplanation, ConfiguredDelay, ConfiguredDelayInMillis, CurrentNode, CanMoveToNode, CanRemainDecisions[]{Decider, Decision, Explanation}, CanRebalanceCluster, CanRebalanceNode, NodeAllocationDecisions[]{NodeDecision{...}}, RemainingDelay, AllocationDelay, AllocationDelayInMillis}`; `ClusterRerouteResponse{Acknowledged, State map[string]any, Explanations []map[string]any}`; `ClusterPendingTasksResponse{Tasks []{InsertOrder, Priority, Source, TimeInQueueMillis, Executing}}`; `ClusterRemoteInfoResponse` = map[string]{Connected, Mode, Seeds, SkipUnavailable, ClusterCredentials}; `ClusterGetComponentTemplateResponse{ComponentTemplates []{Name, ComponentTemplate{Template{Settings,Mappings,Aliases}, Version, Metadata map[string]any}}}`.

### 5.8 NodesService — spec: `nodes.info, nodes.stats, nodes.usage, nodes.hot_threads, nodes.reload_secure_settings`

```go
Info(ctx, req *NodesInfoRequest) (*NodesInfoResponse, error)            // GET /_nodes | /_nodes/{node_id} | /_nodes/{metric} | /_nodes/{node_id}/{metric}
Stats(ctx, req *NodesStatsRequest) (*NodesStatsResponse, error)         // GET /_nodes/stats (+ node_id/metric/index_metric variants)
Usage(ctx, req *NodesUsageRequest) (*NodesUsageResponse, error)         // GET /_nodes/usage variants
HotThreads(ctx, nodeIds []string, params *NodesHotThreadsParams) (string, error) // GET /_nodes/hot_threads | /_nodes/{node_id}/hot_threads (text/plain)
ReloadSecureSettings(ctx, nodeIds []string, body any, params *NodesReloadSecureSettingsParams) (*NodesReloadSecureSettingsResponse, error)
```
- `NodesInfoParams`: flat_settings, timeout. `NodesStatsParams`: completion_fields, fielddata_fields, fields, groups, include_segment_file_sizes, level, timeout, types. `NodesUsageParams`: timeout. `NodesHotThreadsParams`: ignore_idle_threads, interval, snapshots, sort, threads, timeout, type. `NodesReloadSecureSettingsParams`: timeout.
- Responses: `NodesInfoResponse{ClusterName, Nodes map[string]NodeInfo{Name, TransportAddress, Host, IP, Version, BuildHash, BuildType, BuildDate, Roles []string, Attributes map[string]string, Settings map[string]any, OS map[string]any, Process map[string]any, JVM map[string]any, ThreadPool map[string]any, Transport map[string]any, HTTP map[string]any, Plugins []PluginInfo, Modules []PluginInfo, Ingest map[string]any, Aggregations map[string]any}}`; `NodesStatsResponse{ClusterName, Nodes map[string]NodeStats{Timestamp, Name, TransportAddress, Host, IP, Roles, Attributes, Indices NodeIndicesStats{Docs, Store, Indexing, Get, Search, Merges, Refresh, Flush, Warmer, QueryCache, Fielddata, Completion, Segments, Translog, RequestCache, Recovery}, OS, Process, JVM, ThreadPool, FS, Transport, HTTP, Breakers, Script, Discovery, Ingest, AdaptiveSelection, IndexingPressure, ScriptCache}}`; `NodesUsageResponse{ClusterName, Nodes map[string]{Timestamp, Since, RestActions map[string]int}}`; `NodesReloadSecureSettingsResponse{ClusterName, Nodes map[string]any, Failures []FailedNodeException}`.

### 5.9 CatService — spec: `cat.*` (all force `format=json`; typed row structs; every method takes a `*CatXxxParams` embedding CommonParams + the cat-common fields `h []string, help *bool, s []string, v *bool, local *bool` where the spec allows)

```go
type CatService interface {
    Aliases(ctx, names []string, p *CatAliasesParams) (CatAliasesResponse, error)
    Allocation(ctx, nodeIds []string, p *CatAllocationParams) (CatAllocationResponse, error)
    CircuitBreaker(ctx, p *CatCircuitBreakerParams) (CatCircuitBreakerResponse, error)      // verify presence in pinned spec
    ComponentTemplates(ctx, name string, p *CatComponentTemplatesParams) (CatComponentTemplatesResponse, error)
    Count(ctx, indices []string, p *CatCountParams) (CatCountResponse, error)
    Fielddata(ctx, fields []string, p *CatFielddataParams) (CatFielddataResponse, error)
    Health(ctx, p *CatHealthParams) (CatHealthResponse, error)
    Help(ctx) (string, error)                                                                 // plain text
    Indices(ctx, indices []string, p *CatIndicesParams) (CatIndicesResponse, error)
    Master(ctx, p *CatMasterParams) (CatMasterResponse, error)                                // Deprecated: use cluster-level info
    MLDataFrameAnalytics(ctx, id string, p *CatMLDataFrameAnalyticsParams) (CatMLDataFrameAnalyticsResponse, error)
    MLDatafeeds(ctx, datafeedId string, p *CatMLDatafeedsParams) (CatMLDatafeedsResponse, error)
    MLJobs(ctx, jobId string, p *CatMLJobsParams) (CatMLJobsResponse, error)
    MLTrainedModels(ctx, modelId string, p *CatMLTrainedModelsParams) (CatMLTrainedModelsResponse, error)
    NodeAttrs(ctx, p *CatNodeAttrsParams) (CatNodeAttrsResponse, error)
    Nodes(ctx, p *CatNodesParams) (CatNodesResponse, error)
    PendingTasks(ctx, p *CatPendingTasksParams) (CatPendingTasksResponse, error)
    Plugins(ctx, p *CatPluginsParams) (CatPluginsResponse, error)
    Recovery(ctx, indices []string, p *CatRecoveryParams) (CatRecoveryResponse, error)
    Repositories(ctx, p *CatRepositoriesParams) (CatRepositoriesResponse, error)
    Segments(ctx, indices []string, p *CatSegmentsParams) (CatSegmentsResponse, error)
    Shards(ctx, indices []string, p *CatShardsParams) (CatShardsResponse, error)
    Snapshots(ctx, repository string, p *CatSnapshotsParams) (CatSnapshotsResponse, error)
    Tasks(ctx, p *CatTasksParams) (CatTasksResponse, error)
    Templates(ctx, name string, p *CatTemplatesParams) (CatTemplatesResponse, error)
    ThreadPool(ctx, patterns []string, p *CatThreadPoolParams) (CatThreadPoolResponse, error)
    Transforms(ctx, transformId string, p *CatTransformsParams) (CatTransformsResponse, error)
}
```
Per-endpoint extra Q (beyond cat-common h/help/s/v/format/local):
- aliases: expand_wildcards. allocation: bytes. circuit_breaker: bytes. component_templates: none extra. fielddata: bytes. health: ts. indices: bytes, expand_wildcards, health, include_unloaded_segments, pri *bool. ml_data_frame_analytics: allow_no_match, bytes, from, size, time. ml_datafeeds: allow_no_match, from, size, time. ml_jobs: allow_no_match, bytes, from, size. ml_trained_models: allow_no_match, bytes, from, size. nodes: bytes, full_id, time. pending_tasks: none. recovery: active_only, bytes, detailed. repositories: none. segments: bytes, verbose. shards: bytes. snapshots: ignore_unavailable, time. tasks: actions, detailed, node_id, parent_task_id. templates: none. thread_pool: size, time. transforms: allow_no_match, from, size, time.
- Row structs: port opensearch `cat_model.go` shapes and add ES-specific rows: `CatCircuitBreakerRow{Node, Field, Breaker, EstimatedSize, LimitSize, Overhead, Tripped, Limit}`, `CatComponentTemplatesRow{Name, Version}`, `CatMLJobsRow{ID, State, OpenedTime, ...}` etc. (fields = documented cat columns; json tags snake_case).

### 5.10 IngestService — spec: `ingest.*`

```go
GetPipeline(ctx, ids []string, params *IngestGetPipelineParams) (map[string]*IngestPipeline, error) // GET /_ingest/pipeline | /_ingest/pipeline/{id}
PutPipeline(ctx, req *IngestPutPipelineRequest) (*AcknowledgedResponse, error)   // PUT /_ingest/pipeline/{id}
DeletePipeline(ctx, id string, params *IngestDeletePipelineParams) (*AcknowledgedResponse, error)
Simulate(ctx, req *IngestSimulatePipelineRequest) (*IngestSimulateResponse, error) // GET|POST /_ingest/pipeline/_simulate | /_ingest/pipeline/{id}/_simulate
ProcessorGrok(ctx) (*IngestGrokResponse, error)                                    // GET /_ingest/processor/grok
```
- Params: get: summary; put: if_version, timeout; delete: timeout; simulate: verbose.
- `IngestPipeline{Description string, Version int, Processors []map[string]any, OnFailure []map[string]any, Deprecated bool}`; `IngestSimulateResponse{Docs []{Doc map[string]any, ProcessorResults []{Processor, Status, Doc, Tag, Error map[string]any}}}`; `IngestGrokResponse{Patterns map[string]string}`.

### 5.11 SnapshotService — spec: `snapshot.*`

```go
CreateRepository(ctx, name string, body any, params *SnapshotCreateRepositoryParams) (*AcknowledgedResponse, error) // PUT /_snapshot/{repository}
GetRepository(ctx, names []string, params *SnapshotGetRepositoryParams) (map[string]*SnapshotRepository, error)     // GET /_snapshot | /_snapshot/{repository}
DeleteRepository(ctx, names []string, params *SnapshotDeleteRepositoryParams) (*AcknowledgedResponse, error)
VerifyRepository(ctx, name string, params *SnapshotVerifyRepositoryParams) (*SnapshotVerifyRepositoryResponse, error) // POST /_snapshot/{repository}/_verify
RepositoryAnalyze(ctx, name string, params *SnapshotRepositoryAnalyzeParams) (*SnapshotRepositoryAnalyzeResponse, error) // POST /_snapshot/{repository}/_analyze
RepositoryVerifyIntegrity(ctx, name string, params *SnapshotRepositoryVerifyIntegrityParams) (*SnapshotRepositoryVerifyIntegrityResponse, error) // GET /_snapshot/{repository}/_verify_integrity (verify presence)
Create(ctx, req *SnapshotCreateRequest) (*SnapshotCreateResponse, error)   // PUT|POST /_snapshot/{repository}/{snapshot}
Get(ctx, req *SnapshotGetRequest) (*SnapshotGetResponse, error)            // GET /_snapshot/{repository}/{snapshot}
Delete(ctx, req *SnapshotDeleteRequest) (*AcknowledgedResponse, error)     // DELETE /_snapshot/{repository}/{snapshot}
Restore(ctx, req *SnapshotRestoreRequest) (*SnapshotRestoreResponse, error) // POST /_snapshot/{repository}/{snapshot}/_restore
Status(ctx, req *SnapshotStatusRequest) (*SnapshotStatusResponse, error)   // GET /_snapshot/status variants
Clone(ctx, req *SnapshotCloneRequest) (*AcknowledgedResponse, error)       // PUT /_snapshot/{repository}/{snapshot}/_clone/{target_snapshot}
CleanupRepository(ctx, name string, params *SnapshotCleanupRepositoryParams) (*SnapshotCleanupRepositoryResponse, error) // POST /_snapshot/{repository}/_cleanup
```
- Params: create_repository: timeout, verify; get_repository: local; delete_repository: timeout; verify_repository: timeout; repository_analyze: blob_count, concurrency, detailed, max_blob_size, max_total_data_size, seed, timeout (verify rare_action_probability); create: wait_for_completion; get: ignore_unavailable, include_repository, index_details; restore: wait_for_completion; status: ignore_unavailable; clone: none; cleanup: timeout.
- Responses: `SnapshotRepository{Type string, Settings map[string]string, UUID string}`; `SnapshotCreateResponse{Snapshot SnapshotInfo{Snapshot, UUID, Repository, Indices []string, DataStreams []string, State, StartTime, StartTimeInMillis, EndTime, EndTimeInMillis, DurationInMillis, Failures []map[string]any, Shards *ShardsInfo, VersionId, Version, IncludeGlobalState *bool}}`; `SnapshotGetResponse{Snapshots []SnapshotInfo, Total int}`; `SnapshotRestoreResponse{Accepted bool, Snapshot RestoreInfo{Snapshot, Indices []string, Shards *ShardsInfo}}`; `SnapshotStatusResponse{Snapshots []SnapshotStatus{Snapshot, Repository, UUID, State, IncludeGlobalState, Shards *ShardsInfo, Indices map[string]SnapshotIndexStatus{Shards *ShardsInfo, Stats SnapshotStats{Incremental{FileCount,SizeInBytes}, Total{...}, StartTimeInMillis, TimeInMillis}}}}}`; `SnapshotVerifyRepositoryResponse{Nodes map[string]{Name}}`; `SnapshotRepositoryAnalyzeResponse{BlobCount, Concurrency, ReadNodeCount, CoordinatingNode, Summary map[string]any, Details []map[string]any}`; `SnapshotCleanupRepositoryResponse{DeletedBytes int64 json:"deleted_bytes", DeletedBlobs int64 json:"deleted_blobs"}`.

### 5.12 SlmService — spec: `slm.*`

```go
GetLifecycle(ctx, policyIds []string) (map[string]*SlmPolicy, error)     // GET /_slm/policy | /_slm/policy/{policy_id}
PutLifecycle(ctx, policyId string, body any) (*AcknowledgedResponse, error)
DeleteLifecycle(ctx, policyId string) (*AcknowledgedResponse, error)
ExecuteLifecycle(ctx, policyId string) (*SlmExecuteResponse, error)      // PUT /_slm/policy/{policy_id}/_execute
ExecuteRetention(ctx) (*AcknowledgedResponse, error)                     // POST /_slm/_execute_retention
GetStats(ctx) (*SlmStatsResponse, error)                                 // GET /_slm/stats
GetStatus(ctx) (*SlmStatusResponse, error)                               // GET /_slm/status
Start(ctx) (*AcknowledgedResponse, error)                                // POST /_slm/start
Stop(ctx) (*AcknowledgedResponse, error)                                 // POST /_slm/stop
```
- `SlmPolicy{Version, ModifiedDate, ModifiedDateMillis, Policy map[string]any, Config map[string]any, Name, Uuid, SnapshotName, Retention, NextExecutionMillis}`; `SlmExecuteResponse{SnapshotName string}`; `SlmStatsResponse{RetentionRuns, RetentionFailed, RetentionTimedOut, RetentionDeletionTime, RetentionDeletionTimeMillis, TotalSnapshotsTaken, TotalSnapshotsFailed, TotalSnapshotDeletions, TotalSnapshotDeletionFailures, PolicyStats []map[string]any}`; `SlmStatusResponse{OperationMode string}`.

### 5.13 TasksService — spec: `tasks.list, tasks.get, tasks.cancel`

```go
List(ctx, params *TasksListParams) (*TasksListResponse, error)      // GET /_tasks
Get(ctx, taskId string, params *TasksGetParams) (*TaskGetResponse, error) // GET /_tasks/{task_id}
Cancel(ctx, taskId string, params *TasksCancelParams) (*TasksListResponse, error) // POST /_tasks/_cancel | /_tasks/{task_id}/_cancel
```
- `TasksListParams`: actions, detailed, group_by, nodes, parent_task_id, timeout, wait_for_completion. `TasksGetParams`: timeout, wait_for_completion. `TasksCancelParams`: actions, nodes, parent_task_id, wait_for_completion.
- `TaskGetResponse{Completed bool, Task TaskInfo, Error *ElasticsearchErrorDetails, Response map[string]any}`.

### 5.14 IlmService — spec: `ilm.*`

```go
GetLifecycle(ctx, policies []string) (map[string]*IlmPolicy, error)
PutLifecycle(ctx, policy string, body any) (*AcknowledgedResponse, error)
DeleteLifecycle(ctx, policy string) (*AcknowledgedResponse, error)
GetStatus(ctx) (*IlmStatusResponse, error)                            // GET /_ilm/status
Start(ctx) (*AcknowledgedResponse, error)                             // POST /_ilm/start
Stop(ctx) (*AcknowledgedResponse, error)                              // POST /_ilm/stop
ExplainLifecycle(ctx, indices []string, params *IlmExplainParams) (*IlmExplainResponse, error) // GET /{index}/_ilm/explain
MoveToStep(ctx, index string, body any) (*AcknowledgedResponse, error) // POST /_ilm/move/{index}
RemovePolicy(ctx, indices []string) (*IlmRemovePolicyResponse, error)  // POST /{index}/_ilm/remove
Retry(ctx, indices []string) (*AcknowledgedResponse, error)            // POST /{index}/_ilm/retry
MigrateToDataTiers(ctx, dryRun *bool) (*IlmMigrateToDataTiersResponse, error) // Deprecated: POST /_ilm/migrate_to_data_tiers
```
- `IlmExplainParams`: only_errors, only_managed.
- `IlmPolicy{Version, ModifiedDate, ModifiedDateMillis, Policy map[string]any, InUse bool}`; `IlmStatusResponse{OperationMode string}`; `IlmExplainResponse{Indices map[string]{Index, Managed bool, Phase, Action, Step, StepInfo map[string]any, FailedStep, FailedStepRetryable *bool, StepTime, PhaseTime, ActionTime, StepTimeMillis, IsAutoRetryableError *bool}}`; `IlmRemovePolicyResponse{HasFailures bool, FailedIndexes []string}`; `IlmMigrateToDataTiersResponse{DryRun bool, MigratedComponentTemplates, MigratedComposableTemplates, MigratedIndices, MigratedIlmPolicies, MigratedLegacyTemplates []string, RemovedLegacyTemplate, RemovedIndexSettings, RemovedClusterSettings []string}`.

### 5.15 TransformService — spec: `transform.*`

```go
Get(ctx, transformIds []string, params *TransformGetParams) (*TransformGetResponse, error)
GetStats(ctx, transformIds []string, params *TransformGetStatsParams) (*TransformStatsResponse, error)
Put(ctx, transformId string, body any, params *TransformPutParams) (*TransformPutResponse, error)
Preview(ctx, transformId string, body any, params *TransformPreviewParams) (*TransformPreviewResponse, error) // GET|POST /_transform/_preview | /{id}/_preview
Delete(ctx, transformId string, params *TransformDeleteParams) (*AcknowledgedResponse, error)
Start(ctx, transformId string, params *TransformStartParams) (*AcknowledgedResponse, error)
Stop(ctx, transformId string, params *TransformStopParams) (*AcknowledgedResponse, error)
Reset(ctx, transformId string, params *TransformResetParams) (*AcknowledgedResponse, error)
ScheduleNow(ctx, transformId string) (*AcknowledgedResponse, error)    // POST /_transform/{id}/_schedule_now
Upgrade(ctx, params *TransformUpgradeParams) (*TransformUpgradeResponse, error) // POST /_transform/_upgrade
GetNodeStats(ctx) (*TransformNodeStatsResponse, error)                 // GET /_transform/_node_stats
```
- Params: get: allow_no_match, exclude_generated, from, size; stats: allow_no_match, from, size, timeout; put: defer_validation, timeout; preview: timeout; delete: force, timeout; start: timeout; stop: allow_no_match, force, timeout, wait_for_checkpoint, wait_for_completion; reset: force, timeout; upgrade: dry_run, timeout.
- Responses typed: `TransformGetResponse{Count int, Transforms []TransformConfig{ID, Source{Index []string, Query map[string]any}, Dest{Index, Pipeline}, Frequency, Sync map[string]any, Pivot map[string]any, Latest map[string]any, Description, Settings map[string]any, Meta map[string]any, Version, CreateTime, DestIndex? }}`; `TransformStatsResponse{Count, Transforms []{ID, State, Reason, NodeAttributes map[string]string, Stats{PagesProcessed, DocumentsProcessed, DocumentsIndexed, DocumentsDeleted, TriggerCount, ProcessingTimeInMillis, IndexTotal, IndexTimeInMillis, SearchTotal, SearchTimeInMillis, SearchFailure, Checkpoint, CheckpointProgress map[string]any, IndexFailure, DeleteFailure, ExponentialAvgCheckpointProcessingTimeInMillis, ExponentialAvgDocumentsIndexed, ExponentialAvgDocumentsProcessed}, CheckpointInfo{Last{...}, Next{...}}, Health map[string]any}}`; `TransformPutResponse{Acknowledged bool, Created bool}`; `TransformPreviewResponse{Preview []map[string]any}`; `TransformUpgradeResponse{NeedsUpdate bool, DryRun bool, Updated []string, NoAction []string}`; `TransformNodeStatsResponse{Nodes map[string]{Transforms map[string]any}}`.

### 5.16 MlService — spec: `ml.*` (~60 endpoints). Method list (signatures follow the Request/Params conventions; bodies `any`):

Jobs: `PutJob(ctx, jobId string, body any) (*MlPutJobResponse, error)`; `GetJobs(ctx, jobIds []string, params *MlGetJobsParams) (*MlGetJobsResponse, error)` (allow_no_match, exclude_generated); `DeleteJob(ctx, jobId string, params *MlDeleteJobParams) (*MlDeleteJobResponse, error)` (force, wait_for_completion); `OpenJob(ctx, jobId string, params *MlOpenJobParams) (*MlOpenJobResponse, error)` (timeout); `CloseJob(ctx, jobId string, params *MlCloseJobParams) (*MlCloseJobResponse, error)` (allow_no_match, force, timeout); `ResetJob(ctx, jobId string, params *MlResetJobParams) (*AcknowledgedResponse, error)` (wait_for_completion); `UpdateJob(ctx, jobId string, body any) (*MlJobConfig, error)`; `FlushJob(ctx, jobId string, body any, params *MlFlushJobParams) (*MlFlushJobResponse, error)` (advance_time, calc_interim, end, skip_time, start); `Forecast(ctx, jobId string, body any, params *MlForecastParams) (*MlForecastResponse, error)` (duration, expires_in, max_model_memory); `DeleteForecast(ctx, jobId, forecastId string, params *MlDeleteForecastParams) (*AcknowledgedResponse, error)` (allow_no_forecasts, timeout); `GetJobStats(ctx, jobIds []string, params *MlGetJobStatsParams) (*MlJobStatsResponse, error)` (allow_no_match); `GetOverallBuckets(ctx, jobId string, body any, params *MlOverallBucketsParams) (*MlOverallBucketsResponse, error)`; `GetBuckets(ctx, jobId string, body any, params *MlBucketsParams) (*MlBucketsResponse, error)`; `GetCategories(ctx, jobId, categoryId string, body any, params *MlCategoriesParams) (*MlCategoriesResponse, error)`; `GetRecords(ctx, jobId string, body any, params *MlRecordsParams) (*MlRecordsResponse, error)`; `GetInfluencers(ctx, jobId string, body any, params *MlInfluencersParams) (*MlInfluencersResponse, error)`; `PostData(ctx, jobId string, body string, params *MlPostDataParams) (*MlPostDataResponse, error)` (reset_end, reset_start; raw ndjson body); `GetModelSnapshots(ctx, jobId, snapshotId string, body any, params *MlModelSnapshotsParams) (*MlModelSnapshotsResponse, error)`; `DeleteModelSnapshot(ctx, jobId, snapshotId string) (*AcknowledgedResponse, error)`; `UpdateModelSnapshot(ctx, jobId, snapshotId string, body any) (*MlModelSnapshot, error)`; `RevertModelSnapshot(ctx, jobId, snapshotId string, body any, params *MlRevertModelSnapshotParams) (*MlModelSnapshot, error)` (delete_intervening_results); `UpgradeJobSnapshot(ctx, jobId, snapshotId string, params *MlUpgradeSnapshotParams) (*MlUpgradeSnapshotResponse, error)` (timeout, wait_for_completion); `GetModelSnapshotUpgradeStats(ctx, jobIds []string, snapshotIds []string) (*MlSnapshotUpgradeStatsResponse, error)` (verify endpoint name); `Validate(ctx, body any) (map[string]any, error)`; `ValidateDetector(ctx, body any) (map[string]any, error)`; `Info(ctx) (*MlInfoResponse, error)`; `GetMemoryStats(ctx, nodeId string, params *MlMemoryStatsParams) (*MlMemoryStatsResponse, error)` (timeout); `SetUpgradeMode(ctx, params *MlSetUpgradeModeParams) (*AcknowledgedResponse, error)` (enabled, timeout); `DeleteExpiredData(ctx, jobId string, body any, params *MlDeleteExpiredDataParams) (*AcknowledgedResponse, error)` (requests_per_second, timeout); `ClearTrainedModelDeploymentCache(ctx, modelId string) (*AcknowledgedResponse, error)`.

Datafeeds: `PutDatafeed(ctx, datafeedId string, body any) (*MlDatafeedConfig, error)`; `GetDatafeeds(ctx, datafeedIds []string, params *MlGetDatafeedsParams) (*MlGetDatafeedsResponse, error)`; `DeleteDatafeed(ctx, datafeedId string, params *MlDeleteDatafeedParams) (*AcknowledgedResponse, error)` (force); `StartDatafeed(ctx, datafeedId string, body any, params *MlStartDatafeedParams) (*MlStartDatafeedResponse, error)` (end, start, timeout); `StopDatafeed(ctx, datafeedId string, params *MlStopDatafeedParams) (*MlStopDatafeedResponse, error)` (allow_no_match, force, timeout); `UpdateDatafeed(ctx, datafeedId string, body any) (*MlDatafeedConfig, error)`; `GetDatafeedStats(ctx, datafeedIds []string, params *MlGetDatafeedStatsParams) (*MlDatafeedStatsResponse, error)`; `PreviewDatafeed(ctx, datafeedId string, body any, params *MlPreviewDatafeedParams) (json.RawMessage, error)` (start, end).

Calendars: `PutCalendar(ctx, calendarId string, body any) (*MlCalendar, error)`; `GetCalendars(ctx, calendarId string, body any, params *MlGetCalendarsParams) (*MlGetCalendarsResponse, error)` (from, size); `DeleteCalendar(ctx, calendarId string) (*AcknowledgedResponse, error)`; `PutCalendarJob(ctx, calendarId, jobId string) (*MlCalendar, error)`; `DeleteCalendarJob(ctx, calendarId, jobId string) (*MlCalendar, error)`; `GetCalendarEvents(ctx, calendarId string, params *MlCalendarEventsParams) (*MlCalendarEventsResponse, error)` (end, from, job_id, size, start); `PostCalendarEvents(ctx, calendarId string, body any) (*MlCalendarEventsResponse, error)`; `DeleteCalendarEvent(ctx, calendarId, eventId string) (*AcknowledgedResponse, error)`.

Filters: `PutFilter(ctx, filterId string, body any) (*MlFilter, error)`; `GetFilters(ctx, filterId string, params *MlGetFiltersParams) (*MlGetFiltersResponse, error)` (from, size); `UpdateFilter(ctx, filterId string, body any) (*MlFilter, error)`; `DeleteFilter(ctx, filterId string) (*AcknowledgedResponse, error)`.

Data frame analytics: `PutDataFrameAnalytics(ctx, id string, body any) (*MlDataFrameAnalyticsConfig, error)`; `GetDataFrameAnalytics(ctx, ids []string, params *MlGetDFAParams) (*MlGetDFAResponse, error)` (allow_no_match, exclude_generated, from, size); `DeleteDataFrameAnalytics(ctx, id string, params *MlDeleteDFAParams) (*AcknowledgedResponse, error)` (force, timeout); `StartDataFrameAnalytics(ctx, id string, params *MlStartDFAParams) (*MlStartDFAResponse, error)` (timeout); `StopDataFrameAnalytics(ctx, id string, params *MlStopDFAParams) (*MlStopDFAResponse, error)` (allow_no_match, force, timeout); `UpdateDataFrameAnalytics(ctx, id string, body any) (*MlDataFrameAnalyticsConfig, error)`; `GetDataFrameAnalyticsStats(ctx, ids []string, params *MlDFAStatsParams) (*MlDFAStatsResponse, error)` (allow_no_match, from, size, verbose); `ExplainDataFrameAnalytics(ctx, id string, body any) (*MlExplainDFAResponse, error)`; `EvaluateDataFrame(ctx, body any) (*MlEvaluateDFAResponse, error)`; `PreviewDataFrameAnalytics(ctx, id string, body any) (*MlPreviewDFAResponse, error)`.

Trained models: `PutTrainedModel(ctx, modelId string, body any, params *MlPutTrainedModelParams) (*MlTrainedModelConfig, error)` (defer_definition_decompression, wait_for_completion); `GetTrainedModels(ctx, modelIds []string, params *MlGetTrainedModelsParams) (*MlGetTrainedModelsResponse, error)` (allow_no_match, decompress_definition, exclude_generated, from, include, size, tags); `DeleteTrainedModel(ctx, modelId string, params *MlDeleteTrainedModelParams) (*AcknowledgedResponse, error)` (force, timeout); `GetTrainedModelsStats(ctx, modelIds []string, params *MlGetTrainedModelsStatsParams) (*MlTrainedModelsStatsResponse, error)` (allow_no_match, from, size); `CreateTrainedModelsAlias(ctx, modelId, modelAlias string, params *MlCreateAliasParams) (*MlTrainedModelAlias, error)` (reassign); `DeleteTrainedModelsAlias(ctx, modelId, modelAlias string) (*AcknowledgedResponse, error)`; `InferTrainedModel(ctx, modelId string, body any, params *MlInferTrainedModelParams) (json.RawMessage, error)` (timeout); `StartTrainedModelDeployment(ctx, modelId string, params *MlStartDeploymentParams) (*MlStartDeploymentResponse, error)` (cache_size, deployment_id, number_of_allocations, priority, queue_capacity, threads_per_allocation, timeout, wait_for); `StopTrainedModelDeployment(ctx, modelId string, params *MlStopDeploymentParams) (*AcknowledgedResponse, error)` (force); `UpdateTrainedModelDeployment(ctx, modelId string, body any, params *MlUpdateDeploymentParams) (*MlStartDeploymentResponse, error)` (number_of_allocations).

Response structs: fully typed per ES docs (job config: JobId, Description, AnalysisConfig{BucketSpan, Detectors[]{Function, FieldName, ByFieldName, OverFieldName, PartitionFieldName}, Influencers[], SummaryCountField, ModelPlotConfig, Latency, MultivariateByCounts}, DataDescription{TimeField, TimeFormat}, AnalysisLimits, BackgroundPersistInterval, ModelSnapshotRetentionDays, ResultsRetentionDays, CustomSettings, AllowLazyOpen, MaxModelMemory; stats: DataCounts{ProcessedRecords, ProcessedFields, InputBytes,...}, ModelSizeStats, TimingStats, State, OpenTime...). Result types (buckets/records/influencers/categories) typed with map[string]any for the aggregate payloads where ES shape is dynamic.

### 5.17 SecurityService — spec: `security.*` (~70 endpoints). Signatures (grouped):

Users: `GetUser(ctx, usernames []string) (map[string]*SecurityUser, error)`; `PutUser(ctx, username string, body any) (*SecurityCreateResponse, error)`; `DeleteUser(ctx, username string) (*AcknowledgedResponse, error)`; `ChangePassword(ctx, username string, body any) (*AcknowledgedResponse, error)` (POST /_security/user/_password | /{username}/_password); `DisableUser(ctx, username string) (*AcknowledgedResponse, error)`; `EnableUser(ctx, username string) (*AcknowledgedResponse, error)`; `SetRoles(ctx, username string, body any) (*SecurityCreateResponse, error)` (PUT /_security/user/{username}/_roles).
Roles: `GetRole(ctx, names []string) (map[string]*SecurityRole, error)`; `PutRole(ctx, name string, body any) (*SecurityCreateResponse, error)`; `DeleteRole(ctx, name string) (*AcknowledgedResponse, error)`; `ClearRoleCache(ctx, names []string) (*AcknowledgedResponse, error)`.
Role mappings: `GetRoleMapping(ctx, names []string) (map[string]*SecurityRoleMapping, error)`; `PutRoleMapping(ctx, name string, body any) (*SecurityRoleMappingCreateResponse, error)`; `DeleteRoleMapping(ctx, name string) (*AcknowledgedResponse, error)`.
Privileges: `GetPrivileges(ctx, application, name string) (map[string]any, error)`; `PutPrivileges(ctx, body any) (*SecurityPutPrivilegesResponse, error)`; `DeletePrivileges(ctx, application, name string) (*AcknowledgedResponse, error)`; `HasPrivileges(ctx, body any) (*SecurityHasPrivilegesResponse, error)`; `GetUserPrivileges(ctx) (map[string]any, error)`; `GetBuiltinPrivileges(ctx) (*SecurityBuiltinPrivilegesResponse, error)`.
API keys: `CreateAPIKey(ctx, body any) (*SecurityCreateAPIKeyResponse, error)`; `GetAPIKey(ctx, params *SecurityGetAPIKeyParams) (*SecurityGetAPIKeyResponse, error)` (id, name, realm_name, username, owner, active_only, with_limited_privileges, with_profile_uid — verify); `InvalidateAPIKey(ctx, body any) (*SecurityInvalidateAPIKeyResponse, error)`; `UpdateAPIKey(ctx, id string, body any) (*SecurityUpdateAPIKeyResponse, error)`; `QueryAPIKeys(ctx, body any, params *SecurityQueryAPIKeysParams) (*SecurityQueryAPIKeysResponse, error)`; `GrantAPIKey(ctx, body any) (*SecurityCreateAPIKeyResponse, error)`.
Tokens: `GetToken(ctx, body any) (*SecurityGetTokenResponse, error)`; `InvalidateToken(ctx, body any) (*SecurityInvalidateTokenResponse, error)`.
Authenticate: `Authenticate(ctx) (*SecurityAuthenticateResponse, error)`; `ClearRealmCache(ctx, realms []string, params *SecurityClearRealmCacheParams) (*AcknowledgedResponse, error)` (usernames).
Service accounts: `GetServiceAccounts(ctx, namespace, service string) (map[string]*SecurityServiceAccount, error)`; `GetServiceCredentials(ctx, namespace, service string) (*SecurityServiceCredentialsResponse, error)`; `CreateServiceToken(ctx, namespace, service, tokenName string) (*SecurityServiceTokenCreateResponse, error)`; `DeleteServiceToken(ctx, namespace, service, tokenName string) (*AcknowledgedResponse, error)`.
SSO: `SamlAuthenticate(ctx, body any) (*SecuritySamlAuthenticateResponse, error)`; `SamlPrepareAuthentication(ctx, body any) (*SecuritySamlPrepareResponse, error)`; `SamlLogout(ctx, body any) (*SecuritySamlLogoutResponse, error)`; `SamlInvalidate(ctx, body any) (*SecuritySamlInvalidateResponse, error)`; `SamlCompleteLogout(ctx, body any) (string, error)`; `SamlServiceProviderMetadata(ctx, realmName string) (string, error)` (verify name); `OidcAuthenticate(ctx, body any) (*SecurityOidcAuthenticateResponse, error)`; `OidcPrepareAuthentication(ctx, body any) (*SecurityOidcPrepareResponse, error)`; `OidcLogout(ctx, body any) (*SecurityOidcLogoutResponse, error)`; `KerberosAuthenticate(ctx, body any) (*SecurityKerberosAuthenticateResponse, error)`.
Enrollment: `EnrollKibana(ctx) (*SecurityEnrollKibanaResponse, error)`; `EnrollNode(ctx) (*SecurityEnrollNodeResponse, error)`.
Settings: `GetSettings(ctx, names []string) (*SecurityGetSettingsResponse, error)`; `UpdateSettings(ctx, body any, params *SecurityUpdateSettingsParams) (*AcknowledgedResponse, error)` (timeout).
Query: `QueryUser(ctx, body any) (*SecurityQueryResponse, error)`; `QueryRole(ctx, body any) (*SecurityQueryResponse, error)`.
Profiles: `ActivateUserProfile(ctx, body any) (*SecurityProfileResponse, error)`; `GetUserProfile(ctx, uids []string, params *SecurityGetProfileParams) (*SecurityGetProfileResponse, error)` (data); `SuggestUserProfiles(ctx, body any) (*SecuritySuggestProfilesResponse, error)`; `UpdateUserProfileData(ctx, uid string, body any, params *SecurityUpdateProfileDataParams) (*AcknowledgedResponse, error)` (if_seq_no, if_primary_term).

Core response types: `SecurityUser{Username, Roles []string, FullName, Email, Metadata map[string]any, Enabled bool}`; `SecurityRole{Cluster []string, Indices []{Names []string, Privileges []string, FieldSecurity map[string]any, Query string, AllowRestrictedIndices *bool}, Applications []{Application, Privileges []string, Resources []string}, RunAs []string, Metadata map[string]any, TransientMetadata map[string]any}`; `SecurityRoleMapping{Enabled bool, Roles []string, RoleTemplates []map[string]any, Rules map[string]any, Metadata map[string]any}`; `SecurityCreateResponse{Created bool}`; `SecurityCreateAPIKeyResponse{ID, Name, APIKey json:"api_key", Expiration int64, Encoded string}`; `SecurityAuthenticateResponse{Username, Roles []string, FullName, Email, Metadata, AuthenticationRealm{Name,Type}, LookupRealm{Name,Type}, AuthenticationType, ApiKey map[string]any, Token map[string]any}`.

### 5.18 SqlService — spec: `sql.*`

```go
Query(ctx, body any, params *SqlQueryParams) (*SqlQueryResponse, error)        // POST /_sql (params: format)
Translate(ctx, body any) (*SqlTranslateResponse, error)                        // POST /_sql/translate
ClearCursor(ctx, body any) (*SqlClearCursorResponse, error)                    // POST /_sql/close ({cursor})
GetAsync(ctx, id string, params *SqlGetAsyncParams) (*SqlQueryResponse, error) // GET /_sql/async/{id} (delimiter, format, keep_alive, wait_for_completion_timeout)
GetAsyncStatus(ctx, id string) (*SqlAsyncStatusResponse, error)                // GET /_sql/async/status/{id}
DeleteAsync(ctx, id string) (*AcknowledgedResponse, error)                     // DELETE /_sql/async/{id}
```
- `SqlQueryResponse{Columns []{Name, Type}, Rows [][]any, Cursor string, IsRunning bool}`; `SqlTranslateResponse{Size int, Query map[string]any, Sort []map[string]any}`; `SqlClearCursorResponse{Succeeded bool}`; `SqlAsyncStatusResponse{ID, IsRunning, IsPartial, StartInMillis, ExpirationInMillis, CompletionStatus int}`.

### 5.19 WatcherService — spec: `watcher.*`

```go
PutWatch(ctx, id string, body any, params *WatcherPutWatchParams) (*WatcherPutWatchResponse, error) // active, if_primary_term, if_seq_no
GetWatch(ctx, id string) (*WatcherGetWatchResponse, error)
DeleteWatch(ctx, id string) (*WatcherDeleteWatchResponse, error)
ExecuteWatch(ctx, id string, body any, params *WatcherExecuteWatchParams) (*WatcherExecuteWatchResponse, error) // debug
AckWatch(ctx, id, actionId string) (*WatcherAckWatchResponse, error)           // PUT /_watcher/watch/{id}/_ack | /_ack/{action_id}
ActivateWatch(ctx, id string) (*WatcherActivateWatchResponse, error)
DeactivateWatch(ctx, id string) (*WatcherActivateWatchResponse, error)
QueryWatches(ctx, body any) (*WatcherQueryWatchesResponse, error)              // GET|POST /_watcher/_query/watches
Start(ctx) (*AcknowledgedResponse, error)                                      // POST /_watcher/_start
Stop(ctx) (*AcknowledgedResponse, error)                                       // POST /_watcher/_stop
Stats(ctx, metrics []string, params *WatcherStatsParams) (*WatcherStatsResponse, error) // GET /_watcher/stats | /_watcher/stats/{metric} (emit_stacktraces)
```
- Responses: `WatcherPutWatchResponse{ID, Version int64, Created bool}`; `WatcherGetWatchResponse{Found bool, ID, Status map[string]any, Watch map[string]any}`; `WatcherExecuteWatchResponse{ID, WatchRecord map[string]any}`; `WatcherAckWatchResponse{Status map[string]any}`; `WatcherQueryWatchesResponse{Count int, Watches []{ID, Watch map[string]any, Status map[string]any}}`; `WatcherStatsResponse{WatcherState string, Stats map[string]any, Queues map[string]any, CurrentWatches []map[string]any}`.

### 5.20 CcrService — spec: `ccr.*`

```go
PutAutoFollowPattern(ctx, name string, body any) (*AcknowledgedResponse, error)
GetAutoFollowPattern(ctx, names []string) (*CcrGetAutoFollowPatternResponse, error)
DeleteAutoFollowPattern(ctx, name string) (*AcknowledgedResponse, error)
PauseAutoFollowPattern(ctx, name string) (*AcknowledgedResponse, error)
ResumeAutoFollowPattern(ctx, name string) (*AcknowledgedResponse, error)
Follow(ctx, index string, body any, params *CcrFollowParams) (*CcrFollowResponse, error)   // wait_for_active_shards
PauseFollow(ctx, index string) (*AcknowledgedResponse, error)
ResumeFollow(ctx, index string, body any) (*AcknowledgedResponse, error)
Unfollow(ctx, index string) (*AcknowledgedResponse, error)
ForgetFollower(ctx, index string, body any) (*CcrForgetFollowerResponse, error)
FollowInfo(ctx, indices []string) (*CcrFollowInfoResponse, error)
FollowStats(ctx, indices []string) (*CcrFollowStatsResponse, error)
Stats(ctx) (*CcrStatsResponse, error)
```
- Typed responses: `CcrFollowResponse{FollowIndex, ShardsAcknowledged, FollowIndexCreated, FollowIndexShardsAcknowledged}`; `CcrStatsResponse{AutoFollowStats{NumberOfFailedFollowIndices, NumberOfFailedRemoteClusterStateRequests, RecentAutoFollowErrors []{...}, AutoFollowedClusters []{...}}, FollowStats{Indices []{Index, Status map[string]any}}}`; `CcrFollowInfoResponse{Indices []{Index, FollowIndex, RemoteCluster, LeaderIndex}}`; `CcrForgetFollowerResponse{Shards *ShardsInfo}`.

### 5.21 Remaining services (compact)

| Service | Endpoints (method -> HTTP path) | Params of note |
|---|---|---|
| AutoscalingService | GetCapacity GET /_autoscaling/capacity; GetPolicy GET /_autoscaling/policy/{name}; PutPolicy PUT; DeletePolicy DELETE | none (verify master_timeout) |
| EnrichService | GetPolicy GET /_enrich/policy[/{name}]; PutPolicy PUT /_enrich/policy/{name}; DeletePolicy DELETE; ExecutePolicy PUT /_enrich/policy/{name}/_execute (wait_for_completion); Stats GET /_enrich/_stats | — |
| EqlService | Search GET\|POST /{index}/_eql/search (keep_alive, keep_on_completion, wait_for_completion_timeout); Get GET /_eql/search/{id}; GetStatus GET /_eql/search/status/{id}; Delete DELETE /_eql/search/{id} | typed `EqlSearchResponse{ID, IsPartial, IsRunning, TimedOut, Took, Hits{Total{Value,Relation}, Events[]map[string]any, Sequences[]map[string]any}}` |
| GraphService | Explore GET\|POST /{index}/_graph/explore (routing, timeout) | `GraphExploreResponse{Took, TimeOut bool, Failures []map[string]any, Vertices[]{ID, Weight, Depth}, Connections[]{Source, Target, Weight, DocCount}}` |
| LicenseService | Get GET /_license (local); Delete DELETE; Post PUT\|POST /_license (acknowledge); GetBasicStatus GET /_license/basic_status; PostStartBasic POST /_license/start_basic (acknowledge); GetTrialStatus GET /_license/trial_status; PostStartTrial POST /_license/start_trial (acknowledge, type) | `LicenseGetResponse{License{UID,Type,Status,IssueDate,IssueDateInMillis,ExpiryDate,ExpiryDateInMillis,MaxNodes,IssuedTo,Issuer,StartDate,StartDateInMillis,Signature}}` |
| LogstashService | GetPipeline GET /_logstash/pipeline[/{id}]; PutPipeline PUT; DeletePipeline DELETE | responses map[string]map[string]any / Acknowledged |
| MigrationService | Deprecations GET /_migration/deprecations \| /{index}/_migration/deprecations; GetFeatureUpgradeStatus GET /_migration/feature_upgrade; PostFeatureUpgrade POST /_migration/feature_upgrade | `MigrationDeprecationsResponse{ClusterSettings []DeprecationIssue, NodeSettings []..., IndexSettings map[string][]..., MLSettings []...}`; `MigrationFeatureUpgradeStatusResponse{Features []{FeatureName, FeatureStatus, MinimumIndexVersion, Indices []IndexUpgradeInfo}}` |
| ShutdownService | GetNode GET /_nodes/shutdown \| /_nodes/{node_id}/shutdown; PutNode PUT /_nodes/{node_id}/shutdown; DeleteNode DELETE | `ShutdownGetNodeResponse{Nodes []{NodeID, Type, Reason, ShutdownStarted, ShutdownCompleted, AllocationDelay, TargetNodeName, Status map[string]any}}` |
| FeaturesService | GetFeatures GET /_features; ResetFeatures POST /_features/_reset (verify deprecation) | `FeaturesGetResponse{Features []{Name, Description}}` |
| FleetService | GlobalCheckpoints GET /{index}/_fleet/global_checkpoints (checkpoints, wait_for_advance, wait_for_index, timeout) | `FleetGlobalCheckpointsResponse{GlobalCheckpoints []int64, TimedOut bool}` |
| DanglingIndicesService | List GET /_dangling; Import POST /_dangling/{index_uuid} (accept_data_loss **req**); Delete DELETE /_dangling/{index_uuid} (accept_data_loss **req**) | `DanglingIndicesListResponse{DanglingIndices []{IndexUUID, IndexName, CreationDate}}` |
| HealthReportService | Get GET /_health_report \| /_health_report/{feature} (size, time, verbose) | `HealthReportResponse{ClusterName, Status, Indicators map[string]HealthIndicator{Status, Summary, Detail, UserActions []{Action, HelpURL}, Impacts []{ID, Severity, Description, ImpactAreas []}, Diagnoses []{ID, Cause, Action, HelpURL, AffectedResources map[string][]string}}}` |
| QueryRulesService | ListRulesets GET /_query_ruleset (from, size); GetRuleset GET /_query_ruleset/{ruleset_id}; PutRuleset PUT; DeleteRuleset DELETE; GetRule/ PutRule/ DeleteRule /_query_ruleset/{ruleset_id}/_rule/{rule_id} (verify presence); Test GET\|POST /_query_ruleset/{ruleset_id}/_test (query_text, query_params) | typed list/get/test responses |
| SearchApplicationService | List GET /_application/search_application (from, size); Get GET /{name}; Put PUT /{name} (create); Delete DELETE; Search GET\|POST /{name}/_search (typed `querydsl.SearchResult`); GetBehavioralAnalytics GET /_application/analytics[/{name}]; PutBehavioralAnalytics PUT; DeleteBehavioralAnalytics DELETE | typed responses |
| SynonymsService | GetSynonymsSets GET /_synonyms (from, size); GetSynonym GET /_synonyms/{id} (from, size); PutSynonym PUT; DeleteSynonym DELETE; GetSynonymRule GET /_synonyms/{id}/_rule/{rule_id}; PutSynonymRule PUT; DeleteSynonymRule DELETE | `SynonymsSetsResponse{Count, Results []{SynonymSet, Count}}`; `SynonymRulesResponse{Count, Rules []{ID, Type, Synonyms}}` |
| TextStructureService | FindStructure POST /_text_structure/find_structure; FindFieldStructure GET /{index}/_text_structure/field_structure (field **req**); FindMessageStructure GET /{index}/_text_structure/message_structure (query, size); TestGrokPattern GET\|POST /_text_structure/test_grok_pattern (grok_pattern **req**, text) | shared params: charset, column_names, delimiter, explain, format, grok_pattern, has_header_row, line_merge_size_limit, lines_to_sample, quote, should_trim_fields, timestamp_field, timestamp_format; responses typed (`TextStructureResponse{NDJSON bool, NeedClientTimezone bool, HasHeaderRow, HasByteOrderMarker, LinesMerged, Format, Charset, Delimiter, Quote, OriginalQuote, ColumnNames, GrokPattern, TimestampField, TimestampFormat, ShouldTrimFields, Mappings map[string]any, IngestPipeline map[string]any, Explanation []string, SampleStart, SampleSize, NumLinesAnalyzed, NumMessagesAnalyzed}`) |
| InferenceService | Get GET /_inference \| /_inference/{inference_id} \| /_inference/{task_type}/{inference_id}; Put PUT /_inference/{task_type}/{inference_id} (dry_run verify); Delete DELETE (force, dry_run); Update POST /_inference/{inference_id}/_update (verify); Completion POST /_inference/completion/{inference_id}; ChatCompletion POST /_inference/chat_completion/{inference_id} (verify); Rerank POST /_inference/rerank/{inference_id}; Inference POST /_inference/{inference_id} (generic) | responses: `InferenceGetResponse{InferenceID, TaskType, Service, ServiceSettings map[string]any, TaskSettings map[string]any}`; completion/rerank return `json.RawMessage`-backed typed wrappers |
| SearchableSnapshotsService | Mount POST /_snapshot/{repository}/{snapshot}/_mount (master_timeout? verify, wait_for_completion); Stats GET /_searchable_snapshots/stats \| /{index}/... (level); CacheStats GET /_searchable_snapshots/cache/stats \| /{node_id}/...; ClearCache POST /{index}/_searchable_snapshots/cache/clear; RepositoryStats GET /_snapshot/{repository}/_stats (verify deprecation) | typed mount/stats responses |
| SslService | Certificates GET /_ssl/certificates | `SslCertificatesResponse []SslCertificate{Path, SubjectDN, SerialNumber, HasPrivateKey, Alias, IssuerDN, Version, NotBefore, NotAfter}` |
| XpackService | Info GET /_xpack (categories); Usage GET /_xpack/usage | `XpackInfoResponse{Build{Hash,Timestamp}, License{...}, Features map[string]XpackFeature{Available,Enabled,NativeRealms?,Description?}, Tagline}`; `XpackUsageResponse` = map[string]XpackUsage (typed struct with available/enabled + feature-specific map[string]any) |
| StreamsService (if present) | LogsEnable POST /_streams/logs/_enable; LogsDisable POST /_streams/logs/_disable; Status GET /_streams/logs/_status | typed `{Acknowledged bool}` / status map |

**RollupService / KnnSearchService: N/A in ES 9** (removed upstream). Document in README; confirm absence via speccheck.

---

## 6. Edge cases & error handling

1. **Connection failures / timeouts**: resty transport errors -> `wrapNetworkError` (log + return original error, preserving `errors.Is/As` chains). `Config.Timeout` maps to resty `SetTimeout`; zero = no timeout. Context deadline propagates via `SetContext(ctx)` on every request; a canceled ctx surfaces as `context.Canceled`/`context.DeadlineExceeded` wrapped by resty — unit-test both.
2. **Retries**: `RetryCount` + `DefaultRetryConditions` (network errors, 429, 5xx except 501); `ESSearchRetryConditions` adds ES transient-error body sniffing. Retries must NOT retry non-idempotent writes blindly: document that retrying POST/PUT writes (bulk/index) is opt-in via `RetryCount` and can duplicate auto-ID documents; keep defaults `RetryCount: 0` as in opensearch.
3. **ES error responses**: `parseErrorResponse` unmarshals `{"error":{...},"status":N}` into `*types.ElasticsearchError`; must tolerate `"error": "<string>"` legacy shape (custom UnmarshalJSON). Status always backfilled from HTTP status. Helpers `IsNotFound` (404), `IsConflict` (409), `IsUnauthorized` (401).
4. **HEAD semantics**: exists-style methods return `(false, nil)` on 404, error otherwise.
5. **Document GET 404**: return `&GetResult{Found:false}` (parity with opensearch); document DELETE 404 returns parsed response + `fmt.Errorf("document not found")` (opensearch parity) — keep and unit-test.
6. **Response parsing errors**: `wrapUnmarshalError` logs truncated body (500 chars) and returns `fmt.Errorf("unmarshal response from %s %s: %w", method, url, err)`.
7. **NDJSON endpoints**: bulk/msearch/msearch_template set `Content-Type: application/x-ndjson`; body must end with `\n` — helper `ensureTrailingNewline` in options.go; unit-test.
8. **Cat APIs**: always set `format=json` unless caller overrides via params (Help returns raw text).
9. **TLS/certificates**: CACert PEM parse failure -> typed error at `New`; `TLSSkipVerify` documented dev-only; system pool default. HTTP/2 disable path ported verbatim.
10. **Auth conflicts**: more than one of {basic, apikey, bearer} set -> error at `New`.
11. **Product header**: do NOT hard-fail on missing `X-Elastic-Product` (proxies strip headers); log at debug when absent.
12. **Large responses**: responses decoded from `resp.Body()` in memory (resty default) — document; no streaming in v9.
13. **Empty optional slices**: `[]string` params joined with ","; empty slice -> omit path segment or use cluster-wide route (per spec path variants).
14. **Typed-nil bodies**: `normalizeBody` guards typed-nil `bodyProvider` (port opensearch fix).

---

## 7. Validation strategy

- **Required vs optional**: required path parts and required params carry `validate:"required"` (validator v10, `WithRequiredStructEnabled`); `Validate()` called at method entry; errors wrapped as `validation: ...` (opensearch `validationError`). Optional params are zero-value-omitted in `ToMap()`.
- **Numeric optionality**: use pointers (`*int`, `*int64`, `*bool`) wherever `0`/`false` is a legal explicit value (e.g. `from=0`, `realtime=false`, `version=0`).
- **Enums**: typed string constants (§4.3); invalid enum values are rejected by the server (400) — client does not re-validate enum membership except where cheap (Refresh, VersionType, SearchType, ExpandWildcards, HealthStatus) via `Validate()` custom rules.
- **Type safety**: path parts are `string`/`[]string`/`int` typed; time/duration params are `string` ("30s") matching spec `time` type; `list` params are `[]string` joined by comma.
- **Spec-level guarantee**: `internal/speccheck` (§8.4) fails CI if any spec param lacks a struct field or any endpoint lacks a method.

---

## 8. Testing strategy

### 8.1 Unit tests (every component, no build tag)
Pattern from opensearch `*_service_test.go`: `httptest.NewServer` per test, assert method+path+query+body, return canned JSON; cover per method: success, params serialization (assert query string), validation error, HTTP error (500 with ES error body -> `*types.ElasticsearchError`), network error (`deadClient` = resty pointed at closed port), 404 semantics. Shared helpers in `api/test_helpers_test.go`: `restyClient(srv)`, `deadClient()`, `testLogger()`.
Root package: `client_test.go` (New config matrix: auth exclusivity, CACert parse failure, TLS skip, transport/HTTP2 flags, product header logging), `client_retry_test.go` (retry condition table), `errors_test.go`, `types/*_test.go` (error unmarshal both shapes, UnixMilliTime round-trip), `querydsl` tests ported as-is.

### 8.2 Acceptance tests (`//go:build integration`, package `api_test`)
- `integration_helper_test.go`: build client from `ELASTICSEARCH_URL`, `ELASTICSEARCH_USERNAME`, `ELASTICSEARCH_PASSWORD` (+ `TLSSkipVerify: true`); panic with guidance if unset.
- One `*_service_integration_test.go` per service (§3). Test shape: unique index names per test (`fmt.Sprintf("test-%s-%d", name, time.Now().UnixNano())`), cleanup via `defer`, CRUD cycles, 404/409 assertions via `IsNotFound`/`IsConflict`.
- ES-specific flows: scroll + clear scroll, PIT open/close + search_after, bulk mixed ops, reindex + task cancel, snapshot repo(fs)+snapshot+restore+clone+cleanup, SLM policy CRUD, ILM policy + explain, transform put/start/stop/delete, ML job lifecycle (put/open/post data/close/delete) + trained-model list, security role/user/API-key CRUD, watcher put/execute/delete (trial license), SQL query against seeded index, EQL search, enrich policy, CCR skipped gracefully if single node (t.Skip with reason), cat endpoints sanity (non-empty rows), cluster health/stats/settings round-trip, ingest pipeline simulate, scripts CRUD, synonyms/query-rules/search-application CRUD, text_structure find_structure with sample CSV.
- Tests requiring trial features assert license first (`client.License().GetTrialStatus`) and `t.Skip` if unavailable.

### 8.3 Mocks
`generate.go` with mockgen directives for `client.go` + every service interface (opensearch parity). Generated `mock/` package is committed; consumers use it to stub services. Unit tests themselves use httptest (not mocks), matching opensearch.

### 8.4 Spec coverage audit — `internal/speccheck`
- CLI: `go run ./internal/speccheck -spec <rest-api-spec/api dir>`.
- Parses each spec JSON (skip `_internal.*`, `visibility != public` unless feature_flag listed in an allow-list: streams.*, connector.* if present).
- Loads a registry (`internal/speccheck/registry.go`) maintained alongside services: `map[endpointName]EndpointImpl{Service, Method, ParamsStruct, ParamsFields []string, PathParts []string}`. The registry is hand-maintained but verified by a unit test that uses reflection to assert every listed Params field actually exists on the struct (compile-time-ish safety).
- Failures: missing endpoint, missing param field, missing path part. Wired into CI (§9) as `go run ./internal/speccheck -spec /specs/api`.
- Also emits the count report used in PR descriptions.

---

## 9. CI / Dagger pipeline

### 9.1 `dagger.json` (repo root) — mirrors opensearch
```json
{
  "name": "elasticsearch",
  "engineVersion": "v0.20.8",
  "sdk": { "source": "go" },
  "dependencies": [
    { "name": "codecov",    "source": "github.com/disaster37/dagger-library-go/codecov@2.0.0" },
    { "name": "git-module", "source": "github.com/disaster37/dagger-library-go/git-module@2.0.0" },
    { "name": "golang",     "source": "github.com/disaster37/dagger-library-go/golang@2.0.0" }
  ],
  "source": "ci/dagger",
  "disableDefaultFunctionCaching": true
}
```
(Pins: run `dagger init`/`dagger develop` during implementation to fill `pin` hashes, as in opensearch's dagger.json.)

### 9.2 `ci/dagger/main.go` — structure (mirrors opensearch `main.go`)
```go
const (
    ElasticsearchVersion = "<latest 9.x patch>"        // e.g. 9.1.x
    username             = "elastic"
    password             = "<fixed strong CI password>"
    gitUsername          = "ci"
    gitEmail             = "ci@localhost"
    defaultGitBranch     = "main"
)

type Elasticsearch struct {
    Src          *dagger.Directory // +private
    GolangModule *dagger.Golang    // +private
}

func New(ctx, src *dagger.Directory) (*Elasticsearch, error)
    // GolangModule: dag.Golang(src.WithoutDirectory("ci"))

func (h *Elasticsearch) Ci(ctx, ci bool, codeCoveToken *dagger.Secret, gitBranch string, isTag bool, gitToken *dagger.Secret) (*dagger.Directory, error)
    // 1. go mod tidy (re-wrap GolangModule with base container, as opensearch)
    // 2. Build()          -> go build
    // 3. Vet()            -> go vet ./...
    // 4. Lint()           -> golangci-lint (Golang module)
    // 5. Format()         -> gofumpt (Golang module)
    // 6. SpecCheck()      -> clone elastic/elasticsearch (shallow, tag), run go run ./internal/speccheck
    // 7. TestUnit()       -> go test ./... WITHOUT integration tag (coverage.out)
    // 8. TestAcceptance() -> go test -tags integration ./api/... against ES service (coverage merged)
    // 9. if ci: CodeCov upload + git-module CommitAndPush of coverage.out (as opensearch)

func (h *Elasticsearch) Build(ctx) *dagger.Directory            // h.GolangModule.Build()
func (h *Elasticsearch) Lint(ctx) (string, error)               // h.GolangModule.Lint(ctx)
func (h *Elasticsearch) Format(ctx) *dagger.Directory           // h.GolangModule.Format()
func (h *Elasticsearch) Vet(ctx) (string, error)                // h.GolangModule.Container().WithExec(["go","vet","./..."]).Stdout(ctx)
func (h *Elasticsearch) SpecCheck(ctx) (string, error)          // see below
func (h *Elasticsearch) Test(ctx, run string) *dagger.File      // unit tests via GolangModule.Test
func (h *Elasticsearch) TestAcceptance(ctx, run string) *dagger.File
func (h *Elasticsearch) DebugTest(ctx, run string) *dagger.Service  // parity with opensearch
func (h *Elasticsearch) Elasticsearch(ctx) *dagger.Service      // the ES container
func (h *Elasticsearch) CodeCov(ctx, src *dagger.Directory, token *dagger.Secret) (string, error)
```

### 9.3 ES container (`Elasticsearch()` function)
```go
svc := dag.Container().
    From("docker.elastic.co/elasticsearch/elasticsearch:" + ElasticsearchVersion).
    WithUser("root").
    WithExec([]string{"mkdir", "-p", "/usr/share/elasticsearch/backup"}).
    WithExec([]string{"chown", "-R", "elasticsearch:root", "/usr/share/elasticsearch/backup"}).
    WithUser("elasticsearch").
    WithEnvVariable("cluster.name", "elasticsearch-test").
    WithEnvVariable("node.name", "es-node1").
    WithEnvVariable("discovery.type", "single-node").
    WithEnvVariable("network.publish_host", "0.0.0.0").
    WithEnvVariable("bootstrap.memory_lock", "true").
    WithEnvVariable("xpack.security.enabled", "true").
    WithEnvVariable("ELASTIC_PASSWORD", password).
    WithEnvVariable("xpack.license.self_generated.type", "trial").
    WithEnvVariable("path.repo", "/usr/share/elasticsearch/backup").
    WithEnvVariable("ES_JAVA_OPTS", "-Xms1g -Xmx1g").
    WithEnvVariable("CLEAN_CACHE", time.Now().String()).
    WithExposedPort(9200).
    AsService()
```
Health wait (alpine/curl sidecar container, opensearch pattern):
```
curl --fail -k -u elastic:<password> "https://elasticsearch.svc:9200/_cluster/health?wait_for_status=yellow&timeout=500s"
```
Fallback if `bootstrap.memory_lock=true` fails in the Dagger runtime: set `false` (documented in plan risks).

### 9.4 Acceptance test execution
```go
return h.GolangModule.Container().
    WithServiceBinding("elasticsearch.svc", esService).
    WithEnvVariable("ELASTICSEARCH_URL", "https://elasticsearch.svc:9200").
    WithEnvVariable("ELASTICSEARCH_USERNAME", username).
    WithSecretVariable("ELASTICSEARCH_PASSWORD", dag.SetSecret("es-password", password)).
    WithExec(helper.ForgeScript(`
      curl --fail -k -u elastic:<pw> "https://elasticsearch.svc:9200/_cluster/health?wait_for_status=yellow&timeout=500s"
      curl --fail -XPUT -k -u elastic:<pw> -H 'Content-Type: application/json' "https://elasticsearch.svc:9200/_index_template/ci-base" -d '{"index_patterns":["test-*"],"priority":500,"template":{"settings":{"number_of_shards":1,"number_of_replicas":0}}}'
      go install gotest.tools/gotestsum@latest
      gotestsum --format testname -- -tags integration -covermode=atomic -coverprofile coverage.out ./... <run>
    `)).
    File("coverage.out")
```

### 9.5 SpecCheck step
```go
// inside the golang container:
git clone --depth 1 --branch <v9 tag> https://github.com/elastic/elasticsearch /tmp/es-spec
go run ./internal/speccheck -spec /tmp/es-spec/rest-api-spec/src/main/resources/rest-api-spec/api
```

### 9.6 `.github/workflows/workflow.yaml` — mirror opensearch
```yaml
name: "build"
on:
  push: { branches: [main, release-branch.v9], tags: ["*"] }
  pull_request: { branches: [main, release-branch.v9] }
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: dagger/dagger-for-github@8.0.0
        with:
          version: "0.20.8"
          verb: call
          args: --src . ci --ci --code-cove-token env:CODECOV_TOKEN --git-token env:GITHUB_TOKEN --git-branch ${BRANCH_NAME} ${IS_TAG}
        env: { CODECOV_TOKEN: ..., GITHUB_TOKEN: ..., IS_TAG: ..., BRANCH_NAME: ... }
```

---

## 10. Implementation phases (ordered task list)

**Phase 0 — Setup & spec extraction**
1. `git clone` both repos to `/tmp` (§2); check out latest v9 tag; produce endpoint inventory (`ls api/*.json`), confirm rollup/knn_search absence, confirm streams/connector presence; record exact spec tag in README.
2. Scaffold repo: go.mod (`github.com/disaster37/elasticsearch/v9`, go 1.26), .gitignore, LICENSE, doc.go, README skeleton, dagger.json, ci/dagger skeleton, .github workflow.

**Phase 1 — Core plumbing**
3. `types` package (common.go, errors.go, time.go) + tests.
4. Root package: client.go (Config incl. APIKey/BearerToken, Client iface with all accessors, DefaultClient, New, retry conditions), errors.go, common.go + tests (auth exclusivity, TLS, transport).
5. `api` plumbing: errors.go, options.go, param_types.go, common_params.go, test_helpers_test.go + tests.
6. Port `trace/opentelemetry` + tests; port `generate.go` (mockgen list).

**Phase 2 — querydsl port**
7. Copy opensearch `querydsl/` -> rename `github.com/disaster37/opensearch/v4` imports to `github.com/disaster37/elasticsearch/v9`; run tests; fix ES-incompatible bits (none expected beyond imports).

**Phase 3 — Core services (each: model + service + unit + integration test)**
8. InfoService, DocumentService, SearchService, AsyncSearchService, ScriptService.
9. IndicesService, ClusterService, NodesService, CatService.
10. IngestService, SnapshotService, SlmService, TasksService, IlmService, DanglingIndicesService, FeaturesService, FleetService, HealthReportService, LicenseService, LogstashService, MigrationService, ShutdownService, GraphService.

**Phase 4 — X-Pack services**
11. TransformService, MlService, SecurityService, SqlService, WatcherService, SslService, XpackService.
12. CcrService, AutoscalingService, EnrichService, EqlService, QueryRulesService, SearchApplicationService, SynonymsService, TextStructureService, InferenceService, SearchableSnapshotsService, StreamsService (if in spec).

**Phase 5 — Completeness gate**
13. Implement `internal/speccheck` + registry; run against pinned spec; fix all gaps (params/methods) until clean.
14. Generate mocks (`go generate ./...`), commit `mock/`.

**Phase 6 — CI**
15. Implement `ci/dagger/main.go` (§9.2–9.5); run locally: `dagger call build`, `dagger call vet`, `dagger call lint`, `dagger call test`, `dagger call test-acceptance`, `dagger call spec-check`.
16. Wire GitHub workflow; verify codecov/git-module steps behave as opensearch.

**Phase 7 — Docs & release prep**
17. README (services table, examples, auth, retry docs), README_RETRY.md, CONTRIBUTING, example_test.go; tag plan (`v9.0.0` after merge).

---

## 11. Risks & open items

| Risk | Mitigation |
|---|---|
| Spec drift between this plan's tables and the pinned 9.x tag | speccheck audit is the gate; spec always wins; discrepancies fixed in code |
| `master_timeout` removal in ES9 (present in opensearch patterns) | verify per-endpoint in Task 0; do not carry over opensearch's master_timeout fields unless in spec |
| Rollup / `_knn_search` absent in ES9 | confirm via spec listing; document N/A in README |
| streams.* / connector.* / inference.* are beta/experimental and may not exist in the pinned tag | conditional services; speccheck allow-list |
| ES container `bootstrap.memory_lock=true` may fail under Dagger runtime | flip to false if startup fails (health wait catches it) |
| Trial license required for watcher/ML/transform acceptance | `xpack.license.self_generated.type=trial` baked into CI container; tests skip gracefully |
| Huge surface -> long review | phased delivery (§10) with green CI per phase |
| goccy/go-json edge cases with ES payloads (e.g. large ints) | unit tests with real canned ES responses from docs |
| resty v2 + HTTP/2 resets behind proxies | DisableHTTP2 + IdleConnTimeout knobs ported from opensearch |

## 12. Definition of done

- `go build ./...`, `go vet ./...`, `golangci-lint run`, `go test ./...` (unit), `go test -tags integration ./...` (against containerized ES 9 with security on) all green in Dagger.
- `internal/speccheck` reports 0 missing endpoints/params against the pinned spec (excluding `_internal.*`).
- Every public method documented; README examples compile (`example_test.go`).
- Mocks generated; CI workflow + codecov + git push-back operational, mirroring opensearch.
