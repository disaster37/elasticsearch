package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// GetUser
// ---------------------------------------------------------------------------

func TestUnitSecurityGetUser(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"u1":{"username":"u1","roles":["r1"],"full_name":"fn","email":"e","enabled":true}}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.GetUser(context.Background(), []string{"u1"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_security/user/u1", gotPath)
	require.Contains(t, resp, "u1")
	u := resp["u1"]
	assert.Equal(t, "u1", u.Username)
	assert.Equal(t, []string{"r1"}, u.Roles)
	assert.Equal(t, "fn", u.FullName)
	assert.Equal(t, "e", u.Email)
	assert.True(t, u.Enabled)
}

func TestUnitSecurityGetUserAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetUser(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_security/user", gotPath)
}

func TestUnitSecurityGetUserError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetUser(context.Background(), []string{"u1"})
	assert.Error(t, err)
}

func TestUnitSecurityGetUserNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.GetUser(context.Background(), []string{"u1"})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// PutUser
// ---------------------------------------------------------------------------

func TestUnitSecurityPutUser(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"created":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.PutUser(context.Background(), "u1", map[string]any{"password": "p"})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_security/user/u1", gotPath)
	assert.Contains(t, gotBody, "password")
	assert.True(t, resp.Created)
}

func TestUnitSecurityPutUserEmpty(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutUser(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityPutUserError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.PutUser(context.Background(), "u1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityPutUserNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.PutUser(context.Background(), "u1", map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// DeleteUser
// ---------------------------------------------------------------------------

func TestUnitSecurityDeleteUser(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.DeleteUser(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_security/user/u1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSecurityDeleteUserEmpty(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteUser(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSecurityDeleteUserError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.DeleteUser(context.Background(), "u1")
	assert.Error(t, err)
}

func TestUnitSecurityDeleteUserNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.DeleteUser(context.Background(), "u1")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ChangePassword
// ---------------------------------------------------------------------------

func TestUnitSecurityChangePassword(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.ChangePassword(context.Background(), "u1", map[string]any{"password": "p"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_security/user/u1/_password", gotPath)
	assert.Contains(t, gotBody, "password")
	assert.True(t, resp.Acknowledged)
}

func TestUnitSecurityChangePasswordNoUser(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.ChangePassword(context.Background(), "", map[string]any{"password": "p"})
	require.NoError(t, err)
	assert.Equal(t, "/_security/user/_password", gotPath)
}

func TestUnitSecurityChangePasswordError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.ChangePassword(context.Background(), "u1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityChangePasswordNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.ChangePassword(context.Background(), "u1", map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// DisableUser
// ---------------------------------------------------------------------------

func TestUnitSecurityDisableUser(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.DisableUser(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_security/user/u1/_disable", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSecurityDisableUserEmpty(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DisableUser(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSecurityDisableUserError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.DisableUser(context.Background(), "u1")
	assert.Error(t, err)
}

func TestUnitSecurityDisableUserNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.DisableUser(context.Background(), "u1")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// EnableUser
// ---------------------------------------------------------------------------

func TestUnitSecurityEnableUser(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.EnableUser(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_security/user/u1/_enable", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSecurityEnableUserEmpty(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.EnableUser(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSecurityEnableUserError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.EnableUser(context.Background(), "u1")
	assert.Error(t, err)
}

func TestUnitSecurityEnableUserNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.EnableUser(context.Background(), "u1")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetRole
// ---------------------------------------------------------------------------

func TestUnitSecurityGetRole(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"r1":{"cluster":["all"],"run_as":["u1"]}}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.GetRole(context.Background(), []string{"r1"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_security/role/r1", gotPath)
	require.Contains(t, resp, "r1")
	assert.Equal(t, []string{"all"}, resp["r1"].Cluster)
	assert.Equal(t, []string{"u1"}, resp["r1"].RunAs)
}

func TestUnitSecurityGetRoleAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetRole(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_security/role", gotPath)
}

func TestUnitSecurityGetRoleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetRole(context.Background(), []string{"r1"})
	assert.Error(t, err)
}

func TestUnitSecurityGetRoleNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.GetRole(context.Background(), []string{"r1"})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// PutRole
// ---------------------------------------------------------------------------

func TestUnitSecurityPutRole(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"created":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.PutRole(context.Background(), "r1", map[string]any{"cluster": []string{"all"}})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_security/role/r1", gotPath)
	assert.Contains(t, gotBody, "cluster")
	assert.True(t, resp.Created)
}

func TestUnitSecurityPutRoleEmpty(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutRole(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityPutRoleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.PutRole(context.Background(), "r1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityPutRoleNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.PutRole(context.Background(), "r1", map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// DeleteRole
// ---------------------------------------------------------------------------

func TestUnitSecurityDeleteRole(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.DeleteRole(context.Background(), "r1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_security/role/r1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSecurityDeleteRoleEmpty(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteRole(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSecurityDeleteRoleError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.DeleteRole(context.Background(), "r1")
	assert.Error(t, err)
}

func TestUnitSecurityDeleteRoleNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.DeleteRole(context.Background(), "r1")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ClearRoleCache
// ---------------------------------------------------------------------------

func TestUnitSecurityClearRoleCache(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.ClearRoleCache(context.Background(), []string{"r1", "r2"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_security/role/r1,r2/_clear_cache", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSecurityClearRoleCacheEmpty(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.ClearRoleCache(context.Background(), nil)
	assert.Error(t, err)
}

func TestUnitSecurityClearRoleCacheError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.ClearRoleCache(context.Background(), []string{"r1"})
	assert.Error(t, err)
}

func TestUnitSecurityClearRoleCacheNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.ClearRoleCache(context.Background(), []string{"r1"})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetRoleMapping
// ---------------------------------------------------------------------------

func TestUnitSecurityGetRoleMapping(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"rm1":{"enabled":true,"roles":["r1"]}}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.GetRoleMapping(context.Background(), []string{"rm1"})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_security/role_mapping/rm1", gotPath)
	require.Contains(t, resp, "rm1")
	assert.True(t, resp["rm1"].Enabled)
	assert.Equal(t, []string{"r1"}, resp["rm1"].Roles)
}

func TestUnitSecurityGetRoleMappingAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetRoleMapping(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/_security/role_mapping", gotPath)
}

func TestUnitSecurityGetRoleMappingError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetRoleMapping(context.Background(), []string{"rm1"})
	assert.Error(t, err)
}

func TestUnitSecurityGetRoleMappingNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.GetRoleMapping(context.Background(), []string{"rm1"})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// PutRoleMapping
// ---------------------------------------------------------------------------

func TestUnitSecurityPutRoleMapping(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"created":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.PutRoleMapping(context.Background(), "rm1", map[string]any{"roles": []string{"r1"}})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_security/role_mapping/rm1", gotPath)
	assert.Contains(t, gotBody, "roles")
	assert.True(t, resp.Created)
}

func TestUnitSecurityPutRoleMappingEmpty(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.PutRoleMapping(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityPutRoleMappingError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.PutRoleMapping(context.Background(), "rm1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityPutRoleMappingNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.PutRoleMapping(context.Background(), "rm1", map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// DeleteRoleMapping
// ---------------------------------------------------------------------------

func TestUnitSecurityDeleteRoleMapping(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.DeleteRoleMapping(context.Background(), "rm1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_security/role_mapping/rm1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSecurityDeleteRoleMappingEmpty(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeleteRoleMapping(context.Background(), "")
	assert.Error(t, err)
}

func TestUnitSecurityDeleteRoleMappingError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.DeleteRoleMapping(context.Background(), "rm1")
	assert.Error(t, err)
}

func TestUnitSecurityDeleteRoleMappingNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.DeleteRoleMapping(context.Background(), "rm1")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetPrivileges
// ---------------------------------------------------------------------------

func TestUnitSecurityGetPrivileges(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"app1":{"p1":{}}}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.GetPrivileges(context.Background(), "app1", "p1")
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_security/privilege/app1/p1", gotPath)
	require.Contains(t, resp, "app1")
}

func TestUnitSecurityGetPrivilegesAppOnly(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"app1":{}}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetPrivileges(context.Background(), "app1", "")
	require.NoError(t, err)
	assert.Equal(t, "/_security/privilege/app1", gotPath)
}

func TestUnitSecurityGetPrivilegesAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetPrivileges(context.Background(), "", "")
	require.NoError(t, err)
	assert.Equal(t, "/_security/privilege", gotPath)
}

func TestUnitSecurityGetPrivilegesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetPrivileges(context.Background(), "app1", "p1")
	assert.Error(t, err)
}

func TestUnitSecurityGetPrivilegesNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.GetPrivileges(context.Background(), "app1", "p1")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// PutPrivileges
// ---------------------------------------------------------------------------

func TestUnitSecurityPutPrivileges(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"created":{"app1":{"p1":true}}}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.PutPrivileges(context.Background(), map[string]any{"app1": map[string]any{"p1": map[string]any{}}})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_security/privilege", gotPath)
	assert.Contains(t, gotBody, "app1")
	require.Contains(t, resp.Created, "app1")
	assert.True(t, resp.Created["app1"]["p1"])
}

func TestUnitSecurityPutPrivilegesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.PutPrivileges(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityPutPrivilegesNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.PutPrivileges(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// DeletePrivileges
// ---------------------------------------------------------------------------

func TestUnitSecurityDeletePrivileges(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.DeletePrivileges(context.Background(), "app1", "p1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_security/privilege/app1/p1", gotPath)
	assert.True(t, resp.Acknowledged)
}

func TestUnitSecurityDeletePrivilegesEmptyApp(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeletePrivileges(context.Background(), "", "p1")
	assert.Error(t, err)
}

func TestUnitSecurityDeletePrivilegesEmptyName(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.DeletePrivileges(context.Background(), "app1", "")
	assert.Error(t, err)
}

func TestUnitSecurityDeletePrivilegesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.DeletePrivileges(context.Background(), "app1", "p1")
	assert.Error(t, err)
}

func TestUnitSecurityDeletePrivilegesNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.DeletePrivileges(context.Background(), "app1", "p1")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// HasPrivileges
// ---------------------------------------------------------------------------

func TestUnitSecurityHasPrivileges(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"username":"u1","has_privilege":{"app1":{"p1":true}}}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.HasPrivileges(context.Background(), map[string]any{"application": "app1"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_security/user/_has_privileges", gotPath)
	assert.Contains(t, gotBody, "application")
	assert.Equal(t, "u1", resp.Username)
	require.Contains(t, resp.HasPrivilege, "app1")
}

func TestUnitSecurityHasPrivilegesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.HasPrivileges(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityHasPrivilegesNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.HasPrivileges(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetUserPrivileges
// ---------------------------------------------------------------------------

func TestUnitSecurityGetUserPrivileges(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"cluster":["all"]}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.GetUserPrivileges(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_security/user/_privileges", gotPath)
	require.Contains(t, resp, "cluster")
}

func TestUnitSecurityGetUserPrivilegesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetUserPrivileges(context.Background())
	assert.Error(t, err)
}

func TestUnitSecurityGetUserPrivilegesNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.GetUserPrivileges(context.Background())
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// CreateAPIKey
// ---------------------------------------------------------------------------

func TestUnitSecurityCreateAPIKey(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"id":"k1","name":"n","api_key":"secret","expiration":123,"encoded":"base64"}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.CreateAPIKey(context.Background(), map[string]any{"name": "n"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_security/api_key", gotPath)
	assert.Contains(t, gotBody, "name")
	assert.Equal(t, "k1", resp.ID)
	assert.Equal(t, "n", resp.Name)
	assert.Equal(t, "secret", resp.APIKey)
	assert.Equal(t, int64(123), resp.Expiration)
	assert.Equal(t, "base64", resp.Encoded)
}

func TestUnitSecurityCreateAPIKeyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.CreateAPIKey(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityCreateAPIKeyNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.CreateAPIKey(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetAPIKey
// ---------------------------------------------------------------------------

func TestUnitSecurityGetAPIKey(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"api_keys":[{"id":"k1","name":"n","creation":123}]}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.GetAPIKey(context.Background(), &SecurityGetAPIKeyParams{
		ID:                    "k1",
		Name:                  "n",
		RealmName:             "r",
		Username:              "u1",
		Owner:                 ptrBool(true),
		ActiveOnly:            ptrBool(false),
		WithLimitedPrivileges: ptrBool(true),
		WithProfileUid:        ptrBool(false),
	})
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_security/api_key", gotPath)
	assert.Equal(t, "k1", gotQuery["id"])
	assert.Equal(t, "n", gotQuery["name"])
	assert.Equal(t, "r", gotQuery["realm_name"])
	assert.Equal(t, "u1", gotQuery["username"])
	assert.Equal(t, "true", gotQuery["owner"])
	assert.Equal(t, "false", gotQuery["active_only"])
	assert.Equal(t, "true", gotQuery["with_limited_privileges"])
	assert.Equal(t, "false", gotQuery["with_profile_uid"])
	require.Len(t, resp.APIKeys, 1)
	assert.Equal(t, "k1", resp.APIKeys[0].ID)
	assert.Equal(t, "n", resp.APIKeys[0].Name)
	assert.Equal(t, int64(123), resp.APIKeys[0].Creation)
}

func TestUnitSecurityGetAPIKeyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetAPIKey(context.Background(), &SecurityGetAPIKeyParams{})
	assert.Error(t, err)
}

func TestUnitSecurityGetAPIKeyNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.GetAPIKey(context.Background(), &SecurityGetAPIKeyParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// InvalidateAPIKey
// ---------------------------------------------------------------------------

func TestUnitSecurityInvalidateAPIKey(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"invalidated_api_keys":["k1"],"previously_invalidated_api_keys":["k2"]}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.InvalidateAPIKey(context.Background(), map[string]any{"id": "k1"})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_security/api_key", gotPath)
	assert.Contains(t, gotBody, "id")
	assert.Equal(t, []string{"k1"}, resp.InvalidatedAPIKeys)
	assert.Equal(t, []string{"k2"}, resp.PreviouslyInvalidatedAPIKeys)
}

func TestUnitSecurityInvalidateAPIKeyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.InvalidateAPIKey(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityInvalidateAPIKeyNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.InvalidateAPIKey(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// UpdateAPIKey
// ---------------------------------------------------------------------------

func TestUnitSecurityUpdateAPIKey(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"updated":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.UpdateAPIKey(context.Background(), "k1", map[string]any{"role_descriptors": map[string]any{}})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/_security/api_key/k1", gotPath)
	assert.Contains(t, gotBody, "role_descriptors")
	assert.True(t, resp.Updated)
}

func TestUnitSecurityUpdateAPIKeyEmpty(t *testing.T) {
	svc := NewSecurityService(restyClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))), testLogger())
	_, err := svc.UpdateAPIKey(context.Background(), "", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityUpdateAPIKeyError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.UpdateAPIKey(context.Background(), "k1", map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityUpdateAPIKeyNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.UpdateAPIKey(context.Background(), "k1", map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Authenticate
// ---------------------------------------------------------------------------

func TestUnitSecurityAuthenticate(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `{"username":"u1","roles":["r1"],"full_name":"fn","email":"e","authentication_realm":{"name":"rn","type":"rt"},"lookup_realm":{"name":"ln","type":"lt"},"authentication_type":"realm"}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.Authenticate(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_security/_authenticate", gotPath)
	assert.Equal(t, "u1", resp.Username)
	assert.Equal(t, []string{"r1"}, resp.Roles)
	assert.Equal(t, "fn", resp.FullName)
	assert.Equal(t, "e", resp.Email)
	assert.Equal(t, "rn", resp.AuthenticationRealm.Name)
	assert.Equal(t, "rt", resp.AuthenticationRealm.Type)
	assert.Equal(t, "ln", resp.LookupRealm.Name)
	assert.Equal(t, "lt", resp.LookupRealm.Type)
	assert.Equal(t, "realm", resp.AuthenticationType)
}

func TestUnitSecurityAuthenticateError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.Authenticate(context.Background())
	assert.Error(t, err)
}

func TestUnitSecurityAuthenticateNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.Authenticate(context.Background())
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ClearRealmCache
// ---------------------------------------------------------------------------

func TestUnitSecurityClearRealmCache(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = queryMap(r)
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.ClearRealmCache(context.Background(), []string{"r1", "r2"}, &SecurityClearRealmCacheParams{
		Usernames: []string{"u1", "u2"},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_security/realm/r1,r2/_clear_cache", gotPath)
	assert.Equal(t, "u1,u2", gotQuery["usernames"])
	assert.True(t, resp.Acknowledged)
}

func TestUnitSecurityClearRealmCacheAll(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		jsonResponse(w, 200, `{"acknowledged":true}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.ClearRealmCache(context.Background(), nil, &SecurityClearRealmCacheParams{})
	require.NoError(t, err)
	assert.Equal(t, "/_security/realm/_clear_cache", gotPath)
}

func TestUnitSecurityClearRealmCacheError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.ClearRealmCache(context.Background(), []string{"r1"}, &SecurityClearRealmCacheParams{})
	assert.Error(t, err)
}

func TestUnitSecurityClearRealmCacheNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.ClearRealmCache(context.Background(), []string{"r1"}, &SecurityClearRealmCacheParams{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetToken
// ---------------------------------------------------------------------------

func TestUnitSecurityGetToken(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"access_token":"at","type":"Bearer","expires_in":1200,"refresh_token":"rt","scope":"s"}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.GetToken(context.Background(), map[string]any{"grant_type": "password"})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/_security/oauth2/token", gotPath)
	assert.Contains(t, gotBody, "grant_type")
	assert.Equal(t, "at", resp.AccessToken)
	assert.Equal(t, "Bearer", resp.Type)
	assert.Equal(t, 1200, resp.ExpiresIn)
	assert.Equal(t, "rt", resp.RefreshToken)
	assert.Equal(t, "s", resp.Scope)
}

func TestUnitSecurityGetTokenError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.GetToken(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityGetTokenNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.GetToken(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// InvalidateToken
// ---------------------------------------------------------------------------

func TestUnitSecurityInvalidateToken(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody = readBody(r)
		jsonResponse(w, 200, `{"created_tokens":["t1"],"invalidated_tokens":2,"previously_invalidated_tokens":1,"errors_count":0}`)
	}))
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	resp, err := svc.InvalidateToken(context.Background(), map[string]any{"token": "t1"})
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/_security/oauth2/token", gotPath)
	assert.Contains(t, gotBody, "token")
	assert.Equal(t, []string{"t1"}, resp.CreatedTokens)
	assert.Equal(t, 2, resp.InvalidatedTokens)
	assert.Equal(t, 1, resp.PreviouslyInvalidatedTokens)
	assert.Equal(t, 0, resp.ErrorsCount)
}

func TestUnitSecurityInvalidateTokenError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSecurityService(restyClient(srv), testLogger())
	_, err := svc.InvalidateToken(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}

func TestUnitSecurityInvalidateTokenNetworkError(t *testing.T) {
	svc := NewSecurityService(deadClient(), testLogger())
	_, err := svc.InvalidateToken(context.Background(), map[string]any{"x": 1})
	assert.Error(t, err)
}
