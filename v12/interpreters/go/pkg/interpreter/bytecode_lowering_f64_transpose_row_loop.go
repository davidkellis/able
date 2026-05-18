package interpreter

import "able/interpreter-go/pkg/ast"

type bytecodeF64TransposeRowLoopPlan struct {
	indexSlot     int
	boundSlot     int
	receiverSlot  int
	outerSlot     int
	colIndexSlot  int
	successTarget int
	resultPushIP  int
}

func (ctx *bytecodeLoweringContext) setF64TransposeRowLoopPlan(index int, plan bytecodeF64TransposeRowLoopPlan) {
	if ctx == nil || index < 0 {
		return
	}
	if ctx.f64TransposeRowLoops == nil {
		ctx.f64TransposeRowLoops = make(map[int]bytecodeF64TransposeRowLoopPlan, 1)
	}
	ctx.f64TransposeRowLoops[index] = plan
}

func bytecodeF64TransposeRowLoopPlanForLoop(ctx *bytecodeLoweringContext, loop *ast.LoopExpression) (bytecodeF64TransposeRowLoopPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes || loop == nil || loop.Body == nil || len(loop.Body.Body) != 3 {
		return bytecodeF64TransposeRowLoopPlan{}, false
	}
	indexName, boundName, ok := bytecodeF64DotLoopBreakCondition(loop.Body.Body[0])
	if !ok {
		return bytecodeF64TransposeRowLoopPlan{}, false
	}
	call, ok := loop.Body.Body[1].(*ast.FunctionCall)
	if !ok || call == nil {
		return bytecodeF64TransposeRowLoopPlan{}, false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe || bytecodeIdentifierMemberName(member.Member) != "push" || len(call.Arguments) != 1 || len(call.TypeArguments) != 0 {
		return bytecodeF64TransposeRowLoopPlan{}, false
	}
	pushPlan, ok := bytecodeF64NestedArrayGetPushPlanForCall(ctx, member, call.Arguments[0])
	if !ok {
		return bytecodeF64TransposeRowLoopPlan{}, false
	}
	indexSlot, found := ctx.lookupSlot(indexName)
	if !found || pushPlan.rowIndexSlot != indexSlot || pushPlan.colIndexSlot == indexSlot {
		return bytecodeF64TransposeRowLoopPlan{}, false
	}
	if !bytecodeF64DotLoopIncrement(loop.Body.Body[2], indexName) {
		return bytecodeF64TransposeRowLoopPlan{}, false
	}
	boundSlot, found := ctx.lookupSlot(boundName)
	if !found {
		return bytecodeF64TransposeRowLoopPlan{}, false
	}
	return bytecodeF64TransposeRowLoopPlan{
		indexSlot:    indexSlot,
		boundSlot:    boundSlot,
		receiverSlot: pushPlan.receiverSlot,
		outerSlot:    pushPlan.outerSlot,
		colIndexSlot: pushPlan.colIndexSlot,
	}, true
}

func (plan bytecodeF64TransposeRowLoopPlan) validForSlots(slotCount int) bool {
	return plan.indexSlot >= 0 && plan.indexSlot < slotCount &&
		plan.boundSlot >= 0 && plan.boundSlot < slotCount &&
		plan.receiverSlot >= 0 && plan.receiverSlot < slotCount &&
		plan.outerSlot >= 0 && plan.outerSlot < slotCount &&
		plan.colIndexSlot >= 0 && plan.colIndexSlot < slotCount &&
		plan.resultPushIP >= 0 &&
		plan.successTarget > 0
}
