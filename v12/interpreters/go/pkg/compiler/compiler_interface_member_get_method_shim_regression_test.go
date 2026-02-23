package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesInterfaceMemberGetMethodDispatch(t *testing.T) {
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
	end := strings.Index(segment, "func __able_call_compiled_thunk(bytecode any, env *runtime.Environment, args []runtime.Value) (runtime.Value, error, bool) {")
	if end < 0 {
		t.Fatalf("expected __able_member_get_method segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "isIface := false") {
		t.Fatalf("expected legacy isIface tracking branch to be removed from member_get_method")
	}
	if strings.Contains(segment, "if _, ok := obj.(runtime.InterfaceValue); ok {") {
		t.Fatalf("expected legacy interface value assertion branch to be removed from member_get_method")
	}
	if strings.Contains(segment, "if typed, ok := obj.(*runtime.InterfaceValue); ok && typed != nil {") {
		t.Fatalf("expected legacy interface pointer assertion branch to be removed from member_get_method")
	}
	if !strings.Contains(segment, "if iface, ok, nilPtr := __able_callable_interface_value(obj); ok || nilPtr {") {
		t.Fatalf("expected member_get_method to use shared interface unwrapping helper")
	}
}
