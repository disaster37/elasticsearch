package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/disaster37/elasticsearch/v9/internal/speccheck/spec"
)

// MissingEndpoint is a reviewed baseline record for an unimplemented endpoint.
type MissingEndpoint struct {
	Status     string // "not-implemented" or "experimental"
	Reason     string // concrete missing service/method; never merely "skip"
	Stability  string // exact pinned metadata, empty when absent
	Visibility string // exact pinned metadata
	Tracking   string // README backlog anchor or issue URL
}

// IgnoredParam is a narrowly justified, client-controlled query exception.
type IgnoredParam struct {
	Reason          string
	ControlledValue string
	Evidence        string
}

// Policy is the reviewed set of accepted debt and controlled exceptions.
type Policy struct {
	KnownMissingEndpoints map[string]MissingEndpoint
	IgnoredParams         map[string]map[string]IgnoredParam
	FeatureFlagAllowlist  map[string]bool
}

// AuditResult summarizes one audit run.
type AuditResult struct {
	Checked, Implemented, KnownMissing, Ignored int
	MissingEndpoints, ParamGaps, PathGaps       []string
	PolicyErrors, IntegrityErrors               []string
}

// defaultPolicy returns the reviewed production policy.
func defaultPolicy() Policy {
	policy := Policy{
		KnownMissingEndpoints: knownMissingEndpoints,
		IgnoredParams:         map[string]map[string]IgnoredParam{},
		FeatureFlagAllowlist: map[string]bool{
			"streams.logs_enable":  true,
			"streams.logs_disable": true,
			"streams.status":       true,
		},
	}
	for _, name := range catFormatEndpoints {
		policy.IgnoredParams[name] = map[string]IgnoredParam{
			"format": {
				Reason:          "typed JSON decoding requires format=json",
				ControlledValue: "json",
				Evidence:        "catCommonParams.toMap always sets format=json; covered by cat HTTP regression tests",
			},
		}
	}
	return policy
}

// audit compares the catalog against the registry and policy. Registry and
// policy integrity are validated first, and all diagnostics are sorted.
func audit(catalog spec.Catalog, registry map[string]EndpointImpl, policy Policy) AuditResult {
	var res AuditResult
	res.IntegrityErrors = append(res.IntegrityErrors, validateRegistry(registry)...)
	res.PolicyErrors = append(res.PolicyErrors, validatePolicy(catalog, registry, policy)...)

	names := checkedNames(catalog, policy.FeatureFlagAllowlist)
	res.Checked = len(names)
	for _, name := range names {
		ep := catalog.Endpoints[name]
		impl, ok := registry[name]
		if !ok {
			if _, known := policy.KnownMissingEndpoints[name]; known {
				res.KnownMissing++
			} else {
				res.MissingEndpoints = append(res.MissingEndpoints, name)
			}
			continue
		}
		res.Implemented++

		params := sortedParamNames(ep.Endpoint.Params)
		for _, param := range params {
			if ign, ok := policy.IgnoredParams[name][param]; ok {
				if err := validateControlledIgnore(impl, param, ign); err != nil {
					res.IntegrityErrors = append(res.IntegrityErrors, fmt.Sprintf("%s param %q: %v", name, param, err))
				}
				res.Ignored++
				continue
			}
			if _, ok := impl.QueryArgs[param]; ok {
				continue
			}
			field := queryField(impl, param)
			if impl.ParamsStruct != "" && contains(impl.ParamsFields, field) {
				if err := validateParamSerialization(impl, param, field); err != nil {
					res.IntegrityErrors = append(res.IntegrityErrors, fmt.Sprintf("%s param %q: %v", name, param, err))
				}
				continue
			}
			res.ParamGaps = append(res.ParamGaps, fmt.Sprintf("%s: param %q missing from %s", name, param, paramsLabel(impl)))
		}

		for _, part := range spec.CollectParts(ep.Endpoint.URL.Paths) {
			if _, ok := impl.PathSources[part]; ok {
				continue
			}
			res.PathGaps = append(res.PathGaps, fmt.Sprintf("%s: path part %q missing from %s.%s", name, part, impl.Service, impl.Method))
		}
	}

	sort.Strings(res.MissingEndpoints)
	sort.Strings(res.ParamGaps)
	sort.Strings(res.PathGaps)
	sort.Strings(res.PolicyErrors)
	sort.Strings(res.IntegrityErrors)
	return res
}

// validatePolicy checks the reviewed policy against the catalog and registry.
func validatePolicy(catalog spec.Catalog, registry map[string]EndpointImpl, policy Policy) []string {
	var errs []string
	allow := policy.FeatureFlagAllowlist

	baselineNames := make([]string, 0, len(policy.KnownMissingEndpoints))
	for name := range policy.KnownMissingEndpoints {
		baselineNames = append(baselineNames, name)
	}
	sort.Strings(baselineNames)
	for _, name := range baselineNames {
		entry := policy.KnownMissingEndpoints[name]
		ep, ok := catalog.Endpoints[name]
		if !ok {
			errs = append(errs, fmt.Sprintf("baseline %q: absent from catalog", name))
			continue
		}
		if !spec.Checked(ep, allow) {
			errs = append(errs, fmt.Sprintf("baseline %q: excluded or private in catalog", name))
		}
		if _, ok := registry[name]; ok {
			errs = append(errs, fmt.Sprintf("baseline %q: endpoint is also registered", name))
		}
		if entry.Stability != ep.Endpoint.Stability {
			errs = append(errs, fmt.Sprintf("baseline %q: stability %q != catalog %q", name, entry.Stability, ep.Endpoint.Stability))
		}
		if entry.Visibility != ep.Endpoint.Visibility {
			errs = append(errs, fmt.Sprintf("baseline %q: visibility %q != catalog %q", name, entry.Visibility, ep.Endpoint.Visibility))
		}
		if entry.Status != "not-implemented" && entry.Status != "experimental" {
			errs = append(errs, fmt.Sprintf("baseline %q: invalid status %q", name, entry.Status))
		}
		if strings.TrimSpace(entry.Reason) == "" {
			errs = append(errs, fmt.Sprintf("baseline %q: missing reason", name))
		}
		if strings.TrimSpace(entry.Tracking) == "" {
			errs = append(errs, fmt.Sprintf("baseline %q: missing tracking", name))
		}
	}

	ignoreEndpoints := make([]string, 0, len(policy.IgnoredParams))
	for name := range policy.IgnoredParams {
		ignoreEndpoints = append(ignoreEndpoints, name)
	}
	sort.Strings(ignoreEndpoints)
	for _, name := range ignoreEndpoints {
		impl, registered := registry[name]
		ep, inCatalog := catalog.Endpoints[name]
		if !registered || !inCatalog || !spec.Checked(ep, allow) {
			errs = append(errs, fmt.Sprintf("ignored param endpoint %q: not a checked registered endpoint", name))
			continue
		}
		wires := make([]string, 0, len(policy.IgnoredParams[name]))
		for wire := range policy.IgnoredParams[name] {
			wires = append(wires, wire)
		}
		sort.Strings(wires)
		for _, wire := range wires {
			ign := policy.IgnoredParams[name][wire]
			if _, ok := ep.Endpoint.Params[wire]; !ok {
				errs = append(errs, fmt.Sprintf("ignored param %s.%s: not a spec param", name, wire))
				continue
			}
			if _, ok := impl.QueryArgs[wire]; ok {
				errs = append(errs, fmt.Sprintf("ignored param %s.%s: already covered positionally", name, wire))
			} else if impl.ParamsStruct != "" && contains(impl.ParamsFields, queryField(impl, wire)) {
				errs = append(errs, fmt.Sprintf("ignored param %s.%s: already exposed as a field", name, wire))
			}
			if strings.TrimSpace(ign.Reason) == "" || strings.TrimSpace(ign.Evidence) == "" || strings.TrimSpace(ign.ControlledValue) == "" {
				errs = append(errs, fmt.Sprintf("ignored param %s.%s: reasons, evidence, and controlled value are required", name, wire))
			}
		}
	}

	registryNames := sortedImplNames(registry)
	for _, name := range registryNames {
		impl := registry[name]
		ep, ok := catalog.Endpoints[name]
		if !ok {
			if !impl.Conditional {
				errs = append(errs, fmt.Sprintf("binding %q: absent from pinned catalog without Conditional evidence", name))
			}
			continue
		}
		if !spec.Checked(ep, allow) && !impl.Conditional {
			errs = append(errs, fmt.Sprintf("binding %q: excluded or private in catalog", name))
		}
	}
	return errs
}

// validateControlledIgnore asserts the controlled wire value is actually
// emitted by the carrier, so a removed ToMap entry fails the audit.
func validateControlledIgnore(impl EndpointImpl, param string, ign IgnoredParam) error {
	if impl.ParamsStruct == "" {
		return fmt.Errorf("controlled ignore requires a params carrier")
	}
	carrier := resolveType(impl.ParamsStruct)
	if carrier == nil {
		return fmt.Errorf("unknown ParamsStruct %q", impl.ParamsStruct)
	}
	toMap, ok := reflect.PointerTo(carrier).MethodByName("ToMap")
	if !ok {
		return fmt.Errorf("%s has no ToMap method", impl.ParamsStruct)
	}
	out := toMap.Func.Call([]reflect.Value{reflect.New(carrier)})[0].Interface().(map[string]string)
	if got, ok := out[param]; !ok || got != ign.ControlledValue {
		return fmt.Errorf("controlled %q=%q not emitted (got %q, present=%v)", param, ign.ControlledValue, got, ok)
	}
	return nil
}

func paramsLabel(impl EndpointImpl) string {
	if impl.ParamsStruct != "" {
		return impl.ParamsStruct
	}
	return impl.Service + "." + impl.Method
}

// queryField maps a wire param to a Go field, honoring explicit overrides.
func queryField(impl EndpointImpl, wireName string) string {
	if override := impl.QueryFields[wireName]; override != "" {
		return override
	}
	return snakeToCamel(wireName)
}

func checkedNames(catalog spec.Catalog, allowlist map[string]bool) []string {
	names := make([]string, 0, len(catalog.Endpoints))
	for name := range catalog.Endpoints {
		if spec.Checked(catalog.Endpoints[name], allowlist) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
