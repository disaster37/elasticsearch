package elasticsearch

import (
	"errors"
	"testing"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/stretchr/testify/assert"
)

func TestIsNotFound(t *testing.T) {
	assert.True(t, IsNotFound(&types.ElasticsearchError{Status: 404}))
	assert.False(t, IsNotFound(&types.ElasticsearchError{Status: 400}))
	assert.False(t, IsNotFound(errors.New("other")))
	assert.False(t, IsNotFound(nil))
}

func TestIsConflict(t *testing.T) {
	assert.True(t, IsConflict(&types.ElasticsearchError{Status: 409}))
	assert.False(t, IsConflict(&types.ElasticsearchError{Status: 400}))
	assert.False(t, IsConflict(errors.New("other")))
	assert.False(t, IsConflict(nil))
}

func TestIsUnauthorized(t *testing.T) {
	assert.True(t, IsUnauthorized(&types.ElasticsearchError{Status: 401}))
	assert.False(t, IsUnauthorized(&types.ElasticsearchError{Status: 403}))
	assert.False(t, IsUnauthorized(errors.New("other")))
	assert.False(t, IsUnauthorized(nil))
}

// TestIsHelpersWithDetails verifies the helpers work on errors that carry
// structured details.
func TestIsHelpersWithDetails(t *testing.T) {
	e := &types.ElasticsearchError{
		Status: 404,
		Details: &types.ElasticsearchErrorDetails{
			Type:   "index_not_found_exception",
			Reason: "no such index",
		},
	}
	assert.True(t, IsNotFound(e))
	assert.False(t, IsConflict(e))
	assert.False(t, IsUnauthorized(e))
}
