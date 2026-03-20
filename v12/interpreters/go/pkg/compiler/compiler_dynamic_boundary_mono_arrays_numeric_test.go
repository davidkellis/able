package compiler

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerDynamicBoundaryMonoArrayU32CallbackConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic mono-array u32 callback conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_mono_array_u32_callback_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_mono_array_u32_callback_success",
		"",
		"dynimport exec.dynamic_cb_array_u32_mono_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_u32_mono_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(values: Array u32) -> u32 { values[0]! + values[1]! }",
		"  nums: Array u32 := [4_u32, 5_u32]",
		"  print(invoke(callback, nums))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	expectedStdout := []string{"9"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryMonoArrayF32CallbackConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic mono-array f32 callback conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_mono_array_f32_callback_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_mono_array_f32_callback_success",
		"",
		"dynimport exec.dynamic_cb_array_f32_mono_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_f32_mono_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(values: Array f32) -> f32 { values[0]! + values[1]! }",
		"  nums: Array f32 := [1.5_f32, 2.25_f32]",
		"  print(invoke(callback, nums))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcomeWithOptions(t, dir, manifest, Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	expectedStdout := []string{"3.75"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d stderr=%v", tree.Exit, compiled.Exit, compiled.Stderr)
	}
	assertBoundaryCallValueMarkers(t, markers)
}
