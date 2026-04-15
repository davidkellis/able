package compiler

import (
	"strings"
	"testing"
)

func TestCompilerMonoArrayFromHelperStaysDirectAtArrayBoundary(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn cloneish(values: Array i32) -> Array i32 {",
		"  values",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	body, ok := findCompiledFunction(result, "__able_array_i32_from")
	if !ok {
		t.Fatalf("could not find mono-array from helper")
	}
	for _, fragment := range []string{
		"current := __able_unwrap_interface(value)",
		"if raw, ok, nilPtr := __able_runtime_array_value(current); ok || nilPtr {",
		"state, err := runtime.ArrayStoreState(raw.Handle)",
		"inst, ok := current.(*runtime.StructInstanceValue)",
		"__able_array_struct_instance_state(inst)",
		"sourceValues = make([]runtime.Value, len(state.Values), state.Capacity)",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected mono-array from helper to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_struct_Array_from(value)",
		"base, err :=",
		"base.Elements",
		"base.Storage_handle",
		"Storage_handle:",
		"lengthVal, ok := inst.Fields[\"length\"]",
		"capacityVal, ok := inst.Fields[\"capacity\"]",
		"handleVal, ok := inst.Fields[\"storage_handle\"]",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected mono-array from helper to avoid generic Array detour %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerMonoArrayToHelperStaysDirectAtArrayBoundary(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn cloneish(values: Array i32) -> Array i32 {",
		"  values",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	body, ok := findCompiledFunction(result, "__able_array_i32_to")
	if !ok {
		t.Fatalf("could not find mono-array to helper")
	}
	for _, fragment := range []string{
		"return &runtime.ArrayValue{Elements: elems}, nil",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected mono-array to helper to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"base := &Array{",
		"__able_struct_Array_sync(base)",
		"__able_struct_Array_to(rt, base)",
		"Storage_handle",
		"runtime.ArrayStoreEnsureHandle(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected mono-array to helper to avoid generic Array detour %q:\n%s", fragment, body)
		}
	}
}
