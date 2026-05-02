package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execJumpIfFalse(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode jump-if-false missing instruction")
	}
	cond, err := vm.pop()
	if err != nil {
		return err
	}
	if !vm.interp.isTruthy(cond) {
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execJumpIfBoolSlotFalse(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode bool-slot jump-if-false missing instruction")
	}
	slot := instr.argCount
	if slot < 0 || slot >= len(vm.slots) {
		return fmt.Errorf("bytecode bool slot out of range")
	}
	if cond, ok := vm.slots[slot].(runtime.BoolValue); ok {
		if !cond.Val {
			vm.ip = instr.target
			return nil
		}
		vm.ip++
		return nil
	}
	if !vm.interp.isTruthy(vm.slots[slot]) {
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}
