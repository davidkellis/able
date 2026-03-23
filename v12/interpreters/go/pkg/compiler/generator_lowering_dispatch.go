package compiler

import "able/interpreter-go/pkg/ast"

// lowerDispatchCall is the canonical call-dispatch synthesis entrypoint.
func (g *generator) lowerDispatchCall(ctx *compileContext, call *ast.FunctionCall, expected string) ([]string, string, string, bool) {
	return g.compileFunctionCall(ctx, call, expected)
}

// lowerDispatchMember is the canonical member-dispatch synthesis entrypoint.
func (g *generator) lowerDispatchMember(ctx *compileContext, expr *ast.MemberAccessExpression, expected string) ([]string, string, string, bool) {
	return g.compileMemberAccess(ctx, expr, expected)
}

// lowerDispatchIndex is the canonical index-dispatch synthesis entrypoint.
func (g *generator) lowerDispatchIndex(ctx *compileContext, expr *ast.IndexExpression, expected string) ([]string, string, string, bool) {
	return g.compileIndexExpression(ctx, expr, expected)
}

// lowerResolveStaticMethod is the canonical static-method resolution entrypoint.
func (g *generator) lowerResolveStaticMethod(ctx *compileContext, receiver ast.Expression, name string) (*methodInfo, bool) {
	return g.resolveStaticMethodCall(ctx, receiver, name)
}

// lowerResolvedMethodDispatch is the canonical resolved-method dispatch
// synthesis entrypoint.
func (g *generator) lowerResolvedMethodDispatch(ctx *compileContext, call *ast.FunctionCall, expected string, method *methodInfo, objExpr string, objType string, callNode string) ([]string, string, string, bool) {
	return g.compileResolvedMethodCall(ctx, call, expected, method, objExpr, objType, callNode)
}

// lowerNativeInterfaceMethodDispatch is the canonical native interface method
// dispatch entrypoint for statically resolved interface carriers.
func (g *generator) lowerNativeInterfaceMethodDispatch(ctx *compileContext, call *ast.FunctionCall, expected string, objExpr string, objType string, methodName string, callNode string) ([]string, string, string, bool) {
	return g.compileNativeInterfaceMethodCall(ctx, call, expected, objExpr, objType, methodName, callNode)
}

// lowerNativeInterfaceGenericMethodDispatch is the canonical generic native
// interface/default-method dispatch entrypoint.
func (g *generator) lowerNativeInterfaceGenericMethodDispatch(ctx *compileContext, call *ast.FunctionCall, expected string, objExpr string, objType string, methodName string, callNode string) ([]string, string, string, bool) {
	return g.compileNativeInterfaceGenericMethodCall(ctx, call, expected, objExpr, objType, methodName, callNode)
}
