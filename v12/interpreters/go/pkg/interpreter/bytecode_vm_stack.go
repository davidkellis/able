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
