package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// validateRegistry checks that every registry entry resolves against the api
// package and that declared sources and fields are internally consistent.
func validateRegistry(registry map[string]EndpointImpl) []string {
	var errs []string
	names := sortedImplNames(registry)
	for _, name := range names {
		impl := registry[name]
		svcType := resolveService(impl.Service)
		if svcType == nil {
			errs = append(errs, fmt.Sprintf("%s: unknown service %q", name, impl.Service))
			continue
		}
		method, ok := svcType.MethodByName(impl.Method)
		if !ok {
			errs = append(errs, fmt.Sprintf("%s: service %s has no method %q", name, impl.Service, impl.Method))
			continue
		}
		if impl.ParamsStruct == "" {
			if impl.ParamsSource != nil {
				errs = append(errs, fmt.Sprintf("%s: ParamsSource set without ParamsStruct", name))
			}
			if len(impl.ParamsFields) > 0 {
				errs = append(errs, fmt.Sprintf("%s: ParamsFields set without ParamsStruct", name))
			}
		} else {
			carrier := resolveType(impl.ParamsStruct)
			if carrier == nil {
				errs = append(errs, fmt.Sprintf("%s: unknown ParamsStruct %q", name, impl.ParamsStruct))
			} else {
				if impl.ParamsSource == nil {
					errs = append(errs, fmt.Sprintf("%s: ParamsStruct %s requires ParamsSource", name, impl.ParamsStruct))
				} else if reached, err := resolveSource(method, *impl.ParamsSource); err != nil {
					errs = append(errs, fmt.Sprintf("%s: ParamsSource: %v", name, err))
				} else if deref(reached) != carrier {
					errs = append(errs, fmt.Sprintf("%s: ParamsSource reaches %s, want %s", name, deref(reached), carrier))
				}
				if _, ok := reflect.PointerTo(carrier).MethodByName("ToMap"); !ok {
					errs = append(errs, fmt.Sprintf("%s: %s has no ToMap method", name, impl.ParamsStruct))
				}
				want := exportedFields(carrier)
				if !equalStrings(want, impl.ParamsFields) {
					errs = append(errs, fmt.Sprintf("%s: ParamsFields mismatch for %s: generated=%v reflect=%v", name, impl.ParamsStruct, impl.ParamsFields, want))
				}
			}
		}
		fieldSet := map[string]bool{}
		for _, f := range impl.ParamsFields {
			fieldSet[f] = true
		}
		for wire, field := range impl.QueryFields {
			if wire == "" || field == "" {
				errs = append(errs, fmt.Sprintf("%s: empty QueryFields entry", name))
				continue
			}
			if !fieldSet[field] {
				errs = append(errs, fmt.Sprintf("%s: QueryFields[%q]=%q is not an accessible field", name, wire, field))
			}
		}
		for wire, src := range impl.QueryArgs {
			if wire == "" {
				errs = append(errs, fmt.Sprintf("%s: empty QueryArgs key", name))
				continue
			}
			if _, err := resolveSource(method, src); err != nil {
				errs = append(errs, fmt.Sprintf("%s: QueryArgs[%q]: %v", name, wire, err))
			}
		}
		parts := make([]string, 0, len(impl.PathSources))
		for part, src := range impl.PathSources {
			if part == "" {
				errs = append(errs, fmt.Sprintf("%s: empty PathSources key", name))
				continue
			}
			if _, err := resolveSource(method, src); err != nil {
				errs = append(errs, fmt.Sprintf("%s: PathSources[%q]: %v", name, part, err))
			}
			parts = append(parts, part)
		}
		sort.Strings(parts)
		if !equalStrings(parts, impl.PathParts) {
			errs = append(errs, fmt.Sprintf("%s: PathParts %v != PathSources %v", name, impl.PathParts, parts))
		}
		if impl.Conditional && impl.Evidence == "" {
			errs = append(errs, fmt.Sprintf("%s: conditional binding requires evidence", name))
		}
	}
	sort.Strings(errs)
	return errs
}

func sortedImplNames(registry map[string]EndpointImpl) []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// exportedFields returns the sorted set of exported fields accessible through
// Go promotion rules using reflection.
func exportedFields(typ reflect.Type) []string {
	if typ == nil {
		return nil
	}
	var out []string
	for _, f := range reflect.VisibleFields(typ) {
		if f.IsExported() {
			out = append(out, f.Name)
		}
	}
	sort.Strings(out)
	return out
}

// resolveSource resolves a ValueSource against a service method signature.
func resolveSource(method reflect.Method, source ValueSource) (reflect.Type, error) {
	mt := method.Type
	if source.Argument == -1 {
		for i := 0; i < mt.NumIn(); i++ {
			base := deref(mt.In(i))
			if base.Kind() != reflect.Struct {
				continue
			}
			if field, ok := fieldByPath(base, source.Field); ok {
				return field, nil
			}
		}
		return nil, fmt.Errorf("no request argument exposes field %q", source.Field)
	}
	if source.Argument < 0 || source.Argument >= mt.NumIn() {
		return nil, fmt.Errorf("argument %d out of range (0..%d)", source.Argument, mt.NumIn()-1)
	}
	base := mt.In(source.Argument)
	if source.Field == "" {
		return base, nil
	}
	field, ok := fieldByPath(deref(base), source.Field)
	if !ok {
		return nil, fmt.Errorf("argument %d has no field path %q", source.Argument, source.Field)
	}
	return field, nil
}

func fieldByPath(typ reflect.Type, path string) (reflect.Type, bool) {
	if path == "" {
		return typ, true
	}
	cur := typ
	for _, seg := range strings.Split(path, ".") {
		cur = deref(cur)
		if cur.Kind() == reflect.Slice || cur.Kind() == reflect.Array {
			cur = deref(cur.Elem())
		}
		if cur.Kind() != reflect.Struct {
			return nil, false
		}
		f, ok := cur.FieldByName(seg)
		if !ok {
			return nil, false
		}
		cur = f.Type
	}
	return cur, true
}

func deref(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// validateParamSerialization instantiates a fresh carrier, sets exactly one
// accessible field, and asserts ToMap emits the exact wire key. Pointer
// booleans are exercised with both true and false, integers with zero and
// nonzero, and lists with two elements.
func validateParamSerialization(impl EndpointImpl, param, field string) error {
	carrier := resolveType(impl.ParamsStruct)
	if carrier == nil {
		return fmt.Errorf("unknown ParamsStruct %q", impl.ParamsStruct)
	}
	toMap, ok := reflect.PointerTo(carrier).MethodByName("ToMap")
	if !ok {
		return fmt.Errorf("%s has no ToMap method", impl.ParamsStruct)
	}
	return checkFieldSerialization(carrier, toMap, param, field)
}

func checkFieldSerialization(carrier reflect.Type, toMap reflect.Method, param, field string) error {
	fieldType, ok := carrier.FieldByName(field)
	if !ok {
		return fmt.Errorf("%s has no field %q", carrier.Name(), field)
	}
	samples := fieldSamples(fieldType.Type)
	if len(samples) == 0 {
		return fmt.Errorf("unsupported field type for %s.%s", carrier.Name(), field)
	}
	for _, sample := range samples {
		ptr := reflect.New(carrier)
		f := ptr.Elem().FieldByName(field)
		if !f.CanSet() {
			return fmt.Errorf("%s.%s is not settable", carrier.Name(), field)
		}
		if err := sample.set(f); err != nil {
			return fmt.Errorf("set %s.%s: %w", carrier.Name(), field, err)
		}
		out := toMap.Func.Call([]reflect.Value{ptr})[0].Interface().(map[string]string)
		got, present := out[param]
		if !present {
			return fmt.Errorf("%s.%s did not serialize wire key %q", carrier.Name(), field, param)
		}
		if got != sample.value {
			return fmt.Errorf("%s.%s serialized %q=%q, want %q", carrier.Name(), field, param, got, sample.value)
		}
	}
	return nil
}

type fieldSample struct {
	value string
	set   func(reflect.Value) error
}

func fieldSamples(t reflect.Type) []fieldSample {
	switch t.Kind() {
	case reflect.Pointer:
		switch t.Elem().Kind() {
		case reflect.Bool:
			return []fieldSample{
				{value: "true", set: func(v reflect.Value) error { b := true; v.Set(reflect.ValueOf(&b)); return nil }},
				{value: "false", set: func(v reflect.Value) error { b := false; v.Set(reflect.ValueOf(&b)); return nil }},
			}
		case reflect.Int:
			return []fieldSample{
				{value: "0", set: func(v reflect.Value) error { i := 0; v.Set(reflect.ValueOf(&i)); return nil }},
				{value: "7", set: func(v reflect.Value) error { i := 7; v.Set(reflect.ValueOf(&i)); return nil }},
			}
		case reflect.Int64:
			return []fieldSample{
				{value: "0", set: func(v reflect.Value) error { i := int64(0); v.Set(reflect.ValueOf(&i)); return nil }},
				{value: "9", set: func(v reflect.Value) error { i := int64(9); v.Set(reflect.ValueOf(&i)); return nil }},
			}
		case reflect.Float64:
			return []fieldSample{
				{value: "1.5", set: func(v reflect.Value) error { f := 1.5; v.Set(reflect.ValueOf(&f)); return nil }},
			}
		}
	case reflect.Bool:
		return []fieldSample{
			{value: "true", set: func(v reflect.Value) error { v.SetBool(true); return nil }},
			{value: "false", set: func(v reflect.Value) error { v.SetBool(false); return nil }},
		}
	case reflect.String:
		return []fieldSample{
			{value: "sample", set: func(v reflect.Value) error { v.SetString("sample"); return nil }},
		}
	case reflect.Slice:
		if t.Elem().Kind() == reflect.String {
			return []fieldSample{
				{value: "a,b", set: func(v reflect.Value) error { v.Set(reflect.ValueOf([]string{"a", "b"})); return nil }},
			}
		}
	}
	return nil
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// snakeToCamel converts a snake_case spec param name to the CamelCase Go field
// naming convention. It preserves the existing _source handling.
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
