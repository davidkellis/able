package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) typeExprHasWildcard(expr ast.TypeExpression) bool {
	if g == nil || expr == nil {
		return false
	}
	switch t := expr.(type) {
	case *ast.WildcardTypeExpression:
		return true
	case *ast.GenericTypeExpression:
		if t == nil {
			return false
		}
		if g.typeExprHasWildcard(t.Base) {
			return true
		}
		for _, arg := range t.Arguments {
			if g.typeExprHasWildcard(arg) {
				return true
			}
		}
		return false
	case *ast.NullableTypeExpression:
		return g.typeExprHasWildcard(t.InnerType)
	case *ast.ResultTypeExpression:
		return g.typeExprHasWildcard(t.InnerType)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if g.typeExprHasWildcard(member) {
				return true
			}
		}
		return false
	case *ast.FunctionTypeExpression:
		if g.typeExprHasWildcard(t.ReturnType) {
			return true
		}
		for _, param := range t.ParamTypes {
			if g.typeExprHasWildcard(param) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (g *generator) typeExprFullyBound(pkgName string, expr ast.TypeExpression) bool {
	return g.typeExprFullyBoundSeen(pkgName, expr, make(map[string]struct{}))
}

func (g *generator) typeExprFullyBoundSeen(pkgName string, expr ast.TypeExpression, seen map[string]struct{}) bool {
	if g == nil || expr == nil {
		return false
	}
	pkgName, expr = g.normalizeTypeExprContextForPackage(pkgName, expr)
	if expr == nil || g.typeExprHasWildcard(expr) {
		return false
	}
	genericNames := make(map[string]struct{})
	if ifacePkg, _, _, ifaceDef, ok := interfaceExprInfo(g, pkgName, expr); ok && ifaceDef != nil {
		genericNames = addGenericParams(genericNames, ifaceDef.GenericParams)
		pkgName = ifacePkg
	}
	if baseName, ok := typeExprBaseName(expr); ok && baseName != "" {
		if info, ok := g.structInfoForTypeName(pkgName, baseName); ok && info != nil && info.Node != nil {
			genericNames = addGenericParams(genericNames, info.Node.GenericParams)
		}
	}
	return !g.typeExprHasGeneric(expr, genericNames) && !g.typeExprHasUnresolvedSimpleNameSeen(pkgName, expr, genericNames, seen)
}

func (g *generator) simpleTypeNameConcreteInPackage(pkgName string, name string) bool {
	return g.simpleTypeNameConcreteInPackageSeen(pkgName, name, make(map[string]struct{}))
}

func (g *generator) simpleTypeNameResolvableInPackage(pkgName string, name string) bool {
	return g.simpleTypeNameResolvableInPackageSeen(pkgName, name, make(map[string]struct{}))
}

func (g *generator) simpleTypeNameConcreteInPackageSeen(pkgName string, name string, seen map[string]struct{}) bool {
	if g == nil {
		return false
	}
	pkgName = strings.TrimSpace(pkgName)
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	switch name {
	case "bool", "Bool", "String", "string", "char", "Char",
		"i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64",
		"isize", "usize", "f32", "f64", "void", "Void",
		"Error", "Value", "nil":
		return true
	}
	seenKey := pkgName + "::" + name
	if _, ok := seen[seenKey]; ok {
		return false
	}
	seen[seenKey] = struct{}{}
	defer delete(seen, seenKey)

	if info, ok := g.structInfoForTypeName(pkgName, name); ok && info != nil {
		return info.Node == nil || len(info.Node.GenericParams) == 0
	}
	if iface, _, ok := g.interfaceDefinitionForPackage(pkgName, name); ok && iface != nil {
		return len(iface.GenericParams) == 0
	}
	if aliasPkg, sourceName, target, params, ok := g.typeAliasTargetForPackage(pkgName, name); ok {
		if len(params) != 0 {
			return false
		}
		// Imported selector bindings reuse the type-alias resolver even when the
		// source symbol is a struct/interface rather than an alias. In that case
		// the returned target is just the remote simple name, so recurse through
		// the source package to decide whether it is concretely bound there.
		if simple, ok := target.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && strings.TrimSpace(simple.Name.Name) == strings.TrimSpace(sourceName) {
			return g.simpleTypeNameConcreteInPackageSeen(aliasPkg, sourceName, seen)
		}
		return g.typeExprFullyBoundSeen(aliasPkg, target, seen)
	}
	if unionPkg, members, ok := g.expandedUnionMembersInPackage(pkgName, ast.Ty(name)); ok && unionPkg != "" && len(members) != 0 {
		return true
	}
	return g.isConcreteTypeName(name)
}

func (g *generator) simpleTypeNameResolvableInPackageSeen(pkgName string, name string, seen map[string]struct{}) bool {
	if g == nil {
		return false
	}
	pkgName = strings.TrimSpace(pkgName)
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	switch name {
	case "bool", "Bool", "String", "string", "char", "Char",
		"i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64",
		"isize", "usize", "f32", "f64", "void", "Void",
		"Error", "Value", "nil":
		return true
	}
	seenKey := pkgName + "::" + name
	if _, ok := seen[seenKey]; ok {
		return false
	}
	seen[seenKey] = struct{}{}
	defer delete(seen, seenKey)

	if info, ok := g.structInfoForTypeName(pkgName, name); ok && info != nil {
		return true
	}
	if iface, _, ok := g.interfaceDefinitionForPackage(pkgName, name); ok && iface != nil {
		return true
	}
	if aliasPkg, sourceName, target, params, ok := g.typeAliasTargetForPackage(pkgName, name); ok {
		if len(params) != 0 {
			return true
		}
		if simple, ok := target.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && strings.TrimSpace(simple.Name.Name) == strings.TrimSpace(sourceName) {
			return g.simpleTypeNameResolvableInPackageSeen(aliasPkg, sourceName, seen)
		}
		return g.typeExprFullyBoundSeen(aliasPkg, target, seen)
	}
	if unionPkg, members, ok := g.expandedUnionMembersInPackage(pkgName, ast.Ty(name)); ok && unionPkg != "" && len(members) != 0 {
		return true
	}
	return g.isConcreteTypeName(name)
}

func (g *generator) typeExprHasUnresolvedSimpleName(pkgName string, expr ast.TypeExpression, genericNames map[string]struct{}) bool {
	return g.typeExprHasUnresolvedSimpleNameSeen(pkgName, expr, genericNames, make(map[string]struct{}))
}

func (g *generator) typeExprHasUnresolvedSimpleNameSeen(pkgName string, expr ast.TypeExpression, genericNames map[string]struct{}, seen map[string]struct{}) bool {
	if g == nil || expr == nil {
		return false
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || t.Name.Name == "" {
			return false
		}
		if len(genericNames) > 0 {
			if _, ok := genericNames[t.Name.Name]; ok {
				return true
			}
		}
		return !g.simpleTypeNameConcreteInPackageSeen(pkgName, t.Name.Name, seen)
	case *ast.GenericTypeExpression:
		if t == nil {
			return false
		}
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil && base.Name.Name != "" {
			if len(genericNames) > 0 {
				if _, ok := genericNames[base.Name.Name]; ok {
					return true
				}
			}
			if !g.simpleTypeNameResolvableInPackageSeen(pkgName, base.Name.Name, seen) {
				return true
			}
		} else if g.typeExprHasUnresolvedSimpleNameSeen(pkgName, t.Base, genericNames, seen) {
			return true
		}
		for _, arg := range t.Arguments {
			if g.typeExprHasUnresolvedSimpleNameSeen(pkgName, arg, genericNames, seen) {
				return true
			}
		}
		return false
	case *ast.NullableTypeExpression:
		return g.typeExprHasUnresolvedSimpleNameSeen(pkgName, t.InnerType, genericNames, seen)
	case *ast.ResultTypeExpression:
		return g.typeExprHasUnresolvedSimpleNameSeen(pkgName, t.InnerType, genericNames, seen)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if g.typeExprHasUnresolvedSimpleNameSeen(pkgName, member, genericNames, seen) {
				return true
			}
		}
		return false
	case *ast.FunctionTypeExpression:
		if g.typeExprHasUnresolvedSimpleNameSeen(pkgName, t.ReturnType, genericNames, seen) {
			return true
		}
		for _, param := range t.ParamTypes {
			if g.typeExprHasUnresolvedSimpleNameSeen(pkgName, param, genericNames, seen) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
