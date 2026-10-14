// Command speccheck verifies that the Elasticsearch Go client covers every
// endpoint in a pinned Elasticsearch 9.x REST API spec directory.
//
// Usage:
//
//	go run ./internal/speccheck -spec /path/to/rest-api-spec/api
//	go run ./internal/speccheck -spec /path/to/rest-api-spec/api -inventory
//
// Exit codes: 0 when a valid catalog has no unexpected drift, 1 on coverage,
// registry-integrity, or policy drift, and 2 on missing/bad arguments or
// unreadable/malformed input.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/disaster37/elasticsearch/v9/internal/speccheck/spec"
)

func main() {
	os.Exit(execute(os.Args[1:], os.Stdout, os.Stderr))
}

// execute parses arguments and dispatches to the audit or inventory command.
func execute(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("speccheck", flag.ContinueOnError)
	fs.SetOutput(stderr)
	specDir := fs.String("spec", "", "path to the rest-api-spec/api directory")
	inventoryFlag := fs.Bool("inventory", false, "print the machine-readable endpoint inventory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*specDir) == "" {
		_, _ = fmt.Fprintln(stderr, "speccheck: -spec is required")
		return 2
	}
	if *inventoryFlag {
		return runInventory(*specDir, stdout, stderr)
	}
	return run(*specDir, stdout, stderr)
}

// run audits specDir against the production registry and policy.
func run(specDir string, stdout, stderr io.Writer) int {
	return runWithPolicy(specDir, Registry, defaultPolicy(), stdout, stderr)
}

// runWithPolicy audits specDir against an explicit registry and policy so tests
// can exercise fixtures without mutating globals.
func runWithPolicy(specDir string, registry map[string]EndpointImpl, policy Policy, stdout, stderr io.Writer) int {
	catalog, err := spec.LoadDir(specDir)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "speccheck: %v\n", err)
		return 2
	}
	result := audit(catalog, registry, policy)
	printResult(catalog, result, stdout)
	return exitCode(result)
}

// runInventory loads specDir and prints the inventory.
func runInventory(specDir string, stdout, stderr io.Writer) int {
	catalog, err := spec.LoadDir(specDir)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "speccheck: %v\n", err)
		return 2
	}
	rows, err := inventory(catalog, Registry)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "speccheck: inventory: %v\n", err)
		return 2
	}
	printInventory(rows, stdout)
	return 0
}

func exitCode(result AuditResult) int {
	if len(result.MissingEndpoints) > 0 || len(result.ParamGaps) > 0 || len(result.PathGaps) > 0 ||
		len(result.PolicyErrors) > 0 || len(result.IntegrityErrors) > 0 {
		return 1
	}
	return 0
}

func printResult(catalog spec.Catalog, result AuditResult, stdout io.Writer) {
	_, _ = fmt.Fprintf(stdout, "speccheck: files=%d auxiliary=%d parsed=%d checked=%d implemented=%d known-missing=%d ignored-params=%d\n",
		catalog.Files, catalog.AuxiliaryFiles, len(catalog.Endpoints), result.Checked, result.Implemented, result.KnownMissing, result.Ignored)
	if len(result.MissingEndpoints) > 0 {
		_, _ = fmt.Fprintf(stdout, "MISSING ENDPOINTS (%d):\n", len(result.MissingEndpoints))
		for _, m := range result.MissingEndpoints {
			_, _ = fmt.Fprintf(stdout, "  - %s\n", m)
		}
	}
	if len(result.ParamGaps) > 0 {
		_, _ = fmt.Fprintf(stdout, "PARAM GAPS (%d):\n", len(result.ParamGaps))
		for _, g := range result.ParamGaps {
			_, _ = fmt.Fprintf(stdout, "  - %s\n", g)
		}
	}
	if len(result.PathGaps) > 0 {
		_, _ = fmt.Fprintf(stdout, "PATH PART GAPS (%d):\n", len(result.PathGaps))
		for _, g := range result.PathGaps {
			_, _ = fmt.Fprintf(stdout, "  - %s\n", g)
		}
	}
	if len(result.PolicyErrors) > 0 {
		_, _ = fmt.Fprintf(stdout, "POLICY ERRORS (%d):\n", len(result.PolicyErrors))
		for _, g := range result.PolicyErrors {
			_, _ = fmt.Fprintf(stdout, "  - %s\n", g)
		}
	}
	if len(result.IntegrityErrors) > 0 {
		_, _ = fmt.Fprintf(stdout, "REGISTRY ERRORS (%d):\n", len(result.IntegrityErrors))
		for _, g := range result.IntegrityErrors {
			_, _ = fmt.Fprintf(stdout, "  - %s\n", g)
		}
	}
	switch exitCode(result) {
	case 0:
		if result.KnownMissing > 0 {
			_, _ = fmt.Fprintf(stdout, "speccheck: OK - no unexpected drift (%d known missing endpoints, %d controlled-ignored params)\n", result.KnownMissing, result.Ignored)
		} else {
			_, _ = fmt.Fprintln(stdout, "speccheck: OK - no drift and no known debt")
		}
	default:
		_, _ = fmt.Fprintf(stdout, "speccheck: FAILED - unexpected drift detected\n")
	}
}

func printInventory(rows []InventoryRow, stdout io.Writer) {
	for _, row := range rows {
		_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\t%s\t%s\tparams=%s\tparts=%s\tmissingFields=%s\tunsupportedParts=%s\tevidence=%s\n",
			row.Name, row.Classification, row.Source, row.Visibility, row.Stability,
			row.CandidateService, row.CandidateMethod, strings.Join(row.Params, ","),
			strings.Join(row.PathParts, ","), strings.Join(row.MissingFields, ","),
			strings.Join(row.UnsupportedParts, ","), row.Evidence)
	}
}
