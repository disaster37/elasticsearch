//go:build integration

package api_test

// Live integration coverage for the corrected ClusterService.Info route
// (GET /_info/{target}). The allowed targets and response shape follow the
// pinned spec (cluster.info.json): _all, http, ingest, thread_pool, script.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationClusterInfo covers the corrected ClusterService.Info route:
// an empty target list falls back to /_info/_all and succeeds, an explicit
// allowed target succeeds, and an unknown target fails cleanly with an error.
func TestIntegrationClusterInfo(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Empty features -> /_info/_all.
	info, err := c.Cluster().Info(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.NotEmpty(t, *info)

	// Explicit allowed target -> /_info/thread_pool.
	threadPool, err := c.Cluster().Info(ctx, []string{"thread_pool"})
	require.NoError(t, err)
	require.NotNil(t, threadPool)
	assert.Contains(t, *threadPool, "thread_pool")

	// Unknown target fails cleanly (error, no panic).
	_, err = c.Cluster().Info(ctx, []string{"not-a-target"})
	require.Error(t, err)
}
