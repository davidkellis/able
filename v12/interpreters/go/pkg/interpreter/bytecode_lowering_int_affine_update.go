package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeStoreSlotIntMulConstAddLoweringPlan struct {
	instr      bytecodeInstruction
	targetSlot int
	addend     ast.Expression
	loadBase   bool
}

func bytecodeStoreSlotIntMulConstAddPlan(ctx *bytecodeLoweringContext, targetName string, expr ast.Expression, node ast.Node) (bytecodeStoreSlotIntMulConstAddLoweringPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil || targetName == "" {
		return bytecodeStoreSlotIntMulConstAddLoweringPlan{}, false
	}
	add, ok := expr.(*ast.BinaryExpression)
	if !ok || add == nil || add.Operator != "+" || add.Left == nil || add.Right == nil {
		return bytecodeStoreSlotIntMulConstAddLoweringPlan{}, false
	}
	mul, ok := add.Left.(*ast.BinaryExpression)
	if !ok || mul == nil || mul.Operator != "*" || mul.Left == nil || mul.Right == nil {
		return bytecodeStoreSlotIntMulConstAddLoweringPlan{}, false
	}
	left, ok := mul.Left.(*ast.Identifier)
	if !ok || left == nil || left.Name != targetName {
		return bytecodeStoreSlotIntMulConstAddLoweringPlan{}, false
	}
	lit, ok := mul.Right.(*ast.IntegerLiteral)
	if !ok || lit == nil {
		return bytecodeStoreSlotIntMulConstAddLoweringPlan{}, false
	}
	imm, litVal, ok := bytecodeSlotConstIntegerLiteralImmediate(lit)
	if !ok {
		return bytecodeStoreSlotIntMulConstAddLoweringPlan{}, false
	}
	slot, found := ctx.lookupSlot(targetName)
	if !found {
		return bytecodeStoreSlotIntMulConstAddLoweringPlan{}, false
	}
	if ctx.slotKind(slot) == bytecodeCellKindI32 {
		return bytecodeStoreSlotIntMulConstAddLoweringPlan{}, false
	}
	op := bytecodeOpStoreSlotIntMulConstAdd
	loadBase := true
	if bytecodeIntAffineAddendCanRunBeforeBaseRead(add.Right) {
		op = bytecodeOpStoreSlotIntMulConstAddFromSlot
		loadBase = false
	}
	return bytecodeStoreSlotIntMulConstAddLoweringPlan{
		instr: bytecodeInstruction{
			op:              op,
			target:          slot,
			name:            targetName,
			operator:        "+",
			value:           imm,
			intImmediate:    imm,
			intImmediateRaw: litVal,
			hasIntImmediate: true,
			hasIntRaw:       true,
			node:            node,
		},
		targetSlot: slot,
		addend:     add.Right,
		loadBase:   loadBase,
	}, true
}

func bytecodeIntAffineAddendCanRunBeforeBaseRead(expr ast.Expression) bool {
	switch n := expr.(type) {
	case *ast.Identifier, *ast.IntegerLiteral:
		return true
	case *ast.UnaryExpression:
		return n != nil && n.Operator == ast.UnaryOperatorNegate && bytecodeIntAffineAddendCanRunBeforeBaseRead(n.Operand)
	case *ast.BinaryExpression:
		if n == nil {
			return false
		}
		switch n.Operator {
		case "+", "-", "*", "/", "%":
			return bytecodeIntAffineAddendCanRunBeforeBaseRead(n.Left) && bytecodeIntAffineAddendCanRunBeforeBaseRead(n.Right)
		default:
			return false
		}
	case *ast.TypeCastExpression:
		if n == nil {
			return false
		}
		if _, ok := lookupIntegerInfo(runtime.IntegerType(typeExpressionToString(n.TargetType))); !ok {
			return false
		}
		return bytecodeIntAffineAddendCanRunBeforeBaseRead(n.Expression)
	default:
		return false
	}
}
