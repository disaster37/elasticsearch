package main

import (
	"reflect"
	"testing"

	"github.com/disaster37/elasticsearch/v9/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRegistryProductionEntries(t *testing.T) {
	require.NotEmpty(t, Registry)
	assert.Empty(t, validateRegistry(Registry), "every production binding must resolve")
}

func TestValidateRegistryErrors(t *testing.T) {
	good := Registry["search"]
	base := func() map[string]EndpointImpl {
		return map[string]EndpointImpl{"search": good}
	}

	t.Run("unknown service", func(t *testing.T) {
		reg := base()
		reg["search"] = EndpointImpl{Service: "NopeService", Method: "Search"}
		assert.Contains(t, join(validateRegistry(reg)), "unknown service")
	})
	t.Run("unknown method", func(t *testing.T) {
		reg := base()
		reg["search"] = EndpointImpl{Service: "SearchService", Method: "Nope"}
		assert.Contains(t, join(validateRegistry(reg)), "has no method")
	})
	t.Run("unknown params struct", func(t *testing.T) {
		reg := base()
		e := good
		e.ParamsStruct = "api.DoesNotExistParams"
		reg["search"] = e
		assert.Contains(t, join(validateRegistry(reg)), "unknown ParamsStruct")
	})
	t.Run("params source wrong carrier", func(t *testing.T) {
		reg := base()
		e := good
		e.ParamsSource = &ValueSource{Argument: 1, Field: "Body"}
		reg["search"] = e
		assert.Contains(t, join(validateRegistry(reg)), "ParamsSource reaches")
	})
	t.Run("incomplete params fields", func(t *testing.T) {
		reg := base()
		e := good
		e.ParamsFields = []string{"Q"}
		reg["search"] = e
		assert.Contains(t, join(validateRegistry(reg)), "ParamsFields mismatch")
	})
	t.Run("query arg out of range", func(t *testing.T) {
		reg := base()
		e := good
		e.QueryArgs = map[string]ValueSource{"x": {Argument: 99}}
		reg["search"] = e
		assert.Contains(t, join(validateRegistry(reg)), "out of range")
	})
	t.Run("path parts inconsistent", func(t *testing.T) {
		reg := base()
		e := good
		e.PathParts = []string{"index", "extra"}
		reg["search"] = e
		assert.Contains(t, join(validateRegistry(reg)), "PathParts")
	})
	t.Run("path source unreachable", func(t *testing.T) {
		reg := base()
		e := good
		e.PathSources = map[string]ValueSource{"index": {Argument: 1, Field: "Nope"}}
		e.PathParts = []string{"index"}
		reg["search"] = e
		assert.Contains(t, join(validateRegistry(reg)), "PathSources")
	})
	t.Run("conditional without evidence", func(t *testing.T) {
		reg := base()
		e := good
		e.Conditional = true
		e.Evidence = ""
		reg["search"] = e
		assert.Contains(t, join(validateRegistry(reg)), "requires evidence")
	})
	t.Run("params without struct", func(t *testing.T) {
		reg := base()
		e := good
		e.ParamsStruct = ""
		reg["search"] = e
		assert.Contains(t, join(validateRegistry(reg)), "without ParamsStruct")
	})
}

func TestExportedFieldsIncludesPromoted(t *testing.T) {
	fields := exportedFields(reflect.TypeOf(api.CatCountParams{}))
	for _, want := range []string{"H", "Help", "S", "V", "Local", "MasterTimeout", "Time", "Common"} {
		assert.Contains(t, fields, want)
	}
}

func TestResolveSource(t *testing.T) {
	svc := resolveService("SearchService")
	require.NotNil(t, svc)
	method, ok := svc.MethodByName("Search")
	require.True(t, ok)

	t.Run("argument", func(t *testing.T) {
		typ, err := resolveSource(method, ValueSource{Argument: 0})
		require.NoError(t, err)
		assert.Equal(t, "context.Context", typ.String())
	})
	t.Run("request field", func(t *testing.T) {
		typ, err := resolveSource(method, ValueSource{Argument: -1, Field: "Indices"})
		require.NoError(t, err)
		assert.Equal(t, reflect.TypeOf([]string{}), typ)
	})
	t.Run("out of range", func(t *testing.T) {
		_, err := resolveSource(method, ValueSource{Argument: 9})
		assert.Error(t, err)
	})
	t.Run("missing field", func(t *testing.T) {
		_, err := resolveSource(method, ValueSource{Argument: 1, Field: "Nope"})
		assert.Error(t, err)
	})
	t.Run("no request field", func(t *testing.T) {
		_, err := resolveSource(method, ValueSource{Argument: -1, Field: "NoSuchField"})
		assert.Error(t, err)
	})
}

func TestValidateParamSerialization(t *testing.T) {
	impl := Registry["search"]
	t.Run("boolean preserves explicit false", func(t *testing.T) {
		assert.NoError(t, validateParamSerialization(impl, "force_synthetic_source", "ForceSyntheticSource"))
	})
	t.Run("missing wire key fails", func(t *testing.T) {
		assert.Error(t, validateParamSerialization(impl, "not_a_wire_key", "Q"))
	})
	t.Run("list serialized", func(t *testing.T) {
		assert.NoError(t, validateParamSerialization(impl, "docvalue_fields", "DocvalueFields"))
	})
	t.Run("integer zero", func(t *testing.T) {
		assert.NoError(t, validateParamSerialization(impl, "batched_reduce_size", "BatchedReduceSize"))
	})
}

func TestCheckFieldSerializationUnsupported(t *testing.T) {
	impl := Registry["search"]
	// SearchParams.Common is a pointer to a struct and has no wire mapping.
	assert.Error(t, validateParamSerialization(impl, "common", "Common"))
}

func join(errs []string) string {
	out := ""
	for _, e := range errs {
		out += e + "\n"
	}
	return out
}
