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

	switch fn := value.(type) {
	case *runtime.FunctionValue:
		return i.invokeFunction(fn, args, nil)
	case *runtime.BoundMethodValue:
		injected := append([]runtime.Value{fn.Receiver}, args...)
		return i.invokeFunction(fn.Method, injected, nil)
	case runtime.BoundMethodValue:
		injected := append([]runtime.Value{fn.Receiver}, args...)
		return i.invokeFunction(fn.Method, injected, nil)
	default:
		return nil, fmt.Errorf("interpreter: value of kind %s is not callable", value.Kind())
	}
}
