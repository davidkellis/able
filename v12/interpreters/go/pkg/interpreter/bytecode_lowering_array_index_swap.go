package interpreter

import "able/interpreter-go/pkg/ast"

func bytecodeArrayIndexSwapSlotInstruction(ctx *bytecodeLoweringContext, body []ast.Statement) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil || len(body) != 3 {
		return bytecodeInstruction{}, false
	}
	first, ok := body[0].(*ast.AssignmentExpression)
	if !ok || first == nil || first.Operator != ast.AssignmentDeclare {
		return bytecodeInstruction{}, false
	}
	temp, ok := first.Left.(*ast.Identifier)
	if !ok || temp == nil || temp.Name == "" {
		return bytecodeInstruction{}, false
	}
	firstIndex, castTarget, ok := bytecodeArrayIndexSwapCast(first.Right)
	if !ok {
		return bytecodeInstruction{}, false
	}
	firstAssign, ok := body[1].(*ast.AssignmentExpression)
	if !ok || firstAssign == nil || firstAssign.Operator != ast.AssignmentAssign {
		return bytecodeInstruction{}, false
	}
	firstTarget, ok := firstAssign.Left.(*ast.IndexExpression)
	if !ok || !bytecodeArrayIndexSwapSameIndex(firstTarget, firstIndex) {
		return bytecodeInstruction{}, false
	}
	secondIndex, secondCastTarget, ok := bytecodeArrayIndexSwapCast(firstAssign.Right)
	if !ok || typeExpressionToString(secondCastTarget) != typeExpressionToString(castTarget) {
		return bytecodeInstruction{}, false
	}
	secondAssign, ok := body[2].(*ast.AssignmentExpression)
	if !ok || secondAssign == nil || secondAssign.Operator != ast.AssignmentAssign {
		return bytecodeInstruction{}, false
	}
	secondTarget, ok := secondAssign.Left.(*ast.IndexExpression)
	if !ok || !bytecodeArrayIndexSwapSameIndex(secondTarget, secondIndex) {
		return bytecodeInstruction{}, false
	}
	resultIdent, ok := secondAssign.Right.(*ast.Identifier)
	if !ok || resultIdent == nil || resultIdent.Name != temp.Name {
		return bytecodeInstruction{}, false
	}
	receiverSlot, firstSlot, ok := bytecodeArrayIndexSwapSlots(ctx, firstIndex)
	if !ok {
		return bytecodeInstruction{}, false
	}
	receiverSlot2, secondSlot, ok := bytecodeArrayIndexSwapSlots(ctx, secondIndex)
	if !ok || receiverSlot2 != receiverSlot {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:           bytecodeOpArrayIndexSwapSlot,
		argCount:     receiverSlot,
		loopBreak:    firstSlot,
		loopContinue: secondSlot,
		typeExpr:     castTarget,
		node:         secondAssign,
	}, true
}

func bytecodeArrayIndexSwapCast(expr ast.Expression) (*ast.IndexExpression, ast.TypeExpression, bool) {
	cast, ok := expr.(*ast.TypeCastExpression)
	if !ok || cast == nil || cast.TargetType == nil {
		return nil, nil, false
	}
	index, ok := cast.Expression.(*ast.IndexExpression)
	if !ok || index == nil {
		return nil, nil, false
	}
	return index, cast.TargetType, true
}

func bytecodeArrayIndexSwapSameIndex(left *ast.IndexExpression, right *ast.IndexExpression) bool {
	if left == nil || right == nil {
		return false
	}
	leftObj, ok := left.Object.(*ast.Identifier)
	if !ok || leftObj == nil {
		return false
	}
	rightObj, ok := right.Object.(*ast.Identifier)
	if !ok || rightObj == nil || rightObj.Name != leftObj.Name {
		return false
	}
	leftIdx, ok := left.Index.(*ast.Identifier)
	if !ok || leftIdx == nil {
		return false
	}
	rightIdx, ok := right.Index.(*ast.Identifier)
	return ok && rightIdx != nil && rightIdx.Name == leftIdx.Name
}

func bytecodeArrayIndexSwapSlots(ctx *bytecodeLoweringContext, expr *ast.IndexExpression) (int, int, bool) {
	if ctx == nil || expr == nil {
		return 0, 0, false
	}
	objIdent, ok := expr.Object.(*ast.Identifier)
	if !ok || objIdent == nil {
		return 0, 0, false
	}
	idxIdent, ok := expr.Index.(*ast.Identifier)
	if !ok || idxIdent == nil {
		return 0, 0, false
	}
	objSlot, ok := ctx.lookupSlot(objIdent.Name)
	if !ok {
		return 0, 0, false
	}
	idxSlot, ok := ctx.lookupSlot(idxIdent.Name)
	return objSlot, idxSlot, ok
}

func bytecodeArraySlotSwapSlotInstruction(ctx *bytecodeLoweringContext, body []ast.Statement) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil || len(body) != 3 {
		return bytecodeInstruction{}, false
	}
	first, ok := body[0].(*ast.AssignmentExpression)
	if !ok || first == nil || first.Operator != ast.AssignmentDeclare {
		return bytecodeInstruction{}, false
	}
	temp, ok := first.Left.(*ast.Identifier)
	if !ok || temp == nil || temp.Name == "" {
		return bytecodeInstruction{}, false
	}
	receiver, firstIndex, ok := bytecodeArraySlotSwapReadCall(first.Right)
	if !ok {
		return bytecodeInstruction{}, false
	}
	second, ok := body[1].(*ast.FunctionCall)
	if !ok || second == nil {
		return bytecodeInstruction{}, false
	}
	secondReceiver, secondWriteIndex, secondValue, ok := bytecodeArraySlotSwapWriteCall(second)
	if !ok || secondReceiver != receiver || secondWriteIndex != firstIndex {
		return bytecodeInstruction{}, false
	}
	secondReadReceiver, secondIndex, ok := bytecodeArraySlotSwapReadCall(secondValue)
	if !ok || secondReadReceiver != receiver {
		return bytecodeInstruction{}, false
	}
	third, ok := body[2].(*ast.FunctionCall)
	if !ok || third == nil {
		return bytecodeInstruction{}, false
	}
	thirdReceiver, thirdWriteIndex, thirdValue, ok := bytecodeArraySlotSwapWriteCall(third)
	if !ok || thirdReceiver != receiver || thirdWriteIndex != secondIndex {
		return bytecodeInstruction{}, false
	}
	resultIdent, ok := thirdValue.(*ast.Identifier)
	if !ok || resultIdent == nil || resultIdent.Name != temp.Name {
		return bytecodeInstruction{}, false
	}
	receiverSlot, ok := ctx.lookupSlot(receiver)
	if !ok {
		return bytecodeInstruction{}, false
	}
	firstSlot, ok := ctx.lookupSlot(firstIndex)
	if !ok {
		return bytecodeInstruction{}, false
	}
	secondSlot, ok := ctx.lookupSlot(secondIndex)
	if !ok {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:           bytecodeOpArraySlotSwapSlot,
		argCount:     receiverSlot,
		loopBreak:    firstSlot,
		loopContinue: secondSlot,
		node:         third,
	}, true
}

func bytecodeArraySlotSwapReadCall(expr ast.Expression) (string, string, bool) {
	call, ok := expr.(*ast.FunctionCall)
	if !ok || call == nil || len(call.Arguments) != 1 {
		return "", "", false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe || bytecodeIdentifierMemberName(member.Member) != "read_slot" {
		return "", "", false
	}
	receiver, ok := member.Object.(*ast.Identifier)
	if !ok || receiver == nil || receiver.Name == "" {
		return "", "", false
	}
	index, ok := call.Arguments[0].(*ast.Identifier)
	if !ok || index == nil || index.Name == "" {
		return "", "", false
	}
	return receiver.Name, index.Name, true
}

func bytecodeArraySlotSwapWriteCall(call *ast.FunctionCall) (string, string, ast.Expression, bool) {
	if call == nil || len(call.Arguments) != 2 {
		return "", "", nil, false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe || bytecodeIdentifierMemberName(member.Member) != "write_slot" {
		return "", "", nil, false
	}
	receiver, ok := member.Object.(*ast.Identifier)
	if !ok || receiver == nil || receiver.Name == "" {
		return "", "", nil, false
	}
	index, ok := call.Arguments[0].(*ast.Identifier)
	if !ok || index == nil || index.Name == "" {
		return "", "", nil, false
	}
	return receiver.Name, index.Name, call.Arguments[1], true
}
