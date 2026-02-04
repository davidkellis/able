package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// CallFunction invokes an Able function value with the provided arguments.
func (i *Interpreter) CallFunction(value runtime.Value, args []runtime.Value) (runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("interpreter: nil interpreter")
	}
	if value == nil {
		return nil, fmt.Errorf("interpreter: cannot call <nil> value")
	}
	return i.callCallableValue(value, args, nil, nil)
}

// CallFunctionIn invokes an Able function value with the provided arguments and environment.
func (i *Interpreter) CallFunctionIn(value runtime.Value, args []runtime.Value, env *runtime.Environment) (runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("interpreter: nil interpreter")
	}
	if value == nil {
		return nil, fmt.Errorf("interpreter: cannot call <nil> value")
	}
	return i.callCallableValue(value, args, env, nil)
}

// CallFunctionInWithCallNode invokes an Able function value with a call node for diagnostics.
func (i *Interpreter) CallFunctionInWithCallNode(value runtime.Value, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall) (runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("interpreter: nil interpreter")
	}
	if value == nil {
		return nil, fmt.Errorf("interpreter: cannot call <nil> value")
	}
	return i.callCallableValue(value, args, env, call)
}

// PushCallFrame records a call expression in the current eval state for diagnostics.
func (i *Interpreter) PushCallFrame(env *runtime.Environment, call *ast.FunctionCall) {
	if i == nil {
		return
	}
	state := i.stateFromEnv(env)
	state.pushCallFrame(call)
}

// PopCallFrame removes the most recent call frame from the current eval state.
func (i *Interpreter) PopCallFrame(env *runtime.Environment) {
	if i == nil {
		return
	}
	state := i.stateFromEnv(env)
	state.popCallFrame()
}
