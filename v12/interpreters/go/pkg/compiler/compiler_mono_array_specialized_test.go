package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerExperimentalMonoArraysTypedArrayUsesSpecializedWrapper(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn sum(values: Array i32) -> i32 {",
		"  values.push(3)",
		"  values[0]! as i32 + values[1]! as i32 + values[2]! as i32",
		"}",
		"",
		"fn main() -> i32 {",
		"  values: Array i32 = [1, 2]",
		"  sum(values)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_array_i32 struct {",
		"Elements []int32",
		"func __able_array_i32_from(value runtime.Value) (*__able_array_i32, error) {",
		"func __able_array_i32_to(rt *bridge.Runtime, value *__able_array_i32) (runtime.Value, error) {",
		"func __able_compiled_fn_sum(values *__able_array_i32) (int32, *__ableControl) {",
		"func __able_compiled_fn_main() (int32, *__ableControl) {",
		"var values *__able_array_i32 =",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized mono-array lowering to contain %q", fragment)
		}
	}

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var values *__able_array_i32 =",
		"&__able_array_i32{Elements: []int32{int32(1), int32(2)}}",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected mono-array main body to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"runtime.ArrayValue",
		"runtime.ArrayStore",
		"[]runtime.Value{",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected specialized mono-array main body to avoid %q:\n%s", fragment, mainBody)
		}
	}

	sumBody, ok := findCompiledFunction(result, "__able_compiled_fn_sum")
	if !ok {
		t.Fatalf("could not find compiled sum function")
	}
	for _, fragment := range []string{
		"Elements = append(",
	} {
		if !strings.Contains(sumBody, fragment) {
			t.Fatalf("expected mono-array sum body to contain %q:\n%s", fragment, sumBody)
		}
	}
	for _, fragment := range []string{
		"__able_ptr(",
		"__able_nullable_i32_to_value(",
		"__able_array_i32_sync(",
	} {
		if strings.Contains(sumBody, fragment) {
			t.Fatalf("expected mono-array sum body to avoid obsolete nullable-pointer boxing %q:\n%s", fragment, sumBody)
		}
	}
}

func TestCompilerExperimentalMonoArraysF64TypedArrayUsesSpecializedWrapper(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn sum(values: Array f64) -> f64 {",
		"  values.push(3.5)",
		"  values[0]! + values[1]! + values[2]!",
		"}",
		"",
		"fn main() -> f64 {",
		"  values: Array f64 = [1.25, 2.75]",
		"  sum(values)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_array_f64 struct {",
		"Elements []float64",
		"func __able_array_f64_from(value runtime.Value) (*__able_array_f64, error) {",
		"func __able_array_f64_to(rt *bridge.Runtime, value *__able_array_f64) (runtime.Value, error) {",
		"func __able_compiled_fn_sum(values *__able_array_f64) (float64, *__ableControl) {",
		"func __able_compiled_fn_main() (float64, *__ableControl) {",
		"var values *__able_array_f64 =",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized f64 mono-array lowering to contain %q", fragment)
		}
	}

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var values *__able_array_f64 =",
		"&__able_array_f64{Elements: []float64{float64(1.25), float64(2.75)}}",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected f64 mono-array main body to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"runtime.ArrayValue",
		"runtime.ArrayStore",
		"[]runtime.Value{",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected specialized f64 mono-array main body to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerExperimentalMonoArraysCharTypedArrayUsesSpecializedWrapper(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn tail(values: Array char) -> char {",
		"  values.push('z')",
		"  values[2]!",
		"}",
		"",
		"fn main() -> char {",
		"  values: Array char = ['a', 'b']",
		"  tail(values)",
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
		"func __able_array_char_from(value runtime.Value) (*__able_array_char, error) {",
		"func __able_array_char_to(rt *bridge.Runtime, value *__able_array_char) (runtime.Value, error) {",
		"func __able_compiled_fn_tail(values *__able_array_char) (rune, *__ableControl) {",
		"func __able_compiled_fn_main() (rune, *__ableControl) {",
		"var values *__able_array_char =",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized char mono-array lowering to contain %q", fragment)
		}
	}

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var values *__able_array_char =",
		"&__able_array_char{Elements: []rune{rune('a'), rune('b')}}",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected char mono-array main body to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"runtime.ArrayValue",
		"runtime.ArrayStore",
		"[]runtime.Value{",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected specialized char mono-array main body to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerExperimentalMonoArraysStringTypedArrayUsesSpecializedWrapper(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn head(values: Array String) -> String {",
		"  values.push(\"gamma\")",
		"  values[0]!",
		"}",
		"",
		"fn main() -> String {",
		"  values: Array String = [\"alpha\", \"beta\"]",
		"  head(values)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_array_String struct {",
		"Elements []string",
		"func __able_array_String_from(value runtime.Value) (*__able_array_String, error) {",
		"func __able_array_String_to(rt *bridge.Runtime, value *__able_array_String) (runtime.Value, error) {",
		"func __able_compiled_fn_head(values *__able_array_String) (string, *__ableControl) {",
		"func __able_compiled_fn_main() (string, *__ableControl) {",
		"var values *__able_array_String =",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected specialized String mono-array lowering to contain %q", fragment)
		}
	}

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var values *__able_array_String =",
		"&__able_array_String{Elements: []string{\"alpha\", \"beta\"}}",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected String mono-array main body to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"runtime.ArrayValue",
		"runtime.ArrayStore",
		"[]runtime.Value{",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected specialized String mono-array main body to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerExperimentalMonoArraysTypedArrayWrapperUsesSpecializedBoundaryConverters(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn cloneish(values: Array i32) -> Array i32 {",
		"  values.push(3)",
		"  values",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_cloneish(values *__able_array_i32) (*__able_array_i32, *__ableControl)") {
		t.Fatalf("expected cloneish to keep a specialized mono-array signature")
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_cloneish")
	if !ok {
		t.Fatalf("could not find wrapper for cloneish")
	}
	if !strings.Contains(wrapBody, "__able_array_i32_from(arg0Value)") {
		t.Fatalf("expected wrapper arg conversion to use explicit mono-array helper:\n%s", wrapBody)
	}
	if !strings.Contains(wrapBody, "return __able_array_i32_to(rt, compiledResult)") {
		t.Fatalf("expected wrapper return to use explicit mono-array helper:\n%s", wrapBody)
	}
}

func TestCompilerExperimentalMonoArraysTypedArrayExecutes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping mono-array compiler integration test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	source := strings.Join([]string{
		"package demo",
		"",
		"fn sum(values: Array i32) -> i32 {",
		"  values.push(9)",
		"  values[0]! as i32 + values[1]! as i32 + values[2]! as i32",
		"}",
		"",
		"fn main() -> void {",
		"  nums: Array i32 = [4, 5]",
		"  print(sum(nums))",
		"}",
		"",
	}, "\n")
	stdout := compileAndRunSourceWithOptions(t, "ablec-mono-array-", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "18" {
		t.Fatalf("expected compiled mono-array program to print 18, got %q", stdout)
	}
}

func TestCompilerExperimentalMonoArraysF64Executes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping mono-array f64 compiler integration test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	source := strings.Join([]string{
		"package demo",
		"",
		"fn sum(values: Array f64) -> f64 {",
		"  values.push(3.5)",
		"  values[0]! + values[1]! + values[2]!",
		"}",
		"",
		"fn main() -> void {",
		"  nums: Array f64 = [1.25, 2.75]",
		"  print(sum(nums))",
		"}",
		"",
	}, "\n")
	stdout := compileAndRunSourceWithOptions(t, "ablec-mono-array-f64-", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "7.5" {
		t.Fatalf("expected compiled mono-array f64 program to print 7.5, got %q", stdout)
	}
}

func TestCompilerExperimentalMonoArraysTextTypedArraysExecute(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping mono-array text compiler integration test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	source := strings.Join([]string{
		"package demo",
		"",
		"fn build_word(chars: Array char) -> i32 {",
		"  chars.push('!')",
		"  labels: Array String = [\"left\", \"right\"]",
		"  labels.push(\"center\")",
		"  chars.len() + labels.len()",
		"}",
		"",
		"fn main() -> void {",
		"  chars: Array char = ['g', 'o']",
		"  print(build_word(chars))",
		"}",
		"",
	}, "\n")
	stdout := compileAndRunSourceWithOptions(t, "ablec-mono-array-text-", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "6" {
		t.Fatalf("expected compiled mono-array text program to print 6, got %q", stdout)
	}
}

func TestCompilerExperimentalMonoArraysCharResultPropagationStaysNative(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn inner(values: Array char) -> !Array char {",
		"  values",
		"}",
		"",
		"fn outer(values: Array char) -> !Array char {",
		"  inner(values)!",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_outer")
	if !ok {
		t.Fatalf("could not find compiled outer function")
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_outer(values *__able_array_char) (__able_union_",
		"_as_ptr___able_array_char(",
		"_wrap_ptr___able_array_char(",
	} {
		if !strings.Contains(string(result.Files["compiled.go"]), fragment) && !strings.Contains(body, fragment) {
			t.Fatalf("expected char result propagation to contain %q", fragment)
		}
	}
	if strings.Contains(body, "_from_value(__able_runtime, __able_tmp_") {
		t.Fatalf("expected char result propagation to avoid runtime-value reconversion of specialized success branch:\n%s", body)
	}
}

func TestCompilerExperimentalMonoArraysCharResultPropagationExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn inner(values: Array char) -> !Array char {",
		"  values",
		"}",
		"",
		"fn outer(values: Array char) -> !Array char {",
		"  inner(values)!",
		"}",
		"",
		"fn main() -> void {",
		"  chars: Array char = ['a', 'b', 'c']",
		"  print(outer(chars)!.len())",
		"}",
		"",
	}, "\n")
	stdout := compileAndRunSourceWithOptions(t, "ablec-mono-array-char-result-", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "3" {
		t.Fatalf("expected compiled char result propagation program to print 3, got %q", stdout)
	}
}

func compileAndRunSourceWithOptions(t *testing.T, tempPrefix string, source string, opts Options) string {
	t.Helper()
	moduleRoot, workDir := compilerTestWorkDir(t, tempPrefix)

	entryPath := filepath.Join(workDir, "app.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	outputDir := filepath.Join(workDir, "out")
	opts.EntryPath = entryPath
	result, err := New(opts).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(outputDir); err != nil {
		t.Fatalf("write output: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-bin")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = outputDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}

	run := exec.Command(binPath)
	output, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, string(output))
	}
	return string(output)
}
