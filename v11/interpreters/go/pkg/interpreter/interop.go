package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

// CoerceArrayValue exposes array conversion for external tooling (CLI/reporters).
func (i *Interpreter) CoerceArrayValue(val runtime.Value) (*runtime.ArrayValue, error) {
	if i == nil {
		return nil, fmt.Errorf("interpreter: nil interpreter")
	}
	if val == nil {
		return nil, fmt.Errorf("array value is nil")
	}
	return i.toArrayValue(val)
}

// LookupStructMethod returns the method definition for a struct instance, if any.
func (i *Interpreter) LookupStructMethod(val runtime.Value, name string) (runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("interpreter: nil interpreter")
	}
	if name == "" || val == nil {
		return nil, nil
	}
	switch v := val.(type) {
	case *runtime.StructInstanceValue:
		return i.selectStructMethod(v, name)
	default:
		return nil, nil
	}
}
