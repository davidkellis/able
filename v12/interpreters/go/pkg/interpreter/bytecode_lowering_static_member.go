package interpreter

import (
	"able/interpreter-go/pkg/ast"
)

func bytecodeStaticReceiverInstruction(ctx *bytecodeLoweringContext, member *ast.MemberAccessExpression, memberName string, argCount int) (bytecodeInstruction, bool) {
	if ctx == nil || member == nil || member.Safe || memberName == "" {
		return bytecodeInstruction{}, false
	}
	ident, ok := member.Object.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return bytecodeInstruction{}, false
	}
	if _, _, found := ctx.lookupAnySlot(ident.Name); found {
		return bytecodeInstruction{}, false
	}
	if bytecodeArraySlotCallShape(memberName, argCount) && !bytecodeKnownStaticReceiver(ctx, ident.Name) {
		return bytecodeInstruction{}, false
	}
	return bytecodeInstruction{
		op:         bytecodeOpLoadStaticReceiver,
		name:       ident.Name,
		nameSimple: bytecodeSimpleLookupName(ident.Name),
		node:       ident,
	}, true
}

func bytecodeKnownStaticReceiver(ctx *bytecodeLoweringContext, name string) bool {
	if ctx == nil || name == "" {
		return false
	}
	if ctx.structDefs != nil && ctx.structDefs[name] != nil {
		return true
	}
	if ctx.structDefValues != nil && ctx.structDefValues[name] != nil {
		return true
	}
	return false
}

func bytecodeStaticMemberCallInstruction(memberName string, argCount int, node ast.Node) bytecodeInstruction {
	return bytecodeInstruction{
		op:       bytecodeOpCallStaticMember,
		name:     memberName,
		argCount: argCount,
		node:     node,
	}
}
