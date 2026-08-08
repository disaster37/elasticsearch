package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSnapshotCreateRepository(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.CreateRepository(context.Background(), "repo1", map[string]any{"type": "fs", "settings": map[string]any{"location": "/tmp"}}, &SnapshotCreateRepositoryParams{Timeout: "10s", Verify: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_snapshot/repo1", gotPath)
	assert.Equal(t, "10s", gotQuery["timeout"])
	assert.Equal(t, "true", gotQuery["verify"])
	assert.Contains(t, gotBody, "fs")
	assert.True(t, resp.Acknowledged)
}

func TestUnitSnapshotCreateRepositoryEmptyName(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.CreateRepository(context.Background(), "", map[string]any{"x": 1}, &SnapshotCreateRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotCreateRepositoryError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.CreateRepository(context.Background(), "repo", map[string]any{"x": 1}, &SnapshotCreateRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotCreateRepositoryNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.CreateRepository(context.Background(), "repo", map[string]any{"x": 1}, &SnapshotCreateRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotGetRepository(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"repo1":{"type":"fs","settings":{"location":"/tmp"},"uuid":"u1"}}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	result, err := svc.GetRepository(context.Background(), []string{"repo1"}, &SnapshotGetRepositoryParams{Local: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_snapshot/repo1", gotPath)
	assert.Equal(t, "true", gotQuery["local"])
	require.Contains(t, result, "repo1")
	r := result["repo1"]
	require.NotNil(t, r)
	assert.Equal(t, "fs", r.Type)
	assert.Equal(t, "/tmp", r.Settings["location"])
	assert.Equal(t, "u1", r.UUID)
}

func TestUnitSnapshotGetRepositoryAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.GetRepository(context.Background(), nil, &SnapshotGetRepositoryParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_snapshot", gotPath)
}

func TestUnitSnapshotGetRepositoryError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.GetRepository(context.Background(), nil, &SnapshotGetRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotGetRepositoryNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.GetRepository(context.Background(), nil, &SnapshotGetRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotDeleteRepository(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.DeleteRepository(context.Background(), []string{"repo1"}, &SnapshotDeleteRepositoryParams{Timeout: "5s"})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_snapshot/repo1", gotPath)
	assert.Equal(t, "5s", gotQuery["timeout"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitSnapshotDeleteRepositoryEmpty(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteRepository(context.Background(), nil, &SnapshotDeleteRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotDeleteRepositoryError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.DeleteRepository(context.Background(), []string{"repo"}, &SnapshotDeleteRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotDeleteRepositoryNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.DeleteRepository(context.Background(), []string{"repo"}, &SnapshotDeleteRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotVerifyRepository(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"nodes":{"node1":{"name":"node1"}}}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.VerifyRepository(context.Background(), "repo1", &SnapshotVerifyRepositoryParams{Timeout: "10s"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/_verify", gotPath)
	assert.Equal(t, "10s", gotQuery["timeout"])
	require.Contains(t, resp.Nodes, "node1")
	assert.Equal(t, "node1", resp.Nodes["node1"].Name)
}

func TestUnitSnapshotVerifyRepositoryEmptyName(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.VerifyRepository(context.Background(), "", &SnapshotVerifyRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotVerifyRepositoryError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.VerifyRepository(context.Background(), "repo", &SnapshotVerifyRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotVerifyRepositoryNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.VerifyRepository(context.Background(), "repo", &SnapshotVerifyRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotRepositoryAnalyze(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"blob_count":100,"concurrency":4,"read_node_count":2,"coordinating_node":"node1"}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.RepositoryAnalyze(context.Background(), "repo1", &SnapshotRepositoryAnalyzeParams{
		BlobCount: ptrInt(100), Concurrency: ptrInt(4), Detailed: ptrBool(true), MaxBlobSize: "1mb", MaxTotalDataSize: "10mb", Seed: ptrInt(42), Timeout: "30s",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/_analyze", gotPath)
	assert.Equal(t, "100", gotQuery["blob_count"])
	assert.Equal(t, "4", gotQuery["concurrency"])
	assert.Equal(t, "true", gotQuery["detailed"])
	assert.Equal(t, "1mb", gotQuery["max_blob_size"])
	assert.Equal(t, "10mb", gotQuery["max_total_data_size"])
	assert.Equal(t, "42", gotQuery["seed"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Equal(t, 100, resp.BlobCount)
	assert.Equal(t, 4, resp.Concurrency)
	assert.Equal(t, 2, resp.ReadNodeCount)
	assert.Equal(t, "node1", resp.CoordinatingNode)
}

func TestUnitSnapshotRepositoryAnalyzeEmptyName(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.RepositoryAnalyze(context.Background(), "", &SnapshotRepositoryAnalyzeParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotRepositoryAnalyzeError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.RepositoryAnalyze(context.Background(), "repo", &SnapshotRepositoryAnalyzeParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotRepositoryAnalyzeNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.RepositoryAnalyze(context.Background(), "repo", &SnapshotRepositoryAnalyzeParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotRepositoryVerifyIntegrity(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"repositories":[{"repository":"repo1","verified":true}]}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.RepositoryVerifyIntegrity(context.Background(), "repo1", &SnapshotRepositoryVerifyIntegrityParams{
		IndexSnapshotVerification: ptrBool(true), SnapshotVerification: ptrBool(false), Timeout: "30s",
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/_verify_integrity", gotPath)
	assert.Equal(t, "true", gotQuery["index_snapshot_verification"])
	assert.Equal(t, "false", gotQuery["snapshot_verification"])
	assert.Equal(t, "30s", gotQuery["timeout"])
	require.Len(t, resp.Repositories, 1)
	assert.Equal(t, "repo1", resp.Repositories[0].Repository)
	assert.True(t, resp.Repositories[0].Verified)
}

func TestUnitSnapshotRepositoryVerifyIntegrityEmptyName(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.RepositoryVerifyIntegrity(context.Background(), "", &SnapshotRepositoryVerifyIntegrityParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotRepositoryVerifyIntegrityError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.RepositoryVerifyIntegrity(context.Background(), "repo", &SnapshotRepositoryVerifyIntegrityParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotRepositoryVerifyIntegrityNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.RepositoryVerifyIntegrity(context.Background(), "repo", &SnapshotRepositoryVerifyIntegrityParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotCreate(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"snapshot":{"snapshot":"snap1","uuid":"u1","repository":"repo1","indices":["idx"],"state":"SUCCESS","start_time":"2020","start_time_in_millis":100,"end_time":"2020","end_time_in_millis":200,"duration_in_millis":100,"include_global_state":true}}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.Create(context.Background(), &SnapshotCreateRequest{
		Repository: "repo1",
		Snapshot:   "snap1",
		Body:       map[string]any{"indices": []string{"idx"}},
		Params:     &SnapshotCreateParams{WaitForCompletion: ptrBool(true)},
	})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/snap1", gotPath)
	assert.Equal(t, "true", gotQuery["wait_for_completion"])
	assert.Contains(t, gotBody, "idx")
	assert.Equal(t, "snap1", resp.Snapshot.Snapshot)
	assert.Equal(t, "u1", resp.Snapshot.UUID)
	assert.Equal(t, "repo1", resp.Snapshot.Repository)
	assert.Equal(t, []string{"idx"}, resp.Snapshot.Indices)
	assert.Equal(t, "SUCCESS", resp.Snapshot.State)
	assert.Equal(t, int64(100), resp.Snapshot.StartTimeInMillis)
	assert.Equal(t, int64(200), resp.Snapshot.EndTimeInMillis)
	assert.Equal(t, int64(100), resp.Snapshot.DurationInMillis)
	require.NotNil(t, resp.Snapshot.IncludeGlobalState)
	assert.True(t, *resp.Snapshot.IncludeGlobalState)
}

func TestUnitSnapshotCreateValidation(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	// missing repository
	_, err := svc.Create(context.Background(), &SnapshotCreateRequest{Snapshot: "s"})
	assert.Error(t, err)
	// missing snapshot
	_, err = svc.Create(context.Background(), &SnapshotCreateRequest{Repository: "r"})
	assert.Error(t, err)
}

func TestUnitSnapshotCreateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.Create(context.Background(), &SnapshotCreateRequest{Repository: "r", Snapshot: "s"})
	assert.Error(t, err)
}

func TestUnitSnapshotCreateNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.Create(context.Background(), &SnapshotCreateRequest{Repository: "r", Snapshot: "s"})
	assert.Error(t, err)
}

func TestUnitSnapshotGet(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"snapshots":[{"snapshot":"snap1","uuid":"u1","repository":"repo1","state":"SUCCESS"}],"total":1}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), &SnapshotGetRequest{
		Repository: "repo1",
		Snapshots:  []string{"snap1"},
		Params:     &SnapshotGetParams{IgnoreUnavailable: ptrBool(true), IncludeRepository: ptrBool(true), IndexDetails: ptrBool(true)},
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/snap1", gotPath)
	assert.Equal(t, "true", gotQuery["ignore_unavailable"])
	assert.Equal(t, "true", gotQuery["include_repository"])
	assert.Equal(t, "true", gotQuery["index_details"])
	require.Len(t, resp.Snapshots, 1)
	assert.Equal(t, "snap1", resp.Snapshots[0].Snapshot)
	assert.Equal(t, "SUCCESS", resp.Snapshots[0].State)
	assert.Equal(t, 1, resp.Total)
}

func TestUnitSnapshotGetAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"snapshots":[]}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), &SnapshotGetRequest{Repository: "repo1"})
	require.NoError(t, err)
	assert.Equal(t, "/_snapshot/repo1", gotPath)
}

func TestUnitSnapshotGetValidation(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Get(context.Background(), &SnapshotGetRequest{})
	assert.Error(t, err)
}

func TestUnitSnapshotGetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), &SnapshotGetRequest{Repository: "r"})
	assert.Error(t, err)
}

func TestUnitSnapshotGetNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), &SnapshotGetRequest{Repository: "r"})
	assert.Error(t, err)
}

func TestUnitSnapshotDelete(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.Delete(context.Background(), &SnapshotDeleteRequest{Repository: "repo1", Snapshot: "snap1"})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/snap1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSnapshotDeleteValidation(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Delete(context.Background(), &SnapshotDeleteRequest{Repository: "r"})
	assert.Error(t, err)
	_, err = svc.Delete(context.Background(), &SnapshotDeleteRequest{Snapshot: "s"})
	assert.Error(t, err)
}

func TestUnitSnapshotDeleteError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.Delete(context.Background(), &SnapshotDeleteRequest{Repository: "r", Snapshot: "s"})
	assert.Error(t, err)
}

func TestUnitSnapshotDeleteNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.Delete(context.Background(), &SnapshotDeleteRequest{Repository: "r", Snapshot: "s"})
	assert.Error(t, err)
}

func TestUnitSnapshotRestore(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"accepted":true,"snapshot":{"snapshot":"snap1","indices":["idx"]}}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.Restore(context.Background(), &SnapshotRestoreRequest{
		Repository: "repo1",
		Snapshot:   "snap1",
		Body:       map[string]any{"indices": []string{"idx"}},
		Params:     &SnapshotRestoreParams{WaitForCompletion: ptrBool(true)},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/snap1/_restore", gotPath)
	assert.Equal(t, "true", gotQuery["wait_for_completion"])
	assert.Contains(t, gotBody, "idx")
	assert.True(t, resp.Accepted)
	assert.Equal(t, "snap1", resp.Snapshot.Snapshot)
	assert.Equal(t, []string{"idx"}, resp.Snapshot.Indices)
}

func TestUnitSnapshotRestoreValidation(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Restore(context.Background(), &SnapshotRestoreRequest{Snapshot: "s"})
	assert.Error(t, err)
	_, err = svc.Restore(context.Background(), &SnapshotRestoreRequest{Repository: "r"})
	assert.Error(t, err)
}

func TestUnitSnapshotRestoreError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.Restore(context.Background(), &SnapshotRestoreRequest{Repository: "r", Snapshot: "s"})
	assert.Error(t, err)
}

func TestUnitSnapshotRestoreNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.Restore(context.Background(), &SnapshotRestoreRequest{Repository: "r", Snapshot: "s"})
	assert.Error(t, err)
}

func TestUnitSnapshotStatus(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"snapshots":[{"snapshot":"snap1","repository":"repo1","uuid":"u1","state":"SUCCESS"}]}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.Status(context.Background(), &SnapshotStatusRequest{
		Repository: "repo1",
		Snapshots:  []string{"snap1"},
		Params:     &SnapshotStatusParams{IgnoreUnavailable: ptrBool(true)},
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/snap1/_status", gotPath)
	assert.Equal(t, "true", gotQuery["ignore_unavailable"])
	require.Len(t, resp.Snapshots, 1)
	assert.Equal(t, "snap1", resp.Snapshots[0].Snapshot)
	assert.Equal(t, "repo1", resp.Snapshots[0].Repository)
	assert.Equal(t, "SUCCESS", resp.Snapshots[0].State)
}

func TestUnitSnapshotStatusRepoOnly(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"snapshots":[]}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.Status(context.Background(), &SnapshotStatusRequest{Repository: "repo1"})
	require.NoError(t, err)
	assert.Equal(t, "/_snapshot/repo1/_status", gotPath)
}

func TestUnitSnapshotStatusAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"snapshots":[]}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.Status(context.Background(), &SnapshotStatusRequest{})
	require.NoError(t, err)
	assert.Equal(t, "/_snapshot/_status", gotPath)
}

func TestUnitSnapshotStatusError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.Status(context.Background(), &SnapshotStatusRequest{})
	assert.Error(t, err)
}

func TestUnitSnapshotStatusNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.Status(context.Background(), &SnapshotStatusRequest{})
	assert.Error(t, err)
}

func TestUnitSnapshotClone(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.Clone(context.Background(), &SnapshotCloneRequest{
		Repository:     "repo1",
		Snapshot:       "snap1",
		TargetSnapshot: "snap2",
		Body:           map[string]any{"indices": []string{"idx"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/snap1/_clone/snap2", gotPath)
	assert.Contains(t, gotBody, "idx")
	assert.True(t, resp.Acknowledged)
}

func TestUnitSnapshotCloneValidation(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Clone(context.Background(), &SnapshotCloneRequest{Repository: "r", Snapshot: "s"})
	assert.Error(t, err)
	_, err = svc.Clone(context.Background(), &SnapshotCloneRequest{Repository: "r", Snapshot: "s", TargetSnapshot: "t"})
	// valid - should not error on validation (would error on network, but we use a server that does nothing)
	// Actually this will attempt a request; use deadClient to ensure validation passes
	_ = err
	svc2 := NewSnapshotService(deadClient(), testLogger())
	_, err = svc2.Clone(context.Background(), &SnapshotCloneRequest{Repository: "r", Snapshot: "s", TargetSnapshot: "t"})
	// network error, not validation error
	assert.Error(t, err)
}

func TestUnitSnapshotCloneError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.Clone(context.Background(), &SnapshotCloneRequest{Repository: "r", Snapshot: "s", TargetSnapshot: "t"})
	assert.Error(t, err)
}

func TestUnitSnapshotCloneNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.Clone(context.Background(), &SnapshotCloneRequest{Repository: "r", Snapshot: "s", TargetSnapshot: "t"})
	assert.Error(t, err)
}

func TestUnitSnapshotCleanupRepository(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"deleted_bytes":1024,"deleted_blobs":5}`)
	}))
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	resp, err := svc.CleanupRepository(context.Background(), "repo1", &SnapshotCleanupRepositoryParams{Timeout: "30s"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_snapshot/repo1/_cleanup", gotPath)
	assert.Equal(t, "30s", gotQuery["timeout"])
	assert.Equal(t, int64(1024), resp.DeletedBytes)
	assert.Equal(t, int64(5), resp.DeletedBlobs)
}

func TestUnitSnapshotCleanupRepositoryEmptyName(t *testing.T) {
	svc := NewSnapshotService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.CleanupRepository(context.Background(), "", &SnapshotCleanupRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotCleanupRepositoryError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSnapshotService(restyClient(srv), testLogger())
	_, err := svc.CleanupRepository(context.Background(), "repo", &SnapshotCleanupRepositoryParams{})
	assert.Error(t, err)
}

func TestUnitSnapshotCleanupRepositoryNetworkError(t *testing.T) {
	svc := NewSnapshotService(deadClient(), testLogger())
	_, err := svc.CleanupRepository(context.Background(), "repo", &SnapshotCleanupRepositoryParams{})
	assert.Error(t, err)
}
