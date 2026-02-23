package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesRuntimeValueTypeNameUnwrapping(t *testing.T) {
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

	start := strings.Index(compiledSrc, "func __able_runtime_value_type_name(")
	if start < 0 {
		t.Fatalf("expected __able_runtime_value_type_name helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_interface_dispatch_match_equivalent(")
	if end < 0 {
		t.Fatalf("expected __able_runtime_value_type_name segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch typed := value.(type)") {
		t.Fatalf("expected legacy pointer/value switch over value to be removed from __able_runtime_value_type_name")
	}
	if strings.Contains(segment, "case runtime.InterfaceValue:") {
		t.Fatalf("expected legacy interface-value switch branch to be removed from __able_runtime_value_type_name")
	}
	if strings.Contains(segment, "case *runtime.InterfaceValue:") {
		t.Fatalf("expected legacy interface pointer switch branch to be removed from __able_runtime_value_type_name")
	}
	if strings.Contains(segment, "case runtime.StructDefinitionValue:") {
		t.Fatalf("expected legacy struct-definition value switch branch to be removed from __able_runtime_value_type_name")
	}
	if strings.Contains(segment, "case *runtime.StructDefinitionValue:") {
		t.Fatalf("expected legacy struct-definition pointer switch branch to be removed from __able_runtime_value_type_name")
	}
	if strings.Contains(segment, "case runtime.TypeRefValue:") {
		t.Fatalf("expected legacy type-ref value switch branch to be removed from __able_runtime_value_type_name")
	}
	if strings.Contains(segment, "case *runtime.TypeRefValue:") {
		t.Fatalf("expected legacy type-ref pointer switch branch to be removed from __able_runtime_value_type_name")
	}
	if strings.Contains(segment, "case runtime.IntegerValue:") {
		t.Fatalf("expected legacy integer value switch branch to be removed from __able_runtime_value_type_name")
	}
	if strings.Contains(segment, "case *runtime.IntegerValue:") {
		t.Fatalf("expected legacy integer pointer switch branch to be removed from __able_runtime_value_type_name")
	}
	if strings.Contains(segment, "case runtime.FloatValue:") {
		t.Fatalf("expected legacy float value switch branch to be removed from __able_runtime_value_type_name")
	}
	if strings.Contains(segment, "case *runtime.FloatValue:") {
		t.Fatalf("expected legacy float pointer switch branch to be removed from __able_runtime_value_type_name")
	}

	if strings.Contains(segment, "if iface, ok, nilPtr := __able_callable_interface_value(value); ok {") {
		t.Fatalf("expected legacy ok-only interface helper guard to be removed")
	}
	if !strings.Contains(segment, "if iface, ok, nilPtr := __able_callable_interface_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_runtime_value_type_name to use normalized interface unwrapping helper guard")
	}
	if strings.Contains(segment, "if structDef, ok, nilPtr := __able_runtime_struct_definition_value(value); ok {") {
		t.Fatalf("expected legacy ok-only struct-definition helper guard to be removed")
	}
	if !strings.Contains(segment, "if structDef, ok, nilPtr := __able_runtime_struct_definition_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_runtime_value_type_name to use normalized struct-definition unwrapping helper guard")
	}
	if strings.Contains(segment, "if typeRef, ok, nilPtr := __able_runtime_type_ref_value(value); ok {") {
		t.Fatalf("expected legacy ok-only type-ref helper guard to be removed")
	}
	if !strings.Contains(segment, "if typeRef, ok, nilPtr := __able_runtime_type_ref_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_runtime_value_type_name to use normalized type-ref unwrapping helper guard")
	}
	if strings.Contains(segment, "if intVal, ok, nilPtr := __able_runtime_integer_value(value); ok {") {
		t.Fatalf("expected legacy ok-only integer helper guard to be removed")
	}
	if !strings.Contains(segment, "if intVal, ok, nilPtr := __able_runtime_integer_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_runtime_value_type_name to use normalized integer unwrapping helper guard")
	}
	if strings.Contains(segment, "if floatVal, ok, nilPtr := __able_runtime_float_value(value); ok {") {
		t.Fatalf("expected legacy ok-only float helper guard to be removed")
	}
	if !strings.Contains(segment, "if floatVal, ok, nilPtr := __able_runtime_float_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_runtime_value_type_name to use normalized float unwrapping helper guard")
	}
	if !strings.Contains(segment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_runtime_value_type_name to preserve explicit typed-nil rejection")
	}
}
