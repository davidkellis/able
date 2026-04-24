package interpreter

import (
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeBinarySlotConstInstruction(ctx *bytecodeLoweringContext, expr *ast.BinaryExpression) (bytecodeInstruction, bool) {
	if ctx == nil || expr == nil {
		return bytecodeInstruction{}, false
	}
	ident, ok := expr.Left.(*ast.Identifier)
	if !ok || ident == nil {
		return bytecodeInstruction{}, false
	}
	slot, found := ctx.lookupSlot(ident.Name)
	if !found {
		return bytecodeInstruction{}, false
	}
	lit, ok := expr.Right.(*ast.IntegerLiteral)
	if !ok || lit == nil || lit.Value == nil || lit.IntegerType != nil || !lit.Value.IsInt64() {
		return bytecodeInstruction{}, false
	}
	litVal := lit.Value.Int64()
	if litVal < math.MinInt32 || litVal > math.MaxInt32 {
		return bytecodeInstruction{}, false
	}
	imm := runtime.NewSmallInt(litVal, runtime.IntegerI32)
	switch expr.Operator {
	case "+":
		return bytecodeInstruction{
			op:              bytecodeOpBinaryIntAddSlotConst,
			target:          slot,
			value:           imm,
			intImmediate:    imm,
			hasIntImmediate: true,
			operator:        expr.Operator,
			node:            expr,
		}, true
	case "-":
		return bytecodeInstruction{
			op:              bytecodeOpBinaryIntSubSlotConst,
			target:          slot,
			value:           imm,
			intImmediate:    imm,
			hasIntImmediate: true,
			operator:        expr.Operator,
			node:            expr,
		}, true
	case "<=":
		return bytecodeInstruction{
			op:              bytecodeOpBinaryIntLessEqualSlotConst,
			target:          slot,
			value:           imm,
			intImmediate:    imm,
			hasIntImmediate: true,
			operator:        expr.Operator,
			node:            expr,
		}, true
	default:
		return bytecodeInstruction{}, false
	}
}

func bytecodeJumpIfFalseBinarySlotConstInstruction(ctx *bytecodeLoweringContext, expr ast.Expression) (bytecodeInstruction, bool) {
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil {
		return bytecodeInstruction{}, false
	}
	instr, ok := bytecodeBinarySlotConstInstruction(ctx, binary)
	if !ok || instr.op != bytecodeOpBinaryIntLessEqualSlotConst {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:              bytecodeOpJumpIfIntLessEqualSlotConstFalse,
		target:          -1,
		argCount:        instr.target,
		value:           instr.value,
		intImmediate:    instr.intImmediate,
		hasIntImmediate: instr.hasIntImmediate,
		operator:        instr.operator,
		node:            instr.node,
	}, true
}
