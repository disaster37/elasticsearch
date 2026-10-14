package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/disaster37/elasticsearch/v9/internal/speccheck/spec"
)

// InventoryRow is one machine-readable row emitted by -inventory.
type InventoryRow struct {
	Name             string
	Source           string
	Visibility       string
	Stability        string
	CandidateService string
	CandidateMethod  string
	ParamsStruct     string
	Params           []string
	PathParts        []string
	MissingFields    []string
	UnsupportedParts []string
	Classification   string
	Evidence         string
}

// inventory classifies every checked endpoint as either a registered
// implementation or a missing endpoint with a heuristic candidate. Candidates
// are explicitly unverified: a matching method name is not proof of support.
func inventory(catalog spec.Catalog, registry map[string]EndpointImpl) ([]InventoryRow, error) {
	names := make([]string, 0, len(catalog.Endpoints))
	for name := range catalog.Endpoints {
		names = append(names, name)
	}
	sort.Strings(names)

	allow := defaultPolicy().FeatureFlagAllowlist
	rows := make([]InventoryRow, 0, len(names))
	for _, name := range names {
		ep := catalog.Endpoints[name]
		if !spec.Checked(ep, allow) {
			continue
		}
		row := InventoryRow{
			Name:       name,
			Source:     ep.Source,
			Visibility: ep.Endpoint.Visibility,
			Stability:  ep.Endpoint.Stability,
			Params:     sortedParamNames(ep.Endpoint.Params),
			PathParts:  spec.CollectParts(ep.Endpoint.URL.Paths),
		}
		if impl, ok := registry[name]; ok {
			row.Classification = "registered"
			row.CandidateService = impl.Service
			row.CandidateMethod = impl.Method
			row.ParamsStruct = impl.ParamsStruct
			row.Evidence = impl.Evidence
			row.MissingFields = missingRegisteredFields(impl, ep)
			row.UnsupportedParts = missingRegisteredParts(impl, ep)
			rows = append(rows, row)
			continue
		}
		row.Classification = "missing"
		svc, method := candidateBinding(name)
		row.CandidateService = svc
		row.CandidateMethod = method
		if svc == "" {
			row.Evidence = "no service maps to this namespace"
		} else {
			row.Evidence = fmt.Sprintf("heuristic candidate %s.%s (unverified)", svc, method)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func missingRegisteredFields(impl EndpointImpl, ep spec.NamedEndpoint) []string {
	var out []string
	for param := range ep.Endpoint.Params {
		if _, ok := impl.QueryArgs[param]; ok {
			continue
		}
		if _, ok := impl.QueryFields[param]; ok {
			continue
		}
		if contains(impl.ParamsFields, queryField(impl, param)) {
			continue
		}
		out = append(out, param)
	}
	sort.Strings(out)
	return out
}

func missingRegisteredParts(impl EndpointImpl, ep spec.NamedEndpoint) []string {
	var out []string
	for _, part := range spec.CollectParts(ep.Endpoint.URL.Paths) {
		if _, ok := impl.PathSources[part]; !ok {
			out = append(out, part)
		}
	}
	sort.Strings(out)
	return out
}

func sortedParamNames(params map[string]spec.Param) []string {
	out := make([]string, 0, len(params))
	for name := range params {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

var inventoryNamespaceService = map[string]string{
	"async_search": "AsyncSearchService", "autoscaling": "AutoscalingService",
	"cat": "CatService", "ccr": "CcrService", "cluster": "ClusterService",
	"dangling_indices": "DanglingIndicesService", "enrich": "EnrichService",
	"eql": "EqlService", "features": "FeaturesService", "fleet": "FleetService",
	"graph": "GraphService", "health_report": "HealthReportService", "ilm": "IlmService",
	"indices": "IndicesService", "inference": "InferenceService", "ingest": "IngestService",
	"license": "LicenseService", "logstash": "LogstashService", "migration": "MigrationService",
	"ml": "MlService", "nodes": "NodesService", "query_rules": "QueryRulesService",
	"search_application": "SearchApplicationService", "searchable_snapshots": "SearchableSnapshotsService",
	"security": "SecurityService", "shutdown": "ShutdownService", "slm": "SlmService",
	"snapshot": "SnapshotService", "sql": "SqlService", "ssl": "SslService",
	"streams": "StreamsService", "synonyms": "SynonymsService", "tasks": "TasksService",
	"text_structure": "TextStructureService", "transform": "TransformService",
	"watcher": "WatcherService", "xpack": "XpackService",
}

var inventoryRootService = map[string]string{
	"index": "DocumentService", "create": "DocumentService", "get": "DocumentService",
	"get_source": "DocumentService", "exists": "DocumentService", "exists_source": "DocumentService",
	"delete": "DocumentService", "mget": "DocumentService", "update": "DocumentService",
	"bulk": "DocumentService", "delete_by_query": "DocumentService",
	"delete_by_query_rethrottle": "DocumentService", "update_by_query": "DocumentService",
	"update_by_query_rethrottle": "DocumentService", "reindex": "DocumentService",
	"reindex_rethrottle": "DocumentService", "explain": "DocumentService",
	"termvectors": "DocumentService", "mtermvectors": "DocumentService",
	"terms_enum": "DocumentService",
	"search":     "SearchService", "search_mvt": "SearchService", "msearch": "SearchService",
	"msearch_template": "SearchService", "search_template": "SearchService",
	"render_search_template": "SearchService", "rank_eval": "SearchService",
	"count": "SearchService", "scroll": "SearchService", "clear_scroll": "SearchService",
	"validate_query": "SearchService", "search_shards": "SearchService",
	"field_caps": "SearchService", "open_point_in_time": "SearchService",
	"close_point_in_time": "SearchService", "scripts_painless_execute": "SearchService",
	"info": "InfoService", "ping": "InfoService", "capabilities": "InfoService",
	"health_report": "HealthReportService",
	"get_script":    "ScriptService", "put_script": "ScriptService",
	"delete_script": "ScriptService", "get_script_context": "ScriptService",
	"get_script_languages": "ScriptService",
}

// candidateBinding is a best-effort namespace/method guess used only for
// inventory reporting; it never registers support.
func candidateBinding(name string) (service, method string) {
	leaf := name
	if i := strings.Index(name, "."); i >= 0 {
		leaf = name[i+1:]
	}
	ns := ""
	if i := strings.Index(name, "."); i >= 0 {
		ns = name[:i]
	}
	if svc, ok := inventoryNamespaceService[ns]; ok {
		return svc, snakeToCamel(leaf)
	}
	if svc, ok := inventoryRootService[name]; ok {
		return svc, snakeToCamel(leaf)
	}
	return "", ""
}
