package interpreter

import "fmt"

func (vm *bytecodeVM) execDup() error {
	if vm.stackDepth() == 0 {
		return fmt.Errorf("bytecode stack underflow")
	}
	vm.appendStackValue(vm.stackValue(vm.stackDepth() - 1))
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
