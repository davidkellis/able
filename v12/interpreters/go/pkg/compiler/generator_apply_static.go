package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) compileStaticApplyCall(ctx *compileContext, call *ast.FunctionCall, expected string, receiverExpr string, receiverType string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || receiverExpr == "" || receiverType == "" {
		return nil, "", "", false
	}
	if info := g.nativeCallableInfoForGoType(receiverType); info != nil {
		return g.compileNativeCallableCall(ctx, call, expected, receiverExpr, info.GoType, info.TypeExpr, callNode)
	}
	synthetic := ast.NewFunctionCall(ast.NewIdentifier("apply"), call.Arguments, call.TypeArguments, call.IsTrailingLambda)
	if _, ok := g.nativeInterfaceMethodForGoType(receiverType, "apply"); ok {
		return g.lowerNativeInterfaceMethodDispatch(ctx, synthetic, expected, receiverExpr, receiverType, "apply", callNode)
	}
	if method := g.compileableInterfaceMethodForConcreteReceiver(receiverType, "apply"); method != nil {
		return g.lowerResolvedMethodDispatch(ctx, synthetic, expected, method, receiverExpr, receiverType, callNode)
	}
	return nil, "", "", false
}
