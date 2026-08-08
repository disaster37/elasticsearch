package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitCcrPutAutoFollowPattern(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.PutAutoFollowPattern(context.Background(), "my-pattern", map[string]any{"remote_cluster": "remote"})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_ccr/auto_follow/my-pattern", gotPath)
	assert.Contains(t, gotBody, "remote_cluster")
	assert.True(t, resp.Acknowledged)
}

func TestUnitCcrPutAutoFollowPatternEmptyName(t *testing.T) {
	svc := NewCcrService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutAutoFollowPattern(context.Background(), "", map[string]any{})
	assert.Error(t, err)
}

func TestUnitCcrPutAutoFollowPatternError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.PutAutoFollowPattern(context.Background(), "my-pattern", map[string]any{})
	assert.Error(t, err)
}

func TestUnitCcrPutAutoFollowPatternNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.PutAutoFollowPattern(context.Background(), "my-pattern", map[string]any{})
	assert.Error(t, err)
}

func TestUnitCcrGetAutoFollowPattern(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"patterns":[{"name":"p1","pattern":{"remote_cluster":"remote"}}]}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.GetAutoFollowPattern(context.Background(), []string{"p1"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ccr/auto_follow/p1", gotPath)
	require.Len(t, resp.Patterns, 1)
	assert.Equal(t, "p1", resp.Patterns[0].Name)
}

func TestUnitCcrGetAutoFollowPatternAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"patterns":[]}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.GetAutoFollowPattern(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_ccr/auto_follow", gotPath)
}

func TestUnitCcrGetAutoFollowPatternError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.GetAutoFollowPattern(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitCcrGetAutoFollowPatternNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.GetAutoFollowPattern(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitCcrDeleteAutoFollowPattern(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.DeleteAutoFollowPattern(context.Background(), "my-pattern")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_ccr/auto_follow/my-pattern", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitCcrDeleteAutoFollowPatternEmptyName(t *testing.T) {
	svc := NewCcrService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteAutoFollowPattern(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitCcrDeleteAutoFollowPatternError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.DeleteAutoFollowPattern(context.Background(), "my-pattern")
	assert.Error(t, err)
}

func TestUnitCcrDeleteAutoFollowPatternNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.DeleteAutoFollowPattern(context.Background(), "my-pattern")
	assert.Error(t, err)
}

func TestUnitCcrPauseAutoFollowPattern(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.PauseAutoFollowPattern(context.Background(), "my-pattern")
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ccr/auto_follow/my-pattern/pause", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitCcrPauseAutoFollowPatternEmptyName(t *testing.T) {
	svc := NewCcrService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PauseAutoFollowPattern(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitCcrPauseAutoFollowPatternError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.PauseAutoFollowPattern(context.Background(), "my-pattern")
	assert.Error(t, err)
}

func TestUnitCcrPauseAutoFollowPatternNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.PauseAutoFollowPattern(context.Background(), "my-pattern")
	assert.Error(t, err)
}

func TestUnitCcrResumeAutoFollowPattern(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.ResumeAutoFollowPattern(context.Background(), "my-pattern")
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_ccr/auto_follow/my-pattern/resume", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitCcrResumeAutoFollowPatternEmptyName(t *testing.T) {
	svc := NewCcrService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ResumeAutoFollowPattern(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitCcrResumeAutoFollowPatternError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.ResumeAutoFollowPattern(context.Background(), "my-pattern")
	assert.Error(t, err)
}

func TestUnitCcrResumeAutoFollowPatternNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.ResumeAutoFollowPattern(context.Background(), "my-pattern")
	assert.Error(t, err)
}

func TestUnitCcrFollow(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"follow_index":"follower","shards_acknowledged":true,"follow_index_created":true,"follow_index_shards_acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.Follow(context.Background(), "follower", map[string]any{"leader_index": "leader"}, &CcrFollowParams{
		WaitForActiveShards: "1",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/follower/_ccr/follow", gotPath)
	assert.Equal(t, "1", gotQuery["wait_for_active_shards"])
	assert.Contains(t, gotBody, "leader_index")
	assert.Equal(t, "follower", resp.FollowIndex)
	assert.True(t, resp.ShardsAcknowledged)
	assert.True(t, resp.FollowIndexCreated)
	assert.True(t, resp.FollowIndexShardsAcknowledged)
}

func TestUnitCcrFollowEmptyIndex(t *testing.T) {
	svc := NewCcrService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Follow(context.Background(), "", map[string]any{}, nil)
	assert.Error(t, err)
}

func TestUnitCcrFollowError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.Follow(context.Background(), "follower", map[string]any{}, nil)
	assert.Error(t, err)
}

func TestUnitCcrFollowNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.Follow(context.Background(), "follower", map[string]any{}, nil)
	assert.Error(t, err)
}

func TestUnitCcrPauseFollow(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.PauseFollow(context.Background(), "follower")
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/follower/_ccr/pause_follow", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitCcrPauseFollowEmptyIndex(t *testing.T) {
	svc := NewCcrService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PauseFollow(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitCcrPauseFollowError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.PauseFollow(context.Background(), "follower")
	assert.Error(t, err)
}

func TestUnitCcrPauseFollowNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.PauseFollow(context.Background(), "follower")
	assert.Error(t, err)
}

func TestUnitCcrResumeFollow(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.ResumeFollow(context.Background(), "follower", map[string]any{"max_read_request_operation_count": 1024})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/follower/_ccr/resume_follow", gotPath)
	assert.Contains(t, gotBody, "max_read_request_operation_count")
	assert.True(t, resp.Acknowledged)
}

func TestUnitCcrResumeFollowEmptyIndex(t *testing.T) {
	svc := NewCcrService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ResumeFollow(context.Background(), "", map[string]any{})
	assert.Error(t, err)
}

func TestUnitCcrResumeFollowError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.ResumeFollow(context.Background(), "follower", map[string]any{})
	assert.Error(t, err)
}

func TestUnitCcrResumeFollowNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.ResumeFollow(context.Background(), "follower", map[string]any{})
	assert.Error(t, err)
}

func TestUnitCcrUnfollow(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.Unfollow(context.Background(), "follower")
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/follower/_ccr/unfollow", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitCcrUnfollowEmptyIndex(t *testing.T) {
	svc := NewCcrService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Unfollow(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitCcrUnfollowError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.Unfollow(context.Background(), "follower")
	assert.Error(t, err)
}

func TestUnitCcrUnfollowNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.Unfollow(context.Background(), "follower")
	assert.Error(t, err)
}

func TestUnitCcrForgetFollower(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"_shards":{"total":1,"successful":1,"failed":0}}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.ForgetFollower(context.Background(), "leader", map[string]any{"follower_cluster": "remote"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/leader/_ccr/forget_follower", gotPath)
	assert.Contains(t, gotBody, "follower_cluster")
	require.NotNil(t, resp.Shards)
	assert.Equal(t, 1, resp.Shards.Total)
	assert.Equal(t, 1, resp.Shards.Successful)
}

func TestUnitCcrForgetFollowerEmptyIndex(t *testing.T) {
	svc := NewCcrService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ForgetFollower(context.Background(), "", map[string]any{})
	assert.Error(t, err)
}

func TestUnitCcrForgetFollowerError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.ForgetFollower(context.Background(), "leader", map[string]any{})
	assert.Error(t, err)
}

func TestUnitCcrForgetFollowerNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.ForgetFollower(context.Background(), "leader", map[string]any{})
	assert.Error(t, err)
}

func TestUnitCcrFollowInfo(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"follower_indices":[{"index":"follower","follow_index":"follower","remote_cluster":"remote","leader_index":"leader"}]}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.FollowInfo(context.Background(), []string{"follower"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/follower/_ccr/info", gotPath)
	require.Len(t, resp.Indices, 1)
	assert.Equal(t, "follower", resp.Indices[0].Index)
	assert.Equal(t, "remote", resp.Indices[0].RemoteCluster)
}

func TestUnitCcrFollowInfoAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"follower_indices":[]}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.FollowInfo(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_ccr/info", gotPath)
}

func TestUnitCcrFollowInfoError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.FollowInfo(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitCcrFollowInfoNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.FollowInfo(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitCcrFollowStats(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"indices":[{"index":"follower","shards":{"total":1}}]}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.FollowStats(context.Background(), []string{"follower"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/follower/_ccr/stats", gotPath)
	require.Len(t, resp.Indices, 1)
	assert.Equal(t, "follower", resp.Indices[0].Index)
}

func TestUnitCcrFollowStatsAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"indices":[]}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.FollowStats(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_ccr/stats", gotPath)
}

func TestUnitCcrFollowStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.FollowStats(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitCcrFollowStatsNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.FollowStats(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitCcrStats(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"auto_follow_stats":{},"follow_stats":{}}`)
	}))
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	resp, err := svc.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ccr/stats", gotPath)
	assert.NotNil(t, resp.AutoFollowStats)
	assert.NotNil(t, resp.FollowStats)
}

func TestUnitCcrStatsError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewCcrService(restyClient(srv), testLogger())
	_, err := svc.Stats(context.Background())
	assert.Error(t, err)
}

func TestUnitCcrStatsNetworkError(t *testing.T) {
	svc := NewCcrService(deadClient(), testLogger())
	_, err := svc.Stats(context.Background())
	assert.Error(t, err)
}
