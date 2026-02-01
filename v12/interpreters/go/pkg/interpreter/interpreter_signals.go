package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type breakSignal struct {
	label string
	value runtime.Value
}

func (b breakSignal) Error() string {
	if b.label != "" {
		return fmt.Sprintf("break %s", b.label)
	}
	return "break"
}

type continueSignal struct {
	label string
}

func (c continueSignal) Error() string {
	if c.label != "" {
		return fmt.Sprintf("continue %s", c.label)
	}
	return "continue"
}

type raiseSignal struct {
	value   runtime.Value
	context *runtimeDiagnosticContext
}

func (r raiseSignal) Error() string {
	if errVal, ok := r.value.(runtime.ErrorValue); ok {
		return errVal.Message
	}
	return valueToString(r.value)
}

type returnSignal struct {
	value   runtime.Value
	context *runtimeDiagnosticContext
}

func (r returnSignal) Error() string {
	return "return"
}

func (i *Interpreter) makeErrorValue(val runtime.Value, env *runtime.Environment) runtime.ErrorValue {
	if errVal, ok := asErrorValue(val); ok {
		return errVal
	}
	message := valueToString(val)
	payload := map[string]runtime.Value{
		"value": val,
	}
	if i != nil {
		if ifaceVal, err := i.coerceToInterfaceValue("Error", val, nil); err == nil {
			callEnv := env
			if callEnv == nil {
				callEnv = i.global
			}
			if member, err := i.memberAccessOnValue(ifaceVal, ast.NewIdentifier("message"), callEnv); err == nil {
				if msgVal, err := i.callCallableValue(member, nil, callEnv, nil); err == nil {
					if msgStr, ok := msgVal.(runtime.StringValue); ok {
						message = msgStr.Val
					}
				}
			}
			if member, err := i.memberAccessOnValue(ifaceVal, ast.NewIdentifier("cause"), callEnv); err == nil {
				if causeVal, err := i.callCallableValue(member, nil, callEnv, nil); err == nil && !isNilRuntimeValue(causeVal) {
					payload["cause"] = causeVal
				}
			}
		}
	}
	return runtime.ErrorValue{Message: message, Payload: payload}
}

// MakeErrorValue exposes error value construction for compiler interop helpers.
func (i *Interpreter) MakeErrorValue(val runtime.Value, env *runtime.Environment) runtime.ErrorValue {
	return i.makeErrorValue(val, env)
}
