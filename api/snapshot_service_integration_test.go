//go:build integration

package api_test

// Live integration coverage for the corrected SnapshotService.Get route with an
// empty snapshot selector (GET /_snapshot/{repository}/_all). An empty selector
// must succeed with an empty snapshot set rather than a 404/repository-not-found.

import (
	"context"
	"testing"

	"github.com/disaster37/elasticsearch/v9/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationSnapshotGetAll registers a filesystem repository backed by the
// fixture's path.repo, then asserts that Get with no snapshot names returns an
// empty set successfully, and returns the snapshot once one has been created.
func TestIntegrationSnapshotGetAll(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()
	repository := uniqueIndex("snapshot")
	const location = "/usr/share/elasticsearch/backup"

	_, err := c.Snapshot().CreateRepository(ctx, repository, map[string]any{
		"type":     "fs",
		"settings": map[string]any{"location": location},
	}, &api.SnapshotCreateRepositoryParams{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = c.Snapshot().DeleteRepository(ctx, []string{repository}, &api.SnapshotDeleteRepositoryParams{})
	})

	// Empty selector -> /_snapshot/{repository}/_all with no snapshots yet.
	emptyResp, err := c.Snapshot().Get(ctx, &api.SnapshotGetRequest{Repository: repository})
	require.NoError(t, err)
	require.NotNil(t, emptyResp)
	assert.Empty(t, emptyResp.Snapshots)

	// Snapshot a small dedicated index so the empty-selector read is non-trivial.
	index := uniqueIndex("snapshot-data")
	_, err = c.Indices().Create(ctx, index, map[string]any{"settings": map[string]any{"number_of_shards": 1, "number_of_replicas": 0}}, &api.IndicesCreateParams{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = c.Indices().Delete(ctx, []string{index}, &api.IndicesDeleteParams{})
	})

	const snapshot = "test-snap"
	waitForCompletion := true
	_, err = c.Snapshot().Create(ctx, &api.SnapshotCreateRequest{
		Repository: repository,
		Snapshot:   snapshot,
		Body:       map[string]any{"indices": index, "include_global_state": false},
		Params:     &api.SnapshotCreateParams{WaitForCompletion: &waitForCompletion},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = c.Snapshot().Delete(ctx, &api.SnapshotDeleteRequest{Repository: repository, Snapshot: snapshot})
	})

	// The same empty selector now returns the created snapshot.
	allResp, err := c.Snapshot().Get(ctx, &api.SnapshotGetRequest{Repository: repository})
	require.NoError(t, err)
	require.Len(t, allResp.Snapshots, 1)
	assert.Equal(t, snapshot, allResp.Snapshots[0].Snapshot)
}
