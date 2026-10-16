package main

import (
	"strings"
	"testing"

	"github.com/disaster37/elasticsearch/v9/internal/speccheck/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePolicyIgnoredParamExists(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "search.json", `{"search":{"url":{"paths":[{"path":"/{index}/_search","methods":["GET"],"parts":{"index":{"type":"list"}}}]},"params":{"q":{"type":"string"}},"visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	reg := testRegistry("search")

	policy := Policy{
		IgnoredParams: map[string]map[string]IgnoredParam{
			"search": {"q": {Reason: "controlled", ControlledValue: "x", Evidence: "test"}},
		},
		FeatureFlagAllowlist: map[string]bool{},
	}
	errs := validatePolicy(cat, reg, policy)
	assert.Contains(t, strings.Join(errs, "\n"), "already exposed as a field")
}

func TestValidatePolicyIgnoredParamTypo(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "search.json", `{"search":{"url":{"paths":[{"path":"/{index}/_search","methods":["GET"],"parts":{"index":{"type":"list"}}}]},"params":{},"visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	policy := Policy{
		IgnoredParams: map[string]map[string]IgnoredParam{
			"search": {"nonexistent": {Reason: "r", ControlledValue: "v", Evidence: "e"}},
		},
		FeatureFlagAllowlist: map[string]bool{},
	}
	errs := validatePolicy(cat, testRegistry("search"), policy)
	assert.Contains(t, strings.Join(errs, "\n"), "not a spec param")
}

func TestValidatePolicyIgnoredParamUnregisteredEndpoint(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "x.json", `{"x.thing":{"url":{"paths":[{"path":"/_x","methods":["GET"]}]},"params":{"p":{"type":"string"}},"visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	policy := Policy{
		IgnoredParams: map[string]map[string]IgnoredParam{
			"x.thing": {"p": {Reason: "r", ControlledValue: "v", Evidence: "e"}},
		},
		FeatureFlagAllowlist: map[string]bool{},
	}
	errs := validatePolicy(cat, map[string]EndpointImpl{}, policy)
	assert.Contains(t, strings.Join(errs, "\n"), "not a checked registered endpoint")
}

func TestValidatePolicyIgnoredParamMissingEvidence(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "search.json", `{"search":{"url":{"paths":[{"path":"/{index}/_search","methods":["GET"],"parts":{"index":{"type":"list"}}}]},"params":{"brand_new":{"type":"string"}},"visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	policy := Policy{
		IgnoredParams: map[string]map[string]IgnoredParam{
			"search": {"brand_new": {Reason: "", ControlledValue: "", Evidence: ""}},
		},
		FeatureFlagAllowlist: map[string]bool{},
	}
	errs := validatePolicy(cat, testRegistry("search"), policy)
	joined := strings.Join(errs, "\n")
	assert.Contains(t, joined, "reasons, evidence, and controlled value are required")
}

// TestCatFormatIgnore proves the exact controlled Cat format exception works
// and that its controlled wire value is actually emitted by ToMap.
func TestCatFormatIgnore(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "cat.count.json", `{"cat.count":{"url":{"paths":[{"path":"/_cat/count","methods":["GET"]}]},"params":{"format":{"type":"string"}},"stability":"stable","visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	reg := testRegistry("cat.count")
	policy := defaultPolicy()
	res := audit(cat, reg, policy)
	assert.Empty(t, res.ParamGaps, "%v", res.ParamGaps)
	assert.Empty(t, res.IntegrityErrors)
	assert.Equal(t, 1, res.Ignored)
}

func TestValidateControlledIgnoreFailsWhenRemoved(t *testing.T) {
	res := audit(specCatalog(t, "cat.count", `{"format":{"type":"string"}}`), testRegistry("cat.count"), Policy{
		IgnoredParams: map[string]map[string]IgnoredParam{
			"cat.count": {"format": {Reason: "r", ControlledValue: "yaml", Evidence: "e"}},
		},
		FeatureFlagAllowlist: map[string]bool{},
	})
	assert.Contains(t, strings.Join(res.IntegrityErrors, "\n"), "controlled")
}

func specCatalog(t *testing.T, name, paramsJSON string) (c spec.Catalog) {
	t.Helper()
	dir := t.TempDir()
	writeSpec(t, dir, "e.json", `{"`+name+`":{"url":{"paths":[{"path":"/_cat/count","methods":["GET"]}]},"params":`+paramsJSON+`,"visibility":"public"}}`)
	return loadCatalog(t, dir)
}

// TestFeatureFlagParamStillChecked proves parameter visibility metadata never
// suppresses a required endpoint-local parameter.
func TestFeatureFlagParamStillChecked(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "info.json", `{"info":{"url":{"paths":[{"path":"/","methods":["GET"]}]},"params":{"force_thing":{"type":"boolean","visibility":"feature_flag"}},"visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	res := audit(cat, testRegistry("info"), Policy{FeatureFlagAllowlist: map[string]bool{}})
	require.Contains(t, res.ParamGaps, "info: param \"force_thing\" missing from InfoService.Info")
}

func TestDefaultPolicyFeatureFlagAllowlist(t *testing.T) {
	p := defaultPolicy()
	for _, name := range []string{"streams.logs_enable", "streams.logs_disable", "streams.status"} {
		assert.True(t, p.FeatureFlagAllowlist[name], name)
	}
	assert.Len(t, p.FeatureFlagAllowlist, 3, "no broad prefix allowance")
}

func TestBaselineNeverUsesWildcards(t *testing.T) {
	for name := range knownMissingEndpoints {
		assert.NotContains(t, name, "*")
	}
}

func TestQueryFieldOverride(t *testing.T) {
	impl := EndpointImpl{QueryFields: map[string]string{"df": "Df"}}
	assert.Equal(t, "Df", queryField(impl, "df"))
	assert.Equal(t, "ForceSyntheticSource", queryField(EndpointImpl{}, "force_synthetic_source"))
}

func TestInventoryMissingFields(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "search.json", `{"search":{"url":{"paths":[{"path":"/{index}/_search","methods":["GET"],"parts":{"index":{"type":"list"}}}]},"params":{"brand_new":{"type":"string"}},"visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	rows, err := inventory(cat, testRegistry("search"))
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, []string{"brand_new"}, rows[0].MissingFields)
}

// TestMissingRegisteredFieldsQueryFieldsOverride proves a QueryFields override
// never provides imaginary coverage: an override pointing at a non-existent
// field must still be reported missing (matching the audit's behavior).
func TestMissingRegisteredFieldsQueryFieldsOverride(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "search.json", `{"search":{"url":{"paths":[{"path":"/{index}/_search","methods":["GET"],"parts":{"index":{"type":"list"}}}]},"params":{"q":{"type":"string"}},"visibility":"public"}}`)
	cat := loadCatalog(t, dir)
	impl := Registry["search"]
	impl.QueryFields = map[string]string{"q": "DoesNotExist"}
	assert.Equal(t, []string{"q"}, missingRegisteredFields(impl, cat.Endpoints["search"]))
	// A valid override to a real field is recognized as coverage.
	impl.QueryFields = map[string]string{"q": "Q"}
	assert.Empty(t, missingRegisteredFields(impl, cat.Endpoints["search"]))
}
