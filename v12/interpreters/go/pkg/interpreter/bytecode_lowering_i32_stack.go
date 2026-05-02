package interpreter

import (
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeEmitFinalI32StackExpr(ctx *bytecodeLoweringContext, expr ast.Expression) (bool, error) {
	if ctx == nil || !bytecodeCanEmitRawI32StackExprWithSlots(ctx, expr) {
		return false, nil
	}
	bytecodeEmitRawI32StackExpr(ctx, expr)
	ctx.emit(bytecodeInstruction{op: bytecodeOpBoxI32, node: expr})
	return true, nil
}

func bytecodeCanEmitRawI32StackExpr(expr ast.Expression) bool {
	return bytecodeCanEmitRawI32StackExprWithSlots(nil, expr)
}

func bytecodeCanEmitRawI32StackExprWithSlots(ctx *bytecodeLoweringContext, expr ast.Expression) bool {
	switch n := expr.(type) {
	case *ast.IntegerLiteral:
		_, ok := bytecodeI32LiteralRaw(n)
		return ok
	case *ast.Identifier:
		if ctx == nil || n == nil {
			return false
		}
		slot, ok := ctx.lookupSlot(n.Name)
		return ok && ctx.slotKind(slot) == bytecodeCellKindI32
	case *ast.BinaryExpression:
		if n == nil || (n.Operator != "+" && n.Operator != "-") {
			return false
		}
		return bytecodeCanEmitRawI32StackExprWithSlots(ctx, n.Left) && bytecodeCanEmitRawI32StackExprWithSlots(ctx, n.Right)
	default:
		return false
	}
}

func bytecodeEmitRawI32StackExpr(ctx *bytecodeLoweringContext, expr ast.Expression) {
	switch n := expr.(type) {
	case *ast.IntegerLiteral:
		raw, _ := bytecodeI32LiteralRaw(n)
		imm := runtime.NewSmallInt(raw, runtime.IntegerI32)
		ctx.emit(bytecodeInstruction{
			op:              bytecodeOpConstI32,
			value:           imm,
			intImmediate:    imm,
			intImmediateRaw: raw,
			hasIntImmediate: true,
			hasIntRaw:       true,
			node:            n,
		})
	case *ast.Identifier:
		if slot, ok := ctx.lookupSlot(n.Name); ok {
			ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlotI32, target: slot, name: n.Name, node: n})
		}
	case *ast.BinaryExpression:
		bytecodeEmitRawI32StackExpr(ctx, n.Left)
		bytecodeEmitRawI32StackExpr(ctx, n.Right)
		op := bytecodeOpBinaryI32Add
		if n.Operator == "-" {
			op = bytecodeOpBinaryI32Sub
		}
		ctx.emit(bytecodeInstruction{op: op, operator: n.Operator, node: n})
	}
}

func bytecodeI32LiteralRaw(lit *ast.IntegerLiteral) (int64, bool) {
	if lit == nil || lit.Value == nil || !lit.Value.IsInt64() {
		return 0, false
	}
	if lit.IntegerType != nil && runtime.IntegerType(*lit.IntegerType) != runtime.IntegerI32 {
		return 0, false
	}
	raw := lit.Value.Int64()
	if raw < math.MinInt32 || raw > math.MaxInt32 {
		return 0, false
	}
	return raw, true
}
