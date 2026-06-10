package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeFloatCastTargetKind(target ast.TypeExpression) (runtime.FloatType, bool) {
	simple, ok := target.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return "", false
	}
	switch runtime.FloatType(simple.Name.Name) {
	case runtime.FloatF32:
		return runtime.FloatF32, true
	case runtime.FloatF64:
		return runtime.FloatF64, true
	default:
		return "", false
	}
}

func bytecodeBinaryCastSlotFloatConstInstruction(ctx *bytecodeLoweringContext, expr *ast.BinaryExpression) (bytecodeInstruction, bool) {
	return bytecodeBinaryCastSlotFloatConstInstructionForOperator(ctx, expr, "/", bytecodeOpBinaryCastSlotFloatConstDiv)
}

func bytecodeBinaryCastSlotFloatConstInstructionForOperator(ctx *bytecodeLoweringContext, expr *ast.BinaryExpression, operator string, opcode bytecodeOp) (bytecodeInstruction, bool) {
	if ctx == nil || expr == nil || expr.Operator != operator {
		return bytecodeInstruction{}, false
	}

	trySide := func(castExpr ast.Expression, literalExpr ast.Expression) (bytecodeInstruction, bool) {
		cast, ok := castExpr.(*ast.TypeCastExpression)
		if !ok || cast == nil || cast.Expression == nil || cast.TargetType == nil {
			return bytecodeInstruction{}, false
		}
		targetKind, ok := bytecodeFloatCastTargetKind(cast.TargetType)
		if !ok {
			return bytecodeInstruction{}, false
		}
		ident, ok := cast.Expression.(*ast.Identifier)
		if !ok || ident == nil {
			return bytecodeInstruction{}, false
		}
		slot, found := ctx.lookupSlot(ident.Name)
		if !found {
			return bytecodeInstruction{}, false
		}
		lit, ok := literalExpr.(*ast.FloatLiteral)
		if !ok || lit == nil {
			return bytecodeInstruction{}, false
		}
		rightKind := runtime.FloatF64
		if lit.FloatType != nil {
			rightKind = runtime.FloatType(*lit.FloatType)
		}
		if rightKind != runtime.FloatF32 && rightKind != runtime.FloatF64 {
			return bytecodeInstruction{}, false
		}
		return bytecodeInstruction{
			op:       opcode,
			target:   slot,
			value:    runtime.FloatValue{Val: normalizeFloat(rightKind, lit.Value), TypeSuffix: rightKind},
			typeExpr: ast.Ty(string(targetKind)),
			node:     expr,
		}, true
	}

	return trySide(expr.Left, expr.Right)
}
