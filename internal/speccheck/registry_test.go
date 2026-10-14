package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/disaster37/elasticsearch/v9/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistryFields asserts that every ParamsStruct referenced by the registry
// actually exists and exposes the listed ParamsFields. This gives compile-time
// safety that the registry stays in sync with the code.
func TestRegistryFields(t *testing.T) {
	for name, impl := range Registry {
		if impl.ParamsStruct == "" {
			continue
		}
		typ := resolveType(impl.ParamsStruct)
		if typ == nil {
			t.Errorf("%s: unknown ParamsStruct %q", name, impl.ParamsStruct)
			continue
		}
		for _, field := range impl.ParamsFields {
			if _, ok := typ.FieldByName(field); !ok {
				t.Errorf("%s: struct %q has no field %q", name, impl.ParamsStruct, field)
			}
		}
	}
}

// TestRegistryLookupKnownEndpoints verifies that several known spec endpoints
// are present in the registry with the expected service and method.
func TestRegistryLookupKnownEndpoints(t *testing.T) {
	cases := []struct {
		name    string
		service string
		method  string
	}{
		{"index", "DocumentService", "Index"},
		{"create", "DocumentService", "Create"},
		{"get", "DocumentService", "Get"},
		{"delete", "DocumentService", "Delete"},
		{"info", "InfoService", "Info"},
		{"ping", "InfoService", "Ping"},
		{"search", "SearchService", "Search"},
		{"count", "SearchService", "Count"},
		{"cluster.health", "ClusterService", "Health"},
		{"indices.create", "IndicesService", "Create"},
		{"indices.delete", "IndicesService", "Delete"},
	}
	for _, c := range cases {
		impl, ok := Registry[c.name]
		require.True(t, ok, "endpoint %q should be in registry", c.name)
		assert.Equal(t, c.service, impl.Service, "endpoint %q service", c.name)
		assert.Equal(t, c.method, impl.Method, "endpoint %q method", c.name)
	}
}

// TestRegistryMissingEndpoint verifies that an endpoint absent from the
// registry is reported as missing (the speccheck CLI uses this to find gaps).
func TestRegistryMissingEndpoint(t *testing.T) {
	_, ok := Registry["nonexistent.endpoint"]
	assert.False(t, ok, "unknown endpoint should not be in registry")
}

// TestRegistryNoParamsEndpoints verifies endpoints without query params have
// an empty ParamsStruct.
func TestRegistryNoParamsEndpoints(t *testing.T) {
	for _, name := range []string{"info", "ping"} {
		impl, ok := Registry[name]
		require.True(t, ok)
		assert.Empty(t, impl.ParamsStruct, "%s should have no ParamsStruct", name)
		assert.Empty(t, impl.ParamsFields, "%s should have no ParamsFields", name)
	}
}

// TestRegistryPathParamValidation verifies that endpoints with path parts list
// them correctly.
func TestRegistryPathParamValidation(t *testing.T) {
	cases := []struct {
		name      string
		pathParts []string
	}{
		{"index", []string{"index", "id"}},
		{"create", []string{"index", "id"}},
		{"get", []string{"index", "id"}},
		{"delete", []string{"index", "id"}},
		{"indices.create", []string{"index"}},
		{"indices.delete", []string{"index"}},
	}
	for _, c := range cases {
		impl, ok := Registry[c.name]
		require.True(t, ok, c.name)
		assert.Equal(t, c.pathParts, impl.PathParts, "%s path parts", c.name)
	}
}

// TestRegistryParamFieldValidation verifies that the ParamsFields listed for
// each endpoint actually exist on the corresponding struct (deeper check than
// TestRegistryFields, asserting specific known fields).
func TestRegistryParamFieldValidation(t *testing.T) {
	impl, ok := Registry["index"]
	require.True(t, ok)
	typ := resolveType(impl.ParamsStruct)
	require.NotNil(t, typ)
	// Spot-check a few fields.
	for _, field := range []string{"OpType", "Refresh", "Pipeline", "Timeout"} {
		_, ok := typ.FieldByName(field)
		assert.True(t, ok, "IndexParams should have field %q", field)
	}
}

// TestResolveTypeKnown verifies resolveType returns the expected reflect.Type
// for known Params structs.
func TestResolveTypeKnown(t *testing.T) {
	cases := []struct {
		qualified string
		expected  reflect.Type
	}{
		{"api.IndexParams", reflect.TypeOf(api.IndexParams{})},
		{"api.CreateParams", reflect.TypeOf(api.CreateParams{})},
		{"api.GetParams", reflect.TypeOf(api.GetParams{})},
		{"api.DeleteParams", reflect.TypeOf(api.DeleteParams{})},
		{"api.SearchParams", reflect.TypeOf(api.SearchParams{})},
		{"api.CountParams", reflect.TypeOf(api.CountParams{})},
		{"api.ClusterHealthParams", reflect.TypeOf(api.ClusterHealthParams{})},
		{"api.IndicesCreateParams", reflect.TypeOf(api.IndicesCreateParams{})},
		{"api.IndicesDeleteParams", reflect.TypeOf(api.IndicesDeleteParams{})},
	}
	for _, c := range cases {
		assert.Equal(t, c.expected, resolveType(c.qualified), c.qualified)
	}
}

// TestResolveTypeUnknown verifies resolveType returns nil for unknown names.
func TestResolveTypeUnknown(t *testing.T) {
	assert.Nil(t, resolveType("api.NonexistentParams"))
	assert.Nil(t, resolveType("other.IndexParams"))
	assert.Nil(t, resolveType(""))
}

// TestContains verifies the contains helper.
func TestContains(t *testing.T) {
	s := []string{"a", "b", "c"}
	assert.True(t, contains(s, "a"))
	assert.True(t, contains(s, "c"))
	assert.False(t, contains(s, "d"))
	assert.False(t, contains(nil, "a"))
	assert.False(t, contains([]string{}, "a"))
}

// TestSnakeToCamel verifies the snakeToCamel helper converts snake_case to
// CamelCase.
func TestSnakeToCamel(t *testing.T) {
	cases := map[string]string{
		"if_primary_term":        "IfPrimaryTerm",
		"if_seq_no":              "IfSeqNo",
		"op_type":                "OpType",
		"wait_for_active_shards": "WaitForActiveShards",
		"version_type":           "VersionType",
		"force_synthetic_source": "ForceSyntheticSource",
		"single":                 "Single",
		"":                       "",
		"_leading":               "Leading",
		"trailing_":              "Trailing",
	}
	for in, want := range cases {
		assert.Equal(t, want, snakeToCamel(in), in)
	}
}

// TestCollectParts verifies collectParts deduplicates and sorts path parts
// across multiple path entries.
func TestCollectParts(t *testing.T) {
	paths := []struct {
		Path  string   `json:"path"`
		Parts []string `json:"parts"`
	}{
		{Path: "/{index}/_doc/{id}", Parts: []string{"index", "id"}},
		{Path: "/{index}/_search", Parts: []string{"index"}},
		{Path: "/_all/_search", Parts: nil},
	}
	got := collectParts(paths)
	assert.Equal(t, []string{"id", "index"}, got)
}

// TestCollectPartsEmpty verifies collectParts returns an empty slice for no
// path parts.
func TestCollectPartsEmpty(t *testing.T) {
	paths := []struct {
		Path  string   `json:"path"`
		Parts []string `json:"parts"`
	}{
		{Path: "/_search", Parts: nil},
	}
	got := collectParts(paths)
	assert.Empty(t, got)
}

// TestCollectPartsDedup verifies collectParts deduplicates repeated parts.
func TestCollectPartsDedup(t *testing.T) {
	paths := []struct {
		Path  string   `json:"path"`
		Parts []string `json:"parts"`
	}{
		{Parts: []string{"index", "id"}},
		{Parts: []string{"id", "index", "type"}},
	}
	got := collectParts(paths)
	assert.Equal(t, []string{"id", "index", "type"}, got)
}

// TestRunEmptySpecDir verifies run returns exit code 2 when no spec dir is
// provided.
func TestRunEmptySpecDir(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run("", &stdout, &stderr)
	assert.Equal(t, 2, code)
	assert.Contains(t, stderr.String(), "-spec is required")
}

// TestRunNonexistentDir verifies run returns 0 (no files → no gaps) for a
// nonexistent/empty glob.
func TestRunNonexistentDir(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(filepath.Join(t.TempDir(), "nope"), &stdout, &stderr)
	// Glob returns no error and no files for a nonexistent dir.
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "scanned 0 endpoints")
	assert.Contains(t, stdout.String(), "no gaps found")
}

// TestRunKnownEndpointNoGaps verifies run reports no gaps for a spec file that
// matches a registry entry with all params/paths present.
func TestRunKnownEndpointNoGaps(t *testing.T) {
	dir := t.TempDir()
	// "info" is in the registry with no params and no path parts.
	spec := `{"paths":[{"path":"/","parts":[]}],"params":{},"visibility":"public"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "info.json"), []byte(spec), 0644))

	var stdout, stderr bytes.Buffer
	code := run(dir, &stdout, &stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "scanned 1 endpoints")
	assert.Contains(t, stdout.String(), "no gaps found")
}

// TestRunMissingEndpoint verifies run reports a missing endpoint.
func TestRunMissingEndpoint(t *testing.T) {
	dir := t.TempDir()
	spec := `{"paths":[{"path":"/_custom","parts":[]}],"params":{},"visibility":"public"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "custom.endpoint.json"), []byte(spec), 0644))

	var stdout, stderr bytes.Buffer
	code := run(dir, &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stdout.String(), "MISSING ENDPOINTS (1)")
	assert.Contains(t, stdout.String(), "custom.endpoint")
}

// TestRunParamGap verifies run reports a param gap when a spec param is not on
// the registered Params struct.
func TestRunParamGap(t *testing.T) {
	dir := t.TempDir()
	// "info" has no ParamsStruct, so no param gap is reported. Use "index"
	// which has ParamsFields; add a param not in the list.
	spec := `{"paths":[{"path":"/{index}/_doc/{id}","parts":["index","id"]}],"params":{"nonexistent_param":{"type":"string"}},"visibility":"public"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.json"), []byte(spec), 0644))

	var stdout, stderr bytes.Buffer
	code := run(dir, &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stdout.String(), "PARAM GAPS (1)")
	assert.Contains(t, stdout.String(), "nonexistent_param")
}

// TestRunPathPartGap verifies run reports a path-part gap.
func TestRunPathPartGap(t *testing.T) {
	dir := t.TempDir()
	// "info" has no PathParts; add a path part to trigger a gap.
	spec := `{"paths":[{"path":"/{node}","parts":["node"]}],"params":{},"visibility":"public"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "info.json"), []byte(spec), 0644))

	var stdout, stderr bytes.Buffer
	code := run(dir, &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stdout.String(), "PATH PART GAPS (1)")
	assert.Contains(t, stdout.String(), "node")
}

// TestRunSkipsInternal verifies run skips _internal.* endpoints.
func TestRunSkipsInternal(t *testing.T) {
	dir := t.TempDir()
	spec := `{"paths":[],"params":{},"visibility":"public"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "_internal.test.json"), []byte(spec), 0644))

	var stdout, stderr bytes.Buffer
	code := run(dir, &stdout, &stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "scanned 0 endpoints")
}

// TestRunSkipsNonPublic verifies run skips non-public endpoints (unless
// allowlisted).
func TestRunSkipsNonPublic(t *testing.T) {
	dir := t.TempDir()
	spec := `{"paths":[],"params":{},"visibility":"private"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "private.endpoint.json"), []byte(spec), 0644))

	var stdout, stderr bytes.Buffer
	code := run(dir, &stdout, &stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "scanned 0 endpoints")
}

// TestRunFeatureFlagAllowlist verifies run checks allowlisted feature-flag
// endpoints even when visibility != public.
func TestRunFeatureFlagAllowlist(t *testing.T) {
	dir := t.TempDir()
	// streams.status is allowlisted but not in the registry → missing endpoint.
	spec := `{"paths":[{"path":"/_streams/status","parts":[]}],"params":{},"visibility":"feature_flag"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "streams.status.json"), []byte(spec), 0644))

	var stdout, stderr bytes.Buffer
	code := run(dir, &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stdout.String(), "scanned 1 endpoints")
	assert.Contains(t, stdout.String(), "MISSING ENDPOINTS (1)")
	assert.Contains(t, stdout.String(), "streams.status")
}

// TestRunMalformedJSON verifies run tolerates a malformed spec file (logs to
// stderr, continues).
func TestRunMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "info.json"), []byte(`{not valid`), 0644))

	var stdout, stderr bytes.Buffer
	code := run(dir, &stdout, &stderr)
	// info.json failed to parse → 0 endpoints scanned → no gaps → exit 0.
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr.String(), "parse")
}

// TestRunUnreadableFile verifies run logs a read error and continues. This is
// hard to trigger portably, so we skip if the file is readable.
func TestRunUnreadableFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("cannot test unreadable files as root")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "info.json")
	require.NoError(t, os.WriteFile(path, []byte(`{}`), 0644))
	require.NoError(t, os.Chmod(path, 0000))
	defer func() { _ = os.Chmod(path, 0644) }()

	var stdout, stderr bytes.Buffer
	code := run(dir, &stdout, &stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr.String(), "read")
}

// resolveType maps a fully-qualified struct name like "api.IndexParams" to its
// reflect.Type. Only the api package is supported (where all Params live).
func resolveType(qualified string) reflect.Type {
	switch {
	case strings.HasPrefix(qualified, "api."):
		name := strings.TrimPrefix(qualified, "api.")
		switch name {
		case "IndexParams":
			return reflect.TypeOf(api.IndexParams{})
		case "CreateParams":
			return reflect.TypeOf(api.CreateParams{})
		case "GetParams":
			return reflect.TypeOf(api.GetParams{})
		case "DeleteParams":
			return reflect.TypeOf(api.DeleteParams{})
		case "SearchParams":
			return reflect.TypeOf(api.SearchParams{})
		case "CountParams":
			return reflect.TypeOf(api.CountParams{})
		case "ClusterHealthParams":
			return reflect.TypeOf(api.ClusterHealthParams{})
		case "IndicesCreateParams":
			return reflect.TypeOf(api.IndicesCreateParams{})
		case "IndicesDeleteParams":
			return reflect.TypeOf(api.IndicesDeleteParams{})
		}
	}
	return nil
}
