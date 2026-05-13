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
