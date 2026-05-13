package interpreter

import "able/interpreter-go/pkg/ast"

func bytecodeComparisonOperator(op string) bool {
	switch op {
	case "<", "<=", ">", ">=", "==", "!=":
		return true
	default:
		return false
	}
}

func bytecodeJumpIfFalseBinarySlotSlotInstruction(ctx *bytecodeLoweringContext, expr ast.Expression) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil {
		return bytecodeInstruction{}, false
	}
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil || !bytecodeComparisonOperator(binary.Operator) {
		return bytecodeInstruction{}, false
	}
	leftIdent, ok := binary.Left.(*ast.Identifier)
	if !ok || leftIdent == nil {
		return bytecodeInstruction{}, false
	}
	rightIdent, ok := binary.Right.(*ast.Identifier)
	if !ok || rightIdent == nil {
		return bytecodeInstruction{}, false
	}
	leftSlot, found := ctx.lookupSlot(leftIdent.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	rightSlot, found := ctx.lookupSlot(rightIdent.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:        bytecodeOpJumpIfIntCompareSlotFalse,
		target:    -1,
		argCount:  leftSlot,
		loopBreak: rightSlot,
		operator:  binary.Operator,
		node:      binary,
	}, true
}

func bytecodeJumpIfFalseArrayReadSlotCompareSlotInstruction(ctx *bytecodeLoweringContext, expr ast.Expression) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil {
		return bytecodeInstruction{}, false
	}
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil || !bytecodeComparisonOperator(binary.Operator) {
		return bytecodeInstruction{}, false
	}
	call, ok := binary.Left.(*ast.FunctionCall)
	if !ok || call == nil || len(call.Arguments) != 1 {
		return bytecodeInstruction{}, false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe || bytecodeIdentifierMemberName(member.Member) != "read_slot" {
		return bytecodeInstruction{}, false
	}
	receiverIdent, ok := member.Object.(*ast.Identifier)
	if !ok || receiverIdent == nil {
		return bytecodeInstruction{}, false
	}
	indexIdent, ok := call.Arguments[0].(*ast.Identifier)
	if !ok || indexIdent == nil {
		return bytecodeInstruction{}, false
	}
	rightIdent, ok := binary.Right.(*ast.Identifier)
	if !ok || rightIdent == nil {
		return bytecodeInstruction{}, false
	}
	receiverSlot, found := ctx.lookupSlot(receiverIdent.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	indexSlot, found := ctx.lookupSlot(indexIdent.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	rightSlot, found := ctx.lookupSlot(rightIdent.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:           bytecodeOpJumpIfArrayReadSlotCompareSlotFalse,
		target:       -1,
		argCount:     receiverSlot,
		loopBreak:    indexSlot,
		loopContinue: rightSlot,
		name:         "read_slot",
		operator:     binary.Operator,
		node:         binary,
	}, true
}

func bytecodeJumpIfFalseArrayIndexSlotCompareSlotInstruction(ctx *bytecodeLoweringContext, expr ast.Expression) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil {
		return bytecodeInstruction{}, false
	}
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil || !bytecodeComparisonOperator(binary.Operator) {
		return bytecodeInstruction{}, false
	}
	indexExpr, castTarget, ok := bytecodeArrayIndexCompareLeft(binary.Left)
	if !ok || indexExpr == nil {
		return bytecodeInstruction{}, false
	}
	receiverIdent, ok := indexExpr.Object.(*ast.Identifier)
	if !ok || receiverIdent == nil {
		return bytecodeInstruction{}, false
	}
	indexIdent, ok := indexExpr.Index.(*ast.Identifier)
	if !ok || indexIdent == nil {
		return bytecodeInstruction{}, false
	}
	rightIdent, ok := binary.Right.(*ast.Identifier)
	if !ok || rightIdent == nil {
		return bytecodeInstruction{}, false
	}
	receiverSlot, found := ctx.lookupSlot(receiverIdent.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	indexSlot, found := ctx.lookupSlot(indexIdent.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	rightSlot, found := ctx.lookupSlot(rightIdent.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	castName := ""
	if castTarget != nil {
		castName = typeExpressionToString(castTarget)
	}
	return bytecodeInstruction{
		op:           bytecodeOpJumpIfArrayIndexSlotCompareSlotFalse,
		name:         castName,
		target:       -1,
		argCount:     receiverSlot,
		loopBreak:    indexSlot,
		loopContinue: rightSlot,
		operator:     binary.Operator,
		typeExpr:     castTarget,
		node:         binary,
	}, true
}

func bytecodeArrayIndexCompareLeft(expr ast.Expression) (*ast.IndexExpression, ast.TypeExpression, bool) {
	var castTarget ast.TypeExpression
	if cast, ok := expr.(*ast.TypeCastExpression); ok {
		if cast == nil || cast.TargetType == nil || typeExpressionToString(cast.TargetType) != "i32" {
			return nil, nil, false
		}
		castTarget = cast.TargetType
		expr = cast.Expression
	}
	indexExpr, ok := expr.(*ast.IndexExpression)
	if !ok || indexExpr == nil {
		return nil, nil, false
	}
	return indexExpr, castTarget, true
}

func bytecodeArrayReadSlotInstruction(ctx *bytecodeLoweringContext, call *ast.FunctionCall, member *ast.MemberAccessExpression, memberName string) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil || call == nil || member == nil || member.Safe || memberName != "read_slot" || len(call.Arguments) != 1 {
		return bytecodeInstruction{}, false
	}
	receiverIdent, ok := member.Object.(*ast.Identifier)
	if !ok || receiverIdent == nil {
		return bytecodeInstruction{}, false
	}
	indexIdent, ok := call.Arguments[0].(*ast.Identifier)
	if !ok || indexIdent == nil {
		return bytecodeInstruction{}, false
	}
	receiverSlot, found := ctx.lookupSlot(receiverIdent.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	indexSlot, found := ctx.lookupSlot(indexIdent.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:        bytecodeOpArrayReadSlot,
		argCount:  receiverSlot,
		loopBreak: indexSlot,
		name:      "read_slot",
		node:      call,
	}, true
}
