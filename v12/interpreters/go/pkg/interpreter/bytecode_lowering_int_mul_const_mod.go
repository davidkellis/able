package interpreter

import (
	"able/interpreter-go/pkg/ast"
)

func bytecodeStoreSlotIntMulConstModConstInstruction(ctx *bytecodeLoweringContext, targetName string, expr ast.Expression, node ast.Node) (bytecodeInstruction, bool) {
	if ctx == nil || ctx.frameLayout == nil || targetName == "" {
		return bytecodeInstruction{}, false
	}
	mod, ok := expr.(*ast.BinaryExpression)
	if !ok || mod == nil || mod.Operator != "%" || mod.Left == nil || mod.Right == nil {
		return bytecodeInstruction{}, false
	}
	mul, ok := mod.Left.(*ast.BinaryExpression)
	if !ok || mul == nil || mul.Operator != "*" || mul.Left == nil || mul.Right == nil {
		return bytecodeInstruction{}, false
	}
	left, ok := mul.Left.(*ast.Identifier)
	if !ok || left == nil || left.Name != targetName {
		return bytecodeInstruction{}, false
	}
	mulLit, ok := mul.Right.(*ast.IntegerLiteral)
	if !ok || mulLit == nil {
		return bytecodeInstruction{}, false
	}
	modLit, ok := mod.Right.(*ast.IntegerLiteral)
	if !ok || modLit == nil {
		return bytecodeInstruction{}, false
	}
	mulImm, mulRaw, ok := bytecodeSlotConstIntegerLiteralImmediate(mulLit)
	if !ok {
		return bytecodeInstruction{}, false
	}
	modImm, _, ok := bytecodeSlotConstIntegerLiteralImmediate(modLit)
	modRaw, okRaw := modImm.ToInt64()
	if !ok || modImm.TypeSuffix != mulImm.TypeSuffix {
		return bytecodeInstruction{}, false
	}
	if !okRaw {
		return bytecodeInstruction{}, false
	}
	slot, found := ctx.lookupSlot(targetName)
	if !found || ctx.slotKind(slot) == bytecodeCellKindI32 {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:               bytecodeOpStoreSlotIntMulConstModConst,
		target:           slot,
		name:             targetName,
		operator:         "%",
		value:            modImm,
		intImmediate:     mulImm,
		intImmediate2:    modImm,
		intImmediateRaw:  mulRaw,
		intImmediate2Raw: modRaw,
		hasIntImmediate:  true,
		hasIntImmediate2: true,
		hasIntRaw:        true,
		hasIntRaw2:       true,
		node:             node,
	}, true
}
