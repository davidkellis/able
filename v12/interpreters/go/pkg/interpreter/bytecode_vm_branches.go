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

func (vm *bytecodeVM) execJumpIfBinaryCompareFalse(instr *bytecodeInstruction) (bool, error) {
	if instr == nil {
		return false, fmt.Errorf("bytecode binary compare jump missing instruction")
	}
	if vm.stackDepth() < 2 {
		return false, fmt.Errorf("bytecode stack underflow")
	}
	rightIdx := vm.stackDepth() - 1
	leftIdx := rightIdx - 1
	left := vm.stackValue(leftIdx)
	right := vm.stackValue(rightIdx)
	var comparison runtime.BoolValue
	if value, ok := execBinaryDirectIntegerComparisonFast(instr.operator, left, right); ok {
		var resultOK bool
		comparison, resultOK = value.(runtime.BoolValue)
		if !resultOK {
			return false, nil
		}
	} else if value, ok := bytecodeDirectFloatCompareFast(instr.operator, left, right); ok {
		comparison = value
	} else {
		return false, nil
	}
	vm.truncateStack(leftIdx)
	if comparison.Val {
		vm.ip += 2
	} else {
		vm.ip = instr.target
	}
	return true, nil
}

func (vm *bytecodeVM) execJumpIfBinaryCompareFalseOpcode(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable) (bool, error) {
	if handled, err := vm.execJumpIfBinaryCompareFalse(instr); handled || err != nil {
		return handled, err
	}
	fallback := *instr
	fallback.op = bytecodeOpBinary
	return vm.execBinary(&fallback, slotConstIntImmTable)
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
