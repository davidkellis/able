package main

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

func TestFormatRuntimeValueFormatsCallableValuesAsFunction(t *testing.T) {
	values := []runtime.Value{
		&runtime.FunctionValue{},
		&runtime.FunctionOverloadValue{},
		runtime.NativeFunctionValue{},
		&runtime.NativeFunctionValue{},
		runtime.BoundMethodValue{},
		&runtime.BoundMethodValue{},
		runtime.NativeBoundMethodValue{},
		&runtime.NativeBoundMethodValue{},
		runtime.PartialFunctionValue{},
		&runtime.PartialFunctionValue{},
	}

	for _, value := range values {
		if got := formatRuntimeValue(nil, value); got != "<function>" {
			t.Fatalf("formatRuntimeValue(%T) = %q, want %q", value, got, "<function>")
		}
	}
}

func TestFormatRuntimeValueFormatsInteger(t *testing.T) {
	value := runtime.IntegerValue{Val: big.NewInt(42), TypeSuffix: runtime.IntegerI32}
	if got := formatRuntimeValue(nil, value); got != "42" {
		t.Fatalf("formatRuntimeValue(%T) = %q, want %q", value, got, "42")
	}
}

func TestFormatRuntimeValueFormatsStdlibStringStruct(t *testing.T) {
	interp := interpreter.New()
	byteArray := &runtime.ArrayValue{Elements: []runtime.Value{
		runtime.NewSmallInt(72, runtime.IntegerU8),
		runtime.NewSmallInt(105, runtime.IntegerU8),
	}}
	value := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef("String", nil, ast.StructKindNamed, nil, nil, false),
		},
		Fields: map[string]runtime.Value{
			"bytes":     byteArray,
			"len_bytes": runtime.NewSmallInt(2, runtime.IntegerI32),
		},
	}

	if got := formatRuntimeValue(interp, value); got != "Hi" {
		t.Fatalf("formatRuntimeValue(stdlib String) = %q, want %q", got, "Hi")
	}
}
