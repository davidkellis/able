package compiler

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerDynamicBoundaryCallbackArrayConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback array conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_array_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_callback_array_success",
		"",
		"dynimport exec.dynamic_cb_array_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(values: Array i32) -> i32 { values[0]! + values[1]! }",
		"  print(invoke(callback, [4, 5]))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	expectedStdout := []string{"9"}
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

func TestCompilerDynamicBoundaryCallbackArrayConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback array conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_array_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_callback_array_failure",
		"",
		"dynimport exec.dynamic_cb_array_fail.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_array_fail")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(values: Array i32) -> i32 { values[0]! + values[1]! }",
		"  invoke(callback, [\"x\", \"y\"])",
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

func TestCompilerDynamicBoundaryCallbackHashMapConversionSuccessMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback hashmap conversion success in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_hashmap_success\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_callback_hashmap_success",
		"",
		"import able.collections.hash_map.*",
		"",
		"dynimport exec.dynamic_cb_hash_ok.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_hash_ok")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(values: HashMap String i32) -> i32 { values[\"a\"]! + values[\"b\"]! }",
		"  print(invoke(callback, #{ \"a\": 3, \"b\": 4 }))",
		"}",
	)
	if err := os.WriteFile(filepath.Join(dir, "main.able"), []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	manifest := interpreter.FixtureManifest{Entry: "main.able"}
	tree := runTreewalkerFixtureOutcome(t, dir, manifest)
	compiled, markers := runCompiledFixtureBoundaryOutcome(t, dir, manifest)

	expectedStdout := []string{"7"}
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

func TestCompilerDynamicBoundaryCallbackHashMapConversionFailureMarkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler dynamic callback hashmap conversion failure in short mode")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.yml"), []byte("name: exec_dynamic_boundary_callback_hashmap_failure\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := joinLines(
		"package exec_dynamic_boundary_callback_hashmap_failure",
		"",
		"import able.collections.hash_map.*",
		"",
		"dynimport exec.dynamic_cb_hash_fail.{invoke}",
		"",
		`pkg := dyn.def_package("exec.dynamic_cb_hash_fail")!`,
		`pkg.def("fn invoke(f, value) { f(value) }")!`,
		"",
		"fn main() -> void {",
		"  callback := fn(values: HashMap String i32) -> i32 { values[\"a\"]! + values[\"b\"]! }",
		"  invoke(callback, #{ \"a\": \"x\", \"b\": \"y\" })",
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
