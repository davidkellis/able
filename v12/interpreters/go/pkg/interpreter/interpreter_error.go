package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) initErrorBuiltins() {
	i.errorNativeMethods["message"] = runtime.NativeFunctionValue{
		Name:  "Error.message",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Error.message expects only a receiver")
			}
			errVal, ok := asErrorValue(args[0])
			if !ok {
				return nil, fmt.Errorf("Error.message receiver must be an error value")
			}
			return runtime.StringValue{Val: errVal.Message}, nil
		},
	}
	i.errorNativeMethods["cause"] = runtime.NativeFunctionValue{
		Name:  "Error.cause",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Error.cause expects only a receiver")
			}
			errVal, ok := asErrorValue(args[0])
			if !ok {
				return nil, fmt.Errorf("Error.cause receiver must be an error value")
			}
			if errVal.Payload != nil {
				if cause, ok := errVal.Payload["cause"]; ok && cause != nil {
					return cause, nil
				}
			}
			return runtime.NilValue{}, nil
		},
	}
}

func (i *Interpreter) errorMember(err runtime.ErrorValue, member ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Error member access expects identifier")
	}
	if ident.Name == "value" {
		if err.Payload != nil {
			if val, ok := err.Payload["value"]; ok && val != nil {
				return val, nil
			}
		}
		if err.TypeName != nil && err.TypeName.Name != "" {
			return i.errorValueToStructInstance(err), nil
		}
		return runtime.NilValue{}, nil
	}
	if method, ok := i.errorNativeMethods[ident.Name]; ok {
		return &runtime.NativeBoundMethodValue{Receiver: err, Method: method}, nil
	}
	if bound, err := i.resolveMethodFromPool(env, ident.Name, err, ""); err != nil {
		return nil, err
	} else if bound != nil {
		return bound, nil
	}
	return nil, fmt.Errorf("Error value has no member '%s'", ident.Name)
}

func asErrorValue(val runtime.Value) (runtime.ErrorValue, bool) {
	switch v := val.(type) {
	case runtime.ErrorValue:
		return v, true
	case *runtime.ErrorValue:
		if v == nil {
			return runtime.ErrorValue{}, false
		}
		return *v, true
	default:
		return runtime.ErrorValue{}, false
	}
}

// IsErrorValue reports whether a value is an Error or implements the Error interface.
func (i *Interpreter) IsErrorValue(val runtime.Value) bool {
	return i.matchesErrorValue(val)
}

// matchesErrorValue recognizes the language's Error protocol independently of
// whether it is the bootstrap definition or the canonical stdlib definition.
// The bootstrap is available before imports are evaluated, while the stdlib
// definition has a package-qualified identity. Both describe the same
// language-level Error protocol used by Result values and propagation.
func (i *Interpreter) matchesErrorValue(val runtime.Value) bool {
	if val == nil {
		return false
	}
	if _, ok := asErrorValue(val); ok {
		return true
	}
	if i == nil {
		return false
	}
	switch value := val.(type) {
	case runtime.InterfaceValue:
		if value.Interface != nil && value.Interface.Node != nil && value.Interface.Node.ID != nil && value.Interface.Node.ID.Name == "Error" {
			return true
		}
	case *runtime.InterfaceValue:
		if value != nil && value.Interface != nil && value.Interface.Node != nil && value.Interface.Node.ID != nil && value.Interface.Node.ID.Name == "Error" {
			return true
		}
	}
	info, ok := i.getTypeInfoForValue(val)
	if !ok {
		return false
	}
	for _, interfaceName := range i.errorInterfaceNames() {
		implements, err := i.typeImplementsInterface(info, interfaceName, nil, make(map[interfaceImplCacheKey]struct{}))
		if err == nil && implements {
			return true
		}
	}
	return false
}

func (i *Interpreter) errorInterfaceNames() []string {
	names := []string{"Error"}
	if i == nil {
		return names
	}
	const canonicalError = "able.core.interfaces.Error"
	if _, ok := i.interfaces[canonicalError]; ok {
		names = append(names, canonicalError)
	}
	return names
}

func (i *Interpreter) coerceToErrorInterfaceValue(value runtime.Value) (runtime.Value, error) {
	var lastErr error
	for _, interfaceName := range i.errorInterfaceNames() {
		coerced, err := i.coerceToInterfaceValue(interfaceName, value, nil)
		if err == nil {
			return coerced, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("Error interface is not defined")
	}
	return nil, lastErr
}
