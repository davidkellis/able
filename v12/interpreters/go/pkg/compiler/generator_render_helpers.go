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

func (g *generator) callableGenericNames(info *functionInfo) map[string]struct{} {
	if info == nil {
		return nil
	}
	if g != nil && g.implMethodByInfo != nil {
		if impl, ok := g.implMethodByInfo[info]; ok && impl != nil {
			names := genericParamNameSet(info.Definition.GenericParams)
			names = addGenericParams(names, impl.ImplGenerics)
			names = addGenericParams(names, impl.InterfaceGenerics)
			return names
		}
	}
	return g.functionGenericNames(info)
}

func (g *generator) compileContextGenericNames(info *functionInfo) map[string]struct{} {
	names := g.callableGenericNames(info)
	if g == nil || info == nil {
		return names
	}
	names = g.pruneBoundGenericNames(names, g.compileContextTypeBindings(info))
	for _, method := range g.methodList {
		if method == nil || method.Info != info {
			continue
		}
		return mergeGenericNameSets(names, g.methodGenericNames(method))
	}
	return names
}

func (g *generator) pruneBoundGenericNames(names map[string]struct{}, bindings map[string]ast.TypeExpression) map[string]struct{} {
	if g == nil || len(names) == 0 || len(bindings) == 0 {
		return names
	}
	var out map[string]struct{}
	for name := range names {
		expr, ok := bindings[name]
		if !ok || expr == nil {
			continue
		}
		if simple, ok := expr.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == name {
			continue
		}
		if g.typeExprHasGeneric(expr, names) {
			continue
		}
		if out == nil {
			out = make(map[string]struct{}, len(names))
			for existing := range names {
				out[existing] = struct{}{}
			}
		}
		delete(out, name)
	}
	if out == nil {
		return names
	}
	if len(out) == 0 {
		return nil
	}
	return out
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
	if def, ok := g.structInfoForTypeName(method.Info.Package, baseName); ok && def != nil && def.Node != nil {
		names = addGenericParams(names, def.Node.GenericParams)
	}
	if iface, ok := g.interfaces[baseName]; ok && iface != nil {
		names = addGenericParams(names, iface.GenericParams)
	}
	names = mergeGenericNameSets(names, g.typeExprVariableNames(method.TargetType))
	return names
}

func (g *generator) typeExprVariableNames(expr ast.TypeExpression) map[string]struct{} {
	if g == nil || expr == nil {
		return nil
	}
	names := make(map[string]struct{})
	g.collectTypeExprVariableNames(expr, names)
	if len(names) == 0 {
		return nil
	}
	return names
}

func (g *generator) collectTypeExprVariableNames(expr ast.TypeExpression, names map[string]struct{}) {
	if g == nil || expr == nil || names == nil {
		return
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || t.Name.Name == "" || t.Name.Name == "_" {
			return
		}
		if g.isConcreteTypeName(t.Name.Name) {
			return
		}
		names[t.Name.Name] = struct{}{}
	case *ast.GenericTypeExpression:
		if t == nil {
			return
		}
		g.collectTypeExprVariableNames(t.Base, names)
		for _, arg := range t.Arguments {
			g.collectTypeExprVariableNames(arg, names)
		}
	case *ast.NullableTypeExpression:
		g.collectTypeExprVariableNames(t.InnerType, names)
	case *ast.ResultTypeExpression:
		g.collectTypeExprVariableNames(t.InnerType, names)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			g.collectTypeExprVariableNames(member, names)
		}
	case *ast.FunctionTypeExpression:
		for _, param := range t.ParamTypes {
			g.collectTypeExprVariableNames(param, names)
		}
		g.collectTypeExprVariableNames(t.ReturnType, names)
	}
}

func mergeGenericNameSets(left map[string]struct{}, right map[string]struct{}) map[string]struct{} {
	if len(right) == 0 {
		return left
	}
	if len(left) == 0 {
		out := make(map[string]struct{}, len(right))
		for name := range right {
			out[name] = struct{}{}
		}
		return out
	}
	out := make(map[string]struct{}, len(left)+len(right))
	for name := range left {
		out[name] = struct{}{}
	}
	for name := range right {
		out[name] = struct{}{}
	}
	return out
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
		if isPseudoCallableGenericName(gp.Name.Name) {
			continue
		}
		names[gp.Name.Name] = struct{}{}
	}
	return names
}

func isPseudoCallableGenericName(name string) bool {
	switch name {
	case "fn", "()":
		return true
	default:
		return false
	}
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
	case "runtime.ErrorValue":
		return "runtime_error"
	case "any":
		return "any"
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
	if strings.HasPrefix(goType, "[]") {
		return "slice"
	}
	if g != nil {
		if g.isMonoArrayType(goType) {
			return "monoarray"
		}
		if g.nativeCallableInfoForGoType(goType) != nil {
			return "callable"
		}
		if g.nativeInterfaceInfoForGoType(goType) != nil {
			return "interface"
		}
		if g.nativeUnionInfoForGoType(goType) != nil {
			return "union"
		}
	}
	if g != nil && g.structInfoByGoName(goType) != nil {
		return "struct"
	}
	return "unknown"
}

func (g *generator) sortedStructNames() []string {
	return g.sortedStructKeys()
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
