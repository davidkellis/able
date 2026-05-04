package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

type bytecodeSlotMatchPatternKind uint8

const (
	bytecodeSlotMatchPatternWildcard bytecodeSlotMatchPatternKind = iota
	bytecodeSlotMatchPatternNil
	bytecodeSlotMatchPatternTyped
)

type bytecodeSlotMatchClausePlan struct {
	kind        bytecodeSlotMatchPatternKind
	typeExpr    ast.TypeExpression
	bindingName string
	slotKind    bytecodeCellKind
}

func emitSlotMatch(ctx *bytecodeLoweringContext, i *Interpreter, expr *ast.MatchExpression) (bool, error) {
	if ctx == nil || ctx.frameLayout == nil || expr == nil || !bytecodeCanLowerSlotMatch(expr) {
		return false, nil
	}
	if err := emitExpression(ctx, i, expr.Subject); err != nil {
		return false, err
	}
	subjectName := fmt.Sprintf("$match_subject_%d", ctx.nextSlot)
	subjectSlot := ctx.declareSlotWithKind(subjectName, bytecodeCellKindValue)
	ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: subjectSlot, name: subjectName, node: expr})
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})

	endJumps := make([]int, 0, len(expr.Clauses))
	for _, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		plan, ok := bytecodeSlotMatchClausePlanForPattern(clause.Pattern)
		if !ok {
			return false, nil
		}
		nextJump := emitSlotMatchPatternTest(ctx, subjectSlot, plan, clause)
		ctx.enterScope()
		emitSlotMatchPatternBinding(ctx, plan, clause)
		if err := emitExpression(ctx, i, clause.Body); err != nil {
			return false, err
		}
		ctx.exitScope()
		endJumps = append(endJumps, ctx.emit(bytecodeInstruction{op: bytecodeOpJump, target: -1}))
		if nextJump >= 0 {
			ctx.patchJump(nextJump, len(ctx.instructions))
		}
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpMatchNoClause, node: expr})
	end := len(ctx.instructions)
	for _, jump := range endJumps {
		ctx.patchJump(jump, end)
	}
	return true, nil
}

func bytecodeCanLowerSlotMatch(expr *ast.MatchExpression) bool {
	if expr == nil {
		return false
	}
	if len(expr.Clauses) == 0 {
		return false
	}
	for _, clause := range expr.Clauses {
		if clause == nil || clause.Guard != nil {
			return false
		}
		if _, ok := bytecodeSlotMatchClausePlanForPattern(clause.Pattern); !ok {
			return false
		}
	}
	return true
}

func bytecodeSlotMatchClausePlanForPattern(pattern ast.Pattern) (bytecodeSlotMatchClausePlan, bool) {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		if p == nil {
			return bytecodeSlotMatchClausePlan{}, false
		}
		return bytecodeSlotMatchClausePlan{kind: bytecodeSlotMatchPatternWildcard}, true
	case *ast.LiteralPattern:
		if p == nil {
			return bytecodeSlotMatchClausePlan{}, false
		}
		if _, ok := p.Literal.(*ast.NilLiteral); ok {
			return bytecodeSlotMatchClausePlan{kind: bytecodeSlotMatchPatternNil}, true
		}
	case *ast.TypedPattern:
		if p == nil || p.TypeAnnotation == nil {
			return bytecodeSlotMatchClausePlan{}, false
		}
		plan := bytecodeSlotMatchClausePlan{
			kind:     bytecodeSlotMatchPatternTyped,
			typeExpr: p.TypeAnnotation,
			slotKind: bytecodeCellKindForTypeExpr(p.TypeAnnotation),
		}
		switch inner := p.Pattern.(type) {
		case *ast.Identifier:
			if inner != nil && inner.Name != "" && inner.Name != "_" {
				plan.bindingName = inner.Name
			}
			return plan, true
		case *ast.WildcardPattern:
			return plan, inner != nil
		}
	}
	return bytecodeSlotMatchClausePlan{}, false
}

func emitSlotMatchPatternTest(ctx *bytecodeLoweringContext, subjectSlot int, plan bytecodeSlotMatchClausePlan, clause *ast.MatchClause) int {
	switch plan.kind {
	case bytecodeSlotMatchPatternWildcard:
		return -1
	case bytecodeSlotMatchPatternNil:
		ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: subjectSlot, node: clause})
		return ctx.emit(bytecodeInstruction{op: bytecodeOpJumpIfNotNil, target: -1, node: clause})
	case bytecodeSlotMatchPatternTyped:
		ctx.emit(bytecodeInstruction{op: bytecodeOpLoadSlot, target: subjectSlot, node: clause})
		return ctx.emit(bytecodeInstruction{
			op:       bytecodeOpJumpIfNotTypedPattern,
			target:   -1,
			typeExpr: plan.typeExpr,
			node:     clause,
		})
	default:
		return -1
	}
}

func emitSlotMatchPatternBinding(ctx *bytecodeLoweringContext, plan bytecodeSlotMatchClausePlan, clause *ast.MatchClause) {
	if plan.kind != bytecodeSlotMatchPatternTyped {
		return
	}
	if plan.bindingName == "" {
		ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
		return
	}
	slot := ctx.declareSlotWithKind(plan.bindingName, plan.slotKind)
	ctx.emit(bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: slot, name: plan.bindingName, node: clause})
	ctx.emit(bytecodeInstruction{op: bytecodeOpPop})
}
