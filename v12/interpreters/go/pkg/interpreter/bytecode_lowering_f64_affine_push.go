package interpreter

import "able/interpreter-go/pkg/ast"

type bytecodeF64AffineProductPushPlan struct {
	receiverSlot int
	scaleSlot    int
	leftSlot     int
	rightSlot    int
}

func (ctx *bytecodeLoweringContext) setF64AffineProductPushPlan(index int, plan bytecodeF64AffineProductPushPlan) {
	if ctx == nil || index < 0 {
		return
	}
	if ctx.f64AffinePushes == nil {
		ctx.f64AffinePushes = make(map[int]bytecodeF64AffineProductPushPlan, 1)
	}
	ctx.f64AffinePushes[index] = plan
}

func bytecodeEmitTryArrayPushF64AffineProduct(ctx *bytecodeLoweringContext, i *Interpreter, call *ast.FunctionCall, member *ast.MemberAccessExpression, memberName string) (bool, error) {
	if ctx == nil || ctx.frameLayout == nil || ctx.frameLayout.needsEnvScopes || call == nil || member == nil || member.Safe ||
		memberName != "push" || len(call.Arguments) != 1 || len(call.TypeArguments) != 0 {
		return false, nil
	}
	plan, ok := bytecodeF64AffineProductPushPlanForCall(ctx, member, call.Arguments[0])
	if !ok {
		return false, nil
	}
	fastIP := ctx.emit(bytecodeInstruction{
		op:       bytecodeOpTryArrayPushF64AffineProduct,
		name:     "push",
		argCount: 1,
		target:   -1,
		node:     call,
	})
	ctx.setF64AffineProductPushPlan(fastIP, plan)
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

func bytecodeF64AffineProductPushPlanForCall(ctx *bytecodeLoweringContext, member *ast.MemberAccessExpression, arg ast.Expression) (bytecodeF64AffineProductPushPlan, bool) {
	receiver, ok := member.Object.(*ast.Identifier)
	if !ok || receiver == nil || receiver.Name == "" {
		return bytecodeF64AffineProductPushPlan{}, false
	}
	receiverSlot, found := ctx.lookupSlot(receiver.Name)
	if !found {
		return bytecodeF64AffineProductPushPlan{}, false
	}
	factors := bytecodeFlattenProductFactors(arg, nil)
	if len(factors) != 3 {
		return bytecodeF64AffineProductPushPlan{}, false
	}
	scaleName := ""
	leftName := ""
	rightName := ""
	haveDiff := false
	haveSum := false
	for _, factor := range factors {
		if name := bytecodeIdentifierExpressionName(factor); name != "" {
			if scaleName != "" {
				return bytecodeF64AffineProductPushPlan{}, false
			}
			scaleName = name
			continue
		}
		lhs, rhs, ok := bytecodeF64CastBinaryIdentifiers(factor, "-")
		if ok {
			if haveDiff {
				return bytecodeF64AffineProductPushPlan{}, false
			}
			if leftName != "" && (lhs != leftName || rhs != rightName) {
				return bytecodeF64AffineProductPushPlan{}, false
			}
			leftName, rightName = lhs, rhs
			haveDiff = true
			continue
		}
		lhs, rhs, ok = bytecodeF64CastBinaryIdentifiers(factor, "+")
		if ok {
			if haveSum {
				return bytecodeF64AffineProductPushPlan{}, false
			}
			if leftName != "" && (lhs != leftName || rhs != rightName) {
				return bytecodeF64AffineProductPushPlan{}, false
			}
			leftName, rightName = lhs, rhs
			haveSum = true
			continue
		}
		return bytecodeF64AffineProductPushPlan{}, false
	}
	if scaleName == "" || leftName == "" || rightName == "" || !haveDiff || !haveSum {
		return bytecodeF64AffineProductPushPlan{}, false
	}
	scaleSlot, found := ctx.lookupSlot(scaleName)
	if !found {
		return bytecodeF64AffineProductPushPlan{}, false
	}
	leftSlot, found := ctx.lookupSlot(leftName)
	if !found {
		return bytecodeF64AffineProductPushPlan{}, false
	}
	rightSlot, found := ctx.lookupSlot(rightName)
	if !found {
		return bytecodeF64AffineProductPushPlan{}, false
	}
	return bytecodeF64AffineProductPushPlan{
		receiverSlot: receiverSlot,
		scaleSlot:    scaleSlot,
		leftSlot:     leftSlot,
		rightSlot:    rightSlot,
	}, true
}

func bytecodeFlattenProductFactors(expr ast.Expression, factors []ast.Expression) []ast.Expression {
	bin, ok := expr.(*ast.BinaryExpression)
	if !ok || bin == nil || bin.Operator != "*" {
		return append(factors, expr)
	}
	factors = bytecodeFlattenProductFactors(bin.Left, factors)
	return bytecodeFlattenProductFactors(bin.Right, factors)
}

func bytecodeF64CastBinaryIdentifiers(expr ast.Expression, operator string) (string, string, bool) {
	cast, ok := expr.(*ast.TypeCastExpression)
	if !ok || cast == nil || typeExpressionToString(cast.TargetType) != "f64" {
		return "", "", false
	}
	bin, ok := cast.Expression.(*ast.BinaryExpression)
	if !ok || bin == nil || bin.Operator != operator {
		return "", "", false
	}
	left := bytecodeIdentifierExpressionName(bin.Left)
	right := bytecodeIdentifierExpressionName(bin.Right)
	return left, right, left != "" && right != ""
}

func bytecodeIdentifierExpressionName(expr ast.Expression) string {
	ident, ok := expr.(*ast.Identifier)
	if !ok || ident == nil {
		return ""
	}
	return ident.Name
}

func (plan bytecodeF64AffineProductPushPlan) validForSlots(slotCount int) bool {
	return plan.receiverSlot >= 0 && plan.receiverSlot < slotCount &&
		plan.scaleSlot >= 0 && plan.scaleSlot < slotCount &&
		plan.leftSlot >= 0 && plan.leftSlot < slotCount &&
		plan.rightSlot >= 0 && plan.rightSlot < slotCount
}
