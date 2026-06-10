package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeFloatUpdatePairPlan struct {
	firstTargetSlot      int
	firstBaseSlot        int
	firstMulSlot         int
	firstScaleSourceSlot int
	firstScaleImmediate  runtime.FloatValue
	secondTargetSlot     int
	secondBaseSlot       int
	secondSubLeftSlot    int
	secondSubRightSlot   int
}

func (ctx *bytecodeLoweringContext) setFloatUpdatePairPlan(index int, plan bytecodeFloatUpdatePairPlan) {
	if ctx == nil || index < 0 {
		return
	}
	if ctx.floatUpdatePairs == nil {
		ctx.floatUpdatePairs = make(map[int]bytecodeFloatUpdatePairPlan, 1)
	}
	ctx.floatUpdatePairs[index] = plan
}

func emitTryFloatUpdatePair(ctx *bytecodeLoweringContext, i *Interpreter, statements []ast.Statement) (bool, error) {
	if ctx == nil || ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes || len(statements) < 2 {
		return false, nil
	}
	firstAssign, ok := statements[0].(*ast.AssignmentExpression)
	if !ok || firstAssign == nil || firstAssign.Operator != ast.AssignmentAssign {
		return false, nil
	}
	secondAssign, ok := statements[1].(*ast.AssignmentExpression)
	if !ok || secondAssign == nil || secondAssign.Operator != ast.AssignmentAssign {
		return false, nil
	}
	firstTargetName, ok := resolveAssignmentTargetName(firstAssign.Left)
	if !ok {
		return false, nil
	}
	secondTargetName, ok := resolveAssignmentTargetName(secondAssign.Left)
	if !ok {
		return false, nil
	}
	firstPlan, ok := bytecodeStoreSlotFloatAddMulSlotPlan(ctx, firstTargetName, firstAssign.Right, firstAssign)
	if !ok {
		return false, nil
	}
	secondPlan, ok := bytecodeStoreSlotFloatAddSubPlan(ctx, secondTargetName, secondAssign.Right, secondAssign)
	if !ok {
		return false, nil
	}
	if secondPlan.baseSlot == firstPlan.targetSlot || secondPlan.subLeftSlot == firstPlan.targetSlot || secondPlan.subRightSlot == firstPlan.targetSlot {
		return false, nil
	}
	stackExpr, ok := firstPlan.stackExpr.(*ast.BinaryExpression)
	if !ok || stackExpr == nil {
		return false, nil
	}
	scaleInstr, ok := bytecodeBinaryFloatMulSlotConstInstruction(ctx, stackExpr)
	if !ok {
		return false, nil
	}
	scaleVal, scaleKind, ok := bytecodeInstructionFloatImmediateRaw(&scaleInstr)
	if !ok {
		return false, nil
	}

	fastIP := ctx.emit(bytecodeInstruction{
		op:     bytecodeOpTryFloatUpdatePair,
		target: -1,
		node:   firstAssign,
	})
	ctx.setFloatUpdatePairPlan(fastIP, bytecodeFloatUpdatePairPlan{
		firstTargetSlot:      firstPlan.targetSlot,
		firstBaseSlot:        firstPlan.baseSlot,
		firstMulSlot:         firstPlan.mulSlot,
		firstScaleSourceSlot: scaleInstr.target,
		firstScaleImmediate: runtime.FloatValue{
			Val:        scaleVal,
			TypeSuffix: scaleKind,
		},
		secondTargetSlot:   secondPlan.instr.target,
		secondBaseSlot:     secondPlan.baseSlot,
		secondSubLeftSlot:  secondPlan.subLeftSlot,
		secondSubRightSlot: secondPlan.subRightSlot,
	})

	if err := emitStatement(ctx, i, firstAssign, false); err != nil {
		return true, err
	}
	if err := emitStatement(ctx, i, secondAssign, false); err != nil {
		return true, err
	}
	ctx.patchJump(fastIP, len(ctx.instructions))
	return true, nil
}

func (plan bytecodeFloatUpdatePairPlan) validForSlots(slotCount int) bool {
	return plan.firstTargetSlot >= 0 && plan.firstTargetSlot < slotCount &&
		plan.firstBaseSlot >= 0 && plan.firstBaseSlot < slotCount &&
		plan.firstMulSlot >= 0 && plan.firstMulSlot < slotCount &&
		plan.firstScaleSourceSlot >= 0 && plan.firstScaleSourceSlot < slotCount &&
		plan.secondTargetSlot >= 0 && plan.secondTargetSlot < slotCount &&
		plan.secondBaseSlot >= 0 && plan.secondBaseSlot < slotCount &&
		plan.secondSubLeftSlot >= 0 && plan.secondSubLeftSlot < slotCount &&
		plan.secondSubRightSlot >= 0 && plan.secondSubRightSlot < slotCount
}
