package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim(t *testing.T) {
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

	start := strings.Index(compiledSrc, "func __able_member_get_method(obj runtime.Value, member runtime.Value) runtime.Value {")
	if start < 0 {
		t.Fatalf("expected __able_member_get_method helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_member(obj runtime.Value, member runtime.Value) runtime.Value {")
	if end < 0 {
		t.Fatalf("expected __able_member_get_method segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "case runtime.StructDefinitionValue:") {
		t.Fatalf("expected dedicated StructDefinition member_get_method static lookup shim branch to be removed")
	}
	if strings.Contains(segment, "__able_lookup_compiled_method(typed.Node.ID.Name, name, false)") {
		t.Fatalf("expected direct StructDefinition member_get_method lookup shim call to be removed")
	}
	if !strings.Contains(segment, "if typeName := __able_runtime_value_type_name(base); typeName != \"\" {") {
		t.Fatalf("expected shared runtime type-name member_get_method path to remain")
	}
	if !strings.Contains(segment, "if __able_interface_dispatch_static_receiver(base) {") {
		t.Fatalf("expected shared static-receiver member_get_method path to remain")
	}
	if !strings.Contains(segment, "__able_lookup_compiled_method(typeName, name, false)") {
		t.Fatalf("expected shared static member_get_method compiled lookup path to remain")
	}
}
