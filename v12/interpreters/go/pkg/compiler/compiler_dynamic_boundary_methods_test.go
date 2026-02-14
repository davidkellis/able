package compiler

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerDynamicBoundaryBoundMethodCallbackSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic bound-method callback success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_bound_method_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_bound_method_success",
		"",
		"struct Counter { value: i32 }",
		"",
		"methods Counter {",
		"  fn #add(delta: i32) -> i32 { #value + delta }",
		"}",
		"",
		"dynimport exec.dynamic_cb_bound_method_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_bound_method_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  counter := Counter { value: 10 }",
		"  add_fn := counter.add",
		"  print(invoke(add_fn, 5))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	expectedStdout := []string{"15"}
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

func TestCompilerDynamicBoundaryBoundMethodCallbackFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic bound-method callback failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_bound_method_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_bound_method_failure",
		"",
		"struct Counter { value: i32 }",
		"",
		"methods Counter {",
		"  fn #add(delta: i32) -> i32 { #value + delta }",
		"}",
		"",
		"dynimport exec.dynamic_cb_bound_method_fail.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_bound_method_fail")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  counter := Counter { value: 10 }",
		"  add_fn := counter.add",
		"  invoke(add_fn, true)",
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
