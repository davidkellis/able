package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execPlaceholderLambda(instr *bytecodeInstruction) error {
	expr, ok := instr.node.(ast.Expression)
	if !ok || expr == nil {
		return fmt.Errorf("bytecode placeholder lambda expects expression node")
	}
	program, err := vm.interp.lowerExpressionToBytecodeWithOptions(expr, false)
	if err != nil {
		return err
	}
	state := vm.interp.stateFromEnv(vm.env)
	if state.hasPlaceholderFrame() {
		innerVM := vm.interp.acquireBytecodeVM(vm.env)
		val, err := innerVM.run(program)
		vm.interp.releaseBytecodeVM(innerVM)
		if err != nil {
			return err
		}
		if val == nil {
			val = runtime.NilValue{}
		}
		vm.stack = append(vm.stack, val)
		vm.ip++
		return nil
	}
	if instr.argCount <= 0 {
		return fmt.Errorf("bytecode placeholder lambda missing arity")
	}
	closure := &placeholderClosure{
		interpreter: vm.interp,
		expression:  expr,
		env:         vm.env,
		plan:        placeholderPlan{paramCount: instr.argCount},
		bytecode:    program,
	}
	fn := runtime.NativeFunctionValue{
		Name:  "<placeholder>",
		Arity: instr.argCount,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return closure.invoke(args)
		},
	}
	vm.stack = append(vm.stack, fn)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execPlaceholderValue(instr *bytecodeInstruction) error {
	placeholderExpr, ok := instr.node.(*ast.PlaceholderExpression)
	if !ok || placeholderExpr == nil {
		return fmt.Errorf("bytecode placeholder value expects placeholder expression")
	}
	val, err := vm.interp.evaluatePlaceholderExpression(placeholderExpr, vm.env)
	if err != nil {
		return err
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return nil
}
