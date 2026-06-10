package interpreter

import "able/interpreter-go/pkg/ast"

type bytecodeStoreSlotFloatAddSubLoweringPlan struct {
	instr        bytecodeInstruction
	baseSlot     int
	baseName     string
	subLeftSlot  int
	subLeftName  string
	subRightSlot int
	subRightName string
}

func bytecodeStoreSlotFloatAddSubPlan(ctx *bytecodeLoweringContext, targetName string, expr ast.Expression, node ast.Node) (bytecodeStoreSlotFloatAddSubLoweringPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil || targetName == "" {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}
	add, ok := expr.(*ast.BinaryExpression)
	if !ok || add == nil || add.Operator != "+" {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}

	var (
		baseIdent *ast.Identifier
		sub       *ast.BinaryExpression
	)
	if candidate, ok := add.Left.(*ast.Identifier); ok && candidate != nil {
		if subCandidate, ok := add.Right.(*ast.BinaryExpression); ok && subCandidate != nil && subCandidate.Operator == "-" {
			baseIdent = candidate
			sub = subCandidate
		}
	}
	if baseIdent == nil {
		candidate, ok := add.Right.(*ast.Identifier)
		if !ok || candidate == nil {
			return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
		}
		subCandidate, ok := add.Left.(*ast.BinaryExpression)
		if !ok || subCandidate == nil || subCandidate.Operator != "-" {
			return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
		}
		baseIdent = candidate
		sub = subCandidate
	}

	subLeft, ok := sub.Left.(*ast.Identifier)
	if !ok || subLeft == nil {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}
	subRight, ok := sub.Right.(*ast.Identifier)
	if !ok || subRight == nil {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}
	slot, found := ctx.lookupSlot(targetName)
	if !found {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}
	if ctx.slotKind(slot) != bytecodeCellKindValue || !bytecodeExpressionIsKnownFloat(ctx, expr) {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}
	baseSlot, found := ctx.lookupSlot(baseIdent.Name)
	if !found {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}
	if ctx.slotKind(baseSlot) != bytecodeCellKindValue {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}
	subLeftSlot, found := ctx.lookupSlot(subLeft.Name)
	if !found {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}
	if ctx.slotKind(subLeftSlot) != bytecodeCellKindValue {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}
	subRightSlot, found := ctx.lookupSlot(subRight.Name)
	if !found {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}
	if ctx.slotKind(subRightSlot) != bytecodeCellKindValue {
		return bytecodeStoreSlotFloatAddSubLoweringPlan{}, false
	}

	return bytecodeStoreSlotFloatAddSubLoweringPlan{
		instr: bytecodeInstruction{
			op:       bytecodeOpStoreSlotFloatAddSub,
			target:   slot,
			name:     targetName,
			operator: "+",
			node:     node,
		},
		baseSlot:     baseSlot,
		baseName:     baseIdent.Name,
		subLeftSlot:  subLeftSlot,
		subLeftName:  subLeft.Name,
		subRightSlot: subRightSlot,
		subRightName: subRight.Name,
	}, true
}
