package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) inferIndexResultTypeExpr(ctx *compileContext, expr *ast.IndexExpression) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || expr == nil || expr.Object == nil {
		return nil, false
	}
	receiverTypeExpr, ok := g.inferExpressionTypeExpr(ctx, expr.Object, "")
	if !ok || receiverTypeExpr == nil {
		return nil, false
	}
	receiverTypeExpr = g.lowerNormalizedTypeExpr(ctx, receiverTypeExpr)

	if baseName, ok := typeExprBaseName(receiverTypeExpr); ok && baseName == "Array" {
		if generic, ok := receiverTypeExpr.(*ast.GenericTypeExpression); ok &&
			generic != nil &&
			len(generic.Arguments) == 1 &&
			generic.Arguments[0] != nil {
			return g.lowerNormalizedTypeExpr(ctx, generic.Arguments[0]), true
		}
		return nil, false
	}

	receiverGoType, ok := g.lowerCarrierType(ctx, receiverTypeExpr)
	if !ok || receiverGoType == "" || receiverGoType == "runtime.Value" || receiverGoType == "any" {
		return nil, false
	}
	if method, ok := g.nativeInterfaceMethodForGoType(receiverGoType, "get"); ok && method != nil && method.ReturnTypeExpr != nil {
		return g.lowerNormalizedTypeExpr(ctx, method.ReturnTypeExpr), true
	}
	if info := g.compileableInterfaceMethodForReceiverArity(receiverGoType, "Index", "get", 1); info != nil {
		if returnExpr := g.functionReturnTypeExpr(info); returnExpr != nil {
			return g.lowerNormalizedTypeExpr(ctx, returnExpr), true
		}
	}
	return nil, false
}
