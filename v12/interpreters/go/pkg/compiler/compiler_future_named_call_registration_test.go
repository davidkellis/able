package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRegistersBuiltinFutureNamedCalls(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"fn main() -> void {",
			"  _ = future_pending_tasks()",
			"}",
			"",
		}, "\n"),
	})

	if !strings.Contains(compiledSrc, "func __able_builtin_named_future_yield(") {
		t.Fatalf("expected builtin future_yield compiled call helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "func __able_register_builtin_compiled_calls(entryEnv *runtime.Environment, interp *interpreter.Interpreter)") {
		t.Fatalf("expected builtin compiled call registration helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_call(env, \"future_yield\", 0, 0, \"\", __able_builtin_named_future_yield)") {
		t.Fatalf("expected future_yield compiled call registration")
	}
	if !strings.Contains(compiledSrc, "__able_register_builtin_compiled_calls(entryEnv, interp)") {
		t.Fatalf("expected RegisterIn to invoke builtin compiled call registration")
	}
	if strings.Contains(compiledSrc, "case \"future_yield\":") {
		t.Fatalf("expected legacy future_yield __able_call_named shim branch to be removed")
	}
	if strings.Contains(compiledSrc, "case \"future_cancelled\":") {
		t.Fatalf("expected legacy future_cancelled __able_call_named shim branch to be removed")
	}
	if strings.Contains(compiledSrc, "case \"future_flush\":") {
		t.Fatalf("expected legacy future_flush __able_call_named shim branch to be removed")
	}
	if strings.Contains(compiledSrc, "case \"future_pending_tasks\":") {
		t.Fatalf("expected legacy future_pending_tasks __able_call_named shim branch to be removed")
	}
}
