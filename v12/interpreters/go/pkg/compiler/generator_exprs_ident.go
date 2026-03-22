package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileIdentifier(ctx *compileContext, ident *ast.Identifier, expected string) ([]string, string, string, bool) {
	if ident == nil || ident.Name == "" {
		ctx.setReason("missing identifier")
		return nil, "", "", false
	}
	param, ok := ctx.lookup(ident.Name)
	if !ok {
		if info, found := g.structInfoForTypeName(ctx.packageName, ident.Name); found && info != nil && info.Supported && len(info.Fields) == 0 {
			structType := "*" + info.GoName
			if expected == "" || g.typeMatches(expected, structType) {
				return nil, "&" + info.GoName + "{}", structType, true
			}
		}
		nodeName := g.diagNodeName(ident, "*ast.Identifier", "ident")
		valueExpr := fmt.Sprintf("__able_env_get(%q, %s)", ident.Name, nodeName)
		if expected == "" || expected == "runtime.Value" {
			return nil, valueExpr, "runtime.Value", true
		}
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, valueExpr, expected)
		if !ok {
			ctx.setReason("identifier type mismatch")
			return nil, "", "", false
		}
		return convLines, converted, expected, true
	}
	if expected == "" && param.GoType == "runtime.Value" && param.TypeExpr != nil {
		if inferredType, ok := g.joinCarrierTypeFromTypeExpr(ctx, param.TypeExpr); ok {
			convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, param.GoName, inferredType)
			if ok {
				return convLines, converted, inferredType, true
			}
		}
	}
	if !g.typeMatches(expected, param.GoType) {
		if g.nativeNullableWraps(expected, param.GoType) {
			return nil, fmt.Sprintf("__able_ptr(%s)", param.GoName), expected, true
		}
		if expected != "" && expected != "runtime.Value" && expected != "any" && param.GoType != "runtime.Value" && g.canCoerceStaticExpr(expected, param.GoType) {
			return g.coerceExpectedStaticExpr(ctx, nil, param.GoName, param.GoType, expected)
		}
		if expected == "runtime.Value" && param.GoType != "runtime.Value" {
			convLines, converted, ok := g.runtimeValueLines(ctx, param.GoName, param.GoType)
			if !ok {
				ctx.setReason("identifier type mismatch")
				return nil, "", "", false
			}
			return convLines, converted, "runtime.Value", true
		}
		if param.GoType == "runtime.Value" && expected != "" {
			convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, param.GoName, expected)
			if !ok {
				ctx.setReason("identifier type mismatch")
				return nil, "", "", false
			}
			return convLines, converted, expected, true
		}
		if expected != "" && expected != "runtime.Value" && param.GoType != "runtime.Value" {
			// Preserve runtime coercion semantics for typed locals, including
			// integer-width adjustments used by stdlib error structs.
			valConvLines, valueExpr, ok := g.runtimeValueLines(ctx, param.GoName, param.GoType)
			if ok {
				convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, valueExpr, expected)
				if ok {
					allLines := append([]string{}, valConvLines...)
					allLines = append(allLines, convLines...)
					return allLines, converted, expected, true
				}
			}
		}
		ctx.setReason("identifier type mismatch")
		return nil, "", "", false
	}
	return nil, param.GoName, param.GoType, true
}

func (g *generator) compileImplicitMemberExpression(ctx *compileContext, expr *ast.ImplicitMemberExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Member == nil {
		ctx.setReason("implicit member requires identifier")
		return nil, "", "", false
	}
	if ctx == nil || !ctx.hasImplicitReceiver || ctx.implicitReceiver.Name == "" {
		ctx.setReason("implicit member requires receiver")
		return nil, "", "", false
	}
	receiver := ast.NewIdentifier(ctx.implicitReceiver.Name)
	memberExpr := ast.NewMemberAccessExpression(receiver, expr.Member)
	return g.compileMemberAccess(ctx, memberExpr, expected)
}
