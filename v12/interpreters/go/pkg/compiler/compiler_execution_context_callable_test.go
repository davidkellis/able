package compiler

import (
	"strings"
	"testing"
)

func TestCompilerExperimentalExecutionContextCarriesNativeCallables(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn apply(callback: (i32 -> i32), value: i32) -> i32 {",
		"  callback(value)",
		"}",
		"",
		"fn main() -> i32 {",
		"  delta := 1",
		"  handle := spawn { apply(fn(value: i32) -> i32 { value + delta }, 41) }",
		"  await [handle] as i32",
		"}",
		"",
	}, "\n"), Options{
		PackageName:                  "main",
		ExperimentalExecutionContext: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	applyBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_apply_ctx")
	for _, fragment := range []string{
		"type __able_fn_int32_to_int32 func(arg0 int32, __able_exec_ctx *__able_execution_context)",
		"__able_call_value_fast_ctx(value, args, __able_exec_ctx)",
		"__able_context_from_native(callCtx)",
		"func __able_call_value_fast_ctx(fn runtime.Value, args []runtime.Value, __able_exec_ctx *__able_execution_context)",
		"native     runtime.NativeCallContext",
		"return &__able_exec_ctx.native",
		"executionContext atomic.Pointer[__able_execution_context]",
		"if ctx := payload.executionContext.Load(); ctx != nil && ctx.env == native.Env",
		"child.payload.executionContext.Store(child)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("context-aware callable ABI is missing %q", fragment)
		}
	}
	if strings.Contains(compiledSrc, "return &runtime.NativeCallContext{") {
		t.Fatalf("context-aware callable path must reuse its embedded native context")
	}
	if !strings.Contains(applyBody, "(value, __able_exec_ctx)") {
		t.Fatalf("native callable invocation must carry the lexical context:\n%s", applyBody)
	}
}

func TestCompilerDefaultNativeCallableABIHasNoContextParameter(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn apply(callback: (i32 -> i32), value: i32) -> i32 { callback(value) }",
		"fn main() -> i32 { apply(fn(value: i32) -> i32 { value + 1 }, 41) }",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "type __able_fn_int32_to_int32 func(arg0 int32) (int32, *__ableControl)") {
		t.Fatalf("default native callable ABI is missing its compatibility signature")
	}
	if strings.Contains(compiledSrc, "type __able_fn_int32_to_int32 func(arg0 int32, __able_exec_ctx *__able_execution_context)") {
		t.Fatalf("default native callable ABI must not contain an execution-context parameter")
	}
}

func TestCompilerExperimentalExecutionContextSkipsCallableABIWithoutAwait(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn apply(callback: (i32 -> i32), value: i32) -> i32 { callback(value) }",
		"fn main() -> i32 { apply(fn(value: i32) -> i32 { value + 1 }, 41) }",
		"",
	}, "\n"), Options{
		PackageName:                  "main",
		ExperimentalExecutionContext: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "type __able_fn_int32_to_int32 func(arg0 int32) (int32, *__ableControl)") {
		t.Fatalf("await-free program should retain the compatibility callable ABI")
	}
	if strings.Contains(compiledSrc, "func __able_call_value_fast_ctx(") {
		t.Fatalf("await-free program must not emit the scheduler-context callable surface")
	}
	if strings.Contains(compiledSrc, "native     runtime.NativeCallContext") {
		t.Fatalf("await-free program must not carry the scheduler native-call view")
	}
}
