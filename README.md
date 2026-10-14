# elasticsearch

[![build](https://github.com/disaster37/elasticsearch/actions/workflows/workflow.yaml/badge.svg)](https://github.com/disaster37/elasticsearch/actions/workflows/workflow.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/disaster37/elasticsearch/v9.svg)](https://pkg.go.dev/github.com/disaster37/elasticsearch/v9)

A type-safe, ergonomic Go client for **Elasticsearch 9.x**. It mirrors the
architecture of the [disaster37/opensearch](https://github.com/disaster37/opensearch)
Go client: a resty-backed HTTP core, fully typed request/response structs for
every endpoint, a fluent query DSL builder, OpenTelemetry tracing middleware,
and a Dagger-based CI pipeline.

## Quick start

```go
import (
    "github.com/disaster37/elasticsearch/v9"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.NewEntry(logrus.StandardLogger())

    client, err := elasticsearch.New(&elasticsearch.Config{
        Addresses:     []string{"https://localhost:9200"},
        Username:      "elastic",
        Password:      "changeme",
        TLSSkipVerify: true, // development only
    }, logger)
    if err != nil {
        log.Fatal(err)
    }

    info, err := client.Info().Info(ctx)
    // ...
}
```

## Connecting to multiple nodes

`Config.Addresses` accepts a list of node URLs. Requests are distributed across
the addresses in round-robin order. When `RetryCount > 0`, a request that fails
against one address is retried against the next address (see
[README_RETRY.md](README_RETRY.md)). The maximum number of addresses is 1000
(`maxAddresses`); exceeding this limit returns an error.

```go
client, err := elasticsearch.New(&elasticsearch.Config{
    Addresses: []string{
        "https://node1:9200",
        "https://node2:9200",
    },
    RetryCount: 3,
}, logger)
```

## Authentication

Exactly one of the following may be set on `Config` (setting more than one
returns an error from `New`):

| Mechanism     | Fields                  | Header sent                              |
|---------------|-------------------------|------------------------------------------|
| Basic auth    | `Username`, `Password`  | `Authorization: Basic <base64>`         |
| API key       | `APIKey`                | `Authorization: ApiKey <value>`         |
| Bearer token  | `BearerToken`           | `Authorization: Bearer <value>`         |

## Services

The client exposes one accessor per Elasticsearch API group:

| Service              | Accessor                | Spec coverage                                   |
|----------------------|-------------------------|-------------------------------------------------|
| Info                 | `client.Info()`         | `info`, `ping`, `capabilities`                  |
| Document             | `client.Document()`     | index, create, get, get_source, exists, exists_source, delete, mget, update, bulk, delete_by_query, update_by_query, reindex, *_rethrottle, explain, termvectors, mtermvectors, terms_enum |
| Search               | `client.Search()`       | search, search_mvt, msearch, search_template, msearch_template, render_search_template, rank_eval, count, scroll, clear_scroll, validate, search_shards, field_caps, open/close_point_in_time, scripts_painless_exec |
| Async Search         | `client.AsyncSearch()`  | submit, get, status, delete                    |
| Script               | `client.Script()`       | get/put/delete_script, get_script_context, get_script_languages |
| Indices              | `client.Indices()`      | ~45 indices.* endpoints (create, delete, get, mappings, settings, aliases, templates, data streams, resolve, ...) |
| Cluster              | `client.Cluster()`      | health, state, stats, settings, allocation_explain, reroute, pending_tasks, remote_info, voting_config_exclusions, component templates |
| Nodes                | `client.Nodes()`        | info, stats, usage, hot_threads, reload_secure_settings |
| Cat                  | `client.Cat()`          | all cat.* endpoints (typed row structs, format=json) |
| Ingest               | `client.Ingest()`       | get/put/delete pipeline, simulate, processor_grok |
| Snapshot             | `client.Snapshot()`     | repository CRUD, verify, analyze, verify_integrity, snapshot CRUD, restore, status, clone, cleanup |
| SLM                  | `client.SLM()`          | lifecycle CRUD, execute, retention, stats, status, start, stop |
| Tasks                | `client.Tasks()`        | list, get, cancel                              |
| ILM                  | `client.ILM()`          | lifecycle CRUD, status, start, stop, explain, move, remove, retry, migrate_to_data_tiers |
| Transform            | `client.Transform()`    | get, stats, put, preview, delete, start, stop, reset, schedule_now, upgrade, node_stats |
| ML                   | `client.ML()`           | jobs, datafeeds, filters, trained models, infer |
| Security             | `client.Security()`     | users, roles, role mappings, privileges, API keys, tokens, authenticate, realm cache |
| SQL                  | `client.SQL()`          | query, translate, clear_cursor, async get/status/delete |
| Watcher              | `client.Watcher()`      | put/get/delete/execute/ack/activate/deactivate watch, query, start, stop, stats |
| CCR                  | `client.CCR()`          | auto-follow patterns, follow, pause/resume/unfollow/forget, info, stats |
| Autoscaling          | `client.Autoscaling()`  | capacity, policy CRUD                           |
| Enrich               | `client.Enrich()`       | policy CRUD, execute, stats                     |
| EQL                  | `client.EQL()`          | search, get, status, delete                    |
| Graph                | `client.Graph()`        | explore                                        |
| License              | `client.License()`      | get, delete, post, basic/trial status & start  |
| Logstash             | `client.Logstash()`     | pipeline CRUD                                  |
| Migration            | `client.Migration()`    | deprecations, feature_upgrade status & run    |
| Shutdown            | `client.Shutdown()`     | node get/put/delete                            |
| Features             | `client.Features()`     | get, reset                                     |
| Fleet                | `client.Fleet()`        | global_checkpoints                             |
| Dangling Indices     | `client.DanglingIndices()` | list, import, delete                       |
| Health Report        | `client.HealthReport()` | get                                            |
| Query Rules          | `client.QueryRules()`    | ruleset CRUD, test                            |
| Search Application  | `client.SearchApplication()` | list, get, put, delete, search, behavioral analytics |
| Synonyms             | `client.Synonyms()`     | synonyms sets, synonym rules CRUD             |
| Text Structure       | `client.TextStructure()` | find_structure, field/message_structure, test_grok_pattern |
| Inference            | `client.Inference()`    | get, put, delete, update, inference, completion, rerank |
| Searchable Snapshots | `client.SearchableSnapshots()` | mount, stats, cache_stats, clear_cache, repository_stats |
| SSL                  | `client.SSL()`          | certificates                                   |
| XPack                | `client.XPack()`        | info, usage                                    |
| Streams              | `client.Streams()`      | logs enable/disable, status (beta)            |

### N/A in Elasticsearch 9

The following endpoints were removed upstream and are intentionally absent:

- **Rollup** — removed in ES 9.
- **`_knn_search`** — removed in ES 9 (kNN is now part of `_search`).

## Query DSL

Use the `querydsl` package to build type-safe queries:

```go
import "github.com/disaster37/elasticsearch/v9/querydsl"

q := querydsl.NewBoolQuery().
    Must(querydsl.NewMatchQuery("title", "elasticsearch")).
    Filter(querydsl.NewRangeQuery("date").Gte("2024-01-01")).
    MinimumShouldMatch("1")
```

## Error handling

```go
if elasticsearch.IsNotFound(err) {
    // 404
} else if elasticsearch.IsConflict(err) {
    // 409
} else if elasticsearch.IsUnauthorized(err) {
    // 401
}
```

See [README_RETRY.md](README_RETRY.md) for retry behavior.

## CI

CI runs via a Dagger module (`.github/workflows/workflow.yaml` → `dagger call`):
build, vet, lint, format, speccheck, unit tests, acceptance tests against a
containerized Elasticsearch 9 with security enabled, CodeCov upload, and git
push-back of `coverage.out`.

The `internal/speccheck` audit ratchets the pinned 9.x REST spec against the
client. Endpoints are either verified registered implementations or exact,
individually reviewed missing-endpoint debt entries, so newly added endpoints
always fail the audit. Every query parameter and path selector on a registered
endpoint must be covered by a real carrier field, a positional query argument,
or a narrowly controlled exception; the audit performs live `ToMap()`
serialization checks to catch removed wire keys. See
[internal/speccheck/README.md](internal/speccheck/README.md) for the pinned
revision, the inventory and edit matrix, generator commands, and the
opt-in extension APIs (`As<Group>ServiceSpecExtensions`).

## License

MIT, following the reference client.
