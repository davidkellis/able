package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) stackDepth() int {
	return len(vm.stack)
}

func (vm *bytecodeVM) stackCapacity() int {
	return cap(vm.stack)
}

func (vm *bytecodeVM) stackValue(index int) runtime.Value {
	return vm.stack[index]
}

func (vm *bytecodeVM) stackValues(start int, end int) []runtime.Value {
	return vm.stack[start:end]
}

func (vm *bytecodeVM) stackValuesFrom(start int) []runtime.Value {
	return vm.stack[start:]
}

func (vm *bytecodeVM) appendStackValue(value runtime.Value) {
	vm.stack = append(vm.stack, value)
}

func (vm *bytecodeVM) appendStackPair(first runtime.Value, second runtime.Value) {
	vm.stack = append(vm.stack, first, second)
}

func (vm *bytecodeVM) setStackValue(index int, value runtime.Value) {
	vm.stack[index] = value
}

func (vm *bytecodeVM) truncateStack(depth int) {
	vm.stack = vm.stack[:depth]
}

func (vm *bytecodeVM) clearStackFrom(start int) {
	clear(vm.stack[start:])
}

func (vm *bytecodeVM) pop() (runtime.Value, error) {
	if vm.stackDepth() == 0 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	lastIndex := vm.stackDepth() - 1
	last := vm.stackValue(lastIndex)
	vm.truncateStack(lastIndex)
	return last, nil
}

func bytecodeStackResultValue(result runtime.Value) runtime.Value {
	if result == nil {
		return runtime.NilValue{}
	}
	return result
}

func (vm *bytecodeVM) replaceTop1(result runtime.Value) error {
	if vm.stackDepth() < 1 {
		return fmt.Errorf("bytecode stack underflow")
	}
	vm.replaceTop1Unchecked(result)
	return nil
}

func (vm *bytecodeVM) replaceTop2(result runtime.Value) error {
	if vm.stackDepth() < 2 {
		return fmt.Errorf("bytecode stack underflow")
	}
	vm.replaceTop2Unchecked(result)
	return nil
}

func (vm *bytecodeVM) replaceTop3(result runtime.Value) error {
	if vm.stackDepth() < 3 {
		return fmt.Errorf("bytecode stack underflow")
	}
	vm.replaceTop3Unchecked(result)
	return nil
}

func (vm *bytecodeVM) replaceTop1Unchecked(result runtime.Value) {
	vm.setStackValue(vm.stackDepth()-1, bytecodeStackResultValue(result))
}

func (vm *bytecodeVM) replaceTop2Unchecked(result runtime.Value) {
	idx := vm.stackDepth() - 2
	vm.setStackValue(idx, bytecodeStackResultValue(result))
	vm.truncateStack(idx + 1)
}

func (vm *bytecodeVM) replaceTop2RawFloatUnchecked(value float64, kind runtime.FloatType) {
	idx := vm.stackDepth() - 2
	bytecodeSetNormalizedRawFloatValue(&vm.stack[idx], value, kind)
	vm.truncateStack(idx + 1)
}

func (vm *bytecodeVM) replaceTop3Unchecked(result runtime.Value) {
	idx := vm.stackDepth() - 3
	vm.setStackValue(idx, bytecodeStackResultValue(result))
	vm.truncateStack(idx + 1)
}
