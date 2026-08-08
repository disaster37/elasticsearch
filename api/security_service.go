package api

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// SecurityService provides access to the security APIs. Covers users, roles,
// role mappings, privileges, API keys, tokens, authentication, and service
// accounts.
type SecurityService interface {
	GetUser(ctx context.Context, usernames []string) (map[string]*SecurityUser, error)
	PutUser(ctx context.Context, username string, body any) (*SecurityCreateResponse, error)
	DeleteUser(ctx context.Context, username string) (*types.AcknowledgedResponse, error)
	ChangePassword(ctx context.Context, username string, body any) (*types.AcknowledgedResponse, error)
	DisableUser(ctx context.Context, username string) (*types.AcknowledgedResponse, error)
	EnableUser(ctx context.Context, username string) (*types.AcknowledgedResponse, error)
	GetRole(ctx context.Context, names []string) (map[string]*SecurityRole, error)
	PutRole(ctx context.Context, name string, body any) (*SecurityCreateResponse, error)
	DeleteRole(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
	ClearRoleCache(ctx context.Context, names []string) (*types.AcknowledgedResponse, error)
	GetRoleMapping(ctx context.Context, names []string) (map[string]*SecurityRoleMapping, error)
	PutRoleMapping(ctx context.Context, name string, body any) (*SecurityRoleMappingCreateResponse, error)
	DeleteRoleMapping(ctx context.Context, name string) (*types.AcknowledgedResponse, error)
	GetPrivileges(ctx context.Context, application, name string) (map[string]any, error)
	PutPrivileges(ctx context.Context, body any) (*SecurityPutPrivilegesResponse, error)
	DeletePrivileges(ctx context.Context, application, name string) (*types.AcknowledgedResponse, error)
	HasPrivileges(ctx context.Context, body any) (*SecurityHasPrivilegesResponse, error)
	GetUserPrivileges(ctx context.Context) (map[string]any, error)
	CreateAPIKey(ctx context.Context, body any) (*SecurityCreateAPIKeyResponse, error)
	GetAPIKey(ctx context.Context, params *SecurityGetAPIKeyParams) (*SecurityGetAPIKeyResponse, error)
	InvalidateAPIKey(ctx context.Context, body any) (*SecurityInvalidateAPIKeyResponse, error)
	UpdateAPIKey(ctx context.Context, id string, body any) (*SecurityUpdateAPIKeyResponse, error)
	Authenticate(ctx context.Context) (*SecurityAuthenticateResponse, error)
	ClearRealmCache(ctx context.Context, realms []string, params *SecurityClearRealmCacheParams) (*types.AcknowledgedResponse, error)
	GetToken(ctx context.Context, body any) (*SecurityGetTokenResponse, error)
	InvalidateToken(ctx context.Context, body any) (*SecurityInvalidateTokenResponse, error)
}

// DefaultSecurityService is the default SecurityService implementation.
type DefaultSecurityService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewSecurityService creates a new SecurityService.
func NewSecurityService(client *resty.Client, logger *logrus.Entry) SecurityService {
	return &DefaultSecurityService{client: client, logger: logger.WithField("service", "security")}
}

// ---- Params ----

// SecurityGetAPIKeyParams are the query parameters for get API key.
type SecurityGetAPIKeyParams struct {
	Common                *CommonParams
	ID                    string
	Name                  string
	RealmName             string
	Username              string
	Owner                 *bool
	ActiveOnly            *bool
	WithLimitedPrivileges *bool
	WithProfileUid        *bool
}

// ToMap converts SecurityGetAPIKeyParams to a query-parameter map.
func (p *SecurityGetAPIKeyParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.ID != "" {
		m["id"] = p.ID
	}
	if p.Name != "" {
		m["name"] = p.Name
	}
	if p.RealmName != "" {
		m["realm_name"] = p.RealmName
	}
	if p.Username != "" {
		m["username"] = p.Username
	}
	setBoolParam(m, "owner", p.Owner)
	setBoolParam(m, "active_only", p.ActiveOnly)
	setBoolParam(m, "with_limited_privileges", p.WithLimitedPrivileges)
	setBoolParam(m, "with_profile_uid", p.WithProfileUid)
	return mergeParams(m, p.Common)
}

// SecurityClearRealmCacheParams are the query parameters for clear realm cache.
type SecurityClearRealmCacheParams struct {
	Common    *CommonParams
	Usernames []string
}

// ToMap converts SecurityClearRealmCacheParams to a query-parameter map.
func (p *SecurityClearRealmCacheParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setListParam(m, "usernames", p.Usernames)
	return mergeParams(m, p.Common)
}

// ---- Response structs ----

// SecurityUser describes a security user.
type SecurityUser struct {
	Username string         `json:"username"`
	Roles    []string       `json:"roles"`
	FullName string         `json:"full_name,omitempty"`
	Email    string         `json:"email,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Enabled  bool           `json:"enabled"`
}

// SecurityRole describes a security role.
type SecurityRole struct {
	Cluster           []string                  `json:"cluster,omitempty"`
	Indices           []SecurityRoleIndex       `json:"indices,omitempty"`
	Applications      []SecurityRoleApplication `json:"applications,omitempty"`
	RunAs             []string                  `json:"run_as,omitempty"`
	Metadata          map[string]any            `json:"metadata,omitempty"`
	TransientMetadata map[string]any            `json:"transient_metadata,omitempty"`
}

// SecurityRoleIndex describes per-index role permissions.
type SecurityRoleIndex struct {
	Names                  []string       `json:"names"`
	Privileges             []string       `json:"privileges"`
	FieldSecurity          map[string]any `json:"field_security,omitempty"`
	Query                  string         `json:"query,omitempty"`
	AllowRestrictedIndices *bool          `json:"allow_restricted_indices,omitempty"`
}

// SecurityRoleApplication describes application role permissions.
type SecurityRoleApplication struct {
	Application string   `json:"application"`
	Privileges  []string `json:"privileges"`
	Resources   []string `json:"resources"`
}

// SecurityRoleMapping describes a role mapping.
type SecurityRoleMapping struct {
	Enabled       bool             `json:"enabled"`
	Roles         []string         `json:"roles,omitempty"`
	RoleTemplates []map[string]any `json:"role_templates,omitempty"`
	Rules         map[string]any   `json:"rules,omitempty"`
	Metadata      map[string]any   `json:"metadata,omitempty"`
}

// SecurityCreateResponse is the response from create user/role.
type SecurityCreateResponse struct {
	Created bool `json:"created"`
}

// SecurityRoleMappingCreateResponse is the response from create role mapping.
type SecurityRoleMappingCreateResponse struct {
	Created bool `json:"created"`
}

// SecurityPutPrivilegesResponse is the response from put privileges.
type SecurityPutPrivilegesResponse struct {
	Created map[string]map[string]bool `json:"created"`
}

// SecurityHasPrivilegesResponse is the response from has privileges.
type SecurityHasPrivilegesResponse struct {
	Username     string         `json:"username,omitempty"`
	HasPrivilege map[string]any `json:"has_privilege"`
}

// SecurityCreateAPIKeyResponse is the response from create API key.
type SecurityCreateAPIKeyResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	APIKey     string `json:"api_key,omitempty"`
	Expiration int64  `json:"expiration,omitempty"`
	Encoded    string `json:"encoded,omitempty"`
}

// SecurityGetAPIKeyResponse is the response from get API key.
type SecurityGetAPIKeyResponse struct {
	APIKeys []SecurityAPIKey `json:"api_keys"`
}

// SecurityAPIKey describes an API key.
type SecurityAPIKey struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        string           `json:"type,omitempty"`
	Creation    int64            `json:"creation,omitempty"`
	Expiration  int64            `json:"expiration,omitempty"`
	Invalidated bool             `json:"invalidated,omitempty"`
	Username    string           `json:"username,omitempty"`
	Realm       string           `json:"realm,omitempty"`
	Descriptors []map[string]any `json:"role_descriptors,omitempty"`
}

// SecurityInvalidateAPIKeyResponse is the response from invalidate API key.
type SecurityInvalidateAPIKeyResponse struct {
	InvalidatedAPIKeys           []string         `json:"invalidated_api_keys"`
	PreviouslyInvalidatedAPIKeys []string         `json:"previously_invalidated_api_keys"`
	Errors                       []map[string]any `json:"errors,omitempty"`
}

// SecurityUpdateAPIKeyResponse is the response from update API key.
type SecurityUpdateAPIKeyResponse struct {
	Updated bool `json:"updated"`
}

// SecurityAuthenticateResponse is the response from authenticate.
type SecurityAuthenticateResponse struct {
	Username            string         `json:"username"`
	Roles               []string       `json:"roles"`
	FullName            string         `json:"full_name,omitempty"`
	Email               string         `json:"email,omitempty"`
	Metadata            map[string]any `json:"metadata,omitempty"`
	AuthenticationRealm struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"authentication_realm"`
	LookupRealm struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"lookup_realm"`
	AuthenticationType string `json:"authentication_type"`
}

// SecurityGetTokenResponse is the response from get token.
type SecurityGetTokenResponse struct {
	AccessToken  string `json:"access_token"`
	Type         string `json:"type"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// SecurityInvalidateTokenResponse is the response from invalidate token.
type SecurityInvalidateTokenResponse struct {
	CreatedTokens               []string `json:"created_tokens,omitempty"`
	InvalidatedTokens           int      `json:"invalidated_tokens"`
	PreviouslyInvalidatedTokens int      `json:"previously_invalidated_tokens"`
	ErrorsCount                 int      `json:"errors_count,omitempty"`
}

// ---- Implementations ----

func (s *DefaultSecurityService) GetUser(ctx context.Context, usernames []string) (map[string]*SecurityUser, error) {
	var path string
	if len(usernames) > 0 {
		path = fmt.Sprintf("/_security/user/%s", escapedList(usernames))
	} else {
		path = "/_security/user"
	}
	var result map[string]*SecurityUser
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultSecurityService) PutUser(ctx context.Context, username string, body any) (*SecurityCreateResponse, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	var result SecurityCreateResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/user/%s", pathSegment(username)), body, nil, &result)
}

func (s *DefaultSecurityService) DeleteUser(ctx context.Context, username string) (*types.AcknowledgedResponse, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_security/user/%s", pathSegment(username)), nil, nil, &result)
}

func (s *DefaultSecurityService) ChangePassword(ctx context.Context, username string, body any) (*types.AcknowledgedResponse, error) {
	var path string
	if username != "" {
		path = fmt.Sprintf("/_security/user/%s/_password", pathSegment(username))
	} else {
		path = "/_security/user/_password"
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", path, body, nil, &result)
}

func (s *DefaultSecurityService) DisableUser(ctx context.Context, username string) (*types.AcknowledgedResponse, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/user/%s/_disable", pathSegment(username)), nil, nil, &result)
}

func (s *DefaultSecurityService) EnableUser(ctx context.Context, username string) (*types.AcknowledgedResponse, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/user/%s/_enable", pathSegment(username)), nil, nil, &result)
}

func (s *DefaultSecurityService) GetRole(ctx context.Context, names []string) (map[string]*SecurityRole, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_security/role/%s", escapedList(names))
	} else {
		path = "/_security/role"
	}
	var result map[string]*SecurityRole
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultSecurityService) PutRole(ctx context.Context, name string, body any) (*SecurityCreateResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result SecurityCreateResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/role/%s", pathSegment(name)), body, nil, &result)
}

func (s *DefaultSecurityService) DeleteRole(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_security/role/%s", pathSegment(name)), nil, nil, &result)
}

func (s *DefaultSecurityService) ClearRoleCache(ctx context.Context, names []string) (*types.AcknowledgedResponse, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("names is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", fmt.Sprintf("/_security/role/%s/_clear_cache", escapedList(names)), nil, nil, &result)
}

func (s *DefaultSecurityService) GetRoleMapping(ctx context.Context, names []string) (map[string]*SecurityRoleMapping, error) {
	var path string
	if len(names) > 0 {
		path = fmt.Sprintf("/_security/role_mapping/%s", escapedList(names))
	} else {
		path = "/_security/role_mapping"
	}
	var result map[string]*SecurityRoleMapping
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultSecurityService) PutRoleMapping(ctx context.Context, name string, body any) (*SecurityRoleMappingCreateResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result SecurityRoleMappingCreateResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/role_mapping/%s", pathSegment(name)), body, nil, &result)
}

func (s *DefaultSecurityService) DeleteRoleMapping(ctx context.Context, name string) (*types.AcknowledgedResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_security/role_mapping/%s", pathSegment(name)), nil, nil, &result)
}

func (s *DefaultSecurityService) GetPrivileges(ctx context.Context, application, name string) (map[string]any, error) {
	var path string
	switch {
	case application != "" && name != "":
		path = fmt.Sprintf("/_security/privilege/%s/%s", pathSegment(application), pathSegment(name))
	case application != "":
		path = fmt.Sprintf("/_security/privilege/%s", pathSegment(application))
	default:
		path = "/_security/privilege"
	}
	var result map[string]any
	return result, doReq(s.logger, ctx, s.client, "GET", path, nil, nil, &result)
}

func (s *DefaultSecurityService) PutPrivileges(ctx context.Context, body any) (*SecurityPutPrivilegesResponse, error) {
	var result SecurityPutPrivilegesResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", "/_security/privilege", body, nil, &result)
}

func (s *DefaultSecurityService) DeletePrivileges(ctx context.Context, application, name string) (*types.AcknowledgedResponse, error) {
	if application == "" || name == "" {
		return nil, fmt.Errorf("application and name are required")
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", fmt.Sprintf("/_security/privilege/%s/%s", pathSegment(application), pathSegment(name)), nil, nil, &result)
}

func (s *DefaultSecurityService) HasPrivileges(ctx context.Context, body any) (*SecurityHasPrivilegesResponse, error) {
	var result SecurityHasPrivilegesResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_security/user/_has_privileges", body, nil, &result)
}

func (s *DefaultSecurityService) GetUserPrivileges(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	return result, doReq(s.logger, ctx, s.client, "GET", "/_security/user/_privileges", nil, nil, &result)
}

func (s *DefaultSecurityService) CreateAPIKey(ctx context.Context, body any) (*SecurityCreateAPIKeyResponse, error) {
	var result SecurityCreateAPIKeyResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_security/api_key", body, nil, &result)
}

func (s *DefaultSecurityService) GetAPIKey(ctx context.Context, params *SecurityGetAPIKeyParams) (*SecurityGetAPIKeyResponse, error) {
	var result SecurityGetAPIKeyResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_security/api_key", nil, params, &result)
}

func (s *DefaultSecurityService) InvalidateAPIKey(ctx context.Context, body any) (*SecurityInvalidateAPIKeyResponse, error) {
	var result SecurityInvalidateAPIKeyResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", "/_security/api_key", body, nil, &result)
}

func (s *DefaultSecurityService) UpdateAPIKey(ctx context.Context, id string, body any) (*SecurityUpdateAPIKeyResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	var result SecurityUpdateAPIKeyResponse
	return &result, doReq(s.logger, ctx, s.client, "PUT", fmt.Sprintf("/_security/api_key/%s", pathSegment(id)), body, nil, &result)
}

func (s *DefaultSecurityService) Authenticate(ctx context.Context) (*SecurityAuthenticateResponse, error) {
	var result SecurityAuthenticateResponse
	return &result, doReq(s.logger, ctx, s.client, "GET", "/_security/_authenticate", nil, nil, &result)
}

func (s *DefaultSecurityService) ClearRealmCache(ctx context.Context, realms []string, params *SecurityClearRealmCacheParams) (*types.AcknowledgedResponse, error) {
	var path string
	if len(realms) > 0 {
		path = fmt.Sprintf("/_security/realm/%s/_clear_cache", escapedList(realms))
	} else {
		path = "/_security/realm/_clear_cache"
	}
	var result types.AcknowledgedResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", path, nil, params, &result)
}

func (s *DefaultSecurityService) GetToken(ctx context.Context, body any) (*SecurityGetTokenResponse, error) {
	var result SecurityGetTokenResponse
	return &result, doReq(s.logger, ctx, s.client, "POST", "/_security/oauth2/token", body, nil, &result)
}

func (s *DefaultSecurityService) InvalidateToken(ctx context.Context, body any) (*SecurityInvalidateTokenResponse, error) {
	var result SecurityInvalidateTokenResponse
	return &result, doReq(s.logger, ctx, s.client, "DELETE", "/_security/oauth2/token", body, nil, &result)
}

// ensure json import is used.
var _ = json.RawMessage{}
