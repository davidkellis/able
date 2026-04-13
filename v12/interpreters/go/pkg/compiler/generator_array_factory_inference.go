package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) refinedFreshArrayBindingTypeExpr(ctx *compileContext, name string, right ast.Expression, goType string, current ast.TypeExpression) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || name == "" || !g.isArrayStructType(goType) || !g.isForwardInferrableFreshArrayExpr(right) {
		return nil, false
	}
	if current != nil {
		current = g.lowerNormalizedTypeExpr(ctx, current)
		if elem, ok := g.concreteArrayTypeElementExpr(ctx, current); ok && elem != nil {
			return ast.Gen(ast.Ty("Array"), elem), true
		}
		baseName, ok := typeExprBaseName(current)
		if ok && baseName == "Array" && !g.typeExprHasWildcard(current) {
			return nil, false
		}
	}
	refined, ok := g.forwardFreshArrayTypeExpr(ctx, name)
	if !ok || refined == nil {
		return nil, false
	}
	return refined, true
}

func (g *generator) isForwardInferrableFreshArrayExpr(expr ast.Expression) bool {
	if g == nil || expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.ArrayLiteral:
		return e != nil && len(e.Elements) == 0
	case *ast.BlockExpression:
		return g.isForwardInferrableFreshArrayBlock(e)
	case *ast.IfExpression:
		return g.isForwardInferrableFreshArrayIf(e)
	case *ast.MatchExpression:
		return g.isForwardInferrableFreshArrayMatch(e)
	}
	return g.isUntypedArrayFactoryCall(expr)
}

func (g *generator) isForwardInferrableFreshArrayBlock(block *ast.BlockExpression) bool {
	if g == nil || block == nil || len(block.Body) == 0 {
		return false
	}
	resultExpr, ok := g.blockResultExpression(block)
	if !ok || resultExpr == nil {
		return false
	}
	return g.isForwardInferrableFreshArrayExpr(resultExpr)
}

func (g *generator) isForwardInferrableFreshArrayIf(expr *ast.IfExpression) bool {
	if g == nil || expr == nil || expr.IfBody == nil || expr.ElseBody == nil {
		return false
	}
	if !g.isForwardInferrableFreshArrayBlock(expr.IfBody) || !g.isForwardInferrableFreshArrayBlock(expr.ElseBody) {
		return false
	}
	for _, clause := range expr.ElseIfClauses {
		if clause == nil || clause.Body == nil || !g.isForwardInferrableFreshArrayBlock(clause.Body) {
			return false
		}
	}
	return true
}

func (g *generator) isForwardInferrableFreshArrayMatch(expr *ast.MatchExpression) bool {
	if g == nil || expr == nil || len(expr.Clauses) == 0 {
		return false
	}
	for _, clause := range expr.Clauses {
		if clause == nil || clause.Body == nil || !g.isForwardInferrableFreshArrayExpr(clause.Body) {
			return false
		}
	}
	return true
}

func (g *generator) blockResultExpression(block *ast.BlockExpression) (ast.Expression, bool) {
	if block == nil || len(block.Body) == 0 {
		return nil, false
	}
	resultExpr, ok := block.Body[len(block.Body)-1].(ast.Expression)
	if !ok || resultExpr == nil {
		return nil, false
	}
	return resultExpr, true
}

func (g *generator) isUntypedArrayFactoryCall(expr ast.Expression) bool {
	call, ok := expr.(*ast.FunctionCall)
	if !ok || call == nil {
		return false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe {
		return false
	}
	typeIdent, ok := member.Object.(*ast.Identifier)
	if !ok || typeIdent == nil || typeIdent.Name != "Array" {
		return false
	}
	methodIdent, ok := member.Member.(*ast.Identifier)
	if !ok || methodIdent == nil {
		return false
	}
	switch methodIdent.Name {
	case "new":
		return len(call.Arguments) == 0
	case "with_capacity":
		return len(call.Arguments) == 1
	default:
		return false
	}
}

func (g *generator) forwardFreshArrayTypeExpr(ctx *compileContext, name string) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || name == "" || len(ctx.blockStatements) == 0 {
		return nil, false
	}
	var elemExpr ast.TypeExpression
	for idx := ctx.statementIndex + 1; idx < len(ctx.blockStatements); idx++ {
		nextElem, found, stop := g.forwardArrayFactoryElementTypeFromStatement(ctx, name, ctx.blockStatements[idx])
		if found {
			if elemExpr == nil {
				elemExpr = nextElem
			} else if typeExpressionToString(elemExpr) != typeExpressionToString(nextElem) {
				return nil, false
			}
		}
		if stop {
			break
		}
	}
	if elemExpr == nil {
		return nil, false
	}
	return ast.Gen(ast.Ty("Array"), elemExpr), true
}

func (g *generator) forwardArrayFactoryElementTypeFromStatement(ctx *compileContext, name string, stmt ast.Statement) (ast.TypeExpression, bool, bool) {
	if g == nil || ctx == nil || stmt == nil || name == "" {
		return nil, false, false
	}
	switch s := stmt.(type) {
	case *ast.AssignmentExpression:
		if targetName, _, ok := g.assignmentTargetName(s.Left); ok && targetName == name {
			return nil, false, true
		}
		if indexExpr, ok := s.Left.(*ast.IndexExpression); ok && indexExpr != nil && identExprNamed(indexExpr.Object, name) {
			if elemExpr, ok := g.forwardArrayValueTypeExpr(ctx, s.Right); ok {
				return elemExpr, true, false
			}
		}
		if typeExpr := assignmentTypeAnnotation(s.Left); typeExpr != nil && identExprNamed(s.Right, name) {
			if elemExpr, ok := g.concreteArrayTypeElementExpr(ctx, typeExpr); ok {
				return elemExpr, true, false
			}
		}
		if call, ok := s.Right.(*ast.FunctionCall); ok {
			if elemExpr, ok := g.forwardArrayCallArgumentElementTypeExpr(ctx, name, call); ok {
				return elemExpr, true, false
			}
		}
	case *ast.FunctionCall:
		if elemExpr, ok := g.forwardArrayFactoryElementTypeFromCall(ctx, name, s); ok {
			return elemExpr, true, false
		}
	case *ast.ReturnStatement:
		if identExprNamed(s.Argument, name) {
			if elemExpr, ok := g.concreteArrayTypeElementExpr(ctx, ctx.returnTypeExpr); ok {
				return elemExpr, true, false
			}
			return nil, false, true
		}
		if call, ok := s.Argument.(*ast.FunctionCall); ok {
			if elemExpr, ok := g.forwardArrayCallArgumentElementTypeExpr(ctx, name, call); ok {
				return elemExpr, true, false
			}
		}
	}
	return nil, false, false
}

func (g *generator) forwardArrayFactoryElementTypeFromCall(ctx *compileContext, name string, call *ast.FunctionCall) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil {
		return nil, false
	}
	if elemExpr, ok := g.forwardArrayCallArgumentElementTypeExpr(ctx, name, call); ok {
		return elemExpr, true
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe || !identExprNamed(member.Object, name) {
		return nil, false
	}
	methodIdent, ok := member.Member.(*ast.Identifier)
	if !ok || methodIdent == nil {
		return nil, false
	}
	switch methodIdent.Name {
	case "push":
		if len(call.Arguments) != 1 {
			return nil, false
		}
		return g.forwardArrayValueTypeExpr(ctx, call.Arguments[0])
	case "write_slot":
		if len(call.Arguments) != 2 {
			return nil, false
		}
		return g.forwardArrayValueTypeExpr(ctx, call.Arguments[1])
	case "push_all":
		if len(call.Arguments) != 1 {
			return nil, false
		}
		valueExpr, ok := g.forwardArrayArgumentTypeExpr(ctx, call.Arguments[0])
		if !ok {
			return nil, false
		}
		return g.concreteArrayTypeElementExpr(ctx, valueExpr)
	default:
		return nil, false
	}
}

func (g *generator) forwardArrayCallArgumentElementTypeExpr(ctx *compileContext, name string, call *ast.FunctionCall) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || name == "" {
		return nil, false
	}
	var elemExpr ast.TypeExpression
	for idx, arg := range call.Arguments {
		if !identExprNamed(arg, name) {
			continue
		}
		paramTypeExpr, ok := g.forwardStaticCallArgumentTypeExpr(ctx, call, idx)
		if !ok {
			continue
		}
		nextElem, ok := g.concreteArrayTypeElementExpr(ctx, paramTypeExpr)
		if !ok {
			continue
		}
		if elemExpr == nil {
			elemExpr = nextElem
			continue
		}
		if normalizeTypeExprString(g, ctx.packageName, elemExpr) != normalizeTypeExprString(g, ctx.packageName, nextElem) {
			return nil, false
		}
	}
	if elemExpr == nil {
		return nil, false
	}
	return elemExpr, true
}

func (g *generator) forwardStaticCallArgumentTypeExpr(ctx *compileContext, call *ast.FunctionCall, argIdx int) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || argIdx < 0 {
		return nil, false
	}
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		return g.forwardStaticNamedCallArgumentTypeExpr(ctx, call, callee, argIdx)
	case *ast.MemberAccessExpression:
		return g.forwardStaticMethodCallArgumentTypeExpr(ctx, call, callee, argIdx)
	default:
		return nil, false
	}
}

func (g *generator) forwardStaticNamedCallArgumentTypeExpr(ctx *compileContext, call *ast.FunctionCall, callee *ast.Identifier, argIdx int) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || callee == nil || callee.Name == "" {
		return nil, false
	}
	info, _, ok := g.resolveStaticCallable(ctx, callee.Name)
	if !ok || info == nil {
		return nil, false
	}
	info = g.concreteFunctionCallInfo(ctx, call, info, "")
	g.refreshRepresentableFunctionInfo(info)
	return g.forwardConcreteCallParamTypeExpr(info, argIdx)
}

func (g *generator) forwardStaticMethodCallArgumentTypeExpr(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression, argIdx int) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || callee == nil || callee.Safe {
		return nil, false
	}
	memberIdent, ok := callee.Member.(*ast.Identifier)
	if !ok || memberIdent == nil || memberIdent.Name == "" {
		return nil, false
	}
	if method, ok := g.resolveStaticMethodCallForCall(ctx, call, callee.Object, memberIdent.Name); ok && method != nil && method.Info != nil {
		method = g.concreteStaticMethodCallInfo(ctx, call, method, callee.Object, "")
		if method == nil || method.Info == nil {
			return nil, false
		}
		g.refreshRepresentableFunctionInfo(method.Info)
		return g.forwardConcreteCallParamTypeExpr(method.Info, argIdx)
	}
	receiverType := g.forwardCallReceiverGoType(ctx, callee.Object)
	if receiverType == "" {
		return nil, false
	}
	method := g.methodForReceiver(receiverType, memberIdent.Name)
	if method == nil {
		method = g.compileableInterfaceMethodForConcreteReceiverExpr(ctx, callee.Object, receiverType, memberIdent.Name)
	}
	if method == nil || method.Info == nil {
		return nil, false
	}
	paramIdx := argIdx + 1
	if method.ExpectsSelf {
		receiverType := g.forwardCallReceiverGoType(ctx, callee.Object)
		if receiverType == "" {
			return nil, false
		}
		method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, receiverType, "")
	} else {
		method = g.concreteStaticMethodCallInfo(ctx, call, method, callee.Object, "")
		paramIdx = argIdx
	}
	if method == nil || method.Info == nil {
		return nil, false
	}
	g.refreshRepresentableFunctionInfo(method.Info)
	return g.forwardConcreteCallParamTypeExpr(method.Info, paramIdx)
}

func (g *generator) forwardConcreteCallParamTypeExpr(info *functionInfo, idx int) (ast.TypeExpression, bool) {
	if g == nil || info == nil || idx < 0 || idx >= len(info.Params) {
		return nil, false
	}
	if paramTypeExpr := g.functionParamTypeExpr(info, idx); paramTypeExpr != nil {
		return normalizeTypeExprForPackage(g, info.Package, paramTypeExpr), true
	}
	if info.Params[idx].TypeExpr != nil {
		return normalizeTypeExprForPackage(g, info.Package, info.Params[idx].TypeExpr), true
	}
	return nil, false
}

func (g *generator) forwardCallReceiverGoType(ctx *compileContext, expr ast.Expression) string {
	if g == nil || ctx == nil || expr == nil {
		return ""
	}
	if goType := g.ufcsReceiverGoType(ctx, expr); goType != "" {
		return goType
	}
	if inferred, ok := g.inferExpressionTypeExpr(ctx, expr, ""); ok && inferred != nil {
		inferred = g.lowerNormalizedTypeExpr(ctx, inferred)
		if goType, ok := g.lowerCarrierType(ctx, inferred); ok && goType != "" && goType != "runtime.Value" && goType != "any" {
			return goType
		}
	}
	return ""
}

func (g *generator) forwardArrayValueTypeExpr(ctx *compileContext, expr ast.Expression) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || expr == nil {
		return nil, false
	}
	if inferred := g.inferredExpressionTypeExpr(ctx, expr); inferred != nil {
		inferred = g.lowerNormalizedTypeExpr(ctx, inferred)
		if inferred != nil && g.typeExprFullyBound(ctx.packageName, inferred) {
			return inferred, true
		}
	}
	if inferred, ok := g.inferExpressionTypeExpr(ctx, expr, ""); ok && inferred != nil {
		inferred = g.lowerNormalizedTypeExpr(ctx, inferred)
		if inferred != nil && g.typeExprFullyBound(ctx.packageName, inferred) {
			return inferred, true
		}
	}
	return nil, false
}

func (g *generator) forwardArrayArgumentTypeExpr(ctx *compileContext, expr ast.Expression) (ast.TypeExpression, bool) {
	valueExpr, ok := g.forwardArrayValueTypeExpr(ctx, expr)
	if !ok || valueExpr == nil {
		return nil, false
	}
	baseName, ok := typeExprBaseName(valueExpr)
	if !ok || baseName != "Array" {
		return nil, false
	}
	return valueExpr, true
}

func (g *generator) concreteArrayTypeElementExpr(ctx *compileContext, expr ast.TypeExpression) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || expr == nil {
		return nil, false
	}
	expr = g.lowerNormalizedTypeExpr(ctx, expr)
	baseName, ok := typeExprBaseName(expr)
	if !ok || baseName != "Array" {
		return nil, false
	}
	generic, ok := expr.(*ast.GenericTypeExpression)
	if !ok || generic == nil || len(generic.Arguments) != 1 || generic.Arguments[0] == nil {
		return nil, false
	}
	elemExpr := g.lowerNormalizedTypeExpr(ctx, generic.Arguments[0])
	if elemExpr == nil || g.typeExprHasWildcard(elemExpr) || !g.typeExprFullyBound(ctx.packageName, elemExpr) {
		return nil, false
	}
	return elemExpr, true
}

func identExprNamed(expr ast.Expression, name string) bool {
	ident, ok := expr.(*ast.Identifier)
	return ok && ident != nil && ident.Name == name
}

func assignmentTypeAnnotation(target ast.AssignmentTarget) ast.TypeExpression {
	if typed, ok := target.(*ast.TypedPattern); ok && typed != nil {
		return typed.TypeAnnotation
	}
	return nil
}
