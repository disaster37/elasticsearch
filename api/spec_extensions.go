package api

import (
	"context"
	"fmt"
	"reflect"
)

// This file adds opt-in extension APIs for spec endpoints whose original
// service method has no query carrier or no path selector. Existing service
// interfaces and their method signatures are deliberately untouched:
// extensions live on separate <Group>ServiceSpecExtensions interfaces that the
// default implementations satisfy, and callers obtain them through the
// As<Group>ServiceSpecExtensions adapters.

// AsyncSearchServiceSpecExtensions exposes the opt-in async_search.status
// carrier. Obtain it with [AsAsyncSearchServiceSpecExtensions].
type AsyncSearchServiceSpecExtensions interface {
	StatusWithParams(ctx context.Context, id string, params *AsyncSearchStatusParams) (*AsyncSearchStatusResponse, error)
}

// Compile-time assertion that the default implementation supports the
// extension interface.
var _ AsyncSearchServiceSpecExtensions = (*DefaultAsyncSearchService)(nil)

// AsAsyncSearchServiceSpecExtensions adapts an AsyncSearchService to its
// extension interface. It rejects nil and typed-nil services and returns a
// descriptive error for custom implementations without support.
func AsAsyncSearchServiceSpecExtensions(service AsyncSearchService) (AsyncSearchServiceSpecExtensions, error) {
	if isNilService(service) {
		return nil, fmt.Errorf("as async search spec extensions: nil service")
	}
	ext, ok := service.(AsyncSearchServiceSpecExtensions)
	if !ok {
		return nil, fmt.Errorf("as async search spec extensions: %T does not implement AsyncSearchServiceSpecExtensions", service)
	}
	if isNilService(ext) {
		return nil, fmt.Errorf("as async search spec extensions: typed-nil service")
	}
	return ext, nil
}

// StatusWithParams performs GET /_async_search/status/{id} with the optional
// keep_alive query parameter.
func (s *DefaultAsyncSearchService) StatusWithParams(ctx context.Context, id string, params *AsyncSearchStatusParams) (*AsyncSearchStatusResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result AsyncSearchStatusResponse
	err := doReq(s.logger, ctx, s.client, "GET", fmt.Sprintf("/_async_search/status/%s", pathSegment(id)), nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// isNilService reports whether a service interface holds no usable value,
// including a typed-nil pointer.
func isNilService(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return rv.IsNil()
	default:
		return false
	}
}
