package interpreter

import "able/interpreter-go/pkg/runtime"

// Raise converts a value into a runtime raise signal for native/compiled interop.
func Raise(interp *Interpreter, value runtime.Value, env *runtime.Environment) error {
	if interp == nil {
		return raiseSignal{value: value}
	}
	return raiseSignal{value: interp.makeErrorValue(value, env)}
}

// RaiseValue implements the compiler bridge's dynamic-runtime contract.
func (i *Interpreter) RaiseValue(value runtime.Value, env *runtime.Environment) error {
	return Raise(i, value, env)
}
