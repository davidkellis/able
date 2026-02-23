package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesArrayMemberReceiverUnwrap(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_runtime_array_value(value runtime.Value) (*runtime.ArrayValue, bool, bool) {") {
		t.Fatalf("expected shared runtime array unwrapping helper to be emitted")
	}

	memberSetStart := strings.Index(compiledSrc, "func __able_member_set(obj runtime.Value, member runtime.Value, value runtime.Value) runtime.Value {")
	if memberSetStart < 0 {
		t.Fatalf("expected __able_member_set helper to be emitted")
	}
	memberSetSegment := compiledSrc[memberSetStart:]
	memberSetEnd := strings.Index(memberSetSegment, "func __able_member_get(obj runtime.Value, member runtime.Value) runtime.Value {")
	if memberSetEnd < 0 {
		t.Fatalf("expected __able_member_set segment terminator")
	}
	memberSetSegment = memberSetSegment[:memberSetEnd]
	if strings.Contains(memberSetSegment, "if arr, ok := base.(*runtime.ArrayValue); ok && arr != nil {") {
		t.Fatalf("expected legacy direct array pointer assertion to be removed from __able_member_set")
	}
	if !strings.Contains(memberSetSegment, "if arr, ok, _ := __able_runtime_array_value(base); ok {") {
		t.Fatalf("expected __able_member_set to use shared runtime array unwrapping helper")
	}

	memberGetStart := strings.Index(compiledSrc, "func __able_member_get(obj runtime.Value, member runtime.Value) runtime.Value {")
	if memberGetStart < 0 {
		t.Fatalf("expected __able_member_get helper to be emitted")
	}
	memberGetSegment := compiledSrc[memberGetStart:]
	memberGetEnd := strings.Index(memberGetSegment, "func __able_member_get_method(obj runtime.Value, member runtime.Value) runtime.Value {")
	if memberGetEnd < 0 {
		t.Fatalf("expected __able_member_get segment terminator")
	}
	memberGetSegment = memberGetSegment[:memberGetEnd]
	if strings.Contains(memberGetSegment, "if arr, ok := base.(*runtime.ArrayValue); ok && arr != nil {") {
		t.Fatalf("expected legacy direct array pointer assertion to be removed from __able_member_get")
	}
	if !strings.Contains(memberGetSegment, "if arr, ok, _ := __able_runtime_array_value(base); ok {") {
		t.Fatalf("expected __able_member_get to use shared runtime array unwrapping helper")
	}
}

