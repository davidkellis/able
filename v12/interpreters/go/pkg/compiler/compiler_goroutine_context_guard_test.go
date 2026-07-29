package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNestedSpawnImportedEnvironmentIndependentCallUsesRawBody(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.worker",
			"",
			"fn main() -> i32 {",
			"  handle := spawn { worker.answer() }",
			"  handle.value()! as i32",
			"}",
			"",
		}, "\n"),
		"worker/helpers.able": strings.Join([]string{
			"fn answer() -> i32 { 42 }",
			"",
		}, "\n"),
	})

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main_ctx")
	if !strings.Contains(mainBody, "__able_compiled_fn_answer_ctx(") {
		t.Fatalf("nested spawn should use the proven-independent imported context body:\n%s", mainBody)
	}
	if strings.Contains(mainBody, "__able_compiled_entry_fn_answer_ctx(") {
		t.Fatalf("nested spawn should not pay an environment swap for a proven-independent callee:\n%s", mainBody)
	}
	for _, fragment := range []string{"__able_call_named(", "__able_call_value("} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("nested imported static call should avoid dynamic helper %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerDynamicCallHelpersConstructExplicitNativeContext(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": "package demo\n\nfn main() -> void {}\n",
	})

	for _, helper := range []string{"func __able_call_value(", "func __able_call_named("} {
		start := strings.Index(compiledSrc, helper)
		if start < 0 {
			t.Fatalf("generated runtime is missing %s", helper)
		}
		body := compiledSrc[start:]
		if next := strings.Index(body[len(helper):], "\nfunc "); next >= 0 {
			body = body[:len(helper)+next]
		}
		for _, fragment := range []string{
			"env := __able_runtime.Env()",
			"state = env.RuntimeData()",
			"&runtime.NativeCallContext{Env: env, State: state}",
		} {
			if !strings.Contains(body, fragment) {
				t.Fatalf("%s should construct an explicit native context with %q:\n%s", helper, fragment, body)
			}
		}
	}
}

func TestCompilerSpawnSelectsSchedulerExecutionContextByDefault(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  channel := __able_channel_new(1)",
		"  handle := spawn { __able_channel_send(channel, 7) }",
		"  future_flush()",
		"  _ = handle",
		"}",
		"",
	}, "\n"), Options{PackageName: "main"})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_execution_context struct",
		"func __able_compiled_fn_main_ctx(__able_exec_ctx *__able_execution_context)",
		"func __able_compiled_fn_main() (struct{}, *__ableControl)",
		"__able_spawn_context(__able_exec_ctx, func(__able_child_ctx *__able_execution_context)",
		"func __able_channel_send_ctx(args []runtime.Value, __able_exec_ctx *__able_execution_context)",
		"__able_channel_send_ctx([]runtime.Value{",
		"__able_child_ctx)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("default spawn context path is missing %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerExperimentalExecutionContextThreadsStaticSpawnKernelCalls(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  channel := __able_channel_new(1)",
		"  handle := spawn { __able_channel_send(channel, 7) }",
		"  future_flush()",
		"  _ = handle",
		"}",
		"",
	}, "\n"), Options{
		PackageName:                  "main",
		ExperimentalExecutionContext: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_execution_context struct",
		"func __able_compiled_fn_main_ctx(__able_exec_ctx *__able_execution_context)",
		"func __able_compiled_fn_main() (struct{}, *__ableControl)",
		"__able_spawn_context(__able_exec_ctx, func(__able_child_ctx *__able_execution_context)",
		"func __able_channel_send_ctx(args []runtime.Value, __able_exec_ctx *__able_execution_context)",
		"__able_channel_send_ctx([]runtime.Value{",
		"__able_child_ctx)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("experimental generated context path is missing %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerExperimentalExecutionContextUsesFixedPointerForStaticCalls(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn increment(value: i64) -> i64 {",
		"  value + 1",
		"}",
		"",
		"fn main() -> i64 {",
		"  increment(41)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:                  "main",
		ExperimentalExecutionContext: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_compiled_fn_increment_ctx(value int64, __able_exec_ctx *__able_execution_context)",
		"func __able_compiled_fn_increment(value int64) (int64, *__ableControl)",
		"__able_compiled_fn_increment_ctx(int64(41), __able_exec_ctx)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("fixed context ABI is missing %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerExperimentalExecutionContextProvidesFixedHelperSurface(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-fixed-context-helper-surface", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.hash_map.*",
		"",
		"fn main() -> i32 {",
		"  values: HashMap String i32 := #{ \"a\": 1, \"b\": 2 }",
		"  values[\"a\"]! + values[\"b\"]!",
		"}",
		"",
	}, "\n"), Options{
		PackageName:                  "main",
		ExperimentalExecutionContext: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, helper := range runtimeHelperContextAdapterNames {
		fragment := "func " + runtimeHelperContextName(helper) + "(args []runtime.Value, __able_exec_ctx *__able_execution_context)"
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("fixed helper surface is missing %q", fragment)
		}
	}
	for _, fragment := range []string{
		"__able_hash_map_new_ctx(nil, __able_exec_ctx)",
		"__able_hash_map_set_ctx([]runtime.Value{handleVal, __able_map_key_0, __able_map_value_0}, __able_exec_ctx)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("map literal did not use the fixed helper surface %q:\n%s", fragment, compiledSrc)
		}
	}
}
