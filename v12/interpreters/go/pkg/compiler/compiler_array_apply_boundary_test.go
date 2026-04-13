package compiler

import (
	"strings"
	"testing"
)

func TestCompilerArrayApplyHandleFreeRuntimeArrayBoundaryStaysDirect(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  arr := [1, 2, 3]",
		"  arr.len()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_struct_Array_apply")
	if !ok {
		t.Fatalf("could not find __able_struct_Array_apply")
	}

	rawBranchStart := strings.Index(body, "if raw, ok, nilPtr := __able_runtime_array_value(targetCurrent); ok || nilPtr {")
	if rawBranchStart < 0 {
		t.Fatalf("expected raw runtime-array branch in Array_apply:\n%s", body)
	}
	rawBranch := body[rawBranchStart:]
	rawBranchEnd := strings.Index(rawBranch, "inst, ok := targetCurrent.(*runtime.StructInstanceValue)")
	if rawBranchEnd < 0 {
		t.Fatalf("expected struct-instance branch after raw runtime-array branch:\n%s", body)
	}
	rawBranch = rawBranch[:rawBranchEnd]
	directBranchStart := strings.Index(rawBranch, "if preferredHandle == 0 {")
	if directBranchStart < 0 {
		t.Fatalf("expected direct handle-free sub-branch in Array_apply:\n%s", rawBranch)
	}
	directBranch := rawBranch[directBranchStart:]
	directBranchEnd := strings.Index(directBranch, "return nil\n\t\t}\n\t\tcapHint := __able_struct_Array_capacity_hint(value)")
	if directBranchEnd < 0 {
		t.Fatalf("expected direct handle-bearing raw-array branch after handle-free sub-branch:\n%s", rawBranch)
	}
	directBranch = directBranch[:directBranchEnd]

	for _, fragment := range []string{
		"if preferredHandle == 0 {",
		"capHint := __able_struct_Array_capacity_hint(value)",
		"raw.Elements = __able_struct_Array_clone_elements(value.Elements, capHint)",
		"value.Storage_handle = 0",
	} {
		if !strings.Contains(directBranch, fragment) {
			t.Fatalf("expected direct handle-free runtime-array apply branch to contain %q:\n%s", fragment, directBranch)
		}
	}
	for _, fragment := range []string{
		"arr, err := __able_struct_Array_runtime_value(value, preferredHandle)",
		"runtime.ArrayStoreEnsureHandle(",
		"runtime.ArrayStoreNewWithCapacity(",
	} {
		if strings.Contains(directBranch, fragment) {
			t.Fatalf("expected handle-free runtime-array apply branch to avoid %q:\n%s", fragment, directBranch)
		}
	}
}

func TestCompilerArrayApplyHandleBearingRuntimeArrayBoundaryStaysDirect(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  arr := [1, 2, 3]",
		"  arr.len()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_struct_Array_apply")
	if !ok {
		t.Fatalf("could not find __able_struct_Array_apply")
	}

	rawBranchStart := strings.Index(body, "if raw, ok, nilPtr := __able_runtime_array_value(targetCurrent); ok || nilPtr {")
	if rawBranchStart < 0 {
		t.Fatalf("expected raw runtime-array branch in Array_apply:\n%s", body)
	}
	rawBranch := body[rawBranchStart:]
	rawBranchEnd := strings.Index(rawBranch, "inst, ok := targetCurrent.(*runtime.StructInstanceValue)")
	if rawBranchEnd < 0 {
		t.Fatalf("expected struct-instance branch after raw runtime-array branch:\n%s", body)
	}
	rawBranch = rawBranch[:rawBranchEnd]
	handleBranchMarker := "return nil\n\t\t}\n\t\tcapHint := __able_struct_Array_capacity_hint(value)"
	handleBranchStart := strings.Index(rawBranch, handleBranchMarker)
	if handleBranchStart < 0 {
		t.Fatalf("expected handle-bearing raw runtime-array branch in Array_apply:\n%s", rawBranch)
	}
	handleBranch := rawBranch[handleBranchStart+len("return nil\n\t\t}\n"):]

	for _, fragment := range []string{
		"capHint := __able_struct_Array_capacity_hint(value)",
		"elems := __able_struct_Array_clone_elements(value.Elements, capHint)",
		"state, err := runtime.ArrayStoreEnsureHandle(preferredHandle, len(elems), cap(elems))",
		"state.Values = elems",
		"state.Capacity = cap(elems)",
		"raw.Handle = preferredHandle",
		"raw.Elements = state.Values",
		"value.Storage_handle = preferredHandle",
	} {
		if !strings.Contains(handleBranch, fragment) {
			t.Fatalf("expected handle-bearing runtime-array apply branch to contain %q:\n%s", fragment, handleBranch)
		}
	}
	for _, fragment := range []string{
		"arr, err := __able_struct_Array_runtime_value(value, preferredHandle)",
		"raw.Handle = arr.Handle",
		"raw.Elements = arr.Elements",
		"value.Storage_handle = arr.Handle",
	} {
		if strings.Contains(handleBranch, fragment) {
			t.Fatalf("expected handle-bearing runtime-array apply branch to avoid %q:\n%s", fragment, handleBranch)
		}
	}
}

func TestCompilerArrayApplyStructInstanceBoundaryStaysDirect(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  arr := [1, 2, 3]",
		"  arr.len()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_struct_Array_apply")
	if !ok {
		t.Fatalf("could not find __able_struct_Array_apply")
	}

	structBranchStart := strings.Index(body, "inst, ok := targetCurrent.(*runtime.StructInstanceValue)")
	if structBranchStart < 0 {
		t.Fatalf("expected struct-instance branch in Array_apply:\n%s", body)
	}
	structBranch := body[structBranchStart:]

	for _, fragment := range []string{
		"preferredHandle = runtime.ArrayStoreNewWithCapacity(__able_struct_Array_capacity_hint(value))",
		"capHint := __able_struct_Array_capacity_hint(value)",
		"elems := __able_struct_Array_clone_elements(value.Elements, capHint)",
		"state, err := runtime.ArrayStoreEnsureHandle(preferredHandle, len(elems), cap(elems))",
		"state.Values = elems",
		"state.Capacity = cap(elems)",
		"inst.Fields[\"storage_handle\"] = bridge.ToInt(preferredHandle, runtime.IntegerI64)",
	} {
		if !strings.Contains(structBranch, fragment) {
			t.Fatalf("expected struct-instance array apply branch to contain %q:\n%s", fragment, structBranch)
		}
	}
	for _, fragment := range []string{
		"arr, err := __able_struct_Array_runtime_value(value, preferredHandle)",
		"inst.Fields[\"storage_handle\"] = bridge.ToInt(arr.Handle, runtime.IntegerI64)",
		"value.Storage_handle = arr.Handle",
	} {
		if strings.Contains(structBranch, fragment) {
			t.Fatalf("expected struct-instance array apply branch to avoid %q:\n%s", fragment, structBranch)
		}
	}
}
