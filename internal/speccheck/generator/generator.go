// Package generator contains the reusable logic that turns reviewed
// endpoint bindings plus a static scan of the api package into deterministic
// generated resolver tables and registry literals.
//
// Specs can propose gaps but cannot create claims of support: generation only
// consumes bindings.json, which is reviewed by hand. The api package supplies
// type/interface declarations used to derive field lists and to fail
// compilation when a referenced carrier is deleted.
package generator

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	json "github.com/goccy/go-json"
)

// ValueSource identifies where a method reads a query value or path selector.
// Argument is a zero-based method input index including ctx, or -1 when the
// value is a field on the request argument. Field is an exported field path.
type ValueSource struct {
	Argument int    `json:"argument"`
	Field    string `json:"field"`
}

// EndpointBinding is one reviewed endpoint-to-code binding.
type EndpointBinding struct {
	Service      string                 `json:"service"`
	Method       string                 `json:"method"`
	ParamsStruct string                 `json:"paramsStruct"`
	ParamsSource *ValueSource           `json:"paramsSource"`
	QueryArgs    map[string]ValueSource `json:"queryArgs"`
	PathSources  map[string]ValueSource `json:"pathSources"`
	QueryFields  map[string]string      `json:"queryFields"`
	Evidence     string                 `json:"evidence"`
	Conditional  bool                   `json:"conditional"`
}

type bindingRecord struct {
	Name string `json:"name"`
	EndpointBinding
}

// Field is a named struct field.
type Field struct {
	Name     string
	TypeExpr string
	Exported bool
}

// StructInfo describes a scanned struct type.
type StructInfo struct {
	Fields []Field
	Embeds []string // local anonymous embedding type names
}

// MethodInfo describes one method of a scanned interface.
type MethodInfo struct {
	Inputs []string // rendered input type expressions, ctx first
}

// APIModel is the static view of the api package used by generation.
type APIModel struct {
	ParamsTypes []string
	Services    []string
	Structs     map[string]StructInfo
	Aliases     map[string]string
	Methods     map[string]map[string]MethodInfo
}

// ReadBindings loads reviewed bindings from a bindings.json array.
func ReadBindings(path string) (map[string]EndpointBinding, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read bindings %q: %w", path, err)
	}
	var records []bindingRecord
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, fmt.Errorf("parse bindings %q: %w", path, err)
	}
	out := make(map[string]EndpointBinding, len(records))
	for i, rec := range records {
		name := strings.TrimSpace(rec.Name)
		if name == "" {
			return nil, fmt.Errorf("bindings %q: record %d has empty name", path, i)
		}
		if _, ok := out[name]; ok {
			return nil, fmt.Errorf("bindings %q: duplicate binding %q", path, name)
		}
		if rec.Service == "" || rec.Method == "" {
			return nil, fmt.Errorf("bindings %q: %q requires service and method", path, name)
		}
		out[name] = rec.EndpointBinding
	}
	return out, nil
}

// ScanAPI parses every non-test Go file in dir and returns the static model.
func ScanAPI(dir string) (APIModel, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return APIModel{}, fmt.Errorf("parse api dir %q: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	fset := token.NewFileSet()
	model := APIModel{
		Structs: map[string]StructInfo{},
		Aliases: map[string]string{},
		Methods: map[string]map[string]MethodInfo{},
	}
	for _, name := range names {
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.SkipObjectResolution)
		if err != nil {
			return APIModel{}, fmt.Errorf("parse api file %q: %w", name, err)
		}
		scanFile(fset, file, &model)
	}
	sort.Strings(model.ParamsTypes)
	sort.Strings(model.Services)
	return model, nil
}

func scanFile(fset *token.FileSet, file *ast.File, model *APIModel) {
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			switch t := ts.Type.(type) {
			case *ast.InterfaceType:
				if ts.Name.IsExported() {
					model.Services = append(model.Services, ts.Name.Name)
				}
				methods := map[string]MethodInfo{}
				for _, f := range t.Methods.List {
					if len(f.Names) == 0 {
						continue
					}
					ft, ok := f.Type.(*ast.FuncType)
					if !ok {
						continue
					}
					methods[f.Names[0].Name] = MethodInfo{Inputs: funcInputs(fset, ft)}
				}
				model.Methods[ts.Name.Name] = methods
			case *ast.StructType:
				model.Structs[ts.Name.Name] = structInfo(fset, t)
				if ts.Name.IsExported() && strings.HasSuffix(ts.Name.Name, "Params") {
					model.ParamsTypes = append(model.ParamsTypes, ts.Name.Name)
				}
			case *ast.Ident:
				if ts.Assign != 0 {
					model.Aliases[ts.Name.Name] = t.Name
				}
				if ts.Name.IsExported() && strings.HasSuffix(ts.Name.Name, "Params") {
					model.ParamsTypes = append(model.ParamsTypes, ts.Name.Name)
				}
			}
		}
	}
}

func structInfo(fset *token.FileSet, st *ast.StructType) StructInfo {
	info := StructInfo{}
	for _, f := range st.Fields.List {
		typeExpr := exprString(fset, f.Type)
		if len(f.Names) > 0 {
			for _, n := range f.Names {
				info.Fields = append(info.Fields, Field{Name: n.Name, TypeExpr: typeExpr, Exported: n.IsExported()})
			}
			continue
		}
		base, local := embedName(f.Type)
		if base == "" {
			continue
		}
		info.Fields = append(info.Fields, Field{Name: base, TypeExpr: typeExpr, Exported: ast.IsExported(base)})
		if local {
			info.Embeds = append(info.Embeds, base)
		}
	}
	return info
}

func embedName(e ast.Expr) (name string, local bool) {
	switch t := e.(type) {
	case *ast.StarExpr:
		return embedName(t.X)
	case *ast.Ident:
		return t.Name, true
	case *ast.SelectorExpr:
		return t.Sel.Name, false
	default:
		return "", false
	}
}

func funcInputs(fset *token.FileSet, ft *ast.FuncType) []string {
	var out []string
	if ft.Params == nil {
		return out
	}
	for _, f := range ft.Params.List {
		count := len(f.Names)
		if count == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			out = append(out, exprString(fset, f.Type))
		}
	}
	return out
}

func exprString(fset *token.FileSet, e ast.Expr) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, e); err != nil {
		return ""
	}
	return buf.String()
}

// resolveAlias follows local type aliases with cycle detection.
func (m APIModel) resolveAlias(name string) string {
	seen := map[string]bool{}
	for name != "" && !seen[name] {
		seen[name] = true
		next, ok := m.Aliases[name]
		if !ok {
			return name
		}
		name = next
	}
	return name
}

// AccessibleFields returns the sorted exported field names visible on the
// named struct, honoring Go embedding promotion, shadowing, and ambiguity.
func (m APIModel) AccessibleFields(name string) []string {
	type occ struct {
		depth    int
		exported bool
	}
	byName := map[string][]occ{}
	type item struct {
		name  string
		depth int
	}
	queue := []item{{name: name}}
	seen := map[string]int{}
	for len(queue) > 0 {
		it := queue[0]
		queue = queue[1:]
		resolved := m.resolveAlias(it.name)
		if d, ok := seen[resolved]; ok && d <= it.depth {
			continue
		}
		seen[resolved] = it.depth
		si, ok := m.Structs[resolved]
		if !ok {
			continue
		}
		for _, f := range si.Fields {
			byName[f.Name] = append(byName[f.Name], occ{depth: it.depth, exported: f.Exported})
		}
		for _, e := range si.Embeds {
			queue = append(queue, item{name: e, depth: it.depth + 1})
		}
	}
	var out []string
	for field, occs := range byName {
		min := occs[0].depth
		for _, o := range occs[1:] {
			if o.depth < min {
				min = o.depth
			}
		}
		count := 0
		exported := false
		for _, o := range occs {
			if o.depth == min {
				count++
				exported = o.exported
			}
		}
		if count == 1 && exported {
			out = append(out, field)
		}
	}
	sort.Strings(out)
	return out
}

const generatedBanner = "// Code generated by internal/speccheck/generator; DO NOT EDIT.\n"

// GenerateRegistry renders registry_generated.go for the reviewed bindings.
func GenerateRegistry(bindings map[string]EndpointBinding, model APIModel) ([]byte, error) {
	names := make([]string, 0, len(bindings))
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString(generatedBanner)
	b.WriteString("\npackage main\n\n")
	b.WriteString("// Registry maps spec endpoint names to their reviewed implementation.\n")
	b.WriteString("var Registry = map[string]EndpointImpl{\n")
	for _, name := range names {
		binding := bindings[name]
		if _, ok := model.Structs[model.resolveAlias(typeName(binding.ParamsStruct))]; binding.ParamsStruct != "" && !ok {
			return nil, fmt.Errorf("binding %q: unknown params struct %q", name, binding.ParamsStruct)
		}
		fmt.Fprintf(&b, "\t%q: {\n", name)
		fmt.Fprintf(&b, "\t\tService: %q,\n", binding.Service)
		fmt.Fprintf(&b, "\t\tMethod: %q,\n", binding.Method)
		if binding.ParamsStruct != "" {
			fmt.Fprintf(&b, "\t\tParamsStruct: %q,\n", binding.ParamsStruct)
		}
		fields := model.AccessibleFields(typeName(binding.ParamsStruct))
		if binding.ParamsStruct == "" {
			fields = nil
		}
		if len(fields) > 0 {
			fmt.Fprintf(&b, "\t\tParamsFields: []string{%s},\n", quoteList(fields))
		}
		if binding.ParamsSource != nil {
			fmt.Fprintf(&b, "\t\tParamsSource: %s,\n", valueSourceLiteral(*binding.ParamsSource))
		}
		if len(binding.QueryFields) > 0 {
			fmt.Fprintf(&b, "\t\tQueryFields: %s,\n", stringMapLiteral(binding.QueryFields))
		}
		if len(binding.QueryArgs) > 0 {
			fmt.Fprintf(&b, "\t\tQueryArgs: %s,\n", valueSourceMapLiteral(binding.QueryArgs))
		}
		parts := sortedKeys(binding.PathSources)
		if len(parts) > 0 {
			fmt.Fprintf(&b, "\t\tPathParts: []string{%s},\n", quoteList(parts))
			fmt.Fprintf(&b, "\t\tPathSources: %s,\n", valueSourceMapLiteral(binding.PathSources))
		}
		if binding.Evidence != "" {
			fmt.Fprintf(&b, "\t\tEvidence: %q,\n", binding.Evidence)
		}
		if binding.Conditional {
			b.WriteString("\t\tConditional: true,\n")
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n")
	return format.Source([]byte(b.String()))
}

// GenerateTypes renders types_generated.go with the resolver tables.
func GenerateTypes(model APIModel) ([]byte, error) {
	var b strings.Builder
	b.WriteString(generatedBanner)
	b.WriteString("\npackage main\n\n")
	b.WriteString("import (\n\t\"reflect\"\n\n\t\"github.com/disaster37/elasticsearch/v9/api\"\n)\n\n")
	b.WriteString("// paramsTypes resolves qualified api parameter struct names.\n")
	b.WriteString("var paramsTypes = map[string]reflect.Type{\n")
	for _, name := range model.ParamsTypes {
		fmt.Fprintf(&b, "\t\"api.%s\": reflect.TypeOf(api.%s{}),\n", name, name)
	}
	b.WriteString("}\n\n")
	b.WriteString("// serviceTypes resolves api service interface names.\n")
	b.WriteString("var serviceTypes = map[string]reflect.Type{\n")
	for _, name := range model.Services {
		fmt.Fprintf(&b, "\t\"api.%s\": reflect.TypeOf((*api.%s)(nil)).Elem(),\n", name, name)
	}
	b.WriteString("}\n")
	return format.Source([]byte(b.String()))
}

// WriteGenerated replaces only the generated files in dir.
func WriteGenerated(dir string, registry, types []byte) error {
	if len(registry) == 0 || len(types) == 0 {
		return errors.New("refusing to write empty generated output")
	}
	if err := os.WriteFile(filepath.Join(dir, "registry_generated.go"), registry, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "types_generated.go"), types, 0o644)
}

func typeName(qualified string) string {
	if i := strings.LastIndex(qualified, "."); i >= 0 {
		return qualified[i+1:]
	}
	return qualified
}

func quoteList(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return strings.Join(quoted, ", ")
}

func sortedKeys(m map[string]ValueSource) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func valueSourceLiteral(v ValueSource) string {
	if v.Field == "" {
		return fmt.Sprintf("&ValueSource{Argument: %d}", v.Argument)
	}
	return fmt.Sprintf("&ValueSource{Argument: %d, Field: %q}", v.Argument, v.Field)
}

func valueSourceMapLiteral(m map[string]ValueSource) string {
	keys := sortedKeys(m)
	var b strings.Builder
	b.WriteString("map[string]ValueSource{\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "\t\t\t%q: %s,\n", k, strings.TrimPrefix(valueSourceLiteral(m[k]), "&"))
	}
	b.WriteString("\t\t}")
	return b.String()
}

func stringMapLiteral(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("map[string]string{\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "\t\t\t%q: %q,\n", k, m[k])
	}
	b.WriteString("\t\t}")
	return b.String()
}
