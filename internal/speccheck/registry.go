// Command speccheck verifies that the Elasticsearch Go client covers every
// endpoint in the Elasticsearch 9.x REST API spec (excluding _internal.* and
// private-visibility endpoints).
//
// The registry maps each spec endpoint name to its implementation: the service
// interface method and the Params struct that carries its query parameters.
// A unit test uses reflection to assert that every listed Params field actually
// exists on the named struct, giving compile-time-ish safety.
package main

// EndpointImpl describes the client implementation of a single spec endpoint.
type EndpointImpl struct {
	// Service is the service interface name (e.g. "DocumentService").
	Service string
	// Method is the method name on the service interface (e.g. "Index").
	Method string
	// ParamsStruct is the fully-qualified Params struct type name, or "" when
	// the endpoint takes no query parameters.
	ParamsStruct string
	// ParamsFields lists the spec query-parameter names the struct must expose
	// (as Go fields, in CamelCase after snake_case conversion). Empty when the
	// endpoint has no params.
	ParamsFields []string
	// PathParts lists the path-part names the endpoint requires.
	PathParts []string
}

// Registry maps spec endpoint names to their implementation. It is
// hand-maintained alongside the services and verified by TestRegistryFields.
var Registry = map[string]EndpointImpl{
	"index": {
		Service:      "DocumentService",
		Method:       "Index",
		ParamsStruct: "api.IndexParams",
		ParamsFields: []string{"IfPrimaryTerm", "IfSeqNo", "OpType", "Pipeline", "Refresh", "RequireAlias", "Timeout", "Version", "VersionType", "WaitForActiveShards"},
		PathParts:    []string{"index", "id"},
	},
	"create": {
		Service:      "DocumentService",
		Method:       "Create",
		ParamsStruct: "api.CreateParams",
		ParamsFields: []string{"IfPrimaryTerm", "IfSeqNo", "Pipeline", "Refresh", "RequireAlias", "Timeout", "Version", "VersionType", "WaitForActiveShards"},
		PathParts:    []string{"index", "id"},
	},
	"get": {
		Service:      "DocumentService",
		Method:       "Get",
		ParamsStruct: "api.GetParams",
		ParamsFields: []string{"ForceSyntheticSource", "Preference", "Realtime", "Refresh", "Routing", "Source", "SourceExcludes", "SourceIncludes", "StoredFields", "Version", "VersionType"},
		PathParts:    []string{"index", "id"},
	},
	"delete": {
		Service:      "DocumentService",
		Method:       "Delete",
		ParamsStruct: "api.DeleteParams",
		ParamsFields: []string{"IfPrimaryTerm", "IfSeqNo", "Refresh", "Routing", "Timeout", "Version", "VersionType"},
		PathParts:    []string{"index", "id"},
	},
	"info": {
		Service: "InfoService",
		Method:  "Info",
	},
	"ping": {
		Service: "InfoService",
		Method:  "Ping",
	},
	"search": {
		Service:      "SearchService",
		Method:       "Search",
		ParamsStruct: "api.SearchParams",
		PathParts:    []string{"index"},
	},
	"count": {
		Service:      "SearchService",
		Method:       "Count",
		ParamsStruct: "api.CountParams",
		PathParts:    []string{"index"},
	},
	"cluster.health": {
		Service:      "ClusterService",
		Method:       "Health",
		ParamsStruct: "api.ClusterHealthParams",
		PathParts:    []string{"index"},
	},
	"indices.create": {
		Service:      "IndicesService",
		Method:       "Create",
		ParamsStruct: "api.IndicesCreateParams",
		PathParts:    []string{"index"},
	},
	"indices.delete": {
		Service:      "IndicesService",
		Method:       "Delete",
		ParamsStruct: "api.IndicesDeleteParams",
		PathParts:    []string{"index"},
	},
	// The registry is intentionally non-exhaustive in this scaffold; the CI
	// speccheck step reports any spec endpoint missing from the registry so it
	// can be filled in. Endpoints removed in ES 9 (rollup, _knn_search) are
	// documented as N/A in the README and excluded here.
}
