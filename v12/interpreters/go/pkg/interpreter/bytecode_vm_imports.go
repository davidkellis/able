package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execImport(instr bytecodeInstruction) error {
	imp, ok := instr.node.(*ast.ImportStatement)
	if !ok || imp == nil {
		return fmt.Errorf("bytecode import expects import statement")
	}
	val, err := vm.interp.evaluateImportStatement(imp, vm.env)
	if err != nil {
		return vm.interp.attachRuntimeContext(err, imp, vm.interp.stateFromEnv(vm.env))
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execDynImport(instr bytecodeInstruction) error {
	imp, ok := instr.node.(*ast.DynImportStatement)
	if !ok || imp == nil {
		return fmt.Errorf("bytecode dynimport expects dynimport statement")
	}
	val, err := vm.interp.evaluateDynImportStatement(imp, vm.env)
	if err != nil {
		return vm.interp.attachRuntimeContext(err, imp, vm.interp.stateFromEnv(vm.env))
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return nil
}
