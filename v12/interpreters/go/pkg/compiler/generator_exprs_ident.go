package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) compileIdentifier(ctx *compileContext, ident *ast.Identifier, expected string) (string, string, bool) {
	if ident == nil || ident.Name == "" {
		ctx.setReason("missing identifier")
		return "", "", false
	}
	param, ok := ctx.lookup(ident.Name)
	if !ok {
		ctx.setReason("unknown identifier")
		return "", "", false
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
