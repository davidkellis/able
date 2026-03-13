package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesArrayIndexReceiverUnwrap(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  _ = 1",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_runtime_array_value(value runtime.Value) (*runtime.ArrayValue, bool, bool) {") {
		t.Fatalf("expected shared runtime array unwrapping helper to be emitted")
	}

	indexBody, ok := findCompiledFunction(result, "__able_index")
	if !ok {
		t.Fatalf("could not find __able_index helper")
	}
	if strings.Contains(indexBody, "if arr, ok := base.(*runtime.ArrayValue); ok && arr != nil {") {
		t.Fatalf("expected legacy direct array pointer assertion to be removed from __able_index")
	}
	if !strings.Contains(indexBody, "if arr, ok, _ := __able_runtime_array_value(base); ok {") {
		t.Fatalf("expected __able_index to use shared runtime array unwrapping helper")
	}

	indexSetBody, ok := findCompiledFunction(result, "__able_index_set")
	if !ok {
		t.Fatalf("could not find __able_index_set helper")
	}
	if strings.Contains(indexSetBody, "if arr, ok := base.(*runtime.ArrayValue); ok && arr != nil {") {
		t.Fatalf("expected legacy direct array pointer assertion to be removed from __able_index_set")
	}
	if !strings.Contains(indexSetBody, "if arr, ok, _ := __able_runtime_array_value(base); ok {") {
		t.Fatalf("expected __able_index_set to use shared runtime array unwrapping helper")
	}
}
