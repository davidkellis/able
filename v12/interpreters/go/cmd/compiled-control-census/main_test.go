package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeDirectoryResolvesRecursiveControlEffects(t *testing.T) {
	dir := t.TempDir()
	source := `package main
type __ableControl struct{}
func __able_compiled_fn_safe(n int32) (int32, *__ableControl) {
	if n <= 0 { return 0, nil }
	v, control := __able_compiled_fn_safe(n - 1)
	if control != nil { control = __able_append_control_call_frame(control, nil); return 0, control }
	return v + 1, nil
}
func __able_append_control_call_frame(control *__ableControl, _ any) *__ableControl { return control }
func __able_compiled_fn_bad() (int32, *__ableControl) { return 0, __able_raise_overflow() }
func __able_compiled_fn_propagates() (int32, *__ableControl) {
	v, control := __able_compiled_fn_bad()
	if control != nil { return 0, control }
	return v, nil
}
func __able_raise_overflow() *__ableControl { return &__ableControl{} }
`
	if err := os.WriteFile(filepath.Join(dir, "compiled.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := analyzeDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	byName := make(map[string]*functionEffect)
	for _, effect := range report.Functions {
		byName[effect.Name] = effect
	}
	if !byName["__able_compiled_fn_safe"].ControlFree {
		t.Fatalf("recursive safe function classified fallible: %#v", byName["__able_compiled_fn_safe"])
	}
	if byName["__able_compiled_fn_bad"].ControlFree || byName["__able_compiled_fn_propagates"].ControlFree {
		t.Fatalf("fallible chain classified control-free: bad=%#v propagates=%#v", byName["__able_compiled_fn_bad"], byName["__able_compiled_fn_propagates"])
	}
	if report.Summary.DirectCompiledFunctions != 3 || report.Summary.ControlFreeDirect != 1 {
		t.Fatalf("unexpected summary: %#v", report.Summary)
	}
}

func TestAnalyzeDirectoryTreatsHandledControlAsExternallyFree(t *testing.T) {
	dir := t.TempDir()
	source := `package main
type __ableControl struct{}
func __able_compiled_fn_bad() (int32, *__ableControl) { return 0, __able_raise() }
func __able_compiled_fn_handles() (int32, *__ableControl) {
	v, control := __able_compiled_fn_bad()
	if control != nil { return 7, nil }
	return v, nil
}
func __able_raise() *__ableControl { return &__ableControl{} }
`
	if err := os.WriteFile(filepath.Join(dir, "compiled.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := analyzeDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, effect := range report.Functions {
		if effect.Name == "__able_compiled_fn_handles" && !effect.ControlFree {
			t.Fatalf("handled control should not escape: %#v", effect)
		}
	}
}

func TestAnalyzeDirectoryTreatsReassignedConstructedControlAsFallible(t *testing.T) {
	dir := t.TempDir()
	source := `package main
type __ableControl struct{}
func __able_compiled_fn_safe() (int32, *__ableControl) { return 1, nil }
func __able_compiled_fn_reassigns() (int32, *__ableControl) {
	v, control := __able_compiled_fn_safe()
	control = &__ableControl{}
	return v, control
}
`
	if err := os.WriteFile(filepath.Join(dir, "compiled.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := analyzeDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, effect := range report.Functions {
		if effect.Name == "__able_compiled_fn_reassigns" && effect.ControlFree {
			t.Fatalf("constructed control reassignment classified control-free: %#v", effect)
		}
	}
}
