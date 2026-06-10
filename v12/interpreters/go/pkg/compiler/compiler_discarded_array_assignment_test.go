package compiler

import (
	"strings"
	"testing"
)

func TestCompilerDiscardedPrimitiveArrayWriteAvoidsRuntimeValueCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn write(arr: Array i32, index: i32, value: i32) -> void {",
		"  arr[index] = value",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_write")
	if !ok {
		t.Fatal("could not find compiled write function")
	}
	if strings.Contains(body, "runtime.Value") {
		t.Fatalf("discarded primitive array write should not box its result:\n%s", body)
	}
	for _, fragment := range []string{
		".Elements[",
		"__able_index_error(",
		"__able_raise_control(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("discarded primitive array write should contain %q:\n%s", fragment, body)
		}
	}
}
