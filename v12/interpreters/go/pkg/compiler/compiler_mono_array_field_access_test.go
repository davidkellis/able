package compiler

import (
	"strings"
	"testing"
)

func TestCompilerMonoArrayFieldAccessStaysSliceBacked(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  data := [1, 2, 3, 4]",
		"  [first, second, ...tail] := data",
		"  first + second + tail.length + tail.capacity + (tail.storage_handle as i32)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"int32(__able_slice_len(",
		"int32(__able_slice_cap(",
		"int64(0)",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected mono-array field access lowering to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"tail.Length",
		"tail.Capacity",
		"tail.Storage_handle",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected mono-array field access to avoid legacy wrapper metadata %q:\n%s", fragment, body)
		}
	}
}
