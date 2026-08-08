//go:build integration

package api_test

import (
	"context"
	"testing"

	"github.com/disaster37/elasticsearch/v9"
	"github.com/disaster37/elasticsearch/v9/api"
	"github.com/stretchr/testify/assert"
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
