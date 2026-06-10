package interpreter

import "able/interpreter-go/pkg/ast"

// bytecodeArrayOwnershipProgramMetadata is a conservative lowering summary
// for the release-disabled Array ownership observer. It identifies only
// language/kernel Array creation boundaries; it is not a named stdlib
// container rule and does not authorize reclamation.
type bytecodeArrayOwnershipProgramMetadata struct {
	hasCanonicalCreation bool
	hasCaptureBarrier    bool
	hasDynamicBarrier    bool
	hasSpawnBarrier      bool
	hasAggregateBarrier  bool
}

func (metadata bytecodeArrayOwnershipProgramMetadata) observesArrays() bool {
	return metadata.hasCanonicalCreation
}

func (metadata bytecodeArrayOwnershipProgramMetadata) releaseEligible() bool {
	return metadata.hasCanonicalCreation &&
		!metadata.hasCaptureBarrier &&
		!metadata.hasDynamicBarrier &&
		!metadata.hasSpawnBarrier &&
		!metadata.hasAggregateBarrier
}

func bytecodeArrayOwnershipMetadataForInstructions(instructions []bytecodeInstruction) bytecodeArrayOwnershipProgramMetadata {
	var metadata bytecodeArrayOwnershipProgramMetadata
	for _, instruction := range instructions {
		switch instruction.op {
		case bytecodeOpArrayLiteral, bytecodeOpCallMemberArrayNew:
			metadata.hasCanonicalCreation = true
		case bytecodeOpCallStaticMember:
			if instruction.name == "new" && bytecodeArrayOwnershipIsCanonicalArrayNewCall(instruction.node) {
				metadata.hasCanonicalCreation = true
			}
		case bytecodeOpStructLiteralNamedFast:
			if bytecodeArrayOwnershipIsCanonicalArrayLiteral(instruction.node) {
				metadata.hasCanonicalCreation = true
			}
		case bytecodeOpMakeFunction, bytecodeOpPlaceholderLambda, bytecodeOpIteratorLiteral:
			metadata.hasCaptureBarrier = true
		case bytecodeOpDynImport:
			metadata.hasDynamicBarrier = true
		case bytecodeOpSpawn:
			metadata.hasSpawnBarrier = true
		case bytecodeOpStructLiteral, bytecodeOpMapLiteral:
			metadata.hasAggregateBarrier = true
		}
	}
	return metadata
}

func bytecodeArrayOwnershipIsCanonicalArrayLiteral(node ast.Node) bool {
	literal, ok := node.(*ast.StructLiteral)
	return ok && literal != nil && literal.StructType != nil && literal.StructType.Name == "Array"
}

func bytecodeArrayOwnershipIsCanonicalArrayNewCall(node ast.Node) bool {
	call, ok := node.(*ast.FunctionCall)
	if !ok || call == nil {
		return false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || bytecodeIdentifierMemberName(member.Member) != "new" {
		return false
	}
	receiver, ok := member.Object.(*ast.Identifier)
	return ok && receiver != nil && receiver.Name == "Array"
}
