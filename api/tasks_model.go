package api

import "github.com/disaster37/elasticsearch/v9/types"

// TasksListParams are the query parameters for GET /_tasks.
type TasksListParams struct {
	Common            *CommonParams
	Actions           []string
	Detailed          *bool
	GroupBy           GroupBy
	Nodes             []string
	ParentTaskId      string
	Timeout           string
	WaitForCompletion *bool
}

// ToMap converts TasksListParams to a query-parameter map.
func (p *TasksListParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setListParam(m, "actions", p.Actions)
	setBoolParam(m, "detailed", p.Detailed)
	if p.GroupBy != "" {
		m["group_by"] = string(p.GroupBy)
	}
	setListParam(m, "nodes", p.Nodes)
	if p.ParentTaskId != "" {
		m["parent_task_id"] = p.ParentTaskId
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// TasksGetParams are the query parameters for GET /_tasks/{task_id}.
type TasksGetParams struct {
	Common            *CommonParams
	Timeout           string
	WaitForCompletion *bool
}

// ToMap converts TasksGetParams to a query-parameter map.
func (p *TasksGetParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// TasksCancelParams are the query parameters for POST /_tasks/_cancel.
type TasksCancelParams struct {
	Common            *CommonParams
	Actions           []string
	Nodes             []string
	ParentTaskId      string
	WaitForCompletion *bool
}

// ToMap converts TasksCancelParams to a query-parameter map.
func (p *TasksCancelParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setListParam(m, "actions", p.Actions)
	setListParam(m, "nodes", p.Nodes)
	if p.ParentTaskId != "" {
		m["parent_task_id"] = p.ParentTaskId
	}
	setBoolParam(m, "wait_for_completion", p.WaitForCompletion)
	return mergeParams(m, p.Common)
}

// TasksListResponse represents the result of GET /_tasks.
type TasksListResponse struct {
	Nodes map[string]TaskNode `json:"nodes"`
}

// TaskNode describes a node and the tasks it is running.
type TaskNode struct {
	NodeID           string              `json:"node_id"`
	TransportAddress string              `json:"transport_address"`
	Host             string              `json:"host"`
	IP               string              `json:"ip"`
	Name             string              `json:"name"`
	Tasks            map[string]TaskInfo `json:"tasks"`
}

// TaskInfo describes a single task.
type TaskInfo struct {
	Node               string            `json:"node"`
	Id                 int64             `json:"id"`
	Type               string            `json:"type"`
	Action             string            `json:"action"`
	Status             map[string]any    `json:"status,omitempty"`
	Description        string            `json:"description,omitempty"`
	StartTimeInMillis  int64             `json:"start_time_in_millis"`
	RunningTimeInNanos int64             `json:"running_time_in_nanos"`
	Cancellable        bool              `json:"cancellable"`
	Cancelled          bool              `json:"cancelled,omitempty"`
	ParentTaskId       string            `json:"parent_task_id,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
}

// TaskGetResponse represents the result of GET /_tasks/{task_id}.
type TaskGetResponse struct {
	Completed bool                             `json:"completed"`
	Task      TaskInfo                         `json:"task"`
	Error     *types.ElasticsearchErrorDetails `json:"error,omitempty"`
	Response  map[string]any                   `json:"response,omitempty"`
}
