// Package api contains the service interfaces and implementations for the
// Elasticsearch REST API.
//
// Each Elasticsearch API group (documents, search, indices, cluster, etc.) is
// exposed as a service interface with a default implementation backed by a
// resty HTTP client. Services are constructed by the root
// [github.com/disaster37/elasticsearch/v9] package and accessed via the
// [Client] interface.
//
// Request and response types for each service live in the matching *_model.go
// file. Request types carry validate tags (processed by go-playground/validator)
// and Params types provide ToMap() to serialize query parameters.
package api
