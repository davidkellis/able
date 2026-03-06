package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestFunctionScopeFilterSingle(t *testing.T) {
	first := &runtime.FunctionValue{}
	second := &runtime.FunctionValue{}

	filter := functionScopeFilterFromValue(first)
	if !filter.enabled {
		t.Fatalf("expected single-function filter to be enabled")
	}
	if !filter.contains(first) {
		t.Fatalf("expected filter to contain the single scope function")
	}
	if filter.contains(second) {
		t.Fatalf("expected filter to reject non-scope functions")
	}
}

func TestFunctionScopeFilterOverloads(t *testing.T) {
	first := &runtime.FunctionValue{}
	second := &runtime.FunctionValue{}
	other := &runtime.FunctionValue{}

	filter := functionScopeFilterFromValue(&runtime.FunctionOverloadValue{
		Overloads: []*runtime.FunctionValue{first, second},
	})
	if !filter.enabled {
		t.Fatalf("expected overload filter to be enabled")
	}
	if !filter.contains(first) || !filter.contains(second) {
		t.Fatalf("expected overload filter to contain both overload entries")
	}
	if filter.contains(other) {
		t.Fatalf("expected overload filter to reject non-overload functions")
	}
}

func TestFunctionScopeFilterDisabledPassThrough(t *testing.T) {
	filter := functionScopeFilterFromValue(runtime.StringValue{Val: "not callable"})
	if filter.enabled {
		t.Fatalf("expected non-function scope filter to be disabled")
	}
	if !filter.contains(&runtime.FunctionValue{}) {
		t.Fatalf("disabled filter should permit membership checks")
	}
}
