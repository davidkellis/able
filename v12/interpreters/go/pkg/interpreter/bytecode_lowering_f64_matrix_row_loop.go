package interpreter

import "able/interpreter-go/pkg/ast"

type bytecodeF64MatrixRowLoopPlan struct {
	indexSlot          int
	boundSlot          int
	resultReceiverSlot int
	leftReceiverSlot   int
	rightOuterSlot     int
	successTarget      int
	resultPushIP       int
}

func (ctx *bytecodeLoweringContext) setF64MatrixRowLoopPlan(index int, plan bytecodeF64MatrixRowLoopPlan) {
	if ctx == nil || index < 0 {
		return
	}
	if ctx.f64MatrixRowLoops == nil {
		ctx.f64MatrixRowLoops = make(map[int]bytecodeF64MatrixRowLoopPlan, 1)
	}
	ctx.f64MatrixRowLoops[index] = plan
}

func bytecodeF64MatrixRowLoopPlanForLoop(ctx *bytecodeLoweringContext, loop *ast.LoopExpression) (bytecodeF64MatrixRowLoopPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes || loop == nil || loop.Body == nil || len(loop.Body.Body) != 7 {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	indexName, boundName, ok := bytecodeF64DotLoopBreakCondition(loop.Body.Body[0])
	if !ok {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	accumulatorName, ok := bytecodeF64MatrixRowLoopZeroFloatDeclaration(loop.Body.Body[1])
	if !ok {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	rowName, outerName, rowIndexName, ok := bytecodeF64MatrixRowLoopArrayGetDeclaration(loop.Body.Body[2])
	if !ok || rowIndexName != indexName {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	innerIndexName, ok := bytecodeF64MatrixRowLoopZeroIntegerDeclaration(loop.Body.Body[3])
	if !ok {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	innerLoop, ok := loop.Body.Body[4].(*ast.LoopExpression)
	if !ok || innerLoop == nil || innerLoop.Body == nil || len(innerLoop.Body.Body) != 3 {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	innerBreakIndex, innerBoundName, ok := bytecodeF64DotLoopBreakCondition(innerLoop.Body.Body[0])
	if !ok || innerBreakIndex != innerIndexName || innerBoundName != boundName {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	innerAccumulatorName, addMulExpr, ok := bytecodeF64DotLoopAssignment(innerLoop.Body.Body[1])
	if !ok || innerAccumulatorName != accumulatorName {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	leftReceiverName, leftIndexName, rightReceiverName, rightIndexName, ok := bytecodeF64MatrixRowLoopAddMulArrayGetNames(addMulExpr, accumulatorName)
	if !ok || leftIndexName != innerIndexName || rightIndexName != innerIndexName {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	leftName := ""
	switch {
	case leftReceiverName == rowName && rightReceiverName != rowName:
		leftName = rightReceiverName
	case rightReceiverName == rowName && leftReceiverName != rowName:
		leftName = leftReceiverName
	default:
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	resultReceiverName, pushedName, ok := bytecodeF64MatrixRowLoopPush(loop.Body.Body[5])
	if !ok || pushedName != accumulatorName {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	if !bytecodeF64DotLoopIncrement(loop.Body.Body[6], indexName) {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	indexSlot, found := ctx.lookupSlot(indexName)
	if !found {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	boundSlot, found := ctx.lookupSlot(boundName)
	if !found {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	resultReceiverSlot, found := ctx.lookupSlot(resultReceiverName)
	if !found {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	leftReceiverSlot, found := ctx.lookupSlot(leftName)
	if !found {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	rightOuterSlot, found := ctx.lookupSlot(outerName)
	if !found {
		return bytecodeF64MatrixRowLoopPlan{}, false
	}
	return bytecodeF64MatrixRowLoopPlan{
		indexSlot:          indexSlot,
		boundSlot:          boundSlot,
		resultReceiverSlot: resultReceiverSlot,
		leftReceiverSlot:   leftReceiverSlot,
		rightOuterSlot:     rightOuterSlot,
	}, true
}

func bytecodeF64MatrixRowLoopZeroFloatDeclaration(stmt ast.Statement) (string, bool) {
	name, expr, ok := bytecodeF64MatrixRowLoopDeclaration(stmt)
	if !ok {
		return "", false
	}
	lit, ok := expr.(*ast.FloatLiteral)
	if !ok || lit == nil || lit.Value != 0 {
		return "", false
	}
	return name, true
}

func bytecodeF64MatrixRowLoopZeroIntegerDeclaration(stmt ast.Statement) (string, bool) {
	name, expr, ok := bytecodeF64MatrixRowLoopDeclaration(stmt)
	if !ok {
		return "", false
	}
	lit, ok := expr.(*ast.IntegerLiteral)
	if !ok || lit == nil {
		return "", false
	}
	_, raw, ok := bytecodeSlotConstIntegerLiteralImmediate(lit)
	return name, ok && raw == 0
}

func bytecodeF64MatrixRowLoopDeclaration(stmt ast.Statement) (string, ast.Expression, bool) {
	assign, ok := stmt.(*ast.AssignmentExpression)
	if !ok || assign == nil || assign.Operator != ast.AssignmentDeclare || assign.Right == nil {
		return "", nil, false
	}
	target, ok := assign.Left.(*ast.Identifier)
	if !ok || target == nil || target.Name == "" {
		return "", nil, false
	}
	return target.Name, assign.Right, true
}

func bytecodeF64MatrixRowLoopArrayGetDeclaration(stmt ast.Statement) (string, string, string, bool) {
	name, expr, ok := bytecodeF64MatrixRowLoopDeclaration(stmt)
	if !ok {
		return "", "", "", false
	}
	outerName, indexName, ok := bytecodeSingleArrayGetPropagationNames(expr)
	if !ok {
		return "", "", "", false
	}
	return name, outerName, indexName, true
}

func bytecodeF64MatrixRowLoopAddMulArrayGetNames(expr ast.Expression, accumulatorName string) (string, string, string, string, bool) {
	add, ok := expr.(*ast.BinaryExpression)
	if !ok || add == nil || add.Operator != "+" {
		return "", "", "", "", false
	}
	if bytecodeIdentifierExpressionName(add.Left) != accumulatorName {
		return "", "", "", "", false
	}
	mul, ok := add.Right.(*ast.BinaryExpression)
	if !ok || mul == nil || mul.Operator != "*" {
		return "", "", "", "", false
	}
	leftReceiverName, leftIndexName, ok := bytecodeSingleArrayGetPropagationNames(mul.Left)
	if !ok {
		return "", "", "", "", false
	}
	rightReceiverName, rightIndexName, ok := bytecodeSingleArrayGetPropagationNames(mul.Right)
	if !ok {
		return "", "", "", "", false
	}
	return leftReceiverName, leftIndexName, rightReceiverName, rightIndexName, true
}

func bytecodeSingleArrayGetPropagationNames(expr ast.Expression) (string, string, bool) {
	prop, ok := expr.(*ast.PropagationExpression)
	if !ok || prop == nil {
		return "", "", false
	}
	call, ok := prop.Expression.(*ast.FunctionCall)
	if !ok || call == nil || len(call.Arguments) != 1 || len(call.TypeArguments) != 0 {
		return "", "", false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe || bytecodeIdentifierMemberName(member.Member) != "get" {
		return "", "", false
	}
	receiverName := bytecodeIdentifierExpressionName(member.Object)
	indexName := bytecodeIdentifierExpressionName(call.Arguments[0])
	return receiverName, indexName, receiverName != "" && indexName != ""
}

func bytecodeF64MatrixRowLoopPush(stmt ast.Statement) (string, string, bool) {
	call, ok := stmt.(*ast.FunctionCall)
	if !ok || call == nil || len(call.Arguments) != 1 || len(call.TypeArguments) != 0 {
		return "", "", false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe || bytecodeIdentifierMemberName(member.Member) != "push" {
		return "", "", false
	}
	receiverName := bytecodeIdentifierExpressionName(member.Object)
	argName := bytecodeIdentifierExpressionName(call.Arguments[0])
	return receiverName, argName, receiverName != "" && argName != ""
}

func (plan bytecodeF64MatrixRowLoopPlan) validForSlots(slotCount int) bool {
	return plan.indexSlot >= 0 && plan.indexSlot < slotCount &&
		plan.boundSlot >= 0 && plan.boundSlot < slotCount &&
		plan.resultReceiverSlot >= 0 && plan.resultReceiverSlot < slotCount &&
		plan.leftReceiverSlot >= 0 && plan.leftReceiverSlot < slotCount &&
		plan.rightOuterSlot >= 0 && plan.rightOuterSlot < slotCount &&
		plan.resultPushIP >= 0 &&
		plan.successTarget > 0
}
