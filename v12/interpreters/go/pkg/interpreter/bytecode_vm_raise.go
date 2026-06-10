package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (vm *bytecodeVM) execRaise(instr *bytecodeInstruction) error {
	raiseStmt, ok := instr.node.(*ast.RaiseStatement)
	if !ok || raiseStmt == nil {
		return fmt.Errorf("bytecode raise expects node")
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	raiseErr := raiseSignal{value: vm.interp.makeErrorValue(val, vm.env)}
	return vm.interp.attachRuntimeContext(raiseErr, raiseStmt, vm.interp.stateFromEnv(vm.env))
}
