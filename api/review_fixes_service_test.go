package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	json "github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
)

// TestUnitIndicesExistsEmptyIndices ensures Exists does not fall back to
// HEAD / (the ping endpoint, which answers 200) when no indices are given.
func TestUnitIndicesExistsEmptyIndices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected, got %s %s", r.Method, r.URL.Path)
	}))
	defer srv.Close()
	svc := NewIndicesService(restyClient(srv), testLogger())
	ok, err := svc.Exists(context.Background(), nil, nil)
	assert.Error(t, err)
	assert.False(t, ok)
}

// TestUnitIndicesGetEmptyIndices ensures Get requires at least one index.
func TestUnitIndicesGetEmptyIndices(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Get(context.Background(), nil, nil)
	assert.Error(t, err)
}

// TestUnitIndicesOpenEmptyIndices ensures Open requires at least one index.
func TestUnitIndicesOpenEmptyIndices(t *testing.T) {
	svc := NewIndicesService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Open(context.Background(), nil, nil)
	assert.Error(t, err)
	_, err = svc.Close(context.Background(), nil, nil)
	assert.Error(t, err)
}

// TestUnitSearchMVTZeroZoom ensures zoom/x/y of 0 are valid (a plain int with
// validate:"required" would wrongly reject them).
func TestUnitSearchMVTZeroZoom(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_mvt/field/0/0/0", r.URL.Path)
		w.Header().Set("Content-Type", "application/vnd.mapbox-vector-tile")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("binary-tile"))
	}))
	defer srv.Close()
	svc := NewSearchService(restyClient(srv), testLogger())
	resp, err := svc.SearchMVT(context.Background(), &SearchMVTRequest{
		Index: "idx",
		Field: "field",
		Zoom:  ptrInt(0),
		X:     ptrInt(0),
		Y:     ptrInt(0),
	})
	assert.NoError(t, err)
	assert.Equal(t, "binary-tile", string(resp))
}

// TestUnitUpdateByQueryParams ensures the update-specific pipeline and
// version_type parameters are serialized.
func TestUnitUpdateByQueryParams(t *testing.T) {
	m := (&UpdateByQueryParams{
		Pipeline:    "my-pipeline",
		VersionType: VersionTypeExternal,
	}).ToMap()
	assert.Equal(t, "my-pipeline", m["pipeline"])
	assert.Equal(t, "external", m["version_type"])
}

// TestUnitTextStructureFindFieldStructureSendsField ensures the required
// field query parameter is transmitted.
func TestUnitTextStructureFindFieldStructureSendsField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/idx/_text_structure/field_structure", r.URL.Path)
		assert.Equal(t, "message", r.URL.Query().Get("field"))
		jsonResponse(w, 200, `{"format":"json"}`)
	}))
	defer srv.Close()
	svc := NewTextStructureService(restyClient(srv), testLogger())
	_, err := svc.FindFieldStructure(context.Background(), "idx", "message", nil)
	assert.NoError(t, err)
}

// TestUnitXpackUsageUnmarshal ensures feature-specific fields are captured in
// Data despite not being declared on the struct.
func TestUnitXpackUsageUnmarshal(t *testing.T) {
	var usage XpackUsage
	err := json.Unmarshal([]byte(`{"available":true,"enabled":true,"count":3,"jobs":["a","b"]}`), &usage)
	assert.NoError(t, err)
	assert.True(t, usage.Available)
	assert.True(t, usage.Enabled)
	assert.Equal(t, float64(3), usage.Data["count"])
	assert.Equal(t, []any{"a", "b"}, usage.Data["jobs"])
}
