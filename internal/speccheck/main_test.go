package main

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/disaster37/elasticsearch/v9/internal/speccheck/generator"
	"github.com/disaster37/elasticsearch/v9/internal/speccheck/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSpec(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
}

const infoSpec = `{"info":{"url":{"paths":[{"path":"/","methods":["GET"]}]},"params":{},"stability":"stable","visibility":"public"}}`

func loadCatalog(t *testing.T, dir string) spec.Catalog {
	t.Helper()
	cat, err := spec.LoadDir(dir)
	require.NoError(t, err)
	return cat
}

// TestGeneratedFieldsMatchReflection guards that the AST-derived generated
// fields equal reflection's visible exported fields for every declared api
// Params type, including the catCommonParams promoted-field regression.
func TestGeneratedFieldsMatchReflection(t *testing.T) {
	model, err := generator.ScanAPI("../../api")
	require.NoError(t, err)
	require.NotEmpty(t, paramsTypes)
	for qualified, typ := range paramsTypes {
		name := strings.TrimPrefix(qualified, "api.")
		want := exportedFields(typ)
		sort.Strings(want)
		assert.Equal(t, want, model.AccessibleFields(name), "fields mismatch for %s", qualified)
	}
}

func testRegistry(names ...string) map[string]EndpointImpl {
	reg := map[string]EndpointImpl{}
	for _, name := range names {
		reg[name] = Registry[name]
	}
	return reg
}

func TestAuditRegisteredNoDrift(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "info.json", infoSpec)
	writeSpec(t, dir, "search.json", `{"search":{"url":{"paths":[{"path":"/{index}/_search","methods":["GET"],"parts":{"index":{"type":"list"}}}]},"params":{},"visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	res := audit(cat, testRegistry("info", "search"), Policy{FeatureFlagAllowlist: map[string]bool{}})
	assert.Empty(t, res.MissingEndpoints)
	assert.Empty(t, res.ParamGaps)
	assert.Empty(t, res.PathGaps)
	assert.Empty(t, res.IntegrityErrors)
	assert.Empty(t, res.PolicyErrors)
	assert.Equal(t, 2, res.Implemented)
}

func TestAuditNewEndpointFails(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "fictional.json", `{"fictional.endpoint":{"url":{"paths":[{"path":"/_fictional","methods":["GET"]}]},"params":{},"visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	res := audit(cat, map[string]EndpointImpl{}, Policy{FeatureFlagAllowlist: map[string]bool{}})
	require.Contains(t, res.MissingEndpoints, "fictional.endpoint")
	assert.Equal(t, 1, exitCode(res))
}

func TestAuditKnownMissingPasses(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "x.json", `{"x.thing":{"url":{"paths":[{"path":"/_x","methods":["GET"]}]},"params":{},"stability":"stable","visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	policy := Policy{
		KnownMissingEndpoints: map[string]MissingEndpoint{
			"x.thing": {Status: "not-implemented", Reason: "no service", Stability: "stable", Visibility: "public", Tracking: "u"},
		},
		FeatureFlagAllowlist: map[string]bool{},
	}
	res := audit(cat, map[string]EndpointImpl{}, policy)
	require.Empty(t, res.MissingEndpoints)
	assert.Equal(t, 1, res.KnownMissing)
	assert.Equal(t, 0, exitCode(res))
}

func TestAuditPolicyViolations(t *testing.T) {
	registered := map[string]EndpointImpl{
		"info": {Service: "InfoService", Method: "Info"},
	}
	newCatalog := func(t *testing.T, body string) spec.Catalog {
		dir := t.TempDir()
		writeSpec(t, dir, "e.json", body)
		return loadCatalog(t, dir)
	}
	cat := newCatalog(t, infoSpec)

	t.Run("stale baseline", func(t *testing.T) {
		policy := Policy{KnownMissingEndpoints: map[string]MissingEndpoint{"ghost": {Status: "not-implemented", Reason: "r", Tracking: "t"}}, FeatureFlagAllowlist: map[string]bool{}}
		res := audit(cat, registered, policy)
		assert.NotEmpty(t, res.PolicyErrors)
	})
	t.Run("baseline overlap", func(t *testing.T) {
		policy := Policy{KnownMissingEndpoints: map[string]MissingEndpoint{"info": {Status: "not-implemented", Reason: "r", Tracking: "t"}}, FeatureFlagAllowlist: map[string]bool{}}
		res := audit(cat, registered, policy)
		assert.Contains(t, strings.Join(res.PolicyErrors, "\n"), "also registered")
	})
	t.Run("metadata changed", func(t *testing.T) {
		policy := Policy{KnownMissingEndpoints: map[string]MissingEndpoint{"other": {Status: "not-implemented", Reason: "r", Stability: "experimental", Visibility: "public", Tracking: "t"}}, FeatureFlagAllowlist: map[string]bool{}}
		dir := t.TempDir()
		writeSpec(t, dir, "o.json", `{"other":{"url":{"paths":[{"path":"/_o","methods":["GET"]}]},"stability":"stable","visibility":"public"}}`)
		res := audit(loadCatalog(t, dir), registered, policy)
		assert.Contains(t, strings.Join(res.PolicyErrors, "\n"), "stability")
	})
	t.Run("bad status reason tracking", func(t *testing.T) {
		policy := Policy{KnownMissingEndpoints: map[string]MissingEndpoint{"other": {Status: "skip", Reason: "", Visibility: "public", Stability: "stable"}}, FeatureFlagAllowlist: map[string]bool{}}
		dir := t.TempDir()
		writeSpec(t, dir, "o.json", `{"other":{"url":{"paths":[{"path":"/_o","methods":["GET"]}]},"stability":"stable","visibility":"public"}}`)
		res := audit(loadCatalog(t, dir), registered, policy)
		joined := strings.Join(res.PolicyErrors, "\n")
		assert.Contains(t, joined, "invalid status")
		assert.Contains(t, joined, "missing reason")
		assert.Contains(t, joined, "missing tracking")
	})
	t.Run("conditional absent from pin", func(t *testing.T) {
		reg := map[string]EndpointImpl{"absent.thing": {Service: "InfoService", Method: "Info"}}
		res := audit(cat, reg, Policy{FeatureFlagAllowlist: map[string]bool{}})
		assert.Contains(t, strings.Join(res.PolicyErrors, "\n"), "absent from pinned catalog")
		res = audit(cat, map[string]EndpointImpl{"absent.thing": {Service: "InfoService", Method: "Info", Conditional: true, Evidence: "outside pin"}}, Policy{FeatureFlagAllowlist: map[string]bool{}})
		assert.Empty(t, res.PolicyErrors)
	})
}

func TestAuditKnownMissingCountsExcludedFeatureFlags(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.json", `{"streams.status":{"url":{"paths":[{"path":"/_streams/status","methods":["GET"]}]},"params":{},"visibility":"feature_flag"}}`)
	cat := loadCatalog(t, dir)
	res := audit(cat, map[string]EndpointImpl{}, defaultPolicy())
	// streams.status is a conditional allowlisted name; absent from registry it
	// is a known missing endpoint only if baselined, otherwise it is a real gap.
	if _, baselined := knownMissingEndpoints["streams.status"]; !baselined {
		assert.Contains(t, res.MissingEndpoints, "streams.status")
	}
}

func TestExecute(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "info.json", infoSpec)

	t.Run("no args", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		assert.Equal(t, 2, execute(nil, &out, &errBuf))
		assert.Contains(t, errBuf.String(), "-spec is required")
	})
	t.Run("bad flag", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		assert.Equal(t, 2, execute([]string{"-unknown"}, &out, &errBuf))
	})
	t.Run("bad dir", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		assert.Equal(t, 2, execute([]string{"-spec", filepath.Join(t.TempDir(), "nope")}, &out, &errBuf))
	})
	t.Run("valid but registry gap", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		assert.Equal(t, 1, execute([]string{"-spec", dir}, &out, &errBuf))
		assert.Contains(t, out.String(), "speccheck:")
	})
	t.Run("inventory", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		assert.Equal(t, 0, execute([]string{"-spec", dir, "-inventory"}, &out, &errBuf))
		assert.Contains(t, out.String(), "registered")
	})
	t.Run("inventory bad dir", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		assert.Equal(t, 2, execute([]string{"-spec", filepath.Join(t.TempDir(), "nope"), "-inventory"}, &out, &errBuf))
	})
}

func TestRunWithPolicyMutationNewParam(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "search.json", `{"search":{"url":{"paths":[{"path":"/{index}/_search","methods":["GET"],"parts":{"index":{"type":"list"}}}]},"params":{"brand_new_param":{"type":"string"}},"visibility":"public"}}`)
	var out, errBuf bytes.Buffer
	code := runWithPolicy(dir, testRegistry("search"), Policy{FeatureFlagAllowlist: map[string]bool{}}, &out, &errBuf)
	assert.Equal(t, 1, code)
	assert.Contains(t, out.String(), "PARAM GAPS")
	assert.Contains(t, out.String(), "brand_new_param")
}

func TestRunWithPolicyMutationNewPart(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "info.json", `{"info":{"url":{"paths":[{"path":"/{node}","methods":["GET"],"parts":{"node":{"type":"string"}}}]},"params":{},"visibility":"public"}}`)
	var out, errBuf bytes.Buffer
	code := runWithPolicy(dir, testRegistry("info"), Policy{FeatureFlagAllowlist: map[string]bool{}}, &out, &errBuf)
	assert.Equal(t, 1, code)
	assert.Contains(t, out.String(), "PATH PART GAPS")
}

func TestRunWithPolicyMalformedInput(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "bad.json", `{not valid`)
	var out, errBuf bytes.Buffer
	code := runWithPolicy(dir, testRegistry("info"), Policy{FeatureFlagAllowlist: map[string]bool{}}, &out, &errBuf)
	assert.Equal(t, 2, code)
}

func TestRunNoDebtMessage(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "info.json", infoSpec)
	var out, errBuf bytes.Buffer
	code := runWithPolicy(dir, testRegistry("info"), Policy{FeatureFlagAllowlist: map[string]bool{}}, &out, &errBuf)
	require.Equal(t, 0, code, out.String())
	assert.Contains(t, out.String(), "no drift and no known debt")
}

func TestInventoryRowsSorted(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "info.json", infoSpec)
	writeSpec(t, dir, "zzz.json", `{"zzz.thing":{"url":{"paths":[{"path":"/_z","methods":["GET"]}]},"params":{},"visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	rows, err := inventory(cat, Registry)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "info", rows[0].Name)
	assert.Equal(t, "registered", rows[0].Classification)
	assert.Equal(t, "missing", rows[1].Classification)
}
