package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsAsyncSearchServiceSpecExtensionsDefault(t *testing.T) {
	svc := NewAsyncSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	ext, err := AsAsyncSearchServiceSpecExtensions(svc)
	require.NoError(t, err)
	assert.NotNil(t, ext)
}

func TestAsAsyncSearchServiceSpecExtensionsNil(t *testing.T) {
	_, err := AsAsyncSearchServiceSpecExtensions(nil)
	assert.Error(t, err)
}

func TestAsAsyncSearchServiceSpecExtensionsTypedNil(t *testing.T) {
	var svc AsyncSearchService = (*DefaultAsyncSearchService)(nil)
	_, err := AsAsyncSearchServiceSpecExtensions(svc)
	assert.Error(t, err)
}

type customAsyncSearchService struct{}

func (customAsyncSearchService) Submit(context.Context, *AsyncSearchSubmitRequest) (*AsyncSearchResponse, error) {
	return nil, nil
}
func (customAsyncSearchService) Get(context.Context, string, *AsyncSearchGetParams) (*AsyncSearchResponse, error) {
	return nil, nil
}
func (customAsyncSearchService) Status(context.Context, string) (*AsyncSearchStatusResponse, error) {
	return nil, nil
}
func (customAsyncSearchService) Delete(context.Context, string) (*types.AcknowledgedResponse, error) {
	return nil, nil
}

func TestAsAsyncSearchServiceSpecExtensionsCustom(t *testing.T) {
	_, err := AsAsyncSearchServiceSpecExtensions(customAsyncSearchService{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not implement")
}

func TestStatusWithParamsTransport(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"async-id","is_running":true,"is_partial":false}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	ext, err := AsAsyncSearchServiceSpecExtensions(svc)
	require.NoError(t, err)
	resp, err := ext.StatusWithParams(context.Background(), "async-id", &AsyncSearchStatusParams{KeepAlive: "1m"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_async_search/status/async-id", gotPath)
	assert.Equal(t, "1m", gotQuery["keep_alive"])
	assert.Equal(t, "async-id", resp.Id)
}

func TestStatusWithParamsNilParams(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"x"}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	ext, err := AsAsyncSearchServiceSpecExtensions(svc)
	require.NoError(t, err)
	_, err = ext.StatusWithParams(context.Background(), "x", nil)
	require.NoError(t, err)
	assert.NotContains(t, gotQuery, "keep_alive")
}

func TestStatusWrapperUnchanged(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"id":"x"}`)
	}))
	defer srv.Close()
	svc := NewAsyncSearchService(restyClient(srv), testLogger())
	_, err := svc.Status(context.Background(), "x")
	require.NoError(t, err)
	assert.NotContains(t, gotQuery, "keep_alive")
}

func TestStatusWithParamsEmptyID(t *testing.T) {
	svc := NewAsyncSearchService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	ext, err := AsAsyncSearchServiceSpecExtensions(svc)
	require.NoError(t, err)
	_, err = ext.StatusWithParams(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestAsyncSearchStatusParamsToMap(t *testing.T) {
	assert.Nil(t, (*AsyncSearchStatusParams)(nil).ToMap())
	assert.Empty(t, (&AsyncSearchStatusParams{}).ToMap())
	m := (&AsyncSearchStatusParams{KeepAlive: "2m"}).ToMap()
	assert.Equal(t, "2m", m["keep_alive"])
}

// capturedSpecRequest records the last request seen by a test server.
type capturedSpecRequest struct {
	method      string
	path        string
	escapedPath string
	query       map[string]string
	body        string
}

func captureSpecServer(t *testing.T) (*httptest.Server, *capturedSpecRequest) {
	t.Helper()
	got := &capturedSpecRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = capturedSpecRequest{
			method:      r.Method,
			path:        r.URL.Path,
			escapedPath: r.URL.EscapedPath(),
			query:       queryMap(r),
			body:        readBody(r),
		}
		jsonResponse(w, 200, `{}`)
	}))
	return srv, got
}

func TestSharedTimeoutCarriersToMap(t *testing.T) {
	assert.Nil(t, (*MasterTimeoutSpecParams)(nil).ToMap())
	assert.Nil(t, (*TimeoutSpecParams)(nil).ToMap())
	assert.Nil(t, (*MasterTimeoutAndTimeoutSpecParams)(nil).ToMap())

	assert.Empty(t, (&MasterTimeoutSpecParams{}).ToMap())
	assert.Empty(t, (&TimeoutSpecParams{}).ToMap())
	assert.Empty(t, (&MasterTimeoutAndTimeoutSpecParams{}).ToMap())

	m := (&MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1", Timeout: "t1"}).ToMap()
	assert.Equal(t, "m1", m["master_timeout"])
	assert.Equal(t, "t1", m["timeout"])

	// Common params are merged without overriding endpoint-local values.
	mm := (&MasterTimeoutSpecParams{MasterTimeout: "m1", Common: &CommonParams{Pretty: ptrBool(true)}}).ToMap()
	assert.Equal(t, "m1", mm["master_timeout"])
	assert.Equal(t, "true", mm["pretty"])
}

func TestB1SpecExtensionAdapters(t *testing.T) {
	ccr := NewCcrService(nil, testLogger())
	slm := NewSlmService(nil, testLogger())
	ilm := NewIlmService(nil, testLogger())
	enr := NewEnrichService(nil, testLogger())
	feat := NewFeaturesService(nil, testLogger())
	wat := NewWatcherService(nil, testLogger())
	xp := NewXpackService(nil, testLogger())

	if _, err := AsCcrServiceSpecExtensions(ccr); err != nil {
		t.Fatalf("ccr default: %v", err)
	}
	if _, err := AsSlmServiceSpecExtensions(slm); err != nil {
		t.Fatalf("slm default: %v", err)
	}
	if _, err := AsIlmServiceSpecExtensions(ilm); err != nil {
		t.Fatalf("ilm default: %v", err)
	}
	if _, err := AsEnrichServiceSpecExtensions(enr); err != nil {
		t.Fatalf("enrich default: %v", err)
	}
	if _, err := AsFeaturesServiceSpecExtensions(feat); err != nil {
		t.Fatalf("features default: %v", err)
	}
	if _, err := AsWatcherServiceSpecExtensions(wat); err != nil {
		t.Fatalf("watcher default: %v", err)
	}
	if _, err := AsXpackServiceSpecExtensions(xp); err != nil {
		t.Fatalf("xpack default: %v", err)
	}

	if _, err := AsCcrServiceSpecExtensions(nil); err == nil {
		t.Error("ccr nil: want error")
	}
	if _, err := AsSlmServiceSpecExtensions(nil); err == nil {
		t.Error("slm nil: want error")
	}
	if _, err := AsIlmServiceSpecExtensions(nil); err == nil {
		t.Error("ilm nil: want error")
	}
	if _, err := AsEnrichServiceSpecExtensions(nil); err == nil {
		t.Error("enrich nil: want error")
	}
	if _, err := AsFeaturesServiceSpecExtensions(nil); err == nil {
		t.Error("features nil: want error")
	}
	if _, err := AsWatcherServiceSpecExtensions(nil); err == nil {
		t.Error("watcher nil: want error")
	}
	if _, err := AsXpackServiceSpecExtensions(nil); err == nil {
		t.Error("xpack nil: want error")
	}

	var typedNilCcr CcrService = (*DefaultCcrService)(nil)
	if _, err := AsCcrServiceSpecExtensions(typedNilCcr); err == nil {
		t.Error("ccr typed-nil: want error")
	}
	var typedNilSlm SlmService = (*DefaultSlmService)(nil)
	if _, err := AsSlmServiceSpecExtensions(typedNilSlm); err == nil {
		t.Error("slm typed-nil: want error")
	}
	var typedNilIlm IlmService = (*DefaultIlmService)(nil)
	if _, err := AsIlmServiceSpecExtensions(typedNilIlm); err == nil {
		t.Error("ilm typed-nil: want error")
	}
	var typedNilEnr EnrichService = (*DefaultEnrichService)(nil)
	if _, err := AsEnrichServiceSpecExtensions(typedNilEnr); err == nil {
		t.Error("enrich typed-nil: want error")
	}
	var typedNilFeat FeaturesService = (*DefaultFeaturesService)(nil)
	if _, err := AsFeaturesServiceSpecExtensions(typedNilFeat); err == nil {
		t.Error("features typed-nil: want error")
	}
	var typedNilWat WatcherService = (*DefaultWatcherService)(nil)
	if _, err := AsWatcherServiceSpecExtensions(typedNilWat); err == nil {
		t.Error("watcher typed-nil: want error")
	}
	var typedNilXp XpackService = (*DefaultXpackService)(nil)
	if _, err := AsXpackServiceSpecExtensions(typedNilXp); err == nil {
		t.Error("xpack typed-nil: want error")
	}

	// A custom service that embeds the base interface implements the service
	// but not the extension interface.
	if _, err := AsCcrServiceSpecExtensions(customCcr{ccr}); err == nil {
		t.Error("ccr custom: want error")
	}
	if _, err := AsIlmServiceSpecExtensions(customIlm{ilm}); err == nil {
		t.Error("ilm custom: want error")
	}
}

type customCcr struct{ CcrService }
type customIlm struct{ IlmService }

func TestCcrSpecExtensionsTransport(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	base := NewCcrService(restyClient(srv), testLogger())
	ext, err := AsCcrServiceSpecExtensions(base)
	require.NoError(t, err)
	ctx := context.Background()

	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
		query  map[string]string
	}{
		{"put_auto_follow_pattern", func() error {
			_, err := ext.PutAutoFollowPatternWithRequest(ctx, &CcrPutAutoFollowPatternRequest{Name: "n1", Body: map[string]any{"remote_cluster": "rc1"}, Params: &MasterTimeoutSpecParams{MasterTimeout: "m1"}})
			return err
		}, "PUT", "/_ccr/auto_follow/n1", map[string]string{"master_timeout": "m1"}},
		{"delete_auto_follow_pattern", func() error {
			_, err := ext.DeleteAutoFollowPatternWithParams(ctx, "n1", &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "DELETE", "/_ccr/auto_follow/n1", map[string]string{"master_timeout": "m1"}},
		{"get_auto_follow_pattern", func() error {
			_, err := ext.GetAutoFollowPatternWithParams(ctx, []string{"n1"}, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_ccr/auto_follow/n1", map[string]string{"master_timeout": "m1"}},
		{"get_auto_follow_pattern_all", func() error {
			_, err := ext.GetAutoFollowPatternWithParams(ctx, nil, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_ccr/auto_follow", map[string]string{"master_timeout": "m1"}},
		{"pause_auto_follow_pattern", func() error {
			_, err := ext.PauseAutoFollowPatternWithParams(ctx, "n1", &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/_ccr/auto_follow/n1/pause", map[string]string{"master_timeout": "m1"}},
		{"resume_auto_follow_pattern", func() error {
			_, err := ext.ResumeAutoFollowPatternWithParams(ctx, "n1", &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/_ccr/auto_follow/n1/resume", map[string]string{"master_timeout": "m1"}},
		{"pause_follow", func() error {
			_, err := ext.PauseFollowWithParams(ctx, "acme", &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/acme/_ccr/pause_follow", map[string]string{"master_timeout": "m1"}},
		{"resume_follow", func() error {
			_, err := ext.ResumeFollowWithRequest(ctx, &CcrResumeFollowRequest{Index: "acme", Body: map[string]any{"max_read_request_operation_count": 1024}, Params: &MasterTimeoutSpecParams{MasterTimeout: "m1"}})
			return err
		}, "POST", "/acme/_ccr/resume_follow", map[string]string{"master_timeout": "m1"}},
		{"unfollow", func() error {
			_, err := ext.UnfollowWithParams(ctx, "acme", &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/acme/_ccr/unfollow", map[string]string{"master_timeout": "m1"}},
		{"forget_follower", func() error {
			_, err := ext.ForgetFollowerWithRequest(ctx, &CcrForgetFollowerRequest{Index: "acme", Body: map[string]any{"follower_cluster": "remote"}, Params: &TimeoutSpecParams{Timeout: "t1"}})
			return err
		}, "POST", "/acme/_ccr/forget_follower", map[string]string{"timeout": "t1"}},
		{"follow_info", func() error {
			_, err := ext.FollowInfoWithParams(ctx, []string{"acme"}, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/acme/_ccr/info", map[string]string{"master_timeout": "m1"}},
		{"follow_info_all", func() error {
			_, err := ext.FollowInfoWithParams(ctx, nil, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_ccr/info", map[string]string{"master_timeout": "m1"}},
		{"follow_stats", func() error {
			_, err := ext.FollowStatsWithParams(ctx, []string{"acme"}, &TimeoutSpecParams{Timeout: "t1"})
			return err
		}, "GET", "/acme/_ccr/stats", map[string]string{"timeout": "t1"}},
		{"stats", func() error {
			_, err := ext.StatsWithParams(ctx, &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1", Timeout: "t1"})
			return err
		}, "GET", "/_ccr/stats", map[string]string{"master_timeout": "m1", "timeout": "t1"}},
		{"follow_direct_carrier", func() error {
			_, err := base.Follow(ctx, "follower", map[string]any{"leader_index": "leader"}, &CcrFollowParams{MasterTimeout: "m1", WaitForActiveShards: "1"})
			return err
		}, "POST", "/follower/_ccr/follow", map[string]string{"master_timeout": "m1", "wait_for_active_shards": "1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			*got = capturedSpecRequest{}
			require.NoError(t, tc.call())
			assert.Equal(t, tc.method, got.method)
			assert.Equal(t, tc.path, got.path)
			for k, v := range tc.query {
				assert.Equal(t, v, got.query[k], "query %s", k)
			}
		})
	}
}

func TestCcrSpecExtensionEscaping(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	ext, err := AsCcrServiceSpecExtensions(NewCcrService(restyClient(srv), testLogger()))
	require.NoError(t, err)
	_, err = ext.PauseFollowWithParams(context.Background(), "a/b?# c", nil)
	require.NoError(t, err)
	assert.Equal(t, "/a%2Fb%3F%23%20c/_ccr/pause_follow", got.escapedPath)
}

func TestSlmSpecExtensionsTransport(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	base := NewSlmService(restyClient(srv), testLogger())
	ext, err := AsSlmServiceSpecExtensions(base)
	require.NoError(t, err)
	ctx := context.Background()

	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
		query  map[string]string
	}{
		{"get_lifecycle", func() error {
			_, err := ext.GetLifecycleWithParams(ctx, []string{"p1"}, &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1", Timeout: "t1"})
			return err
		}, "GET", "/_slm/policy/p1", map[string]string{"master_timeout": "m1", "timeout": "t1"}},
		{"get_lifecycle_all", func() error {
			_, err := ext.GetLifecycleWithParams(ctx, nil, &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_slm/policy", map[string]string{"master_timeout": "m1"}},
		{"put_lifecycle", func() error {
			_, err := ext.PutLifecycleWithRequest(ctx, &SlmPutLifecycleRequest{PolicyID: "p1", Body: map[string]any{"schedule": "0 30 1 * * ?"}, Params: &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1", Timeout: "t1"}})
			return err
		}, "PUT", "/_slm/policy/p1", map[string]string{"master_timeout": "m1", "timeout": "t1"}},
		{"delete_lifecycle", func() error {
			_, err := ext.DeleteLifecycleWithParams(ctx, "p1", &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1", Timeout: "t1"})
			return err
		}, "DELETE", "/_slm/policy/p1", map[string]string{"master_timeout": "m1", "timeout": "t1"}},
		{"execute_lifecycle", func() error {
			_, err := ext.ExecuteLifecycleWithParams(ctx, "p1", &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1", Timeout: "t1"})
			return err
		}, "PUT", "/_slm/policy/p1/_execute", map[string]string{"master_timeout": "m1", "timeout": "t1"}},
		{"execute_retention", func() error {
			_, err := ext.ExecuteRetentionWithParams(ctx, &MasterTimeoutAndTimeoutSpecParams{Timeout: "t1"})
			return err
		}, "POST", "/_slm/_execute_retention", map[string]string{"timeout": "t1"}},
		{"get_stats", func() error {
			_, err := ext.GetStatsWithParams(ctx, &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_slm/stats", map[string]string{"master_timeout": "m1"}},
		{"get_status", func() error {
			_, err := ext.GetStatusWithParams(ctx, &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_slm/status", map[string]string{"master_timeout": "m1"}},
		{"start", func() error {
			_, err := ext.StartWithParams(ctx, &MasterTimeoutAndTimeoutSpecParams{Timeout: "t1"})
			return err
		}, "POST", "/_slm/start", map[string]string{"timeout": "t1"}},
		{"stop", func() error {
			_, err := ext.StopWithParams(ctx, &MasterTimeoutAndTimeoutSpecParams{Timeout: "t1"})
			return err
		}, "POST", "/_slm/stop", map[string]string{"timeout": "t1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			*got = capturedSpecRequest{}
			require.NoError(t, tc.call())
			assert.Equal(t, tc.method, got.method)
			assert.Equal(t, tc.path, got.path)
			for k, v := range tc.query {
				assert.Equal(t, v, got.query[k], "query %s", k)
			}
		})
	}
}

func TestIlmSpecExtensionsTransport(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	base := NewIlmService(restyClient(srv), testLogger())
	ext, err := AsIlmServiceSpecExtensions(base)
	require.NoError(t, err)
	ctx := context.Background()

	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
		query  map[string]string
	}{
		{"get_lifecycle", func() error {
			_, err := ext.GetLifecycleWithParams(ctx, []string{"p1"}, &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1", Timeout: "t1"})
			return err
		}, "GET", "/_ilm/policy/p1", map[string]string{"master_timeout": "m1", "timeout": "t1"}},
		{"put_lifecycle", func() error {
			_, err := ext.PutLifecycleWithRequest(ctx, &IlmPutLifecycleRequest{Policy: "p1", Body: map[string]any{"phases": map[string]any{}}, Params: &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1", Timeout: "t1"}})
			return err
		}, "PUT", "/_ilm/policy/p1", map[string]string{"master_timeout": "m1", "timeout": "t1"}},
		{"delete_lifecycle", func() error {
			_, err := ext.DeleteLifecycleWithParams(ctx, "p1", &MasterTimeoutAndTimeoutSpecParams{MasterTimeout: "m1", Timeout: "t1"})
			return err
		}, "DELETE", "/_ilm/policy/p1", map[string]string{"master_timeout": "m1", "timeout": "t1"}},
		{"start", func() error {
			_, err := ext.StartWithParams(ctx, &MasterTimeoutAndTimeoutSpecParams{Timeout: "t1"})
			return err
		}, "POST", "/_ilm/start", map[string]string{"timeout": "t1"}},
		{"stop", func() error {
			_, err := ext.StopWithParams(ctx, &MasterTimeoutAndTimeoutSpecParams{Timeout: "t1"})
			return err
		}, "POST", "/_ilm/stop", map[string]string{"timeout": "t1"}},
		{"migrate_to_data_tiers", func() error {
			_, err := ext.MigrateToDataTiersWithParams(ctx, ptrBool(true), &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/_ilm/migrate_to_data_tiers", map[string]string{"master_timeout": "m1", "dry_run": "true"}},
		{"migrate_to_data_tiers_false", func() error {
			_, err := ext.MigrateToDataTiersWithParams(ctx, ptrBool(false), &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/_ilm/migrate_to_data_tiers", map[string]string{"dry_run": "false"}},
		{"migrate_to_data_tiers_nil_dry_run", func() error {
			_, err := ext.MigrateToDataTiersWithParams(ctx, nil, nil)
			return err
		}, "POST", "/_ilm/migrate_to_data_tiers", map[string]string{}},
		{"explain_lifecycle", func() error {
			_, err := base.ExplainLifecycle(ctx, []string{"acme"}, &IlmExplainParams{MasterTimeout: "m1", OnlyErrors: ptrBool(true)})
			return err
		}, "GET", "/acme/_ilm/explain", map[string]string{"master_timeout": "m1", "only_errors": "true"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			*got = capturedSpecRequest{}
			require.NoError(t, tc.call())
			assert.Equal(t, tc.method, got.method)
			assert.Equal(t, tc.path, got.path)
			for k, v := range tc.query {
				assert.Equal(t, v, got.query[k], "query %s", k)
			}
		})
	}
}

func TestEnrichSpecExtensionsTransport(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	base := NewEnrichService(restyClient(srv), testLogger())
	ext, err := AsEnrichServiceSpecExtensions(base)
	require.NoError(t, err)
	ctx := context.Background()

	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
		query  map[string]string
	}{
		{"get_policy", func() error {
			_, err := ext.GetPolicyWithParams(ctx, []string{"p1"}, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_enrich/policy/p1", map[string]string{"master_timeout": "m1"}},
		{"put_policy", func() error {
			_, err := ext.PutPolicyWithRequest(ctx, &EnrichPutPolicyRequest{Name: "p1", Body: map[string]any{"match": map[string]any{}}, Params: &MasterTimeoutSpecParams{MasterTimeout: "m1"}})
			return err
		}, "PUT", "/_enrich/policy/p1", map[string]string{"master_timeout": "m1"}},
		{"delete_policy", func() error {
			_, err := ext.DeletePolicyWithParams(ctx, "p1", &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "DELETE", "/_enrich/policy/p1", map[string]string{"master_timeout": "m1"}},
		{"stats", func() error {
			_, err := ext.StatsWithParams(ctx, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_enrich/_stats", map[string]string{"master_timeout": "m1"}},
		{"execute_policy", func() error {
			_, err := base.ExecutePolicy(ctx, "p1", &EnrichWaitForCompletionParams{MasterTimeout: "m1", WaitForCompletion: ptrBool(false)})
			return err
		}, "PUT", "/_enrich/policy/p1/_execute", map[string]string{"master_timeout": "m1", "wait_for_completion": "false"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			*got = capturedSpecRequest{}
			require.NoError(t, tc.call())
			assert.Equal(t, tc.method, got.method)
			assert.Equal(t, tc.path, got.path)
			for k, v := range tc.query {
				assert.Equal(t, v, got.query[k], "query %s", k)
			}
		})
	}
}

func TestFeaturesWatcherXpackSpecExtensionsTransport(t *testing.T) {
	srv, got := captureSpecServer(t)
	defer srv.Close()
	featExt, err := AsFeaturesServiceSpecExtensions(NewFeaturesService(restyClient(srv), testLogger()))
	require.NoError(t, err)
	watExt, err := AsWatcherServiceSpecExtensions(NewWatcherService(restyClient(srv), testLogger()))
	require.NoError(t, err)
	xpExt, err := AsXpackServiceSpecExtensions(NewXpackService(restyClient(srv), testLogger()))
	require.NoError(t, err)
	ctx := context.Background()

	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
		query  map[string]string
	}{
		{"features_get", func() error {
			_, err := featExt.GetFeaturesWithParams(ctx, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_features", map[string]string{"master_timeout": "m1"}},
		{"features_reset", func() error {
			_, err := featExt.ResetFeaturesWithParams(ctx, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/_features/_reset", map[string]string{"master_timeout": "m1"}},
		{"watcher_start", func() error {
			_, err := watExt.StartWithParams(ctx, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/_watcher/_start", map[string]string{"master_timeout": "m1"}},
		{"watcher_stop", func() error {
			_, err := watExt.StopWithParams(ctx, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "POST", "/_watcher/_stop", map[string]string{"master_timeout": "m1"}},
		{"xpack_usage", func() error {
			_, err := xpExt.UsageWithParams(ctx, &MasterTimeoutSpecParams{MasterTimeout: "m1"})
			return err
		}, "GET", "/_xpack/usage", map[string]string{"master_timeout": "m1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			*got = capturedSpecRequest{}
			require.NoError(t, tc.call())
			assert.Equal(t, tc.method, got.method)
			assert.Equal(t, tc.path, got.path)
			for k, v := range tc.query {
				assert.Equal(t, v, got.query[k], "query %s", k)
			}
		})
	}
}
