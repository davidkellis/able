package interpreter

import "able/interpreter-go/pkg/ast"

func emitGuardedImplicitSlotAssignment(ctx *bytecodeLoweringContext, i *Interpreter, n *ast.AssignmentExpression, name string, typedPattern *ast.TypedPattern, hasTypedStore bool) (bool, error) {
	if ctx == nil || ctx.frameLayout == nil || n == nil || n.Operator != ast.AssignmentAssign {
		return false, nil
	}
	slot, implicit, found := ctx.lookupAnySlot(name)
	if !found || !implicit {
		return false, nil
	}
	if err := emitExpression(ctx, i, n.Right); err != nil {
		return true, err
	}
	ctx.prepareGuardedImplicitSlotStore(slot, n, typedPattern, hasTypedStore)
	ctx.emit(ctx.guardedImplicitSlotStoreInstruction(slot, name, n, typedPattern, hasTypedStore))
	return true, nil
}

func emitGuardedImplicitSlotDeclaration(ctx *bytecodeLoweringContext, n *ast.AssignmentExpression, name string, typedPattern *ast.TypedPattern, hasTypedStore bool) bool {
	if ctx == nil || n == nil || n.Operator != ast.AssignmentAssign || !ctx.canGuardImplicitSlot(name) {
		return false
	}
	slotKind := bytecodeCellKindValue
	if hasTypedStore {
		slotKind = bytecodeCellKindForTypeExpr(typedPattern.TypeAnnotation)
	}
	slot := ctx.declareImplicitSlotWithKind(name, slotKind)
	ctx.prepareGuardedImplicitSlotStore(slot, n, typedPattern, hasTypedStore)
	ctx.emit(ctx.guardedImplicitSlotStoreInstruction(slot, name, n, typedPattern, hasTypedStore))
	return true
}

func (ctx *bytecodeLoweringContext) prepareGuardedImplicitSlotStore(slot int, n *ast.AssignmentExpression, typedPattern *ast.TypedPattern, hasTypedStore bool) {
	if hasTypedStore {
		ctx.setSlotSimpleCheck(slot, bytecodeSimpleTypeCheckForName(cachedSimpleTypeName(typedPattern.TypeAnnotation)))
		ctx.setSlotExactStructDef(slot, bytecodeNominalNamedStructDefinitionForTypeExpr(ctx, typedPattern.TypeAnnotation))
		return
	}
	ctx.setSlotSimpleCheck(slot, bytecodeExpressionSimpleTypeCheck(ctx, n.Right))
	ctx.setSlotExactStructDef(slot, bytecodeNominalNamedStructDefinitionForExpr(ctx, n.Right))
}

func (ctx *bytecodeLoweringContext) guardedImplicitSlotStoreInstruction(slot int, name string, n *ast.AssignmentExpression, typedPattern *ast.TypedPattern, hasTypedStore bool) bytecodeInstruction {
	instr := bytecodeInstruction{
		op:         bytecodeOpStoreImplicitSlot,
		target:     slot,
		name:       name,
		nameSimple: bytecodeSimpleLookupName(name),
		node:       n,
	}
	if hasTypedStore {
		instr.storeTyped = true
		instr.typeExpr = typedPattern.TypeAnnotation
		instr.discardResult = ctx.discardExpressionValue &&
			ctx.discardExpressionNode == n &&
			bytecodeCellKindForTypeExpr(typedPattern.TypeAnnotation) == bytecodeCellKindI32
	}
	return instr
}
