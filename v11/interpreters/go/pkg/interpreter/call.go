package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/runtime"
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
