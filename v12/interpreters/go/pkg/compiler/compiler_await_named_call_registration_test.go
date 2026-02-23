package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRegistersBuiltinAwaitNamedCalls(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"fn main() -> void {",
			"  _ = 1",
			"}",
			"",
		}, "\n"),
	})

	if !strings.Contains(compiledSrc, "func __able_builtin_named_await_default(") {
		t.Fatalf("expected builtin __able_await_default compiled call helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "func __able_builtin_named_await_sleep_ms(") {
		t.Fatalf("expected builtin __able_await_sleep_ms compiled call helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_call(env, \"__able_await_default\", -1, 0, \"\", __able_builtin_named_await_default)") {
		t.Fatalf("expected __able_await_default compiled call registration")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_call(env, \"__able_await_sleep_ms\", -1, 1, \"\", __able_builtin_named_await_sleep_ms)") {
		t.Fatalf("expected __able_await_sleep_ms compiled call registration")
	}
	if strings.Contains(compiledSrc, "case \"__able_await_default\":") {
		t.Fatalf("expected legacy __able_await_default __able_call_named shim branch to be removed")
	}
	if strings.Contains(compiledSrc, "case \"__able_await_sleep_ms\":") {
		t.Fatalf("expected legacy __able_await_sleep_ms __able_call_named shim branch to be removed")
	}
}
