package interpreter

import "able/interpreter-go/pkg/ast"

func bytecodeStoreSlotFloatBinaryInstruction(ctx *bytecodeLoweringContext, expr ast.Expression, node ast.Node) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil {
		return bytecodeInstruction{}, false
	}
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil {
		return bytecodeInstruction{}, false
	}
	switch binary.Operator {
	case "+", "-", "*":
	default:
		return bytecodeInstruction{}, false
	}
	left, ok := binary.Left.(*ast.Identifier)
	if !ok || left == nil {
		return bytecodeInstruction{}, false
	}
	right, ok := binary.Right.(*ast.Identifier)
	if !ok || right == nil {
		return bytecodeInstruction{}, false
	}
	leftSlot, found := ctx.lookupSlot(left.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	rightSlot, found := ctx.lookupSlot(right.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	if ctx.slotKind(leftSlot) != bytecodeCellKindValue || ctx.slotKind(rightSlot) != bytecodeCellKindValue {
		return bytecodeInstruction{}, false
	}
	if !bytecodeExpressionIsKnownFloat(ctx, binary) {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:        bytecodeOpStoreSlotFloatBinary,
		operator:  binary.Operator,
		argCount:  leftSlot,
		loopBreak: rightSlot,
		node:      node,
	}, true
}
