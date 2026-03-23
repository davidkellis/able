package compiler

import (
	"strings"
	"testing"
)

func TestCompilerStaticKernelHelpersUseDirectImplCalls(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  handle := __able_array_new()",
		"  _ = __able_array_set_len(handle, 2)",
		"  _ = __able_array_write(handle, 0, 65)",
		"  _ = __able_array_write(handle, 1, 66)",
		"  bytes := __able_String_from_builtin(\"hi\")",
		"  _ = __able_String_to_builtin(bytes)",
		"  _ = __able_char_to_codepoint(__able_char_from_codepoint(65))",
		"  map_handle := __able_hash_map_new()",
		"  _ = __able_hash_map_set(map_handle, \"a\", 1)",
		"  _ = __able_hash_map_get(map_handle, \"a\")",
		"  ch := __able_channel_new(1)",
		"  _ = __able_channel_send(ch, 1)",
		"  _ = __able_channel_receive(ch)",
		"  mu := __able_mutex_new()",
		"  _ = __able_mutex_lock(mu)",
		"  _ = __able_mutex_unlock(mu)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("main body not found")
	}

	required := []string{
		"__able_array_new_impl(",
		"__able_array_set_len_impl(",
		"__able_array_write_impl(",
		"__able_string_from_builtin_impl(",
		"__able_string_to_builtin_impl(",
		"__able_char_from_codepoint_impl(",
		"__able_char_to_codepoint_impl(",
		"__able_hash_map_new_impl(",
		"__able_hash_map_set_impl(",
		"__able_hash_map_get_impl(",
		"__able_channel_new_impl(",
		"__able_channel_send_impl(",
		"__able_channel_receive_impl(",
		"__able_mutex_new_impl(",
		"__able_mutex_lock_impl(",
		"__able_mutex_unlock_impl(",
		"__able_control_from_error_with_node(",
	}
	for _, fragment := range required {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected static helper call to use %q directly:\n%s", fragment, body)
		}
	}

	forbidden := []string{
		"__able_extern_array_",
		"__able_extern_hash_map_",
		"__able_extern_string_",
		"__able_extern_char_",
		"__able_extern_channel_",
		"__able_extern_mutex_",
		"__able_extern_call(",
	}
	for _, fragment := range forbidden {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected static kernel helper path to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerRuntimeHelperBodiesAvoidExternCallChaining(t *testing.T) {
	sourceResult := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  bytes := __able_String_from_builtin(\"ok\")",
		"  _ = __able_String_to_builtin(bytes)",
		"}",
		"",
	}, "\n"))

	arrayValuesBody, ok := findCompiledFunction(sourceResult, "__able_array_values")
	if !ok {
		t.Fatalf("__able_array_values helper not found")
	}
	for _, fragment := range []string{"__able_extern_array_", "__able_extern_call("} {
		if strings.Contains(arrayValuesBody, fragment) {
			t.Fatalf("expected __able_array_values to avoid %q:\n%s", fragment, arrayValuesBody)
		}
	}
	for _, fragment := range []string{"__able_array_size_impl(", "__able_array_read_impl("} {
		if !strings.Contains(arrayValuesBody, fragment) {
			t.Fatalf("expected __able_array_values to use %q:\n%s", fragment, arrayValuesBody)
		}
	}

	concurrencyResult := compileExecFixtureResult(t, "06_12_19_stdlib_concurrency_channel_mutex_queue")
	channelCommitBody, ok := findCompiledFunction(concurrencyResult, "(a *__able_channel_awaitable) commit")
	if !ok {
		t.Fatalf("channel awaitable commit helper not found")
	}
	mutexCommitBody, ok := findCompiledFunction(concurrencyResult, "(a *__able_mutex_awaitable) commit")
	if !ok {
		t.Fatalf("mutex awaitable commit helper not found")
	}

	for _, body := range []string{channelCommitBody, mutexCommitBody} {
		if strings.Contains(body, "__able_extern_call(") {
			t.Fatalf("expected awaitable commit helper to avoid __able_extern_call:\n%s", body)
		}
	}
	if !strings.Contains(channelCommitBody, "__able_channel_receive_impl(") || !strings.Contains(channelCommitBody, "__able_channel_send_impl(") {
		t.Fatalf("expected channel awaitable commit helper to use direct channel impl helpers:\n%s", channelCommitBody)
	}
	if !strings.Contains(mutexCommitBody, "__able_mutex_lock_impl(") {
		t.Fatalf("expected mutex awaitable commit helper to use direct mutex impl helper:\n%s", mutexCommitBody)
	}
}

func TestCompilerZeroArgCallableTypeSyntaxStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn forty_one() -> i32 { 41 }",
		"",
		"fn invoke(f: fn() -> i32) -> i32 {",
		"  f()",
		"}",
		"",
		"fn main() -> i32 {",
		"  invoke(forty_one)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_invoke")
	if !ok {
		t.Fatalf("invoke body not found")
	}
	if !strings.Contains(body, "__able_fn_void_to_int32") {
		t.Fatalf("expected zero-arg callable syntax to normalize onto the native void-arg callable carrier:\n%s", body)
	}
	for _, fragment := range []string{"__able_call_value(", "__able_method_call_node(", "call arity mismatch"} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected zero-arg callable syntax to stay on native callable dispatch without %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerAwaitDefaultZeroArgCallbackSpecializationStaysNative(t *testing.T) {
	result := compileExecFixtureResult(t, "06_01_compiler_spawn_await")

	body, ok := findCompiledFunction(result, "__able_compiled_method_Await__default_spec")
	if !ok {
		t.Fatalf("Await.default specialization not found")
	}
	if !strings.Contains(body, "callback __able_fn_void_to_string") {
		t.Fatalf("expected Await.default specialization to keep a native zero-arg callback carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_fn_runtime_Value_to_string",
		"__able_fn_runtime_Value_to_string_from_runtime_value",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected Await.default specialization to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerNilableStructExternArgUsesRuntimeNilValue(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Box {",
		"  value: i32",
		"}",
		"",
		"extern go fn sink(value: ?Box) -> void {",
		"  return nil",
		"}",
		"",
		"fn maybe(flag: bool) -> ?Box {",
		"  if flag { Box { value: 1 } } else { nil }",
		"}",
		"",
		"fn main() -> void {",
		"  value := maybe(false)",
		"  sink(value)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("main body not found")
	}
	if !strings.Contains(body, "runtime.NilValue{}") {
		t.Fatalf("expected nilable struct extern arg lowering to emit runtime.NilValue{}:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_struct_Box_to(__able_runtime, (*Box)(nil))",
		"__able_any_to_value(any(nil))",
		"missing Box value",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nilable struct extern arg lowering to avoid %q:\n%s", fragment, body)
		}
	}
}
