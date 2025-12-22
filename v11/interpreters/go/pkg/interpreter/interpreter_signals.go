package interpreter

import (
	"fmt"

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
	value runtime.Value
}

func (r raiseSignal) Error() string {
	if errVal, ok := r.value.(runtime.ErrorValue); ok {
		return errVal.Message
	}
	return valueToString(r.value)
}

type returnSignal struct {
	value runtime.Value
}

func (r returnSignal) Error() string {
	return "return"
}

func makeErrorValue(val runtime.Value) runtime.ErrorValue {
	if errVal, ok := val.(runtime.ErrorValue); ok {
		return errVal
	}
	message := valueToString(val)
	payload := map[string]runtime.Value{
		"value": val,
	}
	return runtime.ErrorValue{Message: message, Payload: payload}
}
