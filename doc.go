// Package elasticsearch provides a Go client for the Elasticsearch REST API.
//
// This client provides a type-safe, ergonomic interface for interacting with
// Elasticsearch 9.x clusters. It supports all major Elasticsearch features
// including document CRUD, full-text search, aggregations, index lifecycle
// management, security, machine learning, SQL, and many more.
//
// # Quick Start
//
// Create a client:
//
//	import (
//	    "github.com/disaster37/elasticsearch/v9"
//	    "github.com/sirupsen/logrus"
//	)
//
//	func main() {
//	    logger := logrus.NewEntry(logrus.StandardLogger())
//
//	    client, err := elasticsearch.New(&elasticsearch.Config{
//	        Addresses:     []string{"https://localhost:9200"},
//	        Username:      "elastic",
//	        Password:      "changeme",
//	        TLSSkipVerify: true, // For development only
//	    }, logger)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Use the client to interact with Elasticsearch
//	    // ...
//	}
//
// Config.Addresses accepts a list of node URLs; requests are distributed across
// them in round-robin order. Configuring more than one address automatically
// enables retry and failover: RetryCount defaults to the number of addresses
// and the default retry conditions (network errors, 429, and 5xx except 501)
// are applied, so every node is tried before a request gives up. Explicit
// RetryCount/RetryConditions values always win. There is currently no opt-out
// field for multi-address configurations; use a single address for strict
// no-retry semantics. The maximum number of addresses is 1000 (maxAddresses);
// exceeding this limit returns an error.
//
// # Authentication
//
// The client supports three mutually-exclusive authentication mechanisms, set
// via [Config]:
//
//   - Basic auth (Username + Password)
//   - API key (APIKey) — sent as "Authorization: ApiKey <value>"
//   - Bearer token (BearerToken) — sent as "Authorization: Bearer <value>"
//
// Setting more than one mechanism returns an error from [New].
//
// # Package Organization
//
// The client is organized into several sub-packages:
//
//   - types: Common types and error handling
//   - api: Service interfaces for Elasticsearch REST APIs
//   - querydsl: Query and aggregation builder DSL
//   - trace/opentelemetry: OpenTelemetry tracing middleware
//
// Most users will only need to import the root package. The sub-packages are
// automatically re-exported for convenience.
//
// # Error Handling
//
// All operations can return an [*ElasticsearchError]. Use the helper functions
// to check for specific error conditions:
//
//	if elasticsearch.IsNotFound(err) {
//	    log.Println("resource not found")
//	} else if elasticsearch.IsConflict(err) {
//	    log.Println("version conflict")
//	} else if elasticsearch.IsUnauthorized(err) {
//	    log.Println("unauthorized")
//	}
package elasticsearch
