package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeBinaryFloatMulSlotConstInstruction(ctx *bytecodeLoweringContext, expr *ast.BinaryExpression) (bytecodeInstruction, bool) {
	if ctx == nil || expr == nil || expr.Operator != "*" {
		return bytecodeInstruction{}, false
	}
	leftSlot, leftIdent := bytecodeExpressionIdentifierSlot(ctx, expr.Left)
	rightSlot, rightIdent := bytecodeExpressionIdentifierSlot(ctx, expr.Right)
	if lit, ok := expr.Right.(*ast.FloatLiteral); ok && lit != nil && leftIdent {
		if immediate, ok := bytecodeFloatLiteralImmediate(lit); ok {
			return bytecodeInstruction{
				op:                 bytecodeOpBinaryFloatMulSlotConst,
				target:             leftSlot,
				value:              immediate,
				floatImmediateRaw:  immediate.Val,
				floatImmediateKind: immediate.TypeSuffix,
				hasFloatImmediate:  true,
				node:               expr,
			}, true
		}
	}
	if lit, ok := expr.Left.(*ast.FloatLiteral); ok && lit != nil && rightIdent {
		if immediate, ok := bytecodeFloatLiteralImmediate(lit); ok {
			return bytecodeInstruction{
				op:                 bytecodeOpBinaryFloatMulSlotConst,
				target:             rightSlot,
				value:              immediate,
				floatImmediateRaw:  immediate.Val,
				floatImmediateKind: immediate.TypeSuffix,
				hasFloatImmediate:  true,
				node:               expr,
			}, true
		}
	}
	return bytecodeInstruction{}, false
}

func bytecodeExpressionIdentifierSlot(ctx *bytecodeLoweringContext, expr ast.Expression) (int, bool) {
	if ctx == nil || expr == nil {
		return 0, false
	}
	ident, ok := expr.(*ast.Identifier)
	if !ok || ident == nil {
		return 0, false
	}
	slot, found := ctx.lookupSlot(ident.Name)
	return slot, found
}

func bytecodeFloatLiteralImmediate(lit *ast.FloatLiteral) (runtime.FloatValue, bool) {
	if lit == nil {
		return runtime.FloatValue{}, false
	}
	kind := runtime.FloatF64
	if lit.FloatType != nil {
		kind = runtime.FloatType(*lit.FloatType)
	}
	if kind != runtime.FloatF32 && kind != runtime.FloatF64 {
		return runtime.FloatValue{}, false
	}
	return runtime.FloatValue{
		Val:        normalizeFloat(kind, lit.Value),
		TypeSuffix: kind,
	}, true
}
