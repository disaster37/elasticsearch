// Command speccheck verifies that the Elasticsearch Go client covers the
// endpoints, query parameters, and path selectors of a pinned Elasticsearch
// REST API spec directory.
//
// The registry is generated from hand-reviewed bindings (bindings.json) plus a
// static scan of the api package; see internal/speccheck/README.md. Generated
// resolver tables live in types_generated.go and the registry literal lives in
// registry_generated.go.
package main

import (
	"reflect"
	"strings"

	"github.com/disaster37/elasticsearch/v9/internal/speccheck/generator"
)

//go:generate go run generate.go -api ../../api -bindings bindings.json -out .

// ValueSource identifies where a method reads a query value or path selector.
type ValueSource = generator.ValueSource

// EndpointBinding is one reviewed endpoint-to-code binding.
type EndpointBinding = generator.EndpointBinding

// EndpointImpl is the resolved implementation of a single spec endpoint.
type EndpointImpl struct {
	// Service is the api service interface name, including an extension
	// interface when the new API lives there.
	Service string
	// Method is the method name on the service interface.
	Method string
	// ParamsStruct is the qualified api params type, or "" when the endpoint
	// has no carrier.
	ParamsStruct string
	// ParamsFields is every accessible exported field of the carrier, sorted.
	ParamsFields []string
	// ParamsSource is the argument or request field that carries the params.
	ParamsSource *ValueSource
	// QueryArgs maps exact wire names supplied positionally as arguments.
	QueryArgs map[string]ValueSource
	// QueryFields overrides the default snake_case-to-Go field mapping for
	// genuine initialism mismatches.
	QueryFields map[string]string
	// PathParts is the sorted list of spec path parts mapped by PathSources.
	PathParts []string
	// PathSources maps exact spec part names to caller-supplied data.
	PathSources map[string]ValueSource
	// Evidence records the source method/helper backing nontrivial aliases.
	Evidence string
	// Conditional marks supported APIs absent from the pinned catalog.
	Conditional bool
}

// resolveType returns the api params type for a qualified name, or nil.
func resolveType(qualified string) reflect.Type {
	return paramsTypes[qualified]
}

// resolveService returns the api service interface type for a name, or nil.
// The name may be given qualified ("api.FooService") or bare.
func resolveService(name string) reflect.Type {
	if !strings.Contains(name, ".") {
		name = "api." + name
	}
	return serviceTypes[name]
}
