package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileStaticPackageSelectorCall(ctx *compileContext, call *ast.FunctionCall, expected string, callee *ast.MemberAccessExpression, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || callee == nil || callee.Safe {
		return nil, "", "", false
	}
	objectIdent, ok := callee.Object.(*ast.Identifier)
	if !ok || objectIdent == nil || objectIdent.Name == "" {
		return nil, "", "", false
	}
	memberIdent, ok := callee.Member.(*ast.Identifier)
	if !ok || memberIdent == nil || memberIdent.Name == "" {
		return nil, "", "", false
	}
	if _, found := ctx.lookup(objectIdent.Name); found {
		return nil, "", "", false
	}
	qualified := fmt.Sprintf("%s.%s", objectIdent.Name, memberIdent.Name)
	info, overload, ok := g.resolveStaticCallable(ctx, qualified)
	if !ok {
		return nil, "", "", false
	}
	return g.compileStaticNamedFunctionCall(ctx, call, expected, qualified, info, overload, callNode)
}
