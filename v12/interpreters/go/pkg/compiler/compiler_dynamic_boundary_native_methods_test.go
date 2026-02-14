package compiler

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerDynamicBoundaryNativeBoundMethodCallbackSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic native bound-method callback success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_native_bound_method_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_native_bound_method_success",
		"",
		"dynimport exec.dynamic_cb_native_runner.{invoke}",
		"",
		`runner := dyn.def_package("exec.dynamic_cb_native_runner")!`,
		`runner.def("fn invoke(f, src) -> String { f(src)!; \"ok\" }")!`,
		"",
		"fn main() -> void {",
		"  target := dyn.def_package(\"exec.dynamic_cb_native_target\")!",
		"  def_fn := target.def",
		"  print(invoke(def_fn, \"fn made() -> i32 { 7 }\"))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	expectedStdout := []string{"ok"}
	if !reflect.DeepEqual(tree.Stdout, expectedStdout) {
		t.Fatalf("treewalker stdout mismatch: expected %v got %v", expectedStdout, tree.Stdout)
	}
	if !reflect.DeepEqual(compiled.Stdout, expectedStdout) {
		t.Fatalf("compiled stdout mismatch: expected %v got %v", expectedStdout, compiled.Stdout)
	}
	if tree.Exit != 0 || compiled.Exit != 0 {
		t.Fatalf("expected successful exit: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}

func TestCompilerDynamicBoundaryNativeBoundMethodCallbackFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic native bound-method callback failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_native_bound_method_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_native_bound_method_failure",
		"",
		"dynimport exec.dynamic_cb_native_runner_fail.{invoke_bad}",
		"",
		`runner := dyn.def_package("exec.dynamic_cb_native_runner_fail")!`,
		`runner.def("fn invoke_bad(f) { f(true)! }")!`,
		"",
		"fn main() -> void {",
		"  target := dyn.def_package(\"exec.dynamic_cb_native_target_fail\")!",
		"  def_fn := target.def",
		"  invoke_bad(def_fn)",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	if tree.Exit == 0 || compiled.Exit == 0 {
		t.Fatalf("expected runtime failure: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
	}
	assertBoundaryCallValueMarkers(t, markers)
}
