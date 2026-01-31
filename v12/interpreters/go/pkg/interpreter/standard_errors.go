package interpreter

import (
	"errors"
	"fmt"
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type standardRuntimeErrorKind string

const (
	standardDivisionByZero  standardRuntimeErrorKind = "DivisionByZeroError"
	standardOverflow        standardRuntimeErrorKind = "OverflowError"
	standardShiftOutOfRange standardRuntimeErrorKind = "ShiftOutOfRangeError"
)

type standardRuntimeError struct {
	kind      standardRuntimeErrorKind
	message   string
	operation string
	shift     int64
}

const (
	minI32 = int64(-1 << 31)
	maxI32 = int64(1<<31 - 1)
)

func (e standardRuntimeError) Error() string {
	return e.message
}

func newDivisionByZeroError() error {
	return standardRuntimeError{kind: standardDivisionByZero, message: "division by zero"}
}

func (i *Interpreter) StandardDivisionByZeroErrorValue() runtime.ErrorValue {
	return i.makeStandardErrorValue(standardRuntimeError{kind: standardDivisionByZero, message: "division by zero"})
}

func (i *Interpreter) StandardOverflowErrorValue(operation string) runtime.ErrorValue {
	message := operation
	if message == "" {
		message = "integer overflow"
	}
	return i.makeStandardErrorValue(standardRuntimeError{kind: standardOverflow, message: message, operation: operation})
}

func (i *Interpreter) StandardShiftOutOfRangeErrorValue(shift int64) runtime.ErrorValue {
	return i.makeStandardErrorValue(standardRuntimeError{kind: standardShiftOutOfRange, message: "shift out of range", shift: shift})
}

func newOverflowError(operation string) error {
	message := operation
	if message == "" {
		message = "integer overflow"
	}
	return standardRuntimeError{kind: standardOverflow, message: message, operation: operation}
}

func newShiftOutOfRangeError(shift int64) error {
	return standardRuntimeError{
		kind:    standardShiftOutOfRange,
		message: "shift out of range",
		shift:   shift,
	}
}

func (i *Interpreter) resolveStandardErrorStruct(name string) *runtime.StructDefinitionValue {
	if def, ok := i.standardErrorStructs[name]; ok && def != nil {
		return def
	}
	candidates := []string{
		name,
		"able.core.errors." + name,
		"core.errors." + name,
		"errors." + name,
	}
	for _, key := range candidates {
		if val, err := i.global.Get(key); err == nil {
			if def, conv := toStructDefinitionValue(val, key); conv == nil {
				i.standardErrorStructs[name] = def
				return def
			}
		}
	}
	for _, bucket := range i.packageRegistry {
		if val, ok := bucket[name]; ok {
			if def, conv := toStructDefinitionValue(val, name); conv == nil {
				i.standardErrorStructs[name] = def
				return def
			}
		}
	}
	placeholder := ast.NewStructDefinition(ast.NewIdentifier(name), nil, ast.StructKindNamed, nil, nil, false)
	def := &runtime.StructDefinitionValue{Node: placeholder}
	i.standardErrorStructs[name] = def
	return def
}

func (i *Interpreter) makeStandardErrorValue(err standardRuntimeError) runtime.ErrorValue {
	def := i.resolveStandardErrorStruct(string(err.kind))
	fields := map[string]runtime.Value{}
	switch err.kind {
	case standardOverflow:
		operation := err.operation
		if operation == "" {
			operation = err.message
		}
		fields["operation"] = runtime.StringValue{Val: operation}
	case standardShiftOutOfRange:
		shift := err.shift
		if shift < minI32 || shift > maxI32 {
			shift = 0
		}
		fields["shift"] = runtime.IntegerValue{Val: big.NewInt(shift), TypeSuffix: runtime.IntegerI32}
	}
	instance := &runtime.StructInstanceValue{
		Definition: def,
		Fields:     fields,
	}
	payload := map[string]runtime.Value{
		"value": instance,
	}
	return runtime.ErrorValue{
		Message: err.message,
		Payload: payload,
	}
}

func (i *Interpreter) makeIndexErrorValue(index int, length int) runtime.ErrorValue {
	def := i.resolveStandardErrorStruct("IndexError")
	fields := map[string]runtime.Value{
		"index": runtime.IntegerValue{
			Val:        big.NewInt(int64(index)),
			TypeSuffix: runtime.IntegerI64,
		},
		"length": runtime.IntegerValue{
			Val:        big.NewInt(int64(length)),
			TypeSuffix: runtime.IntegerI64,
		},
	}
	instance := &runtime.StructInstanceValue{
		Definition: def,
		Fields:     fields,
	}
	message := fmt.Sprintf("index %d out of bounds for length %d", index, length)
	return runtime.ErrorValue{
		Message: message,
		Payload: map[string]runtime.Value{"value": instance},
	}
}

func (i *Interpreter) wrapStandardRuntimeError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(raiseSignal); ok {
		return err
	}
	var standardErr standardRuntimeError
	if ok := errors.As(err, &standardErr); ok {
		return raiseSignal{value: i.makeStandardErrorValue(standardErr)}
	}
	return err
}
