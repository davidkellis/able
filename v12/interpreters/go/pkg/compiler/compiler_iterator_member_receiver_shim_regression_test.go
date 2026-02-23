package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesIteratorBuiltinMemberReceiver(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_runtime_iterator_value(value runtime.Value) (*runtime.IteratorValue, bool, bool) {") {
		t.Fatalf("expected shared runtime iterator unwrapping helper to be emitted")
	}

	start := strings.Index(compiledSrc, "func __able_builtin_iterator_next(_ *bridge.Runtime, _ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {")
	if start < 0 {
		t.Fatalf("expected __able_builtin_iterator_next helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_builtin_future_receiver(args []runtime.Value, member string) (*runtime.FutureValue, error) {")
	if end < 0 {
		t.Fatalf("expected __able_builtin_iterator_next segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "receiver, ok := args[0].(*runtime.IteratorValue)") {
		t.Fatalf("expected legacy direct iterator pointer assertion to be removed from __able_builtin_iterator_next")
	}
	if !strings.Contains(segment, "receiver, ok, nilPtr := __able_runtime_iterator_value(args[0])") {
		t.Fatalf("expected __able_builtin_iterator_next to use shared runtime iterator unwrapping helper")
	}
}

