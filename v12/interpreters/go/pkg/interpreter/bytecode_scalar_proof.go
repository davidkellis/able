package interpreter

import "able/interpreter-go/pkg/ast"

// bytecodeScalarProofKind describes why a generic VM instruction can, or
// cannot, be represented by scalar-only execution. It is diagnostic metadata:
// collecting it never changes opcode selection or runtime semantics.
type bytecodeScalarProofKind uint8

const (
	bytecodeScalarProofNotTarget bytecodeScalarProofKind = iota
	bytecodeScalarProofUnproven
	bytecodeScalarProofExistingLane
	bytecodeScalarProofSlotInteger
	bytecodeScalarProofSlotFloat
	bytecodeScalarProofSlotBool
	bytecodeScalarProofSlotChar
	bytecodeScalarProofFieldInteger
	bytecodeScalarProofFieldFloat
	bytecodeScalarProofFieldBool
	bytecodeScalarProofFieldChar
	bytecodeScalarProofNumericCast
	bytecodeScalarProofIntegerCompare
	bytecodeScalarProofFloatCompare
	bytecodeScalarProofBoolCompare
	bytecodeScalarProofCharCompare
	bytecodeScalarProofCount
)

// BytecodeScalarProofSnapshot reports deterministic static reach and dynamic
// execution counts for one proof class within a source-attributed program.
type BytecodeScalarProofSnapshot struct {
	Opcode              string `json:"opcode"`
	Proof               string `json:"proof"`
	StaticInstructions  int    `json:"static_instructions"`
	DynamicInstructions uint64 `json:"dynamic_instructions"`
}

type bytecodeScalarProofTarget uint8

const (
	bytecodeScalarProofTargetLoadSlot bytecodeScalarProofTarget = iota
	bytecodeScalarProofTargetLoadImplicitSlot
	bytecodeScalarProofTargetStoreSlot
	bytecodeScalarProofTargetStoreSlotNew
	bytecodeScalarProofTargetStoreImplicitSlot
	bytecodeScalarProofTargetLoadSlotStructField
	bytecodeScalarProofTargetCast
	bytecodeScalarProofTargetBinary
	bytecodeScalarProofTargetJumpIfBinaryCompareFalse
	bytecodeScalarProofTargetCount
)

func bytecodeScalarProofTargetForOp(op bytecodeOp) (bytecodeScalarProofTarget, bool) {
	switch op {
	case bytecodeOpLoadSlot:
		return bytecodeScalarProofTargetLoadSlot, true
	case bytecodeOpLoadImplicitSlot:
		return bytecodeScalarProofTargetLoadImplicitSlot, true
	case bytecodeOpStoreSlot:
		return bytecodeScalarProofTargetStoreSlot, true
	case bytecodeOpStoreSlotNew:
		return bytecodeScalarProofTargetStoreSlotNew, true
	case bytecodeOpStoreImplicitSlot:
		return bytecodeScalarProofTargetStoreImplicitSlot, true
	case bytecodeOpLoadSlotStructField:
		return bytecodeScalarProofTargetLoadSlotStructField, true
	case bytecodeOpCast:
		return bytecodeScalarProofTargetCast, true
	case bytecodeOpBinary:
		return bytecodeScalarProofTargetBinary, true
	case bytecodeOpJumpIfBinaryCompareFalse:
		return bytecodeScalarProofTargetJumpIfBinaryCompareFalse, true
	default:
		return 0, false
	}
}

func (target bytecodeScalarProofTarget) String() string {
	switch target {
	case bytecodeScalarProofTargetLoadSlot:
		return "LoadSlot"
	case bytecodeScalarProofTargetLoadImplicitSlot:
		return "LoadImplicitSlot"
	case bytecodeScalarProofTargetStoreSlot:
		return "StoreSlot"
	case bytecodeScalarProofTargetStoreSlotNew:
		return "StoreSlotNew"
	case bytecodeScalarProofTargetStoreImplicitSlot:
		return "StoreImplicitSlot"
	case bytecodeScalarProofTargetLoadSlotStructField:
		return "LoadSlotStructField"
	case bytecodeScalarProofTargetCast:
		return "Cast"
	case bytecodeScalarProofTargetBinary:
		return "Binary"
	case bytecodeScalarProofTargetJumpIfBinaryCompareFalse:
		return "JumpIfBinaryCompareFalse"
	default:
		return "Unknown"
	}
}

func bytecodeScalarProofUsesSlotCheck(op bytecodeOp) bool {
	switch op {
	case bytecodeOpLoadSlot, bytecodeOpLoadImplicitSlot,
		bytecodeOpStoreSlot, bytecodeOpStoreSlotNew, bytecodeOpStoreImplicitSlot:
		return true
	default:
		return false
	}
}

func bytecodeScalarProofForInstruction(program *bytecodeProgram, ip int, instr *bytecodeInstruction, loweringCheck bytecodeSimpleTypeCheck, inferred bytecodeInferenceFacts) bytecodeScalarProofKind {
	if instr == nil {
		return bytecodeScalarProofNotTarget
	}
	if bytecodeScalarProofUsesSlotCheck(instr.op) {
		if bytecodeProgramSlotHasPrimitiveType(program, instr.target) {
			return bytecodeScalarProofExistingLane
		}
		check := bytecodeScalarSlotInferenceCheck(instr, inferred)
		if check == bytecodeSimpleTypeCheckUnknown {
			check = loweringCheck
		}
		return bytecodeScalarSlotProofKind(check)
	}
	switch instr.op {
	case bytecodeOpLoadSlotStructField:
		check := bytecodeInferenceSimpleCheck(inferred, instr.node)
		if check == bytecodeSimpleTypeCheckUnknown {
			check = bytecodeStructFieldSimpleCheck(program, ip)
		}
		return bytecodeScalarFieldProofKind(check)
	case bytecodeOpCast:
		cast, ok := instr.node.(*ast.TypeCastExpression)
		if !ok || cast == nil {
			return bytecodeScalarProofUnproven
		}
		source := bytecodeInferenceSimpleCheck(inferred, cast.Expression)
		if source == bytecodeSimpleTypeCheckUnknown {
			source = bytecodeExpressionSimpleTypeCheck(nil, cast.Expression)
		}
		target := bytecodeInferenceSimpleCheck(inferred, cast)
		if target == bytecodeSimpleTypeCheckUnknown {
			target = bytecodeSimpleTypeCheckForName(cachedSimpleTypeName(cast.TargetType))
		}
		if bytecodeIsNumericSimpleCheck(source) && bytecodeIsNumericSimpleCheck(target) {
			return bytecodeScalarProofNumericCast
		}
		return bytecodeScalarProofUnproven
	case bytecodeOpBinary, bytecodeOpJumpIfBinaryCompareFalse:
		binary, ok := instr.node.(*ast.BinaryExpression)
		if !ok || binary == nil || !bytecodeScalarComparisonOperator(binary.Operator) {
			return bytecodeScalarProofNotTarget
		}
		left := bytecodeInferenceSimpleCheck(inferred, binary.Left)
		right := bytecodeInferenceSimpleCheck(inferred, binary.Right)
		if left == bytecodeSimpleTypeCheckUnknown {
			left = bytecodeExpressionSimpleTypeCheck(nil, binary.Left)
		}
		if right == bytecodeSimpleTypeCheckUnknown {
			right = bytecodeExpressionSimpleTypeCheck(nil, binary.Right)
		}
		switch {
		case bytecodeScalarCheckIsInteger(left) && bytecodeScalarCheckIsInteger(right):
			return bytecodeScalarProofIntegerCompare
		case bytecodeIsNumericSimpleCheck(left) && bytecodeIsNumericSimpleCheck(right):
			return bytecodeScalarProofFloatCompare
		case left == bytecodeSimpleTypeCheckBool && right == bytecodeSimpleTypeCheckBool && (binary.Operator == "==" || binary.Operator == "!="):
			return bytecodeScalarProofBoolCompare
		case left == bytecodeSimpleTypeCheckChar && right == bytecodeSimpleTypeCheckChar:
			return bytecodeScalarProofCharCompare
		default:
			return bytecodeScalarProofUnproven
		}
	default:
		return bytecodeScalarProofNotTarget
	}
}

func bytecodeScalarCheckIsInteger(check bytecodeSimpleTypeCheck) bool {
	if check == bytecodeSimpleTypeCheckAnyInteger {
		return true
	}
	_, ok := check.integerType()
	return ok
}

func bytecodeScalarSlotInferenceCheck(instr *bytecodeInstruction, inferred bytecodeInferenceFacts) bytecodeSimpleTypeCheck {
	if instr == nil {
		return bytecodeSimpleTypeCheckUnknown
	}
	switch instr.op {
	case bytecodeOpLoadSlot, bytecodeOpLoadImplicitSlot:
		if _, ok := instr.node.(*ast.Identifier); ok {
			return bytecodeInferenceSimpleCheck(inferred, instr.node)
		}
	case bytecodeOpStoreSlot, bytecodeOpStoreSlotNew, bytecodeOpStoreImplicitSlot:
		if assignment, ok := instr.node.(*ast.AssignmentExpression); ok && assignment != nil {
			return bytecodeInferenceSimpleCheck(inferred, assignment.Right)
		}
	}
	return bytecodeSimpleTypeCheckUnknown
}

func bytecodeStructFieldSimpleCheck(program *bytecodeProgram, ip int) bytecodeSimpleTypeCheck {
	if program == nil || program.namedStructMembers == nil {
		return bytecodeSimpleTypeCheckUnknown
	}
	plan, ok := program.namedStructMembers[ip]
	if !ok || plan.definition == nil || plan.definition.Node == nil || plan.fieldIndex < 0 || plan.fieldIndex >= len(plan.definition.Node.Fields) {
		return bytecodeSimpleTypeCheckUnknown
	}
	field := plan.definition.Node.Fields[plan.fieldIndex]
	if field == nil {
		return bytecodeSimpleTypeCheckUnknown
	}
	return bytecodeSimpleTypeCheckForName(cachedSimpleTypeName(field.FieldType))
}

func bytecodeScalarSlotProofKind(check bytecodeSimpleTypeCheck) bytecodeScalarProofKind {
	switch {
	case check == bytecodeSimpleTypeCheckAnyInteger:
		return bytecodeScalarProofSlotInteger
	case bytecodeIsFloatSimpleCheck(check):
		return bytecodeScalarProofSlotFloat
	case check == bytecodeSimpleTypeCheckBool:
		return bytecodeScalarProofSlotBool
	case check == bytecodeSimpleTypeCheckChar:
		return bytecodeScalarProofSlotChar
	default:
		if _, ok := check.integerType(); ok {
			return bytecodeScalarProofSlotInteger
		}
		return bytecodeScalarProofUnproven
	}
}

func bytecodeScalarFieldProofKind(check bytecodeSimpleTypeCheck) bytecodeScalarProofKind {
	switch {
	case check == bytecodeSimpleTypeCheckAnyInteger:
		return bytecodeScalarProofFieldInteger
	case bytecodeIsFloatSimpleCheck(check):
		return bytecodeScalarProofFieldFloat
	case check == bytecodeSimpleTypeCheckBool:
		return bytecodeScalarProofFieldBool
	case check == bytecodeSimpleTypeCheckChar:
		return bytecodeScalarProofFieldChar
	default:
		if _, ok := check.integerType(); ok {
			return bytecodeScalarProofFieldInteger
		}
		return bytecodeScalarProofUnproven
	}
}

func bytecodeScalarComparisonOperator(operator string) bool {
	switch operator {
	case "<", "<=", ">", ">=", "==", "!=":
		return true
	default:
		return false
	}
}

func (kind bytecodeScalarProofKind) primitiveEligible() bool {
	return kind > bytecodeScalarProofUnproven && kind < bytecodeScalarProofCount
}

func (kind bytecodeScalarProofKind) String() string {
	switch kind {
	case bytecodeScalarProofUnproven:
		return "unproven-or-boxed"
	case bytecodeScalarProofExistingLane:
		return "existing-scalar-lane"
	case bytecodeScalarProofSlotInteger:
		return "primitive-slot-integer"
	case bytecodeScalarProofSlotFloat:
		return "primitive-slot-float"
	case bytecodeScalarProofSlotBool:
		return "primitive-slot-bool"
	case bytecodeScalarProofSlotChar:
		return "primitive-slot-char"
	case bytecodeScalarProofFieldInteger:
		return "primitive-field-integer"
	case bytecodeScalarProofFieldFloat:
		return "primitive-field-float"
	case bytecodeScalarProofFieldBool:
		return "primitive-field-bool"
	case bytecodeScalarProofFieldChar:
		return "primitive-field-char"
	case bytecodeScalarProofNumericCast:
		return "primitive-numeric-cast"
	case bytecodeScalarProofIntegerCompare:
		return "primitive-integer-compare"
	case bytecodeScalarProofFloatCompare:
		return "primitive-float-compare"
	case bytecodeScalarProofBoolCompare:
		return "primitive-bool-compare"
	case bytecodeScalarProofCharCompare:
		return "primitive-char-compare"
	default:
		return "not-targeted"
	}
}
