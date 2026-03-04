package compiler

import (
	"regexp"
	"strings"
	"testing"
)

func findCompiledFunction(result *Result, funcName string) (string, bool) {
	compiledSrc := string(result.Files["compiled.go"])
	pattern := "func " + funcName + "("
	idx := strings.Index(compiledSrc, pattern)
	if idx == -1 {
		return "", false
	}
	endMarkers := []string{"\n}\n\nfunc ", "\n}\n\nvar ", "\n}\n\nconst "}
	endIdx := len(compiledSrc)
	for _, marker := range endMarkers {
		if pos := strings.Index(compiledSrc[idx:], marker); pos != -1 && idx+pos < endIdx {
			endIdx = idx + pos + 3
		}
	}
	return compiledSrc[idx:endIdx], true
}

func TestCompilerTypedArrayMethodIntrinsics(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Array { length: i32, capacity: i32, storage_handle: i64 }",
		"",
		"methods Array {",
		"  fn push(self: Self, value: i32) -> void {",
		"    self.length = self.length + 1",
		"  }",
		"  fn len(self: Self) -> i32 { self.length }",
		"}",
		"",
		"fn main() -> i32 {",
		"  arr: Array := Array { length: 3, capacity: 3, storage_handle: 1 }",
		"  arr.push(4)",
		"  _ = arr.set(1, 9)",
		"  len := arr.len()",
		"  arr.get(1) match {",
		"    case value: i32 => value + len,",
		"    case nil => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	fnBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}

	if !strings.Contains(fnBody, "__able_compiled_method_Array_push") {
		t.Fatalf("expected typed-array push to compile to static method call")
	}
	if !strings.Contains(fnBody, "__able_compiled_method_Array_len") {
		t.Fatalf("expected typed-array len to compile to static method call")
	}
	if strings.Contains(fnBody, "__able_member_get_method(") {
		t.Fatalf("expected typed-array get/set to avoid __able_member_get_method")
	}
	if strings.Contains(fnBody, "__able_call_value(") {
		t.Fatalf("expected typed-array get/set to avoid __able_call_value")
	}
	if strings.Contains(fnBody, "runtime.ArrayStoreCapacity(") {
		t.Fatalf("expected typed-array get/set intrinsics to avoid redundant ArrayStoreCapacity calls")
	}
}

func TestCompilerTypedArrayLocalsAvoidGlobalHelpers(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Array { length: i32, capacity: i32, storage_handle: i64 }",
		"",
		"methods Array {",
		"  fn push(self: Self, value: i32) -> void {",
		"    self.length = self.length + 1",
		"  }",
		"  fn len(self: Self) -> i32 { self.length }",
		"}",
		"",
		"fn main() -> i32 {",
		"  arr: Array = Array { length: 0, capacity: 0, storage_handle: 1 }",
		"  arr.push(1)",
		"  arr.len()",
		"}",
		"",
	}, "\n"))

	fnBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if strings.Contains(fnBody, "__able_env_get(") {
		t.Fatalf("expected typed-array local declarations to avoid __able_env_get")
	}
	if strings.Contains(fnBody, "__able_env_set(") {
		t.Fatalf("expected typed-array local declarations to avoid __able_env_set")
	}
}

func TestCompilerTypedArrayLoopsAvoidDynamicDispatch(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Array { length: i32, capacity: i32, storage_handle: i64 }",
		"",
		"methods Array {",
		"  fn push(self: Self, value: i32) -> void {",
		"    self.length = self.length + 1",
		"  }",
		"  fn len(self: Self) -> i32 { self.length }",
		"}",
		"",
		"fn main() -> i32 {",
		"  arr: Array := Array { length: 0, capacity: 0, storage_handle: 1 }",
		"  i := 0",
		"  while i < 6 {",
		"    arr.push(i)",
		"    i = i + 1",
		"  }",
		"",
		"  j := 0",
		"  total := 0",
		"  while j < arr.len() {",
		"    _ = arr.set(j, j)",
		"    arr.get(j) match {",
		"      case n: i32 => { total = total + n },",
		"      case nil => {}",
		"    }",
		"    j = j + 1",
		"  }",
		"  total",
		"}",
		"",
	}, "\n"))

	fnBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}

	if strings.Contains(fnBody, "__able_member_get_method(") {
		t.Fatalf("expected typed-array loops to avoid __able_member_get_method for push/get/set/len")
	}
	if strings.Contains(fnBody, "__able_call_value(") {
		t.Fatalf("expected typed-array loops to avoid __able_call_value for push/get/set/len")
	}
	if strings.Contains(fnBody, "runtime.ArrayStoreCapacity(") {
		t.Fatalf("expected typed-array get/set loops to avoid redundant ArrayStoreCapacity calls")
	}
	if strings.Contains(fnBody, "__able_env_get(") || strings.Contains(fnBody, "__able_env_set(") {
		t.Fatalf("expected typed-array loops to avoid global helper routing")
	}
}

func TestCompilerWhileLoopFastPath(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  i := 0",
		"  while i < 10 {",
		"    i = i + 1",
		"  }",
		"  i",
		"}",
		"",
	}, "\n"))

	fnBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(fnBody, "for {") {
		t.Fatalf("expected while-loop fast path to lower into a direct for-loop")
	}
	if strings.Contains(fnBody, "__able_break") || strings.Contains(fnBody, "__able_continue_signal") {
		t.Fatalf("expected while-loop fast path to avoid panic/defer break/continue scaffolding")
	}
}

func TestCompilerMonoArrayPropagationCastAvoidsIndexBoxing(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"struct Array { length: i32, capacity: i32, storage_handle: i64 }",
		"",
		"fn main() -> i64 {",
		"  arr: Array i32 := [1, 2, 3]",
		"  idx := 1",
		"  total: i64 := 0",
		"  while idx < 3 {",
		"    total = total + (arr[idx]! as i64)",
		"    idx = idx + 1",
		"  }",
		"  total",
		"}",
		"",
	}, "\n"), Options{
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "runtime.ArrayStoreMonoReadI32(") {
		t.Fatalf("expected mono i32 reads in compiled output")
	}
	readToRuntimeBox := regexp.MustCompile(`runtime\.ArrayStoreMonoReadI32\([^\n]+\)\n\s*if err != nil \{\n\s*panic\(err\)\n\s*\}\n\s*__able_tmp_[0-9]+ := bridge\.ToInt\(int64\(__able_tmp_[0-9]+\), runtime\.IntegerType\(\"i32\"\)\)`)
	if readToRuntimeBox.MatchString(compiledSrc) {
		t.Fatalf("expected mono propagation/cast path to avoid bridge.ToInt boxing immediately after mono read")
	}
}
