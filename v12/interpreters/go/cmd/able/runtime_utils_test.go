package main

import (
	"math/big"
	"testing"

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
