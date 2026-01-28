package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) pushLoopFrame(breakTarget int, continueTarget int) error {
	if breakTarget < 0 || continueTarget < 0 {
		return fmt.Errorf("bytecode loop targets missing")
	}
	vm.loopStack = append(vm.loopStack, bytecodeLoopFrame{
		breakTarget:    breakTarget,
		continueTarget: continueTarget,
		env:            vm.env,
	})
	return nil
}

func (vm *bytecodeVM) popLoopFrame() error {
	if len(vm.loopStack) == 0 {
		return fmt.Errorf("bytecode loop stack underflow")
	}
	vm.loopStack = vm.loopStack[:len(vm.loopStack)-1]
	return nil
}

func (vm *bytecodeVM) handleLoopSignal(err error) bool {
	if err == nil || len(vm.loopStack) == 0 {
		return false
	}
	frame := vm.loopStack[len(vm.loopStack)-1]
	switch sig := err.(type) {
	case breakSignal:
		if sig.label != "" {
			return false
		}
		vm.env = frame.env
		val := sig.value
		if val == nil {
			val = runtime.NilValue{}
		}
		vm.stack = append(vm.stack, val)
		vm.ip = frame.breakTarget
		return true
	case continueSignal:
		if sig.label != "" {
			return false
		}
		vm.env = frame.env
		vm.ip = frame.continueTarget
		return true
	default:
		return false
	}
}
