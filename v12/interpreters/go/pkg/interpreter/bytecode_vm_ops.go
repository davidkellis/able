package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execBinary(instr bytecodeInstruction) (bool, error) {
	right, err := vm.pop()
	if err != nil {
		return false, err
	}
	left, err := vm.pop()
	if err != nil {
		return false, err
	}
	if instr.operator == "+" {
		rawLeft := unwrapInterfaceValue(left)
		rawRight := unwrapInterfaceValue(right)
		if ls, ok := rawLeft.(runtime.StringValue); ok {
			rs, ok := rawRight.(runtime.StringValue)
			if !ok {
				err := fmt.Errorf("Arithmetic requires numeric operands")
				if instr.node != nil {
					err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
				}
				return false, err
			}
			vm.stack = append(vm.stack, runtime.StringValue{Val: ls.Val + rs.Val})
			vm.ip++
			return false, nil
		}
		if _, ok := rawRight.(runtime.StringValue); ok {
			err := fmt.Errorf("Arithmetic requires numeric operands")
			if instr.node != nil {
				err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			}
			return false, err
		}
	}
	result, err := applyBinaryOperator(vm.interp, instr.operator, left, right)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			if vm.handleLoopSignal(err) {
				return true, nil
			}
		}
		return false, err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return false, nil
}
