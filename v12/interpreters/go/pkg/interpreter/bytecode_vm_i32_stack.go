package interpreter

import (
	"fmt"
	"math"

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

func (vm *bytecodeVM) execBoxI32() error {
	value, err := vm.popBoxedI32()
	if err != nil {
		return err
	}
	vm.stack = append(vm.stack, value)
	vm.ip++
	return nil
}

func bytecodeRawI32Value(value runtime.Value) (int32, bool) {
	var intVal runtime.IntegerValue
	switch v := value.(type) {
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
	raw, err := vm.popI32()
	if err != nil {
		return err
	}
	value := bytecodeBoxedIntegerI32Value(int64(raw))
	vm.slots[instr.target] = value
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Raw(raw)
	}
	vm.stack = append(vm.stack, value)
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
