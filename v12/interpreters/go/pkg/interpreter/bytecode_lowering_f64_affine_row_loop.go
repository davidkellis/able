package interpreter

import "able/interpreter-go/pkg/ast"

type bytecodeF64AffineRowLoopPlan struct {
	indexSlot     int
	boundSlot     int
	receiverSlot  int
	scaleSlot     int
	leftSlot      int
	successTarget int
	resultPushIP  int
}

func (ctx *bytecodeLoweringContext) setF64AffineRowLoopPlan(index int, plan bytecodeF64AffineRowLoopPlan) {
	if ctx == nil || index < 0 {
		return
	}
	if ctx.f64AffineRowLoops == nil {
		ctx.f64AffineRowLoops = make(map[int]bytecodeF64AffineRowLoopPlan, 1)
	}
	ctx.f64AffineRowLoops[index] = plan
}

func bytecodeF64AffineRowLoopPlanForLoop(ctx *bytecodeLoweringContext, loop *ast.LoopExpression) (bytecodeF64AffineRowLoopPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes || loop == nil || loop.Body == nil || len(loop.Body.Body) != 3 {
		return bytecodeF64AffineRowLoopPlan{}, false
	}
	indexName, boundName, ok := bytecodeF64DotLoopBreakCondition(loop.Body.Body[0])
	if !ok {
		return bytecodeF64AffineRowLoopPlan{}, false
	}
	call, ok := loop.Body.Body[1].(*ast.FunctionCall)
	if !ok || call == nil {
		return bytecodeF64AffineRowLoopPlan{}, false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe || bytecodeIdentifierMemberName(member.Member) != "push" || len(call.Arguments) != 1 || len(call.TypeArguments) != 0 {
		return bytecodeF64AffineRowLoopPlan{}, false
	}
	pushPlan, ok := bytecodeF64AffineProductPushPlanForCall(ctx, member, call.Arguments[0])
	if !ok {
		return bytecodeF64AffineRowLoopPlan{}, false
	}
	indexSlot, found := ctx.lookupSlot(indexName)
	if !found || pushPlan.rightSlot != indexSlot || pushPlan.leftSlot == indexSlot {
		return bytecodeF64AffineRowLoopPlan{}, false
	}
	if !bytecodeF64DotLoopIncrement(loop.Body.Body[2], indexName) {
		return bytecodeF64AffineRowLoopPlan{}, false
	}
	boundSlot, found := ctx.lookupSlot(boundName)
	if !found {
		return bytecodeF64AffineRowLoopPlan{}, false
	}
	return bytecodeF64AffineRowLoopPlan{
		indexSlot:    indexSlot,
		boundSlot:    boundSlot,
		receiverSlot: pushPlan.receiverSlot,
		scaleSlot:    pushPlan.scaleSlot,
		leftSlot:     pushPlan.leftSlot,
	}, true
}

func (plan bytecodeF64AffineRowLoopPlan) validForSlots(slotCount int) bool {
	return plan.indexSlot >= 0 && plan.indexSlot < slotCount &&
		plan.boundSlot >= 0 && plan.boundSlot < slotCount &&
		plan.receiverSlot >= 0 && plan.receiverSlot < slotCount &&
		plan.scaleSlot >= 0 && plan.scaleSlot < slotCount &&
		plan.leftSlot >= 0 && plan.leftSlot < slotCount &&
		plan.resultPushIP >= 0 &&
		plan.successTarget > 0
}
