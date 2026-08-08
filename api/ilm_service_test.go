package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitIlmGetLifecycle(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"policy1":{"version":1,"modified_date":"2020","modified_date_millis":100,"in_use":true}}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	result, err := svc.GetLifecycle(context.Background(), []string{"policy1"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ilm/policy/policy1", gotPath)
	require.Contains(t, result, "policy1")
	p := result["policy1"]
	require.NotNil(t, p)
	assert.Equal(t, 1, p.Version)
	assert.Equal(t, "2020", p.ModifiedDate)
	assert.Equal(t, int64(100), p.ModifiedDateMillis)
	assert.True(t, p.InUse)
}

func TestUnitIlmGetLifecycleAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.GetLifecycle(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_ilm/policy", gotPath)
}

func TestUnitIlmGetLifecycleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.GetLifecycle(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitIlmGetLifecycleNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.GetLifecycle(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitIlmPutLifecycle(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	resp, err := svc.PutLifecycle(context.Background(), "policy1", map[string]any{"phases": map[string]any{}})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_ilm/policy/policy1", gotPath)
	assert.Contains(t, gotBody, "phases")
	assert.True(t, resp.Acknowledged)
}

func TestUnitIlmPutLifecycleEmpty(t *testing.T) {
	svc := NewIlmService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutLifecycle(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitIlmPutLifecycleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.PutLifecycle(context.Background(), "policy1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitIlmPutLifecycleNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.PutLifecycle(context.Background(), "policy1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitIlmDeleteLifecycle(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	resp, err := svc.DeleteLifecycle(context.Background(), "policy1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_ilm/policy/policy1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIlmDeleteLifecycleEmpty(t *testing.T) {
	svc := NewIlmService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteLifecycle(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitIlmDeleteLifecycleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.DeleteLifecycle(context.Background(), "policy1")
	assert.Error(t, err)
}

func TestUnitIlmDeleteLifecycleNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.DeleteLifecycle(context.Background(), "policy1")
	assert.Error(t, err)
}

func TestUnitIlmGetStatus(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"operation_mode":"RUNNING"}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	resp, err := svc.GetStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ilm/status", gotPath)
	assert.Equal(t, "RUNNING", resp.OperationMode)
}

func TestUnitIlmGetStatusError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.GetStatus(context.Background())
	assert.Error(t, err)
}

func TestUnitIlmGetStatusNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.GetStatus(context.Background())
	assert.Error(t, err)
}

func TestUnitIlmStart(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	resp, err := svc.Start(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ilm/start", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIlmStartError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.Start(context.Background())
	assert.Error(t, err)
}

func TestUnitIlmStartNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.Start(context.Background())
	assert.Error(t, err)
}

func TestUnitIlmStop(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	resp, err := svc.Stop(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ilm/stop", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIlmStopError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.Stop(context.Background())
	assert.Error(t, err)
}

func TestUnitIlmStopNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.Stop(context.Background())
	assert.Error(t, err)
}

func TestUnitIlmExplainLifecycle(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"indices":{"idx1":{"index":"idx1","managed":true,"phase":"hot","action":"rollover","step":"check-rollover-ready","step_time_millis":100,"phase_time_millis":200,"action_time_millis":300}}}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	resp, err := svc.ExplainLifecycle(context.Background(), []string{"idx1"}, &IlmExplainParams{OnlyErrors: ptrBool(true), OnlyManaged: ptrBool(false)})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/idx1/_ilm/explain", gotPath)
	assert.Equal(t, "true", gotQuery["only_errors"])
	assert.Equal(t, "false", gotQuery["only_managed"])
	require.Contains(t, resp.Indices, "idx1")
	ix := resp.Indices["idx1"]
	assert.Equal(t, "idx1", ix.Index)
	assert.True(t, ix.Managed)
	assert.Equal(t, "hot", ix.Phase)
	assert.Equal(t, "rollover", ix.Action)
	assert.Equal(t, "check-rollover-ready", ix.Step)
	assert.Equal(t, int64(100), ix.StepTime)
	assert.Equal(t, int64(200), ix.PhaseTime)
	assert.Equal(t, int64(300), ix.ActionTime)
}

func TestUnitIlmExplainLifecycleEmpty(t *testing.T) {
	svc := NewIlmService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ExplainLifecycle(context.Background(), nil, &IlmExplainParams{})
	assert.Error(t, err)
}

func TestUnitIlmExplainLifecycleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.ExplainLifecycle(context.Background(), []string{"idx1"}, &IlmExplainParams{})
	assert.Error(t, err)
}

func TestUnitIlmExplainLifecycleNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.ExplainLifecycle(context.Background(), []string{"idx1"}, &IlmExplainParams{})
	assert.Error(t, err)
}

func TestUnitIlmMoveToStep(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	resp, err := svc.MoveToStep(context.Background(), "idx1", map[string]any{"current_step": map[string]any{}})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ilm/move/idx1", gotPath)
	assert.Contains(t, gotBody, "current_step")
	assert.True(t, resp.Acknowledged)
}

func TestUnitIlmMoveToStepEmpty(t *testing.T) {
	svc := NewIlmService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.MoveToStep(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitIlmMoveToStepError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.MoveToStep(context.Background(), "idx1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitIlmMoveToStepNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.MoveToStep(context.Background(), "idx1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitIlmRemovePolicy(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"has_failures":true,"failed_indexes":["idx2"]}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	resp, err := svc.RemovePolicy(context.Background(), []string{"idx1"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx1/_ilm/remove", gotPath)
	assert.True(t, resp.HasFailures)
	assert.Equal(t, []string{"idx2"}, resp.FailedIndexes)
}

func TestUnitIlmRemovePolicyEmpty(t *testing.T) {
	svc := NewIlmService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.RemovePolicy(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitIlmRemovePolicyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.RemovePolicy(context.Background(), []string{"idx1"})
	assert.Error(t, err)
}

func TestUnitIlmRemovePolicyNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.RemovePolicy(context.Background(), []string{"idx1"})
	assert.Error(t, err)
}

func TestUnitIlmRetry(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	resp, err := svc.Retry(context.Background(), []string{"idx1"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/idx1/_ilm/retry", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitIlmRetryEmpty(t *testing.T) {
	svc := NewIlmService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Retry(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitIlmRetryError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.Retry(context.Background(), []string{"idx1"})
	assert.Error(t, err)
}

func TestUnitIlmRetryNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.Retry(context.Background(), []string{"idx1"})
	assert.Error(t, err)
}

func TestUnitIlmMigrateToDataTiers(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"dry_run":true,"migrated_component_templates":["ct1"],"migrated_composable_templates":["cct1"],"migrated_indices":["idx1"],"migrated_ilm_policies":["p1"],"migrated_legacy_templates":["lt1"],"removed_legacy_template":["rlt1"],"removed_index_settings":["ris1"],"removed_cluster_settings":["rcs1"]}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	resp, err := svc.MigrateToDataTiers(context.Background(), ptrBool(true))
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ilm/migrate_to_data_tiers", gotPath)
	assert.Equal(t, "true", gotQuery["dry_run"])
	assert.True(t, resp.DryRun)
	assert.Equal(t, []string{"ct1"}, resp.MigratedComponentTemplates)
	assert.Equal(t, []string{"cct1"}, resp.MigratedComposableTemplates)
	assert.Equal(t, []string{"idx1"}, resp.MigratedIndices)
	assert.Equal(t, []string{"p1"}, resp.MigratedIlmPolicies)
	assert.Equal(t, []string{"lt1"}, resp.MigratedLegacyTemplates)
	assert.Equal(t, []string{"rlt1"}, resp.RemovedLegacyTemplate)
	assert.Equal(t, []string{"ris1"}, resp.RemovedIndexSettings)
	assert.Equal(t, []string{"rcs1"}, resp.RemovedClusterSettings)
}

func TestUnitIlmMigrateToDataTiersNilDryRun(t *testing.T) {
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"dry_run":false}`)
	}))
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.MigrateToDataTiers(context.Background(), nil)
	require.NoError(t, err)
	_, ok := gotQuery["dry_run"]
	assert.False(t, ok)
}

func TestUnitIlmMigrateToDataTiersError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewIlmService(restyClient(srv), testLogger())
	_, err := svc.MigrateToDataTiers(context.Background(), ptrBool(true))
	assert.Error(t, err)
}

func TestUnitIlmMigrateToDataTiersNetworkError(t *testing.T) {
	svc := NewIlmService(deadClient(), testLogger())
	_, err := svc.MigrateToDataTiers(context.Background(), ptrBool(true))
	assert.Error(t, err)
}
