package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileStaticReceiverMethodCall(
	ctx *compileContext,
	receiver ast.Expression,
	receiverExpr string,
	receiverType string,
	methodName string,
	args []ast.Expression,
	expected string,
	callNode string,
) ([]string, string, string, bool) {
	if g == nil || ctx == nil || receiverExpr == "" || receiverType == "" || methodName == "" {
		return nil, "", "", false
	}
	if methodName == "iterator" {
		if info := g.nativeInterfaceInfoForGoType(receiverType); info != nil && info.TypeExpr != nil {
			if baseName, ok := typeExprBaseName(info.TypeExpr); ok && baseName == "Iterator" {
				return nil, receiverExpr, receiverType, true
			}
		}
	}
	synthetic := ast.NewFunctionCall(ast.NewIdentifier(methodName), args, nil, false)
	if _, ok := g.nativeInterfaceMethodForGoType(receiverType, methodName); ok {
		return g.compileNativeInterfaceMethodCall(ctx, synthetic, expected, receiverExpr, receiverType, methodName, callNode)
	}
	if method := g.methodForReceiver(receiverType, methodName); method != nil {
		if receiver != nil {
			method = g.concreteMethodCallInfo(ctx, synthetic, method, receiver, receiverType, expected)
		}
		return g.compileResolvedMethodCall(ctx, synthetic, expected, method, receiverExpr, receiverType, callNode)
	}
	if method := g.compileableInterfaceMethodForConcreteReceiver(receiverType, methodName); method != nil {
		if receiver != nil {
			method = g.concreteMethodCallInfo(ctx, synthetic, method, receiver, receiverType, expected)
		}
		return g.compileResolvedMethodCall(ctx, synthetic, expected, method, receiverExpr, receiverType, callNode)
	}
	return nil, "", "", false
}

func (g *generator) staticReceiverBestEffortCloseDefer(receiverExpr string, receiverType string) (string, bool) {
	if g == nil || receiverExpr == "" || receiverType == "" {
		return "", false
	}
	if method, ok := g.nativeInterfaceMethodForGoType(receiverType, "close"); ok && method != nil {
		return fmt.Sprintf("defer func() { _, _ = %s.%s() }()", receiverExpr, method.GoName), true
	}
	if method := g.methodForReceiver(receiverType, "close"); method != nil && method.Info != nil && method.Info.Compileable {
		return fmt.Sprintf("defer func() { _, _ = __able_compiled_%s(%s) }()", method.Info.GoName, receiverExpr), true
	}
	if method := g.compileableInterfaceMethodForConcreteReceiver(receiverType, "close"); method != nil && method.Info != nil && method.Info.Compileable {
		return fmt.Sprintf("defer func() { _, _ = __able_compiled_%s(%s) }()", method.Info.GoName, receiverExpr), true
	}
	return "", false
}

func (g *generator) compileStaticIterableForLoopInternal(
	ctx *compileContext,
	loop *ast.ForLoop,
	withResult bool,
	iterSource ast.Expression,
	iterLines []string,
	iterExpr string,
	iterType string,
) ([]string, string, bool) {
	if g == nil || ctx == nil || loop == nil {
		return nil, "", false
	}
	iteratorCall := ast.NewFunctionCall(ast.NewIdentifier("iterator"), nil, nil, false)
	iteratorCallNode := g.diagNodeName(iteratorCall, "*ast.FunctionCall", "call")
	iteratorLines, iteratorExpr, iteratorType, ok := g.compileStaticReceiverMethodCall(ctx, iterSource, iterExpr, iterType, "iterator", nil, "", iteratorCallNode)
	if !ok || iteratorType == "" || iteratorType == "runtime.Value" || iteratorType == "any" {
		return nil, "", false
	}

	loopLabelName := ctx.newTemp()
	resultTemp := ""
	if withResult {
		resultTemp = ctx.newTemp()
	}

	bodyCtx := ctx.child()
	bodyCtx.loopDepth++
	bodyCtx.loopLabel = loopLabelName
	if withResult {
		bodyCtx.loopBreakValueTemp = resultTemp
		bodyCtx.loopBreakValueType = "runtime.Value"
	}
	bodyCtx.loopBreakProbe = nil

	iteratorTemp := ctx.newTemp()
	nextCall := ast.NewFunctionCall(ast.NewIdentifier("next"), nil, nil, false)
	nextCallNode := g.diagNodeName(nextCall, "*ast.FunctionCall", "call")
	nextLines, nextExpr, nextType, ok := g.compileStaticReceiverMethodCall(bodyCtx, nil, iteratorTemp, iteratorType, "next", nil, "", nextCallNode)
	if !ok || nextType == "" || nextType == "runtime.Value" || nextType == "any" {
		return nil, "", false
	}

	endPattern := ast.StructP(nil, false, "IteratorEnd")
	endCondLines, endCond, ok := g.compileMatchPatternCondition(bodyCtx, endPattern, nextExpr, nextType)
	if !ok {
		return nil, "", false
	}

	newNames := map[string]struct{}{}
	collectPatternBindingNames(loop.Pattern, newNames)
	mode := patternBindingMode{declare: true, newNames: newNames}
	condLines, cond, ok := g.compileMatchPatternCondition(bodyCtx, loop.Pattern, nextExpr, nextType)
	if !ok {
		return nil, "", false
	}
	bindLines, ok := g.compileAssignmentPatternBindings(bodyCtx, loop.Pattern, nextExpr, nextType, mode)
	if !ok {
		return nil, "", false
	}
	if cond != "true" || len(condLines) > 0 {
		mismatchLine := fmt.Sprintf("break %s", loopLabelName)
		if withResult {
			mismatchLine = fmt.Sprintf("%s = runtime.ErrorValue{Message: \"pattern assignment mismatch\"}; %s", resultTemp, mismatchLine)
		}
		var condPrefix []string
		condPrefix = append(condPrefix, condLines...)
		condPrefix = append(condPrefix, fmt.Sprintf("if !(%s) { %s }", cond, mismatchLine))
		bindLines = append(condPrefix, bindLines...)
	}
	bodyLines, ok := g.compileBlockStatement(bodyCtx, loop.Body)
	if !ok {
		return nil, "", false
	}

	innerLines := append([]string{}, nextLines...)
	innerLines = append(innerLines, endCondLines...)
	if endCond != "false" {
		innerLines = append(innerLines, fmt.Sprintf("if %s { break }", endCond))
	}
	innerLines = append(innerLines, bindLines...)
	innerLines = append(innerLines, bodyLines...)

	lines := append([]string{}, iterLines...)
	lines = append(lines, iteratorLines...)
	lines = append(lines, fmt.Sprintf("var %s %s = %s", iteratorTemp, iteratorType, iteratorExpr))
	if closeLine, ok := g.staticReceiverBestEffortCloseDefer(iteratorTemp, iteratorType); ok {
		lines = append(lines, closeLine)
	}
	if withResult {
		lines = append(lines, fmt.Sprintf("var %s runtime.Value = runtime.VoidValue{}", resultTemp))
	}
	forPrefix := "for {"
	if linesReferenceLabel(innerLines, loopLabelName) {
		forPrefix = fmt.Sprintf("%s: for {", loopLabelName)
	}
	lines = append(lines, forPrefix)
	lines = append(lines, indentLines(innerLines, 1)...)
	lines = append(lines, "}")
	return lines, resultTemp, true
}
