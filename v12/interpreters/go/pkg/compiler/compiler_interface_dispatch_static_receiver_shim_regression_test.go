package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesInterfaceDispatchStaticReceiver(t *testing.T) {
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

	start := strings.Index(compiledSrc, "func __able_interface_dispatch_static_receiver(")
	if start < 0 {
		t.Fatalf("expected __able_interface_dispatch_static_receiver helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_runtime_struct_definition_value(")
	if end < 0 {
		t.Fatalf("expected __able_interface_dispatch_static_receiver segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch receiver.(type)") {
		t.Fatalf("expected legacy pointer/value switch dispatch to be removed from __able_interface_dispatch_static_receiver")
	}
	if strings.Contains(segment, "runtime.StructDefinitionValue") {
		t.Fatalf("expected explicit StructDefinition switch branches to be removed from __able_interface_dispatch_static_receiver")
	}
	if strings.Contains(segment, "runtime.TypeRefValue") {
		t.Fatalf("expected explicit TypeRef switch branches to be removed from __able_interface_dispatch_static_receiver")
	}
	if strings.Contains(segment, "if _, ok, nilPtr := __able_runtime_struct_definition_value(receiver); ok {") {
		t.Fatalf("expected legacy ok-only struct-definition helper guard to be removed")
	}
	if !strings.Contains(segment, "if _, ok, nilPtr := __able_runtime_struct_definition_value(receiver); ok || nilPtr {") {
		t.Fatalf("expected __able_interface_dispatch_static_receiver to use normalized struct-definition helper guard")
	}
	if strings.Contains(segment, "if _, ok, nilPtr := __able_runtime_type_ref_value(receiver); ok {") {
		t.Fatalf("expected legacy ok-only type-ref helper guard to be removed")
	}
	if !strings.Contains(segment, "if _, ok, nilPtr := __able_runtime_type_ref_value(receiver); ok || nilPtr {") {
		t.Fatalf("expected __able_interface_dispatch_static_receiver to use normalized type-ref helper guard")
	}
	if !strings.Contains(segment, "return ok && !nilPtr") {
		t.Fatalf("expected __able_interface_dispatch_static_receiver to preserve typed-nil rejection semantics")
	}
}
