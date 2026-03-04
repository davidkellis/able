package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileIdentifier(ctx *compileContext, ident *ast.Identifier, expected string) (string, string, bool) {
	if ident == nil || ident.Name == "" {
		ctx.setReason("missing identifier")
		return "", "", false
	}
	param, ok := ctx.lookup(ident.Name)
	if !ok {
		if info, found := g.structInfoForTypeName(ctx.packageName, ident.Name); found && info != nil && info.Supported && len(info.Fields) == 0 {
			structType := "*" + info.GoName
			if expected == "" || expected == structType {
				return "&" + info.GoName + "{}", structType, true
			}
		}
		nodeName := g.diagNodeName(ident, "*ast.Identifier", "ident")
		valueExpr := fmt.Sprintf("__able_env_get(%q, %s)", ident.Name, nodeName)
		if expected == "" || expected == "runtime.Value" {
			return valueExpr, "runtime.Value", true
		}
		converted, ok := g.expectRuntimeValueExpr(valueExpr, expected)
		if !ok {
			ctx.setReason("identifier type mismatch")
			return "", "", false
		}
		return converted, expected, true
	}
	if !g.typeMatches(expected, param.GoType) {
		if expected == "runtime.Value" && param.GoType != "runtime.Value" {
			converted, ok := g.runtimeValueExpr(param.GoName, param.GoType)
			if !ok {
				ctx.setReason("identifier type mismatch")
				return "", "", false
			}
			return converted, "runtime.Value", true
		}
		if param.GoType == "runtime.Value" && expected != "" {
			converted, ok := g.expectRuntimeValueExpr(param.GoName, expected)
			if !ok {
				ctx.setReason("identifier type mismatch")
				return "", "", false
			}
			return converted, expected, true
		}
		if expected != "" && expected != "runtime.Value" && param.GoType != "runtime.Value" {
			// Preserve runtime coercion semantics for typed locals, including
			// integer-width adjustments used by stdlib error structs.
			valueExpr, ok := g.runtimeValueExpr(param.GoName, param.GoType)
			if ok {
				converted, ok := g.expectRuntimeValueExpr(valueExpr, expected)
				if ok {
					return converted, expected, true
				}
			}
		}
		ctx.setReason("identifier type mismatch")
		return "", "", false
	}
	return param.GoName, param.GoType, true
}

func (g *generator) compileImplicitMemberExpression(ctx *compileContext, expr *ast.ImplicitMemberExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Member == nil {
		ctx.setReason("implicit member requires identifier")
		return "", "", false
	}
	if ctx == nil || !ctx.hasImplicitReceiver || ctx.implicitReceiver.Name == "" {
		ctx.setReason("implicit member requires receiver")
		return "", "", false
	}
	receiver := ast.NewIdentifier(ctx.implicitReceiver.Name)
	memberExpr := ast.NewMemberAccessExpression(receiver, expr.Member)
	return g.compileMemberAccess(ctx, memberExpr, expected)
}
