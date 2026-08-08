package elasticsearch

import "github.com/disaster37/elasticsearch/v9/types"

// IsNotFound returns true when err is an [*ElasticsearchError] with status 404.
func IsNotFound(err error) bool {
	return types.IsNotFound(err)
}

// IsConflict returns true when err is an [*ElasticsearchError] with status 409.
func IsConflict(err error) bool {
	return types.IsConflict(err)
}

// IsUnauthorized returns true when err is an [*ElasticsearchError] with status 401.
func IsUnauthorized(err error) bool {
	return types.IsUnauthorized(err)
}
