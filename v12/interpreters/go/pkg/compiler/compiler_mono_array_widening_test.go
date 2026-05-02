package compiler

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCompilerExperimentalMonoArraysInferredLiteralLoopAndPatternStaySpecialized(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn summarize() -> i32 {",
		"  nums := [1, 2, 3]",
		"  nums.push(4)",
		"  total := 0",
		"  for value in nums {",
		"    total = total + value",
		"  }",
		"  nums match {",
		"    case [head, ...tail] => total + head + (tail[0] as i32)",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	summaryBody, ok := findCompiledFunction(result, "__able_compiled_fn_summarize")
	if !ok {
		t.Fatalf("could not find compiled summarize function")
	}
	for _, fragment := range []string{
		"var nums *__able_array_i32 =",
		"var value int32",
		"var tail *__able_array_i32 =",
	} {
		if !strings.Contains(summaryBody, fragment) {
			t.Fatalf("expected widened mono-array lowering to contain %q:\n%s", fragment, summaryBody)
		}
	}
	for _, fragment := range []string{
		"__able_array_values(",
		"runtime.ArrayValue",
		"[]runtime.Value{",
		"__able_call_value(",
		"__able_method_call_node(",
		"__able_array_i32_sync(",
	} {
		if strings.Contains(summaryBody, fragment) {
			t.Fatalf("expected widened mono-array lowering to avoid %q:\n%s", fragment, summaryBody)
		}
	}
}

func TestCompilerExperimentalMonoArraysFactoryCloneAndReserveStaySpecialized(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn build() -> Array i32 {",
		"  values: Array i32 = Array.with_capacity(4)",
		"  values.push(7)",
		"  values.push(9)",
		"  values.reserve(8)",
		"  copy := values.clone_shallow()",
		"  copy",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	buildBody, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_build() (*__able_array_i32, *__ableControl)",
		"make([]int32, 0, ",
		"make([]int32, len(",
		"var copy *__able_array_i32 =",
	} {
		if !strings.Contains(compiledSrc, fragment) && !strings.Contains(buildBody, fragment) {
			t.Fatalf("expected widened mono-array factory lowering to contain %q", fragment)
		}
	}
	for _, fragment := range []string{
		"runtime.ArrayValue",
		"[]runtime.Value{",
		"__able_struct_Array_clone_elements(",
		"__able_array_i32_sync(",
	} {
		if strings.Contains(buildBody, fragment) {
			t.Fatalf("expected widened mono-array factory lowering to avoid %q:\n%s", fragment, buildBody)
		}
	}
}

func TestCompilerExperimentalMonoArraysWidenedSliceExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn build() -> Array i32 {",
		"  values: Array i32 = Array.with_capacity(4)",
		"  values.push(7)",
		"  values.push(9)",
		"  values.reserve(8)",
		"  values.clone_shallow()",
		"}",
		"",
		"fn summarize() -> i32 {",
		"  nums := [1, 2, 3]",
		"  nums.push(4)",
		"  total := 0",
		"  for value in nums {",
		"    total = total + value",
		"  }",
		"  nums match {",
		"    case [head, ...tail] => total + head + (tail[0] as i32)",
		"    case _ => 0",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  built := build()",
		"  print((built[1]! as i32) + summarize())",
		"}",
		"",
	}, "\n")
	stdout := compileAndRunSourceWithOptions(t, "ablec-mono-array-wide-", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "22" {
		t.Fatalf("expected widened mono-array program to print 22, got %q", stdout)
	}
}

func TestCompilerExperimentalMonoArraysPropagationComputedIndexStaysHoisted(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn last_value() -> i32 {",
		"  nums := [2, 3, 5]",
		"  count := nums.len()",
		"  nums[count - 1]!",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_last_value")
	if !ok {
		t.Fatalf("could not find compiled last_value function")
	}
	for _, fragment := range []string{
		"var nums *__able_array_i32 =",
		"__able_tmp_",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected propagated mono-array index lowering to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"func() int32 {",
		"func() runtime.Value {",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected computed mono-array propagation index to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerExperimentalMonoArraysNestedF64RowsStaySpecialized(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn build(n: i32) -> Array (Array f64) {",
		"  rows: Array (Array f64) = Array.new()",
		"  i := 0",
		"  loop {",
		"    if i >= n { break }",
		"    row: Array f64 = Array.new()",
		"    row.push((i as f64) + 0.5)",
		"    row.push((i as f64) + 1.5)",
		"    rows.push(row)",
		"    i = i + 1",
		"  }",
		"  rows",
		"}",
		"",
		"fn main() -> f64 {",
		"  matrix := build(2)",
		"  matrix[1]!.get(0)! + matrix[0]!.get(1)!",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_array_array_f64 struct {",
		"Elements []*__able_array_f64",
		"func __able_array_array_f64_from(value runtime.Value) (*__able_array_array_f64, error) {",
		"func __able_array_array_f64_to(rt *bridge.Runtime, value *__able_array_array_f64) (runtime.Value, error) {",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected nested mono-array lowering to contain %q", fragment)
		}
	}

	buildBody, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"var rows *__able_array_array_f64 =",
		"var row *__able_array_f64 =",
		"&__able_array_array_f64{}",
		"&__able_array_f64{}",
		".Elements = append(",
		"float64(i)",
	} {
		if !strings.Contains(buildBody, fragment) {
			t.Fatalf("expected nested f64 mono-array build body to contain %q:\n%s", fragment, buildBody)
		}
	}
	for _, fragment := range []string{
		"[]runtime.Value{",
		"runtime.ArrayValue",
		"__able_array_values(",
		"__able_array_f64_to(__able_runtime, row)",
		"__able_array_f64_from(",
		"__able_cast(",
		"bridge.AsFloat(",
		"__able_array_array_f64_sync(",
		"__able_array_f64_sync(",
	} {
		if strings.Contains(buildBody, fragment) {
			t.Fatalf("expected nested f64 row lowering to avoid %q in build body:\n%s", fragment, buildBody)
		}
	}
}

func TestCompilerExperimentalMonoArraysNestedF64RowsExecute(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn build(n: i32) -> Array (Array f64) {",
		"  rows: Array (Array f64) = Array.new()",
		"  i := 0",
		"  loop {",
		"    if i >= n { break }",
		"    row: Array f64 = Array.new()",
		"    row.push((i as f64) + 0.5)",
		"    row.push((i as f64) + 1.5)",
		"    rows.push(row)",
		"    i = i + 1",
		"  }",
		"  rows",
		"}",
		"",
		"fn main() -> void {",
		"  matrix := build(2)",
		"  print(matrix[1]!.get(0)! + matrix[0]!.get(1)!)",
		"}",
		"",
	}, "\n")
	stdout := compileAndRunSourceWithOptions(t, "ablec-mono-array-f64-nested-", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "3" {
		t.Fatalf("expected nested f64 mono-array program to print 3, got %q", stdout)
	}
}

func TestCompilerExperimentalMonoArraysNestedF64GetPushStaysSpecialized(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn transpose(rows: Array (Array f64)) -> Array (Array f64) {",
		"  n := rows.len()",
		"  out: Array (Array f64) = Array.new()",
		"  i := 0",
		"  loop {",
		"    if i >= n { break }",
		"    col: Array f64 = Array.new()",
		"    j := 0",
		"    loop {",
		"      if j >= n { break }",
		"      col.push(rows.get(j)!.get(i)!)",
		"      j = j + 1",
		"    }",
		"    out.push(col)",
		"    i = i + 1",
		"  }",
		"  out",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_transpose")
	if !ok {
		t.Fatalf("could not find compiled transpose function")
	}
	for _, fragment := range []string{
		"var out *__able_array_array_f64 =",
		"var col *__able_array_f64 =",
		".Elements = append(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected nested f64 get/push path to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_nullable_f64_to_value(",
		"bridge.AsFloat(",
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_array_f64_to(__able_runtime, col)",
		"__able_array_f64_from(",
		"__able_ptr(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nested f64 get/push path to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerExperimentalMonoArraysMatrixMultiplyScalarLoopStaysNative(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn dot(ai: Array f64, cj: Array f64) -> f64 {",
		"  total := 0.0",
		"  k := 0",
		"  loop {",
		"    if k >= ai.len() { break }",
		"    total = total + ai.get(k)! * cj.get(k)!",
		"    k = k + 1",
		"  }",
		"  total",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_dot")
	if !ok {
		t.Fatalf("could not find compiled dot function")
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_dot(ai *__able_array_f64, cj *__able_array_f64) (float64, *__ableControl)",
		"var total float64",
		" * ",
	} {
		if !strings.Contains(string(result.Files["compiled.go"]), fragment) && !strings.Contains(body, fragment) {
			t.Fatalf("expected matrix scalar loop lowering to contain %q", fragment)
		}
	}
	for _, fragment := range []string{
		"__able_nullable_f64_to_value(",
		"bridge.AsFloat(",
		"__able_binary_op(\"*\",",
		"__able_call_value(",
		"__able_method_call_node(",
		"bridge.PushCallFrame(__able_runtime,",
		"bridge.PopCallFrame(__able_runtime)",
		"__able_ptr(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected matrix scalar loop lowering to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerExperimentalMonoArraysMatrixMultiplyMainStaysNative(t *testing.T) {
	sourcePath := filepath.Join(repositoryRoot(), "v12", "examples", "benchmarks", "matrixmultiply.able")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read matrixmultiply benchmark: %v", err)
	}

	result := compileNoFallbackExecSourceWithOptions(t, "ablec-matrixmultiply-main", string(source), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"math.Trunc(",
		"int32(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected matrixmultiply main lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_cast(",
		"bridge.AsInt(",
		"__able_ptr(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected matrixmultiply main lowering to avoid %q:\n%s", fragment, mainBody)
		}
	}

	for _, funcName := range []string{"__able_compiled_fn_build_matrix", "__able_compiled_fn_matmul"} {
		body, ok := findCompiledFunction(result, funcName)
		if !ok {
			t.Fatalf("could not find compiled %s function", funcName)
		}
		for _, fragment := range []string{
			"__able_nullable_f64_to_value(",
			"bridge.AsFloat(",
			"__able_binary_op(\"*\",",
			"__able_binary_op(\"+\",",
			"__able_call_value(",
			"__able_method_call_node(",
			"bridge.PushCallFrame(__able_runtime,",
			"bridge.PopCallFrame(__able_runtime)",
			"__able_ptr(",
		} {
			if strings.Contains(body, fragment) {
				t.Fatalf("expected %s to avoid %q:\n%s", funcName, fragment, body)
			}
		}
	}
}

func TestCompilerExperimentalMonoArraysMatrixMultiplyCountedLoopsStayNative(t *testing.T) {
	sourcePath := filepath.Join(repositoryRoot(), "v12", "examples", "benchmarks", "matrixmultiply.able")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read matrixmultiply benchmark: %v", err)
	}

	result := compileNoFallbackExecSourceWithOptions(t, "ablec-matrixmultiply-counted-loops", string(source), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	buildBody, ok := findCompiledFunction(result, "__able_compiled_fn_build_matrix")
	if !ok {
		t.Fatalf("could not find compiled build_matrix function")
	}
	for _, fragment := range []string{
		"for i < n {",
		"for j < n {",
		"i++",
		"j++",
	} {
		if !strings.Contains(buildBody, fragment) {
			t.Fatalf("expected build_matrix counted-loop lowering to contain %q:\n%s", fragment, buildBody)
		}
	}
	if !regexp.MustCompile(`__able_tmp_[0-9]+ := __able_tmp_[0-9]+ - __able_tmp_[0-9]+`).MatchString(buildBody) {
		t.Fatalf("expected build_matrix to keep proven non-negative subtraction inline:\n%s", buildBody)
	}
	if regexp.MustCompile(`int64\(__able_tmp_[0-9]+\) - int64\(__able_tmp_[0-9]+\)`).MatchString(buildBody) {
		t.Fatalf("expected build_matrix to avoid widened checked subtraction after range proof:\n%s", buildBody)
	}
	if !regexp.MustCompile(`__able_tmp_[0-9]+ := __able_tmp_[0-9]+ \+ __able_tmp_[0-9]+`).MatchString(buildBody) {
		t.Fatalf("expected build_matrix to keep proven bounded addition inline:\n%s", buildBody)
	}
	if regexp.MustCompile(`int64\(__able_tmp_[0-9]+\) \+ int64\(__able_tmp_[0-9]+\)`).MatchString(buildBody) {
		t.Fatalf("expected build_matrix to avoid widened checked addition after upper-bound proof:\n%s", buildBody)
	}
	for _, fragment := range []string{
		"for {",
		"if i >= n {",
		"if j >= n {",
		"__able_runtime_error_value(",
		"__able_checked_add_signed(",
		"__able_checked_sub_signed(",
	} {
		if strings.Contains(buildBody, fragment) {
			t.Fatalf("expected build_matrix counted-loop lowering to avoid %q:\n%s", fragment, buildBody)
		}
	}

	matmulBody, ok := findCompiledFunction(result, "__able_compiled_fn_matmul")
	if !ok {
		t.Fatalf("could not find compiled matmul function")
	}
	for _, fragment := range []string{
		"for i < n {",
		"for j < n {",
		"for k < n {",
		"i++",
		"j++",
		"k++",
	} {
		if !strings.Contains(matmulBody, fragment) {
			t.Fatalf("expected matmul counted-loop lowering to contain %q:\n%s", fragment, matmulBody)
		}
	}
	for _, fragment := range []string{
		"for {",
		"if i >= n {",
		"if j >= n {",
		"if k >= n {",
		"__able_runtime_error_value(",
		"__able_checked_add_signed(",
	} {
		if strings.Contains(matmulBody, fragment) {
			t.Fatalf("expected matmul counted-loop lowering to avoid %q:\n%s", fragment, matmulBody)
		}
	}
}

func TestCompilerExperimentalMonoArraysNestedF64GetPushExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn build(n: i32) -> Array (Array f64) {",
		"  rows: Array (Array f64) = Array.new()",
		"  i := 0",
		"  loop {",
		"    if i >= n { break }",
		"    row: Array f64 = Array.new()",
		"    row.push((i as f64) + 0.5)",
		"    row.push((i as f64) + 1.5)",
		"    rows.push(row)",
		"    i = i + 1",
		"  }",
		"  rows",
		"}",
		"",
		"fn transpose(rows: Array (Array f64)) -> Array (Array f64) {",
		"  n := rows.len()",
		"  out: Array (Array f64) = Array.new()",
		"  i := 0",
		"  loop {",
		"    if i >= n { break }",
		"    col: Array f64 = Array.new()",
		"    j := 0",
		"    loop {",
		"      if j >= n { break }",
		"      col.push(rows.get(j)!.get(i)!)",
		"      j = j + 1",
		"    }",
		"    out.push(col)",
		"    i = i + 1",
		"  }",
		"  out",
		"}",
		"",
		"fn main() -> void {",
		"  matrix := transpose(build(2))",
		"  print(matrix[1]!.get(0)! + matrix[0]!.get(1)!)",
		"}",
		"",
	}, "\n")
	stdout := compileAndRunSourceWithOptions(t, "ablec-mono-array-f64-get-push-", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "3" {
		t.Fatalf("expected nested f64 get/push program to print 3, got %q", stdout)
	}
}

func TestCompilerExperimentalMonoArraysNestedCharRowsStaySpecialized(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn build() -> Array (Array char) {",
		"  rows: Array (Array char) = Array.new()",
		"  row: Array char = ['a', 'b']",
		"  rows.push(row)",
		"  rows",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_array_char struct {",
		"Elements []rune",
		"type __able_array_array_char struct {",
		"Elements []*__able_array_char",
		"func __able_array_array_char_from(value runtime.Value) (*__able_array_array_char, error) {",
		"func __able_array_array_char_to(rt *bridge.Runtime, value *__able_array_array_char) (runtime.Value, error) {",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected nested char mono-array lowering to contain %q", fragment)
		}
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"var rows *__able_array_array_char =",
		"var row *__able_array_char =",
		".Elements = append(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected nested char row build body to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"runtime.ArrayValue",
		"__able_struct_Array_to(__able_runtime, row)",
		"[]runtime.Value{",
		"__able_array_array_char_sync(",
		"__able_array_char_sync(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nested char row build body to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerNestedCarrierArraysDefaultToNativeWrappers(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn build() -> Array (Array char) {",
		"  rows: Array (Array char) = Array.new()",
		"  row: Array char = ['a', 'b']",
		"  rows.push(row)",
		"  rows[0]!.push('c')",
		"  rows",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"const __able_experimental_mono_arrays = true",
		"type __able_array_array_char struct {",
		"Elements []*__able_array_char",
		"func __able_array_array_char_from(value runtime.Value) (*__able_array_array_char, error) {",
		"func __able_array_array_char_to(rt *bridge.Runtime, value *__able_array_array_char) (runtime.Value, error) {",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected default nested carrier-array lowering to contain %q", fragment)
		}
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"var rows *__able_array_array_char =",
		"var row *__able_array_char =",
		".Elements = append(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected default nested carrier-array build body to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_any_to_value(row)",
		"runtime.ArrayValue",
		"__able_struct_Array_from(__able_tmp_",
		"__able_array_array_char_sync(",
		"__able_array_char_sync(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected default nested carrier-array build body to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerNestedCarrierArraysPreserveRowIdentityByDefault(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn build() -> Array (Array char) {",
		"  rows: Array (Array char) = Array.new()",
		"  row: Array char = ['a', 'b']",
		"  rows.push(row)",
		"  rows[0]!.push('c')",
		"  rows",
		"}",
		"",
		"fn main() -> void {",
		"  rows := build()",
		"  print(rows[0]!.len())",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunSourceWithOptions(t, "ablec-mono-off-carrier-array-", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "3" {
		t.Fatalf("expected default nested carrier-array program to print 3, got %q", stdout)
	}
}

func TestCompilerExperimentalMonoArraysInterfaceCarrierArrayStaysSpecialized(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"interface Greeter for Self {",
		"  fn greet(self: Self) -> String",
		"}",
		"",
		"struct Person { name: String }",
		"",
		"impl Greeter for Person {",
		"  fn greet(self: Self) -> String { self.name }",
		"}",
		"",
		"fn join(values: Array Greeter) -> String {",
		"  `${values[0]!.greet()} ${values.get(1)!.greet()}`",
		"}",
		"",
		"fn main() -> String {",
		"  values: Array Greeter = [Person { name: \"Ada\" }, Person { name: \"Grace\" }]",
		"  join(values)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_array_iface_Greeter struct {",
		"Elements []__able_iface_Greeter",
		"func __able_array_iface_Greeter_from(value runtime.Value) (*__able_array_iface_Greeter, error) {",
		"func __able_array_iface_Greeter_to(rt *bridge.Runtime, value *__able_array_iface_Greeter) (runtime.Value, error) {",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected interface-carrier array lowering to contain %q", fragment)
		}
	}

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var values *__able_array_iface_Greeter =",
		"[]__able_iface_Greeter{",
		"__able_iface_Greeter_wrap_ptr_Person(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected interface-carrier array main body to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"runtime.ArrayValue",
		"[]runtime.Value{",
		"__able_iface_Greeter_to_runtime_value(__able_runtime,",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected interface-carrier array main body to avoid %q:\n%s", fragment, mainBody)
		}
	}

	joinBody, ok := findCompiledFunction(result, "__able_compiled_fn_join")
	if !ok {
		t.Fatalf("could not find compiled join function")
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_join(values *__able_array_iface_Greeter) (string, *__ableControl)",
		".greet()",
	} {
		if !strings.Contains(compiledSrc, fragment) && !strings.Contains(joinBody, fragment) {
			t.Fatalf("expected interface-carrier array join lowering to contain %q", fragment)
		}
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(joinBody, fragment) {
			t.Fatalf("expected interface-carrier array join lowering to avoid %q:\n%s", fragment, joinBody)
		}
	}
}

func TestCompilerExperimentalMonoArraysCallableCarrierArrayStaysSpecialized(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  funcs: Array (i32 -> i32) = [",
		"    fn(value: i32) -> i32 { value + 1 },",
		"    fn(value: i32) -> i32 { value + 2 }",
		"  ]",
		"  funcs[0]!(40) + funcs.get(1)!(40)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_array_fn_int32_to_int32 struct {",
		"Elements []__able_fn_int32_to_int32",
		"func __able_array_fn_int32_to_int32_from(value runtime.Value) (*__able_array_fn_int32_to_int32, error) {",
		"func __able_array_fn_int32_to_int32_to(rt *bridge.Runtime, value *__able_array_fn_int32_to_int32) (runtime.Value, error) {",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected callable-carrier array lowering to contain %q", fragment)
		}
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var funcs *__able_array_fn_int32_to_int32 =",
		"[]__able_fn_int32_to_int32{",
		"__able_fn_int32_to_int32(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected callable-carrier array main body to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"runtime.NativeFunctionValue",
		"runtime.ArrayValue",
		"[]runtime.Value{",
		"__able_call_value(",
		"__able_call_value_fast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected callable-carrier array main body to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerExperimentalMonoArraysCarrierArraysExecute(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"interface Greeter for Self {",
		"  fn greet(self: Self) -> String",
		"}",
		"",
		"struct Person { name: String }",
		"",
		"impl Greeter for Person {",
		"  fn greet(self: Self) -> String { self.name }",
		"}",
		"",
		"fn main() -> void {",
		"  greeters: Array Greeter = [Person { name: \"Ada\" }, Person { name: \"Grace\" }]",
		"  funcs: Array (i32 -> i32) = [",
		"    fn(value: i32) -> i32 { value + 1 },",
		"    fn(value: i32) -> i32 { value + 2 }",
		"  ]",
		"  ok := greeters[0]!.greet() == \"Ada\" &&",
		"        greeters.get(1)!.greet() == \"Grace\" &&",
		"        funcs[0]!(40) + funcs.get(1)!(40) == 83",
		"  if ok { print(\"ok\") } else { print(\"bad\") }",
		"}",
		"",
	}, "\n")
	stdout := compileAndRunSourceWithOptions(t, "ablec-mono-array-carriers-", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "ok" {
		t.Fatalf("expected carrier-array program to print \"ok\", got %q", stdout)
	}
}
