package compiler

import (
	"strings"
	"testing"
)

func TestCompilerExperimentalMonoArraysStaticBodyStaysOnCompilerOwnedArrayCarrier(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i64 {",
		"  arr := [1, 2, 3]",
		"  arr.push(4)",
		"  arr[1] = 9",
		"  arr.get(2)! as i64",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "const __able_experimental_mono_arrays = true") {
		t.Fatalf("expected mono-array feature flag constant to be enabled")
	}

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var arr *__able_array_i32 =",
		"append(__able_tmp_1.Elements",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected experimental mono-array static lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"runtime.ArrayStore",
		"runtime.ArrayValue",
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_index(",
		"__able_index_set(",
		"__able_array_values(",
		"__able_array_i32_sync(",
		"Storage_handle",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected experimental mono-array static lowering to avoid %q:\n%s", fragment, mainBody)
		}
	}
}
