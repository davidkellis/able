package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) typeExprIncludesNilInPackage(pkgName string, expr ast.TypeExpression) bool {
	return g.typeExprIncludesNilInPackageSeen(pkgName, expr, make(map[string]struct{}))
}

func (g *generator) typeExprIncludesNilInPackageSeen(pkgName string, expr ast.TypeExpression, seen map[string]struct{}) bool {
	if g == nil || expr == nil {
		return false
	}
	normalized := normalizeTypeExprForPackage(g, pkgName, expr)
	if normalized == nil {
		return false
	}
	key := normalizeTypeExprString(g, pkgName, normalized)
	if key != "" {
		if _, ok := seen[key]; ok {
			return false
		}
		seen[key] = struct{}{}
		defer delete(seen, key)
	}
	if isNilTypeExpression(normalized) {
		return true
	}
	if unionPkg, members, ok := g.expandedUnionMembersInPackage(pkgName, normalized); ok {
		for _, member := range members {
			if g.typeExprIncludesNilInPackageSeen(unionPkg, member, seen) {
				return true
			}
		}
		return false
	}
	switch t := normalized.(type) {
	case *ast.NullableTypeExpression:
		return true
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if g.typeExprIncludesNilInPackageSeen(pkgName, member, seen) {
				return true
			}
		}
	}
	return false
}

func (g *generator) nativeUnionNilExpr(expected string) (string, bool) {
	if g == nil || expected == "" {
		return "", false
	}
	info := g.nativeUnionInfoForGoType(expected)
	if info == nil {
		return "", false
	}
	for _, member := range info.Members {
		if member == nil || member.TypeExpr == nil {
			continue
		}
		if typeExpressionToString(member.TypeExpr) != "nil" {
			continue
		}
		switch member.GoType {
		case "runtime.Value":
			return fmt.Sprintf("%s(runtime.NilValue{})", member.WrapHelper), true
		case "any":
			return fmt.Sprintf("%s(any(nil))", member.WrapHelper), true
		default:
			if typedNil, ok := g.typedNilExpr(member.GoType); ok {
				return fmt.Sprintf("%s(%s)", member.WrapHelper, typedNil), true
			}
		}
	}
	var nilCapableMember *nativeUnionMember
	for _, member := range info.Members {
		if member == nil || member.TypeExpr == nil {
			continue
		}
		if isNilTypeExpression(member.TypeExpr) || !g.typeExprIncludesNilInPackage(info.PackageName, member.TypeExpr) {
			continue
		}
		if nilCapableMember != nil {
			return "", false
		}
		nilCapableMember = member
	}
	if nilCapableMember == nil {
		return "", false
	}
	switch nilCapableMember.GoType {
	case "runtime.Value":
		return fmt.Sprintf("%s(runtime.NilValue{})", nilCapableMember.WrapHelper), true
	case "any":
		return fmt.Sprintf("%s(any(nil))", nilCapableMember.WrapHelper), true
	default:
		if typedNil, ok := g.typedNilExpr(nilCapableMember.GoType); ok {
			return fmt.Sprintf("%s(%s)", nilCapableMember.WrapHelper, typedNil), true
		}
	}
	return "", false
}

func (g *generator) nativeResultVoidSuccessExpr(ctx *compileContext, expected string) (string, bool) {
	if g == nil || expected == "" {
		return "", false
	}
	if info := g.nativeUnionInfoForGoType(expected); info != nil {
		if member, ok := g.nativeUnionMember(info, "struct{}"); ok && member != nil {
			return fmt.Sprintf("%s(struct{}{})", member.WrapHelper), true
		}
	}
	if expr, ok := g.nativeUnionNilExpr(expected); ok {
		return expr, true
	}
	if (expected == "runtime.Value" || expected == "any") && ctx != nil {
		if g.isResultVoidTypeExpr(ctx.expectedTypeExpr) || g.isResultVoidTypeExpr(ctx.returnTypeExpr) {
			return "runtime.VoidValue{}", true
		}
	}
	return "", false
}
