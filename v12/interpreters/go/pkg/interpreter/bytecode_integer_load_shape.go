package interpreter

import (
	"reflect"
	"strings"
	"sync/atomic"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// Integer load shapes are diagnostic-only. They describe the runtime carrier
// and first source-enclosing consumer of a verifier-proven integer LoadSlot.
type bytecodeIntegerLoadCarrier uint8

const (
	bytecodeIntegerLoadCarrierUnknown bytecodeIntegerLoadCarrier = iota
	bytecodeIntegerLoadCarrierRawI64Cell
	bytecodeIntegerLoadCarrierRawIntegerCell
	bytecodeIntegerLoadCarrierRawI32Value
	bytecodeIntegerLoadCarrierRawIntegerValue
	bytecodeIntegerLoadCarrierI32Register
	bytecodeIntegerLoadCarrierI32Sidecar
	bytecodeIntegerLoadCarrierSmallIntegerPointer
	bytecodeIntegerLoadCarrierSmallIntegerValue
	bytecodeIntegerLoadCarrierBigInteger
	bytecodeIntegerLoadCarrierOther
	bytecodeIntegerLoadCarrierCount
)

func (carrier bytecodeIntegerLoadCarrier) String() string {
	switch carrier {
	case bytecodeIntegerLoadCarrierRawI64Cell:
		return "raw-i64-slot-cell"
	case bytecodeIntegerLoadCarrierRawIntegerCell:
		return "raw-integer-slot-cell"
	case bytecodeIntegerLoadCarrierRawI32Value:
		return "raw-i32-slot-value"
	case bytecodeIntegerLoadCarrierRawIntegerValue:
		return "raw-integer-value"
	case bytecodeIntegerLoadCarrierI32Register:
		return "i32-register"
	case bytecodeIntegerLoadCarrierI32Sidecar:
		return "i32-value-sidecar"
	case bytecodeIntegerLoadCarrierSmallIntegerPointer:
		return "small-integer-pointer"
	case bytecodeIntegerLoadCarrierSmallIntegerValue:
		return "small-integer-value"
	case bytecodeIntegerLoadCarrierBigInteger:
		return "big-integer"
	case bytecodeIntegerLoadCarrierOther:
		return "other-or-mismatch"
	default:
		return "unknown-or-uninitialized"
	}
}

type bytecodeIntegerLoadConsumer uint8

const (
	bytecodeIntegerLoadConsumerUnknown bytecodeIntegerLoadConsumer = iota
	bytecodeIntegerLoadConsumerArithmetic
	bytecodeIntegerLoadConsumerComparison
	bytecodeIntegerLoadConsumerCast
	bytecodeIntegerLoadConsumerCallArgument
	bytecodeIntegerLoadConsumerCollectionIndex
	bytecodeIntegerLoadConsumerStore
	bytecodeIntegerLoadConsumerBranchCondition
	bytecodeIntegerLoadConsumerReturn
	bytecodeIntegerLoadConsumerOther
	bytecodeIntegerLoadConsumerCount
)

func (consumer bytecodeIntegerLoadConsumer) String() string {
	switch consumer {
	case bytecodeIntegerLoadConsumerArithmetic:
		return "arithmetic"
	case bytecodeIntegerLoadConsumerComparison:
		return "comparison"
	case bytecodeIntegerLoadConsumerCast:
		return "cast"
	case bytecodeIntegerLoadConsumerCallArgument:
		return "call-argument"
	case bytecodeIntegerLoadConsumerCollectionIndex:
		return "collection-index"
	case bytecodeIntegerLoadConsumerStore:
		return "store"
	case bytecodeIntegerLoadConsumerBranchCondition:
		return "branch-condition"
	case bytecodeIntegerLoadConsumerReturn:
		return "return"
	case bytecodeIntegerLoadConsumerOther:
		return "other"
	default:
		return "unknown"
	}
}

// bytecodeIntegerLoadOperandRole refines call-argument attribution only where
// the lowered instruction has a stable language/kernel operand contract.
type bytecodeIntegerLoadOperandRole uint8

const (
	bytecodeIntegerLoadOperandRoleUnknown bytecodeIntegerLoadOperandRole = iota
	bytecodeIntegerLoadOperandRoleReceiver
	bytecodeIntegerLoadOperandRoleIndex
	bytecodeIntegerLoadOperandRoleValue
	bytecodeIntegerLoadOperandRoleOther
)

func (role bytecodeIntegerLoadOperandRole) String() string {
	switch role {
	case bytecodeIntegerLoadOperandRoleReceiver:
		return "receiver"
	case bytecodeIntegerLoadOperandRoleIndex:
		return "index"
	case bytecodeIntegerLoadOperandRoleValue:
		return "value"
	case bytecodeIntegerLoadOperandRoleOther:
		return "other-call-operand"
	default:
		return ""
	}
}

// BytecodeIntegerLoadShapeSnapshot reports dynamic occurrences of one proven
// integer slot carrier/consumer pair within a source-attributed program.
type BytecodeIntegerLoadShapeSnapshot struct {
	Carrier             string `json:"carrier"`
	Consumer            string `json:"consumer"`
	ConsumerOpcode      string `json:"consumer_opcode,omitempty"`
	ConsumerOperation   string `json:"consumer_operation,omitempty"`
	ConsumerOperandRole string `json:"consumer_operand_role,omitempty"`
	DynamicInstructions uint64 `json:"dynamic_instructions"`
}

type bytecodeIntegerLoadUse struct {
	consumer  bytecodeIntegerLoadConsumer
	op        bytecodeOp
	operation string
	role      bytecodeIntegerLoadOperandRole
}

func bytecodeIntegerLoadCarrierForSlot(vm *bytecodeVM, slot int) bytecodeIntegerLoadCarrier {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return bytecodeIntegerLoadCarrierUnknown
	}
	switch value := vm.slots[slot].(type) {
	case *bytecodeRawI64SlotCell:
		if value != nil {
			return bytecodeIntegerLoadCarrierRawI64Cell
		}
	case *bytecodeRawIntegerSlotCell:
		if value != nil {
			return bytecodeIntegerLoadCarrierRawIntegerCell
		}
	case bytecodeRawI32SlotValue:
		return bytecodeIntegerLoadCarrierRawI32Value
	case *bytecodeRawI32StackCell:
		if value != nil {
			return bytecodeIntegerLoadCarrierRawI32Value
		}
	case bytecodeRawIntegerValue, bytecodeRawU8ResultValue,
		bytecodeRawU16ResultValue, bytecodeRawU32ResultValue,
		bytecodeRawU64ResultValue, bytecodeRawUsizeResultValue,
		bytecodeRawI64ResultValue:
		return bytecodeIntegerLoadCarrierRawIntegerValue
	case *runtime.IntegerValue:
		if value == nil {
			break
		}
		if value.IsSmallRef() {
			return bytecodeIntegerLoadCarrierSmallIntegerPointer
		}
		return bytecodeIntegerLoadCarrierBigInteger
	case runtime.IntegerValue:
		if value.IsSmall() {
			return bytecodeIntegerLoadCarrierSmallIntegerValue
		}
		return bytecodeIntegerLoadCarrierBigInteger
	case nil:
		if _, ok := vm.i32RegisterRaw(slot); ok {
			return bytecodeIntegerLoadCarrierI32Register
		}
		if _, ok := vm.activeValueSlotI32Raw(slot); ok {
			return bytecodeIntegerLoadCarrierI32Sidecar
		}
		return bytecodeIntegerLoadCarrierUnknown
	default:
		return bytecodeIntegerLoadCarrierOther
	}
	return bytecodeIntegerLoadCarrierUnknown
}

func bytecodeIntegerLoadConsumerForProgram(program *bytecodeProgram, loadIP int) bytecodeIntegerLoadUse {
	if program == nil || loadIP < 0 || loadIP >= len(program.instructions) {
		return bytecodeIntegerLoadUse{}
	}
	target := program.instructions[loadIP].node
	if target == nil {
		return bytecodeIntegerLoadUse{}
	}
	const scanLimit = 64
	end := loadIP + scanLimit + 1
	if end > len(program.instructions) {
		end = len(program.instructions)
	}
	for ip := loadIP + 1; ip < end; ip++ {
		instr := &program.instructions[ip]
		consumer, consumes := bytecodeIntegerLoadConsumerForInstruction(instr)
		if consumes && bytecodeASTContainsNode(instr.node, target) {
			return bytecodeIntegerLoadUse{
				consumer:  consumer,
				op:        instr.op,
				operation: bytecodeIntegerLoadConsumerOperation(instr),
				role:      bytecodeIntegerLoadOperandRoleForInstruction(instr, target),
			}
		}
		if ip == loadIP+1 && consumer == bytecodeIntegerLoadConsumerReturn && instr.node == nil {
			return bytecodeIntegerLoadUse{consumer: consumer, op: instr.op}
		}
		if bytecodeIntegerLoadScanBoundary(instr.op) {
			break
		}
	}
	return bytecodeIntegerLoadUse{}
}

func bytecodeIntegerLoadConsumerOperation(instr *bytecodeInstruction) string {
	if instr == nil || instr.op != bytecodeOpCallMemberArraySlot {
		return ""
	}
	switch instr.memberFastPath {
	case bytecodeMemberMethodFastPathArrayReadSlot:
		return "read-slot"
	case bytecodeMemberMethodFastPathArrayWriteSlot:
		return "write-slot"
	case bytecodeMemberMethodFastPathArrayPush:
		return "push"
	case bytecodeMemberMethodFastPathArrayLen:
		return "len"
	default:
		return "other-array-slot-call"
	}
}

func bytecodeIntegerLoadOperandRoleForInstruction(instr *bytecodeInstruction, target ast.Node) bytecodeIntegerLoadOperandRole {
	if instr == nil || instr.op != bytecodeOpCallMemberArraySlot || target == nil {
		return bytecodeIntegerLoadOperandRoleUnknown
	}
	call, ok := instr.node.(*ast.FunctionCall)
	if !ok || call == nil {
		return bytecodeIntegerLoadOperandRoleUnknown
	}
	if member, ok := call.Callee.(*ast.MemberAccessExpression); ok && member != nil &&
		bytecodeASTContainsNode(member.Object, target) {
		return bytecodeIntegerLoadOperandRoleReceiver
	}
	for index, argument := range call.Arguments {
		if !bytecodeASTContainsNode(argument, target) {
			continue
		}
		switch instr.memberFastPath {
		case bytecodeMemberMethodFastPathArrayReadSlot,
			bytecodeMemberMethodFastPathArrayWriteSlot:
			if index == 0 {
				return bytecodeIntegerLoadOperandRoleIndex
			}
			if instr.memberFastPath == bytecodeMemberMethodFastPathArrayWriteSlot && index == 1 {
				return bytecodeIntegerLoadOperandRoleValue
			}
		case bytecodeMemberMethodFastPathArrayPush:
			if index == 0 {
				return bytecodeIntegerLoadOperandRoleValue
			}
		}
		return bytecodeIntegerLoadOperandRoleOther
	}
	return bytecodeIntegerLoadOperandRoleOther
}

func bytecodeIntegerLoadConsumerForInstruction(instr *bytecodeInstruction) (bytecodeIntegerLoadConsumer, bool) {
	if instr == nil {
		return bytecodeIntegerLoadConsumerUnknown, false
	}
	name := bytecodeOpName(instr.op)
	switch {
	case instr.op == bytecodeOpCast:
		return bytecodeIntegerLoadConsumerCast, true
	case instr.op == bytecodeOpBinary:
		if bytecodeScalarComparisonOperator(instr.operator) {
			return bytecodeIntegerLoadConsumerComparison, true
		}
		return bytecodeIntegerLoadConsumerArithmetic, true
	case strings.Contains(name, "Compare") || strings.Contains(name, "LessEqual"):
		return bytecodeIntegerLoadConsumerComparison, true
	case strings.HasPrefix(name, "Binary") || strings.HasPrefix(name, "Unary"):
		return bytecodeIntegerLoadConsumerArithmetic, true
	case strings.HasPrefix(name, "Call") || strings.HasPrefix(name, "TryArrayPush"):
		return bytecodeIntegerLoadConsumerCallArgument, true
	case strings.Contains(name, "Index") || strings.Contains(name, "ArrayRead"):
		return bytecodeIntegerLoadConsumerCollectionIndex, true
	case strings.HasPrefix(name, "Store") || strings.HasPrefix(name, "Assign") || strings.HasPrefix(name, "CompoundAssign"):
		return bytecodeIntegerLoadConsumerStore, true
	case strings.HasPrefix(name, "JumpIf"):
		return bytecodeIntegerLoadConsumerBranchCondition, true
	case strings.HasPrefix(name, "Return"):
		return bytecodeIntegerLoadConsumerReturn, true
	case instr.op == bytecodeOpLoadSlot || instr.op == bytecodeOpLoadImplicitSlot ||
		instr.op == bytecodeOpLoadSlotI32 || instr.op == bytecodeOpConst ||
		instr.op == bytecodeOpConstI32 || instr.op == bytecodeOpDup:
		return bytecodeIntegerLoadConsumerUnknown, false
	default:
		return bytecodeIntegerLoadConsumerOther, instr.node != nil
	}
}

func bytecodeIntegerLoadScanBoundary(op bytecodeOp) bool {
	name := bytecodeOpName(op)
	return op == bytecodeOpJump || strings.HasPrefix(name, "Return") ||
		strings.HasPrefix(name, "Break") || strings.HasPrefix(name, "Continue") ||
		strings.HasPrefix(name, "Raise") || strings.HasPrefix(name, "Rethrow")
}

var bytecodeASTNodeType = reflect.TypeOf((*ast.Node)(nil)).Elem()

func bytecodeASTContainsNode(root, target ast.Node) bool {
	if root == nil || target == nil {
		return false
	}
	visited := make(map[uintptr]struct{})
	return bytecodeASTValueContainsNode(reflect.ValueOf(root), target, visited)
}

func bytecodeASTValueContainsNode(value reflect.Value, target ast.Node, visited map[uintptr]struct{}) bool {
	if !value.IsValid() {
		return false
	}
	for value.Kind() == reflect.Interface {
		if value.IsNil() {
			return false
		}
		value = value.Elem()
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return false
		}
		if value.Type().Implements(bytecodeASTNodeType) && value.CanInterface() {
			if node, ok := value.Interface().(ast.Node); ok && node == target {
				return true
			}
		}
		pointer := value.Pointer()
		if _, ok := visited[pointer]; ok {
			return false
		}
		visited[pointer] = struct{}{}
		return bytecodeASTValueContainsNode(value.Elem(), target, visited)
	}
	switch value.Kind() {
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			field := value.Field(index)
			if field.CanInterface() && bytecodeASTValueContainsNode(field, target, visited) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			if bytecodeASTValueContainsNode(value.Index(index), target, visited) {
				return true
			}
		}
	case reflect.Map:
		iterator := value.MapRange()
		for iterator.Next() {
			if bytecodeASTValueContainsNode(iterator.Value(), target, visited) {
				return true
			}
		}
	}
	return false
}

func (vm *bytecodeVM) recordProvenIntegerLoadShape(program *bytecodeProgram, ip int, instr *bytecodeInstruction) {
	if vm == nil || program == nil || program.reach == nil || instr == nil ||
		instr.op != bytecodeOpLoadSlot || ip < 0 || ip >= len(program.reach.instructionScalarProofs) ||
		program.reach.instructionScalarProofs[ip] != bytecodeScalarProofSlotInteger {
		return
	}
	carrier := bytecodeIntegerLoadCarrierForSlot(vm, instr.target)
	if ip >= len(program.reach.dynamicIntegerLoadCarriers) {
		return
	}
	atomic.AddUint64(&program.reach.dynamicIntegerLoadCarriers[ip][carrier], 1)
}
