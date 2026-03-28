package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) safeNavigationCarrierType(goType string) (string, bool) {
	if g == nil || goType == "" {
		return "", false
	}
	if g.isVoidType(goType) {
		return "struct{}", true
	}
	if g.isNativeNullableValueType(goType) || g.goTypeHasNilZeroValue(goType) {
		return goType, true
	}
	if nullableType, ok := g.nativeNullablePointerType(goType); ok {
		return nullableType, true
	}
	return "", false
}

func (g *generator) safeNavigationNilCheckExpr(expr string, goType string) string {
	if expr == "" {
		return "true"
	}
	switch goType {
	case "runtime.Value", "runtime.ErrorValue":
		return fmt.Sprintf("__able_is_nil(%s)", expr)
	case "any":
		return fmt.Sprintf("(%s) == nil", expr)
	}
	if strings.HasPrefix(goType, "*") || strings.HasPrefix(goType, "[]") {
		return fmt.Sprintf("%s == nil", expr)
	}
	if g != nil && g.goTypeHasNilZeroValue(goType) {
		return fmt.Sprintf("%s == nil", expr)
	}
	return fmt.Sprintf("__able_is_nil(%s)", expr)
}

func (g *generator) safeNavigationCoerceSuccessExpr(ctx *compileContext, expr string, exprType string, resultType string) ([]string, string, bool) {
	if g == nil || ctx == nil || expr == "" || exprType == "" || resultType == "" {
		return nil, "", false
	}
	if resultType == "any" || g.typeMatches(resultType, exprType) {
		return nil, expr, true
	}
	if resultType == "runtime.Value" {
		return g.lowerRuntimeValue(ctx, expr, exprType)
	}
	if g.nativeNullableWraps(resultType, exprType) {
		return nil, fmt.Sprintf("__able_ptr(%s)", expr), true
	}
	if exprType == "runtime.Value" {
		return g.lowerExpectRuntimeValue(ctx, expr, resultType)
	}
	if g.canCoerceStaticExpr(resultType, exprType) {
		lines, coercedExpr, _, ok := g.lowerCoerceExpectedStaticExpr(ctx, nil, expr, exprType, resultType)
		return lines, coercedExpr, ok
	}
	return nil, "", false
}

func (g *generator) compileStaticSafeMemberCall(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression, expected string, objExpr string, objType string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || callee == nil {
		return nil, "", "", false
	}
	memberIdent, ok := callee.Member.(*ast.Identifier)
	if !ok || memberIdent == nil || memberIdent.Name == "" {
		return nil, "", "", false
	}
	method := g.methodForReceiver(objType, memberIdent.Name)
	if method == nil {
		method = g.compileableInterfaceMethodForConcreteReceiverExpr(ctx, callee.Object, objType, memberIdent.Name)
	}
	if method == nil || method.Info == nil {
		return nil, "", "", false
	}
	method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, "")
	if method == nil || method.Info == nil || !method.Info.Compileable {
		return nil, "", "", false
	}

	resultType := expected
	if resultType == "" {
		inferred, ok := g.safeNavigationCarrierType(method.Info.ReturnType)
		if !ok {
			resultType = "runtime.Value"
		} else {
			resultType = inferred
		}
	}

	objTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	nilExpr := safeNilReturnExpr(resultType)
	if wrapped, ok := g.nativeUnionNilExpr(resultType); ok {
		nilExpr = wrapped
	}
	lines := []string{
		fmt.Sprintf("%s := %s", objTemp, objExpr),
		fmt.Sprintf("var %s %s", resultTemp, resultType),
		fmt.Sprintf("if %s {", g.safeNavigationNilCheckExpr(objTemp, objType)),
		fmt.Sprintf("\t%s = %s", resultTemp, nilExpr),
		"} else {",
	}

	callLines, callExpr, callType, ok := g.lowerResolvedMethodDispatch(ctx, call, "", method, objTemp, objType, callNode)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, indentLines(callLines, 1)...)
	coerceLines, coercedExpr, ok := g.safeNavigationCoerceSuccessExpr(ctx, callExpr, callType, resultType)
	if !ok {
		ctx.setReason("safe member call return type mismatch")
		return nil, "", "", false
	}
	lines = append(lines, indentLines(coerceLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, coercedExpr))
	lines = append(lines, "}")
	return lines, resultTemp, resultType, true
}
