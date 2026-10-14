package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadBindings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bindings.json")
	good := `[
	  {"name": "search", "service": "SearchService", "method": "Search", "paramsStruct": "api.SearchParams", "paramsSource": {"argument": -1, "field": "Params"}, "pathSources": {"index": {"argument": -1, "field": "Indices"}}}
	]`
	require.NoError(t, os.WriteFile(path, []byte(good), 0o644))
	bindings, err := ReadBindings(path)
	require.NoError(t, err)
	require.Contains(t, bindings, "search")
	assert.Equal(t, -1, bindings["search"].ParamsSource.Argument)
	assert.Equal(t, "Indices", bindings["search"].PathSources["index"].Field)
}

func TestReadBindingsErrors(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"duplicate":      `[{"name":"a","service":"S","method":"M"},{"name":"a","service":"S","method":"M"}]`,
		"empty name":     `[{"name":"","service":"S","method":"M"}]`,
		"missing svc":    `[{"name":"a","method":"M"}]`,
		"missing method": `[{"name":"a","service":"S"}]`,
		"malformed":      `{`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name+".json")
			require.NoError(t, os.WriteFile(path, []byte(raw), 0o644))
			_, err := ReadBindings(path)
			assert.Error(t, err)
		})
	}
	t.Run("missing file", func(t *testing.T) {
		_, err := ReadBindings(filepath.Join(dir, "nope.json"))
		assert.Error(t, err)
	})
}

func TestAccessibleFieldsPromotion(t *testing.T) {
	model := APIModel{
		Structs: map[string]StructInfo{
			"Outer": {
				Fields: []Field{{Name: "A", Exported: true}, {Name: "X", Exported: true}, {Name: "hidden", Exported: false}},
				Embeds: []string{"inner"},
			},
			"inner": {
				Fields: []Field{{Name: "B", Exported: true}, {Name: "X", Exported: true}},
			},
		},
		Aliases: map[string]string{"OuterAlias": "Outer", "LoopA": "LoopB", "LoopB": "LoopA"},
	}
	assert.Equal(t, []string{"A", "B", "X"}, model.AccessibleFields("Outer"))
	assert.Equal(t, []string{"A", "B", "X"}, model.AccessibleFields("OuterAlias"))
	assert.Empty(t, model.AccessibleFields("LoopA"))
	assert.Empty(t, model.AccessibleFields("Missing"))
}

func TestAccessibleFieldsAmbiguity(t *testing.T) {
	model := APIModel{
		Structs: map[string]StructInfo{
			"Outer": {Embeds: []string{"left", "right"}},
			"left":  {Fields: []Field{{Name: "C", Exported: true}}},
			"right": {Fields: []Field{{Name: "C", Exported: true}}},
		},
	}
	assert.Empty(t, model.AccessibleFields("Outer"), "ambiguous promoted field must be excluded")
}

func TestScanAPI(t *testing.T) {
	model, err := ScanAPI("../../../api")
	require.NoError(t, err)
	assert.Contains(t, model.Services, "SearchService")
	assert.Contains(t, model.ParamsTypes, "SearchParams")
	assert.Contains(t, model.Structs, "catCommonParams")
	methods := model.Methods["SearchService"]
	require.Contains(t, methods, "Search")
	assert.Equal(t, []string{"context.Context", "*SearchRequest"}, methods["Search"].Inputs)
}

func TestGenerateTypesStable(t *testing.T) {
	model, err := ScanAPI("../../../api")
	require.NoError(t, err)
	a, err := GenerateTypes(model)
	require.NoError(t, err)
	b, err := GenerateTypes(model)
	require.NoError(t, err)
	assert.Equal(t, a, b, "generation must be byte-for-byte stable")
	assert.Contains(t, string(a), "reflect.TypeOf(api.SearchParams{})")
	assert.Contains(t, string(a), `reflect.TypeOf((*api.SearchService)(nil)).Elem()`)
}

func TestGenerateRegistryRejectsUnknownParams(t *testing.T) {
	model, err := ScanAPI("../../../api")
	require.NoError(t, err)
	_, err = GenerateRegistry(map[string]EndpointBinding{
		"x": {Service: "SearchService", Method: "Search", ParamsStruct: "api.DoesNotExistParams"},
	}, model)
	assert.Error(t, err)
}

func TestGenerateRegistryDeterministic(t *testing.T) {
	model, err := ScanAPI("../../../api")
	require.NoError(t, err)
	bindings := map[string]EndpointBinding{
		"search": {Service: "SearchService", Method: "Search", ParamsStruct: "api.SearchParams", ParamsSource: &ValueSource{Argument: -1, Field: "Params"}},
	}
	a, err := GenerateRegistry(bindings, model)
	require.NoError(t, err)
	b, err := GenerateRegistry(bindings, model)
	require.NoError(t, err)
	assert.Equal(t, a, b)
	assert.Contains(t, string(a), `"search": {`)
	assert.Contains(t, string(a), "ParamsFields:")
}

// TestCommittedGeneratedMatches is the regeneration check: regenerating from
// the reviewed inputs must reproduce the committed files byte-for-byte without
// touching specs or baselines.
func TestCommittedGeneratedMatches(t *testing.T) {
	bindings, err := ReadBindings("../bindings.json")
	require.NoError(t, err)
	model, err := ScanAPI("../../../api")
	require.NoError(t, err)
	registry, err := GenerateRegistry(bindings, model)
	require.NoError(t, err)
	types, err := GenerateTypes(model)
	require.NoError(t, err)
	committedRegistry, err := os.ReadFile("../registry_generated.go")
	require.NoError(t, err)
	committedTypes, err := os.ReadFile("../types_generated.go")
	require.NoError(t, err)
	assert.Equal(t, string(committedRegistry), string(registry), "run go generate ./internal/speccheck")
	assert.Equal(t, string(committedTypes), string(types), "run go generate ./internal/speccheck")
}

func TestWriteGeneratedRejectsEmpty(t *testing.T) {
	assert.Error(t, WriteGenerated(t.TempDir(), nil, []byte("x")))
	assert.Error(t, WriteGenerated(t.TempDir(), []byte("x"), nil))
}

func TestWriteGeneratedWritesBothFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, WriteGenerated(dir, []byte("// registry\n"), []byte("// types\n")))
	reg, err := os.ReadFile(filepath.Join(dir, "registry_generated.go"))
	require.NoError(t, err)
	assert.Equal(t, "// registry\n", string(reg))
	typ, err := os.ReadFile(filepath.Join(dir, "types_generated.go"))
	require.NoError(t, err)
	assert.Equal(t, "// types\n", string(typ))
}

func TestGenerateRegistryQueryFields(t *testing.T) {
	model, err := ScanAPI("../../../api")
	require.NoError(t, err)
	bindings := map[string]EndpointBinding{
		"search": {
			Service:      "SearchService",
			Method:       "Search",
			ParamsStruct: "api.SearchParams",
			ParamsSource: &ValueSource{Argument: -1, Field: "Params"},
			QueryFields:  map[string]string{"df": "Df"},
		},
	}
	out, err := GenerateRegistry(bindings, model)
	require.NoError(t, err)
	assert.Contains(t, string(out), `QueryFields: map[string]string{`)
	assert.Contains(t, string(out), `"df": "Df"`)
}
