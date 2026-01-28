package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execEnsureStart(instr bytecodeInstruction) error {
	ensureExpr, ok := instr.node.(*ast.EnsureExpression)
	if !ok || ensureExpr == nil {
		return fmt.Errorf("bytecode ensure expects node")
	}
	val, err := vm.evalExpressionWithFallback(ensureExpr.TryExpression, vm.env)
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.ensureStack = append(vm.ensureStack, bytecodeEnsureFrame{result: val, err: err})
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execEnsureEnd(_ bytecodeInstruction) error {
	if len(vm.ensureStack) == 0 {
		return fmt.Errorf("bytecode ensure stack underflow")
	}
	if _, err := vm.pop(); err != nil {
		return err
	}
	frame := vm.ensureStack[len(vm.ensureStack)-1]
	vm.ensureStack = vm.ensureStack[:len(vm.ensureStack)-1]
	if frame.err != nil {
		return frame.err
	}
	result := frame.result
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return nil
}
