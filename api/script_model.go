package api

// ScriptGetParams are the query parameters for GET /_scripts/{id}.
type ScriptGetParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts ScriptGetParams to a query-parameter map.
func (p *ScriptGetParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ScriptPutRequest is the request for PUT|POST /_scripts/{id}.
type ScriptPutRequest struct {
	Id     string `validate:"required"`
	Body   any    `validate:"required"`
	Params *ScriptPutParams
}

// Validate validates the ScriptPutRequest.
func (r *ScriptPutRequest) Validate() error { return validationError(validate.Struct(r)) }

// ScriptPutParams are the query parameters for PUT /_scripts/{id}.
type ScriptPutParams struct {
	Common  *CommonParams
	Context string
	Timeout string
}

// ToMap converts ScriptPutParams to a query-parameter map.
func (p *ScriptPutParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Context != "" {
		m["context"] = p.Context
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ScriptDeleteParams are the query parameters for DELETE /_scripts/{id}.
type ScriptDeleteParams struct {
	Common  *CommonParams
	Timeout string
}

// ToMap converts ScriptDeleteParams to a query-parameter map.
func (p *ScriptDeleteParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// ScriptGetResponse is the response from GET /_scripts/{id}.
type ScriptGetResponse struct {
	Id     string  `json:"_id,omitempty"`
	Found  bool    `json:"found"`
	Script *Script `json:"script,omitempty"`
}

// Script is the stored script body.
type Script struct {
	Lang    string         `json:"lang,omitempty"`
	Source  string         `json:"source,omitempty"`
	Options map[string]any `json:"options,omitempty"`
}

// ScriptContextResponse is the response from GET /_script_context.
type ScriptContextResponse struct {
	Contexts []ScriptContextInfo `json:"contexts"`
}

// ScriptContextInfo describes a script context.
type ScriptContextInfo struct {
	Context string                `json:"context"`
	Methods []ScriptContextMethod `json:"methods"`
}

// ScriptContextMethod describes a method available in a script context.
type ScriptContextMethod struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
}

// ScriptLanguagesResponse is the response from GET /_script_language.
type ScriptLanguagesResponse struct {
	TypesAllowed     []string                `json:"types_allowed"`
	LanguageContexts []ScriptLanguageContext `json:"language_contexts"`
}

// ScriptLanguageContext describes a language and its contexts.
type ScriptLanguageContext struct {
	Language string   `json:"language"`
	Contexts []string `json:"contexts"`
}
