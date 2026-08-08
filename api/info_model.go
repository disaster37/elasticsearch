package api

// InfoResponse represents the cluster info returned by the root endpoint.
type InfoResponse struct {
	Name        string       `json:"name"`
	ClusterName string       `json:"cluster_name"`
	ClusterUUID string       `json:"cluster_uuid"`
	Version     *InfoVersion `json:"version,omitempty"`
	Tagline     string       `json:"tagline"`
}

// InfoVersion contains the Elasticsearch version details.
type InfoVersion struct {
	Number             string `json:"number"`
	BuildFlavor        string `json:"build_flavor,omitempty"`
	BuildType          string `json:"build_type,omitempty"`
	BuildHash          string `json:"build_hash,omitempty"`
	BuildDate          string `json:"build_date,omitempty"`
	BuildSnapshot      bool   `json:"build_snapshot,omitempty"`
	LuceneVersion      string `json:"lucene_version,omitempty"`
	MinimumWireCompat  string `json:"minimum_wire_compatibility_version,omitempty"`
	MinimumIndexCompat string `json:"minimum_index_compatibility_version,omitempty"`
}

// CapabilitiesRequest is the request for GET /_capabilities.
type CapabilitiesRequest struct {
	Method     string
	Path       string
	Parameters []string
	Local      *bool
	Common     *CommonParams
}

// CapabilitiesResponse is the response from GET /_capabilities.
type CapabilitiesResponse struct {
	Capabilities map[string]any `json:"capabilities,omitempty"`
}
