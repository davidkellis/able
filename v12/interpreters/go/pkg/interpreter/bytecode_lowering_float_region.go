package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

const (
	bytecodeFloatRegionMinOperations = 2
	bytecodeFloatRegionMaxDepth      = 16
)

type bytecodeFloatRegionStepKind uint8

const (
	bytecodeFloatRegionLoadSlot bytecodeFloatRegionStepKind = iota
	bytecodeFloatRegionConst
	bytecodeFloatRegionAdd
	bytecodeFloatRegionSub
	bytecodeFloatRegionMul
	bytecodeFloatRegionDiv
)

type bytecodeFloatRegionStep struct {
	kind      bytecodeFloatRegionStepKind
	slot      int
	value     float64
	floatKind runtime.FloatType
}

type bytecodeFloatRegionPlan struct {
	steps    []bytecodeFloatRegionStep
	maxDepth uint8
}

func bytecodeFloatRegionPlanForExpression(ctx *bytecodeLoweringContext, expr ast.Expression) (bytecodeFloatRegionPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil || expr == nil {
		return bytecodeFloatRegionPlan{}, false
	}
	if _, ok := bytecodeExpressionSimpleTypeCheck(ctx, expr).floatType(); !ok {
		return bytecodeFloatRegionPlan{}, false
	}

	steps := make([]bytecodeFloatRegionStep, 0, 8)
	operations := 0
	var appendExpression func(ast.Expression) bool
	appendExpression = func(current ast.Expression) bool {
		switch node := current.(type) {
		case *ast.Identifier:
			if node == nil || node.Name == "" {
				return false
			}
			slot, found := ctx.lookupSlot(node.Name)
			if !found || ctx.slotKind(slot) != bytecodeCellKindValue {
				return false
			}
			if _, ok := ctx.slotSimpleCheck(slot).floatType(); !ok {
				return false
			}
			steps = append(steps, bytecodeFloatRegionStep{kind: bytecodeFloatRegionLoadSlot, slot: slot})
			return true
		case *ast.FloatLiteral:
			value, kind, ok := bytecodeFloatLiteralOperand(node)
			if !ok {
				return false
			}
			steps = append(steps, bytecodeFloatRegionStep{kind: bytecodeFloatRegionConst, value: value, floatKind: kind})
			return true
		case *ast.BinaryExpression:
			if node == nil || node.Left == nil || node.Right == nil {
				return false
			}
			var kind bytecodeFloatRegionStepKind
			switch node.Operator {
			case "+":
				kind = bytecodeFloatRegionAdd
			case "-":
				kind = bytecodeFloatRegionSub
			case "*":
				kind = bytecodeFloatRegionMul
			case "/":
				kind = bytecodeFloatRegionDiv
			default:
				return false
			}
			if _, ok := bytecodeExpressionSimpleTypeCheck(ctx, node).floatType(); !ok {
				return false
			}
			if !appendExpression(node.Left) || !appendExpression(node.Right) {
				return false
			}
			steps = append(steps, bytecodeFloatRegionStep{kind: kind})
			operations++
			return true
		default:
			return false
		}
	}
	if !appendExpression(expr) || operations < bytecodeFloatRegionMinOperations {
		return bytecodeFloatRegionPlan{}, false
	}

	depth := 0
	maxDepth := 0
	for _, step := range steps {
		switch step.kind {
		case bytecodeFloatRegionLoadSlot, bytecodeFloatRegionConst:
			depth++
		default:
			if depth < 2 {
				return bytecodeFloatRegionPlan{}, false
			}
			depth--
		}
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	if depth != 1 || maxDepth > bytecodeFloatRegionMaxDepth {
		return bytecodeFloatRegionPlan{}, false
	}
	return bytecodeFloatRegionPlan{steps: steps, maxDepth: uint8(maxDepth)}, true
}

func (ctx *bytecodeLoweringContext) emitStoreSlotFloatRegion(plan bytecodeFloatRegionPlan, target int, name string, node ast.Node, discardResult bool) {
	if ctx == nil || target < 0 || len(plan.steps) == 0 {
		return
	}
	planIndex := len(ctx.floatRegions)
	ctx.floatRegions = append(ctx.floatRegions, plan)
	ctx.emit(bytecodeInstruction{
		op:            bytecodeOpStoreSlotFloatRegion,
		target:        target,
		name:          name,
		argCount:      planIndex,
		node:          node,
		discardResult: discardResult,
	})
}

func bytecodeTryEmitDeclaredFloatRegion(ctx *bytecodeLoweringContext, name string, assignment *ast.AssignmentExpression, check bytecodeSimpleTypeCheck) bool {
	if assignment == nil || assignment.Operator != ast.AssignmentDeclare {
		return false
	}
	plan, ok := bytecodeFloatRegionPlanForExpression(ctx, assignment.Right)
	if !ok {
		return false
	}
	slot := ctx.declareSlotWithKind(name, bytecodeCellKindValue)
	ctx.setSlotSimpleCheck(slot, check)
	ctx.emitStoreSlotFloatRegion(plan, slot, name, assignment, ctx.discardExpressionValue && ctx.discardExpressionNode == assignment)
	return true
}

func bytecodeTryEmitAssignedFloatRegion(ctx *bytecodeLoweringContext, name string, assignment *ast.AssignmentExpression, check bytecodeSimpleTypeCheck) bool {
	if assignment == nil || assignment.Operator != ast.AssignmentAssign {
		return false
	}
	plan, ok := bytecodeFloatRegionPlanForExpression(ctx, assignment.Right)
	if !ok {
		return false
	}
	slot, found := ctx.lookupSlot(name)
	if !found {
		return false
	}
	ctx.setSlotSimpleCheck(slot, check)
	ctx.emitStoreSlotFloatRegion(plan, slot, name, assignment, ctx.discardExpressionValue && ctx.discardExpressionNode == assignment)
	return true
}
