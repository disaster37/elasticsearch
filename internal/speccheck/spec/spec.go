// Package spec parses Elasticsearch REST API specification files into a
// normalized catalog that the speccheck audit and inventory commands share.
//
// The real spec files are JSON objects that map one or more endpoint names to
// their endpoint bodies, for example:
//
//	{
//	  "info": {
//	    "url":   {"paths": [{"path": "/", "methods": ["GET"]}]},
//	    "params": {"pretty": {"type": "boolean"}},
//	    "visibility": "public"
//	  }
//	}
//
// Endpoint names come only from those top-level keys, never from filenames.
// Input is treated fail-closed: malformed or contradictory specs are errors,
// never silently skipped.
package spec

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	json "github.com/goccy/go-json"
)

// Param describes a single endpoint-local query parameter.
type Param struct {
	Type       string          `json:"type"`
	Options    []string        `json:"options"`
	Visibility string          `json:"visibility"`
	Stability  string          `json:"stability"`
	Default    json.RawMessage `json:"default"`
}

// Path is a single URL template accepted by an endpoint.
type Path struct {
	Path    string                     `json:"path"`
	Methods []string                   `json:"methods"`
	Parts   map[string]json.RawMessage `json:"parts"`
}

// URL groups the paths of an endpoint.
type URL struct {
	Paths []Path `json:"paths"`
}

// Endpoint is the body of a single endpoint entry.
type Endpoint struct {
	URL        URL              `json:"url"`
	Params     map[string]Param `json:"params"`
	Visibility string           `json:"visibility"`
	Stability  string           `json:"stability"`
}

// NamedEndpoint pairs an endpoint name with its parsed body and the file it
// came from.
type NamedEndpoint struct {
	Name     string
	Source   string
	Endpoint Endpoint
}

// Catalog is the parsed set of endpoint files in a directory.
type Catalog struct {
	// Endpoints includes excluded endpoints (internal/private) for validation;
	// callers filter with Checked.
	Endpoints map[string]NamedEndpoint
	// Files is the number of endpoint JSON files parsed.
	Files int
	// AuxiliaryFiles is the number of auxiliary JSON files parsed (_common).
	AuxiliaryFiles int
}

// AuxiliaryFile is the only auxiliary file recognized in the spec directory.
const AuxiliaryFile = "_common.json"

var placeholderRE = regexp.MustCompile(`\{([^{}]+)\}`)

// ParseFile parses the raw bytes of one spec file into its named endpoints.
func ParseFile(filename string, raw []byte) ([]NamedEndpoint, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("%s: parse: %w", filename, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, fmt.Errorf("%s: root must be a JSON object", filename)
	}

	var out []NamedEndpoint
	seen := map[string]bool{}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("%s: parse key: %w", filename, err)
		}
		name, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("%s: non-string endpoint key", filename)
		}
		if name == "" {
			return nil, fmt.Errorf("%s: empty endpoint key", filename)
		}
		if seen[name] {
			return nil, fmt.Errorf("%s: duplicate endpoint key %q", filename, name)
		}
		seen[name] = true

		var body json.RawMessage
		if err := dec.Decode(&body); err != nil {
			return nil, fmt.Errorf("%s: parse endpoint %q: %w", filename, name, err)
		}
		ep, err := parseEndpoint(name, filename, body)
		if err != nil {
			return nil, err
		}
		out = append(out, ep)
	}
	// Consume the closing delimiter.
	if _, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("%s: parse end of object: %w", filename, err)
	}
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("%s: trailing data after JSON object", filename)
		}
		return nil, fmt.Errorf("%s: parse trailing data: %w", filename, err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no endpoints found", filename)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func parseEndpoint(name, filename string, body json.RawMessage) (NamedEndpoint, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return NamedEndpoint{}, fmt.Errorf("%s: endpoint %q is null", filename, name)
	}
	if trimmed[0] != '{' {
		return NamedEndpoint{}, fmt.Errorf("%s: endpoint %q must be a JSON object", filename, name)
	}
	var ep Endpoint
	if err := json.Unmarshal(body, &ep); err != nil {
		return NamedEndpoint{}, fmt.Errorf("%s: parse endpoint %q: %w", filename, name, err)
	}
	var rawParams struct {
		Params map[string]json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(body, &rawParams); err != nil {
		return NamedEndpoint{}, fmt.Errorf("%s: parse endpoint %q params: %w", filename, name, err)
	}
	for param, raw := range rawParams.Params {
		if param == "" {
			return NamedEndpoint{}, fmt.Errorf("%s: endpoint %q: empty param name", filename, name)
		}
		t := bytes.TrimSpace(raw)
		if len(t) == 0 || t[0] != '{' {
			return NamedEndpoint{}, fmt.Errorf("%s: endpoint %q: param %q must be a JSON object", filename, name, param)
		}
	}
	if err := validateEndpoint(name, filename, &ep); err != nil {
		return NamedEndpoint{}, err
	}
	return NamedEndpoint{Name: name, Source: filename, Endpoint: ep}, nil
}

func validateEndpoint(name, filename string, ep *Endpoint) error {
	if len(ep.URL.Paths) == 0 {
		return fmt.Errorf("%s: endpoint %q: missing url.paths", filename, name)
	}
	repeated := map[string]map[string]json.RawMessage{}
	for i := range ep.URL.Paths {
		p := &ep.URL.Paths[i]
		if strings.TrimSpace(p.Path) == "" {
			return fmt.Errorf("%s: endpoint %q: path with empty template", filename, name)
		}
		if len(p.Methods) == 0 {
			return fmt.Errorf("%s: endpoint %q path %q: missing methods", filename, name, p.Path)
		}
		for _, m := range p.Methods {
			if strings.TrimSpace(m) == "" {
				return fmt.Errorf("%s: endpoint %q path %q: empty method", filename, name, p.Path)
			}
		}
		declared := map[string]bool{}
		for key, raw := range p.Parts {
			if key == "" {
				return fmt.Errorf("%s: endpoint %q path %q: empty part name", filename, name, p.Path)
			}
			t := bytes.TrimSpace(raw)
			if len(t) == 0 || t[0] != '{' {
				return fmt.Errorf("%s: endpoint %q path %q: part %q must be a JSON object", filename, name, p.Path, key)
			}
			declared[key] = true
		}
		placeholders := map[string]bool{}
		for _, m := range placeholderRE.FindAllStringSubmatch(p.Path, -1) {
			placeholders[m[1]] = true
		}
		if !sameStringSet(placeholders, declared) {
			return fmt.Errorf("%s: endpoint %q path %q: placeholders %v do not match declared parts %v",
				filename, name, p.Path, sortedSet(placeholders), sortedSet(declared))
		}
		if prev, ok := repeated[p.Path]; ok {
			if !samePartDescriptors(prev, p.Parts) {
				return fmt.Errorf("%s: endpoint %q path %q: contradictory part declarations", filename, name, p.Path)
			}
			continue
		}
		repeated[p.Path] = p.Parts
	}
	return nil
}

func samePartDescriptors(a, b map[string]json.RawMessage) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok || !sameJSON(av, bv) {
			return false
		}
	}
	return true
}

func sameJSON(a, b json.RawMessage) bool {
	var ab, bb bytes.Buffer
	if json.Compact(&ab, a) != nil || json.Compact(&bb, b) != nil {
		return bytes.Equal(a, b)
	}
	return ab.String() == bb.String()
}

func sameStringSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func sortedSet(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CollectParts returns the unique sorted part names across paths.
func CollectParts(paths []Path) []string {
	seen := map[string]bool{}
	for _, p := range paths {
		for part := range p.Parts {
			seen[part] = true
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Checked reports whether an endpoint is subject to the coverage audit:
// endpoints under _common or _internal are skipped, as is any endpoint whose
// visibility is not legacy-public ("") or "public". An explicitly allowlisted
// feature_flag endpoint is checked, but the allowlist never overrides private
// visibility.
func Checked(ep NamedEndpoint, allowlist map[string]bool) bool {
	name := ep.Name
	if name == "_common" || strings.HasPrefix(name, "_internal") {
		return false
	}
	switch ep.Endpoint.Visibility {
	case "", "public":
		return true
	case "feature_flag":
		return allowlist[name]
	default:
		return false
	}
}

// LoadDir reads and parses every spec file in dir. The directory must exist
// and contain at least one endpoint object; all parse errors are aggregated.
func LoadDir(dir string) (Catalog, error) {
	return loadDir(dir, os.ReadDir)
}

func loadDir(dir string, readDir func(string) ([]fs.DirEntry, error)) (Catalog, error) {
	entries, err := readDir(dir)
	if err != nil {
		return Catalog{}, fmt.Errorf("read spec dir %q: %w", dir, err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	catalog := Catalog{Endpoints: map[string]NamedEndpoint{}}
	var errs []error
	seenNames := map[string]string{}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		if !entry.IsDir() && entry.Type()&fs.ModeSymlink == 0 && !entry.Type().IsRegular() {
			errs = append(errs, fmt.Errorf("%s: not a regular file", filepath.Join(dir, name)))
			continue
		}
		full := filepath.Join(dir, name)
		info, statErr := os.Stat(full)
		if statErr != nil {
			errs = append(errs, fmt.Errorf("%s: %w", full, statErr))
			continue
		}
		if info.IsDir() {
			errs = append(errs, fmt.Errorf("%s: is a directory", full))
			continue
		}
		raw, readErr := os.ReadFile(full)
		if readErr != nil {
			errs = append(errs, fmt.Errorf("%s: %w", full, readErr))
			continue
		}
		if name == AuxiliaryFile {
			var anyJSON any
			if err := json.Unmarshal(raw, &anyJSON); err != nil {
				errs = append(errs, fmt.Errorf("%s: parse auxiliary file: %w", full, err))
				continue
			}
			catalog.AuxiliaryFiles++
			continue
		}
		eps, parseErr := ParseFile(full, raw)
		if parseErr != nil {
			errs = append(errs, parseErr)
			continue
		}
		catalog.Files++
		for _, ep := range eps {
			if prev, ok := seenNames[ep.Name]; ok {
				errs = append(errs, fmt.Errorf("duplicate endpoint %q in %s and %s", ep.Name, prev, ep.Source))
				continue
			}
			seenNames[ep.Name] = ep.Source
			catalog.Endpoints[ep.Name] = ep
		}
	}
	if len(errs) > 0 {
		sort.Slice(errs, func(i, j int) bool { return errs[i].Error() < errs[j].Error() })
		return Catalog{}, errors.Join(errs...)
	}
	if len(catalog.Endpoints) == 0 {
		return Catalog{}, fmt.Errorf("spec dir %q contains no endpoints", dir)
	}
	return catalog, nil
}
