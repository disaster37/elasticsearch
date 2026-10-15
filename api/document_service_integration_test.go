//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/disaster37/elasticsearch/v9"
	"github.com/disaster37/elasticsearch/v9/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationDocumentCRUD(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()
	index := uniqueIndex("doc")

	// Create index
	_, err := c.Indices().Create(ctx, index, map[string]any{"settings": map[string]any{"number_of_shards": 1, "number_of_replicas": 0}}, &api.IndicesCreateParams{})
	assert.NoError(t, err)
	defer func() {
		_, _ = c.Indices().Delete(ctx, []string{index}, &api.IndicesDeleteParams{})
	}()

	// Index a document
	_, err = c.Document().Index(ctx, &api.IndexRequest{
		Index:  index,
		Id:     "1",
		Body:   map[string]any{"f": "v"},
		Params: &api.IndexParams{Refresh: api.RefreshTrue},
	})
	assert.NoError(t, err)

	// Get it
	getResp, err := c.Document().Get(ctx, &api.GetRequest{Index: index, Id: "1"})
	assert.NoError(t, err)
	assert.True(t, getResp.Found)

	// Exists
	ok, err := c.Document().Exists(ctx, index, "1", &api.GetParams{})
	assert.NoError(t, err)
	assert.True(t, ok)

	// Delete it
	_, err = c.Document().Delete(ctx, &api.DeleteRequest{Index: index, Id: "1"})
	assert.NoError(t, err)

	// Get -> not found
	getResp, err = c.Document().Get(ctx, &api.GetRequest{Index: index, Id: "1"})
	assert.NoError(t, err)
	assert.False(t, getResp.Found)

	// 404 helper
	_, err = c.Indices().Delete(ctx, []string{"nonexistent-" + index}, &api.IndicesDeleteParams{})
	assert.True(t, elasticsearch.IsNotFound(err))
}

// TestIntegrationDocumentCreate covers the corrected document Create route
// (Create -> PUT /{index}/_create/{id}) against a live cluster. It asserts a
// successful create is readable via Get, and that a duplicate id returns a 409
// conflict without overwriting the original document.
func TestIntegrationDocumentCreate(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()
	index := uniqueIndex("doc-create")

	_, err := c.Indices().Create(ctx, index, map[string]any{"settings": map[string]any{"number_of_shards": 1, "number_of_replicas": 0}}, &api.IndicesCreateParams{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = c.Indices().Delete(ctx, []string{index}, &api.IndicesDeleteParams{})
	})

	// Create with an explicit id.
	createResp, err := c.Document().Create(ctx, &api.CreateRequest{
		Index:  index,
		Id:     "1",
		Body:   map[string]any{"f": "v"},
		Params: &api.CreateParams{Refresh: api.RefreshTrue},
	})
	require.NoError(t, err)
	assert.Equal(t, "created", createResp.Result)

	// Read it back and assert the _source.
	getResp, err := c.Document().Get(ctx, &api.GetRequest{Index: index, Id: "1"})
	require.NoError(t, err)
	require.True(t, getResp.Found)
	assert.Equal(t, map[string]any{"f": "v"}, decodeSource(t, getResp.Source))

	// A second create with the same id must conflict and must not overwrite.
	_, err = c.Document().Create(ctx, &api.CreateRequest{
		Index: index,
		Id:    "1",
		Body:  map[string]any{"f": "v2"},
	})
	require.Error(t, err)
	assert.True(t, elasticsearch.IsConflict(err), "duplicate create should return a 409 conflict, got %v", err)

	getResp, err = c.Document().Get(ctx, &api.GetRequest{Index: index, Id: "1"})
	require.NoError(t, err)
	require.True(t, getResp.Found)
	assert.Equal(t, map[string]any{"f": "v"}, decodeSource(t, getResp.Source))
}

// decodeSource unmarshals a document _source into a generic map for assertions.
func decodeSource(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	return m
}
