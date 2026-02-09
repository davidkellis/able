package compiler

import (
	"go/format"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) functionGenericNames(info *functionInfo) map[string]struct{} {
	if info == nil || info.Definition == nil {
		return nil
	}
	return genericParamNameSet(info.Definition.GenericParams)
}

func (g *generator) methodGenericNames(method *methodInfo) map[string]struct{} {
	if method == nil || method.Info == nil {
		return nil
	}
	names := genericParamNameSet(method.Info.Definition.GenericParams)
	baseName, ok := typeExprBaseName(method.TargetType)
	if !ok {
		return names
	}
	if def, ok := g.structs[baseName]; ok && def != nil && def.Node != nil {
		names = addGenericParams(names, def.Node.GenericParams)
	}
	if iface, ok := g.interfaces[baseName]; ok && iface != nil {
		names = addGenericParams(names, iface.GenericParams)
	}
	return names
}

func genericParamNameSet(params []*ast.GenericParameter) map[string]struct{} {
	if len(params) == 0 {
		return nil
	}
	names := make(map[string]struct{}, len(params))
	for _, gp := range params {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			continue
		}
		names[gp.Name.Name] = struct{}{}
	}
	return names
}

func addGenericParams(names map[string]struct{}, params []*ast.GenericParameter) map[string]struct{} {
	if len(params) == 0 {
		return names
	}
	if names == nil {
		names = make(map[string]struct{}, len(params))
	}
	for _, gp := range params {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			continue
		}
		names[gp.Name.Name] = struct{}{}
	}
	return names
}

func typeExprBaseName(expr ast.TypeExpression) (string, bool) {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return "", false
		}
		return t.Name.Name, true
	case *ast.GenericTypeExpression:
		if t == nil {
			return "", false
		}
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil {
			return base.Name.Name, true
		}
	}
	return "", false
}

func (g *generator) typeExprHasGeneric(expr ast.TypeExpression, genericNames map[string]struct{}) bool {
	if expr == nil || len(genericNames) == 0 {
		return false
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return false
		}
		_, ok := genericNames[t.Name.Name]
		return ok
	case *ast.GenericTypeExpression:
		if t == nil {
			return false
		}
		if g.typeExprHasGeneric(t.Base, genericNames) {
			return true
		}
		for _, arg := range t.Arguments {
			if g.typeExprHasGeneric(arg, genericNames) {
				return true
			}
		}
		return false
	case *ast.NullableTypeExpression:
		return g.typeExprHasGeneric(t.InnerType, genericNames)
	case *ast.ResultTypeExpression:
		return g.typeExprHasGeneric(t.InnerType, genericNames)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if g.typeExprHasGeneric(member, genericNames) {
				return true
			}
		}
		return false
	case *ast.FunctionTypeExpression:
		if g.typeExprHasGeneric(t.ReturnType, genericNames) {
			return true
		}
		for _, param := range t.ParamTypes {
			if g.typeExprHasGeneric(param, genericNames) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (g *generator) typeCategory(goType string) string {
	switch goType {
	case "runtime.Value":
		return "runtime"
	case "struct{}":
		return "void"
	case "bool":
		return "bool"
	case "string":
		return "string"
	case "rune":
		return "rune"
	case "float32":
		return "float32"
	case "float64":
		return "float64"
	case "int":
		return "int"
	case "uint":
		return "uint"
	case "int8", "int16", "int32", "int64":
		return "int" + goType[3:]
	case "uint8", "uint16", "uint32", "uint64":
		return "uint" + goType[4:]
	}
	for _, info := range g.structs {
		if info.GoName == goType {
			return "struct"
		}
		if strings.HasPrefix(goType, "*") && info.GoName == strings.TrimPrefix(goType, "*") {
			return "struct"
		}
	}
	return "unknown"
}

func (g *generator) sortedStructNames() []string {
	names := make([]string, 0, len(g.structs))
	for name := range g.structs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (g *generator) sortedFunctionNames() []string {
	names := make([]string, 0, len(g.functions))
	for pkgName, pkgFuncs := range g.functions {
		for name := range pkgFuncs {
			names = append(names, qualifiedName(pkgName, name))
		}
	}
	sort.Strings(names)
	return names
}

func formatSource(src []byte) ([]byte, error) {
	formatted, err := format.Source(src)
	if err != nil {
		return src, err
	}
	return formatted, nil
}
