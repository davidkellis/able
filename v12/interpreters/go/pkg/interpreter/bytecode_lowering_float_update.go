package interpreter

import (
	"able/interpreter-go/pkg/ast"
)

type bytecodeStoreSlotFloatAddMulLoweringPlan struct {
	instr      bytecodeInstruction
	targetSlot int
	mulLeft    ast.Expression
	mulRight   ast.Expression
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
	left, ok := add.Left.(*ast.Identifier)
	if !ok || left == nil || left.Name != targetName {
		return bytecodeStoreSlotFloatAddMulLoweringPlan{}, false
	}
	mul, ok := add.Right.(*ast.BinaryExpression)
	if !ok || mul == nil || mul.Operator != "*" || mul.Left == nil || mul.Right == nil {
		return bytecodeStoreSlotFloatAddMulLoweringPlan{}, false
	}
	slot, found := ctx.lookupSlot(targetName)
	if !found {
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
		mulLeft:    mul.Left,
		mulRight:   mul.Right,
	}, true
}
