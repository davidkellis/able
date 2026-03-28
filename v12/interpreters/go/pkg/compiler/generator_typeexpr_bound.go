package compiler

import "able/interpreter-go/pkg/ast"

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
	if g == nil || expr == nil {
		return false
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
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
	return !g.typeExprHasGeneric(expr, genericNames) && !g.typeExprHasUnresolvedSimpleName(expr, genericNames)
}

func (g *generator) typeExprHasUnresolvedSimpleName(expr ast.TypeExpression, genericNames map[string]struct{}) bool {
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
		return !g.isConcreteTypeName(t.Name.Name)
	case *ast.GenericTypeExpression:
		if t == nil {
			return false
		}
		if g.typeExprHasUnresolvedSimpleName(t.Base, genericNames) {
			return true
		}
		for _, arg := range t.Arguments {
			if g.typeExprHasUnresolvedSimpleName(arg, genericNames) {
				return true
			}
		}
		return false
	case *ast.NullableTypeExpression:
		return g.typeExprHasUnresolvedSimpleName(t.InnerType, genericNames)
	case *ast.ResultTypeExpression:
		return g.typeExprHasUnresolvedSimpleName(t.InnerType, genericNames)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if g.typeExprHasUnresolvedSimpleName(member, genericNames) {
				return true
			}
		}
		return false
	case *ast.FunctionTypeExpression:
		if g.typeExprHasUnresolvedSimpleName(t.ReturnType, genericNames) {
			return true
		}
		for _, param := range t.ParamTypes {
			if g.typeExprHasUnresolvedSimpleName(param, genericNames) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
