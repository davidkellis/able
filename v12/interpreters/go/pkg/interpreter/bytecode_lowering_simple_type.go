package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (ctx *bytecodeLoweringContext) setSlotSimpleCheck(slot int, check bytecodeSimpleTypeCheck) {
	if slot < 0 {
		return
	}
	for len(ctx.slotSimpleChecks) <= slot {
		ctx.slotSimpleChecks = append(ctx.slotSimpleChecks, bytecodeSimpleTypeCheckUnknown)
	}
	ctx.slotSimpleChecks[slot] = check
}

func (ctx *bytecodeLoweringContext) slotSimpleCheck(slot int) bytecodeSimpleTypeCheck {
	if slot < 0 || slot >= len(ctx.slotSimpleChecks) {
		return bytecodeSimpleTypeCheckUnknown
	}
	return ctx.slotSimpleChecks[slot]
}

func bytecodeIsFloatSimpleCheck(check bytecodeSimpleTypeCheck) bool {
	switch check {
	case bytecodeSimpleTypeCheckAnyFloat, bytecodeSimpleTypeCheckF32, bytecodeSimpleTypeCheckF64:
		return true
	default:
		return false
	}
}

func bytecodeIsNumericSimpleCheck(check bytecodeSimpleTypeCheck) bool {
	if check == bytecodeSimpleTypeCheckAnyInteger || bytecodeIsFloatSimpleCheck(check) {
		return true
	}
	_, ok := check.integerType()
	return ok
}

func bytecodeFloatLiteralSimpleCheck(lit *ast.FloatLiteral) bytecodeSimpleTypeCheck {
	if lit == nil || lit.FloatType == nil {
		return bytecodeSimpleTypeCheckF64
	}
	switch runtime.FloatType(*lit.FloatType) {
	case runtime.FloatF32:
		return bytecodeSimpleTypeCheckF32
	case runtime.FloatF64:
		return bytecodeSimpleTypeCheckF64
	default:
		return bytecodeSimpleTypeCheckUnknown
	}
}

func bytecodeFloatArithmeticResultCheck(left bytecodeSimpleTypeCheck, right bytecodeSimpleTypeCheck) bytecodeSimpleTypeCheck {
	if !bytecodeIsNumericSimpleCheck(left) || !bytecodeIsNumericSimpleCheck(right) {
		return bytecodeSimpleTypeCheckUnknown
	}
	if !bytecodeIsFloatSimpleCheck(left) && !bytecodeIsFloatSimpleCheck(right) {
		return bytecodeSimpleTypeCheckUnknown
	}
	if left == bytecodeSimpleTypeCheckF64 || right == bytecodeSimpleTypeCheckF64 {
		return bytecodeSimpleTypeCheckF64
	}
	if left == bytecodeSimpleTypeCheckF32 && right == bytecodeSimpleTypeCheckF32 {
		return bytecodeSimpleTypeCheckF32
	}
	return bytecodeSimpleTypeCheckAnyFloat
}

func bytecodeExpressionSimpleTypeCheck(ctx *bytecodeLoweringContext, expr ast.Expression) bytecodeSimpleTypeCheck {
	switch n := expr.(type) {
	case nil:
		return bytecodeSimpleTypeCheckUnknown
	case *ast.Identifier:
		if ctx == nil {
			return bytecodeSimpleTypeCheckUnknown
		}
		slot, ok := ctx.lookupSlot(n.Name)
		if !ok {
			return bytecodeSimpleTypeCheckUnknown
		}
		return ctx.slotSimpleCheck(slot)
	case *ast.FloatLiteral:
		return bytecodeFloatLiteralSimpleCheck(n)
	case *ast.IntegerLiteral:
		if n == nil || n.IntegerType == nil {
			return bytecodeSimpleTypeCheckAnyInteger
		}
		return bytecodeSimpleTypeCheckForName(string(*n.IntegerType))
	case *ast.BooleanLiteral:
		return bytecodeSimpleTypeCheckBool
	case *ast.StringLiteral:
		return bytecodeSimpleTypeCheckString
	case *ast.TypeCastExpression:
		if n == nil {
			return bytecodeSimpleTypeCheckUnknown
		}
		return bytecodeSimpleTypeCheckForName(cachedSimpleTypeName(n.TargetType))
	case *ast.UnaryExpression:
		if n == nil || n.Operator != ast.UnaryOperatorNegate {
			return bytecodeSimpleTypeCheckUnknown
		}
		operand := bytecodeExpressionSimpleTypeCheck(ctx, n.Operand)
		if _, ok := operand.floatType(); ok {
			return operand
		}
		return bytecodeSimpleTypeCheckUnknown
	case *ast.BinaryExpression:
		if n == nil {
			return bytecodeSimpleTypeCheckUnknown
		}
		left := bytecodeExpressionSimpleTypeCheck(ctx, n.Left)
		right := bytecodeExpressionSimpleTypeCheck(ctx, n.Right)
		switch n.Operator {
		case "+", "-", "*":
			return bytecodeFloatArithmeticResultCheck(left, right)
		case "/":
			if !bytecodeIsNumericSimpleCheck(left) || !bytecodeIsNumericSimpleCheck(right) {
				return bytecodeSimpleTypeCheckUnknown
			}
			if left == bytecodeSimpleTypeCheckF64 || right == bytecodeSimpleTypeCheckF64 {
				return bytecodeSimpleTypeCheckF64
			}
			if left == bytecodeSimpleTypeCheckF32 && right == bytecodeSimpleTypeCheckF32 {
				return bytecodeSimpleTypeCheckF32
			}
			if bytecodeIsFloatSimpleCheck(left) || bytecodeIsFloatSimpleCheck(right) || left == bytecodeSimpleTypeCheckAnyInteger || right == bytecodeSimpleTypeCheckAnyInteger {
				return bytecodeSimpleTypeCheckAnyFloat
			}
			return bytecodeSimpleTypeCheckF64
		case "<", "<=", ">", ">=", "==", "!=":
			return bytecodeSimpleTypeCheckBool
		default:
			return bytecodeSimpleTypeCheckUnknown
		}
	default:
		return bytecodeSimpleTypeCheckUnknown
	}
}

func bytecodeExpressionIsKnownFloat(ctx *bytecodeLoweringContext, expr ast.Expression) bool {
	return bytecodeIsFloatSimpleCheck(bytecodeExpressionSimpleTypeCheck(ctx, expr))
}
