package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeStoreSlotFloatAffinePlan struct {
	sourceSlot  int
	divisorSlot int
	divisorName string
	targetKind  runtime.FloatType
	scaleVal    float64
	scaleKind   runtime.FloatType
	offsetVal   float64
	offsetKind  runtime.FloatType
}

type bytecodeStoreSlotFloatAffineLoweringPlan struct {
	instr bytecodeInstruction
	plan  bytecodeStoreSlotFloatAffinePlan
}

func (plan bytecodeStoreSlotFloatAffinePlan) validForSlots(slotCount int) bool {
	if plan.sourceSlot < 0 || plan.sourceSlot >= slotCount {
		return false
	}
	if plan.divisorSlot >= 0 {
		return plan.divisorSlot < slotCount
	}
	return plan.divisorName != ""
}

func (ctx *bytecodeLoweringContext) setFloatAffineStorePlan(index int, plan bytecodeStoreSlotFloatAffinePlan) {
	if ctx == nil || index < 0 {
		return
	}
	if ctx.floatAffineStores == nil {
		ctx.floatAffineStores = make(map[int]bytecodeStoreSlotFloatAffinePlan, 1)
	}
	ctx.floatAffineStores[index] = plan
}

func bytecodeStoreSlotFloatAffineInstruction(ctx *bytecodeLoweringContext, expr ast.Expression, node ast.Node) (bytecodeStoreSlotFloatAffineLoweringPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil {
		return bytecodeStoreSlotFloatAffineLoweringPlan{}, false
	}
	sub, ok := expr.(*ast.BinaryExpression)
	if !ok || sub == nil || sub.Operator != "-" {
		return bytecodeStoreSlotFloatAffineLoweringPlan{}, false
	}
	offsetVal, offsetKind, ok := bytecodeFloatLiteralOperand(sub.Right)
	if !ok {
		return bytecodeStoreSlotFloatAffineLoweringPlan{}, false
	}
	div, ok := sub.Left.(*ast.BinaryExpression)
	if !ok || div == nil || div.Operator != "/" {
		return bytecodeStoreSlotFloatAffineLoweringPlan{}, false
	}
	divisorName, targetKind, ok := bytecodeFloatCastIdentifierOperand(div.Right)
	if !ok {
		return bytecodeStoreSlotFloatAffineLoweringPlan{}, false
	}
	mul, ok := div.Left.(*ast.BinaryExpression)
	if !ok || mul == nil || mul.Operator != "*" {
		return bytecodeStoreSlotFloatAffineLoweringPlan{}, false
	}
	sourceName, scaleVal, scaleKind, ok := bytecodeFloatCastSlotMulOperand(ctx, mul, targetKind)
	if !ok {
		return bytecodeStoreSlotFloatAffineLoweringPlan{}, false
	}
	sourceSlot, found := ctx.lookupSlot(sourceName)
	if !found {
		return bytecodeStoreSlotFloatAffineLoweringPlan{}, false
	}

	divisorSlot := -1
	if slot, found := ctx.lookupSlot(divisorName); found {
		divisorSlot = slot
	} else if !bytecodeSimpleLookupName(divisorName) {
		return bytecodeStoreSlotFloatAffineLoweringPlan{}, false
	}

	return bytecodeStoreSlotFloatAffineLoweringPlan{
		instr: bytecodeInstruction{
			op:   bytecodeOpStoreSlotFloatAffine,
			node: node,
		},
		plan: bytecodeStoreSlotFloatAffinePlan{
			sourceSlot:  sourceSlot,
			divisorSlot: divisorSlot,
			divisorName: divisorName,
			targetKind:  targetKind,
			scaleVal:    scaleVal,
			scaleKind:   scaleKind,
			offsetVal:   offsetVal,
			offsetKind:  offsetKind,
		},
	}, true
}

func bytecodeFloatCastSlotMulOperand(ctx *bytecodeLoweringContext, expr *ast.BinaryExpression, targetKind runtime.FloatType) (string, float64, runtime.FloatType, bool) {
	if ctx == nil || expr == nil || expr.Operator != "*" {
		return "", 0, runtime.FloatF64, false
	}
	trySide := func(castExpr ast.Expression, literalExpr ast.Expression) (string, float64, runtime.FloatType, bool) {
		name, castKind, ok := bytecodeFloatCastIdentifierOperand(castExpr)
		if !ok || castKind != targetKind {
			return "", 0, runtime.FloatF64, false
		}
		if _, found := ctx.lookupSlot(name); !found {
			return "", 0, runtime.FloatF64, false
		}
		scaleVal, scaleKind, ok := bytecodeFloatLiteralOperand(literalExpr)
		if !ok {
			return "", 0, runtime.FloatF64, false
		}
		return name, scaleVal, scaleKind, true
	}
	if name, scaleVal, scaleKind, ok := trySide(expr.Left, expr.Right); ok {
		return name, scaleVal, scaleKind, true
	}
	return trySide(expr.Right, expr.Left)
}

func bytecodeFloatCastIdentifierOperand(expr ast.Expression) (string, runtime.FloatType, bool) {
	cast, ok := expr.(*ast.TypeCastExpression)
	if !ok || cast == nil || cast.Expression == nil || cast.TargetType == nil {
		return "", runtime.FloatF64, false
	}
	targetKind, ok := bytecodeFloatCastTargetKind(cast.TargetType)
	if !ok {
		return "", runtime.FloatF64, false
	}
	ident, ok := cast.Expression.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return "", runtime.FloatF64, false
	}
	return ident.Name, targetKind, true
}

func bytecodeFloatLiteralOperand(expr ast.Expression) (float64, runtime.FloatType, bool) {
	lit, ok := expr.(*ast.FloatLiteral)
	if !ok || lit == nil {
		return 0, runtime.FloatF64, false
	}
	kind := runtime.FloatF64
	if lit.FloatType != nil {
		kind = runtime.FloatType(*lit.FloatType)
	}
	switch kind {
	case runtime.FloatF32, runtime.FloatF64:
		return normalizeFloat(kind, lit.Value), kind, true
	default:
		return 0, runtime.FloatF64, false
	}
}
