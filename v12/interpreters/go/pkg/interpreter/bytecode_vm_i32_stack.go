package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) pushI32(value int32) {
	vm.i32Stack = append(vm.i32Stack, value)
}

func (vm *bytecodeVM) popI32() (int32, error) {
	if len(vm.i32Stack) == 0 {
		return 0, fmt.Errorf("bytecode i32 stack underflow")
	}
	idx := len(vm.i32Stack) - 1
	value := vm.i32Stack[idx]
	vm.i32Stack = vm.i32Stack[:idx]
	return value, nil
}

func (vm *bytecodeVM) popBoxedI32() (runtime.Value, error) {
	value, err := vm.popI32()
	if err != nil {
		return nil, err
	}
	return bytecodeBoxedIntegerI32Value(int64(value)), nil
}

func (vm *bytecodeVM) execConstI32(instr *bytecodeInstruction) error {
	if instr == nil || !instr.hasIntRaw || instr.intImmediateRaw < math.MinInt32 || instr.intImmediateRaw > math.MaxInt32 {
		return fmt.Errorf("bytecode i32 const missing valid immediate")
	}
	vm.pushI32(int32(instr.intImmediateRaw))
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execUnboxI32(instr *bytecodeInstruction) error {
	if vm == nil || vm.interp == nil {
		return fmt.Errorf("bytecode i32 unbox missing interpreter")
	}
	value, err := vm.pop()
	if err != nil {
		return err
	}
	raw, fallback, ok, err := vm.unboxAssignableI32Value(value)
	if err != nil {
		return err
	}
	if !ok {
		vm.i32UnboxFallbackValue = fallback
		vm.i32UnboxFallbackSet = true
		vm.ip++
		return nil
	}
	vm.i32UnboxFallbackValue = nil
	vm.i32UnboxFallbackSet = false
	vm.pushI32(raw)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execBoxI32() error {
	value, err := vm.popBoxedI32()
	if err != nil {
		return err
	}
	vm.appendStackValue(value)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) unboxAssignableI32Value(value runtime.Value) (int32, runtime.Value, bool, error) {
	storeVal, fallback, shouldStore, err := vm.typedSlotAssignmentValues(bytecodeInstruction{
		storeTyped: true,
		typeExpr:   cachedSimpleTypeExpression("i32"),
	}, value)
	if err != nil {
		return 0, nil, false, err
	}
	if !shouldStore {
		return 0, fallback, false, nil
	}
	raw, ok := bytecodeRawI32Value(storeVal)
	if !ok {
		return 0, nil, false, fmt.Errorf("bytecode i32 unbox expected i32 value")
	}
	return raw, nil, true, nil
}

func bytecodeRawI32Value(value runtime.Value) (int32, bool) {
	var intVal runtime.IntegerValue
	switch v := value.(type) {
	case bytecodeRawI32SlotValue:
		return int32(v), true
	case *bytecodeRawI32StackCell:
		if v != nil {
			return v.Val, true
		}
		return 0, false
	case runtime.IntegerValue:
		intVal = v
	case *runtime.IntegerValue:
		if v == nil {
			return 0, false
		}
		intVal = *v
	default:
		return 0, false
	}
	if intVal.TypeSuffix != runtime.IntegerI32 {
		return 0, false
	}
	raw, ok := intVal.ToInt64()
	if !ok || raw < math.MinInt32 || raw > math.MaxInt32 {
		return 0, false
	}
	return int32(raw), true
}

func (vm *bytecodeVM) execLoadSlotI32(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode i32 slot load missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	if value, ok := vm.i32RegisterRaw(instr.target); ok {
		vm.pushI32(value)
		vm.ip++
		return nil
	}
	if value, ok := vm.activeValueSlotI32Raw(instr.target); ok {
		vm.pushI32(value)
		vm.ip++
		return nil
	}
	value, ok := bytecodeRawI32Value(vm.slots[instr.target])
	if !ok {
		return fmt.Errorf("bytecode i32 slot load expected i32 value")
	}
	vm.pushI32(value)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execStoreSlotI32(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode i32 slot store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	if vm.i32UnboxFallbackSet {
		fallback := vm.i32UnboxFallbackValue
		vm.i32UnboxFallbackValue = nil
		vm.i32UnboxFallbackSet = false
		if !instr.discardResult {
			if fallback == nil {
				fallback = runtime.NilValue{}
			}
			vm.appendStackValue(fallback)
		}
		vm.ip++
		return nil
	}
	raw, err := vm.popI32()
	if err != nil {
		return err
	}
	if instr.discardResult {
		if vm.setI32RegisterRaw(instr.target, raw) {
			vm.clearActiveValueSlotI32(instr.target)
			vm.clearActiveValueSlotFloat(instr.target)
			vm.slots[instr.target] = nil
		} else {
			vm.clearActiveValueSlotI32(instr.target)
			vm.clearActiveValueSlotFloat(instr.target)
			vm.slots[instr.target] = bytecodeRawI32SlotCachedValue(raw)
		}
		if instr.target == 0 {
			vm.setSelfFastSlot0I32Raw(raw)
		}
		vm.ip++
		return nil
	}
	value := bytecodeBoxedIntegerI32Value(int64(raw))
	vm.clearActiveValueSlotI32(instr.target)
	vm.clearActiveValueSlotFloat(instr.target)
	vm.slots[instr.target] = value
	vm.setI32RegisterRaw(instr.target, raw)
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Raw(raw)
	}
	vm.appendStackValue(value)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execCompoundAssignSlotI32(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode i32 compound assignment missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	right, err := vm.popI32()
	if err != nil {
		return err
	}
	left, ok := vm.i32RegisterRaw(instr.target)
	if !ok {
		left, ok = vm.activeValueSlotI32Raw(instr.target)
	}
	if !ok {
		left, ok = bytecodeRawI32Value(vm.slots[instr.target])
	}
	if !ok {
		return fmt.Errorf("bytecode i32 compound assignment expected i32 slot value")
	}
	assignmentOp := ast.AssignmentOperator(instr.operator)
	binaryOp, isCompound := binaryOpForAssignment(assignmentOp)
	if !isCompound {
		return fmt.Errorf("unsupported assignment operator %s", assignmentOp)
	}
	var result int64
	switch binaryOp {
	case "+":
		result = int64(left) + int64(right)
	case "-":
		result = int64(left) - int64(right)
	default:
		return fmt.Errorf("bytecode i32 compound assignment unsupported operator %q", binaryOp)
	}
	if result < math.MinInt32 || result > math.MaxInt32 {
		err := newOverflowError("integer overflow")
		if vm.interp != nil {
			err = vm.interp.wrapStandardRuntimeError(err)
			if instr.node != nil {
				return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			}
		}
		return err
	}
	raw := int32(result)
	if instr.discardResult {
		if vm.setI32RegisterRaw(instr.target, raw) {
			vm.clearActiveValueSlotI32(instr.target)
			vm.clearActiveValueSlotFloat(instr.target)
			vm.slots[instr.target] = nil
		} else {
			vm.clearActiveValueSlotI32(instr.target)
			vm.clearActiveValueSlotFloat(instr.target)
			vm.slots[instr.target] = bytecodeRawI32SlotCachedValue(raw)
		}
		if instr.target == 0 {
			vm.setSelfFastSlot0I32Raw(raw)
		}
		vm.ip++
		return nil
	}
	value := bytecodeBoxedIntegerI32Value(int64(raw))
	vm.clearActiveValueSlotI32(instr.target)
	vm.clearActiveValueSlotFloat(instr.target)
	vm.slots[instr.target] = value
	vm.setI32RegisterRaw(instr.target, raw)
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Raw(raw)
	}
	vm.appendStackValue(value)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execBinaryI32Opcode(instr *bytecodeInstruction) (bool, error) {
	if err := vm.execBinaryI32(instr); err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr != nil && instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			if vm.handleLoopSignal(err) {
				return true, nil
			}
		}
		return false, err
	}
	vm.ip++
	return false, nil
}

func (vm *bytecodeVM) execBinaryI32(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode i32 binary missing instruction")
	}
	right, err := vm.popI32()
	if err != nil {
		return err
	}
	left, err := vm.popI32()
	if err != nil {
		return err
	}
	var result int64
	switch instr.op {
	case bytecodeOpBinaryI32Add:
		result = int64(left) + int64(right)
	case bytecodeOpBinaryI32Sub:
		result = int64(left) - int64(right)
	default:
		return fmt.Errorf("bytecode i32 binary opcode %d unsupported", instr.op)
	}
	if result < math.MinInt32 || result > math.MaxInt32 {
		return newOverflowError("integer overflow")
	}
	vm.pushI32(int32(result))
	return nil
}
