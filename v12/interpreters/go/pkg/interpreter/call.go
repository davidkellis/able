package interpreter

import (
	"fmt"

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
