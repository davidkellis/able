package interpreter

import (
	"able/interpreter-go/pkg/ast"
)

type bytecodeStoreSlotFloatAddMulLoweringPlan struct {
	instr      bytecodeInstruction
	targetSlot int
	baseSlot   int
	baseName   string
	mulLeft    ast.Expression
	mulRight   ast.Expression
}

type bytecodeStoreSlotFloatAddMulSlotLoweringPlan struct {
	instr      bytecodeInstruction
	targetSlot int
	baseSlot   int
	baseName   string
	mulSlot    int
	mulName    string
	stackExpr  ast.Expression
}

type bytecodeArrayGetPropagationOperandPlan struct {
	receiverSlot int
	receiverName string
	indexSlot    int
	indexName    string
}

type bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan struct {
	instr             bytecodeInstruction
	targetSlot        int
	leftReceiverSlot  int
	leftReceiverName  string
	leftIndexSlot     int
	leftIndexName     string
	rightReceiverSlot int
	rightReceiverName string
	rightIndexSlot    int
	rightIndexName    string
}

func bytecodeStoreSlotFloatAddMulArrayGetPlan(ctx *bytecodeLoweringContext, targetName string, expr ast.Expression, node ast.Node) (bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil || targetName == "" {
		return bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan{}, false
	}
	add, ok := expr.(*ast.BinaryExpression)
	if !ok || add == nil || add.Operator != "+" {
		return bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan{}, false
	}
	left, ok := add.Left.(*ast.Identifier)
	if !ok || left == nil || left.Name != targetName {
		return bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan{}, false
	}
	mul, ok := add.Right.(*ast.BinaryExpression)
	if !ok || mul == nil || mul.Operator != "*" {
		return bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan{}, false
	}
	leftGet, ok := bytecodeArrayGetPropagationOperand(ctx, mul.Left)
	if !ok {
		return bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan{}, false
	}
	rightGet, ok := bytecodeArrayGetPropagationOperand(ctx, mul.Right)
	if !ok {
		return bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan{}, false
	}
	slot, found := ctx.lookupSlot(targetName)
	if !found {
		return bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan{}, false
	}
	if ctx.slotKind(slot) != bytecodeCellKindValue {
		return bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan{}, false
	}
	return bytecodeStoreSlotFloatAddMulArrayGetLoweringPlan{
		instr: bytecodeInstruction{
			op:       bytecodeOpStoreSlotFloatAddMulArrayGet,
			target:   slot,
			name:     targetName,
			operator: "+",
			node:     node,
		},
		targetSlot:        slot,
		leftReceiverSlot:  leftGet.receiverSlot,
		leftReceiverName:  leftGet.receiverName,
		leftIndexSlot:     leftGet.indexSlot,
		leftIndexName:     leftGet.indexName,
		rightReceiverSlot: rightGet.receiverSlot,
		rightReceiverName: rightGet.receiverName,
		rightIndexSlot:    rightGet.indexSlot,
		rightIndexName:    rightGet.indexName,
	}, true
}

func bytecodeArrayGetPropagationOperand(ctx *bytecodeLoweringContext, expr ast.Expression) (bytecodeArrayGetPropagationOperandPlan, bool) {
	prop, ok := expr.(*ast.PropagationExpression)
	if !ok || prop == nil {
		return bytecodeArrayGetPropagationOperandPlan{}, false
	}
	call, ok := prop.Expression.(*ast.FunctionCall)
	if !ok || call == nil || len(call.Arguments) != 1 || len(call.TypeArguments) != 0 {
		return bytecodeArrayGetPropagationOperandPlan{}, false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe || bytecodeIdentifierMemberName(member.Member) != "get" {
		return bytecodeArrayGetPropagationOperandPlan{}, false
	}
	receiver, ok := member.Object.(*ast.Identifier)
	if !ok || receiver == nil {
		return bytecodeArrayGetPropagationOperandPlan{}, false
	}
	index, ok := call.Arguments[0].(*ast.Identifier)
	if !ok || index == nil {
		return bytecodeArrayGetPropagationOperandPlan{}, false
	}
	receiverSlot, found := ctx.lookupSlot(receiver.Name)
	if !found {
		return bytecodeArrayGetPropagationOperandPlan{}, false
	}
	indexSlot, found := ctx.lookupSlot(index.Name)
	if !found {
		return bytecodeArrayGetPropagationOperandPlan{}, false
	}
	return bytecodeArrayGetPropagationOperandPlan{
		receiverSlot: receiverSlot,
		receiverName: receiver.Name,
		indexSlot:    indexSlot,
		indexName:    index.Name,
	}, true
}

func bytecodeStoreSlotFloatAddMulPlan(ctx *bytecodeLoweringContext, targetName string, expr ast.Expression, node ast.Node) (bytecodeStoreSlotFloatAddMulLoweringPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil || targetName == "" {
		return bytecodeStoreSlotFloatAddMulLoweringPlan{}, false
	}
	add, ok := expr.(*ast.BinaryExpression)
	if !ok || add == nil || add.Operator != "+" {
		return bytecodeStoreSlotFloatAddMulLoweringPlan{}, false
	}
	var (
		baseIdent *ast.Identifier
		mul       *ast.BinaryExpression
	)
	if candidate, ok := add.Left.(*ast.Identifier); ok && candidate != nil {
		if mulCandidate, ok := add.Right.(*ast.BinaryExpression); ok && mulCandidate != nil && mulCandidate.Operator == "*" && mulCandidate.Left != nil && mulCandidate.Right != nil {
			baseIdent = candidate
			mul = mulCandidate
		}
	}
	if baseIdent == nil {
		candidate, ok := add.Right.(*ast.Identifier)
		if !ok || candidate == nil {
			return bytecodeStoreSlotFloatAddMulLoweringPlan{}, false
		}
		mulCandidate, ok := add.Left.(*ast.BinaryExpression)
		if !ok || mulCandidate == nil || mulCandidate.Operator != "*" || mulCandidate.Left == nil || mulCandidate.Right == nil {
			return bytecodeStoreSlotFloatAddMulLoweringPlan{}, false
		}
		baseIdent = candidate
		mul = mulCandidate
	}
	slot, found := ctx.lookupSlot(targetName)
	if !found {
		return bytecodeStoreSlotFloatAddMulLoweringPlan{}, false
	}
	if ctx.slotKind(slot) != bytecodeCellKindValue || !bytecodeExpressionIsKnownFloat(ctx, expr) {
		return bytecodeStoreSlotFloatAddMulLoweringPlan{}, false
	}
	baseSlot, found := ctx.lookupSlot(baseIdent.Name)
	if !found {
		return bytecodeStoreSlotFloatAddMulLoweringPlan{}, false
	}
	if ctx.slotKind(baseSlot) != bytecodeCellKindValue {
		return bytecodeStoreSlotFloatAddMulLoweringPlan{}, false
	}
	return bytecodeStoreSlotFloatAddMulLoweringPlan{
		instr: bytecodeInstruction{
			op:       bytecodeOpStoreSlotFloatAddMul,
			target:   slot,
			name:     targetName,
			operator: "+",
			node:     node,
		},
		targetSlot: slot,
		baseSlot:   baseSlot,
		baseName:   baseIdent.Name,
		mulLeft:    mul.Left,
		mulRight:   mul.Right,
	}, true
}

func bytecodeStoreSlotFloatAddMulSlotPlan(ctx *bytecodeLoweringContext, targetName string, expr ast.Expression, node ast.Node) (bytecodeStoreSlotFloatAddMulSlotLoweringPlan, bool) {
	if ctx == nil || ctx.frameLayout == nil || targetName == "" {
		return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
	}
	add, ok := expr.(*ast.BinaryExpression)
	if !ok || add == nil || add.Operator != "+" {
		return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
	}

	var (
		baseIdent *ast.Identifier
		mul       *ast.BinaryExpression
	)
	if candidate, ok := add.Left.(*ast.Identifier); ok && candidate != nil {
		if mulCandidate, ok := add.Right.(*ast.BinaryExpression); ok && mulCandidate != nil && mulCandidate.Operator == "*" && mulCandidate.Left != nil && mulCandidate.Right != nil {
			baseIdent = candidate
			mul = mulCandidate
		}
	}
	if baseIdent == nil {
		candidate, ok := add.Right.(*ast.Identifier)
		if !ok || candidate == nil {
			return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
		}
		mulCandidate, ok := add.Left.(*ast.BinaryExpression)
		if !ok || mulCandidate == nil || mulCandidate.Operator != "*" || mulCandidate.Left == nil || mulCandidate.Right == nil {
			return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
		}
		baseIdent = candidate
		mul = mulCandidate
	}

	var (
		mulIdent  *ast.Identifier
		stackExpr ast.Expression
	)
	if candidate, ok := mul.Left.(*ast.Identifier); ok && candidate != nil {
		mulIdent = candidate
		stackExpr = mul.Right
	}
	if mulIdent == nil {
		candidate, ok := mul.Right.(*ast.Identifier)
		if !ok || candidate == nil {
			return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
		}
		mulIdent = candidate
		stackExpr = mul.Left
	}
	if stackExpr == nil {
		return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
	}

	slot, found := ctx.lookupSlot(targetName)
	if !found {
		return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
	}
	if ctx.slotKind(slot) != bytecodeCellKindValue || !bytecodeExpressionIsKnownFloat(ctx, expr) {
		return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
	}
	baseSlot, found := ctx.lookupSlot(baseIdent.Name)
	if !found {
		return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
	}
	if ctx.slotKind(baseSlot) != bytecodeCellKindValue {
		return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
	}
	mulSlot, found := ctx.lookupSlot(mulIdent.Name)
	if !found {
		return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
	}
	if ctx.slotKind(mulSlot) != bytecodeCellKindValue {
		return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{}, false
	}

	return bytecodeStoreSlotFloatAddMulSlotLoweringPlan{
		instr: bytecodeInstruction{
			op:        bytecodeOpStoreSlotFloatAddMulSlot,
			target:    slot,
			name:      targetName,
			operator:  "+",
			argCount:  baseSlot,
			loopBreak: mulSlot,
			node:      node,
		},
		targetSlot: slot,
		baseSlot:   baseSlot,
		baseName:   baseIdent.Name,
		mulSlot:    mulSlot,
		mulName:    mulIdent.Name,
		stackExpr:  stackExpr,
	}, true
}
