package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesFutureBuiltinMemberReceiver(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_runtime_future_value(value runtime.Value) (*runtime.FutureValue, bool, bool) {") {
		t.Fatalf("expected shared runtime future unwrapping helper to be emitted")
	}

	start := strings.Index(compiledSrc, "func __able_builtin_future_receiver(args []runtime.Value, member string) (*runtime.FutureValue, error) {")
	if start < 0 {
		t.Fatalf("expected __able_builtin_future_receiver helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_builtin_future_status(")
	if end < 0 {
		t.Fatalf("expected __able_builtin_future_receiver segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "recv, ok := args[0].(*runtime.FutureValue)") {
		t.Fatalf("expected legacy direct future pointer assertion to be removed from __able_builtin_future_receiver")
	}
	if !strings.Contains(segment, "recv, ok, nilPtr := __able_runtime_future_value(args[0])") {
		t.Fatalf("expected __able_builtin_future_receiver to use shared runtime future unwrapping helper")
	}
}

