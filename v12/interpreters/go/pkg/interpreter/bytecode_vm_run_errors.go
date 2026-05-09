package interpreter

import "able/interpreter-go/pkg/ast"

func (vm *bytecodeVM) handleBytecodeJumpRuntimeError(err error, node ast.Node) (bool, error) {
	if err == nil {
		return false, nil
	}
	err = vm.interp.wrapStandardRuntimeError(err)
	if node != nil {
		err = vm.interp.attachRuntimeContext(err, node, vm.interp.stateFromEnv(vm.env))
		if vm.handleLoopSignal(err) {
			return true, nil
		}
	}
	return false, err
}
