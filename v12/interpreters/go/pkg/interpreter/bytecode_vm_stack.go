package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) pop() (runtime.Value, error) {
	if len(vm.stack) == 0 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	last := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return last, nil
}

func bytecodeStackResultValue(result runtime.Value) runtime.Value {
	if result == nil {
		return runtime.NilValue{}
	}
	return result
}

func (vm *bytecodeVM) replaceTop1(result runtime.Value) error {
	if len(vm.stack) < 1 {
		return fmt.Errorf("bytecode stack underflow")
	}
	vm.replaceTop1Unchecked(result)
	return nil
}

func (vm *bytecodeVM) replaceTop2(result runtime.Value) error {
	if len(vm.stack) < 2 {
		return fmt.Errorf("bytecode stack underflow")
	}
	vm.replaceTop2Unchecked(result)
	return nil
}

func (vm *bytecodeVM) replaceTop3(result runtime.Value) error {
	if len(vm.stack) < 3 {
		return fmt.Errorf("bytecode stack underflow")
	}
	vm.replaceTop3Unchecked(result)
	return nil
}

func (vm *bytecodeVM) replaceTop1Unchecked(result runtime.Value) {
	vm.stack[len(vm.stack)-1] = bytecodeStackResultValue(result)
}

func (vm *bytecodeVM) replaceTop2Unchecked(result runtime.Value) {
	idx := len(vm.stack) - 2
	vm.stack[idx] = bytecodeStackResultValue(result)
	vm.stack = vm.stack[:idx+1]
}

func (vm *bytecodeVM) replaceTop3Unchecked(result runtime.Value) {
	idx := len(vm.stack) - 3
	vm.stack[idx] = bytecodeStackResultValue(result)
	vm.stack = vm.stack[:idx+1]
}
