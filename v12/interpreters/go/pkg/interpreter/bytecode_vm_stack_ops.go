package interpreter

import "fmt"

func (vm *bytecodeVM) execDup() error {
	if len(vm.stack) == 0 {
		return fmt.Errorf("bytecode stack underflow")
	}
	vm.stack = append(vm.stack, vm.stack[len(vm.stack)-1])
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execPop() error {
	if _, err := vm.pop(); err != nil {
		return err
	}
	vm.ip++
	return nil
}
