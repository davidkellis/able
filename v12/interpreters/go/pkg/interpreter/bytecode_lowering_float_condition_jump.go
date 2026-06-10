package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeFloatMulAddMulCompareConstJumpPlan struct {
	leftMulLeftSlot   int
	leftMulRightSlot  int
	rightMulLeftSlot  int
	rightMulRightSlot int
	rightImmediate    runtime.FloatValue
	storeProducts     bool
	leftTargetSlot    int
	rightTargetSlot   int
}

func (ctx *bytecodeLoweringContext) setFloatMulAddMulCompareConstJumpPlan(index int, plan bytecodeFloatMulAddMulCompareConstJumpPlan) {
	if ctx == nil || index < 0 {
		return
	}
	if ctx.floatMulAddMulJumps == nil {
		ctx.floatMulAddMulJumps = make(map[int]bytecodeFloatMulAddMulCompareConstJumpPlan, 1)
	}
	ctx.floatMulAddMulJumps[index] = plan
}

func bytecodeJumpIfFalseFloatMulAddMulCompareConstInstruction(ctx *bytecodeLoweringContext, expr ast.Expression) (bytecodeInstruction, bytecodeFloatMulAddMulCompareConstJumpPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil || !bytecodeComparisonOperator(binary.Operator) {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	add, ok := binary.Left.(*ast.BinaryExpression)
	if !ok || add == nil || add.Operator != "+" {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	leftMulLeftSlot, leftMulRightSlot, ok := bytecodeFloatIdentifierMultiplySlots(ctx, add.Left)
	if !ok {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	rightMulLeftSlot, rightMulRightSlot, ok := bytecodeFloatIdentifierMultiplySlots(ctx, add.Right)
	if !ok {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	rightLiteral, ok := binary.Right.(*ast.FloatLiteral)
	if !ok || rightLiteral == nil {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	rightKind := runtime.FloatF64
	if rightLiteral.FloatType != nil {
		rightKind = runtime.FloatType(*rightLiteral.FloatType)
	}
	if rightKind != runtime.FloatF32 && rightKind != runtime.FloatF64 {
		return bytecodeInstruction{}, bytecodeFloatMulAddMulCompareConstJumpPlan{}, false
	}
	plan := bytecodeFloatMulAddMulCompareConstJumpPlan{
		leftMulLeftSlot:   leftMulLeftSlot,
		leftMulRightSlot:  leftMulRightSlot,
		rightMulLeftSlot:  rightMulLeftSlot,
		rightMulRightSlot: rightMulRightSlot,
		rightImmediate: runtime.FloatValue{
			Val:        normalizeFloat(rightKind, rightLiteral.Value),
			TypeSuffix: rightKind,
		},
	}
	return bytecodeInstruction{
		op:       bytecodeOpJumpIfFloatMulAddMulCompareConstFalse,
		target:   -1,
		operator: binary.Operator,
		node:     binary,
	}, plan, true
}

func bytecodeFloatIdentifierMultiplySlots(ctx *bytecodeLoweringContext, expr ast.Expression) (int, int, bool) {
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil || binary.Operator != "*" {
		return 0, 0, false
	}
	leftIdent, ok := binary.Left.(*ast.Identifier)
	if !ok || leftIdent == nil {
		return 0, 0, false
	}
	rightIdent, ok := binary.Right.(*ast.Identifier)
	if !ok || rightIdent == nil {
		return 0, 0, false
	}
	leftSlot, found := ctx.lookupSlot(leftIdent.Name)
	if !found {
		return 0, 0, false
	}
	rightSlot, found := ctx.lookupSlot(rightIdent.Name)
	if !found {
		return 0, 0, false
	}
	return leftSlot, rightSlot, true
}
