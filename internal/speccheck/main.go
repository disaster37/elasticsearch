// Command speccheck verifies that the Elasticsearch Go client covers every
// endpoint in a pinned Elasticsearch 9.x REST API spec directory.
//
// Usage:
//
//	go run ./internal/speccheck -spec /path/to/rest-api-spec/api
//
// It parses each spec JSON file (skipping _internal.* and visibility != public
// unless feature-flag-allowlisted), and reports:
//   - any endpoint missing from the registry (no service method),
//   - any spec query parameter missing a field on the registered Params struct,
//   - any path part missing from the registered PathParts.
//
// Exit code is non-zero when any gap is found.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// specEndpoint is the subset of the REST spec JSON we care about.
type specEndpoint struct {
	Paths []struct {
		Path  string   `json:"path"`
		Parts []string `json:"parts"`
	} `json:"paths"`
	Params map[string]struct {
		Type string `json:"type"`
	} `json:"params"`
	Visibility string `json:"visibility"`
}

// featureFlagAllowlist lists feature-flag endpoints that are still checked
// even when visibility != public.
var featureFlagAllowlist = map[string]bool{
	"streams.logs_enable":  true,
	"streams.logs_disable": true,
	"streams.status":       true,
}

func main() {
	specDir := flag.String("spec", "", "path to the rest-api-spec/api directory")
	flag.Parse()
	os.Exit(run(*specDir, os.Stdout, os.Stderr))
}

// run executes the speccheck logic against specDir and returns the process
// exit code: 0 when no gaps are found, 1 when gaps are found, 2 when the
// arguments or spec directory are invalid. It is extracted from main so the
// core logic is unit-testable without os.Exit.
func run(specDir string, stdout, stderr io.Writer) int {
	if specDir == "" {
		_, _ = fmt.Fprintln(stderr, "speccheck: -spec is required")
		return 2
	}

	files, err := filepath.Glob(filepath.Join(specDir, "*.json"))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "glob spec dir: %v\n", err)
		return 2
	}
	sort.Strings(files)

	var missing []string
	var paramGaps []string
	var pathGaps []string
	count := 0

	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".json")
		if strings.HasPrefix(name, "_internal") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "read %s: %v\n", f, err)
			continue
		}
		var ep specEndpoint
		if err := json.Unmarshal(raw, &ep); err != nil {
			_, _ = fmt.Fprintf(stderr, "parse %s: %v\n", f, err)
			continue
		}
		if ep.Visibility != "" && ep.Visibility != "public" && !featureFlagAllowlist[name] {
			continue
		}
		count++

		impl, ok := Registry[name]
		if !ok {
			missing = append(missing, name)
			continue
		}
		// Check params.
		for p := range ep.Params {
			if !contains(impl.ParamsFields, snakeToCamel(p)) && impl.ParamsStruct != "" {
				paramGaps = append(paramGaps, fmt.Sprintf("%s: param %q missing from %s", name, p, impl.ParamsStruct))
			}
		}
		// Check path parts.
		parts := collectParts(ep.Paths)
		for _, part := range parts {
			if !contains(impl.PathParts, part) {
				pathGaps = append(pathGaps, fmt.Sprintf("%s: path part %q missing from %s.%s", name, part, impl.Service, impl.Method))
			}
		}
	}

	_, _ = fmt.Fprintf(stdout, "speccheck: scanned %d endpoints\n", count)
	if len(missing) > 0 {
		_, _ = fmt.Fprintf(stdout, "MISSING ENDPOINTS (%d):\n", len(missing))
		for _, m := range missing {
			_, _ = fmt.Fprintf(stdout, "  - %s\n", m)
		}
	}
	if len(paramGaps) > 0 {
		_, _ = fmt.Fprintf(stdout, "PARAM GAPS (%d):\n", len(paramGaps))
		for _, g := range paramGaps {
			_, _ = fmt.Fprintf(stdout, "  - %s\n", g)
		}
	}
	if len(pathGaps) > 0 {
		_, _ = fmt.Fprintf(stdout, "PATH PART GAPS (%d):\n", len(pathGaps))
		for _, g := range pathGaps {
			_, _ = fmt.Fprintf(stdout, "  - %s\n", g)
		}
	}
	if len(missing)+len(paramGaps)+len(pathGaps) == 0 {
		_, _ = fmt.Fprintln(stdout, "speccheck: OK - no gaps found")
		return 0
	}
	return 1
}

func collectParts(paths []struct {
	Path  string   `json:"path"`
	Parts []string `json:"parts"`
}) []string {
	seen := map[string]bool{}
	for _, p := range paths {
		for _, part := range p.Parts {
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

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// snakeToCamel converts a snake_case spec param name to CamelCase (the Go field
// naming convention used by the Params structs).
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			b.WriteString(p[1:])
		}
	}
	return b.String()
}
