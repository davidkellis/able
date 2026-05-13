package interpreter

import "able/interpreter-go/pkg/ast"

type bytecodeF64DotLoopPlan struct {
	accumulatorSlot    int
	indexSlot          int
	boundSlot          int
	leftReceiverSlot   int
	rightReceiverSlot  int
	successTarget      int
	resultAppend       bool
	resultReceiverSlot int
	resultPushIP       int
	resultTarget       int
}

func (ctx *bytecodeLoweringContext) setF64DotLoopPlan(index int, plan bytecodeF64DotLoopPlan) {
	if ctx == nil || index < 0 {
		return
	}
	if ctx.f64DotLoops == nil {
		ctx.f64DotLoops = make(map[int]bytecodeF64DotLoopPlan, 1)
	}
	ctx.f64DotLoops[index] = plan
}

func bytecodeF64DotLoopPlanForLoop(ctx *bytecodeLoweringContext, loop *ast.LoopExpression) (bytecodeF64DotLoopPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes || loop == nil || loop.Body == nil || len(loop.Body.Body) != 3 {
		return bytecodeF64DotLoopPlan{}, false
	}
	indexName, boundName, ok := bytecodeF64DotLoopBreakCondition(loop.Body.Body[0])
	if !ok {
		return bytecodeF64DotLoopPlan{}, false
	}
	accumulatorName, addMulExpr, ok := bytecodeF64DotLoopAssignment(loop.Body.Body[1])
	if !ok {
		return bytecodeF64DotLoopPlan{}, false
	}
	if !bytecodeF64DotLoopIncrement(loop.Body.Body[2], indexName) {
		return bytecodeF64DotLoopPlan{}, false
	}
	addMulPlan, ok := bytecodeStoreSlotFloatAddMulArrayGetPlan(ctx, accumulatorName, addMulExpr, loop.Body.Body[1])
	if !ok || addMulPlan.leftIndexName != indexName || addMulPlan.rightIndexName != indexName {
		return bytecodeF64DotLoopPlan{}, false
	}
	indexSlot, found := ctx.lookupSlot(indexName)
	if !found {
		return bytecodeF64DotLoopPlan{}, false
	}
	boundSlot, found := ctx.lookupSlot(boundName)
	if !found {
		return bytecodeF64DotLoopPlan{}, false
	}
	return bytecodeF64DotLoopPlan{
		accumulatorSlot:   addMulPlan.targetSlot,
		indexSlot:         indexSlot,
		boundSlot:         boundSlot,
		leftReceiverSlot:  addMulPlan.leftReceiverSlot,
		rightReceiverSlot: addMulPlan.rightReceiverSlot,
	}, true
}

func bytecodeF64DotLoopResultAppendReceiverSlot(ctx *bytecodeLoweringContext, loop *ast.LoopExpression, stmt ast.Statement) (int, bool) {
	if ctx == nil || ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes || loop == nil || loop.Body == nil || len(loop.Body.Body) != 3 || stmt == nil {
		return -1, false
	}
	accumulatorName, _, ok := bytecodeF64DotLoopAssignment(loop.Body.Body[1])
	if !ok || accumulatorName == "" {
		return -1, false
	}
	call, ok := stmt.(*ast.FunctionCall)
	if !ok || call == nil || len(call.Arguments) != 1 || len(call.TypeArguments) != 0 {
		return -1, false
	}
	if bytecodeIdentifierExpressionName(call.Arguments[0]) != accumulatorName {
		return -1, false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe || bytecodeIdentifierMemberName(member.Member) != "push" {
		return -1, false
	}
	receiverName := bytecodeIdentifierExpressionName(member.Object)
	if receiverName == "" {
		return -1, false
	}
	slot, found := ctx.lookupSlot(receiverName)
	return slot, found
}

func bytecodeF64DotLoopBreakCondition(stmt ast.Statement) (string, string, bool) {
	ifExpr, ok := stmt.(*ast.IfExpression)
	if !ok || ifExpr == nil || ifExpr.ElseBody != nil || len(ifExpr.ElseIfClauses) != 0 || ifExpr.IfBody == nil || len(ifExpr.IfBody.Body) != 1 {
		return "", "", false
	}
	br, ok := ifExpr.IfBody.Body[0].(*ast.BreakStatement)
	if !ok || br == nil || br.Label != nil || br.Value != nil {
		return "", "", false
	}
	cond, ok := ifExpr.IfCondition.(*ast.BinaryExpression)
	if !ok || cond == nil || cond.Operator != ">=" {
		return "", "", false
	}
	left, ok := cond.Left.(*ast.Identifier)
	if !ok || left == nil {
		return "", "", false
	}
	right, ok := cond.Right.(*ast.Identifier)
	if !ok || right == nil {
		return "", "", false
	}
	return left.Name, right.Name, left.Name != "" && right.Name != ""
}

func bytecodeF64DotLoopAssignment(stmt ast.Statement) (string, ast.Expression, bool) {
	assign, ok := stmt.(*ast.AssignmentExpression)
	if !ok || assign == nil || assign.Operator != ast.AssignmentAssign {
		return "", nil, false
	}
	target, ok := assign.Left.(*ast.Identifier)
	if !ok || target == nil || target.Name == "" || assign.Right == nil {
		return "", nil, false
	}
	return target.Name, assign.Right, true
}

func bytecodeF64DotLoopIncrement(stmt ast.Statement, indexName string) bool {
	targetName, expr, ok := bytecodeF64DotLoopAssignment(stmt)
	if !ok || targetName != indexName {
		return false
	}
	add, ok := expr.(*ast.BinaryExpression)
	if !ok || add == nil || add.Operator != "+" {
		return false
	}
	left, ok := add.Left.(*ast.Identifier)
	if !ok || left == nil || left.Name != indexName {
		return false
	}
	lit, ok := add.Right.(*ast.IntegerLiteral)
	if !ok || lit == nil {
		return false
	}
	_, raw, ok := bytecodeSlotConstIntegerLiteralImmediate(lit)
	return ok && raw == 1
}
