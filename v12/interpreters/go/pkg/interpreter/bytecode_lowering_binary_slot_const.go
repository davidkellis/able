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
			intImmediateRaw: litVal,
			hasIntImmediate: true,
			hasIntRaw:       true,
			operator:        expr.Operator,
			node:            expr,
		}, true
	case "-":
		return bytecodeInstruction{
			op:              bytecodeOpBinaryIntSubSlotConst,
			target:          slot,
			value:           imm,
			intImmediate:    imm,
			intImmediateRaw: litVal,
			hasIntImmediate: true,
			hasIntRaw:       true,
			operator:        expr.Operator,
			node:            expr,
		}, true
	case "<=":
		return bytecodeInstruction{
			op:              bytecodeOpBinaryIntLessEqualSlotConst,
			target:          slot,
			value:           imm,
			intImmediate:    imm,
			intImmediateRaw: litVal,
			hasIntImmediate: true,
			hasIntRaw:       true,
			operator:        expr.Operator,
			node:            expr,
		}, true
	case ">", ">=":
		return bytecodeInstruction{
			op:              bytecodeOpBinaryIntCompareSlotConst,
			target:          slot,
			value:           imm,
			intImmediate:    imm,
			intImmediateRaw: litVal,
			hasIntImmediate: true,
			hasIntRaw:       true,
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
	if !ok {
		return bytecodeInstruction{}, false
	}
	op := bytecodeOpJumpIfIntLessEqualSlotConstFalse
	if instr.op == bytecodeOpBinaryIntCompareSlotConst {
		op = bytecodeOpJumpIfIntCompareSlotConstFalse
	} else if instr.op != bytecodeOpBinaryIntLessEqualSlotConst {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:              op,
		target:          -1,
		argCount:        instr.target,
		value:           instr.value,
		intImmediate:    instr.intImmediate,
		intImmediateRaw: instr.intImmediateRaw,
		hasIntImmediate: instr.hasIntImmediate,
		hasIntRaw:       instr.hasIntRaw,
		operator:        instr.operator,
		node:            instr.node,
	}, true
}

func bytecodeStoreSlotBinarySlotConstInstruction(ctx *bytecodeLoweringContext, targetName string, expr ast.Expression, node ast.Node) (bytecodeInstruction, bool) {
	if ctx == nil || targetName == "" {
		return bytecodeInstruction{}, false
	}
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil {
		return bytecodeInstruction{}, false
	}
	left, ok := binary.Left.(*ast.Identifier)
	if !ok || left == nil || left.Name != targetName {
		return bytecodeInstruction{}, false
	}
	instr, ok := bytecodeBinarySlotConstInstruction(ctx, binary)
	if !ok {
		return bytecodeInstruction{}, false
	}
	if ctx.slotKind(instr.target) == bytecodeCellKindI32 {
		return bytecodeInstruction{}, false
	}
	switch instr.op {
	case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst:
	default:
		return bytecodeInstruction{}, false
	}
	instr.op = bytecodeOpStoreSlotBinaryIntSlotConst
	instr.name = targetName
	instr.node = node
	return instr, true
}

func bytecodeReturnIfBinarySlotConstInstruction(ctx *bytecodeLoweringContext, condition ast.Expression, body *ast.BlockExpression) (bytecodeInstruction, bool) {
	if ctx == nil || body == nil || len(body.Body) != 1 {
		return bytecodeInstruction{}, false
	}
	ret, ok := body.Body[0].(*ast.ReturnStatement)
	if !ok || ret == nil {
		return bytecodeInstruction{}, false
	}
	instr, ok := bytecodeJumpIfFalseBinarySlotConstInstruction(ctx, condition)
	if !ok {
		return bytecodeInstruction{}, false
	}
	if returnIdent, ok := ret.Argument.(*ast.Identifier); ok && returnIdent != nil {
		returnSlot, found := ctx.lookupSlot(returnIdent.Name)
		if !found {
			return bytecodeInstruction{}, false
		}
		return bytecodeInstruction{
			op:              bytecodeOpReturnIfIntLessEqualSlotConst,
			target:          returnSlot,
			argCount:        instr.argCount,
			value:           instr.value,
			intImmediate:    instr.intImmediate,
			intImmediateRaw: instr.intImmediateRaw,
			hasIntImmediate: instr.hasIntImmediate,
			hasIntRaw:       instr.hasIntRaw,
			operator:        instr.operator,
			node:            ret,
		}, true
	}
	if lit, ok := ret.Argument.(*ast.IntegerLiteral); ok && lit != nil && lit.Value != nil && lit.IntegerType == nil && lit.Value.IsInt64() {
		litVal := lit.Value.Int64()
		if litVal >= math.MinInt32 && litVal <= math.MaxInt32 {
			return bytecodeInstruction{
				op:              bytecodeOpReturnConstIfIntLessEqualSlotConst,
				target:          -1,
				argCount:        instr.argCount,
				value:           runtime.NewSmallInt(litVal, runtime.IntegerI32),
				intImmediate:    instr.intImmediate,
				intImmediateRaw: instr.intImmediateRaw,
				hasIntImmediate: instr.hasIntImmediate,
				hasIntRaw:       instr.hasIntRaw,
				operator:        instr.operator,
				node:            ret,
			}, true
		}
	}
	return bytecodeInstruction{}, false
}
