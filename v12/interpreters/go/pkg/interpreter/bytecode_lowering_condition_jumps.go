package interpreter

import "able/interpreter-go/pkg/ast"

func emitJumpIfFalseCondition(ctx *bytecodeLoweringContext, i *Interpreter, expr ast.Expression) ([]int, error) {
	if jumps, ok := bytecodeEmitJumpIfFalseSlotConstConjunction(ctx, expr); ok {
		return jumps, nil
	}
	if instr, plan, ok := bytecodeJumpIfFalseFloatMulAddMulCompareConstInstruction(ctx, expr); ok {
		jump := ctx.emit(instr)
		ctx.setFloatMulAddMulCompareConstJumpPlan(jump, plan)
		return []int{jump}, nil
	}
	if instr, plan, ok := bytecodeJumpIfFalseFloatAddCompareConstInstruction(ctx, expr); ok {
		jump := ctx.emit(instr)
		ctx.setFloatAddCompareConstJumpPlan(jump, plan)
		return []int{jump}, nil
	}
	if instr, ok := bytecodeJumpIfFalseBinarySlotSlotInstruction(ctx, expr); ok {
		return []int{ctx.emit(instr)}, nil
	}
	if instr, ok := bytecodeJumpIfFalseArrayReadSlotCompareSlotInstruction(ctx, expr); ok {
		return []int{ctx.emit(instr)}, nil
	}
	if instr, ok := bytecodeJumpIfFalseArrayIndexSlotCompareSlotInstruction(ctx, expr); ok {
		return []int{ctx.emit(instr)}, nil
	}
	if instr, ok := bytecodeJumpIfFalseBinarySlotConstInstruction(ctx, expr); ok {
		return []int{ctx.emit(instr)}, nil
	}
	if instr, ok := bytecodeJumpIfFalseBoolSlotInstruction(ctx, expr); ok {
		return []int{ctx.emit(instr)}, nil
	}
	if err := emitExpression(ctx, i, expr); err != nil {
		return nil, err
	}
	return []int{ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfFalse, target: -1})}, nil
}

func bytecodeEmitJumpIfFalseSlotConstConjunction(ctx *bytecodeLoweringContext, expr ast.Expression) ([]int, bool) {
	instrs, ok := bytecodeJumpIfFalseSlotConstConjunctionInstructions(ctx, expr)
	if !ok {
		return nil, false
	}
	jumps := make([]int, 0, len(instrs))
	for _, instr := range instrs {
		jumps = append(jumps, ctx.emit(instr))
	}
	return jumps, true
}

func bytecodeJumpIfFalseSlotConstConjunctionInstructions(ctx *bytecodeLoweringContext, expr ast.Expression) ([]bytecodeInstruction, bool) {
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil || binary.Operator != "&&" {
		instr, ok := bytecodeJumpIfFalseBinarySlotConstInstruction(ctx, expr)
		if !ok {
			return nil, false
		}
		return []bytecodeInstruction{instr}, true
	}
	left, ok := bytecodeJumpIfFalseSlotConstConjunctionInstructions(ctx, binary.Left)
	if !ok {
		return nil, false
	}
	right, ok := bytecodeJumpIfFalseSlotConstConjunctionInstructions(ctx, binary.Right)
	if !ok {
		return nil, false
	}
	return append(left, right...), true
}

func patchConditionalFalseJumps(ctx *bytecodeLoweringContext, jumps []int, target int) {
	for _, jump := range jumps {
		ctx.patchJump(jump, target)
	}
}
