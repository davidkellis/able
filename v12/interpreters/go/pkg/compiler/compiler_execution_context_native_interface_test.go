package compiler

import (
	"strings"
	"testing"
)

func TestCompilerExperimentalExecutionContextThreadsNativeInterfaceCalls(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"interface Transformer {",
		"  fn transform(self: Self, value: i32) -> i32",
		"}",
		"",
		"struct Offset { amount: i32 }",
		"",
		"impl Transformer for Offset {",
		"  fn transform(self: Self, value: i32) -> i32 {",
		"    value + self.amount",
		"  }",
		"}",
		"",
		"fn apply(transformer: Transformer, value: i32) -> i32 {",
		"  transformer.transform(value)",
		"}",
		"",
		"fn main() -> i32 {",
		"  apply(Offset { amount: 2 }, 40)",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackSourceWithCompilerOptions(t, source, Options{
		PackageName:                  "main",
		ExperimentalExecutionContext: true,
	})
	compiledSrc := string(result.Files["compiled.go"])
	applyBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_apply_ctx")

	if !strings.Contains(applyBody, ".__able_ctx_transform(value, __able_exec_ctx)") {
		t.Fatalf("compiled interface dispatch must pass its execution context:\n%s", applyBody)
	}
	for _, fragment := range []string{
		"__able_ctx_transform(arg0 int32, __able_exec_ctx *__able_execution_context) (int32, *__ableControl)",
		"return w.__able_ctx_transform(arg0, __able_context_from_args())",
		"func __able_context_with_environment(ctx *__able_execution_context, env *runtime.Environment, local *__able_execution_context) *__able_execution_context",
		"var __able_exec_ctx_package __able_execution_context",
		"__able_context_with_environment(__able_exec_ctx, __able_pkg_env_",
		"child.packageEnv = parent.packageEnv",
		"__able_compiled_impl_Transformer_transform_0_ctx(",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("context-aware native interface adapter is missing %q", fragment)
		}
	}
	if strings.Contains(compiledSrc, "result, control = __able_compiled_entry_impl_Transformer_transform_0_ctx(") {
		t.Fatalf("environment-independent interface implementation should bypass its package entry wrapper")
	}
}

func TestCompilerDefaultNativeInterfaceABIHasNoContextSibling(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Transformer {",
		"  fn transform(self: Self, value: i32) -> i32",
		"}",
		"",
		"struct Offset { amount: i32 }",
		"",
		"impl Transformer for Offset {",
		"  fn transform(self: Self, value: i32) -> i32 { value + self.amount }",
		"}",
		"",
		"fn main() -> i32 {",
		"  value: Transformer := Offset { amount: 2 }",
		"  value.transform(40)",
		"}",
		"",
	}, "\n"))

	if compiledSrc := string(result.Files["compiled.go"]); strings.Contains(compiledSrc, "__able_ctx_transform") {
		t.Fatalf("default native interface ABI must not contain the experimental context sibling")
	}
}

func TestCompilerExperimentalExecutionContextLocalizesEnvironmentIndependentCrossPackageInterfaceBodies(t *testing.T) {
	result := compileNoFallbackPackageWithOptions(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader, SeedReader}",
			"",
			"fn read(reader: Reader) -> i32 { reader.read() }",
			"fn main() -> i32 { read(SeedReader {}) }",
			"",
		}, "\n"),
		"remote/reader.able": strings.Join([]string{
			"SEED: i32 := 42",
			"",
			"interface Reader {",
			"  fn read(self: Self) -> i32",
			"}",
			"",
			"struct SeedReader {}",
			"",
			"impl Reader for SeedReader {",
			"  fn read(self: Self) -> i32 { SEED }",
			"}",
			"",
		}, "\n"),
	}, Options{
		PackageName:                  "main",
		RequireNoFallbacks:           true,
		EmitMain:                     true,
		ExperimentalExecutionContext: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	readBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_read_ctx")
	if !strings.Contains(readBody, ".__able_ctx_read(__able_exec_ctx)") {
		t.Fatalf("cross-package interface dispatch must pass its execution context:\n%s", readBody)
	}
	for _, fragment := range []string{
		"__able_compiled_impl_Reader_read_0_ctx(",
		"__able_compiled_entry_impl_Reader_read_0_ctx(",
		"result, control = __able_compiled_impl_Reader_read_0_ctx(w.Value, __able_exec_ctx)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("cross-package interface adapter is missing %q", fragment)
		}
	}
	if strings.Contains(compiledSrc, "if __able_exec_ctx != nil && __able_exec_ctx.packageEnv ==") {
		t.Fatalf("environment-independent cross-package interface body must not retain a package-environment guard")
	}
}

func TestCompilerExperimentalExecutionContextSpawnedInterfaceCapturedCallbackExecutes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generated execution-context binary in short mode")
	}
	t.Setenv("ABLE_EXECUTOR", "goroutine")
	stdout := compileAndRunSourceWithOptions(t, "ablec-context-interface-callback-", strings.Join([]string{
		"package demo",
		"",
		"interface Worker {",
		"  fn apply(self: Self, value: i32, callback: (i32 -> i32)) -> i32",
		"}",
		"",
		"struct CallbackWorker {}",
		"",
		"impl Worker for CallbackWorker {",
		"  fn apply(self: Self, value: i32, callback: (i32 -> i32)) -> i32 {",
		"    callback(value)",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  delta := 2",
		"  callback := fn(value: i32) -> i32 { value + delta }",
		"  worker: Worker := CallbackWorker {}",
		"  handle := spawn { worker.apply(40, callback) }",
		"  future_flush()",
		"  print(handle.value())",
		"}",
		"",
	}, "\n"), Options{
		PackageName:                  "main",
		EmitMain:                     true,
		RequireNoFallbacks:           true,
		ExperimentalExecutionContext: true,
	})

	if got := strings.TrimSpace(stdout); got != "42" {
		t.Fatalf("spawned interface callback stdout = %q, want 42", got)
	}
}
