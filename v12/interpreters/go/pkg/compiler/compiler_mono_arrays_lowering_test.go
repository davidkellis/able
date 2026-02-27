package compiler

import (
	"strings"
	"testing"
)

func TestCompilerMonoArraysFlagEnablesTypedArrayLowering(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"struct Array T { length: i32, capacity: i32, storage_handle: i64 }",
		"",
		"fn main() -> i32 {",
		"  flags: Array bool := [true, true]",
		"  flags[1] = false",
		"  if flags[0] { 1 } else { 0 }",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	fnBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(fnBody, "runtime.ArrayStoreMonoNewWithCapacityBool(") {
		t.Fatalf("expected typed bool array literal lowering under mono flag")
	}
	if !strings.Contains(fnBody, "runtime.ArrayStoreMonoWriteBool(") {
		t.Fatalf("expected typed bool index assignment lowering under mono flag")
	}
	if !strings.Contains(fnBody, "runtime.ArrayStoreMonoReadBool(") {
		t.Fatalf("expected typed bool index read lowering under mono flag")
	}
}

func TestCompilerMonoArraysFlagDisabledKeepsLegacyArrayLowering(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"struct Array T { length: i32, capacity: i32, storage_handle: i64 }",
		"",
		"fn main() -> i32 {",
		"  flags: Array bool := [true, true]",
		"  flags[1] = false",
		"  if flags[0] { 1 } else { 0 }",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: false,
	})

	fnBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if strings.Contains(fnBody, "runtime.ArrayStoreMono") {
		t.Fatalf("expected mono array lowering to stay disabled without flag")
	}
}

func TestCompilerMonoArraysFlagEnablesPushLenIntrinsics(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"struct Array T { length: i32, capacity: i32, storage_handle: i64 }",
		"",
		"fn main() -> i32 {",
		"  flags: Array bool := []",
		"  flags.push(true)",
		"  flags.push(false)",
		"  if flags.len() == 2 { 1 } else { 0 }",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	fnBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(fnBody, "runtime.ArrayStoreMonoWriteBool(") {
		t.Fatalf("expected mono push lowering to use runtime.ArrayStoreMonoWriteBool")
	}
	if !strings.Contains(fnBody, "runtime.ArrayStoreSize(") {
		t.Fatalf("expected mono len lowering to use runtime.ArrayStoreSize")
	}
	if strings.Contains(fnBody, "__able_compiled_method_Array_push") {
		t.Fatalf("expected mono push intrinsic to bypass compiled method call")
	}
	if strings.Contains(fnBody, "__able_compiled_method_Array_len") {
		t.Fatalf("expected mono len intrinsic to bypass compiled method call")
	}
}

func TestCompilerMonoArraysFlagEnablesI64Lowering(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"struct Array T { length: i32, capacity: i32, storage_handle: i64 }",
		"",
		"fn main() -> i64 {",
		"  values: Array i64 := [1_i64, 2_i64]",
		"  values.push(3_i64)",
		"  values[1] = 8_i64",
		"  values[1]! + values[2]!",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	fnBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(fnBody, "runtime.ArrayStoreMonoNewWithCapacityI64(") {
		t.Fatalf("expected typed i64 array literal lowering under mono flag")
	}
	if !strings.Contains(fnBody, "runtime.ArrayStoreMonoWriteI64(") {
		t.Fatalf("expected typed i64 write lowering under mono flag")
	}
	if !strings.Contains(fnBody, "runtime.ArrayStoreMonoReadI64(") {
		t.Fatalf("expected typed i64 read lowering under mono flag")
	}
}
