package interpreter

import "able/interpreter-go/pkg/ast"

func bytecodeStoreSlotCastSlotFloatConstDivInstruction(ctx *bytecodeLoweringContext, expr ast.Expression, node ast.Node) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil {
		return bytecodeInstruction{}, false
	}
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil {
		return bytecodeInstruction{}, false
	}
	instr, ok := bytecodeBinaryCastSlotFloatConstInstruction(ctx, binary)
	if !ok {
		return bytecodeInstruction{}, false
	}
	sourceSlot := instr.target
	instr.op = bytecodeOpStoreSlotCastSlotFloatConstDiv
	instr.target = -1
	instr.argCount = sourceSlot
	instr.node = node
	return instr, true
}
