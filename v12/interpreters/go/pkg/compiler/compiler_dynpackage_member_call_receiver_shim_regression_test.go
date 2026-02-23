package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesDynPackageBuiltinMemberCallReceiver(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_runtime_dynpackage_value(value runtime.Value) (runtime.DynPackageValue, bool, bool) {") {
		t.Fatalf("expected shared runtime dynpackage unwrapping helper to be emitted")
	}

	start := strings.Index(compiledSrc, "func __able_builtin_dynpackage_member_call(")
	if start < 0 {
		t.Fatalf("expected dynpackage builtin member-call helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_builtin_dynpackage_def(")
	if end < 0 {
		t.Fatalf("expected dynpackage builtin helper segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch typed := receiver.(type)") {
		t.Fatalf("expected legacy dynpackage receiver switch shim to be removed")
	}
	if strings.Contains(segment, "if typed, ok := receiver.(*runtime.DynPackageValue); ok && typed != nil {") {
		t.Fatalf("expected legacy dynpackage pointer receiver assertion to be removed")
	}
	if !strings.Contains(segment, "if dyn, ok, nilPtr := __able_runtime_dynpackage_value(receiver); ok || nilPtr {") {
		t.Fatalf("expected dynpackage receiver normalization to use shared runtime helper")
	}
}
