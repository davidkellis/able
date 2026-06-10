package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeFloatAddCompareConstJumpPlan struct {
	leftSlot       int
	rightSlot      int
	rightImmediate runtime.FloatValue
}

func (ctx *bytecodeLoweringContext) setFloatAddCompareConstJumpPlan(index int, plan bytecodeFloatAddCompareConstJumpPlan) {
	if ctx == nil || index < 0 {
		return
	}
	if ctx.floatAddCompareConstJumps == nil {
		ctx.floatAddCompareConstJumps = make(map[int]bytecodeFloatAddCompareConstJumpPlan, 1)
	}
	ctx.floatAddCompareConstJumps[index] = plan
}

func bytecodeJumpIfFalseFloatAddCompareConstInstruction(ctx *bytecodeLoweringContext, expr ast.Expression) (bytecodeInstruction, bytecodeFloatAddCompareConstJumpPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil {
		return bytecodeInstruction{}, bytecodeFloatAddCompareConstJumpPlan{}, false
	}
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil || !bytecodeComparisonOperator(binary.Operator) {
		return bytecodeInstruction{}, bytecodeFloatAddCompareConstJumpPlan{}, false
	}
	add, ok := binary.Left.(*ast.BinaryExpression)
	if !ok || add == nil || add.Operator != "+" {
		return bytecodeInstruction{}, bytecodeFloatAddCompareConstJumpPlan{}, false
	}
	leftSlot, rightSlot, ok := bytecodeFloatIdentifierAddSlots(ctx, add)
	if !ok {
		return bytecodeInstruction{}, bytecodeFloatAddCompareConstJumpPlan{}, false
	}
	rightLiteral, ok := binary.Right.(*ast.FloatLiteral)
	if !ok || rightLiteral == nil {
		return bytecodeInstruction{}, bytecodeFloatAddCompareConstJumpPlan{}, false
	}
	rightKind := runtime.FloatF64
	if rightLiteral.FloatType != nil {
		rightKind = runtime.FloatType(*rightLiteral.FloatType)
	}
	if rightKind != runtime.FloatF32 && rightKind != runtime.FloatF64 {
		return bytecodeInstruction{}, bytecodeFloatAddCompareConstJumpPlan{}, false
	}
	plan := bytecodeFloatAddCompareConstJumpPlan{
		leftSlot:  leftSlot,
		rightSlot: rightSlot,
		rightImmediate: runtime.FloatValue{
			Val:        normalizeFloat(rightKind, rightLiteral.Value),
			TypeSuffix: rightKind,
		},
	}
	return bytecodeInstruction{
		op:       bytecodeOpJumpIfFloatAddCompareConstFalse,
		target:   -1,
		operator: binary.Operator,
		node:     binary,
	}, plan, true
}

func bytecodeFloatIdentifierAddSlots(ctx *bytecodeLoweringContext, expr ast.Expression) (int, int, bool) {
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil || binary.Operator != "+" {
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
