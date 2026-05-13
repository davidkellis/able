package interpreter

import "able/interpreter-go/pkg/ast"

type bytecodeF64NestedArrayGetPushPlan struct {
	receiverSlot int
	outerSlot    int
	rowIndexSlot int
	colIndexSlot int
}

func (ctx *bytecodeLoweringContext) setF64NestedArrayGetPushPlan(index int, plan bytecodeF64NestedArrayGetPushPlan) {
	if ctx == nil || index < 0 {
		return
	}
	if ctx.f64NestedGetPushes == nil {
		ctx.f64NestedGetPushes = make(map[int]bytecodeF64NestedArrayGetPushPlan, 1)
	}
	ctx.f64NestedGetPushes[index] = plan
}

func bytecodeEmitTryArrayPushF64NestedGet(ctx *bytecodeLoweringContext, i *Interpreter, call *ast.FunctionCall, member *ast.MemberAccessExpression, memberName string) (bool, error) {
	if ctx == nil || ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes || call == nil || member == nil || member.Safe ||
		memberName != "push" || len(call.Arguments) != 1 || len(call.TypeArguments) != 0 {
		return false, nil
	}
	plan, ok := bytecodeF64NestedArrayGetPushPlanForCall(ctx, member, call.Arguments[0])
	if !ok {
		return false, nil
	}
	fastIP := ctx.emit(bytecodeInstruction{
		op:       bytecodeOpTryArrayPushF64NestedGet,
		name:     "push",
		argCount: 1,
		target:   -1,
		node:     call,
	})
	ctx.setF64NestedArrayGetPushPlan(fastIP, plan)
	if err := emitExpression(ctx, i, member.Object); err != nil {
		return true, err
	}
	if err := emitExpression(ctx, i, call.Arguments[0]); err != nil {
		return true, err
	}
	ctx.emit(bytecodeInstruction{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1, node: call})
	ctx.patchJump(fastIP, len(ctx.instructions))
	return true, nil
}

func bytecodeF64NestedArrayGetPushPlanForCall(ctx *bytecodeLoweringContext, member *ast.MemberAccessExpression, arg ast.Expression) (bytecodeF64NestedArrayGetPushPlan, bool) {
	receiver := bytecodeIdentifierExpressionName(member.Object)
	if receiver == "" {
		return bytecodeF64NestedArrayGetPushPlan{}, false
	}
	receiverSlot, found := ctx.lookupSlot(receiver)
	if !found {
		return bytecodeF64NestedArrayGetPushPlan{}, false
	}
	outerName, rowName, colName, ok := bytecodeNestedArrayGetPropagationNames(arg)
	if !ok {
		return bytecodeF64NestedArrayGetPushPlan{}, false
	}
	outerSlot, found := ctx.lookupSlot(outerName)
	if !found {
		return bytecodeF64NestedArrayGetPushPlan{}, false
	}
	rowSlot, found := ctx.lookupSlot(rowName)
	if !found {
		return bytecodeF64NestedArrayGetPushPlan{}, false
	}
	colSlot, found := ctx.lookupSlot(colName)
	if !found {
		return bytecodeF64NestedArrayGetPushPlan{}, false
	}
	return bytecodeF64NestedArrayGetPushPlan{
		receiverSlot: receiverSlot,
		outerSlot:    outerSlot,
		rowIndexSlot: rowSlot,
		colIndexSlot: colSlot,
	}, true
}

func bytecodeNestedArrayGetPropagationNames(expr ast.Expression) (string, string, string, bool) {
	innerProp, ok := expr.(*ast.PropagationExpression)
	if !ok || innerProp == nil {
		return "", "", "", false
	}
	innerCall, ok := innerProp.Expression.(*ast.FunctionCall)
	if !ok || innerCall == nil || len(innerCall.Arguments) != 1 || len(innerCall.TypeArguments) != 0 {
		return "", "", "", false
	}
	innerMember, ok := innerCall.Callee.(*ast.MemberAccessExpression)
	if !ok || innerMember == nil || innerMember.Safe || bytecodeIdentifierMemberName(innerMember.Member) != "get" {
		return "", "", "", false
	}
	colName := bytecodeIdentifierExpressionName(innerCall.Arguments[0])
	if colName == "" {
		return "", "", "", false
	}
	outerProp, ok := innerMember.Object.(*ast.PropagationExpression)
	if !ok || outerProp == nil {
		return "", "", "", false
	}
	outerCall, ok := outerProp.Expression.(*ast.FunctionCall)
	if !ok || outerCall == nil || len(outerCall.Arguments) != 1 || len(outerCall.TypeArguments) != 0 {
		return "", "", "", false
	}
	outerMember, ok := outerCall.Callee.(*ast.MemberAccessExpression)
	if !ok || outerMember == nil || outerMember.Safe || bytecodeIdentifierMemberName(outerMember.Member) != "get" {
		return "", "", "", false
	}
	outerName := bytecodeIdentifierExpressionName(outerMember.Object)
	rowName := bytecodeIdentifierExpressionName(outerCall.Arguments[0])
	if outerName == "" || rowName == "" {
		return "", "", "", false
	}
	return outerName, rowName, colName, true
}

func (plan bytecodeF64NestedArrayGetPushPlan) validForSlots(slotCount int) bool {
	return plan.receiverSlot >= 0 && plan.receiverSlot < slotCount &&
		plan.outerSlot >= 0 && plan.outerSlot < slotCount &&
		plan.rowIndexSlot >= 0 && plan.rowIndexSlot < slotCount &&
		plan.colIndexSlot >= 0 && plan.colIndexSlot < slotCount
}
