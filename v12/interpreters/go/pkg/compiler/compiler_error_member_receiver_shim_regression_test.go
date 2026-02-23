package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesErrorBuiltinMemberReceivers(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_runtime_error_value(value runtime.Value) (runtime.ErrorValue, bool, bool) {") {
		t.Fatalf("expected shared runtime error unwrapping helper to be emitted")
	}

	helperStart := strings.Index(compiledSrc, "func __able_builtin_error_receiver(args []runtime.Value, member string)")
	if helperStart < 0 {
		t.Fatalf("expected shared Error receiver helper to be emitted")
	}
	helperSegment := compiledSrc[helperStart:]
	helperEnd := strings.Index(helperSegment, "func __able_builtin_error_message(")
	if helperEnd < 0 {
		t.Fatalf("expected Error.message helper segment terminator")
	}
	helperSegment = helperSegment[:helperEnd]
	if strings.Contains(helperSegment, "if typed, ok := args[0].(*runtime.ErrorValue); ok && typed != nil {") {
		t.Fatalf("expected legacy direct pointer assertion to be removed from shared Error receiver helper")
	}
	if strings.Contains(helperSegment, "if recv, ok, nilPtr := __able_runtime_error_value(args[0]); ok {") {
		t.Fatalf("expected legacy ok-only error receiver guard to be removed")
	}
	if !strings.Contains(helperSegment, "if recv, ok, nilPtr := __able_runtime_error_value(args[0]); ok || nilPtr {") {
		t.Fatalf("expected shared Error receiver helper to use normalized runtime error unwrapping guard")
	}
	if !strings.Contains(helperSegment, "if !ok || nilPtr {") {
		t.Fatalf("expected shared Error receiver helper to treat typed-nil receiver as invalid")
	}

	messageStart := strings.Index(compiledSrc, "func __able_builtin_error_message(")
	if messageStart < 0 {
		t.Fatalf("expected Error.message builtin helper to be emitted")
	}
	messageSegment := compiledSrc[messageStart:]
	messageEnd := strings.Index(messageSegment, "func __able_builtin_error_cause(")
	if messageEnd < 0 {
		t.Fatalf("expected Error.message segment terminator")
	}
	messageSegment = messageSegment[:messageEnd]
	if strings.Contains(messageSegment, "switch recv := args[0].(type)") {
		t.Fatalf("expected legacy Error.message receiver switch shim to be removed")
	}
	if !strings.Contains(messageSegment, "__able_builtin_error_receiver(args, \"message\")") {
		t.Fatalf("expected Error.message to use shared Error receiver helper")
	}

	causeStart := strings.Index(compiledSrc, "func __able_builtin_error_cause(")
	if causeStart < 0 {
		t.Fatalf("expected Error.cause builtin helper to be emitted")
	}
	causeSegment := compiledSrc[causeStart:]
	causeEnd := strings.Index(causeSegment, "func __able_builtin_named_future_yield(")
	if causeEnd < 0 {
		t.Fatalf("expected Error.cause segment terminator")
	}
	causeSegment = causeSegment[:causeEnd]
	if strings.Contains(causeSegment, "switch typed := args[0].(type)") {
		t.Fatalf("expected legacy Error.cause receiver switch shim to be removed")
	}
	if !strings.Contains(causeSegment, "__able_builtin_error_receiver(args, \"cause\")") {
		t.Fatalf("expected Error.cause to use shared Error receiver helper")
	}
}
