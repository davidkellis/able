package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRemovesErrorValueMemberGetMethodShim(t *testing.T) {
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

	if strings.Contains(compiledSrc, "if hasErrorValue && name == \"value\" {") {
		t.Fatalf("expected legacy Error.value member_get_method shim branch to be removed")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"Error\", \"message\", true, 0, 0, __able_builtin_error_message)") {
		t.Fatalf("expected Error.message builtin method registration to remain")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"Error\", \"cause\", true, 0, 0, __able_builtin_error_cause)") {
		t.Fatalf("expected Error.cause builtin method registration to remain")
	}
	segmentStart := strings.Index(compiledSrc, "func __able_try_member_get_method(obj runtime.Value, member runtime.Value) (runtime.Value, error) {")
	if segmentStart < 0 {
		t.Fatalf("expected member_get_method helper to be emitted")
	}
	segment := compiledSrc[segmentStart:]
	segmentEnd := strings.Index(segment, "func __able_member_get_method(")
	if segmentEnd < 0 {
		t.Fatalf("expected member_get_method helper segment terminator")
	}
	segment = segment[:segmentEnd]
	if !strings.Contains(segment, "if _, ok, nilPtr := __able_runtime_error_value(base); ok || nilPtr {") {
		t.Fatalf("expected Error values to preserve a dedicated runtime.ErrorValue member lookup branch")
	}
	if !strings.Contains(segment, "if entry := __able_lookup_compiled_method(\"Error\", name, true); entry != nil && entry.fn != nil {") {
		t.Fatalf("expected runtime.ErrorValue member lookup to bind the builtin Error methods directly")
	}
	errorIdx := strings.Index(segment, "if _, ok, nilPtr := __able_runtime_error_value(base); ok || nilPtr {")
	typeIdx := strings.Index(segment, "if typeName := __able_runtime_value_type_name(base); typeName != \"\" {")
	if typeIdx < 0 || errorIdx < 0 || errorIdx > typeIdx {
		t.Fatalf("expected Error interface dispatch guard before runtime value type-name lookup")
	}
}
