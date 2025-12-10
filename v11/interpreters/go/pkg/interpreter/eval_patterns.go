package interpreter

import (
	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func analyzePatternDeclarationNames(env *runtime.Environment, pattern ast.Pattern) (map[string]struct{}, bool) {
	names := make(map[string]struct{})
	collectPatternIdentifiers(pattern, names)
	newNames := make(map[string]struct{})
	for name := range names {
		if !env.HasInCurrentScope(name) {
			newNames[name] = struct{}{}
		}
	}
	return newNames, len(names) > 0
}

func collectPatternIdentifiers(pattern ast.Pattern, into map[string]struct{}) {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p.Name != "" {
			into[p.Name] = struct{}{}
		}
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil && field.Binding.Name != "" {
				into[field.Binding.Name] = struct{}{}
			}
			if inner, ok := field.Pattern.(ast.Pattern); ok {
				collectPatternIdentifiers(inner, into)
			}
		}
	case *ast.ArrayPattern:
		for _, elem := range p.Elements {
			if elem == nil {
				continue
			}
			if inner, ok := elem.(ast.Pattern); ok {
				collectPatternIdentifiers(inner, into)
			}
		}
		if rest := p.RestPattern; rest != nil {
			if inner, ok := rest.(ast.Pattern); ok {
				collectPatternIdentifiers(inner, into)
			} else if ident, ok := rest.(*ast.Identifier); ok && ident.Name != "" {
				into[ident.Name] = struct{}{}
			}
		}
	case *ast.TypedPattern:
		if inner, ok := p.Pattern.(ast.Pattern); ok {
			collectPatternIdentifiers(inner, into)
		}
	}
}
