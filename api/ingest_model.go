package api

// IngestGetPipelineParams are the query parameters for GET /_ingest/pipeline.
type IngestGetPipelineParams struct {
	Common        *CommonParams
	MasterTimeout string
	Summary       *bool
}

// ToMap converts IngestGetPipelineParams to a query-parameter map.
func (p *IngestGetPipelineParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	setBoolParam(m, "summary", p.Summary)
	return mergeParams(m, p.Common)
}

// IngestPutPipelineRequest is the request for PUT /_ingest/pipeline/{id}.
type IngestPutPipelineRequest struct {
	Id     string `validate:"required"`
	Body   any    `validate:"required"`
	Params *IngestPutPipelineParams
}

// Validate validates the IngestPutPipelineRequest.
func (r *IngestPutPipelineRequest) Validate() error { return validationError(validate.Struct(r)) }

// IngestPutPipelineParams are the query parameters for put pipeline.
type IngestPutPipelineParams struct {
	Common        *CommonParams
	IfVersion     *int
	MasterTimeout string
	Timeout       string
}

// ToMap converts IngestPutPipelineParams to a query-parameter map.
func (p *IngestPutPipelineParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.IfVersion != nil {
		m["if_version"] = intStr(*p.IfVersion)
	}
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// IngestDeletePipelineParams are the query parameters for delete pipeline.
type IngestDeletePipelineParams struct {
	Common        *CommonParams
	MasterTimeout string
	Timeout       string
}

// ToMap converts IngestDeletePipelineParams to a query-parameter map.
func (p *IngestDeletePipelineParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	if p.MasterTimeout != "" {
		m["master_timeout"] = p.MasterTimeout
	}
	if p.Timeout != "" {
		m["timeout"] = p.Timeout
	}
	return mergeParams(m, p.Common)
}

// IngestSimulatePipelineRequest is the request for POST /_ingest/pipeline/_simulate.
type IngestSimulatePipelineRequest struct {
	Id     string
	Body   any `validate:"required"`
	Params *IngestSimulatePipelineParams
}

// Validate validates the IngestSimulatePipelineRequest.
func (r *IngestSimulatePipelineRequest) Validate() error { return validationError(validate.Struct(r)) }

// IngestSimulatePipelineParams are the query parameters for simulate.
type IngestSimulatePipelineParams struct {
	Common  *CommonParams
	Verbose *bool
}

// ToMap converts IngestSimulatePipelineParams to a query-parameter map.
func (p *IngestSimulatePipelineParams) ToMap() map[string]string {
	if p == nil {
		return nil
	}
	m := make(map[string]string)
	setBoolParam(m, "verbose", p.Verbose)
	return mergeParams(m, p.Common)
}

// IngestPipeline describes a stored ingest pipeline.
type IngestPipeline struct {
	Description string           `json:"description,omitempty"`
	Version     int              `json:"version,omitempty"`
	Processors  []map[string]any `json:"processors"`
	OnFailure   []map[string]any `json:"on_failure,omitempty"`
	Deprecated  bool             `json:"deprecated,omitempty"`
}

// IngestSimulateResponse is the response from simulate pipeline.
type IngestSimulateResponse struct {
	Docs []struct {
		Doc              map[string]any `json:"doc"`
		ProcessorResults []struct {
			Processor map[string]any `json:"processor_type"`
			Status    string         `json:"status"`
			Doc       map[string]any `json:"doc"`
			Tag       string         `json:"tag,omitempty"`
			Error     map[string]any `json:"error,omitempty"`
		} `json:"processor_results"`
	} `json:"docs"`
}

// IngestGrokResponse is the response from GET /_ingest/processor/grok.
type IngestGrokResponse struct {
	Patterns map[string]string `json:"patterns"`
}
