package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
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
