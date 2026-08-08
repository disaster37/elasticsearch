package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitTasksList(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"nodes":{"node1":{"node_id":"node1","transport_address":"ta","host":"h","ip":"127.0.0.1","name":"n","tasks":{"o1T...:1":{"node":"node1","id":1,"type":"transport","action":"action1","start_time_in_millis":100,"running_time_in_nanos":200,"cancellable":true}}}}}`)
	}))
	defer srv.Close()
	svc := NewTasksService(restyClient(srv), testLogger())
	resp, err := svc.List(context.Background(), &TasksListParams{
		Actions:           []string{"action1", "action2"},
		Detailed:           ptrBool(true),
		GroupBy:           GroupByNodes,
		Nodes:             []string{"node1"},
		ParentTaskId:      "parent1",
		Timeout:           "10s",
		WaitForCompletion: ptrBool(false),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_tasks", gotPath)
	assert.Equal(t, "action1,action2", gotQuery["actions"])
	assert.Equal(t, "true", gotQuery["detailed"])
	assert.Equal(t, "nodes", gotQuery["group_by"])
	assert.Equal(t, "node1", gotQuery["nodes"])
	assert.Equal(t, "parent1", gotQuery["parent_task_id"])
	assert.Equal(t, "10s", gotQuery["timeout"])
	assert.Equal(t, "false", gotQuery["wait_for_completion"])
	require.Contains(t, resp.Nodes, "node1")
	n := resp.Nodes["node1"]
	assert.Equal(t, "node1", n.NodeID)
	assert.Equal(t, "ta", n.TransportAddress)
	assert.Equal(t, "h", n.Host)
	assert.Equal(t, "127.0.0.1", n.IP)
	assert.Equal(t, "n", n.Name)
	require.Contains(t, n.Tasks, "o1T...:1")
	tk := n.Tasks["o1T...:1"]
	assert.Equal(t, "node1", tk.Node)
	assert.Equal(t, int64(1), tk.Id)
	assert.Equal(t, "transport", tk.Type)
	assert.Equal(t, "action1", tk.Action)
	assert.Equal(t, int64(100), tk.StartTimeInMillis)
	assert.Equal(t, int64(200), tk.RunningTimeInNanos)
	assert.True(t, tk.Cancellable)
}

func TestUnitTasksListNilParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, `{"nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewTasksService(restyClient(srv), testLogger())
	_, err := svc.List(context.Background(), nil)
	require.NoError(t, err)
}

func TestUnitTasksListError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTasksService(restyClient(srv), testLogger())
	_, err := svc.List(context.Background(), &TasksListParams{})
	assert.Error(t, err)
}

func TestUnitTasksListNetworkError(t *testing.T) {
	svc := NewTasksService(deadClient(), testLogger())
	_, err := svc.List(context.Background(), &TasksListParams{})
	assert.Error(t, err)
}

func TestUnitTasksGet(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"completed":false,"task":{"node":"node1","id":1,"type":"transport","action":"action1","cancellable":true}}`)
	}))
	defer srv.Close()
	svc := NewTasksService(restyClient(srv), testLogger())
	resp, err := svc.Get(context.Background(), "node1:1", &TasksGetParams{Timeout: "5s", WaitForCompletion: ptrBool(true)})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_tasks/node1:1", gotPath)
	assert.Equal(t, "5s", gotQuery["timeout"])
	assert.Equal(t, "true", gotQuery["wait_for_completion"])
	assert.False(t, resp.Completed)
	assert.Equal(t, "node1", resp.Task.Node)
	assert.Equal(t, int64(1), resp.Task.Id)
	assert.Equal(t, "action1", resp.Task.Action)
	assert.True(t, resp.Task.Cancellable)
}

func TestUnitTasksGetEmpty(t *testing.T) {
	svc := NewTasksService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.Get(context.Background(), "", &TasksGetParams{})
	assert.Error(t, err)
}

func TestUnitTasksGetError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTasksService(restyClient(srv), testLogger())
	_, err := svc.Get(context.Background(), "node1:1", &TasksGetParams{})
	assert.Error(t, err)
}

func TestUnitTasksGetNetworkError(t *testing.T) {
	svc := NewTasksService(deadClient(), testLogger())
	_, err := svc.Get(context.Background(), "node1:1", &TasksGetParams{})
	assert.Error(t, err)
}

func TestUnitTasksCancel(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewTasksService(restyClient(srv), testLogger())
	resp, err := svc.Cancel(context.Background(), "node1:1", &TasksCancelParams{
		Actions:           []string{"action1"},
		Nodes:             []string{"node1"},
		ParentTaskId:      "parent1",
		WaitForCompletion: ptrBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_tasks/node1:1/_cancel", gotPath)
	assert.Equal(t, "action1", gotQuery["actions"])
	assert.Equal(t, "node1", gotQuery["nodes"])
	assert.Equal(t, "parent1", gotQuery["parent_task_id"])
	assert.Equal(t, "true", gotQuery["wait_for_completion"])
	require.NotNil(t, resp)
}

func TestUnitTasksCancelAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"nodes":{}}`)
	}))
	defer srv.Close()
	svc := NewTasksService(restyClient(srv), testLogger())
	_, err := svc.Cancel(context.Background(), "", &TasksCancelParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_tasks/_cancel", gotPath)
}

func TestUnitTasksCancelError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewTasksService(restyClient(srv), testLogger())
	_, err := svc.Cancel(context.Background(), "node1:1", &TasksCancelParams{})
	assert.Error(t, err)
}

func TestUnitTasksCancelNetworkError(t *testing.T) {
	svc := NewTasksService(deadClient(), testLogger())
	_, err := svc.Cancel(context.Background(), "node1:1", &TasksCancelParams{})
	assert.Error(t, err)
}
