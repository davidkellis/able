package driver

import (
	"reflect"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func collectDynImports(node ast.Node, into map[string]struct{}, visited map[uintptr]struct{}) {
	collectDynImportsFromValue(reflect.ValueOf(node), into, visited)
}

func collectDynImportsFromValue(val reflect.Value, into map[string]struct{}, visited map[uintptr]struct{}) {
	if !val.IsValid() {
		return
	}
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return
		}
		if ptr := val.Pointer(); ptr != 0 {
			if _, ok := visited[ptr]; ok {
				return
			}
			visited[ptr] = struct{}{}
		}
		if node, ok := val.Interface().(ast.Node); ok {
			if dyn, ok := node.(*ast.DynImportStatement); ok && dyn != nil {
				if name := joinIdentifiers(dyn.PackagePath); name != "" {
					into[name] = struct{}{}
				}
			}
		}
		collectDynImportsFromValue(val.Elem(), into, visited)
		return
	}
	if val.Kind() == reflect.Interface {
		if val.IsNil() {
			return
		}
		collectDynImportsFromValue(val.Elem(), into, visited)
		return
	}
	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			collectDynImportsFromValue(val.Field(i), into, visited)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			collectDynImportsFromValue(val.Index(i), into, visited)
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			collectDynImportsFromValue(val.MapIndex(key), into, visited)
		}
	}
}

func joinIdentifiers(ids []*ast.Identifier) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == nil || id.Name == "" {
			continue
		}
		parts = append(parts, id.Name)
	}
	return strings.Join(parts, ".")
}

func buildIdentifiers(parts []string) []*ast.Identifier {
	out := make([]*ast.Identifier, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, ast.NewIdentifier(part))
	}
	return out
}

func copyIdentifiers(ids []*ast.Identifier) []*ast.Identifier {
	if len(ids) == 0 {
		return nil
	}
	out := make([]*ast.Identifier, 0, len(ids))
	for _, id := range ids {
		if id == nil {
			continue
		}
		out = append(out, ast.NewIdentifier(id.Name))
	}
	return out
}

func importKey(imp *ast.ImportStatement) string {
	if imp == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(joinIdentifiers(imp.PackagePath))
	sb.WriteString("|")
	if imp.IsWildcard {
		sb.WriteString("*")
	}
	sb.WriteString("|")
	if imp.Alias != nil {
		sb.WriteString(imp.Alias.Name)
	}
	if len(imp.Selectors) > 0 {
		sb.WriteString("|")
		for _, sel := range imp.Selectors {
			if sel == nil || sel.Name == nil {
				continue
			}
			sb.WriteString(sel.Name.Name)
			if sel.Alias != nil {
				sb.WriteString("::")
				sb.WriteString(sel.Alias.Name)
			}
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func sanitizeSegment(seg string) string {
	seg = strings.TrimSpace(seg)
	seg = strings.ReplaceAll(seg, "-", "_")
	return seg
}
