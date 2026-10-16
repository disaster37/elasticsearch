//go:build ignore

// Command generate renders the speccheck resolver tables and registry literal
// from reviewed bindings plus a static scan of the api package.
//
//go:generate go run generate.go -api ../../api -bindings bindings.json -out .
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/disaster37/elasticsearch/v9/internal/speccheck/generator"
)

func main() {
	apiDir := flag.String("api", "../../api", "api package directory")
	bindingsPath := flag.String("bindings", "bindings.json", "reviewed bindings file")
	outDir := flag.String("out", ".", "output directory")
	flag.Parse()

	if err := run(*apiDir, *bindingsPath, *outDir); err != nil {
		fmt.Fprintln(os.Stderr, "speccheck generate:", err)
		os.Exit(1)
	}
}

func run(apiDir, bindingsPath, outDir string) error {
	bindings, err := generator.ReadBindings(bindingsPath)
	if err != nil {
		return err
	}
	model, err := generator.ScanAPI(apiDir)
	if err != nil {
		return err
	}
	registry, err := generator.GenerateRegistry(bindings, model)
	if err != nil {
		return err
	}
	types, err := generator.GenerateTypes(model)
	if err != nil {
		return err
	}
	return generator.WriteGenerated(outDir, registry, types)
}
